import axios from 'axios'

export interface UploadResult {
  url: string
  filename: string
  size: number
}

/**
 * 上传使用独立 client：Vercel Serverless 代理不支持 multipart，
 * 生产环境请设置 VITE_API_ORIGIN 直连后端（后端已开启 CORS）。
 */
const uploadClient = axios.create({
  baseURL: getUploadBaseURL(),
  timeout: 60000,
})

function getUploadBaseURL(): string {
  const origin = import.meta.env.VITE_API_ORIGIN as string | undefined
  if (origin) {
    return `${origin.replace(/\/$/, '')}/api/v1`
  }
  return (import.meta.env.VITE_API_BASE_URL as string) || '/api/v1'
}

uploadClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('delicious_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

uploadClient.interceptors.response.use(
  (res) => res.data,
  (err) => Promise.reject(err.response?.data ?? err),
)

export async function uploadImage(file: File): Promise<UploadResult> {
  const form = new FormData()
  form.append('file', file)
  return uploadClient.post<unknown, UploadResult>('/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function uploadImages(files: File[]): Promise<UploadResult[]> {
  const form = new FormData()
  files.forEach((f) => form.append('files', f))
  const res = await uploadClient.post<unknown, { files: UploadResult[] }>('/upload/batch', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.files
}

/** 将相对路径转为可访问的完整 URL */
export function resolveImageUrl(url: string): string {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('data:')) {
    return url
  }
  const origin = import.meta.env.VITE_API_ORIGIN as string | undefined
  if (origin) {
    return `${origin.replace(/\/$/, '')}${url}`
  }
  return url
}
