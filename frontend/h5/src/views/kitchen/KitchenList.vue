<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRecipeStore } from '@/stores/recipe'
import { resolveImageUrl } from '@/api/upload'

const router = useRouter()
const store = useRecipeStore()

onMounted(() => {
  store.fetchList().catch(() => {})
})

function goCreate() {
  router.push('/kitchen/new')
}

function goDetail(id: number) {
  router.push(`/kitchen/${id}`)
}

function renderStars(rating: number) {
  return '★'.repeat(rating) + '☆'.repeat(Math.max(0, 5 - rating))
}
</script>

<template>
  <div class="page kitchen-page">
    <header class="page-header">
      <h1 class="page-title">我的厨房</h1>
      <p class="page-subtitle">记录每一道用心做的菜</p>
    </header>

    <button class="fab" aria-label="新增菜品" @click="goCreate">
      <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path d="M12 5V19M5 12H19" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
      </svg>
      <span>新增菜品</span>
    </button>

    <div v-if="store.loading" class="state-msg">加载中…</div>

    <ul v-else-if="store.items.length" class="recipe-flow">
      <li
        v-for="item in store.items"
        :key="item.id"
        class="recipe-card"
        @click="goDetail(item.id)"
      >
        <div class="recipe-card__cover">
          <img v-if="item.cover_image_url" :src="resolveImageUrl(item.cover_image_url)" :alt="item.name" />
          <div v-else class="recipe-card__placeholder">{{ item.name.charAt(0) }}</div>
        </div>
        <div class="recipe-card__body">
          <h3 class="recipe-card__name">{{ item.name }}</h3>
          <p v-if="item.user_rating" class="recipe-card__rating">{{ renderStars(item.user_rating) }}</p>
          <p class="recipe-card__meta">v{{ item.current_version_number }}</p>
        </div>
      </li>
    </ul>

    <div v-else class="empty-state">
      <p class="empty-state__title">厨房还是空的</p>
      <p class="empty-state__desc">点击上方按钮，记录你的第一道菜</p>
    </div>
  </div>
</template>

<style scoped>
.kitchen-page {
  padding: var(--page-padding);
  padding-top: max(var(--page-padding), env(safe-area-inset-top));
}

.page-header {
  margin-bottom: 20px;
}

.fab {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 14px;
  margin-bottom: 24px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  color: #fff;
  border-radius: var(--radius-md);
  font-size: 15px;
  font-weight: 500;
  box-shadow: 0 4px 16px var(--color-shadow);
  transition: transform var(--transition-fast);
}

.fab svg {
  width: 20px;
  height: 20px;
}

.fab:active {
  transform: scale(0.98);
}

.recipe-flow {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.recipe-card {
  display: flex;
  gap: 14px;
  padding: 12px;
  background: var(--color-bg-elevated);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  box-shadow: 0 2px 8px var(--color-shadow);
  cursor: pointer;
}

.recipe-card:active {
  opacity: 0.92;
}

.recipe-card__cover {
  width: 80px;
  height: 80px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  flex-shrink: 0;
  background: var(--color-bg-muted);
}

.recipe-card__cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.recipe-card__placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-display);
  font-size: 1.75rem;
  color: var(--color-primary-light);
}

.recipe-card__name {
  font-family: var(--font-display);
  font-size: 1.05rem;
}

.recipe-card__rating {
  color: var(--color-accent);
  font-size: 13px;
  letter-spacing: 1px;
  margin-top: 4px;
}

.recipe-card__meta {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-top: 4px;
}

.state-msg {
  text-align: center;
  padding: 48px;
  color: var(--color-text-muted);
}

.empty-state {
  text-align: center;
  padding: 64px 24px;
}

.empty-state__title {
  font-family: var(--font-display);
  font-size: 1.2rem;
  margin-bottom: 8px;
}

.empty-state__desc {
  font-size: 14px;
  color: var(--color-text-muted);
}
</style>
