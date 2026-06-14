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
        <span v-if="activeTab === tab.key" class="tabbar__dot" />
      </span>
      <span class="tabbar__label">{{ tab.label }}</span>
    </button>
  </nav>
</template>

<style scoped>
.tabbar {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 100;
  display: flex;
  align-items: stretch;
  height: calc(var(--tabbar-height) + var(--safe-bottom));
  padding-bottom: var(--safe-bottom);
  background: rgba(255, 248, 240, 0.92);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-top: 1px solid var(--color-border);
  box-shadow: 0 -4px 24px var(--color-shadow);
}

.tabbar__item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 8px 0 6px;
  color: var(--color-text-muted);
}

.tabbar__item--active {
  color: var(--color-primary);
}

.tabbar__icon-wrap {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
}

.tabbar__icon {
  width: 24px;
  height: 24px;
}

.tabbar__dot {
  position: absolute;
  bottom: -2px;
  left: 50%;
  transform: translateX(-50%);
  width: 4px;
  height: 4px;
  border-radius: 999px;
  background: var(--color-primary);
}

.tabbar__label {
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.04em;
}

.tabbar__item--active .tabbar__label {
  font-weight: 600;
}
</style>
