<script setup lang="ts">
import { computed } from 'vue'
import type { VersionDiffResult } from '@/types/diff'
import DiffIngredientList from './DiffIngredientList.vue'
import DiffProcessList from './DiffProcessList.vue'

const props = defineProps<{
  diff: VersionDiffResult
  baseLabel: string
  targetLabel: string
  showUnchanged?: boolean
}>()

const hasChanges = computed(() => {
  const filter = (t: string) => t !== 'unchanged'
  const ing = props.diff.ingredient_diffs.some((d) => {
    const type = typeof d.type === 'number'
      ? ['', 'unchanged', 'added', 'removed', 'modified'][d.type]
      : d.type
    return filter(type as string)
  })
  const proc = props.diff.process_diffs.some((d) => {
    const type = typeof d.type === 'number'
      ? ['', 'unchanged', 'added', 'removed', 'modified'][d.type]
      : d.type
    return filter(type as string)
  })
  return ing || proc
})
</script>

<template>
  <div class="diff-panel">
    <div class="diff-panel__legend">
      <span class="legend-item legend-item--added">+ 新增</span>
      <span class="legend-item legend-item--removed">− 删除</span>
      <span class="legend-item legend-item--modified">~ 修改</span>
    </div>

    <div class="diff-panel__labels">
      <span>{{ baseLabel }}</span>
      <span>{{ targetLabel }}</span>
    </div>

    <p class="diff-panel__summary">{{ diff.summary }}</p>

    <section v-if="hasChanges || showUnchanged" class="diff-section">
      <h3 class="diff-section__title">配料</h3>
      <DiffIngredientList :diffs="diff.ingredient_diffs" :show-unchanged="showUnchanged" />
    </section>

    <section v-if="hasChanges || showUnchanged" class="diff-section">
      <h3 class="diff-section__title">步骤</h3>
      <DiffProcessList :diffs="diff.process_diffs" :show-unchanged="showUnchanged" />
    </section>

    <div v-if="!hasChanges && !showUnchanged" class="diff-panel__no-change">
      ✓ 完全一致，没有差异
    </div>
  </div>
</template>

<style scoped>
.diff-panel {
  padding: 4px 0;
}

.diff-panel__legend {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.legend-item {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: var(--radius-full);
}

.legend-item--added {
  color: #2d6b47;
  background: rgba(74, 157, 111, 0.15);
}

.legend-item--removed {
  color: #9c3333;
  background: rgba(201, 68, 68, 0.12);
}

.legend-item--modified {
  color: #8a6420;
  background: rgba(212, 146, 42, 0.15);
}

.diff-panel__labels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  padding: 0 4px 0 28px;
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text);
}

.diff-panel__summary {
  font-size: 13px;
  color: var(--color-text-secondary);
  padding: 10px 14px;
  margin-bottom: 16px;
  background: var(--color-bg-muted);
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--color-primary);
}

.diff-section {
  margin-bottom: 20px;
}

.diff-section__title {
  font-family: var(--font-display);
  font-size: 1rem;
  margin-bottom: 10px;
  color: var(--color-text);
}

.diff-panel__no-change {
  text-align: center;
  padding: 32px;
  color: var(--color-success);
  font-size: 14px;
  font-weight: 500;
}
</style>
