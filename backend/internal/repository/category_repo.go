package repository

import (
	"errors"

	"github.com/delicious/delicious/pkg/model"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) ListAll() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Order("id ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) Create(name string) (*model.Category, error) {
	c := model.Category{Name: name}
	if err := r.db.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) Update(id uint64, name string) (*model.Category, error) {
	var c model.Category
	if err := r.db.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := r.db.Model(&c).Update("name", name).Error; err != nil {
		return nil, err
	}
	c.Name = name
	return &c, nil
}

func (r *CategoryRepository) Delete(id uint64) error {
	res := r.db.Delete(&model.Category{}, id)
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return res.Error
}

func (r *CategoryRepository) Exists(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Category{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}
