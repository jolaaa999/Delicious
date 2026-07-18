import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import AdminLayout from '@/layouts/AdminLayout.vue'
import MobileLayout from '@/layouts/MobileLayout.vue'

const mobilePrefix = '/m'

const isMobileDevice = () => {
  if (typeof navigator === 'undefined') return false
  return /Android|iPhone|iPad|iPod|Mobile/i.test(navigator.userAgent)
}

const stripMobilePrefix = (path: string) => {
  if (path === mobilePrefix) return '/'
  if (path.startsWith(`${mobilePrefix}/`)) return path.slice(mobilePrefix.length)
  return path
}

const withMobilePrefix = (path: string) => {
  const normalized = path === '/' ? '' : path
  return `${mobilePrefix}${normalized}` || mobilePrefix
}

const resolveTitle = (to: RouteLocationNormalized) => {
  const title = (to.meta.title as string) || '管理端'
  return `${title} · 人间烟火`
}

const desktopRoutes = [
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
]

const mobileRoutes = [
  {
    path: '/m',
    component: MobileLayout,
    redirect: '/m/inspiration',
    children: [
      {
        path: 'inspiration',
        name: 'MobileInspiration',
        component: () => import('@/views/mobile/InspirationSearchView.vue'),
        meta: { tab: 'inspiration', title: '找灵感', isMobile: true },
      },
      {
        path: 'inspiration/:id',
        name: 'MobileEncyclopediaDetail',
        component: () => import('@/views/mobile/EncyclopediaDetailView.vue'),
        meta: { title: '百科详情', hideTabBar: true, isMobile: true },
      },
      {
        path: 'kitchen',
        name: 'MobileKitchen',
        component: () => import('@/views/mobile/KitchenListView.vue'),
        meta: { tab: 'kitchen', title: '我的厨房', isMobile: true },
      },
      {
        path: 'kitchen/trash',
        name: 'MobileTrash',
        component: () => import('@/views/mobile/TrashView.vue'),
        meta: { title: '回收站', hideTabBar: true, isMobile: true },
      },
      {
        path: 'kitchen/new',
        name: 'MobileRecipeCreate',
        component: () => import('@/views/mobile/RecipeCreateView.vue'),
        meta: { title: '新增菜品', hideTabBar: true, isMobile: true },
      },
      {
        path: 'kitchen/:id',
        name: 'MobileRecipeDetail',
        component: () => import('@/views/mobile/RecipeDetailView.vue'),
        meta: { title: '菜品详情', hideTabBar: true, isMobile: true },
      },
      {
        path: 'kitchen/:id/edit',
        name: 'MobileRecipeEdit',
        component: () => import('@/views/mobile/RecipeEditView.vue'),
        meta: { title: '编辑菜品', hideTabBar: true, isMobile: true },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes: [...mobileRoutes, ...desktopRoutes],
  scrollBehavior(_to, _from, saved) {
    return saved ?? { top: 0 }
  },
})

router.beforeEach((to, _from, next) => {
  const onMobileRoute = to.path === mobilePrefix || to.path.startsWith(`${mobilePrefix}/`)
  const mobileTarget = stripMobilePrefix(to.path)

  if (isMobileDevice() && !onMobileRoute) {
    next({ path: withMobilePrefix(mobileTarget), query: to.query, hash: to.hash })
    return
  }

  if (!isMobileDevice() && onMobileRoute) {
    next({ path: mobileTarget, query: to.query, hash: to.hash })
    return
  }

  next()
})

router.afterEach((to) => {
  document.title = resolveTitle(to)
})

export default router
