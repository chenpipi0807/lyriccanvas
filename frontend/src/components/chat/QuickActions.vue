<script setup lang="ts">
import { useChatStore } from '@/stores/chat'
import { useCanvasStore } from '@/stores/canvas'
import type { ActionType } from '@/types'
import { ACTION_TYPES } from '@/types'

const chatStore = useChatStore()
const canvasStore = useCanvasStore()

function selectedLyrics(): string {
  const ids = canvasStore.selectedNodeIds
  if (ids.length === 0) return ''
  const nodes = canvasStore.nodes.filter((n) => ids.includes(n.id))
  return nodes.map((n) => n.text).join('\n')
}

function buildPrompt(action: ActionType): string {
  const sel = selectedLyrics()
  switch (action) {
    case 'polish':
      return sel ? `请润色以下选中的歌词：\n${sel}` : '请润色当前项目的歌词，提升文学性和韵律感。'
    case 'rewrite':
      return sel ? `请用不同的表达方式重写以下歌词：\n${sel}` : '请用不同的表达方式重写当前歌词。'
    case 'continue':
      return sel ? `请根据已有内容续写歌词：\n${sel}` : '请根据已有内容续写歌词，保持风格一致。'
    case 'generate':
      return sel ? `请参考以下歌词风格，创作一段新的歌词：\n${sel}` : '请根据项目风格创作一段歌词。'
    case 'complete_create':
      return '请根据当前项目完成完整创作（风格描述 + 歌词 + 歌曲名称）。'
    default:
      return ''
  }
}

function handleAction(action: ActionType) {
  chatStore.actionType = action
  const msg = buildPrompt(action)
  if (msg) {
    chatStore.sendMessage(msg)
  }
}
</script>

<template>
  <div class="quick-actions">
    <button
      v-for="act in ACTION_TYPES"
      :key="act.type"
      class="quick-btn"
      :class="{ active: chatStore.actionType === act.type }"
      :title="act.hint"
      @click="handleAction(act.type)"
    >
      {{ act.label }}
    </button>
  </div>
</template>

<style scoped>
.quick-actions {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 6px;
  margin-top: 12px;
}

.quick-btn {
  font-size: 12px;
  font-weight: 500;
  padding: 7px 10px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  transition: all 0.2s ease;
  color: var(--color-text-secondary);
  background: rgba(255,255,255,0.03);
  cursor: pointer;
  white-space: nowrap;
}

.quick-btn:hover {
  border-color: var(--color-accent);
  color: var(--color-accent);
  background: var(--color-accent-soft);
  transform: translateY(-1px);
}

.quick-btn.active {
  border-color: var(--color-accent);
  color: var(--color-accent);
  background: var(--color-accent-soft);
  box-shadow: 0 0 8px var(--color-accent-soft);
}
</style>
