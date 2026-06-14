package main

import (
	"log"

	"github.com/delicious/delicious/internal/app"
	"github.com/delicious/delicious/internal/config"
)

func main() {
	cfg := config.Load()
	r := app.Bootstrap()

	log.Printf("人间烟火 API listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
