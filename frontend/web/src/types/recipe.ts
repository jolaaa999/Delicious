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
}

export interface EncyclopediaListItem {
  id: number
  name: string
  description?: string
  category?: string
  cover_image_url?: string
}

export interface RecipeVersion {
  id: number
  recipe_id: number
  version_number: number
  commit_msg?: string
  created_at: string
  images?: string[]
  ingredients: Ingredient[]
  process_steps: ProcessStep[]
  process_text?: string
}

export interface MyRecipe {
  id: number
  user_id: number
  name: string
  current_version_id: number
  current_version: RecipeVersion
  user_rating?: number
  cover_image_url?: string
  created_at: string
  updated_at: string
  encyclopedia_recipe_id?: number
}

export interface RecipeListItem {
  id: number
  name: string
  user_rating?: number
  cover_image_url?: string
  current_version_number: number
  updated_at: string
}
