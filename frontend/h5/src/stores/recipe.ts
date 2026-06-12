import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RecipeListItem } from '@/types/recipe'
import { listRecipes } from '@/api/recipe'

export const useRecipeStore = defineStore('recipe', () => {
  const items = ref<RecipeListItem[]>([])
  const loading = ref(false)

  async function fetchList() {
    loading.value = true
    try {
      const res = await listRecipes({ page: 1, page_size: 50, order_by: 'updated_at', desc: true })
      items.value = res.items
    } finally {
      loading.value = false
    }
  }

  return { items, loading, fetchList }
})
