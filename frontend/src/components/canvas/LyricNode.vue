<script setup lang="ts">
import { ref, computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { useCanvasStore } from '@/stores/canvas'
import type { LyricNode } from '@/types'

const props = defineProps<{
  id: string
  data: LyricNode
}>()

const canvasStore = useCanvasStore()

const isEditing = ref(false)
const editText = ref(props.data.text)
const isHovered = ref(false)

const isSelected = computed(() => canvasStore.selectedNodeIds.includes(props.id))
const isConnected = computed(() => !!(props.data as any).isConnected)

const nodeStyle = computed(() => {
  const connected = isConnected.value
  if (isEditing.value) {
    return {
      background: 'rgba(255,255,255,0.06)',
      borderColor: 'var(--color-accent)',
      borderWidth: '2px',
      borderStyle: 'solid',
      borderRadius: '12px',
      padding: '12px 16px',
      minWidth: '220px',
      maxWidth: '360px',
      cursor: 'text',
      userSelect: 'text' as const,
      opacity: 1,
      boxShadow: '0 0 0 3px var(--color-accent-soft), 0 4px 20px rgba(0,0,0,0.4)',
      backdropFilter: 'blur(12px)',
      WebkitBackdropFilter: 'blur(12px)',
    }
  }
  if (isSelected.value) {
    return {
      background: 'linear-gradient(135deg, rgba(59,130,246,0.18) 0%, rgba(59,130,246,0.06) 100%)',
      borderColor: 'rgba(59,130,246,0.7)',
      borderWidth: '2px',
      borderStyle: 'solid',
      borderRadius: '12px',
      padding: '12px 16px',
      minWidth: '140px',
      maxWidth: '300px',
      cursor: props.data.data.locked ? 'default' : 'grab',
      userSelect: 'none' as const,
      opacity: 1,
      boxShadow: '0 0 0 2px rgba(59,130,246,0.2), 0 4px 16px rgba(0,0,0,0.35)',
      backdropFilter: 'blur(12px)',
      WebkitBackdropFilter: 'blur(12px)',
    }
  }
  if (connected) {
    return {
      background: 'linear-gradient(135deg, rgba(255,255,255,0.06) 0%, rgba(255,255,255,0.02) 100%)',
      borderColor: 'rgba(255,255,255,0.18)',
      borderWidth: '1.5px',
      borderStyle: 'solid',
      borderRadius: '12px',
      padding: '12px 16px',
      minWidth: '140px',
      maxWidth: '300px',
      cursor: props.data.data.locked ? 'default' : 'grab',
      userSelect: 'none' as const,
      opacity: 1,
      boxShadow: '0 2px 12px rgba(0,0,0,0.25), 0 1px 3px rgba(0,0,0,0.2)',
      backdropFilter: 'blur(12px)',
      WebkitBackdropFilter: 'blur(12px)',
    }
  }
  // 未连接：置灰
  return {
    background: 'rgba(255,255,255,0.01)',
    borderColor: 'rgba(255,255,255,0.04)',
    borderWidth: '1px',
    borderStyle: 'dashed',
    borderRadius: '12px',
    padding: '12px 16px',
    minWidth: '140px',
    maxWidth: '300px',
    cursor: props.data.data.locked ? 'default' : 'grab',
    userSelect: 'none' as const,
    opacity: 0.65,
    boxShadow: '0 1px 4px rgba(0,0,0,0.15)',
    backdropFilter: 'none',
    WebkitBackdropFilter: 'none',
  }
})

function startEdit() {
  if (props.data.data.locked) return
  isEditing.value = true
  editText.value = props.data.text
  // 延迟添加全局点击监听，避免当前双击事件触发退出
  setTimeout(() => {
    document.addEventListener('mousedown', onClickOutside, true)
  }, 50)
}

function onClickOutside(e: MouseEvent) {
  const el = document.querySelector(`[data-node-id="${props.id}"]`)
  if (el && !el.contains(e.target as Node)) {
    finishEdit()
    document.removeEventListener('mousedown', onClickOutside, true)
  }
}

function finishEdit() {
  if (editText.value.trim()) {
    canvasStore.updateNode(props.id, { text: editText.value.trim() })
  }
  isEditing.value = false
  document.removeEventListener('mousedown', onClickOutside, true)
}

function cancelEdit() {
  editText.value = props.data.text
  isEditing.value = false
  document.removeEventListener('mousedown', onClickOutside, true)
}

function handleDelete() {
  canvasStore.removeNode(props.id)
}

function handleDuplicate() {
  canvasStore.duplicateNode(props.id)
}

function handleToggleLock() {
  canvasStore.updateNode(props.id, {
    data: { ...props.data.data, locked: !props.data.data.locked },
  } as any)
}
</script>

<template>
  <div
    class="lyric-node"
    :class="{ 'selected-node': isSelected, 'unconnected': !isConnected }"
    :style="nodeStyle"
    :data-node-id="id"
    @mouseenter="isHovered = true"
    @mouseleave="isHovered = false"
    @dblclick.stop="startEdit"
    @contextmenu.prevent
  >
    <!-- 编辑模式 -->
    <div v-if="isEditing" class="edit-mode" @mousedown.stop @pointerdown.stop @click.stop>
      <textarea
        v-model="editText"
        class="edit-input"
        @keyup.enter="finishEdit"
        @keyup.esc="cancelEdit"
        @blur="finishEdit"
        autofocus
        rows="2"
      />
    </div>

    <!-- 显示模式 -->
    <div v-else class="display-mode">
      <div class="node-text">{{ data.text }}</div>
      <div v-if="isConnected" class="connected-dot" title="已连接" />
      <div v-if="data.data.groupId" class="group-indicator" :title="'已分组'">
        📌
      </div>
      <div v-if="data.data.locked" class="lock-indicator">🔒</div>
    </div>

    <!-- 连接点（无实际连线） -->
    <Handle type="source" :position="Position.Bottom" :style="{ visibility: 'hidden' }" />
    <Handle type="target" :position="Position.Top" :style="{ visibility: 'hidden' }" />

    <!-- 右键菜单 -->
    <div v-if="isHovered && !isEditing" class="node-actions">
      <button class="action-btn" @click.stop="handleDuplicate" title="复制">📋</button>
      <button class="action-btn" @click.stop="handleToggleLock" title="锁定/解锁">
        {{ data.data.locked ? '🔓' : '🔒' }}
      </button>
      <button class="action-btn danger" @click.stop="handleDelete" title="删除">🗑️</button>
    </div>
  </div>
</template>

<style scoped>
.lyric-node {
  position: relative;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  font-size: 15px;
  line-height: 1.7;
  color: var(--color-text);
}

.lyric-node:hover {
  border-color: var(--color-accent) !important;
  box-shadow: 0 4px 24px rgba(59,130,246,0.2), 0 1px 3px rgba(0,0,0,0.3) !important;
}

.lyric-node:hover:not(.unconnected) {
  transform: translateY(-1px);
}

.lyric-node.selected-node {
  border-color: var(--color-accent) !important;
  box-shadow: 0 0 0 3px rgba(59,130,246,0.25), 0 4px 20px rgba(0,0,0,0.4) !important;
}

.lyric-node.unconnected {
  border-style: dashed !important;
  opacity: 0.65;
}

.lyric-node.unconnected .node-text {
  color: rgba(255,255,255,0.55);
}

.display-mode {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.node-text {
  flex: 1;
  word-break: break-word;
  font-weight: 450;
  letter-spacing: 0.01em;
}

.group-indicator,
.lock-indicator {
  flex-shrink: 0;
  font-size: 11px;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.connected-dot {
  flex-shrink: 0;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #34D399;
  margin-top: 5px;
  box-shadow: 0 0 6px rgba(52,211,153,0.5);
}

.lyric-node:hover .group-indicator,
.lyric-node:hover .lock-indicator {
  opacity: 0.8;
}

.edit-mode {
  width: 100%;
  min-width: 200px;
}

.edit-input {
  width: 100%;
  background: rgba(255,255,255,0.05);
  border: none;
  border-radius: 8px;
  color: var(--color-text);
  padding: 8px 10px;
  font-size: 15px;
  line-height: 1.7;
  resize: none;
  outline: none;
  font-family: inherit;
}

.node-actions {
  position: absolute;
  top: -32px;
  right: 0;
  display: flex;
  gap: 3px;
  background: #18181B;
  border-radius: 8px;
  padding: 3px;
  border: 1px solid rgba(255,255,255,0.12);
  box-shadow: 0 4px 16px rgba(0,0,0,0.5);
  opacity: 0;
  transform: translateY(4px);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.lyric-node:hover .node-actions {
  opacity: 1;
  transform: translateY(0);
}

.action-btn {
  background: transparent;
  border: none;
  color: var(--color-text);
  cursor: pointer;
  padding: 4px 8px;
  font-size: 13px;
  border-radius: 6px;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 28px;
}

.action-btn:hover {
  background: rgba(255,255,255,0.12);
}

.action-btn.danger:hover {
  background: rgba(239,68,68,0.2);
  color: #EF4444;
}
</style>
