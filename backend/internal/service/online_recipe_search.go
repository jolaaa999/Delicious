package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/delicious/delicious/internal/config"
	"github.com/delicious/delicious/internal/repository"
)

type onlineRecipeProvider interface {
	Name() string
	Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error)
	Fetch(ctx context.Context, externalID string) (*OnlineRecipeHit, error)
}

// OnlineRecipeSearch 聚合多个公开菜谱 API 进行联网搜索。
type OnlineRecipeSearch struct {
	providers []onlineRecipeProvider
	enabled   bool
}

func NewOnlineRecipeSearch(cfg config.Config) *OnlineRecipeSearch {
	var providers []onlineRecipeProvider
	if cfg.SpoonacularAPIKey != "" {
		providers = append(providers, newSpoonacularProvider(cfg.SpoonacularAPIKey, &http.Client{Timeout: 12 * time.Second}))
	}
	providers = append(providers, newMealDBProvider(&http.Client{Timeout: 12 * time.Second}))
	return &OnlineRecipeSearch{
		providers: providers,
		enabled:   cfg.OnlineSearchEnabled,
	}
}

func (s *OnlineRecipeSearch) Enabled() bool {
	return s != nil && s.enabled && len(s.providers) > 0
}

func (s *OnlineRecipeSearch) Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error) {
	if !s.Enabled() || keyword == "" {
		return nil, 0, fmt.Errorf("online search disabled")
	}
	var lastErr error
	for _, p := range s.providers {
		hits, total, err := p.Search(ctx, keyword, page, pageSize)
		if err != nil {
			lastErr = err
			continue
		}
		if len(hits) > 0 {
			return hits, total, nil
		}
	}
	if lastErr != nil {
		return nil, 0, lastErr
	}
	return nil, 0, nil
}

func (s *OnlineRecipeSearch) Fetch(ctx context.Context, source, externalID string) (*OnlineRecipeHit, error) {
	if !s.Enabled() {
		return nil, fmt.Errorf("online search disabled")
	}
	for _, p := range s.providers {
		if p.Name() != source {
			continue
		}
		return p.Fetch(ctx, externalID)
	}
	return nil, fmt.Errorf("unknown online source: %s", source)
}

func cacheOnlineHits(repo *repository.EncyclopediaRepository, hits []OnlineRecipeHit) ([]repository.CachedOnlineHit, error) {
	out := make([]repository.CachedOnlineHit, 0, len(hits))
	for _, hit := range hits {
		item, err := repo.UpsertExternal(hit.toModel())
		if err != nil {
			return nil, err
		}
		out = append(out, repository.CachedOnlineHit{ID: item.ID, Recipe: *item})
	}
	return out, nil
}
