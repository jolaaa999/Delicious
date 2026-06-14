import client from './client'
import type { MyRecipe, RecipeListItem, RecipeVersion } from '@/types/recipe'
import type { VersionDiffResult, VersionListItem } from '@/types/diff'

interface PageInfo {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export function createRecipe(payload: Record<string, unknown>) {
  return client.post<unknown, { recipe: MyRecipe }>('/recipes', payload)
}

export function updateRecipe(id: number, payload: Record<string, unknown>) {
  return client.put<unknown, { recipe: MyRecipe }>(`/recipes/${id}`, payload)
}

export function getRecipe(id: number) {
  return client.get<unknown, { recipe: MyRecipe }>(`/recipes/${id}`)
}

export function listRecipes(params?: Record<string, unknown>) {
  return client.get<unknown, { items: RecipeListItem[]; page_info: PageInfo }>('/recipes', { params })
}

export function listVersions(recipeId: number) {
  return client.get<unknown, { versions: VersionListItem[] }>(`/recipes/${recipeId}/versions`)
}

export function getVersion(recipeId: number, versionId: number) {
  return client.get<unknown, { version: RecipeVersion }>(`/recipes/${recipeId}/versions/${versionId}`)
}

export function compareVersions(recipeId: number, baseId: number, targetId: number) {
  return client.get<unknown, { diff: VersionDiffResult }>(`/recipes/${recipeId}/diff`, {
    params: { base_version_id: baseId, target_version_id: targetId },
  })
}

export function compareWithEncyclopedia(recipeId: number) {
  return client.get<unknown, { diff: VersionDiffResult; encyclopedia_name: string }>(
    `/recipes/${recipeId}/diff/encyclopedia`,
  )
}
