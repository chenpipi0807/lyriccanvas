// ============================================================
// LyricCanvas — TypeScript 类型定义
// ============================================================

/** 项目 */
export interface Project {
  id: string
  name: string
  createdAt: string
  updatedAt: string
  canvas: Canvas
  chatHistory: ChatMessage[]
}

/** 项目列表项（摘要） */
export interface ProjectListItem {
  id: string
  name: string
  createdAt: string
  updatedAt: string
  nodeCount: number
}

/** 画布 */
export interface Canvas {
  viewport: Viewport
  nodes: LyricNode[]
  groups: GroupLabel[]
  draftZone: DraftZone
}

/** 视口 */
export interface Viewport {
  x: number
  y: number
  zoom: number
}

/** 歌词节点 */
export interface LyricNode {
  id: string
  type: 'lyric'
  text: string
  position: Position
  style?: NodeStyle
  data: LyricNodeData
}

export interface LyricNodeData {
  groupId: string | null
  order: number
  locked: boolean
  rhymeHint?: string
}

export interface NodeStyle {
  color?: string
  fontSize?: number
  width?: number
}

/** 结构标签 */
export interface GroupLabel {
  id: string
  type: 'structure'
  label: string
  category: StructureCategory
  position: Position
  nodeIds: string[]
  color: string
}

export type StructureCategory =
  | 'intro'
  | 'verse'
  | 'pre-chorus'
  | 'chorus'
  | 'bridge'
  | 'outro'
  | 'interlude'

/** 位置 */
export interface Position {
  x: number
  y: number
}

/** 草稿区 */
export interface DraftZone {
  x: number
  y: number
  width: number
  height: number
}

/** 对话消息 */
export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
}

/** 对话请求 */
export interface ChatRequest {
  projectId: string
  model?: string
  systemPrompt?: string
  message: string
  actionType?: ActionType
  targetFields?: TargetField[]
  temperature?: number
  maxTokens?: number
  stream?: boolean
}

// ========== Prompt 模板相关类型 ==========

/** 操作类型 */
export type ActionType = 'polish' | 'rewrite' | 'continue' | 'generate' | 'complete_create'

/** 目标字段 */
export type TargetField = 'songIdea' | 'lyrics' | 'songName'

/** 操作类型项 */
export interface ActionTypeItem {
  type: ActionType
  label: string
  hint: string
}

/** 快捷模板 */
export interface QuickTemplate {
  id: string
  category: string
  label: string
  content: string
  hint?: string
}

/** 模板分类 */
export interface TemplateCategory {
  key: string
  label: string
  templates: QuickTemplate[]
}

/** 押韵结果 */
export interface RhymeResult {
  char: string
  pinyin: string
  final: string
  rhymes: RhymeChar[]
  commonWords?: string[]
}

export interface RhymeChar {
  char: string
  pinyin: string
}

/** 预设结构类别 */
export const STRUCTURE_CATEGORIES: { value: StructureCategory; label: string; color: string }[] = [
  { value: 'intro', label: '前奏', color: '#9B9B9B' },
  { value: 'verse', label: '主歌', color: '#4A90D9' },
  { value: 'pre-chorus', label: '导歌', color: '#7B68EE' },
  { value: 'chorus', label: '副歌', color: '#E85D75' },
  { value: 'bridge', label: '桥段', color: '#F5A623' },
  { value: 'outro', label: '尾声', color: '#9B9B9B' },
  { value: 'interlude', label: '间奏', color: '#50C878' },
]

/** 可用操作类型列表 */
export const ACTION_TYPES: ActionTypeItem[] = [
  { type: 'polish', label: '✨ 润色', hint: '润色当前内容，提升文学性和韵律感' },
  { type: 'rewrite', label: '🔄 重写', hint: '换一种表达方式重写，保持主题不变' },
  { type: 'continue', label: '💡 续写', hint: '根据已有内容续写/扩展' },
  { type: 'generate', label: '📝 生成', hint: '根据描述全新创作' },
  { type: 'complete_create', label: '🎵 完整创作', hint: '同时生成风格描述+歌词+歌曲名称' },
]
