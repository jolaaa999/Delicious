package app

import (
	"fmt"
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
	once    sync.Once
	engine  *gin.Engine
	initErr error
)

func initApp() {
	once.Do(func() {
		gin.SetMode(gin.ReleaseMode)
		cfg := config.Load()

		db, err := database.Connect(cfg.DatabaseURL)
		if err != nil {
			initErr = fmt.Errorf("database: %w", err)
			return
		}

		if cfg.AutoMigrate {
			if err := database.AutoMigrate(db); err != nil {
				initErr = fmt.Errorf("migrate: %w", err)
				return
			}
			if err := database.Seed(db, cfg.DefaultUID); err != nil {
				initErr = fmt.Errorf("seed: %w", err)
				return
			}
		}

		uploadSvc := service.NewUploadService(cfg)
		if err := uploadSvc.EnsureDir(); err != nil {
			initErr = fmt.Errorf("upload dir: %w", err)
			return
		}

		recipeRepo := repository.NewRecipeRepository(db)
		encyRepo := repository.NewEncyclopediaRepository(db)
		categoryRepo := repository.NewCategoryRepository(db)
		tagRepo := repository.NewTagRepository(db)

		recipeSvc := service.NewRecipeService(recipeRepo, encyRepo)
		encySvc := service.NewEncyclopediaService(encyRepo, tagRepo, cfg)
		dashSvc := service.NewDashboardService(recipeRepo)
		categorySvc := service.NewCategoryService(categoryRepo)
		tagSvc := service.NewTagService(tagRepo)

		r := gin.Default()
		handler.Register(r, cfg, handler.Handlers{
			Recipe:       handler.NewRecipeHandler(recipeSvc),
			Encyclopedia: handler.NewEncyclopediaHandler(encySvc),
			Dashboard:    handler.NewDashboardHandler(recipeSvc, dashSvc),
			Upload:       handler.NewUploadHandler(uploadSvc, recipeRepo),
			Category:     handler.NewCategoryHandler(categorySvc),
			Tag:          handler.NewTagHandler(tagSvc),
		})
		engine = r
	})
}

// Bootstrap 初始化数据库、服务与路由，供本地 main 使用。
func Bootstrap() *gin.Engine {
	initApp()
	if initErr != nil {
		log.Fatal(initErr)
	}
	return engine
}

// Engine 供 Vercel Serverless 使用，初始化失败时返回可读错误而非崩溃。
func Engine() (*gin.Engine, error) {
	initApp()
	return engine, initErr
}
