// ============================================================
// LyricCanvas — DeepSeek 对话 HTTP 处理器（含 SSE 流式）
// 增强：prompt 模板引擎 + 快捷模板/操作类型端点 + 智能歌词拆句
// ============================================================
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"lyriccanvas/models"
	"lyriccanvas/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	dsClient *services.DeepSeekClient
	store    *services.ProjectStore
}

func NewChatHandler(dsClient *services.DeepSeekClient, store *services.ProjectStore) *ChatHandler {
	return &ChatHandler{dsClient: dsClient, store: store}
}

// Send 发送对话（POST /api/chat）
func (h *ChatHandler) Send(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}
	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息不能为空"})
		return
	}

	model := req.Model
	if model == "" {
		model = "deepseek-v4-flash"
	}
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 1.0
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	ctx := services.ChatContext{}
	if req.ActionType != "" {
		ctx.TargetFields = parseTargetFields(req.TargetFields)
	}

	if req.ProjectID != "" && h.store != nil {
		proj, err := h.store.Get(req.ProjectID)
		if err == nil {
			var lyricsLines []string
			for _, node := range proj.Canvas.Nodes {
				lyricsLines = append(lyricsLines, node.Text)
			}
			ctx.Lyrics = strings.Join(lyricsLines, "\n")
			ctx.SongName = proj.Name
		}
	}

	var messages []services.ChatMessageDS
	var history []services.ChatMessageDS
	if req.ProjectID != "" && h.store != nil {
		proj, err := h.store.Get(req.ProjectID)
		if err == nil && len(proj.ChatHistory) > 0 {
			historyStart := 0
			if len(proj.ChatHistory) > 20 {
				historyStart = len(proj.ChatHistory) - 20
			}
			for _, msg := range proj.ChatHistory[historyStart:] {
				history = append(history, services.ChatMessageDS{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}
		}
	}

	actionType := services.ActionType(req.ActionType)
	if actionType == "" {
		actionType = services.ActionGenerate
	}
	messages = services.BuildChatMessages(actionType, ctx, req.Message, history)

	if req.Stream {
		h.handleStream(c, model, messages, temperature, maxTokens, req.ProjectID)
		return
	}

	reply, err := h.dsClient.Chat(model, messages, temperature, maxTokens)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.saveChatHistory(req.ProjectID, req.Message, reply)
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

func (h *ChatHandler) handleStream(c *gin.Context, model string, messages []services.ChatMessageDS, temperature float64, maxTokens int, projectID string) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	contentCh, errCh := h.dsClient.ChatStream(model, messages, temperature, maxTokens)
	var fullReply strings.Builder
	flusher, _ := c.Writer.(http.Flusher)

	for {
		select {
		case content, ok := <-contentCh:
			if !ok {
				fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
				if flusher != nil {
					flusher.Flush()
				}
				h.saveChatHistory(projectID, messages[len(messages)-1].Content, fullReply.String())
				return
			}
			fullReply.WriteString(content)
			data, _ := json.Marshal(map[string]string{"content": content})
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}
		case err, ok := <-errCh:
			if ok && err != nil {
				fmt.Fprintf(c.Writer, "data: {\"error\":\"%s\"}\n\n", err.Error())
				if flusher != nil {
					flusher.Flush()
				}
			}
			return
		}
	}
}

func (h *ChatHandler) saveChatHistory(projectID, userMsg, aiReply string) {
	if projectID == "" || h.store == nil {
		return
	}
	now := time.Now()
	_ = h.store.AppendChatMessage(projectID, models.ChatMessage{
		ID: "msg_" + uuid.New().String()[:8], Role: "user", Content: userMsg, Timestamp: now,
	})
	_ = h.store.AppendChatMessage(projectID, models.ChatMessage{
		ID: "msg_" + uuid.New().String()[:8], Role: "assistant", Content: aiReply, Timestamp: now,
	})
}

func (h *ChatHandler) GetHistory(c *gin.Context) {
	projectID := c.Param("projectId")
	proj, err := h.store.Get(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}
	c.JSON(http.StatusOK, proj.ChatHistory)
}

func (h *ChatHandler) ClearHistory(c *gin.Context) {
	projectID := c.Param("projectId")
	_, err := h.store.Update(projectID, &models.UpdateProjectRequest{
		ChatHistory: []models.ChatMessage{},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *ChatHandler) GetSystemPrompt(c *gin.Context) {
	ctx := services.ChatContext{}
	prompt := services.BuildSystemPrompt(ctx)
	c.JSON(http.StatusOK, gin.H{"systemPrompt": prompt})
}

func (h *ChatHandler) GetTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"categories": services.GetTemplateCategories()})
}

func (h *ChatHandler) GetActions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"actions": services.GetActionTypes()})
}

// ---------- 智能歌词拆句 ----------

// ParseLyrics 智能拆句 —— 用 DeepSeek 清洗歌词文本，拆分为干净句子
// POST /api/lyrics/parse  body: { "text": "原始歌词文本" }
func (h *ChatHandler) ParseLyrics(c *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Text) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供歌词文本"})
		return
	}

	raw := strings.TrimSpace(req.Text)
	if len(raw) > 8000 {
		raw = raw[:8000]
	}

	systemPrompt := "你是一个歌词文本清洗工具。你的唯一任务是将原始歌词文本拆分为独立的句子。\n\n规则：\n1. 去除所有特殊符号、markdown 标记（如 ## ** ~~ 等）、HTML 标签、编号前缀（如 1. 2. ① 等）\n2. 去除所有英文、数字、URL\n3. 将文本按自然断句拆分为独立句子，每行一句\n4. 每句末尾统一加上中文标点（。或！或？）\n5. 只输出拆分后的句子，每行一句，不要序号，不要任何额外解释\n6. 如果某行不是中文歌词内容（纯符号、纯数字等），直接跳过"

	userPrompt := fmt.Sprintf("请清洗并拆分以下歌词文本：\n\n%s", raw)

	messages := []services.ChatMessageDS{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	reply, err := h.dsClient.Chat("deepseek-chat", messages, 0.3, 2048)
	if err != nil {
		lines := fallbackSplit(raw)
		c.JSON(http.StatusOK, gin.H{"lines": lines, "fallback": true})
		return
	}

	lines := cleanLines(reply)
	if len(lines) == 0 {
		lines = fallbackSplit(raw)
		c.JSON(http.StatusOK, gin.H{"lines": lines, "fallback": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lines": lines})
}

func cleanLines(reply string) []string {
	var result []string
	for _, line := range strings.Split(reply, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimLeft(line, "0123456789.、．①②③④⑤⑥⑦⑧⑨⑩ ")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lastChar := line[len(line)-1:]
		if !strings.Contains("。！？，；、）)", lastChar) {
			line += "。"
		}
		result = append(result, line)
	}
	return result
}

func fallbackSplit(text string) []string {
	var result []string
	text = strings.NewReplacer(
		"##", "", "**", "", "~~", "", "```", "",
		"（", "", "）", "", "(", "", ")", "",
		"「", "", "」", "", "『", "", "』", "",
		"《", "", "》", "", "〈", "", "〉", "",
	).Replace(text)

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.FieldsFunc(line, func(r rune) bool {
			return strings.ContainsRune("。！？；，、\n", r)
		})
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if len(p) < 2 {
				continue
			}
			hasChinese := false
			for _, r := range p {
				if unicode.Is(unicode.Han, r) {
					hasChinese = true
					break
				}
			}
			if !hasChinese {
				continue
			}
			result = append(result, p+"。")
		}
	}
	return result
}

func parseTargetFields(raw []string) []services.TargetField {
	var result []services.TargetField
	for _, s := range raw {
		switch s {
		case "songIdea":
			result = append(result, services.FieldSongIdea)
		case "lyrics":
			result = append(result, services.FieldLyrics)
		case "songName":
			result = append(result, services.FieldSongName)
		}
	}
	return result
}

func extractJSON(content string) (map[string]string, bool) {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		var result map[string]string
		if err := json.Unmarshal([]byte(content), &result); err == nil {
			return result, true
		}
	}
	return nil, false
}

func ReadBody(r *http.Request) string {
	body, _ := io.ReadAll(r.Body)
	return string(body)
}
