<script setup lang="ts">
import type { VersionListItem } from '@/types/diff'

defineProps<{
  open: boolean
  versions: VersionListItem[]
  currentVersionId?: number
  loading?: boolean
}>()

const emit = defineEmits<{
  close: []
  select: [version: VersionListItem]
}>()

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString('zh-CN', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return iso
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="drawer-root">
        <div class="drawer-mask" @click="emit('close')" />
        <aside class="drawer" role="dialog" aria-label="历史版本">
          <div class="drawer__handle" />

          <header class="drawer__header">
            <h2 class="drawer__title">历史版本</h2>
            <button class="drawer__close" aria-label="关闭" @click="emit('close')">×</button>
          </header>

          <div v-if="loading" class="drawer__loading">加载中…</div>

          <ul v-else class="version-list">
            <li
              v-for="ver in versions"
              :key="ver.id"
              class="version-item"
              :class="{ 'version-item--current': ver.id === currentVersionId }"
              @click="emit('select', ver)"
            >
              <div class="version-item__top">
                <span class="version-item__num">v{{ ver.version_number }}</span>
                <span v-if="ver.id === currentVersionId" class="version-item__tag">当前</span>
              </div>
              <p v-if="ver.commit_msg" class="version-item__msg">{{ ver.commit_msg }}</p>
              <time class="version-item__time">{{ formatDate(ver.created_at) }}</time>
            </li>
          </ul>

          <p v-if="!loading && versions.length === 0" class="drawer__empty">暂无历史版本</p>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-root {
  position: fixed;
  inset: 0;
  z-index: 200;
  display: flex;
  align-items: flex-end;
}

.drawer-mask {
  position: absolute;
  inset: 0;
  background: rgba(61, 44, 30, 0.4);
  backdrop-filter: blur(2px);
}

.drawer {
  position: relative;
  width: 100%;
  max-height: 75vh;
  background: var(--color-bg-elevated);
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  padding-bottom: max(16px, env(safe-area-inset-bottom));
  box-shadow: 0 -8px 32px var(--color-shadow);
  display: flex;
  flex-direction: column;
  animation: slide-up 0.32s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes slide-up {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

.drawer__handle {
  width: 36px;
  height: 4px;
  background: var(--color-border);
  border-radius: var(--radius-full);
  margin: 10px auto 0;
}

.drawer__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px var(--page-padding) 8px;
}

.drawer__title {
  font-family: var(--font-display);
  font-size: 1.15rem;
}

.drawer__close {
  font-size: 28px;
  line-height: 1;
  color: var(--color-text-muted);
  padding: 0 4px;
}

.drawer__loading,
.drawer__empty {
  padding: 32px;
  text-align: center;
  color: var(--color-text-muted);
  font-size: 14px;
}

.version-list {
  list-style: none;
  overflow-y: auto;
  padding: 0 var(--page-padding) 8px;
  flex: 1;
}

.version-item {
  padding: 14px 16px;
  margin-bottom: 8px;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: border-color var(--transition-fast), transform var(--transition-fast);
}

.version-item:active {
  transform: scale(0.99);
}

.version-item--current {
  border-color: var(--color-primary);
  background: rgba(212, 98, 42, 0.06);
}

.version-item__top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.version-item__num {
  font-family: var(--font-display);
  font-size: 1rem;
  color: var(--color-primary);
}

.version-item__tag {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 8px;
  color: #fff;
  background: var(--color-primary);
  border-radius: var(--radius-full);
}

.version-item__msg {
  font-size: 14px;
  color: var(--color-text);
  margin-bottom: 4px;
}

.version-item__time {
  font-size: 12px;
  color: var(--color-text-muted);
}

.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 0.25s ease;
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}
</style>
