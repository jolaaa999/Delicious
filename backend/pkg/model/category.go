package model

import "time"

// Category 菜谱分类字典表
type Category struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:64;not null;uniqueIndex" json:"name"`
	CreatedAt time.Time `gorm:"precision:3" json:"created_at"`
}

func (Category) TableName() string { return "categories" }
