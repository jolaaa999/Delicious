<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createRecipe, getRecipe, updateRecipe } from '@/api/recipe'
import { getEncyclopedia } from '@/api/encyclopedia'
import type { Ingredient, ProcessStep } from '@/types/recipe'
import StarRating from '@/components/recipe/StarRating.vue'
import ImageUploader from '@/components/recipe/ImageUploader.vue'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => route.name === 'RecipeEdit')
const recipeId = computed(() => (isEdit.value ? Number(route.params.id) : 0))

const saving = ref(false)
const loading = ref(false)
const errorMsg = ref('')

const name = ref('')
const userRating = ref(0)
const commitMsg = ref('')
const processText = ref('')
const imageUrls = ref<string[]>([])
const encyclopediaRecipeId = ref<number | undefined>()

const ingredients = ref<Ingredient[]>([
  { name: '', amount: 0, unit: 'g', note: '' },
])

const processSteps = ref<ProcessStep[]>([
  { order: 1, content: '' },
])

onMounted(async () => {
  const encyId = route.query.encyclopedia_id
  if (encyId) {
    encyclopediaRecipeId.value = Number(encyId)
    await prefillFromEncyclopedia(Number(encyId))
  } else if (isEdit.value) {
    await loadRecipe()
  }
})

async function prefillFromEncyclopedia(id: number) {
  loading.value = true
  try {
    const res = await getEncyclopedia(id) as { recipe: {
      name: string
      ingredients: Ingredient[]
      process_steps: ProcessStep[]
    }}
    name.value = res.recipe.name
    ingredients.value = res.recipe.ingredients.map((i) => ({ ...i }))
    processSteps.value = res.recipe.process_steps.map((s) => ({ ...s }))
    commitMsg.value = '从百科导入'
  } catch {
    errorMsg.value = '加载百科数据失败'
  } finally {
    loading.value = false
  }
}

async function loadRecipe() {
  loading.value = true
  try {
    const res = await getRecipe(recipeId.value)
    const r = res.recipe
    name.value = r.name
    userRating.value = r.user_rating || 0
    encyclopediaRecipeId.value = r.encyclopedia_recipe_id
    const ver = r.current_version
    if (ver) {
      ingredients.value = ver.ingredients?.length
        ? ver.ingredients.map((i) => ({ ...i }))
        : [{ name: '', amount: 0, unit: 'g' }]
      processSteps.value = ver.process_steps?.length
        ? ver.process_steps.map((s) => ({ ...s }))
        : [{ order: 1, content: '' }]
      processText.value = ver.process_text || ''
      imageUrls.value = ver.images?.length ? [...ver.images] : []
    }
  } catch {
    errorMsg.value = '加载失败'
  } finally {
    loading.value = false
  }
}

function addIngredient() {
  ingredients.value.push({ name: '', amount: 0, unit: 'g', note: '' })
}

function removeIngredient(idx: number) {
  if (ingredients.value.length > 1) {
    ingredients.value.splice(idx, 1)
  }
}

function addStep() {
  processSteps.value.push({ order: processSteps.value.length + 1, content: '' })
}

function removeStep(idx: number) {
  if (processSteps.value.length > 1) {
    processSteps.value.splice(idx, 1)
    renumberSteps()
  }
}

function renumberSteps() {
  processSteps.value.forEach((s, i) => {
    s.order = i + 1
  })
}

function validate(): boolean {
  if (!name.value.trim()) {
    errorMsg.value = '请填写菜名'
    return false
  }
  if (ingredients.value.some((i) => !i.name.trim())) {
    errorMsg.value = '请填写所有食材名称'
    return false
  }
  if (processSteps.value.some((s) => !s.content.trim())) {
    errorMsg.value = '请填写所有步骤内容'
    return false
  }
  if (isEdit.value && !commitMsg.value.trim()) {
    errorMsg.value = '编辑时请填写修改备注'
    return false
  }
  errorMsg.value = ''
  return true
}

async function handleSubmit() {
  if (!validate()) return

  saving.value = true
  const images = imageUrls.value.map((u) => u.trim()).filter(Boolean)
  const payload = {
    name: name.value.trim(),
    ingredients: ingredients.value.map((i) => ({
      name: i.name.trim(),
      amount: Number(i.amount) || 0,
      unit: i.unit.trim() || 'g',
      note: i.note?.trim(),
    })),
    process_steps: processSteps.value.map((s, i) => ({
      order: i + 1,
      content: s.content.trim(),
      duration_minutes: s.duration_minutes,
    })),
    process_text: processText.value.trim() || undefined,
    images: images.length ? images : undefined,
    user_rating: userRating.value || undefined,
    encyclopedia_recipe_id: encyclopediaRecipeId.value,
    commit_msg: commitMsg.value.trim() || '初次记录',
  }

  try {
    if (isEdit.value) {
      await updateRecipe(recipeId.value, { ...payload, commit_msg: commitMsg.value.trim() })
      router.replace(`/kitchen/${recipeId.value}`)
    } else {
      const res = await createRecipe(payload)
      router.replace(`/kitchen/${res.recipe.id}`)
    }
  } catch (e: unknown) {
    errorMsg.value = (e as { message?: string })?.message || '保存失败'
  } finally {
    saving.value = false
  }
}

function goBack() {
  router.back()
}
</script>

<template>
  <div class="edit-page">
    <header class="nav-bar">
      <button class="nav-bar__back" type="button" @click="goBack">←</button>
      <span class="nav-bar__title">{{ isEdit ? '编辑菜品' : '新增菜品' }}</span>
      <button
        class="nav-bar__save"
        type="button"
        :disabled="saving || loading"
        @click="handleSubmit"
      >
        {{ saving ? '保存中…' : '保存' }}
      </button>
    </header>

    <div v-if="loading" class="state-msg">加载中…</div>

    <form v-else class="form" @submit.prevent="handleSubmit">
      <p v-if="errorMsg" class="form-error">{{ errorMsg }}</p>

      <section class="form-section">
        <label class="field-label">菜名</label>
        <input v-model="name" class="field-input" type="text" placeholder="如：红烧肉" required />
      </section>

      <section class="form-section">
        <label class="field-label">个人评分</label>
        <StarRating v-model="userRating" />
      </section>

      <section class="form-section">
        <div class="section-head">
          <label class="field-label">食材与重量</label>
          <button type="button" class="link-btn" @click="addIngredient">+ 添加</button>
        </div>
        <div v-for="(ing, idx) in ingredients" :key="idx" class="ingredient-row">
          <input v-model="ing.name" class="field-input field-input--name" placeholder="食材" />
          <input v-model.number="ing.amount" class="field-input field-input--amount" type="number" min="0" step="any" placeholder="量" />
          <input v-model="ing.unit" class="field-input field-input--unit" placeholder="单位" />
          <button type="button" class="row-remove" aria-label="删除" @click="removeIngredient(idx)">×</button>
        </div>
      </section>

      <section class="form-section">
        <div class="section-head">
          <label class="field-label">制作步骤</label>
          <button type="button" class="link-btn" @click="addStep">+ 添加</button>
        </div>
        <div v-for="(step, idx) in processSteps" :key="idx" class="step-row">
          <span class="step-row__num">{{ idx + 1 }}</span>
          <textarea
            v-model="step.content"
            class="field-textarea"
            rows="2"
            placeholder="描述这一步…"
          />
          <button type="button" class="row-remove" @click="removeStep(idx)">×</button>
        </div>
      </section>

      <section class="form-section">
        <label class="field-label">步骤备注（可选）</label>
        <textarea v-model="processText" class="field-textarea" rows="3" placeholder="整体心得…" />
      </section>

      <section class="form-section">
        <label class="field-label">菜品图片</label>
        <ImageUploader v-model="imageUrls" />
      </section>

      <section v-if="isEdit" class="form-section">
        <label class="field-label">修改备注 <span class="required">*</span></label>
        <input
          v-model="commitMsg"
          class="field-input"
          type="text"
          placeholder="如：减盐、调整炖煮时间"
        />
      </section>

      <section v-else class="form-section">
        <label class="field-label">备注（可选）</label>
        <input v-model="commitMsg" class="field-input" type="text" placeholder="初次记录说明" />
      </section>

      <button type="submit" class="submit-btn" :disabled="saving">
        {{ saving ? '保存中…' : '保存菜品' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.edit-page {
  min-height: 100vh;
  min-height: 100dvh;
  padding-bottom: max(32px, env(safe-area-inset-bottom));
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
}

.nav-bar__title {
  flex: 1;
  font-family: var(--font-display);
  font-size: 1.05rem;
}

.nav-bar__save {
  padding: 6px 16px;
  font-size: 14px;
  font-weight: 500;
  color: #fff;
  background: var(--color-primary);
  border-radius: var(--radius-full);
}

.nav-bar__save:disabled {
  opacity: 0.6;
}

.state-msg {
  padding: 48px;
  text-align: center;
  color: var(--color-text-muted);
}

.form {
  padding: var(--page-padding);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-error {
  padding: 10px 14px;
  background: rgba(201, 68, 68, 0.1);
  color: var(--color-danger);
  border-radius: var(--radius-sm);
  font-size: 14px;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.field-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.required {
  color: var(--color-danger);
}

.field-hint {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-top: -6px;
}

.field-input,
.field-textarea {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-text);
  outline: none;
  transition: border-color var(--transition-fast);
}

.field-input:focus,
.field-textarea:focus {
  border-color: var(--color-primary-light);
}

.link-btn {
  font-size: 13px;
  color: var(--color-primary);
  font-weight: 500;
}

.ingredient-row {
  display: grid;
  grid-template-columns: 1fr 64px 52px 32px;
  gap: 6px;
  align-items: center;
}

.step-row {
  display: flex;
  gap: 8px;
  align-items: flex-start;
}

.step-row__num {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  background: var(--color-primary);
  border-radius: var(--radius-full);
}

.image-row {
  display: flex;
  gap: 6px;
  align-items: center;
}

.row-remove {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  font-size: 20px;
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
}

.submit-btn {
  margin-top: 8px;
  padding: 16px;
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  border-radius: var(--radius-md);
  box-shadow: 0 4px 16px var(--color-shadow);
}

.submit-btn:disabled {
  opacity: 0.7;
}
</style>
