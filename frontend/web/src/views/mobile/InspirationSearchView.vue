<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { searchEncyclopedia } from '@/api/encyclopedia'
import type { EncyclopediaListItem } from '@/types/recipe'

const router = useRouter()
const keyword = ref('')
const items = ref<EncyclopediaListItem[]>([])
const loading = ref(false)
const searched = ref(false)

async function handleSearch() {
  if (!keyword.value.trim()) return
  loading.value = true
  searched.value = true
  try {
    const res = await searchEncyclopedia({ keyword: keyword.value.trim(), page: 1, page_size: 20 })
    items.value = res.items
  } catch {
    items.value = []
  } finally {
    loading.value = false
  }
}

function goDetail(id: number) {
  router.push(`/m/inspiration/${id}`)
}
</script>

<template>
  <div class="page inspiration-page">
    <header class="page-header">
      <p class="page-eyebrow">Delicious</p>
      <h1 class="page-title">找灵感</h1>
      <p class="page-subtitle">搜索百科菜谱，发现新味道</p>
    </header>

    <div class="search-card">
      <div class="search-bar">
        <svg class="search-bar__icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <circle cx="11" cy="11" r="7" stroke="currentColor" stroke-width="1.6" />
          <path d="M20 20L16.5 16.5" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
        </svg>
        <input
          v-model="keyword"
          class="search-bar__input"
          type="search"
          placeholder="搜索菜名，如「红烧肉」"
          enterkeyhint="search"
          @keyup.enter="handleSearch"
        />
        <button class="search-bar__btn" type="button" @click="handleSearch">搜索</button>
      </div>
      <p class="search-tip">支持按菜名快速检索百科菜谱</p>
    </div>

    <div v-if="loading" class="state-card">正在寻找灵感…</div>
    <div v-else-if="searched && items.length === 0" class="state-card">暂无相关菜谱</div>

    <ul v-else class="recipe-flow">
      <li v-for="item in items" :key="item.id" class="recipe-card" @click="goDetail(item.id)">
        <div class="recipe-card__cover">
          <img v-if="item.cover_image_url" :src="item.cover_image_url" :alt="item.name" />
          <div v-else class="recipe-card__placeholder">{{ item.name.charAt(0) }}</div>
        </div>
        <div class="recipe-card__body">
          <h3 class="recipe-card__name">{{ item.name }}</h3>
          <p v-if="item.description" class="recipe-card__desc">{{ item.description }}</p>
          <div v-if="item.category" class="recipe-card__tag">{{ item.category }}</div>
        </div>
      </li>
    </ul>

    <div v-if="!searched && !loading" class="empty-hint">
      <p>输入菜名，从百科中寻找烹饪灵感</p>
    </div>
  </div>
</template>
