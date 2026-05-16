// ============================================================
// LyricCanvas — 项目状态管理
// ============================================================
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Project, ProjectListItem } from '@/types'
import * as projectsApi from '@/api/projects'

export const useProjectStore = defineStore('project', () => {
  const list = ref<ProjectListItem[]>([])
  const current = ref<Project | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function loadList() {
    loading.value = true
    error.value = null
    try {
      list.value = await projectsApi.fetchProjects()
    } catch (e: any) {
      error.value = e?.message || '加载项目列表失败'
    } finally {
      loading.value = false
    }
  }

  async function loadProject(id: string) {
    loading.value = true
    error.value = null
    try {
      current.value = await projectsApi.fetchProject(id)
    } catch (e: any) {
      error.value = e?.message || '加载项目失败'
    } finally {
      loading.value = false
    }
  }

  async function create(name: string) {
    loading.value = true
    error.value = null
    try {
      const proj = await projectsApi.createProject(name)
      list.value.unshift({
        id: proj.id,
        name: proj.name,
        createdAt: proj.createdAt,
        updatedAt: proj.updatedAt,
        nodeCount: 0,
      })
      current.value = proj
      return proj
    } catch (e: any) {
      error.value = e?.message || '创建项目失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function save() {
    if (!current.value) return
    try {
      await projectsApi.updateProject(current.value.id, {
        name: current.value.name,
        canvas: current.value.canvas,
        chatHistory: current.value.chatHistory,
      })
    } catch (e: any) {
      console.error('保存失败:', e)
    }
  }

  async function remove(id: string) {
    error.value = null
    try {
      await projectsApi.deleteProject(id)
      list.value = list.value.filter((p) => p.id !== id)
      if (current.value?.id === id) {
        current.value = null
      }
    } catch (e: any) {
      error.value = e?.message || '删除失败'
    }
  }

  return { list, current, loading, error, loadList, loadProject, create, save, remove }
})
