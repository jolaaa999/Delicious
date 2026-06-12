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
  user_id: number
  name: string
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
  name: string
  cover_image_url?: string
  user_rating: number
  current_version_number: number
  created_at: string
  updated_at: string
}

export interface EncyclopediaListItem {
  id: number
  name: string
  cover_image_url?: string
  category?: string
  tags?: string[]
  description?: string
}

export interface PageInfo {
  page: number
  page_size: number
  total: number
  total_pages: number
}
