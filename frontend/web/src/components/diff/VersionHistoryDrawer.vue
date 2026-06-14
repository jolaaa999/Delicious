<script setup lang="ts">
import type { VersionListItem } from '@/types/diff'

defineProps<{
  open: boolean
  versions: VersionListItem[]
  currentVersionId?: number
  loading: boolean
}>()

const emit = defineEmits<{
  close: []
  select: [version: VersionListItem]
}>()

function formatDate(iso: string) {
  const d = new Date(iso)
  return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="drawer-overlay" @click.self="emit('close')">
        <div class="drawer" role="dialog" aria-label="历史版本">
          <div class="drawer__handle" />
          <header class="drawer__header">
            <h2 class="drawer__title">历史版本</h2>
            <button class="drawer__close" type="button" aria-label="关闭" @click="emit('close')">×</button>
          </header>

          <div v-if="loading" class="drawer__loading">加载版本记录…</div>

          <ul v-else class="version-list">
            <li
              v-for="ver in versions"
              :key="ver.id"
              class="version-item"
              :class="{ 'version-item--current': ver.id === currentVersionId }"
              @click="emit('select', ver)"
            >
              <div class="version-item__head">
                <span class="version-item__num">v{{ ver.version_number }}</span>
                <span v-if="ver.id === currentVersionId" class="version-item__tag">当前</span>
              </div>
              <p v-if="ver.commit_msg" class="version-item__msg">{{ ver.commit_msg }}</p>
              <time class="version-item__time">{{ formatDate(ver.created_at) }}</time>
            </li>
          </ul>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-overlay {
  position: fixed;
  inset: 0;
  z-index: 200;
  background: rgba(44, 36, 32, 0.4);
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.drawer {
  width: 100%;
  max-height: 72vh;
  padding: 8px var(--space-page-x) calc(20px + var(--safe-bottom));
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
  background: var(--color-surface-elevated);
  box-shadow: 0 -8px 40px var(--color-shadow-lg);
  overflow-y: auto;
}

.drawer__handle {
  width: 36px;
  height: 4px;
  margin: 4px auto 16px;
  border-radius: var(--radius-full);
  background: var(--color-border-strong);
}

.drawer__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.drawer__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 600;
}

.drawer__close {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: var(--radius-full);
  background: var(--color-surface-muted);
  font-size: 22px;
  color: var(--color-text-muted);
  cursor: pointer;
}

.drawer__loading {
  padding: 32px;
  text-align: center;
  font-size: 14px;
  color: var(--color-text-muted);
}

.version-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.version-item {
  padding: 14px 16px;
  border-radius: var(--radius-md);
  background: var(--color-surface-muted);
  border: 1px solid var(--color-border);
  cursor: pointer;
  transition: border-color var(--duration-fast), background var(--duration-fast);
}

.version-item:active {
  background: var(--color-primary-soft);
}

.version-item--current {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
}

.version-item__head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.version-item__num {
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.version-item__tag {
  padding: 2px 8px;
  border-radius: var(--radius-full);
  background: var(--color-primary);
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-inverse);
}

.version-item__msg {
  margin: 0 0 4px;
  font-size: 14px;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.version-item__time {
  font-size: 12px;
  color: var(--color-text-muted);
}

.drawer-enter-active,
.drawer-leave-active {
  transition: opacity var(--duration-normal) var(--ease-out);
}

.drawer-enter-active .drawer,
.drawer-leave-active .drawer {
  transition: transform var(--duration-normal) var(--ease-out);
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}

.drawer-enter-from .drawer,
.drawer-leave-to .drawer {
  transform: translateY(100%);
}
</style>
