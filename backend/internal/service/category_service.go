package service

import (
	"github.com/delicious/delicious/internal/dto"
	"github.com/delicious/delicious/internal/repository"
)

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List() ([]dto.CategoryDTO, error) {
	categories, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryDTO, 0, len(categories))
	for _, c := range categories {
		result = append(result, dto.CategoryDTO{
			ID:   c.ID,
			Name: c.Name,
		})
	}
	return result, nil
}

func (s *CategoryService) Create(name string) (*dto.CategoryDTO, error) {
	c, err := s.repo.Create(name)
	if err != nil {
		return nil, err
	}
	return &dto.CategoryDTO{ID: c.ID, Name: c.Name}, nil
}

func (s *CategoryService) Update(id uint64, name string) (*dto.CategoryDTO, error) {
	c, err := s.repo.Update(id, name)
	if err != nil {
		return nil, err
	}
	return &dto.CategoryDTO{ID: c.ID, Name: c.Name}, nil
}

func (s *CategoryService) Delete(id uint64) error {
	return s.repo.Delete(id)
}
