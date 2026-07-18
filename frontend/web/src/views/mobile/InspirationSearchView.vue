<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { searchEncyclopedia, listCategories } from '@/api/encyclopedia'
import type { CategoryDTO } from '@/api/encyclopedia'
import { resolveImageUrl } from '@/api/upload'
import LangSwitch from '@/components/mobile/LangSwitch.vue'
import type { EncyclopediaListItem } from '@/types/recipe'
import { splitHighlightParts } from '@/utils/highlight'

const PAGE_SIZE = 12

const router = useRouter()
const keyword = ref('')
const items = ref<EncyclopediaListItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const searched = ref(false)
const searchError = ref('')
const displayLang = ref<'en' | 'zh'>('zh')
const selectedCategory = ref('')
const categories = ref<CategoryDTO[]>([])
const filtersExpanded = ref(true)
const page = ref(1)
const hasMore = ref(false)
const loadMoreError = ref('')
const sentinel = ref<HTMLElement | null>(null)
const activeKeyword = ref('')

let observer: IntersectionObserver | null = null

onMounted(async () => {
  try {
    const res = await listCategories()
    categories.value = res.items ?? []
  } catch {
    categories.value = []
  }
  setupObserver()
})

onUnmounted(() => {
  observer?.disconnect()
  observer = null
})

function setupObserver() {
  observer?.disconnect()
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) {
        loadMore()
      }
    },
    { root: null, rootMargin: '160px 0px', threshold: 0 },
  )
  if (sentinel.value) {
    observer.observe(sentinel.value)
  }
}

watch(sentinel, async (el) => {
  await nextTick()
  if (!observer) setupObserver()
  observer?.disconnect()
  if (el) observer?.observe(el)
})

async function fetchPage(nextPage: number, append: boolean) {
  const params: Record<string, unknown> = {
    keyword: keyword.value.trim(),
    page: nextPage,
    page_size: PAGE_SIZE,
    lang: displayLang.value,
  }
  if (selectedCategory.value) {
    params.category = selectedCategory.value
  }

  const res = await searchEncyclopedia(params)
  const batch = res.items ?? []
  if (!append) {
    activeKeyword.value = keyword.value.trim()
  }
  if (append) {
    const seen = new Set(items.value.map((i) => i.id))
    items.value.push(...batch.filter((i) => !seen.has(i.id)))
  } else {
    items.value = batch
  }

  page.value = nextPage
  // 本页满额则认为可能还有下一页；下一页不足再自然停住
  hasMore.value = batch.length >= PAGE_SIZE
}

async function handleSearch() {
  if (!keyword.value.trim()) return
  loading.value = true
  loadingMore.value = false
  searched.value = true
  searchError.value = ''
  loadMoreError.value = ''
  hasMore.value = false
  page.value = 1
  try {
    await fetchPage(1, false)
  } catch (e: unknown) {
    items.value = []
    hasMore.value = false
    const msg = (e as { message?: string })?.message || ''
    if (msg.includes('timeout')) {
      searchError.value = '搜索超时，联网翻译较慢，请稍后重试'
    } else if (/402|429|daily points limit|quota/i.test(msg)) {
      searchError.value = '部分菜谱源今日额度已用尽，请换个关键词或稍后再试'
    } else {
      searchError.value = msg || '搜索失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (!searched.value || !hasMore.value || loading.value || loadingMore.value) return
  if (!keyword.value.trim()) return

  loadingMore.value = true
  loadMoreError.value = ''
  try {
    await fetchPage(page.value + 1, true)
  } catch (e: unknown) {
    const msg = (e as { message?: string })?.message || ''
    loadMoreError.value = msg.includes('timeout')
      ? '加载超时，请上滑重试'
      : (msg || '加载失败，请重试')
  } finally {
    loadingMore.value = false
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

function nameParts(name: string) {
  // 只按用户输入关键词与菜名的公共片段高亮（如「鸡柳」→「鸡肉…」中的「鸡」）
  if (!activeKeyword.value) return [{ text: name, hit: false }]
  return splitHighlightParts(name, [activeKeyword.value])
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
        <LangSwitch v-model="displayLang" :disabled="loading || loadingMore" />
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
        <button class="search-bar__btn" type="button" :disabled="loading" @click="handleSearch">搜索</button>
      </div>
      <p class="search-tip">支持中文菜名，优先检索 HowToCook / 厨房计划，并聚合 Spoonacular / TheMealDB / Forkify / DummyJSON</p>

      <div
        v-if="searched && !searchError && items.length > 0 && categories.length"
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

    <template v-else-if="items.length > 0">
      <ul class="recipe-flow">
        <li v-for="item in items" :key="item.id" class="recipe-card" @click="goDetail(item.id)">
          <div class="recipe-card__cover">
            <img v-if="item.cover_image_url" :src="resolveImageUrl(item.cover_image_url)" :alt="item.name" />
            <div v-else class="recipe-card__placeholder">{{ item.name.charAt(0) }}</div>
          </div>
          <div class="recipe-card__body">
            <h3 class="recipe-card__name">
              <template v-for="(part, idx) in nameParts(item.name)" :key="idx">
                <mark v-if="part.hit" class="name-hit">{{ part.text }}</mark>
                <template v-else>{{ part.text }}</template>
              </template>
            </h3>
            <p v-if="item.description" class="recipe-card__desc">{{ item.description }}</p>
            <div v-if="item.category" class="recipe-card__tag">{{ item.category }}</div>
          </div>
        </li>
      </ul>

      <div ref="sentinel" class="load-more" aria-live="polite">
        <p v-if="loadingMore" class="load-more__text">正在加载更多…</p>
        <button
          v-else-if="loadMoreError"
          type="button"
          class="load-more__retry"
          @click="loadMore"
        >{{ loadMoreError }}</button>
        <p v-else-if="hasMore" class="load-more__text">下滑加载更多</p>
        <p v-else class="load-more__text load-more__text--muted">已经到底了</p>
      </div>
    </template>

    <div v-if="!searched && !loading" class="empty-hint">
      <p>输入菜名，从百科中寻找烹饪灵感</p>
    </div>
  </div>
</template>
