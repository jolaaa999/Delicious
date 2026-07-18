import client from './client'
import type { MyRecipe, RecipeListItem, RecipeVersion } from '@/types/recipe'
import type { VersionDiffResult, VersionListItem } from '@/types/diff'
import type { PageInfo, Ingredient, ProcessStep, RecipeVersion as RecipeVersionFull } from '@/types'

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

export interface EncyclopediaCompareResult {
  encyclopedia_recipe_id: number
  encyclopedia_name: string
  encyclopedia_ingredients: Ingredient[]
  encyclopedia_process_steps: ProcessStep[]
  my_version: RecipeVersionFull
  diff: VersionDiffResult
}

export function compareWithEncyclopedia(recipeId: number) {
  return client.get<unknown, EncyclopediaCompareResult>(
    `/recipes/${recipeId}/diff/encyclopedia`,
  )
}

export function deleteRecipe(id: number) {
  return client.delete(`/recipes/${id}`)
}

// ── 回收站 ──
export function listTrashRecipes(params?: Record<string, unknown>) {
  return client.get<unknown, { items: RecipeListItem[]; page_info: PageInfo }>('/recipes/trash', { params })
}

export function restoreRecipe(id: number) {
  return client.post(`/recipes/${id}/restore`)
}

export function permanentDeleteRecipe(id: number) {
  return client.delete(`/recipes/${id}/permanent`)
}

// ── 导出/导入 ──
export interface ExportRecipe {
  name: string
  user_rating?: number
  encyclopedia_recipe_id?: number
  ingredients: { name: string; amount: number; unit: string; note?: string }[]
  process_steps: { order: number; content: string; duration_minutes?: number }[]
  process_text?: string
  images?: string[]
  commit_msg?: string
}

export interface ImportResult {
  total: number
  created: number
  updated: number
  skipped: number
  errors?: string[]
}

export function exportRecipes() {
  return client.get<unknown, { recipes: ExportRecipe[] }>('/recipes/export')
}

export function importRecipes(recipes: ExportRecipe[]) {
  return client.post<unknown, { result: ImportResult }>('/recipes/import', { recipes })
}
