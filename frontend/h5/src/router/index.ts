import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/inspiration',
  },
  // ── 找灵感 ──
  {
    path: '/inspiration',
    name: 'Inspiration',
    component: () => import('@/views/inspiration/InspirationSearch.vue'),
    meta: { tab: 'inspiration', title: '找灵感' },
  },
  {
    path: '/inspiration/:id',
    name: 'EncyclopediaDetail',
    component: () => import('@/views/inspiration/EncyclopediaDetail.vue'),
    meta: { title: '百科详情', hideTabBar: true },
  },
  // ── 我的厨房 ──
  {
    path: '/kitchen',
    name: 'Kitchen',
    component: () => import('@/views/kitchen/KitchenList.vue'),
    meta: { tab: 'kitchen', title: '我的厨房' },
  },
  {
    path: '/kitchen/new',
    name: 'RecipeCreate',
    component: () => import('@/views/kitchen/RecipeEdit.vue'),
    meta: { title: '新增菜品', hideTabBar: true },
  },
  {
    path: '/kitchen/:id',
    name: 'RecipeDetail',
    component: () => import('@/views/kitchen/RecipeDetail.vue'),
    meta: { title: '菜品详情', hideTabBar: true },
  },
  {
    path: '/kitchen/:id/edit',
    name: 'RecipeEdit',
    component: () => import('@/views/kitchen/RecipeEdit.vue'),
    meta: { title: '编辑菜品', hideTabBar: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_to, _from, saved) {
    return saved ?? { top: 0 }
  },
})

router.afterEach((to) => {
  const title = (to.meta.title as string) || '人间烟火'
  document.title = title === '人间烟火' ? title : `${title} · 人间烟火`
})

export default router
