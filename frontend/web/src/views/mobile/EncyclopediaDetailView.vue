<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getEncyclopedia } from '@/api/encyclopedia'
import type { Ingredient, ProcessStep } from '@/types/recipe'

interface EncyclopediaDetail {
  id: number
  name: string
  description?: string
  category?: string
  tags?: string[]
  ingredients: Ingredient[]
  process_steps: ProcessStep[]
}

const route = useRoute()
const router = useRouter()
const id = Number(route.params.id)

const recipe = ref<EncyclopediaDetail | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getEncyclopedia(id) as { recipe: EncyclopediaDetail }
    recipe.value = res.recipe
  } catch {
    recipe.value = null
  } finally {
    loading.value = false
  }
})

function goBack() {
  router.back()
}

function addToKitchen() {
  router.push({ path: '/m/kitchen/new', query: { encyclopedia_id: String(id) } })
}
</script>

<template>
  <div class="detail-page">
    <header class="nav-bar">
      <button class="nav-bar__back" aria-label="返回" @click="goBack">←</button>
      <span class="nav-bar__title">{{ recipe?.name ?? '百科详情' }}</span>
    </header>

    <div v-if="loading" class="state-msg">加载中…</div>
    <div v-else-if="!recipe" class="state-msg">未找到该菜谱</div>

    <template v-else>
      <section class="hero">
        <h1 class="hero__name">{{ recipe.name }}</h1>
        <p v-if="recipe.description" class="hero__desc">{{ recipe.description }}</p>
        <div v-if="recipe.category" class="hero__tag">{{ recipe.category }}</div>
      </section>

      <button class="add-btn" type="button" @click="addToKitchen">
        加入我的厨房
      </button>

      <section class="content">
        <h2 class="section-title">标准配料</h2>
        <ul class="ingredient-list">
          <li v-for="(ing, i) in recipe.ingredients" :key="i" class="ingredient-item">
            <span>{{ ing.name }}</span>
            <span class="amount">{{ ing.amount }}{{ ing.unit }}</span>
          </li>
        </ul>

        <h2 class="section-title">标准步骤</h2>
        <ol class="step-list">
          <li v-for="step in recipe.process_steps" :key="step.order" class="step-item">
            <span class="step-num">{{ step.order }}</span>
            <span>{{ step.content }}</span>
          </li>
        </ol>
      </section>
    </template>
  </div>
</template>
