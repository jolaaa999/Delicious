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
    } catch {
      items.value = [
        {
          id: 1,
          name: '红烧肉',
          user_rating: 4,
          current_version_number: 3,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      ]
    } finally {
      loading.value = false
    }
  }

  return { items, loading, fetchList }
})
