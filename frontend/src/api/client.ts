// ============================================================
// LyricCanvas — HTTP 客户端
// ============================================================
import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 120000,
  headers: { 'Content-Type': 'application/json' },
})

export default api
