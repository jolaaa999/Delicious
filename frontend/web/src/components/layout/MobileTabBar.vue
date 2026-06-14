<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

type TabKey = 'inspiration' | 'kitchen'

interface TabItem {
  key: TabKey
  label: string
  path: string
  icon: 'spark' | 'pot'
}

const tabs: TabItem[] = [
  { key: 'inspiration', label: '找灵感', path: '/m/inspiration', icon: 'spark' },
  { key: 'kitchen', label: '我的厨房', path: '/m/kitchen', icon: 'pot' },
]

const route = useRoute()
const router = useRouter()

const activeTab = computed<TabKey>(() => {
  const tab = route.meta.tab as TabKey | undefined
  if (tab) return tab
  return route.path.startsWith('/m/kitchen') ? 'kitchen' : 'inspiration'
})

function navigate(tab: TabItem) {
  if (activeTab.value !== tab.key) {
    router.push(tab.path)
  }
}
</script>

<template>
  <nav class="tabbar" role="tablist" aria-label="主导航">
    <button
      v-for="tab in tabs"
      :key="tab.key"
      class="tabbar__item"
      :class="{ 'tabbar__item--active': activeTab === tab.key }"
      role="tab"
      :aria-selected="activeTab === tab.key"
      @click="navigate(tab)"
    >
      <span class="tabbar__pill" :class="{ 'tabbar__pill--visible': activeTab === tab.key }" />
      <span class="tabbar__icon-wrap">
        <svg
          v-if="tab.icon === 'spark'"
          class="tabbar__icon"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <path
            d="M12 2C12 2 8 8 8 12.5C8 15.5 9.8 18 12 18C14.2 18 16 15.5 16 12.5C16 8 12 2 12 2Z"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linejoin="round"
          />
          <path
            d="M12 18V22M9 21H15"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linecap="round"
          />
        </svg>
        <svg
          v-else
          class="tabbar__icon"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <path
            d="M4 11C4 8.2 6.2 6 9 6H15C17.8 6 20 8.2 20 11V13H4V11Z"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linejoin="round"
          />
          <path
            d="M6 13V16C6 18.2 7.8 20 10 20H14C16.2 20 18 18.2 18 16V13"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linejoin="round"
          />
          <path
            d="M2 13H22M9 6V4M15 6V4"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linecap="round"
          />
        </svg>
      </span>
      <span class="tabbar__label">{{ tab.label }}</span>
    </button>
  </nav>
</template>

<style scoped>
.tabbar {
  position: fixed;
  left: 12px;
  right: 12px;
  bottom: calc(10px + var(--safe-bottom));
  z-index: 100;
  display: flex;
  align-items: stretch;
  height: var(--tabbar-height);
  padding: 4px;
  border-radius: var(--radius-xl);
  background: rgba(255, 253, 249, 0.94);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--color-border);
  box-shadow: 0 8px 32px var(--color-shadow-lg), 0 2px 8px var(--color-shadow);
}

.tabbar__item {
  position: relative;
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  padding: 6px 0;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color var(--duration-fast) var(--ease-out);
}

.tabbar__item--active {
  color: var(--color-primary);
}

.tabbar__pill {
  position: absolute;
  inset: 2px 4px;
  border-radius: calc(var(--radius-xl) - 4px);
  background: var(--color-primary-soft);
  opacity: 0;
  transform: scale(0.92);
  transition: opacity var(--duration-normal) var(--ease-out), transform var(--duration-normal) var(--ease-out);
  pointer-events: none;
}

.tabbar__pill--visible {
  opacity: 1;
  transform: scale(1);
}

.tabbar__icon-wrap {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
}

.tabbar__icon {
  width: 22px;
  height: 22px;
}

.tabbar__label {
  position: relative;
  z-index: 1;
  font-family: var(--font-body);
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.06em;
}

.tabbar__item--active .tabbar__label {
  font-weight: 600;
}
</style>
