<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listTrashRecipes, restoreRecipe, permanentDeleteRecipe } from '@/api/recipe'
import type { RecipeListItem } from '@/types/recipe'
import ConfirmSheet from '@/components/mobile/ConfirmSheet.vue'

const router = useRouter()
const items = ref<RecipeListItem[]>([])
const loading = ref(true)
const purgeOpen = ref(false)
const purging = ref(false)
const purgeTarget = ref<RecipeListItem | null>(null)

onMounted(() => { loadTrash() })

async function loadTrash() {
  loading.value = true
  try {
    const res = await listTrashRecipes({ page: 1, page_size: 50 })
    items.value = res.items ?? []
  } catch {
    items.value = []
  } finally {
    loading.value = false
  }
}

async function handleRestore(id: number) {
  try {
    await restoreRecipe(id)
    items.value = items.value.filter(i => i.id !== id)
  } catch { /* ignore */ }
}

function openPermanentDelete(item: RecipeListItem) {
  purgeTarget.value = item
  purgeOpen.value = true
}

async function confirmPermanentDelete() {
  if (!purgeTarget.value) return
  purging.value = true
  try {
    await permanentDeleteRecipe(purgeTarget.value.id)
    items.value = items.value.filter(i => i.id !== purgeTarget.value!.id)
    purgeOpen.value = false
    purgeTarget.value = null
  } catch { /* ignore */ }
  finally {
    purging.value = false
  }
}

function goBack() { router.back() }
</script>

<template>
  <div class="page">
    <header class="nav-bar" style="display:flex;align-items:center;gap:12px;padding:12px 0;margin-bottom:16px;">
      <button class="nav-bar__back" aria-label="返回" @click="goBack" style="border:none;background:none;font-size:20px;cursor:pointer;color:var(--color-text);">←</button>
      <span class="nav-bar__title" style="font-family:var(--font-display);font-size:18px;font-weight:600;">回收站</span>
    </header>

    <div v-if="loading" class="state-card">加载中…</div>
    <div v-else-if="items.length === 0" class="empty-state">
      <p class="empty-state__title">回收站是空的</p>
      <p class="empty-state__desc">删除的菜谱会出现在这里，30天内可恢复</p>
    </div>

    <ul v-else class="recipe-flow">
      <li v-for="item in items" :key="item.id" class="recipe-card">
        <div class="recipe-card__body" style="flex:1;">
          <h3 class="recipe-card__name">{{ item.recipe_name }}</h3>
          <p class="recipe-card__meta">v{{ item.current_version_number }}</p>
        </div>
        <div style="display:flex;gap:8px;align-items:center;">
          <button class="toolbar-btn" type="button" @click="handleRestore(item.id)">恢复</button>
          <button class="toolbar-btn" type="button" @click="openPermanentDelete(item)" style="color:var(--color-danger);border-color:rgba(193,102,107,0.35);">彻底删除</button>
        </div>
      </li>
    </ul>

    <ConfirmSheet
      :open="purgeOpen"
      title="彻底删除"
      :message="`「${purgeTarget?.recipe_name ?? ''}」将被永久删除，此操作不可恢复。`"
      confirm-text="彻底删除"
      cancel-text="再想想"
      tone="danger"
      :loading="purging"
      @close="purgeOpen = false"
      @confirm="confirmPermanentDelete"
    />
  </div>
</template>
