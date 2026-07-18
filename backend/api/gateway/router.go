package gateway

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route 描述一条 REST → gRPC 映射
type Route struct {
	Method      string
	Path        string
	GRPCService string
	GRPCMethod  string
}

// AllRoutes 网关路由注册表（与 routes.yaml 保持同步）
var AllRoutes = []Route{
	// Encyclopedia
	{http.MethodGet, "/encyclopedia/search", "delicious.v1.EncyclopediaService", "SearchRecipes"},
	{http.MethodGet, "/encyclopedia/:id", "delicious.v1.EncyclopediaService", "GetRecipe"},
	{http.MethodGet, "/encyclopedia/category/:category", "delicious.v1.EncyclopediaService", "ListByCategory"},

	// Categories
	{http.MethodGet, "/categories", "delicious.v1.CategoryService", "ListCategories"},
	{http.MethodPost, "/categories", "delicious.v1.CategoryService", "CreateCategory"},
	{http.MethodPut, "/categories/:id", "delicious.v1.CategoryService", "UpdateCategory"},
	{http.MethodDelete, "/categories/:id", "delicious.v1.CategoryService", "DeleteCategory"},

	// Tags
	{http.MethodGet, "/tags", "delicious.v1.TagService", "ListTags"},
	{http.MethodPost, "/tags", "delicious.v1.TagService", "CreateTag"},
	{http.MethodDelete, "/tags/:id", "delicious.v1.TagService", "DeleteTag"},
	{http.MethodGet, "/encyclopedia/:id/tags", "delicious.v1.TagService", "ListRecipeTags"},
	{http.MethodPost, "/encyclopedia/:id/tags", "delicious.v1.TagService", "AddRecipeTag"},
	{http.MethodDelete, "/encyclopedia/:id/tags/:tag_id", "delicious.v1.TagService", "RemoveRecipeTag"},

	// Recipe CRUD
	{http.MethodPost, "/recipes", "delicious.v1.RecipeService", "CreateRecipe"},
	{http.MethodGet, "/recipes", "delicious.v1.RecipeService", "ListRecipes"},
	{http.MethodGet, "/recipes/trash", "delicious.v1.RecipeService", "ListTrash"},
	{http.MethodGet, "/recipes/export", "delicious.v1.RecipeService", "ExportRecipes"},
	{http.MethodPost, "/recipes/import", "delicious.v1.RecipeService", "ImportRecipes"},
	{http.MethodGet, "/recipes/:id", "delicious.v1.RecipeService", "GetRecipe"},
	{http.MethodPut, "/recipes/:id", "delicious.v1.RecipeService", "UpdateRecipe"},
	{http.MethodDelete, "/recipes/:id", "delicious.v1.RecipeService", "DeleteRecipe"},
	{http.MethodPost, "/recipes/:id/restore", "delicious.v1.RecipeService", "RestoreRecipe"},
	{http.MethodDelete, "/recipes/:id/permanent", "delicious.v1.RecipeService", "PermanentDelete"},

	// Version & Diff
	{http.MethodGet, "/recipes/:id/versions", "delicious.v1.RecipeService", "ListVersions"},
	{http.MethodGet, "/recipes/:id/versions/:version_id", "delicious.v1.RecipeService", "GetVersion"},
	{http.MethodGet, "/recipes/:id/diff", "delicious.v1.RecipeService", "CompareVersions"},
	{http.MethodGet, "/recipes/:id/diff/encyclopedia", "delicious.v1.RecipeService", "CompareWithEncyclopedia"},
	{http.MethodGet, "/recipes/:id/compare-encyclopedia", "delicious.v1.RecipeService", "CompareWithEncyclopedia"},

	// Upload
	{http.MethodPost, "/upload", "", "Upload"},
	{http.MethodPost, "/upload/image", "", "Upload"},
	{http.MethodPost, "/upload/batch", "", "UploadMultiple"},

	// Dashboard
	{http.MethodGet, "/dashboard/stats", "delicious.v1.DashboardService", "GetStats"},
	{http.MethodGet, "/dashboard/recipes/:id/timeline", "delicious.v1.DashboardService", "GetRecipeTimeline"},
}

// ServiceEndpoints 内网 gRPC 服务发现地址
var ServiceEndpoints = map[string]string{
	"delicious.v1.RecipeService":       "recipe-service:50051",
	"delicious.v1.EncyclopediaService": "encyclopedia-service:50052",
	"delicious.v1.DashboardService":    "dashboard-service:50053",
}

// Register 将路由挂到 Gin 引擎（handler 由 grpc-gateway 或自定义代理实现）
func Register(r *gin.Engine, proxy Handler) {
	v1 := r.Group("/api/v1")
	v1.Use(JWTAuthMiddleware())

	for _, route := range AllRoutes {
		v1.Handle(route.Method, route.Path, proxy.Handle(route))
	}
}

// Handler REST 请求代理到 gRPC 的接口
type Handler interface {
	Handle(route Route) gin.HandlerFunc
}

// JWTAuthMiddleware 从 Authorization: Bearer <token> 解析 user_id
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现 JWT 校验，c.Set("user_id", uid)
		c.Next()
	}
}
