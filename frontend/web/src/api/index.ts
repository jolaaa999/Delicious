import client from './client'
import type {
  DashboardStats,
  MyRecipe,
  PageInfo,
  RecipeListItem,
  TimelineNode,
  VersionDiffResult,
} from '@/types'

export function getDashboardStats() {
  return client.get<unknown, DashboardStats>('/dashboard/stats')
}

export function listRecipes(params?: Record<string, unknown>) {
  return client.get<unknown, { items: RecipeListItem[]; page_info: PageInfo }>('/recipes', { params })
}

export function getRecipe(id: number) {
  return client.get<unknown, { recipe: MyRecipe }>(`/recipes/${id}`)
}

export function getRecipeTimeline(id: number) {
  return client.get<unknown, { recipe: MyRecipe; timeline: TimelineNode[] }>(
    `/dashboard/recipes/${id}/timeline`,
  )
}

export function compareVersions(recipeId: number, baseId: number, targetId: number) {
  return client.get<unknown, { diff: VersionDiffResult }>(`/recipes/${recipeId}/diff`, {
    params: { base_version_id: baseId, target_version_id: targetId },
  })
}

export function deleteRecipe(id: number) {
  return client.delete(`/recipes/${id}`)
}
