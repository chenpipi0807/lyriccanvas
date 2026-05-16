// ============================================================
// LyricCanvas — 对话 API（含 SSE 流式）
// ============================================================
import api from './client'
import type { ChatMessage, ChatRequest, ActionTypeItem, TemplateCategory } from '@/types'

/** 非流式对话 */
export async function sendChat(req: ChatRequest): Promise<string> {
  const { data } = await api.post('/chat', { ...req, stream: false })
  return data.reply
}

/** 流式对话 — 通过 ReadableStream 逐 token 返回 */
export async function sendChatStream(
  req: ChatRequest,
  onToken: (token: string) => void,
  onDone: () => void,
  onError: (err: Error) => void
): Promise<void> {
  try {
    const body = JSON.stringify({ ...req, stream: true })

    const response = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body,
    })

    if (!response.ok) {
      const errText = await response.text()
      onError(new Error(`请求失败: ${response.status} ${errText}`))
      return
    }

    const reader = response.body?.getReader()
    if (!reader) {
      onError(new Error('不支持流式响应'))
      return
    }

    const decoder = new TextDecoder()
    let buffer = ''

    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line === 'data: [DONE]') {
            onDone()
            return
          }
          if (!line.startsWith('data: ')) continue

          const jsonStr = line.slice(6)
          try {
            const parsed = JSON.parse(jsonStr)
            if (parsed.content) {
              onToken(parsed.content)
            }
            if (parsed.error) {
              onError(new Error(parsed.error))
              return
            }
          } catch {
            // skip malformed chunks
          }
        }
      }
      onDone()
    } catch (err) {
      onError(err instanceof Error ? err : new Error('流式读取失败'))
    }
  } catch (err: any) {
    onError(err instanceof Error ? err : new Error(err?.message || '请求失败，请检查后端是否启动'))
  }
}

/** 获取对话历史 */
export async function fetchChatHistory(projectId: string): Promise<ChatMessage[]> {
  const { data } = await api.get(`/chat/history/${projectId}`)
  return data
}

/** 清空对话历史 */
export async function clearChatHistory(projectId: string): Promise<void> {
  await api.delete(`/chat/history/${projectId}`)
}

/** 获取系统提示词 */
export async function fetchSystemPrompt(): Promise<string> {
  const { data } = await api.get('/chat/system-prompt')
  return data.systemPrompt
}

/** 获取快捷模板分类 */
export async function fetchTemplates(): Promise<TemplateCategory[]> {
  const { data } = await api.get('/chat/templates')
  return data.categories ?? []
}

/** 获取可用操作类型 */
export async function fetchActions(): Promise<ActionTypeItem[]> {
  const { data } = await api.get('/chat/actions')
  return data.actions ?? []
}

/** 智能解析歌词 —— 调 DeepSeek 拆分句子、去除杂符 */
export async function parseLyrics(text: string): Promise<{ lines: string[]; fallback?: boolean }> {
  const { data } = await api.post('/lyrics/parse', { text })
  return data
}
