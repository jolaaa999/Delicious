package service

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/delicious/delicious/internal/dto"
	"github.com/delicious/delicious/pkg/cache"
)

const listTranslateTimeout = 45 * time.Second

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
	const maxWorkers = 4
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
			if translated := translateWithRetry(ctx, s.transCache, client, out[idx].Name, "en|zh-CN"); translated != "" {
				out[idx].Name = translated
			}
			if out[idx].Description != nil && !isMostlyChinese(*out[idx].Description) {
				if desc := translateWithRetry(ctx, s.transCache, client, *out[idx].Description, "en|zh-CN"); desc != "" {
					out[idx].Description = &desc
				}
			}
		}(i)
	}
	wg.Wait()
	return out
}

func translateWithRetry(ctx context.Context, c *cache.MemoryCache, client *http.Client, text, langPair string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	var last string
	for attempt := 0; attempt < 3; attempt++ {
		if ctx.Err() != nil {
			return last
		}
		translated, err := CachedTranslate(ctx, c, client, text, langPair)
		if err == nil && strings.TrimSpace(translated) != "" {
			return translated
		}
		select {
		case <-ctx.Done():
			return last
		case <-time.After(time.Duration(150*(attempt+1)) * time.Millisecond):
		}
	}
	return last
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
		return sanitizeIngredientUnits(recipe, "zh-CN|en")
	}
	var out *dto.EncyclopediaRecipeDTO
	if isMostlyChinese(recipe.Name) {
		out = recipe
	} else {
		out = s.translateRecipeDTO(ctx, recipe, "en|zh-CN")
	}
	return sanitizeIngredientUnits(out, "en|zh-CN")
}

// sanitizeIngredientUnits 清洗已缓存/已翻译的脏单位（广告文案等）。
func sanitizeIngredientUnits(recipe *dto.EncyclopediaRecipeDTO, langPair string) *dto.EncyclopediaRecipeDTO {
	if recipe == nil {
		return nil
	}
	out := *recipe
	ings := make([]dto.Ingredient, len(out.Ingredients))
	copy(ings, out.Ingredients)
	for i := range ings {
		ings[i].Unit = localizeUnit(ings[i].Unit, langPair)
	}
	out.Ingredients = ings
	return &out
}

func (s *EncyclopediaService) translateRecipeDTO(ctx context.Context, recipe *dto.EncyclopediaRecipeDTO, langPair string) *dto.EncyclopediaRecipeDTO {
	client := s.translateClient()
	out := *recipe
	// 深拷贝切片，避免并发改写污染调用方/缓存中的原文
	if len(recipe.Ingredients) > 0 {
		ings := make([]dto.Ingredient, len(recipe.Ingredients))
		copy(ings, recipe.Ingredients)
		out.Ingredients = ings
	}
	if len(recipe.ProcessSteps) > 0 {
		steps := make([]dto.ProcessStep, len(recipe.ProcessSteps))
		copy(steps, recipe.ProcessSteps)
		out.ProcessSteps = steps
	}

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
			origName := out.Ingredients[idx].Name
			origUnit := out.Ingredients[idx].Unit
			if name, err := CachedTranslate(ctx, s.transCache, client, origName, langPair); err == nil {
				out.Ingredients[idx].Name = name
			}
			// 单位绝不能走免费翻译 API（cup/clove/份 等短词常被灌成广告文）
			out.Ingredients[idx].Unit = localizeUnit(origUnit, langPair)
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

// localizeUnit 用本地词典处理烹饪单位，避免短词被 MyMemory 灌广告。
func localizeUnit(unit, langPair string) string {
	unit = strings.TrimSpace(unit)
	toZh := strings.HasPrefix(langPair, "en") && strings.Contains(langPair, "zh")
	fallback := "适量"
	if !toZh {
		fallback = "serving"
	}
	if unit == "" {
		return fallback
	}
	// dirty / spam
	if unit == "待定" || strings.Contains(unit, "千锋") || len([]rune(unit)) > 12 {
		return fallback
	}
	key := strings.ToLower(unit)

	enToZh := map[string]string{
		"g": "克", "gram": "克", "grams": "克",
		"kg": "千克", "kilogram": "千克", "kilograms": "千克",
		"ml": "毫升", "milliliter": "毫升", "milliliters": "毫升",
		"l": "升", "liter": "升", "litre": "升", "liters": "升", "litres": "升",
		"cup": "杯", "cups": "杯",
		"tbsp": "汤匙", "tbs": "汤匙", "tablespoon": "汤匙", "tablespoons": "汤匙",
		"tsp": "茶匙", "teaspoon": "茶匙", "teaspoons": "茶匙",
		"oz": "盎司", "ounce": "盎司", "ounces": "盎司",
		"lb": "磅", "lbs": "磅", "pound": "磅", "pounds": "磅",
		"clove": "瓣", "cloves": "瓣",
		"piece": "个", "pieces": "个",
		"slice": "片", "slices": "片",
		"can": "罐", "cans": "罐",
		"pinch": "少许", "dash": "少许",
		"to taste": "适量", "taste": "适量",
		"handful": "一把", "bunch": "把",
		"stick": "根", "sticks": "根",
		"serving": "份", "servings": "份",
		"份": "份",
	}

	switch {
	case strings.HasPrefix(langPair, "en") && strings.Contains(langPair, "zh"):
		if v, ok := enToZh[key]; ok {
			return v
		}
		if isMostlyChinese(unit) {
			return unit
		}
		return unit
	case strings.HasPrefix(langPair, "zh") && strings.Contains(langPair, "en"):
		zhToEn := map[string]string{
			"克": "g", "千克": "kg", "公斤": "kg", "毫升": "ml", "升": "l",
			"杯": "cup", "汤匙": "tbsp", "茶匙": "tsp", "大勺": "tbsp", "小勺": "tsp",
			"盎司": "oz", "磅": "lb", "瓣": "clove", "个": "piece", "片": "slice",
			"罐": "can", "少许": "pinch", "适量": "to taste", "一把": "handful",
			"把": "bunch", "根": "stick", "份": "serving",
		}
		if v, ok := zhToEn[unit]; ok {
			return v
		}
		return unit
	default:
		return unit
	}
}
