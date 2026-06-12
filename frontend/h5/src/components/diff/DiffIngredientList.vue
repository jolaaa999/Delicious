<script setup lang="ts">
import { computed } from 'vue'
import type { IngredientDiff } from '@/types/diff'
import { normalizeDiffType } from '@/types/diff'
import { formatDelta, formatIngredient } from '@/utils/diff'

const props = defineProps<{
  diffs: IngredientDiff[]
  showUnchanged?: boolean
}>()

const visibleDiffs = computed(() => {
  if (props.showUnchanged) return props.diffs
  return props.diffs.filter((d) => normalizeDiffType(d.type) !== 'unchanged')
})
</script>

<template>
  <div class="diff-ingredients">
    <div class="diff-ingredients__header">
      <span class="col col--base">基准</span>
      <span class="col col--target">对比</span>
    </div>

    <div v-if="visibleDiffs.length === 0" class="diff-empty">配料无差异</div>

    <div
      v-for="(item, idx) in visibleDiffs"
      :key="idx"
      class="diff-row"
      :class="`diff-row--${normalizeDiffType(item.type)}`"
    >
      <span class="diff-row__badge">{{ normalizeDiffType(item.type) === 'added' ? '+' : normalizeDiffType(item.type) === 'removed' ? '−' : normalizeDiffType(item.type) === 'modified' ? '~' : '=' }}</span>

      <div class="diff-row__cells">
        <div class="cell cell--base">
          <span v-if="item.base">{{ formatIngredient(item.base) }}</span>
          <span v-else class="cell--empty">—</span>
        </div>
        <div class="cell cell--target">
          <span v-if="item.target">{{ formatIngredient(item.target) }}</span>
          <span v-else class="cell--empty">—</span>
          <span v-if="item.amount_delta != null && item.amount_delta !== 0" class="cell__delta">
            {{ formatDelta(item.amount_delta) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.diff-ingredients__header {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 8px;
  padding: 0 4px 0 28px;
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.diff-empty {
  padding: 16px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
}

.diff-row {
  display: flex;
  align-items: stretch;
  gap: 8px;
  margin-bottom: 6px;
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.diff-row__badge {
  flex-shrink: 0;
  width: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  font-family: ui-monospace, monospace;
}

.diff-row__cells {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
  min-width: 0;
}

.cell {
  padding: 10px 12px;
  font-size: 13px;
  line-height: 1.45;
  border-radius: var(--radius-sm);
  word-break: break-word;
}

.cell--empty {
  color: var(--color-text-muted);
}

.cell__delta {
  display: inline-block;
  margin-left: 6px;
  padding: 1px 6px;
  font-size: 11px;
  font-weight: 600;
  border-radius: var(--radius-full);
  background: rgba(0, 0, 0, 0.06);
}

/* Git-diff 风格配色 */
.diff-row--unchanged .diff-row__badge { color: var(--color-text-muted); }
.diff-row--unchanged .cell { background: var(--color-bg-muted); color: var(--color-text-secondary); }

.diff-row--added .diff-row__badge { color: var(--color-success); }
.diff-row--added .cell--base { background: transparent; }
.diff-row--added .cell--target {
  background: rgba(74, 157, 111, 0.15);
  border: 1px solid rgba(74, 157, 111, 0.35);
  color: #2d6b47;
}

.diff-row--removed .diff-row__badge { color: var(--color-danger); }
.diff-row--removed .cell--base {
  background: rgba(201, 68, 68, 0.12);
  border: 1px solid rgba(201, 68, 68, 0.3);
  color: #9c3333;
  text-decoration: line-through;
  text-decoration-color: rgba(201, 68, 68, 0.5);
}
.diff-row--removed .cell--target { background: transparent; }

.diff-row--modified .diff-row__badge { color: var(--color-warning); }
.diff-row--modified .cell--base {
  background: rgba(201, 68, 68, 0.08);
  border: 1px solid rgba(201, 68, 68, 0.2);
}
.diff-row--modified .cell--target {
  background: rgba(74, 157, 111, 0.1);
  border: 1px solid rgba(74, 157, 111, 0.25);
}
.diff-row--modified .cell__delta {
  background: rgba(212, 146, 42, 0.2);
  color: var(--color-warning);
}
</style>
