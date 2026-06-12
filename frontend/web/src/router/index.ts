import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '@/layouts/AdminLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: AdminLayout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { title: '数据总览' },
        },
        {
          path: 'recipes',
          name: 'Recipes',
          component: () => import('@/views/RecipeListView.vue'),
          meta: { title: '菜谱管理' },
        },
        {
          path: 'recipes/:id',
          name: 'RecipeDetail',
          component: () => import('@/views/RecipeDetailView.vue'),
          meta: { title: '菜谱详情' },
        },
      ],
    },
  ],
})

router.afterEach((to) => {
  document.title = `${to.meta.title || '管理端'} · 人间烟火`
})

export default router
