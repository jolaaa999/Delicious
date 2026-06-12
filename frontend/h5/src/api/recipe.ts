import client from './client'
import type { Ingredient, MyRecipe, ProcessStep, RecipeListItem, PageInfo } from '@/types/recipe'

export interface ListParams {
  page?: number
  page_size?: number
  min_rating?: number
  max_rating?: number
  keyword?: string
  order_by?: string
  desc?: boolean
}

export interface ListResult {
  items: RecipeListItem[]
  page_info: PageInfo
}

export interface CreateRecipePayload {
  name: string
  ingredients: Ingredient[]
  process_steps: ProcessStep[]
  process_text?: string
  images?: string[]
  user_rating?: number
  commit_msg?: string
  encyclopedia_recipe_id?: number
}

export function listRecipes(params?: ListParams) {
  return client.get<unknown, ListResult>('/recipes', { params })
}

export function getRecipe(id: number) {
  return client.get<unknown, { recipe: MyRecipe }>(`/recipes/${id}`)
}

export function createRecipe(data: CreateRecipePayload) {
  return client.post<unknown, { recipe: MyRecipe }>('/recipes', data)
}

export function updateRecipe(id: number, data: CreateRecipePayload & { commit_msg: string }) {
  return client.put<unknown, { recipe: MyRecipe }>(`/recipes/${id}`, data)
}

export function deleteRecipe(id: number) {
  return client.delete(`/recipes/${id}`)
}

export function listVersions(recipeId: number) {
  return client.get(`/recipes/${recipeId}/versions`)
}

export function compareVersions(
  recipeId: number,
  baseVersionId: number,
  targetVersionId: number,
) {
  return client.get(`/recipes/${recipeId}/diff`, {
    params: { base_version_id: baseVersionId, target_version_id: targetVersionId },
  })
}

export function compareWithEncyclopedia(recipeId: number, encyclopediaRecipeId?: number) {
  return client.get(`/recipes/${recipeId}/diff/encyclopedia`, {
    params: encyclopediaRecipeId ? { encyclopedia_recipe_id: encyclopediaRecipeId } : {},
  })
}
