import client from './client'

export function uploadImage(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return client.post<unknown, { url: string }>('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function resolveImageUrl(url?: string) {
  if (!url) return ''
  if (/^https?:\/\//i.test(url)) return url

  const normalized = url.startsWith('/') ? url : `/${url}`
  // 本地上传路径由后端静态目录提供，不在 /api/v1 下
  if (normalized.startsWith('/uploads/')) {
    const origin = import.meta.env.VITE_API_ORIGIN || ''
    return origin ? `${origin.replace(/\/$/, '')}${normalized}` : normalized
  }

  const base = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  return `${base.replace(/\/$/, '')}${normalized}`
}
