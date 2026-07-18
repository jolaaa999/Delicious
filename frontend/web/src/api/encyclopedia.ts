import client from './client'
import type { EncyclopediaListItem, Ingredient, ProcessStep } from '@/types/recipe'

const SEARCH_TIMEOUT_MS = 60000
const DETAIL_TIMEOUT_MS = 90000

export interface SearchResult {
  items: EncyclopediaListItem[]
  page_info: { page: number; page_size: number; total: number; total_pages: number }
  highlight_terms?: string[]
}

export function searchEncyclopedia(params: Record<string, unknown>) {
  return client.get<unknown, SearchResult>('/encyclopedia/search', {
    params,
    timeout: SEARCH_TIMEOUT_MS,
  })
}

export interface EncyclopediaDetail {
  recipe: {
    id: number
    name: string
    description?: string
    cover_image_url?: string
    category?: string
    tags?: string[]
    source?: string
    view_count: number
    ingredients: Ingredient[]
    process_steps: ProcessStep[]
  }
}

export function getEncyclopedia(id: number, params?: { lang?: string }) {
  return client.get<unknown, EncyclopediaDetail['recipe']>(`/encyclopedia/${id}`, { params, timeout: DETAIL_TIMEOUT_MS })
}

export function listByCategory(category: string, params?: Record<string, unknown>) {
  return client.get<unknown, SearchResult>(`/encyclopedia/category/${encodeURIComponent(category)}`, {
    params,
    timeout: SEARCH_TIMEOUT_MS,
  })
}

// ── 分类管理 ──
export interface CategoryDTO {
  id: number
  name: string
}

export function listCategories() {
  return client.get<unknown, { items: CategoryDTO[] }>('/categories')
}

export function createCategory(name: string) {
  return client.post<unknown, { category: CategoryDTO }>('/categories', { name })
}

export function updateCategory(id: number, name: string) {
  return client.put<unknown, { category: CategoryDTO }>(`/categories/${id}`, { name })
}

export function deleteCategory(id: number) {
  return client.delete(`/categories/${id}`)
}

// ── 标签管理 ──
export interface TagDTO {
  id: number
  name: string
}

export function listTags() {
  return client.get<unknown, { items: TagDTO[] }>('/tags')
}

export function createTag(name: string) {
  return client.post<unknown, { tag: TagDTO }>('/tags', { name })
}

export function deleteTag(id: number) {
  return client.delete(`/tags/${id}`)
}

export function getEncyclopediaTags(recipeId: number) {
  return client.get<unknown, { items: TagDTO[] }>(`/encyclopedia/${recipeId}/tags`)
}

export function addEncyclopediaTag(recipeId: number, tagId: number) {
  return client.post<unknown, unknown>(`/encyclopedia/${recipeId}/tags`, { tag_id: tagId })
}

export function removeEncyclopediaTag(recipeId: number, tagId: number) {
  return client.delete(`/encyclopedia/${recipeId}/tags/${tagId}`)
}
