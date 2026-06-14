import type { VersionDiffItem, VersionDiffResult } from '@/types/diff'
import type { Ingredient, ProcessStep } from '@/types/recipe'

interface VersionSnapshot {
  ingredients: Ingredient[]
  process_steps: ProcessStep[]
}

const AMOUNT_EPSILON = 1e-6

export function formatDiffText(text: string) {
  return text
}

export function compareVersions(base: VersionSnapshot, target: VersionSnapshot): VersionDiffResult {
  return {
    ingredient_diff: diffIngredients(base.ingredients, target.ingredients),
    process_diff: diffProcessSteps(base.process_steps, target.process_steps),
  }
}

interface ApiIngredientDiff {
  type: string
  base?: Ingredient
  target?: Ingredient
}

interface ApiProcessDiff {
  type: string
  order: number
  base?: ProcessStep
  target?: ProcessStep
}

interface ApiDiffResult {
  ingredient_diff?: VersionDiffItem[]
  process_diff?: VersionDiffItem[]
  ingredient_diffs?: ApiIngredientDiff[]
  process_diffs?: ApiProcessDiff[]
}

function mapDiffType(type: string): VersionDiffItem['type'] {
  if (type === 'modified') return 'changed'
  if (type === 'added' || type === 'removed' || type === 'changed' || type === 'unchanged') {
    return type
  }
  return 'unchanged'
}

export function normalizeApiDiff(diff: unknown): VersionDiffResult {
  if (!diff || typeof diff !== 'object') {
    return { ingredient_diff: [], process_diff: [] }
  }

  const apiDiff = diff as ApiDiffResult
  if (Array.isArray(apiDiff.ingredient_diff) && Array.isArray(apiDiff.process_diff)) {
    return {
      ingredient_diff: apiDiff.ingredient_diff,
      process_diff: apiDiff.process_diff,
    }
  }

  return {
    ingredient_diff: (apiDiff.ingredient_diffs ?? []).map((item) => ({
      type: mapDiffType(item.type),
      name: item.target?.name ?? item.base?.name ?? '',
      base: item.base ? formatIngredient(item.base) : undefined,
      target: item.target ? formatIngredient(item.target) : undefined,
    })),
    process_diff: (apiDiff.process_diffs ?? []).map((item) => ({
      type: mapDiffType(item.type),
      name: stepLabel(item.target ?? item.base ?? { order: item.order, content: '' }),
      base: item.base ? formatProcessStep(item.base) : undefined,
      target: item.target ? formatProcessStep(item.target) : undefined,
    })),
  }
}

function diffIngredients(base: Ingredient[], target: Ingredient[]): VersionDiffItem[] {
  const baseMap = indexIngredients(base)
  const seen = new Set<string>()
  const result: VersionDiffItem[] = []

  for (const item of target) {
    const key = ingredientKey(item.name)
    seen.add(key)
    const baseItem = baseMap.get(key)
    if (!baseItem) {
      result.push({ type: 'added', name: item.name, target: formatIngredient(item) })
      continue
    }
    result.push(compareIngredientPair(baseItem, item))
  }

  for (const item of base) {
    const key = ingredientKey(item.name)
    if (seen.has(key)) continue
    result.push({ type: 'removed', name: item.name, base: formatIngredient(item) })
  }

  return result
}

function compareIngredientPair(base: Ingredient, target: Ingredient): VersionDiffItem {
  if (ingredientsEqual(base, target)) {
    return {
      type: 'unchanged',
      name: target.name,
      base: formatIngredient(base),
      target: formatIngredient(target),
    }
  }
  return {
    type: 'changed',
    name: target.name,
    base: formatIngredient(base),
    target: formatIngredient(target),
  }
}

function diffProcessSteps(base: ProcessStep[], target: ProcessStep[]): VersionDiffItem[] {
  const baseMap = indexProcessSteps(base)
  const seen = new Set<number>()
  const result: VersionDiffItem[] = []

  for (const item of target) {
    seen.add(item.order)
    const baseItem = baseMap.get(item.order)
    if (!baseItem) {
      result.push({ type: 'added', name: stepLabel(item), target: formatProcessStep(item) })
      continue
    }
    result.push(compareProcessPair(baseItem, item))
  }

  for (const item of base) {
    if (seen.has(item.order)) continue
    result.push({ type: 'removed', name: stepLabel(item), base: formatProcessStep(item) })
  }

  return result
}

function compareProcessPair(base: ProcessStep, target: ProcessStep): VersionDiffItem {
  if (processStepsEqual(base, target)) {
    return {
      type: 'unchanged',
      name: stepLabel(target),
      base: formatProcessStep(base),
      target: formatProcessStep(target),
    }
  }
  return {
    type: 'changed',
    name: stepLabel(target),
    base: formatProcessStep(base),
    target: formatProcessStep(target),
  }
}

function indexIngredients(items: Ingredient[]) {
  const map = new Map<string, Ingredient>()
  for (const item of items) {
    const key = ingredientKey(item.name)
    if (!map.has(key)) map.set(key, item)
  }
  return map
}

function indexProcessSteps(items: ProcessStep[]) {
  const map = new Map<number, ProcessStep>()
  for (const item of items) {
    if (!map.has(item.order)) map.set(item.order, item)
  }
  return map
}

function ingredientKey(name: string) {
  return name.trim().toLowerCase()
}

function normalizeUnit(unit: string) {
  const u = unit.trim().toLowerCase()
  switch (u) {
    case 'g':
    case '克':
      return 'g'
    case 'kg':
    case '千克':
    case '公斤':
      return 'kg'
    case 'ml':
    case '毫升':
      return 'ml'
    case 'l':
    case '升':
      return 'l'
    default:
      return unit.trim()
  }
}

function ingredientsEqual(a: Ingredient, b: Ingredient) {
  return (
    ingredientKey(a.name) === ingredientKey(b.name) &&
    Math.abs(a.amount - b.amount) < AMOUNT_EPSILON &&
    normalizeUnit(a.unit) === normalizeUnit(b.unit) &&
    (a.note ?? '').trim() === (b.note ?? '').trim()
  )
}

function processStepsEqual(a: ProcessStep, b: ProcessStep) {
  return (
    a.order === b.order &&
    a.content.trim() === b.content.trim() &&
    (a.duration_minutes ?? null) === (b.duration_minutes ?? null)
  )
}

function formatIngredient(item: Ingredient) {
  const note = item.note?.trim() ? `（${item.note.trim()}）` : ''
  return `${item.amount}${item.unit}${note}`
}

function formatProcessStep(item: ProcessStep) {
  const duration = item.duration_minutes ? ` · ${item.duration_minutes}分钟` : ''
  return item.content.trim() + duration
}

function stepLabel(item: ProcessStep) {
  return `步骤 ${item.order}`
}
