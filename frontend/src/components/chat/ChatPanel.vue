<script setup lang="ts">
import { ref, nextTick, watch, onMounted } from 'vue'
import { useChatStore } from '@/stores/chat'
import { useProjectStore } from '@/stores/project'
import { useCanvasStore } from '@/stores/canvas'
import ChatMessageComp from './ChatMessage.vue'
import ChatInputComp from './ChatInput.vue'
import TemplatePanel from './TemplatePanel.vue'
import Modal from '@/components/layout/Modal.vue'

const chatStore = useChatStore()
const projectStore = useProjectStore()
const canvasStore = useCanvasStore()

const chatBody = ref<HTMLElement | null>(null)
const chatInputRef = ref<InstanceType<typeof ChatInputComp> | null>(null)
const showTemplatePanel = ref(false)

// Modal 状态
const modal = ref<{ show: boolean; type: 'info' | 'danger' | 'input'; title: string; message?: string; resolve?: (v?: any) => void }>({
  show: false,
  type: 'info',
  title: '',
})

function showModal(type: 'danger' | 'info', title: string, message?: string): Promise<any> {
  return new Promise((resolve) => {
    modal.value = { show: true, type, title, message, resolve }
  })
}

function onModalConfirm(value?: string) {
  modal.value.show = false
  modal.value.resolve?.(true)
}

function onModalCancel() {
  modal.value.show = false
  modal.value.resolve?.(false)
}

// 选中节点文本
function selectedLyricsText(): string {
  const ids = canvasStore.selectedNodeIds
  if (ids.length === 0) return ''
  const nodes = canvasStore.nodes.filter((n) => ids.includes(n.id))
  return nodes.map((n) => n.text).join('\n')
}

// 自动滚到底部
watch(
  () => chatStore.messages?.length ?? 0,
  async () => {
    await nextTick()
    if (chatBody.value) {
      chatBody.value.scrollTop = chatBody.value.scrollHeight
    }
  }
)

// 流式内容滚动
watch(
  () => chatStore.streamContent,
  async () => {
    await nextTick()
    if (chatBody.value) {
      chatBody.value.scrollTop = chatBody.value.scrollHeight
    }
  }
)

async function handleClear() {
  const ok = await showModal('danger', '清空对话', '确定要清空当前对话历史吗？')
  if (!ok) return
  await chatStore.clearHistory()
}

// 将 AI 回复中的歌词拆分为画布节点
function handleAddToCanvas(content: string) {
  const sentences = content
    .split(/[。！？，；、\n]/)
    .map((s) => s.trim())
    .filter((s) => s.length > 0 && s.length < 100)
    .map((s) => {
      const lastChar = s[s.length - 1]
      return /[。！？，；、]/.test(lastChar) ? s : s + '。'
    })

  if (sentences.length === 0) {
    alert('未检测到有效歌词句子')
    return
  }

  const startX = 300 + canvasStore.nodes.length * 20
  let y = 150

  for (const sentence of sentences) {
    canvasStore.addNode(sentence, { x: startX, y })
    y += 70
  }
}

function toggleTemplatePanel() {
  showTemplatePanel.value = !showTemplatePanel.value
  if (showTemplatePanel.value) {
    chatStore.loadTemplates()
  }
}

function onTemplateSelect(content: string) {
  chatInputRef.value?.fillInput(content)
  showTemplatePanel.value = false
}

onMounted(() => {
  chatStore.loadSystemPrompt()
  chatStore.loadActionTypes()
})
</script>

<template>
  <aside class="panel chat-panel" :style="{ width: 'var(--chat-width)', minWidth: 'var(--chat-width)' }">
    <div class="panel-header">
      <span>🤖 创作助手</span>
      <div class="header-actions">
        <select v-model="chatStore.model" class="model-select">
          <option value="deepseek-chat">V3</option>
          <option value="deepseek-reasoner">R1</option>
          <option value="deepseek-v4-flash">v4-flash</option>
          <option value="deepseek-v4-pro">v4-pro</option>
        </select>
        <button class="btn-ghost" @click="toggleTemplatePanel" title="快捷模板">📋</button>
        <button class="btn-ghost" @click="handleClear" title="清空历史">🗑️</button>
      </div>
    </div>

    <!-- 选中节点信息 -->
    <div v-if="canvasStore.selectedNodeIds.length > 0" class="selection-bar">
      已选 {{ canvasStore.selectedNodeIds.length }} 句
      <span class="selection-hint">— 快捷操作将引用选中内容</span>
    </div>

    <!-- 错误提示 -->
    <div v-if="chatStore.lastError" class="error-bar">
      ❌ {{ chatStore.lastError }}
      <button class="btn-ghost" @click="chatStore.lastError = null">✕</button>
    </div>

    <!-- 模板面板 -->
    <TemplatePanel v-if="showTemplatePanel" @select="onTemplateSelect" />

    <!-- 对话消息 -->
    <div ref="chatBody" class="panel-body chat-body">
      <div v-if="!projectStore.current && !chatStore.messages?.length" class="chat-empty warn">
        <p>⚠️ 尚未选择项目，请先在左侧创建或选择一个项目</p>
      </div>
      <div v-else-if="chatStore.messages?.length === 0 && !chatStore.isStreaming" class="chat-empty">
        <p>👋 输入创作想法开始对话</p>
      </div>

      <ChatMessageComp
        v-for="msg in chatStore.messages"
        :key="msg.id"
        :message="msg"
        @add-to-canvas="handleAddToCanvas"
      />

      <!-- 流式输出中 -->
      <div v-if="chatStore.isStreaming" class="streaming-msg">
        <div class="msg-role">🤖 AI</div>
        <div class="msg-content streaming">{{ chatStore.streamContent }}<span class="cursor">▌</span></div>
      </div>
    </div>

    <!-- 输入区 -->
    <ChatInputComp ref="chatInputRef" />
  </aside>

  <Modal
    :show="modal.show"
    :type="modal.type"
    :title="modal.title"
    :message="modal.message"
    @confirm="onModalConfirm"
    @cancel="onModalCancel"
  />
</template>

<style scoped>
.chat-panel {
  border-left: 1px solid var(--color-border);
  background: linear-gradient(180deg, rgba(18,18,26,0.95) 0%, rgba(9,9,11,0.98) 100%);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  display: flex;
  flex-direction: column;
}

.chat-body {
  flex: 1;
  overflow-y: auto;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.model-select {
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text);
  font-size: 11px;
  font-family: inherit;
  font-weight: 500;
  padding: 5px 10px;
  outline: none;
  cursor: pointer;
  transition: all 0.2s;
}

.model-select:hover {
  border-color: var(--color-border-strong);
}

.model-select:focus {
  border-color: var(--color-accent);
}

.model-select option {
  background: var(--color-bg-elevated);
  color: var(--color-text);
}

.selection-bar {
  background: var(--color-accent-soft);
  border-bottom: 1px solid var(--color-accent);
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 500;
  color: var(--color-accent);
  flex-shrink: 0;
}

.selection-hint {
  color: var(--color-text-muted);
  font-weight: 400;
}

.error-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: rgba(220, 38, 38, 0.15);
  border-bottom: 1px solid rgba(220, 38, 38, 0.3);
  font-size: 11px;
  color: #f87171;
  flex-shrink: 0;
}

.chat-empty {
  padding: 24px 16px;
  text-align: center;
  font-size: 13px;
  color: var(--color-text-muted);
}

.chat-empty.warn {
  color: #fbbf24;
}

.streaming-msg {
  padding: 12px 16px;
}

.msg-role {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-accent);
  margin-bottom: 4px;
}

.msg-content {
  font-size: 13px;
  line-height: 1.7;
  color: var(--color-text);
  white-space: pre-wrap;
  word-break: break-word;
}

.msg-content.streaming {
  color: var(--color-text-secondary);
}

.cursor {
  color: var(--color-accent);
  animation: blink 0.8s infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
</style>
