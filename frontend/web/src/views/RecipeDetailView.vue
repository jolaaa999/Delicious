<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { compareVersions, getRecipeTimeline } from '@/api'
import type { MyRecipe, TimelineNode, VersionDiffResult } from '@/types'

const route = useRoute()
const router = useRouter()
const id = Number(route.params.id)

const loading = ref(true)
const recipe = ref<MyRecipe | null>(null)
const timeline = ref<TimelineNode[]>([])
const diffResult = ref<VersionDiffResult | null>(null)
const diffLoading = ref(false)
const compareBaseId = ref<number>()

onMounted(async () => {
  try {
    const res = await getRecipeTimeline(id)
    recipe.value = res.recipe
    timeline.value = res.timeline
  } finally {
    loading.value = false
  }
})

async function handleCompare(node: TimelineNode) {
  if (!recipe.value?.current_version_id || node.is_current) {
    diffResult.value = null
    return
  }
  compareBaseId.value = node.version_id
  diffLoading.value = true
  try {
    const res = await compareVersions(id, node.version_id, recipe.value.current_version_id)
    diffResult.value = res.diff
  } finally {
    diffLoading.value = false
  }
}

function diffTagType(type: string) {
  if (type === 'added') return 'success'
  if (type === 'removed') return 'danger'
  if (type === 'modified') return 'warning'
  return 'info'
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString('zh-CN')
}
</script>

<template>
  <div v-loading="loading">
    <el-page-header @back="router.push('/recipes')">
      <template #content>
        <span class="detail-title">{{ recipe?.name ?? '菜谱详情' }}</span>
      </template>
    </el-page-header>

    <el-row v-if="recipe" :gutter="20" class="content-row">
      <el-col :span="10">
        <el-card shadow="never">
          <template #header>版本时间轴</template>
          <el-timeline>
            <el-timeline-item
              v-for="node in timeline"
              :key="node.version_id"
              :type="node.is_current ? 'primary' : 'info'"
              :timestamp="formatDate(node.created_at)"
              :hollow="!node.is_current"
            >
              <div class="timeline-node" @click="handleCompare(node)">
                <strong>v{{ node.version_number }}</strong>
                <el-tag v-if="node.is_current" size="small" type="success">当前</el-tag>
                <p class="timeline-msg">{{ node.commit_msg || '—' }}</p>
                <el-button
                  v-if="!node.is_current"
                  link
                  type="primary"
                  size="small"
                  @click.stop="handleCompare(node)"
                >
                  与当前对比
                </el-button>
              </div>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>

      <el-col :span="14">
        <el-card v-if="recipe.current_version" shadow="never" class="detail-card">
          <template #header>当前版本 v{{ recipe.current_version.version_number }}</template>

          <h4>配料</h4>
          <el-table :data="recipe.current_version.ingredients" size="small" class="mini-table">
            <el-table-column prop="name" label="食材" />
            <el-table-column label="用量" width="120">
              <template #default="{ row }">{{ row.amount }}{{ row.unit }}</template>
            </el-table-column>
          </el-table>

          <h4>步骤</h4>
          <el-steps direction="vertical" :active="recipe.current_version.process_steps.length">
            <el-step
              v-for="step in recipe.current_version.process_steps"
              :key="step.order"
              :title="`步骤 ${step.order}`"
              :description="step.content"
            />
          </el-steps>
        </el-card>

        <el-card v-loading="diffLoading" shadow="never" class="diff-card">
          <template #header>
            版本对比
            <span v-if="compareBaseId" class="diff-hint">（历史 vs 当前）</span>
          </template>

          <el-empty v-if="!diffResult" description="点击时间轴节点查看差异" />

          <template v-else>
            <el-alert :title="diffResult.summary" type="info" show-icon :closable="false" class="diff-summary" />

            <h4>配料差异</h4>
            <el-table :data="diffResult.ingredient_diffs.filter(d => d.type !== 'unchanged')" size="small">
              <el-table-column label="类型" width="80">
                <template #default="{ row }">
                  <el-tag :type="diffTagType(row.type)" size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="基准">
                <template #default="{ row }">
                  {{ row.base ? `${row.base.name} ${row.base.amount}${row.base.unit}` : '—' }}
                </template>
              </el-table-column>
              <el-table-column label="当前">
                <template #default="{ row }">
                  {{ row.target ? `${row.target.name} ${row.target.amount}${row.target.unit}` : '—' }}
                  <el-tag v-if="row.amount_delta" size="small" type="warning" class="delta">
                    {{ row.amount_delta > 0 ? '+' : '' }}{{ row.amount_delta }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>

            <h4>步骤差异</h4>
            <el-table :data="diffResult.process_diffs.filter(d => d.type !== 'unchanged')" size="small">
              <el-table-column label="类型" width="80">
                <template #default="{ row }">
                  <el-tag :type="diffTagType(row.type)" size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="order" label="#" width="50" />
              <el-table-column label="基准" prop="base.content" />
              <el-table-column label="当前" prop="target.content" />
            </el-table>
          </template>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.detail-title {
  font-size: 18px;
  font-weight: 600;
}

.content-row {
  margin-top: 20px;
}

.timeline-node {
  cursor: pointer;
}

.timeline-msg {
  margin: 4px 0;
  color: #606266;
  font-size: 13px;
}

.detail-card h4,
.diff-card h4 {
  margin: 16px 0 8px;
  font-size: 14px;
  color: #303133;
}

.mini-table {
  margin-bottom: 16px;
}

.diff-card {
  margin-top: 16px;
}

.diff-hint {
  font-size: 13px;
  color: #909399;
  font-weight: normal;
}

.diff-summary {
  margin-bottom: 16px;
}

.delta {
  margin-left: 6px;
}
</style>
