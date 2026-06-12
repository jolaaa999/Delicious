import type { Ingredient, ProcessStep, RecipeVersion } from './recipe'

export type DiffType = 'unchanged' | 'added' | 'removed' | 'modified'

/** API 可能返回数字枚举 (1-4) 或字符串 */
export type DiffTypeInput = DiffType | 1 | 2 | 3 | 4

export interface IngredientDiff {
  type: DiffTypeInput
  base?: Ingredient
  target?: Ingredient
  amount_delta?: number
}

export interface ProcessStepDiff {
  type: DiffTypeInput
  order: number
  base?: ProcessStep
  target?: ProcessStep
}

export interface VersionDiffResult {
  ingredient_diffs: IngredientDiff[]
  process_diffs: ProcessStepDiff[]
  summary: string
}

export interface VersionListItem {
  id: number
  version_number: number
  commit_msg: string
  created_at: string
}

export interface CompareVersionsResponse {
  base_version: RecipeVersion
  target_version: RecipeVersion
  diff: VersionDiffResult
}

export interface CompareEncyclopediaResponse {
  encyclopedia_recipe_id: number
  encyclopedia_name: string
  encyclopedia_ingredients: Ingredient[]
  encyclopedia_process_steps: ProcessStep[]
  my_version: RecipeVersion
  diff: VersionDiffResult
}

export function normalizeDiffType(type: DiffTypeInput): DiffType {
  if (typeof type === 'string') return type
  switch (type) {
    case 1:
      return 'unchanged'
    case 2:
      return 'added'
    case 3:
      return 'removed'
    case 4:
      return 'modified'
    default:
      return 'unchanged'
  }
}
