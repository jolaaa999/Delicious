package service

import (
	"github.com/delicious/delicious/internal/dto"
	"github.com/delicious/delicious/internal/repository"
)

type DashboardService struct {
	repo *repository.RecipeRepository
}

func NewDashboardService(repo *repository.RecipeRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) Stats(userID uint64) (*dto.DashboardStatsDTO, error) {
	total, err := s.repo.CountByUser(userID)
	if err != nil {
		return nil, err
	}
	avg, err := s.repo.AverageRating(userID)
	if err != nil {
		return nil, err
	}
	verCount, err := s.repo.CountVersionsByUser(userID)
	if err != nil {
		return nil, err
	}
	distMap, err := s.repo.RatingDistribution(userID)
	if err != nil {
		return nil, err
	}
	latest, _ := s.repo.LatestRecipeAt(userID)

	dist := make([]dto.RatingDistributionDTO, 0, 5)
	for rating := uint8(1); rating <= 5; rating++ {
		dist = append(dist, dto.RatingDistributionDTO{
			Rating: rating,
			Count:  distMap[rating],
		})
	}

	return &dto.DashboardStatsDTO{
		TotalRecipes:       total,
		AverageRating:      avg,
		TotalVersions:      verCount,
		RatingDistribution: dist,
		LatestRecipeAt:     latest,
	}, nil
}
