// ============================================================
// LyricCanvas — 数据模型定义
// ============================================================
package models

import "time"

// Project 歌曲项目
type Project struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	Canvas      Canvas        `json:"canvas"`
	ChatHistory []ChatMessage `json:"chatHistory"`
}

// Canvas 画布数据
type Canvas struct {
	Viewport  Viewport     `json:"viewport"`
	Nodes     []LyricNode  `json:"nodes"`
	Groups    []GroupLabel `json:"groups"`
	DraftZone DraftZone    `json:"draftZone"`
}

// Viewport 视口状态
type Viewport struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

// LyricNode 歌词节点
type LyricNode struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Text     string        `json:"text"`
	Position Position      `json:"position"`
	Style    NodeStyle     `json:"style,omitempty"`
	Data     LyricNodeData `json:"data"`
}

// LyricNodeData 节点附加数据
type LyricNodeData struct {
	GroupID   *string `json:"groupId"`
	Order     int     `json:"order"`
	Locked    bool    `json:"locked"`
	RhymeHint string  `json:"rhymeHint,omitempty"`
}

// NodeStyle 节点样式
type NodeStyle struct {
	Color    string `json:"color,omitempty"`
	FontSize int    `json:"fontSize,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// GroupLabel 结构标签（主歌/副歌等）
type GroupLabel struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Label    string   `json:"label"`
	Category string   `json:"category"`
	Position Position `json:"position"`
	NodeIDs  []string `json:"nodeIds"`
	Color    string   `json:"color"`
}

// Position 画布坐标
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// DraftZone 草稿区
type DraftZone struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ChatMessage 对话消息
type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ---------- API 请求/响应类型 ----------

// CreateProjectRequest 新建项目请求
type CreateProjectRequest struct {
	Name string `json:"name"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        string        `json:"name,omitempty"`
	Canvas      *Canvas       `json:"canvas,omitempty"`
	ChatHistory []ChatMessage `json:"chatHistory,omitempty"`
}

// ProjectListItem 项目列表项（不含画布和对话数据）
type ProjectListItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	NodeCount int       `json:"nodeCount"`
}

// ChatRequest 对话请求
type ChatRequest struct {
	ProjectID    string   `json:"projectId"`
	Model        string   `json:"model,omitempty"`
	SystemPrompt string   `json:"systemPrompt,omitempty"`
	Message      string   `json:"message"`
	ActionType   string   `json:"actionType,omitempty"`
	TargetFields []string `json:"targetFields,omitempty"`
	Temperature  float64  `json:"temperature,omitempty"`
	MaxTokens    int      `json:"maxTokens,omitempty"`
	Stream       bool     `json:"stream,omitempty"`
}

// RhymeResult 押韵查询结果
type RhymeResult struct {
	Char        string      `json:"char"`
	Pinyin      string      `json:"pinyin"`
	Final       string      `json:"final"`
	Rhymes      []RhymeChar `json:"rhymes"`
	CommonWords []string    `json:"commonWords,omitempty"`
}

// RhymeChar 押韵字
type RhymeChar struct {
	Char   string `json:"char"`
	Pinyin string `json:"pinyin"`
}

// BatchRhymeRequest 批量押韵请求
type BatchRhymeRequest struct {
	Chars []string `json:"chars"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status       string `json:"status"`
	DeepSeekOK   bool   `json:"deepseekOk"`
	ProjectCount int    `json:"projectCount"`
}
