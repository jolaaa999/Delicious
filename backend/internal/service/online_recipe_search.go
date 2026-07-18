package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
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
	// 中文免费源优先，提升家常菜命中率
	providers = append(providers,
		newHowToCookProvider(client),
		newProjKitchenProvider(client),
	)
	if cfg.SpoonacularAPIKey != "" {
		providers = append(providers, newSpoonacularProvider(cfg.SpoonacularAPIKey, client))
	}
	providers = append(providers,
		newMealDBProvider(client),
		newForkifyProvider(client),
		newDummyJSONProvider(client),
	)
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
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	}

	keywords := expandSearchKeywords(ctx, s.httpClient, keyword)
	// 拉取足够候选后再全局排序分页，保证跨页排序一致
	fetchSize := page * pageSize
	if fetchSize < pageSize {
		fetchSize = pageSize
	}
	if fetchSize > 80 {
		fetchSize = 80
	}

	seen := map[string]bool{}
	allHits := make([]OnlineRecipeHit, 0, fetchSize)
	totalMax := 0

	for _, kw := range keywords {
		for _, p := range s.providers {
			hits, total, err := p.Search(ctx, kw, 1, fetchSize)
			if err != nil {
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
		}
	}

	if len(allHits) == 0 {
		// 各源失败（含 Spoonacular 402 额度用尽）不向客户端暴露原始 API 错误
		return nil, 0, nil
	}

	s.enrichHitsForRanking(ctx, allHits, keywords)
	sortHitsByRelevance(allHits, keywords)

	if totalMax < len(allHits) {
		totalMax = len(allHits)
	}

	start := (page - 1) * pageSize
	if start >= len(allHits) {
		return nil, totalMax, nil
	}
	end := start + pageSize
	if end > len(allHits) {
		end = len(allHits)
	}
	return allHits[start:end], totalMax, nil
}

// enrichHitsForRanking 为标题未命中的结果补全材料/步骤，便于内容匹配排序。
func (s *OnlineRecipeSearch) enrichHitsForRanking(ctx context.Context, hits []OnlineRecipeHit, keywords []string) {
	type job struct{ idx int }
	jobs := make([]job, 0)
	for i := range hits {
		if needsContentEnrichment(hits[i], keywords) {
			jobs = append(jobs, job{idx: i})
		}
	}
	if len(jobs) == 0 {
		return
	}
	if len(jobs) > 24 {
		jobs = jobs[:24]
	}

	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()

			hit := hits[j.idx]
			detail, err := s.Fetch(ctx, hit.Source, hit.ExternalID)
			if err != nil || detail == nil {
				return
			}
			mu.Lock()
			if len(hits[j.idx].Ingredients) == 0 {
				hits[j.idx].Ingredients = detail.Ingredients
			}
			if len(hits[j.idx].ProcessSteps) == 0 {
				hits[j.idx].ProcessSteps = detail.ProcessSteps
			}
			if hits[j.idx].Description == nil && detail.Description != nil {
				hits[j.idx].Description = detail.Description
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
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

func cacheOnlineHits(repo *repository.EncyclopediaRepository, tagRepo *repository.TagRepository, hits []OnlineRecipeHit) ([]repository.CachedOnlineHit, error) {
	out := make([]repository.CachedOnlineHit, 0, len(hits))
	for _, hit := range hits {
		// 映射英文分类→中文分类
		if hit.Category != nil {
			mapped := MapCategory(*hit.Category)
			hit.Category = &mapped
		}
		// 映射英文标签→中文标签
		hit.Tags = MapTags(hit.Tags)

		model := hit.toModel()
		item, err := repo.UpsertExternal(model)
		if err != nil {
			return nil, err
		}
		// 同步中文标签到关联表
		_ = tagRepo.SyncTagsForRecipe(item.ID, hit.Tags)
		out = append(out, repository.CachedOnlineHit{ID: item.ID, Recipe: *item})
	}
	return out, nil
}
