package dto

import (
	"time"

	"github.com/delicious/delicious/pkg/model"
)

type Ingredient = model.Ingredient
type ProcessStep = model.ProcessStep

type PageInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type RecipeVersionDTO struct {
	ID            uint64        `json:"id"`
	RecipeID      uint64        `json:"recipe_id"`
	VersionNumber uint32        `json:"version_number"`
	Ingredients   []Ingredient  `json:"ingredients"`
	ProcessSteps  []ProcessStep `json:"process_steps"`
	ProcessText   *string       `json:"process_text,omitempty"`
	Images        []string      `json:"images,omitempty"`
	CommitMsg     string        `json:"commit_msg"`
	CreatedAt     time.Time     `json:"created_at"`
}

type MyRecipeDTO struct {
	ID                   uint64            `json:"id"`
	UserID               uint64            `json:"user_id"`
	Name                 string            `json:"name"`
	CurrentVersionID     uint64            `json:"current_version_id"`
	UserRating           uint8             `json:"user_rating"`
	CoverImageURL        *string           `json:"cover_image_url,omitempty"`
	EncyclopediaRecipeID *uint64           `json:"encyclopedia_recipe_id,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	CurrentVersion       *RecipeVersionDTO `json:"current_version,omitempty"`
}

type RecipeListItemDTO struct {
	ID                    uint64    `json:"id"`
	Name                  string    `json:"name"`
	CoverImageURL         *string   `json:"cover_image_url,omitempty"`
	UserRating            uint8     `json:"user_rating"`
	CurrentVersionNumber  uint32    `json:"current_version_number"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type VersionListItemDTO struct {
	ID            uint64    `json:"id"`
	VersionNumber uint32    `json:"version_number"`
	CommitMsg     string    `json:"commit_msg"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateRecipeRequest struct {
	Name                 string        `json:"name" binding:"required"`
	Ingredients          []Ingredient  `json:"ingredients" binding:"required"`
	ProcessSteps         []ProcessStep `json:"process_steps" binding:"required"`
	ProcessText          *string       `json:"process_text"`
	Images               []string      `json:"images"`
	UserRating           *uint8        `json:"user_rating"`
	CommitMsg            string        `json:"commit_msg"`
	EncyclopediaRecipeID *uint64       `json:"encyclopedia_recipe_id"`
}

type UpdateRecipeRequest struct {
	Name                 string        `json:"name" binding:"required"`
	Ingredients          []Ingredient  `json:"ingredients" binding:"required"`
	ProcessSteps         []ProcessStep `json:"process_steps" binding:"required"`
	ProcessText          *string       `json:"process_text"`
	Images               []string      `json:"images"`
	UserRating           *uint8        `json:"user_rating"`
	CommitMsg            string        `json:"commit_msg" binding:"required"`
	EncyclopediaRecipeID *uint64       `json:"encyclopedia_recipe_id"`
}

type IngredientDiffDTO struct {
	Type        string      `json:"type"`
	Base        *Ingredient `json:"base,omitempty"`
	Target      *Ingredient `json:"target,omitempty"`
	AmountDelta *float64    `json:"amount_delta,omitempty"`
}

type ProcessStepDiffDTO struct {
	Type   string       `json:"type"`
	Order  int          `json:"order"`
	Base   *ProcessStep `json:"base,omitempty"`
	Target *ProcessStep `json:"target,omitempty"`
}

type VersionDiffResultDTO struct {
	IngredientDiffs []IngredientDiffDTO  `json:"ingredient_diffs"`
	ProcessDiffs    []ProcessStepDiffDTO `json:"process_diffs"`
	Summary         string               `json:"summary"`
}

type TimelineNodeDTO struct {
	VersionID     uint64    `json:"version_id"`
	VersionNumber uint32    `json:"version_number"`
	CommitMsg     string    `json:"commit_msg"`
	CreatedAt     time.Time `json:"created_at"`
	IsCurrent     bool      `json:"is_current"`
}

type RatingDistributionDTO struct {
	Rating uint8 `json:"rating"`
	Count  int64 `json:"count"`
}

type DashboardStatsDTO struct {
	TotalRecipes        int64                   `json:"total_recipes"`
	AverageRating       float64                 `json:"average_rating"`
	TotalVersions       int64                   `json:"total_versions"`
	RatingDistribution  []RatingDistributionDTO `json:"rating_distribution"`
	LatestRecipeAt      *time.Time              `json:"latest_recipe_at,omitempty"`
}

type EncyclopediaListItemDTO struct {
	ID            uint64   `json:"id"`
	Name          string   `json:"name"`
	CoverImageURL *string  `json:"cover_image_url,omitempty"`
	Category      *string  `json:"category,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Description   *string  `json:"description,omitempty"`
}

type EncyclopediaRecipeDTO struct {
	ID            uint64        `json:"id"`
	Name          string        `json:"name"`
	Description   *string       `json:"description,omitempty"`
	CoverImageURL *string       `json:"cover_image_url,omitempty"`
	Category      *string       `json:"category,omitempty"`
	Tags          []string      `json:"tags,omitempty"`
	Ingredients   []Ingredient  `json:"ingredients"`
	ProcessSteps  []ProcessStep `json:"process_steps"`
	Source        *string       `json:"source,omitempty"`
	ViewCount     uint32        `json:"view_count"`
	CreatedAt     time.Time     `json:"created_at"`
}

func ToVersionDTO(v *model.RecipeVersion) RecipeVersionDTO {
	imgs := []string(v.Images)
	if imgs == nil {
		imgs = []string{}
	}
	return RecipeVersionDTO{
		ID:            v.ID,
		RecipeID:      v.RecipeID,
		VersionNumber: v.VersionNumber,
		Ingredients:   []Ingredient(v.Ingredients),
		ProcessSteps:  []ProcessStep(v.ProcessSteps),
		ProcessText:   v.ProcessText,
		Images:        imgs,
		CommitMsg:     v.CommitMsg,
		CreatedAt:     v.CreatedAt,
	}
}

func ToRecipeDTO(r *model.MyRecipe, ver *model.RecipeVersion) MyRecipeDTO {
	dto := MyRecipeDTO{
		ID:                   r.ID,
		UserID:               r.UserID,
		Name:                 r.Name,
		CoverImageURL:        r.CoverImageURL,
		EncyclopediaRecipeID: r.EncyclopediaRecipeID,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
	if r.CurrentVersionID != nil {
		dto.CurrentVersionID = *r.CurrentVersionID
	}
	if r.UserRating != nil {
		dto.UserRating = *r.UserRating
	}
	if ver != nil {
		v := ToVersionDTO(ver)
		dto.CurrentVersion = &v
	}
	return dto
}

// CategoryDTO 分类字典
type CategoryDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// TagDTO 标签字典
type TagDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// ExportRecipeDTO 导出菜谱结构
type ExportRecipeDTO struct {
	Name                 string        `json:"name"`
	UserRating           *uint8        `json:"user_rating,omitempty"`
	EncyclopediaRecipeID *uint64       `json:"encyclopedia_recipe_id,omitempty"`
	Ingredients          []Ingredient  `json:"ingredients"`
	ProcessSteps         []ProcessStep `json:"process_steps"`
	ProcessText          *string       `json:"process_text,omitempty"`
	Images               []string      `json:"images,omitempty"`
	CommitMsg            string        `json:"commit_msg,omitempty"`
}

// ImportResultDTO 导入结果
type ImportResultDTO struct {
	Total   int      `json:"total"`
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}
