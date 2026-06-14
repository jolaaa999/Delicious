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
	providers  []onlineRecipeProvider
	enabled    bool
	httpClient *http.Client
}

func NewOnlineRecipeSearch(cfg config.Config) *OnlineRecipeSearch {
	client := &http.Client{Timeout: 15 * time.Second}
	var providers []onlineRecipeProvider
	if cfg.SpoonacularAPIKey != "" {
		providers = append(providers, newSpoonacularProvider(cfg.SpoonacularAPIKey, client))
	}
	providers = append(providers, newMealDBProvider(client))
	return &OnlineRecipeSearch{
		providers:  providers,
		enabled:    cfg.OnlineSearchEnabled,
		httpClient: client,
	}
}

func (s *OnlineRecipeSearch) Enabled() bool {
	return s != nil && s.enabled && len(s.providers) > 0
}

func (s *OnlineRecipeSearch) Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error) {
	if !s.Enabled() || keyword == "" {
		return nil, 0, fmt.Errorf("online search disabled")
	}

	keywords := expandSearchKeywords(ctx, s.httpClient, keyword)
	seen := map[string]bool{}
	allHits := make([]OnlineRecipeHit, 0, pageSize)
	totalMax := 0
	var lastErr error

	for _, kw := range keywords {
		for _, p := range s.providers {
			hits, total, err := p.Search(ctx, kw, page, pageSize)
			if err != nil {
				lastErr = err
				continue
			}
			if total > totalMax {
				totalMax = total
			}
			for _, hit := range hits {
				key := hit.Source + ":" + hit.ExternalID
				if seen[key] {
					continue
				}
				seen[key] = true
				allHits = append(allHits, hit)
			}
			if len(allHits) >= pageSize {
				break
			}
		}
		if len(allHits) >= pageSize {
			break
		}
	}

	if len(allHits) > pageSize {
		allHits = allHits[:pageSize]
	}
	if len(allHits) == 0 {
		if lastErr != nil {
			return nil, 0, lastErr
		}
		return nil, 0, nil
	}
	if totalMax == 0 {
		totalMax = len(allHits)
	}
	return allHits, totalMax, nil
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
