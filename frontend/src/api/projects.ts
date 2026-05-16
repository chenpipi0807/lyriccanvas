// ============================================================
// LyricCanvas — 项目 API
// ============================================================
import api from './client'
import type { Project, ProjectListItem, Canvas, ChatMessage } from '@/types'

export async function fetchProjects(): Promise<ProjectListItem[]> {
  const { data } = await api.get('/projects')
  return data
}

export async function fetchProject(id: string): Promise<Project> {
  const { data } = await api.get(`/projects/${id}`)
  return data
}

export async function createProject(name: string): Promise<Project> {
  const { data } = await api.post('/projects', { name })
  return data
}

export async function updateProject(
  id: string,
  updates: { name?: string; canvas?: Canvas; chatHistory?: ChatMessage[] }
): Promise<Project> {
  const { data } = await api.put(`/projects/${id}`, updates)
  return data
}

export async function deleteProject(id: string): Promise<void> {
  await api.delete(`/projects/${id}`)
}
