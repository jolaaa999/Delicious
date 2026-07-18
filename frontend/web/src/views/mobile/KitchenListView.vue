<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRecipeStore } from '@/stores/recipe'
import { resolveImageUrl } from '@/api/upload'
import { exportRecipes, importRecipes } from '@/api/recipe'
import type { ExportRecipe } from '@/api/recipe'

const router = useRouter()
const store = useRecipeStore()
const fileInput = ref<HTMLInputElement | null>(null)
const importing = ref(false)

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

function goTrash() {
  router.push('/m/kitchen/trash')
}

async function handleExport() {
  try {
    const res = await exportRecipes()
    const blob = new Blob([JSON.stringify(res.recipes, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `delicious-recipes-${new Date().toISOString().slice(0, 10)}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch {
    // silently fail
  }
}

function handleImportClick() {
  fileInput.value?.click()
}

async function handleImportFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  importing.value = true
  try {
    const text = await file.text()
    const data = JSON.parse(text) as ExportRecipe[]
    await importRecipes(data)
    await store.fetchList()
  } catch {
    // silently fail
  } finally {
    importing.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}
</script>

<template>
  <div class="page kitchen-page">
    <header class="page-header">
      <p class="page-eyebrow">Delicious</p>
      <h1 class="page-title">我的厨房</h1>
      <p class="page-subtitle">记录每一道用心做的菜</p>
    </header>

    <div class="toolbar">
      <button class="toolbar-btn" type="button" @click="handleExport" :disabled="store.items.length === 0">
        <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" width="16" height="16">
          <path d="M12 4V16M8 8L12 4L16 8M4 16V19C4 19.55 4.45 20 5 20H19C19.55 20 20 19.55 20 19V16" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        导出
      </button>
      <button class="toolbar-btn" type="button" @click="handleImportClick" :disabled="importing">
        <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" width="16" height="16">
          <path d="M12 20V8M8 12L12 8L16 12M4 16V19C4 19.55 4.45 20 5 20H19C19.55 20 20 19.55 20 19V16" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        {{ importing ? '导入中…' : '导入' }}
      </button>
      <button class="toolbar-btn" type="button" @click="goTrash">
        <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" width="16" height="16">
          <path d="M3 6H5H21M8 6V4C8 3.45 8.45 3 9 3H15C15.55 3 16 3.45 16 4V6M19 6V20C19 20.55 18.55 21 18 21H6C5.45 21 5 20.55 5 20V6" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        回收站
      </button>
    </div>

    <input ref="fileInput" type="file" accept=".json" class="file-input-hidden" @change="handleImportFile" />

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
          <img v-if="item.cover_image_url" :src="resolveImageUrl(item.cover_image_url)" :alt="item.recipe_name" />
          <div v-else class="recipe-card__placeholder">{{ item.recipe_name.charAt(0) }}</div>
        </div>
        <div class="recipe-card__body">
          <h3 class="recipe-card__name">{{ item.recipe_name }}</h3>
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
