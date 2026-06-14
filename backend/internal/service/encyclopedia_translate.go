package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/delicious/delicious/internal/dto"
)

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

func (s *EncyclopediaService) applyListLang(ctx context.Context, items []dto.EncyclopediaListItemDTO, lang string) []dto.EncyclopediaListItemDTO {
	lang = normalizeDisplayLang(lang)
	if lang == "en" {
		return items
	}
	client := s.translateClient()
	out := make([]dto.EncyclopediaListItemDTO, len(items))
	for i, item := range items {
		out[i] = item
		if !isMostlyChinese(item.Name) {
			if translated, err := translateWithPair(ctx, client, item.Name, "en|zh-CN"); err == nil {
				out[i].Name = translated
			}
		}
		if item.Description != nil && !isMostlyChinese(*item.Description) {
			if translated, err := translateLongText(ctx, client, *item.Description, "en|zh-CN"); err == nil {
				out[i].Description = &translated
			}
		}
		if item.Category != nil && !isMostlyChinese(*item.Category) {
			if translated, err := translateWithPair(ctx, client, *item.Category, "en|zh-CN"); err == nil {
				out[i].Category = &translated
			}
		}
	}
	return out
}

func (s *EncyclopediaService) applyRecipeLang(ctx context.Context, recipe *dto.EncyclopediaRecipeDTO, lang string) *dto.EncyclopediaRecipeDTO {
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

	if name, err := translateWithPair(ctx, client, out.Name, langPair); err == nil {
		out.Name = name
	}
	if out.Description != nil {
		if desc, err := translateLongText(ctx, client, *out.Description, langPair); err == nil {
			out.Description = &desc
		}
	}
	if out.Category != nil {
		if cat, err := translateWithPair(ctx, client, *out.Category, langPair); err == nil {
			out.Category = &cat
		}
	}
	for i := range out.Ingredients {
		if name, err := translateWithPair(ctx, client, out.Ingredients[i].Name, langPair); err == nil {
			out.Ingredients[i].Name = name
		}
		if unit, err := translateWithPair(ctx, client, out.Ingredients[i].Unit, langPair); err == nil {
			out.Ingredients[i].Unit = unit
		}
	}
	for i := range out.ProcessSteps {
		if content, err := translateLongText(ctx, client, out.ProcessSteps[i].Content, langPair); err == nil {
			out.ProcessSteps[i].Content = content
		}
	}
	if len(out.Tags) > 0 {
		tags := make([]string, len(out.Tags))
		for i, tag := range out.Tags {
			if translated, err := translateWithPair(ctx, client, tag, langPair); err == nil {
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
