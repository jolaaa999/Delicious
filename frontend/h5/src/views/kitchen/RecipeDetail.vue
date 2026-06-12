<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goBack() {
  router.back()
}
</script>

<template>
  <div class="page detail-page">
    <header class="nav-bar">
      <button class="nav-bar__back" @click="goBack">←</button>
      <span class="nav-bar__title">菜品详情</span>
    </header>
    <div class="placeholder">详情 · 版本历史 · Diff · 阶段 5 完善</div>
  </div>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  min-height: 100dvh;
}

.nav-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: max(12px, env(safe-area-inset-top)) var(--page-padding) 12px;
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
}

.nav-bar__back {
  font-size: 20px;
  padding: 4px 8px;
}

.nav-bar__title {
  font-family: var(--font-display);
  font-size: 1.1rem;
}

.placeholder {
  padding: 48px;
  text-align: center;
  color: var(--color-text-muted);
}
</style>
