package model

import "time"

// EncyclopediaRecipe 百科菜谱表
type EncyclopediaRecipe struct {
	ID            uint64                   `gorm:"primaryKey;autoIncrement" json:"id"`
	EncyclopediaRecipesID   *string                  `gorm:"size:128" json:"encyclopedia_recipes_id,omitempty"`
	Name                    string                   `gorm:"column:encyclopedia_recipes_name;size:128;not null;index:idx_encyclopedia_name" json:"encyclopedia_recipes_name"`
	Description   *string                  `gorm:"type:text" json:"description,omitempty"`
	CoverImageURL *string                  `gorm:"size:512" json:"cover_image_url,omitempty"`
	Category      *string                  `gorm:"size:64;index:idx_encyclopedia_category" json:"category,omitempty"`
	Tags          StringSlice              `gorm:"type:json" json:"tags,omitempty"`
	Ingredients   JSONSlice[Ingredient]    `gorm:"type:json;not null" json:"ingredients"`
	ProcessSteps  JSONSlice[ProcessStep]   `gorm:"type:json;not null" json:"process_steps"`
	Source        *string                  `gorm:"size:128" json:"source,omitempty"`
	ExternalSource *string                 `gorm:"size:32;uniqueIndex:idx_encyclopedia_external" json:"external_source,omitempty"`
	ExternalID     *string                 `gorm:"size:64;uniqueIndex:idx_encyclopedia_external" json:"external_id,omitempty"`
	ViewCount     uint32                   `gorm:"not null;default:0" json:"view_count"`
	CreatedAt     time.Time                `gorm:"precision:3" json:"created_at"`
	UpdatedAt     time.Time                `gorm:"precision:3" json:"updated_at"`
}

func (EncyclopediaRecipe) TableName() string { return "encyclopedia_recipes" }
