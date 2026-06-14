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
  router.push('/m/kitchen/new')
}

function goDetail(id: number) {
  router.push(`/m/kitchen/${id}`)
}

function renderStars(rating: number) {
  return '★'.repeat(rating) + '☆'.repeat(Math.max(0, 5 - rating))
}
</script>

<template>
  <div class="page kitchen-page">
    <header class="page-header">
      <p class="page-eyebrow">Delicious</p>
      <h1 class="page-title">我的厨房</h1>
      <p class="page-subtitle">记录每一道用心做的菜</p>
    </header>

    <button class="fab" type="button" aria-label="新增菜品" @click="goCreate">
      <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path d="M12 5V19M5 12H19" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
      </svg>
      <span>新增菜品</span>
    </button>

    <div v-if="store.loading" class="state-card">加载中…</div>

    <ul v-else-if="store.items.length" class="recipe-flow">
      <li v-for="item in store.items" :key="item.id" class="recipe-card" @click="goDetail(item.id)">
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
