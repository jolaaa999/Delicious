package service

import (
	"context"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/delicious/delicious/internal/config"
	"github.com/delicious/delicious/internal/dto"
	"github.com/delicious/delicious/internal/repository"
	"github.com/delicious/delicious/pkg/cache"
	"github.com/delicious/delicious/pkg/model"
)

type EncyclopediaService struct {
	repo       *repository.EncyclopediaRepository
	online     *OnlineRecipeSearch
	httpClient *http.Client
	transCache *cache.MemoryCache
}

func NewEncyclopediaService(repo *repository.EncyclopediaRepository, cfg config.Config) *EncyclopediaService {
	return &EncyclopediaService{
		repo:       repo,
		online:     NewOnlineRecipeSearch(cfg),
		httpClient: &http.Client{Timeout: 15 * time.Second},
		transCache: cache.New(1 * time.Hour),
	}
}

func (s *EncyclopediaService) Search(keyword, category, lang string, page, pageSize int) ([]dto.EncyclopediaListItemDTO, dto.PageInfo, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	highlightTerms := expandSearchKeywords(ctx, s.httpClient, keyword)

	if !s.online.Enabled() {
		return []dto.EncyclopediaListItemDTO{}, pageInfoFrom(page, pageSize, 0), highlightTerms, nil
	}
	if strings.TrimSpace(keyword) == "" {
		return []dto.EncyclopediaListItemDTO{}, pageInfoFrom(page, pageSize, 0), nil, nil
	}

	items, pageInfo, err := s.searchOnline(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, dto.PageInfo{}, highlightTerms, err
	}
	if category != "" {
		filtered := make([]dto.EncyclopediaListItemDTO, 0, len(items))
		for _, item := range items {
			if item.Category != nil && strings.EqualFold(*item.Category, category) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
		pageInfo.Total = int64(len(filtered))
		pageInfo.TotalPages = 1
		if pageSize > 0 {
			pageInfo.TotalPages = int(math.Ceil(float64(len(filtered)) / float64(pageSize)))
		}
	}
	return s.applyListLang(ctx, items, lang), pageInfo, highlightTerms, nil
}

func (s *EncyclopediaService) searchOnline(ctx context.Context, keyword string, page, pageSize int) ([]dto.EncyclopediaListItemDTO, dto.PageInfo, error) {
	hits, total, err := s.online.Search(ctx, keyword, page, pageSize)
	if err != nil || len(hits) == 0 {
		return nil, dto.PageInfo{}, err
	}
	cached, err := cacheOnlineHits(s.repo, hits)
	if err != nil {
		return nil, dto.PageInfo{}, err
	}
	items := make([]dto.EncyclopediaListItemDTO, 0, len(cached))
	for _, hit := range cached {
		items = append(items, toListItemDTO(&hit.Recipe))
	}
	return items, pageInfoFrom(page, pageSize, int64(total)), nil
}

func (s *EncyclopediaService) Get(id uint64, lang string) (*dto.EncyclopediaRecipeDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	e, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if s.shouldRefreshOnline(e) {
		if hit, fetchErr := s.online.Fetch(ctx, *e.ExternalSource, *e.ExternalID); fetchErr == nil {
			if updated, upsertErr := s.repo.UpsertExternal(hit.toModel()); upsertErr == nil {
				e = updated
			}
		}
	}
	out := toEncyclopediaDetailDTO(e)
	if out.Ingredients == nil {
		out.Ingredients = []dto.Ingredient{}
	}
	if out.ProcessSteps == nil {
		out.ProcessSteps = []dto.ProcessStep{}
	}
	return s.applyRecipeLang(ctx, &out, lang), nil
}

func (s *EncyclopediaService) shouldRefreshOnline(e *model.EncyclopediaRecipe) bool {
	if e.ExternalSource == nil || e.ExternalID == nil || !s.online.Enabled() {
		return false
	}
	if len(e.Ingredients) == 0 || len(e.ProcessSteps) == 0 {
		return true
	}
	// Forkify 等源搜索阶段可能只有占位步骤，进入详情时补全
	if len(e.ProcessSteps) == 1 && strings.Contains(e.ProcessSteps[0].Content, "参考原菜谱") {
		return true
	}
	return false
}

func (s *EncyclopediaService) ListByCategory(category, lang string, page, pageSize int) ([]dto.EncyclopediaListItemDTO, dto.PageInfo, error) {
	items, pageInfo, _, err := s.Search("", category, lang, page, pageSize)
	return items, pageInfo, err
}

func toListItemDTO(e *model.EncyclopediaRecipe) dto.EncyclopediaListItemDTO {
	tags := []string(e.Tags)
	return dto.EncyclopediaListItemDTO{
		ID:            e.ID,
		Name:          e.Name,
		CoverImageURL: e.CoverImageURL,
		Category:      e.Category,
		Tags:          tags,
		Description:   e.Description,
	}
}

func pageInfoFrom(page, pageSize int, total int64) dto.PageInfo {
	if pageSize <= 0 {
		pageSize = 20
	}
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	return dto.PageInfo{Page: page, PageSize: pageSize, Total: total, TotalPages: pages}
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
