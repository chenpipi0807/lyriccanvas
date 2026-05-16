// ============================================================
// LyricCanvas — 画布状态管理
// ============================================================
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { LyricNode, GroupLabel, DraftZone, Position } from '@/types'
import { useProjectStore } from './project'

export const useCanvasStore = defineStore('canvas', () => {
  const projectStore = useProjectStore()

  // 防抖自动保存
  let saveTimer: ReturnType<typeof setTimeout> | null = null
  function scheduleAutoSave() {
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => {
      projectStore.save()
    }, 800)
  }

  // ---- 撤销栈 ----
  interface CanvasSnapshot {
    nodes: LyricNode[]
    groups: GroupLabel[]
  }
  const undoStack = ref<CanvasSnapshot[]>([])
  const maxUndo = 50

  function snapshot(): CanvasSnapshot {
    return {
      nodes: JSON.parse(JSON.stringify(nodes.value)),
      groups: JSON.parse(JSON.stringify(groups.value)),
    }
  }

  function pushUndo() {
    undoStack.value.push(snapshot())
    if (undoStack.value.length > maxUndo) {
      undoStack.value.shift()
    }
  }

  function undo() {
    const last = undoStack.value.pop()
    if (!last) return
    nodes.value = last.nodes
    groups.value = last.groups
    scheduleAutoSave()
  }

  // 画布数据快捷访问
  const nodes = computed({
    get: () => projectStore.current?.canvas.nodes ?? [],
    set: (v: LyricNode[]) => {
      if (projectStore.current) {
        projectStore.current.canvas.nodes = v
      }
    },
  })

  const groups = computed({
    get: () => projectStore.current?.canvas.groups ?? [],
    set: (v: GroupLabel[]) => {
      if (projectStore.current) {
        projectStore.current.canvas.groups = v
      }
    },
  })

  const draftZone = computed({
    get: () => projectStore.current?.canvas.draftZone ?? { x: 100, y: 100, width: 600, height: 800 },
    set: (v: DraftZone) => {
      if (projectStore.current) {
        projectStore.current.canvas.draftZone = v
      }
    },
  })

  // 选中节点
  const selectedNodeIds = ref<string[]>([])

  function addNode(text: string, position?: Position): LyricNode {
    pushUndo()
    const node: LyricNode = {
      id: 'node_' + crypto.randomUUID().slice(0, 8),
      type: 'lyric',
      text,
      position: position ?? { x: 300, y: 200 },
      style: { color: '#333', fontSize: 16, width: 200 },
      data: { groupId: null, order: 0, locked: false },
    }
    nodes.value = [...nodes.value, node]
    scheduleAutoSave()
    return node
  }

  function removeNode(id: string) {
    pushUndo()
    nodes.value = nodes.value.filter((n) => n.id !== id)
    selectedNodeIds.value = selectedNodeIds.value.filter((sid) => sid !== id)
    scheduleAutoSave()
  }

  function updateNode(id: string, updates: Partial<LyricNode>) {
    pushUndo()
    nodes.value = nodes.value.map((n) => (n.id === id ? { ...n, ...updates } : n))
    scheduleAutoSave()
  }

  function updateNodePosition(id: string, position: Position) {
    updateNode(id, { position })
  }

  function duplicateNode(id: string, offset?: Position): LyricNode | null {
    const original = nodes.value.find((n) => n.id === id)
    if (!original) return null
    const copy = {
      ...JSON.parse(JSON.stringify(original)),
      id: 'node_' + crypto.randomUUID().slice(0, 8),
      position: offset
        ? { x: original.position.x + offset.x, y: original.position.y + offset.y }
        : { x: original.position.x + 30, y: original.position.y + 30 },
    }
    nodes.value = [...nodes.value, copy]
    scheduleAutoSave()
    return copy
  }

  function addGroup(category: string, label: string, position: Position): GroupLabel {
    const group: GroupLabel = {
      id: 'group_' + crypto.randomUUID().slice(0, 8),
      type: 'structure',
      label,
      category: category as any,
      position,
      nodeIds: [],
      color: '#3B82F6',
    }
    groups.value = [...groups.value, group]
    scheduleAutoSave()
    return group
  }

  function assignNodeToGroup(nodeId: string, groupId: string | null) {
    updateNode(nodeId, { data: { ...nodes.value.find((n) => n.id === nodeId)!.data, groupId } } as any)
    // 更新 group 的 nodeIds
    groups.value = groups.value.map((g) => {
      if (g.id === groupId && !g.nodeIds.includes(nodeId)) {
        return { ...g, nodeIds: [...g.nodeIds, nodeId] }
      }
      if (g.id !== groupId && g.nodeIds.includes(nodeId)) {
        return { ...g, nodeIds: g.nodeIds.filter((id) => id !== nodeId) }
      }
      return g
    })
  }

  // 是否在草稿区内
  function isInDraftZone(pos: Position): boolean {
    const dz = draftZone.value
    return pos.x >= dz.x && pos.x <= dz.x + dz.width && pos.y >= dz.y && pos.y <= dz.y + dz.height
  }

  return {
    nodes,
    groups,
    draftZone,
    selectedNodeIds,
    undoStack,
    pushUndo,
    undo,
    addNode,
    removeNode,
    updateNode,
    updateNodePosition,
    duplicateNode,
    addGroup,
    assignNodeToGroup,
    isInDraftZone,
  }
})
