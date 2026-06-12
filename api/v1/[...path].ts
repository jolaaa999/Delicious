import type { VercelRequest, VercelResponse } from '@vercel/node'

/**
 * Vercel Serverless 代理（仓库根目录部署时使用）
 * 环境变量 BACKEND_URL = Go API 网关地址，如 https://api.example.com
 */
export default async function handler(req: VercelRequest, res: VercelResponse) {
  const backend = process.env.BACKEND_URL
  if (!backend) {
    res.status(503).json({
      message: 'BACKEND_URL 未配置。请在 Vercel → Settings → Environment Variables 中设置。',
    })
    return
  }

  const segments = req.query.path
  const pathStr = Array.isArray(segments) ? segments.join('/') : segments || ''
  const url = new URL(req.url || '/', 'http://localhost')
  const target = `${backend.replace(/\/$/, '')}/api/v1/${pathStr}${url.search}`

  const headers: Record<string, string> = {
    Accept: (req.headers.accept as string) || 'application/json',
  }
  if (req.headers.authorization) {
    headers.Authorization = req.headers.authorization as string
  }
  if (req.headers['content-type']) {
    headers['Content-Type'] = req.headers['content-type'] as string
  }

  const init: RequestInit = { method: req.method, headers }

  if (req.method && !['GET', 'HEAD'].includes(req.method) && req.body) {
    init.body = typeof req.body === 'string' ? req.body : JSON.stringify(req.body)
  }

  try {
    const upstream = await fetch(target, init)
    const contentType = upstream.headers.get('content-type')
    if (contentType) {
      res.setHeader('Content-Type', contentType)
    }
    res.status(upstream.status).send(await upstream.text())
  } catch (err) {
    res.status(502).json({
      message: '无法连接后端服务',
      error: err instanceof Error ? err.message : String(err),
    })
  }
}
