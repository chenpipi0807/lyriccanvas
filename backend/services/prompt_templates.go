// ============================================================
// LyricCanvas — Prompt 模板引擎
// 参考 ai-roundtable-extension prompt-templates.js
// ============================================================
package services

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ---------- 操作类型与目标字段 ----------

type ActionType string

const (
	ActionPolish         ActionType = "polish"
	ActionRewrite        ActionType = "rewrite"
	ActionContinue       ActionType = "continue"
	ActionGenerate       ActionType = "generate"
	ActionCompleteCreate ActionType = "complete_create"
)

type TargetField string

const (
	FieldSongIdea TargetField = "songIdea"
	FieldLyrics   TargetField = "lyrics"
	FieldSongName TargetField = "songName"
)

var FieldLabelMap = map[TargetField]string{
	FieldSongIdea: "风格描述",
	FieldLyrics:   "歌词",
	FieldSongName: "歌曲名称",
}

type ChatContext struct {
	Styles       []string      `json:"styles,omitempty"`
	SongIdea     string        `json:"songIdea,omitempty"`
	Lyrics       string        `json:"lyrics,omitempty"`
	SongName     string        `json:"songName,omitempty"`
	LockedFields []TargetField `json:"lockedFields,omitempty"`
	TargetFields []TargetField `json:"targetFields,omitempty"`
}

type PromptPair struct {
	System string `json:"system"`
	User   string `json:"user"`
}

type QuickTemplate struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Label    string `json:"label"`
	Content  string `json:"content"`
	Hint     string `json:"hint,omitempty"`
}

type TemplateCategory struct {
	Key       string          `json:"key"`
	Label     string          `json:"label"`
	Templates []QuickTemplate `json:"templates"`
}

// ---------- 创作法则 ----------

const songwritingRules = `创作法则（全局遵守）:

【结构完整】
每首歌必须具备完整的专业结构层次，不可简单敷衍：
前奏段 → 主歌A → 主歌B → 副歌(高潮) → 间奏 → 主歌B'(变体) → 副歌(重复) → 尾声
主歌与副歌之间应有明显的情感与节奏递进。

【副歌传唱】
副歌至少包含 2-4 句核心重复句式，该句式必须：
- 朗朗上口、节奏鲜明、易于记忆
- 具备高传唱度（普通人听一遍就能跟唱）
- 重复时可有微小变化（如换一两个字强化递进），但主旋律句式保持一致

【段落对仗】
相邻主歌段落（主歌A与主歌B）应句数相近、句长对称，
形成结构上的对仗美感，不可一段臃肿一段单薄。

【标点必加】★
每句歌词末尾必须添加标点符号（。，！？），
标点是演唱的换气点和节奏断句，绝不可连续多句无标点，
否则 AI 演唱时会因缺乏断句而急促"咬嘴"。

【尾韵优先】
相邻句尾字尽量押韵（脚韵），但优先级为：
自然流畅的表达 > 押韵
不可为凑韵脚而扭曲句意或填入生僻词。
能自然押韵时务必保留；需要牺牲表达则放弃押韵。

【文辞质量】
歌词用词优美、有文学感和画面感，避免口语大白话。
重复出现的关键词或句式应有递进、转折或深化意义，
不可机械填充。意象具体，少用抽象空洞的形容词堆砌。`

// ---------- System Prompt ----------

func BuildSystemPrompt(ctx ChatContext) string {
	styles := "未指定"
	if len(ctx.Styles) > 0 {
		styles = strings.Join(ctx.Styles, "、")
	}
	hasIdea := strings.TrimSpace(ctx.SongIdea) != ""
	hasLyrics := strings.TrimSpace(ctx.Lyrics) != ""
	hasName := strings.TrimSpace(ctx.SongName) != ""
	var contextSummary string
	if hasIdea || hasLyrics || hasName {
		contextSummary = "\n当前已有内容状态:\n"
		if hasIdea {
			contextSummary += "- 风格描述: 已有内容\n"
		} else {
			contextSummary += "- 风格描述: 空（需要生成）\n"
		}
		if hasLyrics {
			contextSummary += "- 歌词: 已有内容\n"
		} else {
			contextSummary += "- 歌词: 空（需要生成）\n"
		}
		if hasName {
			contextSummary += "- 歌曲名称: 已有内容\n"
		} else {
			contextSummary += "- 歌曲名称: 空（需要生成）\n"
		}
	}
	var lockConstraint string
	if len(ctx.LockedFields) > 0 {
		var labels []string
		for _, f := range ctx.LockedFields {
			if label, ok := FieldLabelMap[f]; ok {
				labels = append(labels, label)
			}
		}
		if len(labels) > 0 {
			lockConstraint = fmt.Sprintf(
				"\n⚠️ 以下字段已被用户锁定，【严禁修改，必须原样输出其现有内容】：%s\n",
				strings.Join(labels, "、"),
			)
		}
	}
	targetNote := targetFieldsNote(ctx)
	return fmt.Sprintf(
		"你是一位资深音乐创作助手，专注于协助用户创作中文歌词、歌曲概念和歌曲名称。\n请根据用户的创作意图，生成高质量、有韵律感、情感真挚的内容。\n\n%s\n\n当前歌曲风格: %s\n%s%s%s\n输出要求:\n- 直接给出完整内容，不要添加额外解释\n- 严格按照用户问题中指定的格式输出\n- 保持与原文相同的行数结构（如果是对已有内容的修改）\n- 如果是多字段生成（同时生成风格描述、歌词、歌曲名称），请使用 JSON 格式输出，键名分别为 songIdea（风格描述）、lyrics（歌词）、songName（歌曲名称）。JSON 中的换行符保持原样（\\n），不要使用真实换行打断 JSON 结构。重要：确保输出为合法 JSON，不要添加 markdown 代码块标记。\n- 如果只生成歌词，每句歌词末尾必须添加标点符号（。，！？等），确保断句清晰，不要通篇无标点。\n \n重要：如果某部分已有内容，在其基础上优化或保持一致性；如果某部分为空，需要全新创作。",
		songwritingRules, styles, contextSummary, lockConstraint, targetNote,
	)
}

func targetFieldsNote(ctx ChatContext) string {
	if len(ctx.TargetFields) == 0 || len(ctx.TargetFields) >= 3 {
		return ""
	}
	var labels []string
	for _, f := range ctx.TargetFields {
		if label, ok := FieldLabelMap[f]; ok {
			labels = append(labels, label)
		}
	}
	if len(labels) == 0 {
		return ""
	}
	return fmt.Sprintf("\n本次仅需生成/修改: %s（不要改动其他字段）", strings.Join(labels, "、"))
}

// ---------- User Prompt 分发 ----------

func BuildUserPrompt(actionType ActionType, ctx ChatContext, userInput string) string {
	if actionType == ActionCompleteCreate {
		return buildCompleteCreatePrompt(ctx, userInput)
	}
	switch {
	case hasTargetField(ctx.TargetFields, FieldSongIdea):
		return buildIdeaPrompt(ctx, userInput)
	case hasTargetField(ctx.TargetFields, FieldSongName):
		return buildSongNamePrompt(ctx, userInput)
	default:
		return buildLyricsPrompt(actionType, ctx, userInput)
	}
}

func BuildPrompt(actionType ActionType, ctx ChatContext, userInput string) PromptPair {
	return PromptPair{System: BuildSystemPrompt(ctx), User: BuildUserPrompt(actionType, ctx, userInput)}
}

// ---------- 各场景 Prompt ----------

func buildCompleteCreatePrompt(ctx ChatContext, userInput string) string {
	styles := "未指定"
	if len(ctx.Styles) > 0 {
		styles = strings.Join(ctx.Styles, "、")
	}
	extra := ""
	if userInput != "" {
		extra = fmt.Sprintf("\n用户额外要求: %s", userInput)
	}
	var existingParts, missingParts []string
	if strings.TrimSpace(ctx.SongIdea) != "" {
		existingParts = append(existingParts, "风格描述")
	} else {
		missingParts = append(missingParts, "风格描述")
	}
	if strings.TrimSpace(ctx.Lyrics) != "" {
		existingParts = append(existingParts, "歌词")
	} else {
		missingParts = append(missingParts, "歌词")
	}
	if strings.TrimSpace(ctx.SongName) != "" {
		existingParts = append(existingParts, "歌曲名称")
	} else {
		missingParts = append(missingParts, "歌曲名称")
	}
	var contextNote string
	if len(existingParts) > 0 {
		contextNote = "\n已有内容（请参考并保持一致性）:\n"
		if strings.TrimSpace(ctx.SongIdea) != "" {
			contextNote += fmt.Sprintf("风格描述: %s\n", strings.TrimSpace(ctx.SongIdea))
		}
		if strings.TrimSpace(ctx.Lyrics) != "" {
			contextNote += fmt.Sprintf("歌词:\n%s\n", strings.TrimSpace(ctx.Lyrics))
		}
		if strings.TrimSpace(ctx.SongName) != "" {
			contextNote += fmt.Sprintf("歌曲名称: %s\n", strings.TrimSpace(ctx.SongName))
		}
	}
	if len(missingParts) > 0 {
		contextNote += fmt.Sprintf("\n需要生成的部分: %s", strings.Join(missingParts, "、"))
	}
	target := ctx.TargetFields
	if len(target) == 0 {
		target = []TargetField{FieldSongIdea, FieldLyrics, FieldSongName}
	}
	needIdea := hasTargetField(target, FieldSongIdea)
	needLyrics := hasTargetField(target, FieldLyrics)
	needName := hasTargetField(target, FieldSongName)
	var fieldInstructions string
	if needIdea {
		fieldInstructions += "- songIdea: 风格描述（50-200字）\n"
	}
	if needLyrics {
		fieldInstructions += "- lyrics: 歌词正文（每句必须带标点符号）\n"
	}
	if needName {
		fieldInstructions += "- songName: 歌曲名称\n"
	}
	n := 1
	notes := fmt.Sprintf("%d. 输出必须是合法 JSON，键名严格为 songIdea、lyrics、songName\n", n)
	n++
	notes += fmt.Sprintf("%d. 歌词每句末尾必须有标点（。，！？），确保断句清晰\n", n)
	n++
	notes += fmt.Sprintf("%d. JSON 中的换行符必须用 \\n 表示，不能用真实换行\n", n)
	n++
	if !needIdea && strings.TrimSpace(ctx.SongIdea) != "" {
		notes += fmt.Sprintf("%d. 风格描述已有内容，不要输出 songIdea\n", n)
		n++
	}
	if !needLyrics && strings.TrimSpace(ctx.Lyrics) != "" {
		notes += fmt.Sprintf("%d. 歌词已有内容，不要输出 lyrics\n", n)
		n++
	}
	if !needName && strings.TrimSpace(ctx.SongName) != "" {
		notes += fmt.Sprintf("%d. 歌曲名称已有内容，不要输出 songName\n", n)
		n++
	}
	notes += fmt.Sprintf("%d. 如果某部分已有内容，在其基础上优化或保持一致性；如果某部分需要生成，请全新创作", n)
	return fmt.Sprintf(
		"请根据以下描述，完成歌曲创作。\n\n风格: %s\n创作想法: %s\n%s\n%s\n\n请严格按照 JSON 格式输出，键名分别为 songIdea、lyrics、songName。只输出纯 JSON，不要加代码块标记。JSON 中的换行使用 \\n 转义，不要插入真实换行打断 JSON 结构。\n只需包含以下字段：\n%s\n示例格式：\n{\"songIdea\":\"一首温暖的流行歌曲，...\",\"lyrics\":\"第一句歌词。\\n第二句歌词。\",\"songName\":\"歌名\"}\n\n注意：\n%s",
		styles, orDefault(ctx.SongIdea, "(由用户直接描述)"), contextNote, extra, fieldInstructions, notes,
	)
}

func buildLyricsPrompt(actionType ActionType, ctx ChatContext, userInput string) string {
	styles := "未指定"
	if len(ctx.Styles) > 0 {
		styles = strings.Join(ctx.Styles, "、")
	}
	extra := ""
	if userInput != "" {
		extra = fmt.Sprintf("\n用户额外要求: %s", userInput)
	}
	var contextNote string
	if strings.TrimSpace(ctx.SongName) != "" {
		contextNote = fmt.Sprintf("\n歌曲名称: %s", strings.TrimSpace(ctx.SongName))
	}
	switch actionType {
	case ActionPolish:
		return fmt.Sprintf("请润色以下歌词，提升文学性和韵律感，保持原有主题和情感。\n\n风格: %s\n歌曲想法: %s%s\n当前歌词:\n%s\n%s\n\n请直接输出润色后的完整歌词。每句末尾必须添加标点符号（。，！？等）。",
			styles, orDefault(ctx.SongIdea, "(无)"), contextNote, orDefault(ctx.Lyrics, "(无内容)"), extra)
	case ActionRewrite:
		return fmt.Sprintf("请用不同的表达方式重写以下歌词，保持主题不变。\n\n风格: %s\n歌曲想法: %s%s\n当前歌词:\n%s\n%s\n\n请直接输出重写后的完整歌词。每句末尾必须添加标点符号（。，！？等）。",
			styles, orDefault(ctx.SongIdea, "(无)"), contextNote, orDefault(ctx.Lyrics, "(无内容)"), extra)
	case ActionContinue:
		return fmt.Sprintf("请根据已有内容续写歌词，保持风格一致。\n\n风格: %s\n歌曲想法: %s%s\n已有歌词:\n%s\n%s\n\n请直接输出续写后的完整歌词（包含已有内容 + 续写部分）。每句末尾必须添加标点符号（。，！？等）。",
			styles, orDefault(ctx.SongIdea, "(无)"), contextNote, orDefault(ctx.Lyrics, "(无内容)"), extra)
	default:
		return fmt.Sprintf("请根据以下描述创作歌词。\n\n风格: %s\n歌曲想法: %s%s\n%s\n\n请直接输出创作的完整歌词。每句末尾必须添加标点符号（。，！？等）。",
			styles, orDefault(ctx.SongIdea, "(无)"), contextNote, extra)
	}
}

func buildSongNamePrompt(ctx ChatContext, userInput string) string {
	styles := "未指定"
	if len(ctx.Styles) > 0 {
		styles = strings.Join(ctx.Styles, "、")
	}
	extra := ""
	if userInput != "" {
		extra = fmt.Sprintf("\n用户额外要求: %s", userInput)
	}
	var contextNote string
	if strings.TrimSpace(ctx.SongName) != "" {
		contextNote = fmt.Sprintf("\n已有歌曲名称（可参考或改进）: %s", strings.TrimSpace(ctx.SongName))
	}
	return fmt.Sprintf(
		"请为这首歌起 3-5 个有创意的歌曲名称。\n\n风格: %s\n歌曲想法: %s\n歌词:\n%s%s\n%s\n\n请直接输出歌曲名称列表，每行一个。",
		styles, orDefault(ctx.SongIdea, "(无)"), orDefault(ctx.Lyrics, "(无)"), contextNote, extra,
	)
}

func buildIdeaPrompt(ctx ChatContext, userInput string) string {
	styles := "未指定"
	if len(ctx.Styles) > 0 {
		styles = strings.Join(ctx.Styles, "、")
	}
	extra := ""
	if userInput != "" {
		extra = fmt.Sprintf("\n用户额外要求: %s", userInput)
	}
	var contextNote string
	if strings.TrimSpace(ctx.Lyrics) != "" {
		contextNote += fmt.Sprintf("\n已有歌词（可参考）:\n%s", strings.TrimSpace(ctx.Lyrics))
	}
	if strings.TrimSpace(ctx.SongName) != "" {
		contextNote += fmt.Sprintf("\n歌曲名称: %s", strings.TrimSpace(ctx.SongName))
	}
	return fmt.Sprintf(
		"请优化以下歌曲创作想法，使其更具体、更有画面感、更具音乐性。\n\n风格: %s\n当前想法: %s%s\n%s\n\n请直接输出优化后的歌曲想法。",
		styles, orDefault(ctx.SongIdea, "(无)"), contextNote, extra,
	)
}

// ---------- 辅助 ----------

func hasTargetField(fields []TargetField, target TargetField) bool {
	for _, f := range fields {
		if f == target {
			return true
		}
	}
	return false
}

func orDefault(val, defaultVal string) string {
	if strings.TrimSpace(val) == "" {
		return defaultVal
	}
	return val
}

// ---------- 快捷模板数据 ----------

func GetTemplateCategories() []TemplateCategory {
	return []TemplateCategory{
		{
			Key: "structure", Label: "歌词结构", Templates: []QuickTemplate{
				{ID: "struct-pop", Category: "structure", Label: "流行歌曲结构", Hint: "主歌+副歌经典结构", Content: "请按照以下结构创作歌词：\n【前奏】2句引入氛围\n【主歌A】4句叙述故事\n【主歌B】4句情感递进\n【副歌】4句高潮传唱句\n【主歌B'】4句变体呼应\n【副歌重复】\n【尾声】2句收束"},
				{ID: "struct-ballad", Category: "structure", Label: "民谣叙事结构", Hint: "叙事+抒情交替", Content: "请按照以下民谣叙事结构创作：\n【主歌A】4句场景描写\n【主歌B】4句人物内心\n【副歌】4句核心情感\n【主歌A'】4句场景呼应\n【副歌重复】"},
				{ID: "struct-ancient", Category: "structure", Label: "古风结构", Hint: "四段式古风", Content: "请按照以下古风结构创作：\n【起】4句铺陈意境\n【承】4句深化主题\n【转】4句转折升华\n【合】4句收束余韵"},
				{ID: "struct-rap", Category: "structure", Label: "说唱结构", Hint: "verse+hook", Content: "请按照说唱结构创作：\n【Intro】2句开场\n【Verse 1】8句押韵叙述\n【Hook】4句重复核心\n【Verse 2】8句递进\n【Hook重复】\n【Outro】2句结尾"},
			},
		},
		{
			Key: "emotion", Label: "情感主题", Templates: []QuickTemplate{
				{ID: "emo-love", Category: "emotion", Label: "爱情 - 甜蜜", Hint: "初恋、告白、陪伴", Content: "请创作一首关于甜蜜爱情的歌词，主题围绕初恋的心动、告白的勇气或陪伴的温暖。用具体场景描写代替抽象形容词。"},
				{ID: "emo-heartbreak", Category: "emotion", Label: "爱情 - 伤感", Hint: "分手、思念、遗憾", Content: "请创作一首关于失恋的歌词，主题围绕分手的痛苦、对过去的思念或未说出口的遗憾。情感克制而深沉，避免直白的哭诉。"},
				{ID: "emo-nostalgia", Category: "emotion", Label: "怀旧 / 青春", Hint: "校园、故乡、时光", Content: "请创作一首关于怀旧或青春的歌词，主题围绕校园时光、故乡记忆或流逝的岁月。用具体的物品和场景唤起共鸣。"},
				{ID: "emo-dream", Category: "emotion", Label: "励志 / 梦想", Hint: "追梦、坚持、成长", Content: "请创作一首励志主题的歌词，主题围绕追逐梦想、坚持不懈或自我成长。语言有力但不空洞，用具体经历支撑主题。"},
				{ID: "emo-nature", Category: "emotion", Label: "自然 / 季节", Hint: "春夏秋冬、山水", Content: "请创作一首以自然或季节为主题的歌词，用景物描写寄托情感，将自然意象与人的心境融合。"},
				{ID: "emo-friendship", Category: "emotion", Label: "友情 / 亲情", Hint: "兄弟、父母、陪伴", Content: "请创作一首关于友情或亲情的歌词，主题围绕真挚的关系、陪伴与感恩。语言温暖真挚。"},
			},
		},
		{
			Key: "rhetoric", Label: "修辞手法", Templates: []QuickTemplate{
				{ID: "rhet-metaphor", Category: "rhetoric", Label: "隐喻手法", Hint: "用自然意象表达情感", Content: "请在歌词中大量使用隐喻手法，将抽象情感映射到具体自然意象上（如：用'雨'喻思念，用'风'喻自由，用'灯塔'喻希望）。"},
				{ID: "rhet-antithesis", Category: "rhetoric", Label: "对仗/对比", Hint: "前后句对仗工整", Content: "请在歌词中运用对仗和对比手法，相邻句式结构对称，通过对比强化情感张力。"},
				{ID: "rhet-repetition", Category: "rhetoric", Label: "反复/叠句", Hint: "核心句式反复出现", Content: "请在歌词中使用反复和叠句手法，一个核心句式在副歌中反复出现（每次可有微调），强化记忆点和情感冲击。"},
				{ID: "rhet-allusion", Category: "rhetoric", Label: "典故/化用", Hint: "化用古诗词意境", Content: "请在歌词中化用中国古典诗词的意境和名句，融入现代语境，使歌词兼具古典韵味和现代感。"},
			},
		},
		{
			Key: "style", Label: "风格流派", Templates: []QuickTemplate{
				{ID: "style-pop", Category: "style", Label: "流行", Hint: "旋律优先、通俗易懂", Content: "请以流行音乐风格创作歌词，注重旋律感和传唱度，语言通俗但不失优美。"},
				{ID: "style-rock", Category: "style", Label: "摇滚", Hint: "力量感、反叛精神", Content: "请以摇滚风格创作歌词，语言有力直接，富有反叛精神和力量感。"},
				{ID: "style-folk", Category: "style", Label: "民谣", Hint: "叙事性、生活化", Content: "请以民谣风格创作歌词，注重叙事性和生活细节，语言朴素但有画面感。"},
				{ID: "style-rnb", Category: "style", Label: "R&B / 灵魂", Hint: "律动感、情感细腻", Content: "请以 R&B 风格创作歌词，注重律动感和情感的细腻表达，语言流畅有韵味。"},
				{ID: "style-ancient", Category: "style", Label: "中国风/古风", Hint: "古典意象、典雅", Content: "请以中国风/古风风格创作歌词，大量使用古典意象和典雅词藻，营造诗画意境。"},
				{ID: "style-electronic", Category: "style", Label: "电子/氛围", Hint: "迷幻、未来感", Content: "请以电子/氛围音乐风格创作歌词，语言迷幻朦胧，富有未来感和空间感。"},
			},
		},
	}
}

func GetActionTypes() []map[string]string {
	return []map[string]string{
		{"type": string(ActionPolish), "label": "✨ 润色", "hint": "润色当前内容，提升文学性和韵律感"},
		{"type": string(ActionRewrite), "label": "🔄 重写", "hint": "换一种表达方式重写，保持主题不变"},
		{"type": string(ActionContinue), "label": "💡 续写", "hint": "根据已有内容续写/扩展"},
		{"type": string(ActionGenerate), "label": "📝 生成", "hint": "根据描述全新创作"},
		{"type": string(ActionCompleteCreate), "label": "🎵 完整创作", "hint": "同时生成风格描述+歌词+歌曲名称"},
	}
}

func BuildChatMessages(actionType ActionType, ctx ChatContext, userInput string, history []ChatMessageDS) []ChatMessageDS {
	pair := BuildPrompt(actionType, ctx, userInput)
	messages := []ChatMessageDS{{Role: "system", Content: pair.System}}
	for _, msg := range history {
		messages = append(messages, msg)
	}
	messages = append(messages, ChatMessageDS{Role: "user", Content: pair.User})
	return messages
}

func MarshalTemplateCategories() ([]byte, error) {
	return json.Marshal(GetTemplateCategories())
}

func MarshalActionTypes() ([]byte, error) {
	return json.Marshal(GetActionTypes())
}
