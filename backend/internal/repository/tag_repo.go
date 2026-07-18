package repository

import (
	"github.com/delicious/delicious/pkg/model"
	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// ── 标签字典 CRUD ──

func (r *TagRepository) ListAll() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Order("id ASC").Find(&tags).Error
	return tags, err
}

func (r *TagRepository) Create(name string) (*model.Tag, error) {
	t := model.Tag{Name: name}
	if err := r.db.Create(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TagRepository) Delete(id uint64) error {
	res := r.db.Delete(&model.Tag{}, id)
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return res.Error
}

// ── 百科菜谱-标签关联 ──

func (r *TagRepository) ListByRecipe(recipeID uint64) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Joins("JOIN encyclopedia_recipe_tags ON encyclopedia_recipe_tags.tag_id = tags.id").
		Where("encyclopedia_recipe_tags.recipe_id = ?", recipeID).
		Order("tags.id ASC").
		Find(&tags).Error
	return tags, err
}

func (r *TagRepository) AddToRecipe(recipeID, tagID uint64) error {
	rel := model.EncyclopediaRecipeTag{RecipeID: recipeID, TagID: tagID}
	return r.db.Create(&rel).Error
}

func (r *TagRepository) RemoveFromRecipe(recipeID, tagID uint64) error {
	res := r.db.Where("recipe_id = ? AND tag_id = ?", recipeID, tagID).
		Delete(&model.EncyclopediaRecipeTag{})
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return res.Error
}

func (r *TagRepository) RecipeTagIDs(recipeID uint64) ([]uint64, error) {
	var rels []model.EncyclopediaRecipeTag
	if err := r.db.Where("recipe_id = ?", recipeID).Find(&rels).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(rels))
	for _, rel := range rels {
		ids = append(ids, rel.TagID)
	}
	return ids, nil
}

// BatchRecipeTags 批量获取多个食谱的标签映射。recipeID → []Tag
func (r *TagRepository) BatchRecipeTags(recipeIDs []uint64) (map[uint64][]model.Tag, error) {
	if len(recipeIDs) == 0 {
		return map[uint64][]model.Tag{}, nil
	}
	type row struct {
		RecipeID uint64
		Tag      model.Tag
	}
	var rows []row
	err := r.db.Table("encyclopedia_recipe_tags").
		Select("encyclopedia_recipe_tags.recipe_id, tags.*").
		Joins("JOIN tags ON tags.id = encyclopedia_recipe_tags.tag_id").
		Where("encyclopedia_recipe_tags.recipe_id IN ?", recipeIDs).
		Order("tags.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint64][]model.Tag, len(recipeIDs))
	for _, row := range rows {
		m[row.RecipeID] = append(m[row.RecipeID], row.Tag)
	}
	return m, nil
}

// TagExists 检查标签名是否已存在
func (r *TagRepository) TagExists(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Tag{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}
