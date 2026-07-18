package database

import (
	"fmt"
	"log"
	"strings"

	"github.com/delicious/delicious/pkg/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL 未配置：请在 Vercel 项目 Settings → Environment Variables 中添加 Postgres 连接串")
	}
	databaseURL = withConnectTimeout(databaseURL, 5)

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	// 先建被引用表，再建含外键的表；my_recipes ↔ recipe_versions 的 current_version_id 不设 DB 外键，避免循环依赖。
	if err := db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.EncyclopediaRecipe{},
		&model.MyRecipe{},
	); err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.RecipeVersion{}); err != nil {
		return err
	}
	return db.AutoMigrate(&model.EncyclopediaRecipeTag{})
}

func Seed(db *gorm.DB, defaultUID uint64) error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count == 0 {
		nickname := "家人"
		user := model.User{
			ID:       defaultUID,
			Username: "family",
			Nickname: nickname,
			Status:   1,
		}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
		// 同步 serial，避免后续插入 ID 冲突
		db.Exec("SELECT setval(pg_get_serial_sequence('users', 'id'), GREATEST((SELECT MAX(id) FROM users), 1))")
		log.Println("seed: default owner user created (id=1, no login required)")
	}

	// 清理无外部来源的本地百科菜谱（仅保留联网缓存）
	if res := db.Where("external_source IS NULL OR external_source = ''").Delete(&model.EncyclopediaRecipe{}); res.Error != nil {
		return res.Error
	} else if res.RowsAffected > 0 {
		log.Printf("seed: removed %d local-only encyclopedia recipes", res.RowsAffected)
	}

	// Seed 分类字典
	db.Model(&model.Category{}).Count(&count)
	if count == 0 {
		categories := []model.Category{
			{Name: "家常菜"}, {Name: "川菜"}, {Name: "粤菜"}, {Name: "鲁菜"},
			{Name: "苏菜"}, {Name: "湘菜"}, {Name: "西餐"}, {Name: "日料"},
			{Name: "韩餐"}, {Name: "烘焙"}, {Name: "甜点"}, {Name: "快手菜"},
			{Name: "硬菜"}, {Name: "汤羹"}, {Name: "凉菜"},
		}
		if err := db.Create(&categories).Error; err != nil {
			return err
		}
		log.Println("seed: categories created")
	}

	// Seed 标签字典
	db.Model(&model.Tag{}).Count(&count)
	if count == 0 {
		tags := []model.Tag{
			{Name: "快手菜"}, {Name: "下饭菜"}, {Name: "硬菜"}, {Name: "宴客菜"},
			{Name: "清淡"}, {Name: "高蛋白"}, {Name: "低脂"}, {Name: "素食"},
			{Name: "养生"}, {Name: "早餐"}, {Name: "午餐"}, {Name: "晚餐"},
		}
		if err := db.Create(&tags).Error; err != nil {
			return err
		}
		log.Println("seed: tags created")
	}

	return nil
}

func withConnectTimeout(databaseURL string, seconds int) string {
	if strings.Contains(databaseURL, "connect_timeout=") {
		return databaseURL
	}
	sep := "?"
	if strings.Contains(databaseURL, "?") {
		sep = "&"
	}
	return databaseURL + sep + fmt.Sprintf("connect_timeout=%d", seconds)
}
