package model

import "time"

// Tag 标签字典表
type Tag struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:64;not null;uniqueIndex" json:"name"`
	CreatedAt time.Time `gorm:"precision:3" json:"created_at"`
}

func (Tag) TableName() string { return "tags" }

// EncyclopediaRecipeTag 百科菜谱-标签关联表
type EncyclopediaRecipeTag struct {
	ID       uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	RecipeID uint64              `gorm:"not null;uniqueIndex:uk_recipe_tag,priority:1;index;constraint:OnDelete:CASCADE" json:"recipe_id"`
	TagID    uint64              `gorm:"not null;uniqueIndex:uk_recipe_tag,priority:2;index;constraint:OnDelete:CASCADE" json:"tag_id"`
	Recipe   EncyclopediaRecipe  `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"-"`
	Tag      Tag                 `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE" json:"-"`
}

func (EncyclopediaRecipeTag) TableName() string { return "encyclopedia_recipe_tags" }
