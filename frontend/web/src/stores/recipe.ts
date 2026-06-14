import { defineStore } from 'pinia'
import { listRecipes } from '@/api/recipe'
import type { RecipeListItem } from '@/types/recipe'

export const useRecipeStore = defineStore('recipe', {
  state: () => ({
    loading: false,
    items: [] as RecipeListItem[],
  }),
  actions: {
    async fetchList() {
      this.loading = true
      try {
        const res = await listRecipes({ page: 1, page_size: 20, desc: true })
        this.items = res.items
      } finally {
        this.loading = false
      }
    },
  },
})
