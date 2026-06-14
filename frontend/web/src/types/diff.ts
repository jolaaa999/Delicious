export interface VersionListItem {
  id: number
  version_number: number
  commit_msg?: string
  created_at: string
}

export interface VersionDiffItem {
  type: 'added' | 'removed' | 'changed' | 'unchanged'
  name: string
  base?: string
  target?: string
}

export interface VersionDiffResult {
  ingredient_diff: VersionDiffItem[]
  process_diff: VersionDiffItem[]
}
