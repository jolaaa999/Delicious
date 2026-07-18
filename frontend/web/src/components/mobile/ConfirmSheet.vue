<script setup lang="ts">
defineProps<{
  open: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  tone?: 'danger' | 'default'
  loading?: boolean
}>()

const emit = defineEmits<{
  close: []
  confirm: []
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="confirm-sheet">
      <div
        v-if="open"
        class="confirm-overlay"
        role="presentation"
        @click.self="emit('close')"
      >
        <div
          class="confirm-sheet"
          role="dialog"
          aria-modal="true"
          :aria-label="title"
        >
          <div class="confirm-sheet__handle" />
          <h2 class="confirm-sheet__title">{{ title }}</h2>
          <p class="confirm-sheet__message">{{ message }}</p>
          <div class="confirm-sheet__actions">
            <button
              type="button"
              class="confirm-sheet__btn confirm-sheet__btn--ghost"
              :disabled="loading"
              @click="emit('close')"
            >
              {{ cancelText || '取消' }}
            </button>
            <button
              type="button"
              class="confirm-sheet__btn"
              :class="tone === 'danger' ? 'confirm-sheet__btn--danger' : 'confirm-sheet__btn--primary'"
              :disabled="loading"
              @click="emit('confirm')"
            >
              {{ loading ? '处理中…' : (confirmText || '确认') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 220;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgba(44, 36, 32, 0.4);
}

.confirm-sheet {
  width: 100%;
  max-width: 480px;
  padding: 8px var(--space-page-x) calc(20px + var(--safe-bottom));
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  background: var(--color-surface-elevated);
  box-shadow: 0 -8px 40px var(--color-shadow-lg);
}

.confirm-sheet__handle {
  width: 36px;
  height: 4px;
  margin: 4px auto 18px;
  border-radius: var(--radius-full);
  background: var(--color-border-strong);
}

.confirm-sheet__title {
  margin: 0 0 10px;
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text);
  letter-spacing: 0.02em;
}

.confirm-sheet__message {
  margin: 0 0 24px;
  font-family: var(--font-body);
  font-size: 14px;
  line-height: 1.65;
  color: var(--color-text-secondary);
}

.confirm-sheet__actions {
  display: flex;
  gap: 10px;
}

.confirm-sheet__btn {
  flex: 1;
  padding: 14px 16px;
  border: none;
  border-radius: var(--radius-md);
  font-family: var(--font-body);
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: transform var(--duration-fast), background var(--duration-fast), opacity var(--duration-fast);
}

.confirm-sheet__btn:active:not(:disabled) {
  transform: scale(0.98);
}

.confirm-sheet__btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.confirm-sheet__btn--ghost {
  border: 1px solid var(--color-border-strong);
  background: var(--color-surface);
  color: var(--color-text-secondary);
}

.confirm-sheet__btn--primary {
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  color: var(--color-text-inverse);
  box-shadow: 0 8px 24px rgba(196, 92, 38, 0.25);
}

.confirm-sheet__btn--danger {
  border: 1px solid rgba(193, 102, 107, 0.35);
  background: rgba(193, 102, 107, 0.12);
  color: var(--color-danger);
}

.confirm-sheet-enter-active,
.confirm-sheet-leave-active {
  transition: opacity var(--duration-normal) var(--ease-out);
}

.confirm-sheet-enter-active .confirm-sheet,
.confirm-sheet-leave-active .confirm-sheet {
  transition: transform var(--duration-normal) var(--ease-out);
}

.confirm-sheet-enter-from,
.confirm-sheet-leave-to {
  opacity: 0;
}

.confirm-sheet-enter-from .confirm-sheet,
.confirm-sheet-leave-to .confirm-sheet {
  transform: translateY(100%);
}
</style>
