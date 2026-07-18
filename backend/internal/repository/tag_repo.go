package repository

import (
	"errors"

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

// SyncTagsForRecipe 同步食谱标签：对每个标签名 find-or-create，然后全量替换关联关系
func (r *TagRepository) SyncTagsForRecipe(recipeID uint64, tagNames []string) error {
	if len(tagNames) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 清除旧关联
		if err := tx.Where("recipe_id = ?", recipeID).Delete(&model.EncyclopediaRecipeTag{}).Error; err != nil {
			return err
		}
		for _, name := range tagNames {
			if name == "" {
				continue
			}
			var tag model.Tag
			err := tx.Where("name = ?", name).First(&tag).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tag = model.Tag{Name: name}
				if err := tx.Create(&tag).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			if err := tx.Create(&model.EncyclopediaRecipeTag{
				RecipeID: recipeID,
				TagID:    tag.ID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
