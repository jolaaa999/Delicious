<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getDashboardStats } from '@/api'
import type { DashboardStats } from '@/types'

const loading = ref(true)
const stats = ref<DashboardStats | null>(null)

onMounted(async () => {
  try {
    stats.value = await getDashboardStats()
  } finally {
    loading.value = false
  }
})

function formatDate(iso?: string) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString('zh-CN')
}
</script>

<template>
  <div v-loading="loading">
    <div class="page-header">
      <h1>数据总览</h1>
      <p>我的厨房整体情况</p>
    </div>

    <el-row v-if="stats" :gutter="16" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="菜品总数" :value="stats.total_recipes" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="平均评分" :value="stats.average_rating" :precision="1" suffix="/ 5" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="版本总数" :value="stats.total_versions" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-label">最近录入</div>
          <div class="stat-value-sm">{{ formatDate(stats.latest_recipe_at) }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card v-if="stats" class="chart-card" shadow="never">
      <template #header>评分分布</template>
      <div class="rating-bars">
        <div v-for="item in stats.rating_distribution" :key="item.rating" class="rating-bar">
          <span class="rating-bar__label">{{ item.rating }} 星</span>
          <el-progress
            :percentage="stats.total_recipes ? Math.round((item.count / stats.total_recipes) * 100) : 0"
            :stroke-width="16"
            :format="() => String(item.count)"
          />
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.stat-row {
  margin-bottom: 20px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value-sm {
  font-size: 20px;
  font-weight: 600;
}

.chart-card {
  margin-top: 8px;
}

.rating-bars {
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-width: 480px;
}

.rating-bar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.rating-bar__label {
  width: 40px;
  font-size: 13px;
  color: #606266;
}

.rating-bar .el-progress {
  flex: 1;
}
</style>
