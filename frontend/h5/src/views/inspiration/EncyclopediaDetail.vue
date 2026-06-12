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
  router.push({ path: '/kitchen/new', query: { encyclopedia_id: String(id) } })
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

<style scoped>
.detail-page {
  min-height: 100vh;
  min-height: 100dvh;
  padding-bottom: 32px;
}

.nav-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: max(12px, env(safe-area-inset-top)) var(--page-padding) 12px;
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 10;
}

.nav-bar__back {
  font-size: 20px;
  padding: 4px 8px;
}

.nav-bar__title {
  flex: 1;
  font-family: var(--font-display);
  font-size: 1.05rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.state-msg {
  padding: 48px;
  text-align: center;
  color: var(--color-text-muted);
}

.hero {
  padding: var(--page-padding);
}

.hero__name {
  font-family: var(--font-display);
  font-size: 1.5rem;
  margin-bottom: 8px;
}

.hero__desc {
  font-size: 14px;
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.hero__tag {
  display: inline-block;
  margin-top: 10px;
  padding: 4px 12px;
  font-size: 12px;
  color: var(--color-primary);
  background: rgba(212, 98, 42, 0.1);
  border-radius: var(--radius-full);
}

.add-btn {
  display: block;
  width: calc(100% - var(--page-padding) * 2);
  margin: 0 var(--page-padding) 20px;
  padding: 14px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  border-radius: var(--radius-md);
  box-shadow: 0 4px 16px var(--color-shadow);
}

.content {
  padding: 0 var(--page-padding);
}

.section-title {
  font-family: var(--font-display);
  font-size: 1rem;
  margin-bottom: 12px;
}

.ingredient-list {
  list-style: none;
  margin-bottom: 24px;
}

.ingredient-item {
  display: flex;
  justify-content: space-between;
  padding: 10px 14px;
  margin-bottom: 6px;
  background: var(--color-bg-elevated);
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
  font-size: 14px;
}

.amount {
  color: var(--color-primary);
  font-weight: 500;
}

.step-list {
  list-style: none;
}

.step-item {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border);
  font-size: 14px;
  line-height: 1.55;
}

.step-num {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  background: var(--color-primary);
  border-radius: var(--radius-full);
}
</style>
