<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getRecipe } from '@/api/recipe'
import { useRecipeDiff, getDemoRecipe } from '@/composables/useRecipeDiff'
import type { MyRecipe } from '@/types/recipe'
import type { VersionListItem } from '@/types/diff'
import VersionHistoryDrawer from '@/components/diff/VersionHistoryDrawer.vue'
import DiffPanel from '@/components/diff/DiffPanel.vue'
import { resolveImageUrl } from '@/api/upload'

const route = useRoute()
const router = useRouter()
const recipeId = Number(route.params.id)

const recipe = ref<MyRecipe | null>(null)
const loading = ref(true)
const showUnchanged = ref(false)

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
  router.push(`/kitchen/${recipeId}/edit`)
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
</script>

<template>
  <div class="detail-page">
    <header class="nav-bar">
      <button class="nav-bar__back" aria-label="返回" @click="goBack">←</button>
      <span class="nav-bar__title">{{ recipe?.name ?? '菜品详情' }}</span>
      <button class="nav-bar__edit" @click="goEdit">编辑</button>
    </header>

    <div v-if="loading" class="state-msg">加载中…</div>

    <template v-else-if="recipe && currentVersion">
      <!-- 封面与评分 -->
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

      <!-- 操作栏 -->
      <div class="action-bar">
        <button class="action-btn" @click="openHistoryDrawer">
          <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M12 8V12L15 15M21 12C21 16.97 16.97 21 12 21C7.03 21 3 16.97 3 12C3 7.03 7.03 3 12 3C16.97 3 21 7.03 21 12Z" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/>
          </svg>
          历史版本
        </button>
        <button class="action-btn action-btn--accent" @click="onCompareEncyclopedia">
          <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M9 5H7C5.9 5 5 5.9 5 7V19C5 20.1 5.9 21 7 21H17C18.1 21 19 20.1 19 19V7C19 5.9 18.1 5 17 5H15M9 5C9 6.1 9.9 7 11 7H13C14.1 7 15 6.1 15 5M9 5C9 3.9 9.9 3 11 3H13C14.1 3 15 3.9 15 5" stroke="currentColor" stroke-width="1.6"/>
          </svg>
          对比百科
        </button>
      </div>

      <!-- Diff 对比区 -->
      <section v-if="diffMode && diffResult" class="diff-section">
        <div class="diff-section__head">
          <h2 class="diff-section__title">
            {{ diffMode === 'encyclopedia' ? '百科基准对比' : '版本对比' }}
          </h2>
          <button class="diff-section__close" @click="clearDiff">关闭对比</button>
        </div>

        <div v-if="diffLoading" class="state-msg">对比计算中…</div>
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

      <!-- 当前版本内容 -->
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
    </template>

    <VersionHistoryDrawer
      :open="drawerOpen"
      :versions="versions"
      :current-version-id="recipe?.current_version_id"
      :loading="versionsLoading"
      @close="closeHistoryDrawer"
      @select="onSelectVersion"
    />
  </div>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  min-height: 100dvh;
  padding-bottom: max(24px, env(safe-area-inset-bottom));
}

.nav-bar {
  display: flex;
  align-items: center;
  gap: 8px;
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
  flex-shrink: 0;
}

.nav-bar__title {
  flex: 1;
  font-family: var(--font-display);
  font-size: 1.05rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-bar__edit {
  flex-shrink: 0;
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  color: var(--color-primary);
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-full);
}

.state-msg {
  padding: 48px;
  text-align: center;
  color: var(--color-text-muted);
}

.hero {
  display: flex;
  gap: 16px;
  padding: var(--page-padding);
}

.hero__cover {
  width: 100px;
  height: 100px;
  border-radius: var(--radius-md);
  overflow: hidden;
  flex-shrink: 0;
  background: var(--color-bg-muted);
  box-shadow: 0 4px 12px var(--color-shadow);
}

.hero__cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.hero__placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-display);
  font-size: 2.5rem;
  color: var(--color-primary-light);
}

.hero__name {
  font-family: var(--font-display);
  font-size: 1.35rem;
  margin-bottom: 4px;
}

.hero__rating {
  color: var(--color-accent);
  letter-spacing: 2px;
  font-size: 14px;
}

.hero__version {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-top: 4px;
}

.hero__commit {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-top: 6px;
  font-style: italic;
}

.action-bar {
  display: flex;
  gap: 10px;
  padding: 0 var(--page-padding) 16px;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: transform var(--transition-fast);
}

.action-btn svg {
  width: 18px;
  height: 18px;
}

.action-btn:active {
  transform: scale(0.98);
}

.action-btn--accent {
  color: #fff;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  border-color: transparent;
}

.diff-section,
.content-section {
  padding: 0 var(--page-padding) 24px;
}

.diff-section {
  background: var(--color-bg-elevated);
  margin: 0 var(--page-padding) 16px;
  padding: var(--page-padding);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  box-shadow: 0 4px 16px var(--color-shadow);
}

.diff-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.diff-section__title {
  font-family: var(--font-display);
  font-size: 1.1rem;
}

.diff-section__close {
  font-size: 13px;
  color: var(--color-text-muted);
  padding: 4px 8px;
}

.toggle-unchanged {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  font-size: 13px;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.section-title {
  font-family: var(--font-display);
  font-size: 1rem;
  margin-bottom: 12px;
  color: var(--color-text);
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

.ingredient-item__amount {
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

.step-item__num {
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

.step-item__text {
  flex: 1;
  padding-top: 2px;
}
</style>
