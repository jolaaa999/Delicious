package service

import (
	"github.com/delicious/delicious/internal/dto"
	"github.com/delicious/delicious/internal/repository"
)

type TagService struct {
	repo *repository.TagRepository
}

func NewTagService(repo *repository.TagRepository) *TagService {
	return &TagService{repo: repo}
}

// ── 标签字典 ──

func (s *TagService) List() ([]dto.TagDTO, error) {
	tags, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.TagDTO, 0, len(tags))
	for _, t := range tags {
		result = append(result, dto.TagDTO{ID: t.ID, Name: t.Name})
	}
	return result, nil
}

func (s *TagService) Create(name string) (*dto.TagDTO, error) {
	t, err := s.repo.Create(name)
	if err != nil {
		return nil, err
	}
	return &dto.TagDTO{ID: t.ID, Name: t.Name}, nil
}

func (s *TagService) Delete(id uint64) error {
	return s.repo.Delete(id)
}

// ── 百科菜谱标签关联 ──

func (s *TagService) ListByRecipe(recipeID uint64) ([]dto.TagDTO, error) {
	tags, err := s.repo.ListByRecipe(recipeID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.TagDTO, 0, len(tags))
	for _, t := range tags {
		result = append(result, dto.TagDTO{ID: t.ID, Name: t.Name})
	}
	return result, nil
}

func (s *TagService) AddToRecipe(recipeID, tagID uint64) error {
	return s.repo.AddToRecipe(recipeID, tagID)
}

func (s *TagService) RemoveFromRecipe(recipeID, tagID uint64) error {
	return s.repo.RemoveFromRecipe(recipeID, tagID)
}
