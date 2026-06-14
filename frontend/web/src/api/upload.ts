import client from './client'

export function uploadImage(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return client.post<unknown, { url: string }>('/upload/image', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function resolveImageUrl(url?: string) {
  if (!url) return ''
  if (/^https?:\/\//i.test(url)) return url
  const base = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  return `${base.replace(/\/$/, '')}/${url.replace(/^\//, '')}`
}
