<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { deleteRecipe, getRecipe } from '@/api/recipe'
import { useRecipeDiff, getDemoRecipe } from '@/composables/useRecipeDiff'
import { useRecipeStore } from '@/stores/recipe'
import type { MyRecipe } from '@/types/recipe'
import type { VersionListItem } from '@/types/diff'
import VersionHistoryDrawer from '@/components/diff/VersionHistoryDrawer.vue'
import DiffPanel from '@/components/diff/DiffPanel.vue'
import ConfirmSheet from '@/components/mobile/ConfirmSheet.vue'
import { resolveImageUrl } from '@/api/upload'

const route = useRoute()
const router = useRouter()
const recipeId = Number(route.params.id)

const recipe = ref<MyRecipe | null>(null)
const loading = ref(true)
const deleting = ref(false)
const deleteOpen = ref(false)
const fromApi = ref(false)
const showUnchanged = ref(false)
const store = useRecipeStore()

const {
  versions,
  versionsLoading,
  diffLoading,
  diffMode,
  diffResult,
  baseLabel,
  targetLabel,
  drawerOpen,
  loadVersions,
  openHistoryDrawer,
  closeHistoryDrawer,
  compareWithVersion,
  compareEncyclopedia,
  clearDiff,
} = useRecipeDiff(recipeId)

const currentVersion = computed(() => recipe.value?.current_version)

const stars = computed(() => {
  const r = recipe.value?.user_rating ?? 0
  return '★'.repeat(r) + '☆'.repeat(Math.max(0, 5 - r))
})

onMounted(async () => {
  try {
    const res = await getRecipe(recipeId)
    recipe.value = res.recipe
    fromApi.value = true
  } catch {
    recipe.value = getDemoRecipe()
  } finally {
    loading.value = false
    if (recipe.value) {
      await loadVersions(recipe.value.current_version_id)
    }
  }
})

function goBack() {
  router.back()
}

function goEdit() {
  router.push(`/m/kitchen/${recipeId}/edit`)
}

async function onSelectVersion(ver: VersionListItem) {
  if (!currentVersion.value) return
  if (ver.id === recipe.value?.current_version_id) {
    clearDiff()
    return
  }
  await compareWithVersion(ver, currentVersion.value)
}

async function onCompareEncyclopedia() {
  if (!currentVersion.value || !recipe.value) return
  await compareEncyclopedia(currentVersion.value, recipe.value.name)
}

function openDeleteConfirm() {
  if (!recipe.value || !fromApi.value || deleting.value) return
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!recipe.value || !fromApi.value) return
  deleting.value = true
  try {
    await deleteRecipe(recipeId)
    store.items = store.items.filter((item) => item.id !== recipeId)
    deleteOpen.value = false
    ElMessage.success('已移入回收站')
    router.replace('/m/kitchen')
  } catch (e: unknown) {
    ElMessage.error((e as { message?: string })?.message || '删除失败')
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div class="detail-page">
    <header class="nav-bar">
      <button class="nav-bar__back" aria-label="返回" @click="goBack">←</button>
      <span class="nav-bar__title">{{ recipe?.name ?? '菜品详情' }}</span>
      <button class="nav-bar__edit" @click="goEdit">编辑</button>
    </header>

    <div v-if="loading" class="state-card">加载中…</div>

    <template v-else-if="recipe && currentVersion">
      <section class="hero">
        <div class="hero__cover">
          <img
            v-if="currentVersion.images?.[0] || recipe.cover_image_url"
            :src="resolveImageUrl(currentVersion.images?.[0] || recipe.cover_image_url || '')"
            :alt="recipe.name"
          />
          <div v-else class="hero__placeholder">{{ recipe.name.charAt(0) }}</div>
        </div>
        <div class="hero__info">
          <h1 class="hero__name">{{ recipe.name }}</h1>
          <p v-if="recipe.user_rating" class="hero__rating">{{ stars }}</p>
          <p class="hero__version">当前 v{{ currentVersion.version_number }}</p>
          <p v-if="currentVersion.commit_msg" class="hero__commit">{{ currentVersion.commit_msg }}</p>
        </div>
      </section>

      <div class="action-bar">
        <button class="action-btn" type="button" @click="openHistoryDrawer">历史版本</button>
        <button class="action-btn action-btn--accent" type="button" @click="onCompareEncyclopedia">对比百科</button>
      </div>

      <section v-if="diffMode && diffResult" class="diff-section">
        <div class="diff-section__head">
          <h2 class="diff-section__title">
            {{ diffMode === 'encyclopedia' ? '百科基准对比' : '版本对比' }}
          </h2>
          <button class="diff-section__close" type="button" @click="clearDiff">关闭对比</button>
        </div>

        <div v-if="diffLoading" class="state-card">对比计算中…</div>
        <DiffPanel
          v-else
          :diff="diffResult"
          :base-label="baseLabel"
          :target-label="targetLabel"
          :show-unchanged="showUnchanged"
        />

        <label class="toggle-unchanged">
          <input v-model="showUnchanged" type="checkbox" />
          显示未变更项
        </label>
      </section>

      <section v-if="!diffMode" class="content-section">
        <h2 class="section-title">配料</h2>
        <ul class="ingredient-list">
          <li v-for="(ing, i) in currentVersion.ingredients" :key="i" class="ingredient-item">
            <span class="ingredient-item__name">{{ ing.name }}</span>
            <span class="ingredient-item__amount">{{ ing.amount }}{{ ing.unit }}</span>
          </li>
        </ul>

        <h2 class="section-title">步骤</h2>
        <ol class="step-list">
          <li v-for="step in currentVersion.process_steps" :key="step.order" class="step-item">
            <span class="step-item__num">{{ step.order }}</span>
            <span class="step-item__text">{{ step.content }}</span>
          </li>
        </ol>
      </section>

      <section v-if="fromApi" class="danger-section">
        <button
          type="button"
          class="delete-btn"
          :disabled="deleting"
          @click="openDeleteConfirm"
        >
          {{ deleting ? '删除中…' : '删除此菜品' }}
        </button>
      </section>
    </template>

    <VersionHistoryDrawer
      :open="drawerOpen"
      :versions="versions"
      :current-version-id="recipe?.current_version_id"
      :loading="versionsLoading"
      @close="closeHistoryDrawer"
      @select="onSelectVersion"
    />

    <ConfirmSheet
      :open="deleteOpen"
      title="移入回收站"
      :message="`「${recipe?.name ?? ''}」将移入回收站，30 天内可恢复。`"
      confirm-text="确认删除"
      cancel-text="再想想"
      tone="danger"
      :loading="deleting"
      @close="deleteOpen = false"
      @confirm="confirmDelete"
    />
  </div>
</template>
