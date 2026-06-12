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
  router.push(`/inspiration/${id}`)
}
</script>

<template>
  <div class="page inspiration-page">
    <header class="page-header">
      <h1 class="page-title">找灵感</h1>
      <p class="page-subtitle">搜索百科菜谱，发现新味道</p>
    </header>

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
      <button class="search-bar__btn" @click="handleSearch">搜索</button>
    </div>

    <div v-if="loading" class="state-msg">正在寻找灵感…</div>
    <div v-else-if="searched && items.length === 0" class="state-msg">暂无相关菜谱</div>

    <ul v-else class="recipe-flow">
      <li
        v-for="item in items"
        :key="item.id"
        class="recipe-card"
        @click="goDetail(item.id)"
      >
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

<style scoped>
.inspiration-page {
  padding: var(--page-padding);
  padding-top: max(var(--page-padding), env(safe-area-inset-top));
}

.page-header {
  margin-bottom: 20px;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 4px 4px 14px;
  background: var(--color-bg-elevated);
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border);
  box-shadow: 0 2px 12px var(--color-shadow);
  margin-bottom: 24px;
}

.search-bar__icon {
  width: 18px;
  height: 18px;
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.search-bar__input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  color: var(--color-text);
  min-width: 0;
}

.search-bar__input::placeholder {
  color: var(--color-text-muted);
}

.search-bar__btn {
  padding: 10px 18px;
  background: var(--color-primary);
  color: #fff;
  border-radius: var(--radius-full);
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  transition: background var(--transition-fast);
}

.search-bar__btn:active {
  background: var(--color-primary-dark);
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
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.recipe-card:active {
  transform: scale(0.98);
}

.recipe-card__cover {
  width: 88px;
  height: 88px;
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
  font-size: 2rem;
  color: var(--color-primary-light);
}

.recipe-card__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.recipe-card__name {
  font-family: var(--font-display);
  font-size: 1.1rem;
  margin-bottom: 4px;
}

.recipe-card__desc {
  font-size: 13px;
  color: var(--color-text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.recipe-card__tag {
  display: inline-block;
  margin-top: 6px;
  padding: 2px 8px;
  font-size: 11px;
  color: var(--color-primary);
  background: rgba(212, 98, 42, 0.1);
  border-radius: var(--radius-full);
  align-self: flex-start;
}

.state-msg,
.empty-hint {
  text-align: center;
  padding: 48px 16px;
  color: var(--color-text-muted);
  font-size: 14px;
}
</style>
