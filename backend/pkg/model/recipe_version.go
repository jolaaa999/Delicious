package model

import "time"

// RecipeVersion 菜谱版本详情表（不可变记录）
type RecipeVersion struct {
	ID            uint64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	RecipeID      uint64                 `gorm:"not null;uniqueIndex:uk_recipe_version,priority:1;index:idx_recipe_versions_recipe_id" json:"recipe_id"`
	VersionNumber uint32                 `gorm:"not null;uniqueIndex:uk_recipe_version,priority:2" json:"version_number"`
	Ingredients   JSONSlice[Ingredient]  `gorm:"type:json;not null" json:"ingredients"`
	ProcessSteps  JSONSlice[ProcessStep] `gorm:"type:json;not null" json:"process_steps"`
	ProcessText   *string                `gorm:"type:text" json:"process_text,omitempty"`
	Images        StringSlice            `gorm:"type:json" json:"images,omitempty"`
	CommitMsg     string                 `gorm:"size:255;not null;default:''" json:"commit_msg"`
	CreatedAt     time.Time              `gorm:"precision:3;index:idx_recipe_versions_created_at" json:"created_at"`

	Recipe MyRecipe `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE" json:"recipe,omitempty"`
}

func (RecipeVersion) TableName() string { return "recipe_versions" }
