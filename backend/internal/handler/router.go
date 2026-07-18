package handler

import (
	"github.com/delicious/delicious/internal/config"
	"github.com/delicious/delicious/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/delicious/delicious/docs" // swag init 生成的 docs 包
)

type Handlers struct {
	Recipe       *RecipeHandler
	Encyclopedia *EncyclopediaHandler
	Dashboard    *DashboardHandler
	Upload       *UploadHandler
	Category     *CategoryHandler
	Tag          *TagHandler
}

func Register(r *gin.Engine, cfg config.Config, h Handlers) {
	r.Use(corsMiddleware())

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 本地开发时提供静态文件；Vercel 使用 Blob 存储，不走本地目录
	if cfg.UploadDir != "" {
		r.Static("/uploads", cfg.UploadDir)
	}

	v1 := r.Group("/api/v1")
	v1.Use(middleware.InjectOwner(cfg))

	// Upload
	v1.POST("/upload", h.Upload.Upload)
	v1.POST("/upload/image", h.Upload.Upload) // 兼容旧版前端路径
	v1.POST("/upload/batch", h.Upload.UploadMultiple)
	v1.GET("/media", h.Upload.ProxyMedia)

	// Encyclopedia
	v1.GET("/encyclopedia/search", h.Encyclopedia.Search)
	v1.GET("/encyclopedia/category/:category", h.Encyclopedia.ListByCategory)
	v1.GET("/encyclopedia/:id", h.Encyclopedia.Get)

	// Categories
	v1.GET("/categories", h.Category.List)
	v1.POST("/categories", h.Category.Create)
	v1.PUT("/categories/:id", h.Category.Update)
	v1.DELETE("/categories/:id", h.Category.Delete)

	// Tags
	v1.GET("/tags", h.Tag.List)
	v1.POST("/tags", h.Tag.Create)
	v1.DELETE("/tags/:id", h.Tag.Delete)
	v1.GET("/encyclopedia/:id/tags", h.Tag.ListByRecipe)
	v1.POST("/encyclopedia/:id/tags", h.Tag.AddToRecipe)
	v1.DELETE("/encyclopedia/:id/tags/:tag_id", h.Tag.RemoveFromRecipe)

	// Recipes — 无参数路由放在 :id 前避免匹配冲突
	v1.POST("/recipes", h.Recipe.Create)
	v1.GET("/recipes", h.Recipe.List)
	v1.GET("/recipes/trash", h.Recipe.ListTrash)
	v1.GET("/recipes/export", h.Recipe.Export)
	v1.POST("/recipes/import", h.Recipe.Import)
	v1.GET("/recipes/:id", h.Recipe.Get)
	v1.PUT("/recipes/:id", h.Recipe.Update)
	v1.DELETE("/recipes/:id", h.Recipe.Delete)
	v1.POST("/recipes/:id/restore", h.Recipe.Restore)
	v1.DELETE("/recipes/:id/permanent", h.Recipe.PermanentDelete)
	v1.GET("/recipes/:id/versions", h.Recipe.ListVersions)
	v1.GET("/recipes/:id/versions/:version_id", h.Recipe.GetVersion)
	v1.GET("/recipes/:id/diff", h.Recipe.CompareVersions)
	v1.GET("/recipes/:id/diff/encyclopedia", h.Recipe.CompareEncyclopedia)
	v1.GET("/recipes/:id/compare-encyclopedia", h.Recipe.CompareEncyclopedia) // 兼容旧版

	// Admin — 图片清理
	v1.POST("/admin/cleanup-images", h.Upload.CleanupScan)
	v1.POST("/admin/cleanup-images/execute", h.Upload.CleanupExecute)

	// Dashboard
	v1.GET("/dashboard/stats", h.Dashboard.Stats)
	v1.GET("/dashboard/recipes/:id/timeline", h.Dashboard.Timeline)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
