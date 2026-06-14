package app

import (
	"log"
	"sync"

	"github.com/delicious/delicious/internal/config"
	"github.com/delicious/delicious/internal/database"
	"github.com/delicious/delicious/internal/handler"
	"github.com/delicious/delicious/internal/repository"
	"github.com/delicious/delicious/internal/service"
	"github.com/gin-gonic/gin"
)

var (
	once   sync.Once
	engine *gin.Engine
)

// Bootstrap 初始化数据库、服务与路由，供本地 main 与 Vercel Serverless 共用。
func Bootstrap() *gin.Engine {
	once.Do(func() {
		gin.SetMode(gin.ReleaseMode)
		cfg := config.Load()

		db, err := database.Connect(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("database: %v", err)
		}

		if cfg.AutoMigrate {
			if err := database.AutoMigrate(db); err != nil {
				log.Fatalf("migrate: %v", err)
			}
			if err := database.Seed(db, cfg.DefaultUID); err != nil {
				log.Fatalf("seed: %v", err)
			}
		}

		uploadSvc := service.NewUploadService(cfg)
		if err := uploadSvc.EnsureDir(); err != nil {
			log.Fatalf("upload dir: %v", err)
		}

		recipeRepo := repository.NewRecipeRepository(db)
		encyRepo := repository.NewEncyclopediaRepository(db)

		recipeSvc := service.NewRecipeService(recipeRepo, encyRepo)
		encySvc := service.NewEncyclopediaService(encyRepo)
		dashSvc := service.NewDashboardService(recipeRepo)

		r := gin.Default()
		handler.Register(r, cfg, handler.Handlers{
			Recipe:       handler.NewRecipeHandler(recipeSvc),
			Encyclopedia: handler.NewEncyclopediaHandler(encySvc),
			Dashboard:    handler.NewDashboardHandler(recipeSvc, dashSvc),
			Upload:       handler.NewUploadHandler(uploadSvc),
		})
		engine = r
	})
	return engine
}
