export interface Ingredient {
  name: string
  amount: number
  unit: string
  note?: string
}

export interface ProcessStep {
  order: number
  content: string
  duration_minutes?: number
  image_url?: string
}

export interface RecipeVersion {
  id: number
  recipe_id: number
  version_number: number
  ingredients: Ingredient[]
  process_steps: ProcessStep[]
  process_text?: string
  images?: string[]
  commit_msg: string
  created_at: string
}

export interface MyRecipe {
  id: number
  user_id: string
  recipe_name: string
  current_version_id: number
  user_rating: number
  cover_image_url?: string
  encyclopedia_recipe_id?: number
  created_at: string
  updated_at: string
  current_version?: RecipeVersion
}

export interface RecipeListItem {
  id: number
  recipe_name: string
  cover_image_url?: string
  user_rating: number
  current_version_number: number
  created_at: string
  updated_at: string
}

export interface PageInfo {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface TimelineNode {
  version_id: number
  version_number: number
  commit_msg: string
  created_at: string
  is_current: boolean
}

export interface EncyclopediaListItem {
  id: number
  encyclopedia_recipes_name: string
  description?: string
  category?: string
  cover_image_url?: string
}

export interface DashboardStats {
  total_recipes: number
  average_rating: number
  total_versions: number
  rating_distribution: { rating: number; count: number }[]
  latest_recipe_at?: string
}

export interface VersionDiffResult {
  ingredient_diffs: Array<{
    type: string
    base?: Ingredient
    target?: Ingredient
    amount_delta?: number
  }>
  process_diffs: Array<{
    type: string
    order: number
    base?: ProcessStep
    target?: ProcessStep
  }>
  summary: string
}
