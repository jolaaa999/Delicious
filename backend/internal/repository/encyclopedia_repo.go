package repository

import (
	"errors"
	"fmt"

	"github.com/delicious/delicious/pkg/model"
	"gorm.io/gorm"
)

type EncyclopediaRepository struct {
	db *gorm.DB
}

func NewEncyclopediaRepository(db *gorm.DB) *EncyclopediaRepository {
	return &EncyclopediaRepository{db: db}
}

type SearchFilter struct {
	Keyword  string
	Category string
	Page     int
	PageSize int
}

func (r *EncyclopediaRepository) Search(f SearchFilter) ([]model.EncyclopediaRecipe, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}

	q := r.db.Model(&model.EncyclopediaRecipe{})
	if f.Keyword != "" {
		like := fmt.Sprintf("%%%s%%", f.Keyword)
		q = q.Where("name LIKE ? OR description LIKE ?", like, like)
	}
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.EncyclopediaRecipe
	err := q.Order("view_count DESC, id ASC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *EncyclopediaRepository) GetByID(id uint64) (*model.EncyclopediaRecipe, error) {
	var item model.EncyclopediaRecipe
	err := r.db.First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err == nil {
		r.db.Model(&item).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
	}
	return &item, err
}

func (r *EncyclopediaRepository) FindByName(name string) (*model.EncyclopediaRecipe, error) {
	var item model.EncyclopediaRecipe
	err := r.db.Where("name = ?", name).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &item, err
}

func (r *EncyclopediaRepository) ListByCategory(category string, page, pageSize int) ([]model.EncyclopediaRecipe, int64, error) {
	return r.Search(SearchFilter{Category: category, Page: page, PageSize: pageSize})
}

type CachedOnlineHit struct {
	ID     uint64
	Recipe model.EncyclopediaRecipe
}

func (r *EncyclopediaRepository) UpsertExternal(recipe model.EncyclopediaRecipe) (*model.EncyclopediaRecipe, error) {
	if recipe.ExternalSource == nil || recipe.ExternalID == nil {
		return nil, fmt.Errorf("external source/id required")
	}
	var existing model.EncyclopediaRecipe
	err := r.db.Where("external_source = ? AND external_id = ?", *recipe.ExternalSource, *recipe.ExternalID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.db.Create(&recipe).Error; err != nil {
			return nil, err
		}
		return &recipe, nil
	}
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"name":            recipe.Name,
		"description":     recipe.Description,
		"cover_image_url": recipe.CoverImageURL,
		"category":        recipe.Category,
		"tags":            recipe.Tags,
		"ingredients":     recipe.Ingredients,
		"process_steps":   recipe.ProcessSteps,
		"source":          recipe.Source,
	}
	if err := r.db.Model(&existing).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := r.db.First(&existing, existing.ID).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// GetByExternal 按外部来源+ID查找记录
func (r *EncyclopediaRepository) GetByExternal(source, externalID string) (*model.EncyclopediaRecipe, error) {
	var recipe model.EncyclopediaRecipe
	err := r.db.Where("external_source = ? AND external_id = ?", source, externalID).First(&recipe).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &recipe, err
}
