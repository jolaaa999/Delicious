import { ref, shallowRef } from 'vue'
import type { VersionDiffResult, VersionListItem } from '@/types/diff'
import type { MyRecipe, RecipeVersion } from '@/types/recipe'
import {
  compareVersions as apiCompareVersions,
  compareWithEncyclopedia as apiCompareEncyclopedia,
  getVersion,
  listVersions,
} from '@/api/recipe'
import { compareVersions as localCompare, normalizeApiDiff } from '@/utils/diff'

export type DiffMode = 'version' | 'encyclopedia' | null

export function useRecipeDiff(recipeId: number) {
  const versions = ref<VersionListItem[]>([])
  const versionsLoading = ref(false)
  const diffLoading = ref(false)
  const diffMode = ref<DiffMode>(null)
  const diffResult = shallowRef<VersionDiffResult | null>(null)
  const baseLabel = ref('')
  const targetLabel = ref('')
  const selectedVersion = ref<VersionListItem | null>(null)
  const drawerOpen = ref(false)

  async function loadVersions(currentVersionId?: number) {
    versionsLoading.value = true
    try {
      const res = await listVersions(recipeId)
      versions.value = res.versions ?? []
    } catch {
      versions.value = getDemoVersions(currentVersionId)
    } finally {
      versionsLoading.value = false
    }
  }

  function openHistoryDrawer() {
    drawerOpen.value = true
  }

  function closeHistoryDrawer() {
    drawerOpen.value = false
  }

  async function compareWithVersion(version: VersionListItem, currentVersion: RecipeVersion) {
    selectedVersion.value = version
    diffMode.value = 'version'
    baseLabel.value = `v${version.version_number}（历史）`
    targetLabel.value = `v${currentVersion.version_number}（当前）`
    drawerOpen.value = false
    diffLoading.value = true

    try {
      const res = await apiCompareVersions(recipeId, version.id, currentVersion.id)
      diffResult.value = normalizeApiDiff(res.diff)
    } catch {
      const baseVer = await fetchVersionOrDemo(version.id, version.version_number)
      diffResult.value = localCompare(
        {
          ingredients: baseVer.ingredients ?? [],
          process_steps: baseVer.process_steps ?? [],
        },
        {
          ingredients: currentVersion.ingredients ?? [],
          process_steps: currentVersion.process_steps ?? [],
        },
      )
    } finally {
      diffLoading.value = false
    }
  }

  async function compareEncyclopedia(currentVersion: RecipeVersion, recipeName: string) {
    diffMode.value = 'encyclopedia'
    baseLabel.value = `百科 · ${recipeName}`
    targetLabel.value = '我的配比'
    diffLoading.value = true

    try {
      const res = await apiCompareEncyclopedia(recipeId)
      baseLabel.value = `百科 · ${res.encyclopedia_name || recipeName}`
      diffResult.value = normalizeApiDiff(res.diff)
    } catch {
      const demoBase = getDemoEncyclopediaSnapshot()
      diffResult.value = localCompare(
        demoBase,
        {
          ingredients: currentVersion.ingredients ?? [],
          process_steps: currentVersion.process_steps ?? [],
        },
      )
    } finally {
      diffLoading.value = false
    }
  }

  function clearDiff() {
    diffMode.value = null
    diffResult.value = null
    selectedVersion.value = null
  }

  async function fetchVersionOrDemo(versionId: number, versionNumber: number): Promise<RecipeVersion> {
    try {
      const res = await getVersion(recipeId, versionId)
      return res.version
    } catch {
      return getDemoVersionDetail(versionId, versionNumber)
    }
  }

  return {
    versions,
    versionsLoading,
    diffLoading,
    diffMode,
    diffResult,
    baseLabel,
    targetLabel,
    selectedVersion,
    drawerOpen,
    loadVersions,
    openHistoryDrawer,
    closeHistoryDrawer,
    compareWithVersion,
    compareEncyclopedia,
    clearDiff,
  }
}

function getDemoVersions(currentVersionId?: number): VersionListItem[] {
  const now = Date.now()
  const items: VersionListItem[] = [
    { id: 101, version_number: 1, commit_msg: '初次记录', created_at: new Date(now - 86400000 * 14).toISOString() },
    { id: 102, version_number: 2, commit_msg: '减盐，多放一点糖', created_at: new Date(now - 86400000 * 7).toISOString() },
    { id: 103, version_number: 3, commit_msg: '调整火候时间', created_at: new Date(now - 86400000).toISOString() },
  ]
  return currentVersionId ? items.filter((v) => v.id !== currentVersionId).concat(items.filter((v) => v.id === currentVersionId)) : items
}

function getDemoVersionDetail(id: number, versionNumber: number): RecipeVersion {
  const base = { id, recipe_id: 1, version_number: versionNumber, commit_msg: '', created_at: new Date().toISOString(), images: [] }
  if (versionNumber === 1) return { ...base, commit_msg: '初次记录', ingredients: [{ name: '五花肉', amount: 500, unit: 'g' }], process_steps: [{ order: 1, content: '切块焯水' }] }
  return { ...base, commit_msg: '调整火候时间', ingredients: [{ name: '五花肉', amount: 500, unit: 'g' }], process_steps: [{ order: 1, content: '切块焯水' }] }
}

function getDemoEncyclopediaSnapshot() {
  return { ingredients: [{ name: '五花肉', amount: 500, unit: 'g' }], process_steps: [{ order: 1, content: '切块焯水' }] }
}

export function getDemoRecipe(): MyRecipe {
  const version = getDemoVersionDetail(103, 3)
  return {
    id: 1,
    user_id: '1',
    recipe_name: '红烧肉',
    current_version_id: 103,
    user_rating: 4,
    cover_image_url: undefined,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    current_version: version,
  }
}
