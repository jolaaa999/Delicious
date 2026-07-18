<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { searchEncyclopedia, listCategories } from '@/api/encyclopedia'
import type { CategoryDTO } from '@/api/encyclopedia'
import { resolveImageUrl } from '@/api/upload'
import LangSwitch from '@/components/mobile/LangSwitch.vue'
import type { EncyclopediaListItem } from '@/types/recipe'

const router = useRouter()
const keyword = ref('')
const items = ref<EncyclopediaListItem[]>([])
const loading = ref(false)
const searched = ref(false)
const searchError = ref('')
const displayLang = ref<'en' | 'zh'>('zh')
const selectedCategory = ref('')
const categories = ref<CategoryDTO[]>([])
const filtersExpanded = ref(true)

onMounted(async () => {
  try {
    const res = await listCategories()
    categories.value = res.items ?? []
  } catch {
    categories.value = []
  }
})

async function handleSearch() {
  if (!keyword.value.trim()) return
  loading.value = true
  searched.value = true
  searchError.value = ''
  try {
    const params: Record<string, unknown> = {
      keyword: keyword.value.trim(),
      page: 1,
      page_size: 12,
      lang: displayLang.value,
    }
    if (selectedCategory.value) {
      params.category = selectedCategory.value
    }
    const res = await searchEncyclopedia(params)
    items.value = res.items ?? []
  } catch (e: unknown) {
    items.value = []
    const msg = (e as { message?: string })?.message || ''
    if (msg.includes('timeout')) {
      searchError.value = '搜索超时，联网翻译较慢，请稍后重试'
    } else {
      searchError.value = msg || '搜索失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}

function toggleCategory(name: string) {
  selectedCategory.value = selectedCategory.value === name ? '' : name
  if (searched.value && keyword.value.trim()) {
    handleSearch()
  }
}

watch(displayLang, () => {
  if (searched.value && keyword.value.trim()) {
    handleSearch()
  }
})

function goDetail(id: number) {
  router.push({ path: `/m/inspiration/${id}`, query: { lang: displayLang.value } })
}
</script>

<template>
  <div class="page inspiration-page">
    <header class="page-header">
      <div class="page-header__row">
        <div>
          <p class="page-eyebrow">Delicious</p>
          <h1 class="page-title">找灵感</h1>
          <p class="page-subtitle">联网搜索公开菜谱，发现新味道</p>
        </div>
        <LangSwitch v-model="displayLang" :disabled="loading" />
      </div>
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
      <p class="search-tip">支持中文菜名，结果来自 Spoonacular / TheMealDB / Forkify / DummyJSON</p>

      <div
        v-if="searched && !loading && !searchError && items.length > 0 && categories.length"
        class="filter-panel"
      >
        <button
          type="button"
          class="filter-panel__toggle"
          :aria-expanded="filtersExpanded"
          @click="filtersExpanded = !filtersExpanded"
        >
          <span class="filter-panel__label">
            分类筛选
            <span v-if="selectedCategory" class="filter-panel__active">· {{ selectedCategory }}</span>
          </span>
          <span class="filter-panel__chevron" :class="{ 'filter-panel__chevron--open': filtersExpanded }">▾</span>
        </button>
        <div v-show="filtersExpanded" class="category-chips">
          <button
            v-for="cat in categories"
            :key="cat.id"
            type="button"
            class="chip"
            :class="{ 'chip--active': selectedCategory === cat.name }"
            @click="toggleCategory(cat.name)"
          >{{ cat.name }}</button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="state-card">正在寻找灵感…</div>
    <div v-else-if="searchError" class="state-card form-error">{{ searchError }}</div>
    <div v-else-if="searched && items.length === 0" class="state-card">未找到相关菜谱，换个关键词试试</div>

    <ul v-else class="recipe-flow">
      <li v-for="item in items" :key="item.id" class="recipe-card" @click="goDetail(item.id)">
        <div class="recipe-card__cover">
          <img v-if="item.cover_image_url" :src="resolveImageUrl(item.cover_image_url)" :alt="item.name" />
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
