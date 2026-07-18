package model

import (
	"time"

	"gorm.io/gorm"
)

// MyRecipe 我的菜谱主表
type MyRecipe struct {
	ID                   uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID               uint64         `gorm:"not null;index:idx_my_recipes_user_id" json:"user_id"`
	Name                 string         `gorm:"column:recipe_name;size:128;not null" json:"recipe_name"`
	CurrentVersionID     *uint64        `gorm:"index" json:"current_version_id,omitempty"`
	UserRating           *uint8         `gorm:"index:idx_my_recipes_user_rating" json:"user_rating,omitempty"` // 1-5
	CoverImageURL        *string        `gorm:"size:512" json:"cover_image_url,omitempty"`
	EncyclopediaRecipeID *uint64        `gorm:"index:idx_my_recipes_encyclopedia" json:"encyclopedia_recipe_id,omitempty"`
	CreatedAt            time.Time      `gorm:"precision:3;index:idx_my_recipes_created_at" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"precision:3" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index:idx_my_recipes_deleted_at" json:"-"`

	User               User                `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	CurrentVersion     *RecipeVersion      `gorm:"foreignKey:CurrentVersionID;constraint:-" json:"current_version,omitempty"`
	EncyclopediaRecipe *EncyclopediaRecipe `gorm:"foreignKey:EncyclopediaRecipeID;constraint:OnDelete:SET NULL" json:"encyclopedia_recipe,omitempty"`
	Versions           []RecipeVersion     `gorm:"foreignKey:RecipeID" json:"versions,omitempty"`
}

func (MyRecipe) TableName() string { return "my_recipes" }
