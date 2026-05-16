// ============================================================
// LyricCanvas — 对话状态管理 (增强：action type / templates)
// ============================================================
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ChatMessage, ChatRequest, ActionType, TargetField, ActionTypeItem, TemplateCategory } from '@/types'
import * as chatApi from '@/api/chat'
import { useProjectStore } from './project'

export const useChatStore = defineStore('chat', () => {
  const projectStore = useProjectStore()

  const messages = computed({
    get: () => projectStore.current?.chatHistory ?? [],
    set: (v: ChatMessage[]) => {
      if (projectStore.current) {
        projectStore.current.chatHistory = v
      }
    },
  })

  const isStreaming = ref(false)
  const streamContent = ref('')
  const systemPrompt = ref('')
  const model = ref('deepseek-v4-flash')
  const lastError = ref<string | null>(null)

  // 新增状态
  const actionType = ref<ActionType>('generate')
  const targetFields = ref<TargetField[]>(['lyrics'])
  const actionTypes = ref<ActionTypeItem[]>([])
  const templateCategories = ref<TemplateCategory[]>([])
  const templatesLoaded = ref(false)

  // 外部填充输入框（不发送）
  const pendingInput = ref('')

  function fillInput(text: string) {
    pendingInput.value = text
  }

  async function loadSystemPrompt() {
    try {
      systemPrompt.value = await chatApi.fetchSystemPrompt()
    } catch {
      systemPrompt.value = '你是一位资深音乐创作助手，专注于协助用户创作中文歌词。'
    }
  }

  async function loadActionTypes() {
    try {
      actionTypes.value = await chatApi.fetchActions()
    } catch {
      // 使用前端默认值
    }
  }

  async function loadTemplates() {
    if (templatesLoaded.value) return
    try {
      templateCategories.value = await chatApi.fetchTemplates()
      templatesLoaded.value = true
    } catch {
      // 静默失败
    }
  }

  async function sendMessage(userMessage: string, selectedLyrics = '') {
    if (!projectStore.current) return

    const projectId = projectStore.current.id

    const req: ChatRequest = {
      projectId,
      model: model.value,
      systemPrompt: systemPrompt.value,
      message: userMessage,
      actionType: actionType.value,
      targetFields: targetFields.value,
      temperature: 1.0,
      maxTokens: 4096,
      stream: true,
    }

    // 添加用户消息到历史
    const userMsg: ChatMessage = {
      id: 'msg_' + crypto.randomUUID().slice(0, 8),
      role: 'user',
      content: userMessage,
      timestamp: new Date().toISOString(),
    }
    messages.value = [...messages.value, userMsg]

    // 流式获取回复
    lastError.value = null
    isStreaming.value = true
    streamContent.value = ''

    await chatApi.sendChatStream(
      req,
      (token) => {
        streamContent.value += token
      },
      () => {
        const aiMsg: ChatMessage = {
          id: 'msg_' + crypto.randomUUID().slice(0, 8),
          role: 'assistant',
          content: streamContent.value,
          timestamp: new Date().toISOString(),
        }
        messages.value = [...messages.value, aiMsg]
        isStreaming.value = false
        streamContent.value = ''
      },
      (err) => {
        isStreaming.value = false
        lastError.value = err.message || '对话请求失败'
        console.error('流式对话错误:', err)
      }
    )
  }

  async function clearHistory() {
    if (!projectStore.current) return
    await chatApi.clearChatHistory(projectStore.current.id)
    messages.value = []
  }

  async function loadHistory() {
    if (!projectStore.current) return
    const history = await chatApi.fetchChatHistory(projectStore.current.id)
    messages.value = history
  }

  return {
    messages,
    isStreaming,
    streamContent,
    systemPrompt,
    model,
    lastError,
    actionType,
    targetFields,
    actionTypes,
    templateCategories,
    templatesLoaded,
    pendingInput,
    fillInput,
    loadSystemPrompt,
    loadActionTypes,
    loadTemplates,
    sendMessage,
    clearHistory,
    loadHistory,
  }
})
