// ============================================================
// LyricCanvas — 押韵 API
// ============================================================
import api from './client'
import type { RhymeResult } from '@/types'

export async function fetchRhyme(char: string): Promise<RhymeResult> {
  const { data } = await api.get('/rhyme', { params: { char } })
  return data
}

export async function fetchRhymeBatch(chars: string[]): Promise<RhymeResult[]> {
  const { data } = await api.post('/rhyme/batch', { chars })
  return data
}
