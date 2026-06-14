package handler

import (
	"github.com/delicious/delicious/internal/config"
	"github.com/delicious/delicious/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Recipe       *RecipeHandler
	Encyclopedia *EncyclopediaHandler
	Dashboard    *DashboardHandler
	Upload       *UploadHandler
}

func Register(r *gin.Engine, cfg config.Config, h Handlers) {
	r.Use(corsMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 本地开发时提供静态文件；Vercel 使用 Blob 存储，不走本地目录
	if cfg.UploadDir != "" {
		r.Static("/uploads", cfg.UploadDir)
	}

	v1 := r.Group("/api/v1")
	v1.Use(middleware.InjectOwner(cfg))

	// Upload
	v1.POST("/upload", h.Upload.Upload)
	v1.POST("/upload/batch", h.Upload.UploadMultiple)

	// Encyclopedia
	v1.GET("/encyclopedia/search", h.Encyclopedia.Search)
	v1.GET("/encyclopedia/category/:category", h.Encyclopedia.ListByCategory)
	v1.GET("/encyclopedia/:id", h.Encyclopedia.Get)

	// Recipes
	v1.POST("/recipes", h.Recipe.Create)
	v1.GET("/recipes", h.Recipe.List)
	v1.GET("/recipes/:id", h.Recipe.Get)
	v1.PUT("/recipes/:id", h.Recipe.Update)
	v1.DELETE("/recipes/:id", h.Recipe.Delete)
	v1.GET("/recipes/:id/versions", h.Recipe.ListVersions)
	v1.GET("/recipes/:id/versions/:version_id", h.Recipe.GetVersion)
	v1.GET("/recipes/:id/diff", h.Recipe.CompareVersions)
	v1.GET("/recipes/:id/diff/encyclopedia", h.Recipe.CompareEncyclopedia)

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
