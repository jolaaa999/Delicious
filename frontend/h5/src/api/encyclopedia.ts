import client from './client'
import type { EncyclopediaListItem, PageInfo } from '@/types/recipe'

export interface SearchParams {
  keyword?: string
  page?: number
  page_size?: number
  category?: string
}

export interface SearchResult {
  items: EncyclopediaListItem[]
  page_info: PageInfo
}

export function searchEncyclopedia(params: SearchParams) {
  return client.get<unknown, SearchResult>('/encyclopedia/search', { params })
}

export function getEncyclopedia(id: number) {
  return client.get(`/encyclopedia/${id}`)
}
