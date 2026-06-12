package service

import (
	"math"

	"github.com/delicious/delicious/internal/dto"
	"github.com/delicious/delicious/internal/repository"
	"github.com/delicious/delicious/pkg/model"
)

type EncyclopediaService struct {
	repo *repository.EncyclopediaRepository
}

func NewEncyclopediaService(repo *repository.EncyclopediaRepository) *EncyclopediaService {
	return &EncyclopediaService{repo: repo}
}

func (s *EncyclopediaService) Search(keyword, category string, page, pageSize int) ([]dto.EncyclopediaListItemDTO, dto.PageInfo, error) {
	items, total, err := s.repo.Search(repository.SearchFilter{
		Keyword:  keyword,
		Category: category,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, dto.PageInfo{}, err
	}
	result := make([]dto.EncyclopediaListItemDTO, 0, len(items))
	for _, e := range items {
		tags := []string(e.Tags)
		result = append(result, dto.EncyclopediaListItemDTO{
			ID:            e.ID,
			Name:          e.Name,
			CoverImageURL: e.CoverImageURL,
			Category:      e.Category,
			Tags:          tags,
			Description:   e.Description,
		})
	}
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	return result, dto.PageInfo{Page: page, PageSize: pageSize, Total: total, TotalPages: pages}, nil
}

func (s *EncyclopediaService) Get(id uint64) (*dto.EncyclopediaRecipeDTO, error) {
	e, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	out := toEncyclopediaDetailDTO(e)
	return &out, nil
}

func (s *EncyclopediaService) ListByCategory(category string, page, pageSize int) ([]dto.EncyclopediaListItemDTO, dto.PageInfo, error) {
	return s.Search("", category, page, pageSize)
}

func toEncyclopediaDetailDTO(e *model.EncyclopediaRecipe) dto.EncyclopediaRecipeDTO {
	tags := []string(e.Tags)
	if tags == nil {
		tags = []string{}
	}
	return dto.EncyclopediaRecipeDTO{
		ID:            e.ID,
		Name:          e.Name,
		Description:   e.Description,
		CoverImageURL: e.CoverImageURL,
		Category:      e.Category,
		Tags:          tags,
		Ingredients:   []dto.Ingredient(e.Ingredients),
		ProcessSteps:  []dto.ProcessStep(e.ProcessSteps),
		Source:        e.Source,
		ViewCount:     e.ViewCount,
		CreatedAt:     e.CreatedAt,
	}
}
