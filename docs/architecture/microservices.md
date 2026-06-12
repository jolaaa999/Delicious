# 微服务架构概览

```
                    ┌─────────────────┐
   H5 / Web PC ───► │   API Gateway   │ :8080  REST /api/v1
                    └────────┬────────┘
                             │ gRPC
         ┌───────────────────┼───────────────────┐
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ recipe-service  │ │ encyclopedia-   │ │ dashboard-      │
│     :50051      │ │ service :50052  │ │ service :50053  │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             ▼
                    ┌─────────────────┐
                    │  MySQL + Redis  │
                    └─────────────────┘
```

## 服务职责

| 服务 | Proto | 职责 |
|------|-------|------|
| recipe-service | `recipe.proto` | 我的菜谱 CRUD、版本写入、Diff |
| encyclopedia-service | `encyclopedia.proto` | 百科搜索、详情（Redis 缓存热点数据） |
| dashboard-service | `dashboard.proto` | PC 看板统计、版本时间轴 |
| api-gateway | `routes.yaml` | JWT 鉴权、REST 转 gRPC |

## Proto 生成

```bash
cd backend/api/proto
buf generate
```

输出目录：`backend/api/gen/delicious/v1/`

## 核心 RPC 一览

### RecipeService
- `CreateRecipe` / `GetRecipe` / `ListRecipes` / `UpdateRecipe` / `DeleteRecipe`
- `ListVersions` / `GetVersion`
- `CompareVersions` — 历史版本 Diff
- `CompareWithEncyclopedia` — 基准对比

### EncyclopediaService
- `SearchRecipes` — 关键词搜索
- `GetRecipe` — 百科详情
- `ListByCategory` — 分类浏览

### DashboardService
- `GetStats` — 菜品总数、平均评分
- `GetRecipeTimeline` — 版本时间轴（PC 详情穿透）
