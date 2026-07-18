package service

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/delicious/delicious/internal/dto"
)

const listTranslateTimeout = 25 * time.Second

func (s *EncyclopediaService) applyListLang(ctx context.Context, items []dto.EncyclopediaListItemDTO, lang string) []dto.EncyclopediaListItemDTO {
	lang = normalizeDisplayLang(lang)
	if lang == "en" {
		return items
	}

	out := make([]dto.EncyclopediaListItemDTO, len(items))
	copy(out, items)

	ctx, cancel := context.WithTimeout(ctx, listTranslateTimeout)
	defer cancel()

	client := s.translateClient()
	const maxWorkers = 6
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for i := range out {
		if isMostlyChinese(out[i].Name) {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			if translated, err := CachedTranslate(ctx, s.transCache, client, out[idx].Name, "en|zh-CN"); err == nil {
				out[idx].Name = translated
			}
		}(i)
	}
	wg.Wait()
	return out
}

func normalizeDisplayLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "zh", "zh-cn", "zh-hans", "chinese":
		return "zh"
	default:
		return "en"
	}
}

func isMostlyChinese(text string) bool {
	var chinese, latin int
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			chinese++
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			latin++
		}
	}
	return chinese > latin
}

func (s *EncyclopediaService) applyRecipeLang(ctx context.Context, recipe *dto.EncyclopediaRecipeDTO, lang string) *dto.EncyclopediaRecipeDTO {
	ctx, cancel := context.WithTimeout(ctx, 75*time.Second)
	defer cancel()

	lang = normalizeDisplayLang(lang)
	if lang == "en" {
		if isMostlyChinese(recipe.Name) {
			return s.translateRecipeDTO(ctx, recipe, "zh-CN|en")
		}
		return recipe
	}
	if isMostlyChinese(recipe.Name) {
		return recipe
	}
	return s.translateRecipeDTO(ctx, recipe, "en|zh-CN")
}

func (s *EncyclopediaService) translateRecipeDTO(ctx context.Context, recipe *dto.EncyclopediaRecipeDTO, langPair string) *dto.EncyclopediaRecipeDTO {
	client := s.translateClient()
	out := *recipe

	if name, err := CachedTranslate(ctx, s.transCache, client, out.Name, langPair); err == nil {
		out.Name = name
	}
	if out.Description != nil {
		if desc, err := CachedTranslateLong(ctx, s.transCache, client, *out.Description, langPair); err == nil {
			out.Description = &desc
		}
	}
	if out.Category != nil {
		if cat, err := CachedTranslate(ctx, s.transCache, client, *out.Category, langPair); err == nil {
			out.Category = &cat
		}
	}

	var wg sync.WaitGroup
	for i := range out.Ingredients {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if name, err := CachedTranslate(ctx, s.transCache, client, out.Ingredients[idx].Name, langPair); err == nil {
				out.Ingredients[idx].Name = name
			}
			if unit, err := CachedTranslate(ctx, s.transCache, client, out.Ingredients[idx].Unit, langPair); err == nil {
				out.Ingredients[idx].Unit = unit
			}
		}(i)
	}
	for i := range out.ProcessSteps {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if content, err := CachedTranslateLong(ctx, s.transCache, client, out.ProcessSteps[idx].Content, langPair); err == nil {
				out.ProcessSteps[idx].Content = content
			}
		}(i)
	}
	wg.Wait()

	if len(out.Tags) > 0 {
		tags := make([]string, len(out.Tags))
		for i, tag := range out.Tags {
			if translated, err := CachedTranslate(ctx, s.transCache, client, tag, langPair); err == nil {
				tags[i] = translated
			} else {
				tags[i] = tag
			}
		}
		out.Tags = tags
	}
	return &out
}

func (s *EncyclopediaService) translateClient() *http.Client {
	if s.httpClient != nil {
		return s.httpClient
	}
	return &http.Client{Timeout: 12 * time.Second}
}
