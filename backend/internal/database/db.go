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
		&model.EncyclopediaRecipe{},
		&model.MyRecipe{},
	); err != nil {
		return err
	}
	return db.AutoMigrate(&model.RecipeVersion{})
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

	db.Model(&model.EncyclopediaRecipe{}).Count(&count)
	if count == 0 {
		recipes := []model.EncyclopediaRecipe{
			{
				Name:        "红烧肉",
				Description: strPtr("经典家常菜，肥而不腻、入口即化"),
				Category:    strPtr("家常菜"),
				Tags:        model.StringSlice{"硬菜", "下饭菜"},
				Ingredients: model.JSONSlice[model.Ingredient]{
					{Name: "五花肉", Amount: 500, Unit: "g", Note: "切块"},
					{Name: "冰糖", Amount: 30, Unit: "g"},
					{Name: "生抽", Amount: 2, Unit: "勺"},
					{Name: "老抽", Amount: 1, Unit: "勺"},
					{Name: "料酒", Amount: 2, Unit: "勺"},
				},
				ProcessSteps: model.JSONSlice[model.ProcessStep]{
					{Order: 1, Content: "五花肉切块，冷水下锅焯水去腥，捞出沥干"},
					{Order: 2, Content: "少油小火炒冰糖至枣红色（炒糖色）"},
					{Order: 3, Content: "下肉块翻炒上色，加生抽、老抽、料酒"},
					{Order: 4, Content: "加热水没过食材，大火烧开转小火炖 45 分钟"},
					{Order: 5, Content: "大火收汁至浓稠，出锅装盘"},
				},
				Source: strPtr("百科"),
			},
			{
				Name:        "番茄炒蛋",
				Description: strPtr("国民家常菜，酸甜开胃"),
				Category:    strPtr("家常菜"),
				Tags:        model.StringSlice{"快手菜"},
				Ingredients: model.JSONSlice[model.Ingredient]{
					{Name: "番茄", Amount: 2, Unit: "个"},
					{Name: "鸡蛋", Amount: 3, Unit: "个"},
					{Name: "盐", Amount: 3, Unit: "g"},
					{Name: "糖", Amount: 5, Unit: "g"},
				},
				ProcessSteps: model.JSONSlice[model.ProcessStep]{
					{Order: 1, Content: "鸡蛋打散，加 pinch 盐，热油滑炒至半熟盛出"},
					{Order: 2, Content: "番茄切块，下锅炒出汁水"},
					{Order: 3, Content: "倒回鸡蛋，加盐和糖快速翻炒均匀"},
				},
				Source: strPtr("百科"),
			},
			{
				Name:        "清蒸鲈鱼",
				Description: strPtr("粤式经典，保留食材本味"),
				Category:    strPtr("粤菜"),
				Tags:        model.StringSlice{"清淡", "宴客"},
				Ingredients: model.JSONSlice[model.Ingredient]{
					{Name: "鲈鱼", Amount: 1, Unit: "条", Note: "约500g"},
					{Name: "姜", Amount: 20, Unit: "g"},
					{Name: "葱", Amount: 2, Unit: "根"},
					{Name: "蒸鱼豉油", Amount: 2, Unit: "勺"},
				},
				ProcessSteps: model.JSONSlice[model.ProcessStep]{
					{Order: 1, Content: "鱼处理干净，两面划刀，抹少许料酒"},
					{Order: 2, Content: "铺姜丝，水开后大火蒸 8-10 分钟"},
					{Order: 3, Content: "倒掉蒸鱼水，铺葱丝，淋蒸鱼豉油，泼热油"},
				},
				Source: strPtr("百科"),
			},
		}
		if err := db.Create(&recipes).Error; err != nil {
			return err
		}
		log.Println("seed: encyclopedia recipes created")
	}
	return nil
}

func strPtr(s string) *string { return &s }

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
