<script setup lang="ts">
import { ref, computed } from 'vue'
import { useProjectStore } from '@/stores/project'
import { useCanvasStore } from '@/stores/canvas'
import { useChatStore } from '@/stores/chat'
import ProjectCard from './ProjectCard.vue'
import Modal from '@/components/layout/Modal.vue'

const projectStore = useProjectStore()
const canvasStore = useCanvasStore()
const chatStore = useChatStore()

const newName = ref('')
const showLyricsPreview = ref(true)

// 获取已连接的歌词（有 groupId 且组内 ≥2）
const connectedLyrics = computed(() => {
  if (!projectStore.current) return []
  // 统计每个 groupId 的节点数
  const groupCount = new Map<string, number>()
  for (const n of canvasStore.nodes) {
    const gid = n.data?.groupId
    if (!gid) continue
    groupCount.set(gid, (groupCount.get(gid) || 0) + 1)
  }
  // 筛选组内 ≥2 的节点
  return canvasStore.nodes
    .filter((n) => {
      const gid = n.data?.groupId
      return gid && (groupCount.get(gid) || 0) >= 2
    })
    .sort((a, b) => a.position.y - b.position.y)
    .map((n) => n.text)
})

const lyricsPreviewText = computed(() => connectedLyrics.value.join('\n'))

function copyLyricsPreview() {
  if (lyricsPreviewText.value) {
    navigator.clipboard.writeText(lyricsPreviewText.value)
  }
}

// Modal 状态
const modal = ref<{ show: boolean; type: 'info' | 'danger' | 'input'; title: string; message?: string; placeholder?: string; defaultValue?: string; resolve?: (v?: any) => void }>({
  show: false,
  type: 'info',
  title: '',
})

function showModal(type: 'danger' | 'input', title: string, opts?: { message?: string; placeholder?: string; defaultValue?: string }): Promise<any> {
  return new Promise((resolve) => {
    modal.value = { show: true, type, title, ...opts, resolve }
  })
}

function onModalConfirm(value?: string) {
  modal.value.show = false
  modal.value.resolve?.(value ?? true)
}

function onModalCancel() {
  modal.value.show = false
  modal.value.resolve?.(false)
}

async function handleCreate() {
  const name = newName.value.trim() || '未命名歌曲'
  await projectStore.create(name)
  newName.value = ''
  await chatStore.loadHistory()
}

async function handleSelect(id: string) {
  await projectStore.loadProject(id)
  await chatStore.loadHistory()
}

async function handleDelete(id: string) {
  const ok = await showModal('danger', '删除项目', { message: '确定要删除此项目吗？此操作不可撤销。' })
  if (!ok) return
  await projectStore.remove(id)
}

async function handleRename(id: string) {
  const proj = projectStore.list.find((p) => p.id === id)
  if (!proj) return
  const name = await showModal('input', '重命名项目', { defaultValue: proj.name, placeholder: '输入新名称...' })
  if (name && typeof name === 'string' && name.trim()) {
    projectStore.current!.name = name.trim()
    await projectStore.save()
    await projectStore.loadList()
  }
}
</script>

<template>
  <aside class="panel sidebar" :style="{ width: 'var(--sidebar-width)', minWidth: 'var(--sidebar-width)' }">
    <div class="panel-header">
      <span>📁 我的项目</span>
    </div>

    <!-- 新建 -->
    <div class="create-bar">
      <input
        v-model="newName"
        placeholder="歌曲名称..."
        @keyup.enter="handleCreate"
      />
      <button class="btn-primary" @click="handleCreate">+ 新建</button>
    </div>

    <!-- 项目列表 -->
    <div class="panel-body">
      <div v-if="projectStore.loading" class="status-text">加载中...</div>
      <div v-else-if="projectStore.error" class="status-text error">
        ❌ {{ projectStore.error }}
        <br/><small>请确认后端已启动 (cd backend && lyriccanvas.exe)</small>
      </div>
      <div v-else-if="projectStore.list.length === 0" class="status-text">暂无项目，点击上方新建</div>
      <template v-else>
        <ProjectCard
          v-for="proj in projectStore.list"
          :key="proj.id"
          :project="proj"
          :active="projectStore.current?.id === proj.id"
          @select="handleSelect(proj.id)"
          @delete="handleDelete(proj.id)"
          @rename="handleRename(proj.id)"
        />
      </template>
    </div>

    <!-- 歌词预览 -->
    <div v-if="projectStore.current && connectedLyrics.length > 0" class="lyrics-preview">
      <div class="lyrics-preview-header" @click="showLyricsPreview = !showLyricsPreview">
        <span>📜 完整歌词 · {{ connectedLyrics.length }} 句</span>
        <span class="preview-toggle">{{ showLyricsPreview ? '▼' : '▶' }}</span>
      </div>
      <div v-if="showLyricsPreview" class="lyrics-preview-body">
        <div class="lyrics-text">{{ lyricsPreviewText }}</div>
        <button class="btn-ghost copy-btn" @click="copyLyricsPreview">📋 复制</button>
      </div>
    </div>
    <div v-else-if="projectStore.current" class="lyrics-preview empty">
      <div class="lyrics-preview-header">
        <span>📜 完整歌词</span>
      </div>
      <div class="lyrics-preview-body hint">
        选中节点后点击 🔗打组 即可连接
      </div>
    </div>
  </aside>

  <!-- 自定义弹窗 -->
  <Modal
    :show="modal.show"
    :type="modal.type"
    :title="modal.title"
    :message="modal.message"
    :placeholder="modal.placeholder"
    :default-value="modal.defaultValue"
    @confirm="onModalConfirm"
    @cancel="onModalCancel"
  />
</template>

<style scoped>
.sidebar {
  border-right: 1px solid var(--color-border);
  user-select: none;
  background: linear-gradient(180deg, rgba(18,18,26,0.95) 0%, rgba(9,9,11,0.98) 100%);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
}

.create-bar {
  padding: 12px 14px;
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--color-border-light);
}

.create-bar input {
  flex: 1;
  min-width: 0;
  padding: 9px 12px;
  border-radius: var(--radius-sm);
  font-size: 13px;
}

.create-bar button {
  white-space: nowrap;
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 600;
}

.status-text {
  color: var(--color-text-muted);
  font-size: 13px;
  text-align: center;
  padding: 32px 16px;
  line-height: 1.6;
}

.status-text.error {
  color: var(--color-danger);
  background: var(--color-danger-soft);
  border-radius: var(--radius-sm);
  margin: 8px;
  padding: 16px;
}

/* 歌词预览 */
.lyrics-preview {
  border-top: 1px solid var(--color-border-light);
  flex-shrink: 0;
  max-height: 40%;
  display: flex;
  flex-direction: column;
}

.lyrics-preview.empty {
  max-height: none;
}

.lyrics-preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
  cursor: pointer;
  user-select: none;
  letter-spacing: 0.02em;
  transition: color 0.15s;
}

.lyrics-preview-header:hover {
  color: var(--color-text);
}

.preview-toggle {
  font-size: 10px;
  opacity: 0.5;
}

.lyrics-preview-body {
  padding: 0 14px 10px;
  overflow-y: auto;
  flex: 1;
}

.lyrics-text {
  font-size: 13px;
  line-height: 2;
  color: var(--color-text);
  white-space: pre-wrap;
  word-break: break-word;
}

.lyrics-preview-body.hint {
  font-size: 12px;
  color: var(--color-text-muted);
  padding: 8px 14px 14px;
  line-height: 1.6;
}

.copy-btn {
  margin-top: 8px;
  font-size: 11px;
  padding: 4px 10px;
  border-radius: 6px;
}
</style>
