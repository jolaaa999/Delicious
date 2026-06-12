import axios from 'axios'

/**
 * 本地开发：/api/v1 → Vite proxy → localhost:8080
 * Vercel 部署：/api/v1 → api/v1/[...path].ts → BACKEND_URL
 * 直连后端：设置 VITE_API_BASE_URL=https://your-api.com/api/v1
 */
const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

const client = axios.create({
  baseURL,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('delicious_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (res) => res.data,
  (err) => Promise.reject(err.response?.data ?? err),
)

export default client
