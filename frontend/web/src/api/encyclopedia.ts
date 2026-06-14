import client from './client'
import type { EncyclopediaListItem, Ingredient, ProcessStep } from '@/types/recipe'

export function searchEncyclopedia(params: Record<string, unknown>) {
  return client.get<unknown, { items: EncyclopediaListItem[] }>('/encyclopedia/search', { params })
}

export function getEncyclopedia(id: number) {
  return client.get<unknown, {
    recipe: {
      id: number
      name: string
      description?: string
      category?: string
      tags?: string[]
      ingredients: Ingredient[]
      process_steps: ProcessStep[]
    }
  }>(`/encyclopedia/${id}`)
}
