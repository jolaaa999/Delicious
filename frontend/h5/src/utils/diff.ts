import type { Ingredient, ProcessStep } from '@/types/recipe'
import type { DiffType, IngredientDiff, ProcessStepDiff, VersionDiffResult } from '@/types/diff'

const AMOUNT_EPSILON = 1e-6

export interface VersionSnapshot {
  ingredients: Ingredient[]
  process_steps: ProcessStep[]
}

/** 客户端 Diff（与 backend/pkg/diff 算法一致，API 不可用时本地计算） */
export function compareVersions(base: VersionSnapshot, target: VersionSnapshot): VersionDiffResult {
  const ingredientDiffs = diffIngredients(base.ingredients, target.ingredients)
  const processDiffs = diffProcessSteps(base.process_steps, target.process_steps)
  return {
    ingredient_diffs: ingredientDiffs,
    process_diffs: processDiffs,
    summary: buildSummary(ingredientDiffs, processDiffs),
  }
}

function diffIngredients(base: Ingredient[], target: Ingredient[]): IngredientDiff[] {
  const baseMap = indexIngredients(base)
  const seen = new Set<string>()
  const result: IngredientDiff[] = []

  for (const t of target) {
    const key = ingredientKey(t.name)
    seen.add(key)
    const b = baseMap.get(key)
    if (!b) {
      result.push({ type: 'added', target: { ...t } })
    } else {
      result.push(compareIngredientPair(b, t))
    }
  }

  for (const b of base) {
    const key = ingredientKey(b.name)
    if (!seen.has(key)) {
      result.push({ type: 'removed', base: { ...b } })
    }
  }

  return result
}

function compareIngredientPair(base: Ingredient, target: Ingredient): IngredientDiff {
  if (ingredientsEqual(base, target)) {
    return { type: 'unchanged', base: { ...base }, target: { ...target } }
  }
  return {
    type: 'modified',
    base: { ...base },
    target: { ...target },
    amount_delta: target.amount - base.amount,
  }
}

function diffProcessSteps(base: ProcessStep[], target: ProcessStep[]): ProcessStepDiff[] {
  const baseMap = indexProcessSteps(base)
  const seen = new Set<number>()
  const result: ProcessStepDiff[] = []

  for (const t of target) {
    seen.add(t.order)
    const b = baseMap.get(t.order)
    if (!b) {
      result.push({ type: 'added', order: t.order, target: { ...t } })
    } else {
      result.push(compareProcessPair(b, t))
    }
  }

  for (const b of base) {
    if (!seen.has(b.order)) {
      result.push({ type: 'removed', order: b.order, base: { ...b } })
    }
  }

  return result
}

function compareProcessPair(base: ProcessStep, target: ProcessStep): ProcessStepDiff {
  if (processStepsEqual(base, target)) {
    return { type: 'unchanged', order: base.order, base: { ...base }, target: { ...target } }
  }
  return { type: 'modified', order: base.order, base: { ...base }, target: { ...target } }
}

function indexIngredients(items: Ingredient[]): Map<string, Ingredient> {
  const m = new Map<string, Ingredient>()
  for (const item of items) {
    const key = ingredientKey(item.name)
    if (!m.has(key)) m.set(key, item)
  }
  return m
}

function indexProcessSteps(items: ProcessStep[]): Map<number, ProcessStep> {
  const m = new Map<number, ProcessStep>()
  for (const item of items) {
    if (!m.has(item.order)) m.set(item.order, item)
  }
  return m
}

function ingredientKey(name: string): string {
  return name.trim().toLowerCase()
}

function normalizeUnit(unit: string): string {
  const u = unit.trim()
  switch (u.toLowerCase()) {
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
      return u
  }
}

function ingredientsEqual(a: Ingredient, b: Ingredient): boolean {
  return (
    ingredientKey(a.name) === ingredientKey(b.name) &&
    floatEqual(a.amount, b.amount) &&
    normalizeUnit(a.unit) === normalizeUnit(b.unit) &&
    (a.note ?? '').trim() === (b.note ?? '').trim()
  )
}

function processStepsEqual(a: ProcessStep, b: ProcessStep): boolean {
  return (
    a.order === b.order &&
    a.content.trim() === b.content.trim() &&
    (a.duration_minutes ?? null) === (b.duration_minutes ?? null) &&
    (a.image_url ?? null) === (b.image_url ?? null)
  )
}

function floatEqual(a: number, b: number): boolean {
  return Math.abs(a - b) < AMOUNT_EPSILON
}

function buildSummary(ingDiffs: IngredientDiff[], procDiffs: ProcessStepDiff[]): string {
  let addedIng = 0,
    removedIng = 0,
    modifiedIng = 0
  let addedProc = 0,
    removedProc = 0,
    modifiedProc = 0

  for (const d of ingDiffs) {
    if (d.type === 'added') addedIng++
    else if (d.type === 'removed') removedIng++
    else if (d.type === 'modified') modifiedIng++
  }
  for (const d of procDiffs) {
    if (d.type === 'added') addedProc++
    else if (d.type === 'removed') removedProc++
    else if (d.type === 'modified') modifiedProc++
  }

  if (addedIng + removedIng + modifiedIng + addedProc + removedProc + modifiedProc === 0) {
    return '两个版本完全一致'
  }

  const parts: string[] = []
  if (addedIng) parts.push(`新增 ${addedIng} 项配料`)
  if (removedIng) parts.push(`删除 ${removedIng} 项配料`)
  if (modifiedIng) parts.push(`修改 ${modifiedIng} 项配料`)
  if (addedProc) parts.push(`新增 ${addedProc} 个步骤`)
  if (removedProc) parts.push(`删除 ${removedProc} 个步骤`)
  if (modifiedProc) parts.push(`修改 ${modifiedProc} 个步骤`)
  return parts.join('，')
}

export function diffTypeLabel(type: DiffType): string {
  switch (type) {
    case 'added':
      return '新增'
    case 'removed':
      return '删除'
    case 'modified':
      return '修改'
    default:
      return ''
  }
}

export function formatIngredient(ing?: Ingredient): string {
  if (!ing) return '—'
  const note = ing.note ? `（${ing.note}）` : ''
  return `${ing.name}  ${ing.amount}${ing.unit}${note}`
}

export function formatDelta(delta?: number): string {
  if (delta == null || delta === 0) return ''
  return delta > 0 ? `+${delta}` : `${delta}`
}
