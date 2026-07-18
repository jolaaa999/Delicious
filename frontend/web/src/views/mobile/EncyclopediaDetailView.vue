<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getEncyclopedia, getEncyclopediaTags } from '@/api/encyclopedia'
import type { TagDTO } from '@/api/encyclopedia'
import LangSwitch from '@/components/mobile/LangSwitch.vue'
import type { Ingredient, ProcessStep } from '@/types/recipe'

interface EncyclopediaDetail {
  id: number
  name: string
  description?: string
  category?: string
  tags?: string[]
  source?: string
  ingredients: Ingredient[]
  process_steps: ProcessStep[]
}

const route = useRoute()
const router = useRouter()
const id = Number(route.params.id)

const recipe = ref<EncyclopediaDetail | null>(null)
const loading = ref(true)
const translating = ref(false)
const displayLang = ref<'en' | 'zh'>(
  route.query.lang === 'en' ? 'en' : 'zh',
)
const recipeCache = ref<Partial<Record<'en' | 'zh', EncyclopediaDetail>>>({})
const ready = ref(false)
const tags = ref<TagDTO[]>([])
const loadError = ref('')

async function loadTags() {
  try {
    const res = await getEncyclopediaTags(id)
    tags.value = res.items ?? []
  } catch {
    tags.value = []
  }
}

const sourceLabel = computed(() => {
  const src = recipe.value?.source
  if (!src) return ''
  if (src === 'spoonacular') return '来源：Spoonacular 公开菜谱'
  if (src === 'themealdb') return '来源：TheMealDB 公开菜谱'
  if (src === 'forkify') return '来源：Forkify 公开菜谱'
  if (src === 'dummyjson') return '来源：DummyJSON 示例菜谱'
  if (src === 'howtocook') return '来源：HowToCook 开源中文菜谱'
  if (src === 'projkitchen') return '来源：厨房计划 Proj.Kitchen'
  return `来源：${src}`
})

async function loadRecipe(lang: 'en' | 'zh', options?: { silent?: boolean }) {
  if (recipeCache.value[lang]) {
    recipe.value = recipeCache.value[lang]
    loadError.value = ''
    return
  }
  if (!options?.silent) {
    translating.value = true
  }
  try {
    const res = await getEncyclopedia(id, { lang })
    if (!res?.name) {
      throw new Error('菜谱内容为空')
    }
    const normalized: EncyclopediaDetail = {
      id: res.id,
      name: res.name,
      description: res.description,
      category: res.category,
      tags: res.tags,
      source: res.source,
      ingredients: res.ingredients ?? [],
      process_steps: res.process_steps ?? [],
    }
    recipeCache.value[lang] = normalized
    recipe.value = normalized
    loadError.value = ''
  } catch (e: unknown) {
    if (!recipe.value) recipe.value = null
    const msg = (e as { message?: string })?.message || ''
    loadError.value = msg || '加载失败，请稍后重试'
  } finally {
    translating.value = false
  }
}

onMounted(async () => {
  loading.value = true
  try {
    const initialLang = route.query.lang === 'en' || route.query.lang === 'zh'
      ? route.query.lang as 'en' | 'zh'
      : 'zh'
    displayLang.value = initialLang
    await loadRecipe(initialLang, { silent: true })
  } finally {
    loading.value = false
    ready.value = true
  }
  loadTags()
})

watch(displayLang, async (lang) => {
  if (!ready.value) return
  await loadRecipe(lang)
})

function goBack() {
  router.back()
}

function addToKitchen() {
  router.push({ path: '/m/kitchen/new', query: { encyclopedia_id: String(id) } })
}

function formatIngredientAmount(ing: { amount?: number; unit?: string }) {
  const amount = Number(ing.amount)
  const unit = (ing.unit || '').trim()
  if (!amount || amount === 0) {
    return unit === '适量' ? '适量' : (unit && unit !== '份' ? unit : '适量')
  }
  return `${amount}${unit}`
}
</script>

<template>
  <div class="detail-page">
    <header class="nav-bar">
      <button class="nav-bar__back" aria-label="返回" @click="goBack">←</button>
      <span class="nav-bar__title">{{ recipe?.name ?? '百科详情' }}</span>
      <LangSwitch v-if="recipe" v-model="displayLang" :disabled="translating" />
    </header>

    <div v-if="loading" class="state-msg">加载中…</div>
    <div v-else-if="!recipe" class="state-msg">{{ loadError || '未找到该菜谱' }}</div>

    <template v-else>
      <div v-if="translating" class="translate-hint">正在翻译…</div>

      <section class="hero">
        <div class="hero__info">
          <h1 class="hero__name">{{ recipe.name }}</h1>
          <p v-if="recipe.description" class="hero__desc">{{ recipe.description }}</p>
          <div v-if="recipe.category" class="hero__tag">{{ recipe.category }}</div>
          <div v-if="tags.length" class="hero__tags">
            <span v-for="t in tags" :key="t.id" class="hero__tag hero__tag--secondary">{{ t.name }}</span>
          </div>
          <p v-if="sourceLabel" class="hero__source">{{ sourceLabel }}</p>
        </div>
      </section>

      <button class="add-btn" type="button" @click="addToKitchen">
        加入我的厨房
      </button>

      <section class="content">
        <h2 class="section-title">标准配料</h2>
        <ul class="ingredient-list">
          <li v-for="(ing, i) in recipe.ingredients" :key="i" class="ingredient-item">
            <span>{{ ing.name }}</span>
            <span class="amount">{{ formatIngredientAmount(ing) }}</span>
          </li>
        </ul>

        <h2 class="section-title">标准步骤</h2>
        <ol class="step-list">
          <li v-for="step in recipe.process_steps" :key="step.order" class="step-item">
            <span class="step-num">{{ step.order }}</span>
            <span class="step-item__text">{{ step.content }}</span>
          </li>
        </ol>
      </section>
    </template>
  </div>
</template>
