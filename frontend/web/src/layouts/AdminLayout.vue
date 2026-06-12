<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const activeMenu = computed(() => {
  if (route.path.startsWith('/recipes')) return '/recipes'
  return route.path
})

function navigate(path: string) {
  router.push(path)
}
</script>

<template>
  <el-container class="admin-layout">
    <el-aside width="220px" class="aside">
      <div class="brand">
        <span class="brand__icon">🍳</span>
        <span class="brand__text">人间烟火</span>
      </div>
      <el-menu :default-active="activeMenu" @select="navigate">
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>数据总览</span>
        </el-menu-item>
        <el-menu-item index="/recipes">
          <el-icon><Dish /></el-icon>
          <span>菜谱管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <span class="header__title">{{ route.meta.title }}</span>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin-layout {
  height: 100vh;
}

.aside {
  background: #fff;
  border-right: 1px solid #ebeef5;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.04);
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 20px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.brand__icon {
  font-size: 24px;
}

.brand__text {
  font-size: 18px;
  font-weight: 700;
  color: var(--delicious-primary);
}

.header {
  display: flex;
  align-items: center;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  height: 56px;
}

.header__title {
  font-size: 16px;
  font-weight: 600;
}

.main {
  background: var(--delicious-bg);
  padding: 24px;
}
</style>
