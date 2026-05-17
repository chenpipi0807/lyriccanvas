<script setup lang="ts">
import { ref, watch, computed, markRaw, onMounted, onUnmounted } from 'vue'
import { VueFlow, useVueFlow, SelectionMode } from '@vue-flow/core'
import { Background, BackgroundVariant } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { useProjectStore } from '@/stores/project'
import { useCanvasStore } from '@/stores/canvas'
import { useChatStore } from '@/stores/chat'
import LyricNodeComp from './LyricNode.vue'
import Modal from '@/components/layout/Modal.vue'
import type { LyricNode, Position } from '@/types'
import { parseLyrics } from '@/api/chat'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'

const projectStore = useProjectStore()
const canvasStore = useCanvasStore()
const chatStore = useChatStore()
const { screenToFlowCoordinate, getSelectedNodes, onNodesChange } = useVueFlow()

// 监听 Vue Flow 节点变化 → 同步框选状态到 canvasStore
onNodesChange(() => {
  const sel = (getSelectedNodes.value || []) as any[]
  if (sel.length === 0) return  // 不覆盖手动选中
  const ids = sel.map((n: any) => n.id)
  if (ids.join(',') !== canvasStore.selectedNodeIds.join(',')) {
    canvasStore.selectedNodeIds = ids
  }
})

// 定时备份：确保框选状态不漏（仅当 VueFlow 有选中时才同步）
const syncInterval = setInterval(() => {
  const sel = (getSelectedNodes.value || []) as any[]
  if (sel.length === 0) return  // 不覆盖手动选中（Ctrl+A / Ctrl+点击）
  const ids = sel.map((n: any) => n.id)
  if (ids.join(',') !== canvasStore.selectedNodeIds.join(',')) {
    canvasStore.selectedNodeIds = ids
  }
}, 300)

// Modal 状态
const importModal = ref<{ show: boolean; resolve?: (value?: string) => void }>({ show: false })

function showImportModal(): Promise<string | undefined> {
  return new Promise((resolve) => {
    importModal.value = { show: true, resolve }
  })
}

function onImportConfirm(value?: string) {
  importModal.value.show = false
  importModal.value.resolve?.(value)
}

function onImportCancel() {
  importModal.value.show = false
  importModal.value.resolve?.()
}

const isParsing = ref(false)
async function handleSmartParse() {
  const text = await showImportModal()
  if (!text || !text.trim()) return

  isParsing.value = true
  try {
    const result = await parseLyrics(text.trim())
    const sentences = result.lines.filter((s) => s.length > 1)

    const startX = 300 + canvasStore.nodes.length * 20
    let y = 150

    for (const sentence of sentences) {
      canvasStore.addNode(sentence, { x: startX, y })
      y += 70
    }
  } catch (e: any) {
    // 降级：本地简单拆分
    console.warn('智能解析失败，使用本地拆分', e)
    parseLyricsToNodes(text.trim())
  } finally {
    isParsing.value = false
  }
}

// 后端连接状态
const backendStatus = ref<'checking' | 'ok' | 'down'>('checking')
async function checkBackend() {
  try {
    const ctrl = new AbortController()
    const timer = setTimeout(() => ctrl.abort(), 5000)
    const res = await fetch('/api/health', { signal: ctrl.signal })
    clearTimeout(timer)
    backendStatus.value = res.ok ? 'ok' : 'down'
  } catch {
    backendStatus.value = 'down'
  }
}
checkBackend()
setInterval(checkBackend, 10000)

// 将 store 数据转为 Vue Flow 格式
const connectedGroupIds = computed(() => {
  const count = new Map<string, number>()
  for (const n of canvasStore.nodes) {
    const gid = n.data?.groupId
    if (!gid) continue
    count.set(gid, (count.get(gid) || 0) + 1)
  }
  return new Set([...count.entries()].filter(([_, c]) => c >= 2).map(([gid]) => gid))
})

const flowNodes = computed(() => {
  return canvasStore.nodes.map((n) => ({
    id: n.id,
    type: 'lyric',
    position: n.position,
    data: {
      ...n,
      isConnected: !!n.data?.groupId && connectedGroupIds.value.has(n.data.groupId),
    },
  }))
})

const flowEdges = computed(() => {
  const edges: any[] = []
  const groupMap = new Map<string, { id: string; y: number }[]>()
  for (const n of canvasStore.nodes) {
    const gid = n.data?.groupId
    if (!gid) continue
    if (!groupMap.has(gid)) groupMap.set(gid, [])
    groupMap.get(gid)!.push({ id: n.id, y: n.position.y })
  }
  for (const [, members] of groupMap) {
    members.sort((a, b) => a.y - b.y)
    for (let i = 0; i < members.length - 1; i++) {
      edges.push({
        id: `edge_${members[i].id}_${members[i + 1].id}`,
        source: members[i].id,
        target: members[i + 1].id,
        type: 'default',
        style: { stroke: 'rgba(59,130,246,0.22)', strokeWidth: 0.8 },
      })
    }
  }
  return edges
})

const dragStartPos = ref<Position | null>(null)
const dragStartSelected = ref<string[]>([])

function onNodeDragStart(event: any) {
  dragStartPos.value = { ...event.node.position }
  dragStartSelected.value = [...canvasStore.selectedNodeIds]
}

function onNodeDragStop(event: any) {
  const { node } = event
  if (isAltHeld.value && dragStartPos.value) {
    // Alt+拖拽：复制所有选中节点
    const dx = node.position.x - dragStartPos.value.x
    const dy = node.position.y - dragStartPos.value.y
    canvasStore.updateNodePosition(node.id, dragStartPos.value)

    const ids = dragStartSelected.value.includes(node.id) && dragStartSelected.value.length > 1
      ? dragStartSelected.value
      : [node.id]

    for (const id of ids) {
      const orig = canvasStore.nodes.find((n) => n.id === id)
      if (orig) {
        canvasStore.addNode(orig.text, { x: orig.position.x + dx, y: orig.position.y + dy })
      }
    }
    dragStartPos.value = null
    dragStartSelected.value = []
    return
  }
  dragStartPos.value = null
  dragStartSelected.value = []
  // 多选拖拽时，同步所有选中节点的位置到 store（VueFlow 内部已移动它们，但只回调了拖拽的那个节点）
  const sel = (getSelectedNodes.value || []) as any[]
  if (sel.length > 1) {
    for (const sn of sel) {
      canvasStore.updateNodePosition(sn.id, sn.position)
    }
  } else {
    canvasStore.updateNodePosition(node.id, node.position)
  }
  checkSnap(node)
}

const snapThresholdX = 50
const snapThresholdY = 30
const rowSpacing = 68
const breakGroupDistance = 150  // 拖远超过此距离自动拆组

function checkSnap(movedNode: any) {
  const movedData = canvasStore.nodes.find((n) => n.id === movedNode.id)
  const movedGid = movedData?.data?.groupId

  // 检查是否拖远 → 自动拆组
  if (movedGid) {
    const groupMembers = canvasStore.nodes.filter((n) => n.data?.groupId === movedGid && n.id !== movedNode.id)
    const tooFar = groupMembers.every((m) => {
      const dx = Math.abs(movedNode.position.x - m.position.x)
      const dy = Math.abs(movedNode.position.y - m.position.y)
      return Math.sqrt(dx * dx + dy * dy) > breakGroupDistance
    })
    if (tooFar && groupMembers.length > 0) {
      // 拆组
      canvasStore.updateNode(movedNode.id, { data: { ...movedData!.data, groupId: null } } as any)
      // 如果组只剩 1 个节点，也拆掉它
      if (groupMembers.length === 1) {
        const last = groupMembers[0]
        canvasStore.updateNode(last.id, { data: { ...last.data, groupId: null } } as any)
      }
      return
    }
    // 如果在组内且不拆组，检查吸附到同组其他节点
  }

  const allNodes = canvasStore.nodes
  for (const other of allNodes) {
    if (other.id === movedNode.id) continue
    const dx = Math.abs(movedNode.position.x - other.position.x)
    const dy = Math.abs(movedNode.position.y - other.position.y)
    if (dx < snapThresholdX && dy < snapThresholdY * 4) {
      snapAlign(movedNode, other)
      break
    }
  }
}

function snapAlign(a: any, b: any) {
  const avgX = (a.position.x + b.position.x) / 2
  if (a.position.y <= b.position.y) {
    canvasStore.updateNodePosition(a.id, { x: avgX, y: a.position.y })
    canvasStore.updateNodePosition(b.id, { x: avgX, y: a.position.y + rowSpacing })
  } else {
    canvasStore.updateNodePosition(b.id, { x: avgX, y: b.position.y })
    canvasStore.updateNodePosition(a.id, { x: avgX, y: b.position.y + rowSpacing })
  }
  // 自动打组：吸附后归入同一组
  const nodeA = canvasStore.nodes.find((n) => n.id === a.id)
  const nodeB = canvasStore.nodes.find((n) => n.id === b.id)
  const gidA = nodeA?.data?.groupId
  const gidB = nodeB?.data?.groupId
  if (!gidA && !gidB) {
    // 两个都没组 → 新建组
    const newGid = 'group_' + (crypto.randomUUID ? crypto.randomUUID().slice(0, 8) : Math.random().toString(36).slice(2, 10))
    canvasStore.updateNode(a.id, { data: { ...nodeA!.data, groupId: newGid } } as any)
    canvasStore.updateNode(b.id, { data: { ...nodeB!.data, groupId: newGid } } as any)
  } else if (gidA && !gidB) {
    canvasStore.updateNode(b.id, { data: { ...nodeB!.data, groupId: gidA } } as any)
  } else if (!gidA && gidB) {
    canvasStore.updateNode(a.id, { data: { ...nodeA!.data, groupId: gidB } } as any)
  }
  // 都有组 → 保持各自组
}

function onNodeClick(event: any) {
  const { node } = event
  const evt = event.event as MouseEvent
  const ids = [...canvasStore.selectedNodeIds]
  if (evt.ctrlKey || evt.metaKey) {
    const idx = ids.indexOf(node.id)
    if (idx >= 0) ids.splice(idx, 1)
    else ids.push(node.id)
  } else {
    if (ids.length === 1 && ids[0] === node.id) {
      ids.length = 0
    } else {
      ids.length = 0
      ids.push(node.id)
    }
  }
  canvasStore.selectedNodeIds = ids
}

function onPaneClickHandler(ev: any) {
  canvasStore.selectedNodeIds = []
  onPaneClick(ev)
}

const editingNodeId = ref<string | null>(null)
function onNodeDoubleClick(event: any) {
  editingNodeId.value = event.node.id
}

function finishEditing(id: string, text: string) {
  canvasStore.updateNode(id, { text })
  editingNodeId.value = null
}

const isAltHeld = ref(false)

const nodeTypes = markRaw({
  lyric: LyricNodeComp,
}) as Record<string, any>

let paneClickTimer: ReturnType<typeof setTimeout> | null = null
function onPaneClick(ev: any) {
  if (paneClickTimer) {
    clearTimeout(paneClickTimer)
    paneClickTimer = null
    const mouseEvent = ev.event || ev
    const pos = screenToFlowCoordinate({ x: mouseEvent.clientX, y: mouseEvent.clientY })
    if (!projectStore.current) return
    canvasStore.addNode('对你的喜欢不言而渝', { x: pos.x - 70, y: pos.y - 25 })
  } else {
    paneClickTimer = setTimeout(() => { paneClickTimer = null }, 300)
  }
}

// Ctrl+G 打组 + 自动排版
function groupSelected() {
  const ids = canvasStore.selectedNodeIds
  if (ids.length < 2) return

  // 获取选中节点并按 Y 排序
  const selected = canvasStore.nodes
    .filter((n) => ids.includes(n.id))
    .sort((a, b) => a.position.y - b.position.y)

  // 计算平均 X
  const avgX = selected.reduce((s, n) => s + n.position.x, 0) / selected.length

  const groupId = 'group_' + (crypto.randomUUID ? crypto.randomUUID().slice(0, 8) : Math.random().toString(36).slice(2, 10))

  // 排版：纵向等距排列
  let y = selected[0].position.y
  for (const node of selected) {
    canvasStore.updateNode(node.id, {
      position: { x: avgX, y },
      data: { ...node.data, groupId },
    } as any)
    y += rowSpacing
  }

  canvasStore.selectedNodeIds = []
}

function exportSelected() {
  const ids = canvasStore.selectedNodeIds
  if (ids.length === 0) return
  const nodes = canvasStore.nodes
    .filter((n) => ids.includes(n.id))
    .sort((a, b) => a.position.y - b.position.y)
  const text = nodes.map((n) => n.text).join('\n')
  navigator.clipboard.writeText(text)
}

// 导出完整歌词为 txt 文件
function exportTxt() {
  const groupCount = new Map<string, number>()
  for (const n of canvasStore.nodes) {
    const gid = n.data?.groupId
    if (!gid) continue
    groupCount.set(gid, (groupCount.get(gid) || 0) + 1)
  }
  const lines = canvasStore.nodes
    .filter((n) => {
      const gid = n.data?.groupId
      return gid && (groupCount.get(gid) || 0) >= 2
    })
    .sort((a, b) => a.position.y - b.position.y)
    .map((n) => n.text)

  if (lines.length === 0) return
  const text = lines.join('\n')
  const name = projectStore.current?.name || '歌词'
  const blob = new Blob([text], { type: 'text/plain' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = name + '.txt'
  a.click()
  URL.revokeObjectURL(a.href)
}

function sendSelectedToChat() {
  const ids = canvasStore.selectedNodeIds
  if (ids.length === 0) return
  const nodes = canvasStore.nodes
    .filter((n) => ids.includes(n.id))
    .sort((a, b) => a.position.y - b.position.y)
  const text = nodes.map((n) => n.text).join('\n')
  chatStore.fillInput('请根据以下歌词内容提供建议：\n' + text)
}

const selectionBarStyle = computed(() => {
  const ids = canvasStore.selectedNodeIds
  if (ids.length < 2) return { display: 'none' }
  const selectedNodes = canvasStore.nodes.filter((n) => ids.includes(n.id))
  if (selectedNodes.length === 0) return { display: 'none' }
  const minX = Math.min(...selectedNodes.map((n) => n.position.x))
  const maxY = Math.max(...selectedNodes.map((n) => n.position.y))
  return {
    left: `${minX}px`,
    top: `${maxY + 70}px`,
  }
})

// 本地降级拆分
function parseLyricsToNodes(lyricsText: string) {
  // 清理杂符
  const cleaned = lyricsText
    .replace(/[#\*~`\[\]{}()<>（）《》「」『』〈〉]/g, '')
    .replace(/[0-9]+\s*[.、．]/g, '')
    .replace(/\n{2,}/g, '\n')

  const sentences = cleaned
    .split(/[。！？\n]/)
    .map((s) => s.trim())
    .filter((s) => s.length > 1)
    .map((s) => {
      const lastChar = s[s.length - 1]
      const needsPunct = !/[。！？]/.test(lastChar)
      return s + (needsPunct ? '。' : '')
    })

  const startX = 300 + canvasStore.nodes.length * 20
  let y = 150

  for (const sentence of sentences) {
    canvasStore.addNode(sentence, { x: startX, y })
    y += 70
  }
}

defineExpose({ parseLyricsToNodes })

watch(
  () => canvasStore.selectedNodeIds,
  (ids) => {
    if (ids.length === 0) return
  }
)

// 键盘快捷键 → window 级别监听（Vue Flow 会吞掉 div 上的 keydown）
function handleGlobalKeydown(e: KeyboardEvent) {
  // 实时同步 Alt 状态（比 keyup 可靠）
  isAltHeld.value = e.altKey

  const focused = document.activeElement
  if (focused && (focused.tagName === 'INPUT' || focused.tagName === 'TEXTAREA' || focused.tagName === 'SELECT')) return

  if (e.key === 'Backspace' || e.key === 'Delete') {
    const ids = [...canvasStore.selectedNodeIds]
    if (ids.length === 0) return
    e.preventDefault()
    for (const id of ids) canvasStore.removeNode(id)
    return
  }

  if (e.ctrlKey && (e.key === 'z' || e.code === 'KeyZ')) {
    e.preventDefault()
    canvasStore.undo()
    return
  }

  if (e.ctrlKey && (e.key === 'g' || e.code === 'KeyG')) {
    e.preventDefault()
    if (canvasStore.selectedNodeIds.length >= 2) groupSelected()
    return
  }

  if (e.ctrlKey && (e.key === 'a' || e.code === 'KeyA')) {
    e.preventDefault()
    canvasStore.selectedNodeIds = canvasStore.nodes.map((n) => n.id)
    return
  }
}

function handleGlobalKeyup(e: KeyboardEvent) {
  isAltHeld.value = e.altKey
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeydown)
  window.addEventListener('keyup', handleGlobalKeyup)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
  window.removeEventListener('keyup', handleGlobalKeyup)
})
</script>

<template>
  <div class="canvas-wrapper" tabindex="0">
    <!-- 顶部工具栏 -->
    <div class="canvas-toolbar">
      <span class="toolbar-title" v-if="projectStore.current">
        🎵 {{ projectStore.current.name }}
        <span class="backend-dot" :class="backendStatus" :title="'后端: ' + backendStatus" />
      </span>
      <span class="toolbar-title" v-else>
        未选择项目
        <span class="backend-dot" :class="backendStatus" :title="'后端: ' + backendStatus" />
      </span>
      <div class="toolbar-actions">
        <button
          class="btn-primary"
          @click="handleSmartParse"
          :disabled="isParsing"
          :title="isParsing ? 'AI 解析中...' : '粘贴歌词文本，AI 智能拆句'"
        >
          {{ isParsing ? '⏳ 解析中...' : '🤖 智能解析' }}
        </button>
        <button class="btn-ghost" @click="exportTxt" title="导出完整歌词为 txt 文件">📥 导出 TXT</button>
      </div>
    </div>

    <!-- Vue Flow 画布 -->
    <VueFlow
      v-if="projectStore.current"
      :nodes="flowNodes"
      :edges="flowEdges"
      :node-types="nodeTypes"
      :default-viewport="{ x: 0, y: 0, zoom: 1 }"
      :min-zoom="0.1"
      :max-zoom="4"
      :snap-to-grid="false"
      :selection-key-code="'Control'"
      :selection-mode="SelectionMode.Partial"
      :nodes-draggable="true"
      :nodes-connectable="false"
      :edges-updatable="false"
      @node-click="onNodeClick"
      @node-drag-start="onNodeDragStart"
      @node-drag-stop="onNodeDragStop"
      @node-double-click="onNodeDoubleClick"
      @pane-click="onPaneClickHandler"
      class="vue-flow-instance"
    >
      <Background :variant="BackgroundVariant.Dots" :gap="20" :size="1" />
      <Controls />
    </VueFlow>

    <!-- 多选浮动操作栏 -->
    <div
      v-if="canvasStore.selectedNodeIds.length >= 2"
      class="selection-float-bar"
      :style="selectionBarStyle"
    >
      <span class="sel-count">{{ canvasStore.selectedNodeIds.length }} 句已选</span>
      <button class="sel-action" @click="groupSelected">🔗 打组</button>
      <button class="sel-action" @click="exportSelected">📋 复制</button>
      <button class="sel-action" @click="sendSelectedToChat">💬 加入对话</button>
    </div>

    <!-- 空白提示 -->
    <div v-else-if="!projectStore.current" class="empty-hint">
      <template v-if="backendStatus === 'down'">
        <p>🔴 后端未连接</p>
        <p class="hint-sub">请先启动后端: cd backend && .\lyriccanvas.exe</p>
      </template>
      <template v-else-if="projectStore.error">
        <p>❌ {{ projectStore.error }}</p>
      </template>
      <template v-else>
        <p>👈 在左侧新建或选择一个项目开始创作</p>
      </template>
    </div>

  </div>

  <Modal
    :show="importModal.show"
    type="input"
    title="智能解析歌词"
    message="粘贴歌词文本（支持带格式、杂符），AI 将自动拆分为干净的句子："
    placeholder="粘贴歌词内容..."
    @confirm="onImportConfirm"
    @cancel="onImportCancel"
  />
</template>

<style scoped>
.canvas-wrapper {
  width: 100%;
  height: 100%;
  position: relative;
  outline: none;
}

.canvas-toolbar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 20px;
  background: rgba(18, 18, 26, 0.8);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--color-border);
  pointer-events: auto;
}

.toolbar-title {
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--color-text);
}

.backend-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 0 8px currentColor;
}

.backend-dot.checking { background: #FBBF24; box-shadow: 0 0 8px #FBBF24; }
.backend-dot.ok { background: #34D399; box-shadow: 0 0 8px #34D399; }
.backend-dot.down { background: #FB7185; box-shadow: 0 0 8px #FB7185; }

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.vue-flow-instance {
  width: 100%;
  height: 100%;
}

.empty-hint {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: var(--color-text-secondary);
  font-size: 16px;
  text-align: center;
  pointer-events: none;
}

.hint-sub {
  font-size: 12px;
  opacity: 0.6;
  margin-top: 10px;
  font-family: 'SF Mono', 'Consolas', 'Fira Code', monospace;
}

/* 多选浮动操作栏 */
.selection-float-bar {
  position: absolute;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: #18181B;
  border: 1px solid rgba(255,255,255,0.10);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.5);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  pointer-events: auto;
  transform: translateX(-50%);
}

.sel-count {
  font-size: 11px;
  color: var(--color-text-muted);
  font-weight: 500;
  padding-right: 6px;
  border-right: 1px solid rgba(255,255,255,0.08);
}

.sel-action {
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  font-size: 12px;
  font-family: inherit;
  font-weight: 500;
  padding: 5px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.sel-action:hover {
  background: var(--color-accent-soft);
  color: var(--color-accent);
}
</style>
