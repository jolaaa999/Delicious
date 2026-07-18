package main

import (
	"log"

	"github.com/delicious/delicious/internal/app"
	"github.com/delicious/delicious/internal/config"
)

// @title           人间烟火 (Delicious) API
// @version         2.0
// @description     家庭菜谱管理系统 — 支持菜谱版本控制、百科搜索、翻译、导入导出
// @description     所有接口前缀 /api/v1，单用户模式（无需认证）
// @contact.name    Delicious Team
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https
// @produce         json
// @consumes        json

func main() {
	cfg := config.Load()
	r := app.Bootstrap()

	log.Printf("人间烟火 API listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
