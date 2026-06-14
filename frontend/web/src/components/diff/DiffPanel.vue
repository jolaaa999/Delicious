<script setup lang="ts">
import { computed } from 'vue'
import type { VersionDiffResult } from '@/types/diff'

const props = defineProps<{
  diff: VersionDiffResult
  baseLabel: string
  targetLabel: string
  showUnchanged: boolean
}>()

const typeLabels: Record<string, string> = {
  added: '新增',
  removed: '删除',
  changed: '修改',
  unchanged: '未变',
}

function filterItems(items?: VersionDiffResult['ingredient_diff']) {
  if (!items?.length) return []
  if (props.showUnchanged) return items
  return items.filter((i) => i.type !== 'unchanged')
}

const ingredients = computed(() => filterItems(props.diff.ingredient_diff))
const steps = computed(() => filterItems(props.diff.process_diff))
</script>

<template>
  <div class="diff-panel">
    <div class="diff-panel__labels">
      <span class="diff-panel__label diff-panel__label--base">{{ baseLabel }}</span>
      <span class="diff-panel__arrow">→</span>
      <span class="diff-panel__label diff-panel__label--target">{{ targetLabel }}</span>
    </div>

    <section v-if="ingredients.length" class="diff-panel__section">
      <h3 class="diff-panel__section-title">配料差异</h3>
      <ul class="diff-list">
        <li
          v-for="(item, i) in ingredients"
          :key="'ing-' + i"
          class="diff-item"
          :class="`diff-item--${item.type}`"
        >
          <span class="diff-item__badge">{{ typeLabels[item.type] }}</span>
          <div class="diff-item__content">
            <span class="diff-item__name">{{ item.name }}</span>
            <div v-if="item.type === 'changed'" class="diff-item__values">
              <span class="diff-item__old">{{ item.base }}</span>
              <span class="diff-item__arrow">→</span>
              <span class="diff-item__new">{{ item.target }}</span>
            </div>
            <span v-else-if="item.type === 'added'" class="diff-item__new">{{ item.target }}</span>
            <span v-else-if="item.type === 'removed'" class="diff-item__old">{{ item.base }}</span>
          </div>
        </li>
      </ul>
    </section>

    <section v-if="steps.length" class="diff-panel__section">
      <h3 class="diff-panel__section-title">步骤差异</h3>
      <ul class="diff-list">
        <li
          v-for="(item, i) in steps"
          :key="'step-' + i"
          class="diff-item"
          :class="`diff-item--${item.type}`"
        >
          <span class="diff-item__badge">{{ typeLabels[item.type] }}</span>
          <div class="diff-item__content">
            <span class="diff-item__name">{{ item.name }}</span>
            <div v-if="item.type === 'changed'" class="diff-item__values diff-item__values--stack">
              <span class="diff-item__old">{{ item.base }}</span>
              <span class="diff-item__new">{{ item.target }}</span>
            </div>
            <span v-else-if="item.type === 'added'" class="diff-item__new">{{ item.target }}</span>
            <span v-else-if="item.type === 'removed'" class="diff-item__old">{{ item.base }}</span>
          </div>
        </li>
      </ul>
    </section>

    <p v-if="!ingredients.length && !steps.length" class="diff-panel__empty">两个版本完全一致</p>
  </div>
</template>

<style scoped>
.diff-panel__labels {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}

.diff-panel__label {
  padding: 6px 14px;
  border-radius: var(--radius-full);
  font-size: 12px;
  font-weight: 600;
}

.diff-panel__label--base {
  background: var(--color-surface-muted);
  color: var(--color-text-secondary);
}

.diff-panel__label--target {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.diff-panel__arrow {
  font-size: 14px;
  color: var(--color-text-muted);
}

.diff-panel__section {
  margin-bottom: 16px;
}

.diff-panel__section-title {
  margin: 0 0 10px;
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.diff-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.diff-item {
  display: flex;
  gap: 10px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--color-surface-muted);
  border-left: 3px solid var(--color-border-strong);
}

.diff-item--added { border-left-color: var(--color-diff-added); }
.diff-item--removed { border-left-color: var(--color-diff-removed); }
.diff-item--changed { border-left-color: var(--color-diff-changed); }

.diff-item__badge {
  flex-shrink: 0;
  padding: 2px 8px;
  height: fit-content;
  border-radius: var(--radius-full);
  font-size: 11px;
  font-weight: 600;
}

.diff-item--added .diff-item__badge { background: rgba(90, 143, 94, 0.15); color: var(--color-diff-added); }
.diff-item--removed .diff-item__badge { background: rgba(193, 102, 107, 0.15); color: var(--color-diff-removed); }
.diff-item--changed .diff-item__badge { background: rgba(212, 160, 84, 0.2); color: var(--color-diff-changed); }

.diff-item__content {
  flex: 1;
  min-width: 0;
}

.diff-item__name {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 4px;
}

.diff-item__values {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  flex-wrap: wrap;
}

.diff-item__values--stack {
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}

.diff-item__old {
  color: var(--color-text-muted);
  text-decoration: line-through;
}

.diff-item__new {
  color: var(--color-text-secondary);
  line-height: 1.45;
}

.diff-item__arrow {
  color: var(--color-text-muted);
  font-size: 12px;
}

.diff-panel__empty {
  margin: 0;
  padding: 20px;
  text-align: center;
  font-size: 14px;
  color: var(--color-text-muted);
}
</style>
