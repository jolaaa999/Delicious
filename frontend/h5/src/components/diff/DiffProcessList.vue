<script setup lang="ts">
import { computed } from 'vue'
import type { ProcessStepDiff } from '@/types/diff'
import { normalizeDiffType } from '@/types/diff'

const props = defineProps<{
  diffs: ProcessStepDiff[]
  showUnchanged?: boolean
}>()

const visibleDiffs = computed(() => {
  if (props.showUnchanged) return props.diffs
  return props.diffs.filter((d) => normalizeDiffType(d.type) !== 'unchanged')
})

function stepText(step?: { content: string; duration_minutes?: number }): string {
  if (!step) return '—'
  const dur = step.duration_minutes ? ` · ${step.duration_minutes}分钟` : ''
  return step.content + dur
}
</script>

<template>
  <div class="diff-process">
    <div v-if="visibleDiffs.length === 0" class="diff-empty">步骤无差异</div>

    <div
      v-for="(item, idx) in visibleDiffs"
      :key="idx"
      class="diff-row"
      :class="`diff-row--${normalizeDiffType(item.type)}`"
    >
      <div class="diff-row__order">{{ item.order }}</div>
      <span class="diff-row__badge">
        {{ normalizeDiffType(item.type) === 'added' ? '+' : normalizeDiffType(item.type) === 'removed' ? '−' : normalizeDiffType(item.type) === 'modified' ? '~' : '=' }}
      </span>

      <div class="diff-row__cells">
        <div class="cell cell--base">
          {{ stepText(item.base) }}
        </div>
        <div class="cell cell--target">
          {{ stepText(item.target) }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.diff-empty {
  padding: 16px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 13px;
}

.diff-row {
  display: flex;
  align-items: stretch;
  gap: 6px;
  margin-bottom: 8px;
}

.diff-row__order {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 600;
  color: var(--color-primary);
  background: rgba(212, 98, 42, 0.12);
  border-radius: var(--radius-full);
  margin-top: 10px;
}

.diff-row__badge {
  flex-shrink: 0;
  width: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  font-family: ui-monospace, monospace;
  margin-top: 8px;
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
  line-height: 1.5;
  border-radius: var(--radius-sm);
  word-break: break-word;
}

.diff-row--added .diff-row__badge { color: var(--color-success); }
.diff-row--added .cell--base { opacity: 0.3; }
.diff-row--added .cell--target {
  background: rgba(74, 157, 111, 0.15);
  border: 1px solid rgba(74, 157, 111, 0.35);
}

.diff-row--removed .diff-row__badge { color: var(--color-danger); }
.diff-row--removed .cell--base {
  background: rgba(201, 68, 68, 0.12);
  border: 1px solid rgba(201, 68, 68, 0.3);
  text-decoration: line-through;
}
.diff-row--removed .cell--target { opacity: 0.3; }

.diff-row--modified .diff-row__badge { color: var(--color-warning); }
.diff-row--modified .cell--base {
  background: rgba(201, 68, 68, 0.08);
  border: 1px solid rgba(201, 68, 68, 0.2);
}
.diff-row--modified .cell--target {
  background: rgba(74, 157, 111, 0.1);
  border: 1px solid rgba(74, 157, 111, 0.25);
}

.diff-row--unchanged .cell {
  background: var(--color-bg-muted);
  color: var(--color-text-secondary);
}
</style>
