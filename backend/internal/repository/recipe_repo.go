package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/delicious/delicious/pkg/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type RecipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

type ListRecipesFilter struct {
	UserID        uint64
	Page          int
	PageSize      int
	MinRating     *uint8
	MaxRating     *uint8
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Keyword       string
	OrderBy       string
	Desc          bool
}

func (r *RecipeRepository) CreateWithVersion(recipe *model.MyRecipe, version *model.RecipeVersion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(recipe).Error; err != nil {
			return err
		}
		version.RecipeID = recipe.ID
		version.VersionNumber = 1
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		cover := pickCover(version.Images, recipe.CoverImageURL)
		return tx.Model(recipe).Updates(map[string]interface{}{
			"current_version_id": version.ID,
			"cover_image_url":    cover,
		}).Error
	})
}

func (r *RecipeRepository) AddVersion(recipe *model.MyRecipe, version *model.RecipeVersion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var maxVer uint32
		if err := tx.Model(&model.RecipeVersion{}).
			Where("recipe_id = ?", recipe.ID).
			Select("COALESCE(MAX(version_number), 0)").
			Scan(&maxVer).Error; err != nil {
			return err
		}
		version.RecipeID = recipe.ID
		version.VersionNumber = maxVer + 1
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		cover := pickCover(version.Images, recipe.CoverImageURL)
		updates := map[string]interface{}{
			"name":                 recipe.Name,
			"current_version_id":   version.ID,
			"cover_image_url":      cover,
			"encyclopedia_recipe_id": recipe.EncyclopediaRecipeID,
		}
		if recipe.UserRating != nil {
			updates["user_rating"] = *recipe.UserRating
		}
		return tx.Model(recipe).Updates(updates).Error
	})
}

func pickCover(images model.StringSlice, existing *string) *string {
	if len(images) > 0 {
		s := images[0]
		return &s
	}
	return existing
}

func (r *RecipeRepository) GetByID(id, userID uint64) (*model.MyRecipe, error) {
	var recipe model.MyRecipe
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&recipe).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &recipe, err
}

func (r *RecipeRepository) GetVersion(id uint64) (*model.RecipeVersion, error) {
	var ver model.RecipeVersion
	err := r.db.First(&ver, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &ver, err
}

func (r *RecipeRepository) GetVersionByRecipe(recipeID, versionID, userID uint64) (*model.RecipeVersion, error) {
	if _, err := r.GetByID(recipeID, userID); err != nil {
		return nil, err
	}
	var ver model.RecipeVersion
	err := r.db.Where("id = ? AND recipe_id = ?", versionID, recipeID).First(&ver).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &ver, err
}

func (r *RecipeRepository) ListVersions(recipeID, userID uint64) ([]model.RecipeVersion, error) {
	if _, err := r.GetByID(recipeID, userID); err != nil {
		return nil, err
	}
	var versions []model.RecipeVersion
	err := r.db.Where("recipe_id = ?", recipeID).
		Order("version_number DESC").
		Find(&versions).Error
	return versions, err
}

func (r *RecipeRepository) List(filter ListRecipesFilter) ([]model.MyRecipe, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	q := r.db.Model(&model.MyRecipe{}).Where("user_id = ?", filter.UserID)
	if filter.MinRating != nil {
		q = q.Where("user_rating >= ?", *filter.MinRating)
	}
	if filter.MaxRating != nil {
		q = q.Where("user_rating <= ?", *filter.MaxRating)
	}
	if filter.CreatedAfter != nil {
		q = q.Where("created_at >= ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		q = q.Where("created_at <= ?", *filter.CreatedBefore)
	}
	if filter.Keyword != "" {
		q = q.Where("name LIKE ?", fmt.Sprintf("%%%s%%", filter.Keyword))
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderCol := "updated_at"
	switch filter.OrderBy {
	case "created_at", "updated_at", "user_rating":
		orderCol = filter.OrderBy
	}
	order := orderCol + " DESC"
	if !filter.Desc {
		order = orderCol + " ASC"
	}

	var items []model.MyRecipe
	err := q.Order(order).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *RecipeRepository) SoftDelete(id, userID uint64) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.MyRecipe{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *RecipeRepository) CountVersionsByUser(userID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.RecipeVersion{}).
		Joins("JOIN my_recipes ON my_recipes.id = recipe_versions.recipe_id").
		Where("my_recipes.user_id = ? AND my_recipes.deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (r *RecipeRepository) RatingDistribution(userID uint64) (map[uint8]int64, error) {
	type row struct {
		Rating uint8
		Count  int64
	}
	var rows []row
	err := r.db.Model(&model.MyRecipe{}).
		Select("user_rating as rating, COUNT(*) as count").
		Where("user_id = ? AND user_rating IS NOT NULL", userID).
		Group("user_rating").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint8]int64)
	for _, row := range rows {
		m[row.Rating] = row.Count
	}
	return m, nil
}

func (r *RecipeRepository) AverageRating(userID uint64) (float64, error) {
	var avg *float64
	err := r.db.Model(&model.MyRecipe{}).
		Where("user_id = ? AND user_rating IS NOT NULL", userID).
		Select("AVG(user_rating)").
		Scan(&avg).Error
	if err != nil || avg == nil {
		return 0, err
	}
	return *avg, nil
}

func (r *RecipeRepository) LatestRecipeAt(userID uint64) (*time.Time, error) {
	var t time.Time
	err := r.db.Model(&model.MyRecipe{}).
		Where("user_id = ?", userID).
		Select("MAX(created_at)").
		Scan(&t).Error
	if err != nil || t.IsZero() {
		return nil, err
	}
	return &t, nil
}

func (r *RecipeRepository) CountByUser(userID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.MyRecipe{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
