<script setup lang="ts">
import { ref, watch } from 'vue'
import { useChatStore } from '@/stores/chat'

const chatStore = useChatStore()

const input = ref('')
const textareaHeight = ref(80)

// 监听外部填充请求
watch(() => chatStore.pendingInput, (val) => {
  if (val) {
    input.value = val
    chatStore.pendingInput = ''
  }
})

function handleSend() {
  const msg = input.value.trim()
  if (!msg) return
  chatStore.sendMessage(msg)
  input.value = ''
}

// 自定义拖拽拉高输入框
const isDragging = ref(false)
let dragStartY = 0
let dragStartH = 0

function onDragStart(e: MouseEvent) {
  isDragging.value = true
  dragStartY = e.clientY
  dragStartH = textareaHeight.value
  document.addEventListener('mousemove', onDragMove)
  document.addEventListener('mouseup', onDragEnd)
}

function onDragMove(e: MouseEvent) {
  if (!isDragging.value) return
  const dy = dragStartY - e.clientY  // 向上为正
  textareaHeight.value = Math.max(60, Math.min(280, dragStartH + dy))
}

function onDragEnd() {
  isDragging.value = false
  document.removeEventListener('mousemove', onDragMove)
  document.removeEventListener('mouseup', onDragEnd)
}

function fillInput(text: string) {
  input.value = text
}

defineExpose({ fillInput })
</script>

<template>
  <div class="chat-input-area">
    <div class="input-row">
      <div class="textarea-wrapper">
        <div class="resize-handle" @mousedown.prevent="onDragStart">
          <div class="resize-grip" />
        </div>
        <textarea
          v-model="input"
          class="chat-textarea"
          :style="{ height: textareaHeight + 'px' }"
          placeholder="输入创作需求或自由提问... (Enter 发送, Shift+Enter 换行)"
          @keydown.enter.exact.prevent="handleSend"
        />
      </div>
      <div class="input-actions">
        <button
          class="btn-send"
          :disabled="!input.trim()"
          @click="handleSend"
        >
          发送
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-input-area {
  flex-shrink: 0;
  border-top: 1px solid var(--color-border-light);
  background: rgba(0,0,0,0.15);
}

.target-fields {
  display: flex;
  gap: 10px;
  padding: 4px 12px;
  font-size: 11px;
}

.field-check {
  display: flex;
  align-items: center;
  gap: 3px;
  color: var(--color-text-muted);
  cursor: pointer;
}

.field-check input[type="checkbox"] {
  accent-color: var(--color-accent);
  cursor: pointer;
}

.template-select {
  margin-left: auto;
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-muted);
  font-size: 11px;
  font-family: inherit;
  padding: 3px 8px;
  outline: none;
  cursor: pointer;
  max-width: 130px;
}

.template-select:hover,
.template-select:focus {
  border-color: var(--color-accent);
  color: var(--color-text);
}

.template-select option,
.template-select optgroup {
  background: #18181B;
  color: var(--color-text);
  font-size: 12px;
}

.input-row {
  display: flex;
  gap: 8px;
  padding: 0 10px 10px;
  align-items: flex-end;
}

.textarea-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.resize-handle {
  height: 14px;
  cursor: ns-resize;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 2px;
  user-select: none;
  flex-shrink: 0;
}

.resize-handle:hover .resize-grip,
.resize-handle:active .resize-grip {
  background: var(--color-accent);
  width: 60px;
}

.resize-grip {
  width: 40px;
  height: 4px;
  border-radius: 2px;
  background: rgba(255,255,255,0.15);
  transition: all 0.15s;
}

.chat-textarea {
  resize: none;
  font-family: inherit;
  font-size: 13px;
  line-height: 1.6;
  padding: 10px 14px;
  border-radius: 10px;
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  transition: border-color 0.2s, box-shadow 0.2s;
  width: 100%;
}

.chat-textarea:focus {
  outline: none;
  border-color: var(--color-accent);
  background: rgba(255,255,255,0.06);
  box-shadow: 0 0 0 3px var(--color-accent-soft);
}

.chat-textarea::placeholder {
  color: var(--color-text-muted);
}

.input-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: stretch;
}

.btn-send {
  flex: 1;
  padding: 6px 16px;
  font-weight: 600;
  font-size: 12px;
  border-radius: 8px;
  border: none;
  background: var(--color-accent);
  color: #fff;
  cursor: pointer;
  transition: all 0.15s;
  min-width: 48px;
}

.btn-send:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-send:disabled {
  opacity: 0.3;
  cursor: not-allowed;
  transform: none;
}

.btn-tmpl {
  padding: 4px 8px;
  font-size: 11px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.btn-tmpl:hover {
  border-color: var(--color-accent);
  color: var(--color-accent);
  background: var(--color-accent-soft);
}
</style>
