// ============================================================
// LyricCanvas — 配置 API
// ============================================================
import api from './client'

export interface ConfigInfo {
  hasApiKey: boolean
  apiKeyPreview: string
  port: string
  baseUrl: string
}

export async function getConfig(): Promise<ConfigInfo> {
  const res = await api.get<ConfigInfo>('/config')
  return res.data
}

export async function setApiKey(apiKey: string): Promise<void> {
  await api.put('/config/apikey', { apiKey })
}
