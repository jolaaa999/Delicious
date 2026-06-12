<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteRecipe, listRecipes } from '@/api'
import type { RecipeListItem } from '@/types'

const router = useRouter()
const loading = ref(false)
const items = ref<RecipeListItem[]>([])
const total = ref(0)

const filters = ref({
  keyword: '',
  min_rating: undefined as number | undefined,
  page: 1,
  page_size: 10,
  order_by: 'updated_at',
})

async function fetchList() {
  loading.value = true
  try {
    const res = await listRecipes({
      ...filters.value,
      desc: true,
    })
    items.value = res.items
    total.value = res.page_info.total
  } finally {
    loading.value = false
  }
}

onMounted(fetchList)

function handleSearch() {
  filters.value.page = 1
  fetchList()
}

function goDetail(row: RecipeListItem) {
  router.push(`/recipes/${row.id}`)
}

async function handleDelete(row: RecipeListItem) {
  await ElMessageBox.confirm(`确定删除「${row.name}」？`, '确认删除', { type: 'warning' })
  await deleteRecipe(row.id)
  ElMessage.success('已删除')
  fetchList()
}

function renderStars(rating: number) {
  return '★'.repeat(rating) + '☆'.repeat(Math.max(0, 5 - rating))
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString('zh-CN')
}
</script>

<template>
  <div>
    <div class="page-header">
      <h1>菜谱管理</h1>
      <p>所有录入的菜品，支持筛选与详情穿透</p>
    </div>

    <el-card shadow="never">
      <el-form :inline="true" @submit.prevent="handleSearch">
        <el-form-item label="菜名">
          <el-input v-model="filters.keyword" placeholder="搜索" clearable @clear="handleSearch" />
        </el-form-item>
        <el-form-item label="最低评分">
          <el-select v-model="filters.min_rating" placeholder="全部" clearable style="width: 100px">
            <el-option v-for="n in 5" :key="n" :label="`${n} 星`" :value="n" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="items" stripe @row-click="goDetail">
        <el-table-column label="封面" width="80">
          <template #default="{ row }">
            <el-image
              v-if="row.cover_image_url"
              :src="row.cover_image_url"
              fit="cover"
              style="width: 48px; height: 48px; border-radius: 6px"
            />
            <div v-else class="thumb-placeholder">{{ row.name.charAt(0) }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="菜名" min-width="120" />
        <el-table-column label="评分" width="120">
          <template #default="{ row }">
            <span class="stars">{{ renderStars(row.user_rating || 0) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="current_version_number" label="版本" width="80">
          <template #default="{ row }">v{{ row.current_version_number }}</template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="160">
          <template #default="{ row }">{{ formatDate(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click.stop="goDetail(row)">详情</el-button>
            <el-button link type="danger" @click.stop="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="filters.page"
          :page-size="filters.page_size"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchList"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.thumb-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  background: #fdf0e8;
  color: var(--delicious-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 600;
}

.stars {
  color: #e6a23c;
  letter-spacing: 1px;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-table__row) {
  cursor: pointer;
}
</style>
