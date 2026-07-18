package service

import (
	"sort"
	"strings"
)

const (
	rankExactTitle    = 300
	rankPartialTitle  = 200
	rankPartialContent = 100
)

func normalizeSearchText(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// scoreRecipeHit 按标题完全匹配 > 标题部分匹配 > 材料/步骤部分匹配打分。
func scoreRecipeHit(hit OnlineRecipeHit, keywords []string) int {
	best := 0
	name := normalizeSearchText(hit.Name)
	for _, kw := range keywords {
		k := normalizeSearchText(kw)
		if k == "" {
			continue
		}
		if name == k {
			return rankExactTitle
		}
		if titlePartiallyMatches(name, k) {
			if rankPartialTitle > best {
				best = rankPartialTitle
			}
			continue
		}
		if best >= rankPartialContent {
			continue
		}
		if recipeContentContains(hit, k) {
			best = rankPartialContent
		}
	}
	return best
}

func recipeContentContains(hit OnlineRecipeHit, keyword string) bool {
	if keyword == "" {
		return false
	}
	if hit.Description != nil && textPartiallyMatches(normalizeSearchText(*hit.Description), keyword) {
		return true
	}
	for _, tag := range hit.Tags {
		if textPartiallyMatches(normalizeSearchText(tag), keyword) {
			return true
		}
	}
	for _, ing := range hit.Ingredients {
		text := normalizeSearchText(ing.Name)
		if ing.Note != "" {
			text += " " + normalizeSearchText(ing.Note)
		}
		if textPartiallyMatches(text, keyword) {
			return true
		}
	}
	for _, step := range hit.ProcessSteps {
		if textPartiallyMatches(normalizeSearchText(step.Content), keyword) {
			return true
		}
	}
	return false
}

func titlePartiallyMatches(name, keyword string) bool {
	return textPartiallyMatches(name, keyword)
}

// textPartiallyMatches：整句包含，或关键词中任一有意义片段命中。
func textPartiallyMatches(haystack, needle string) bool {
	if needle == "" || haystack == "" {
		return false
	}
	if strings.Contains(haystack, needle) || strings.Contains(needle, haystack) {
		return true
	}
	for _, p := range keywordTokens(needle) {
		if strings.Contains(haystack, p) {
			return true
		}
	}
	return false
}

func keywordTokens(keyword string) []string {
	parts := strings.Fields(keyword)
	out := make([]string, 0, len(parts)+2)
	seen := map[string]bool{}
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			return
		}
		runes := []rune(p)
		if len(runes) < 2 {
			return
		}
		seen[p] = true
		out = append(out, p)
		if strings.HasSuffix(p, "s") && len(runes) > 2 {
			stem := strings.TrimSuffix(p, "s")
			if !seen[stem] {
				seen[stem] = true
				out = append(out, stem)
			}
		}
	}
	for _, p := range parts {
		add(p)
	}
	// 无空格的中文词：也按整词加入（已在 Fields 空时处理）
	if len(parts) == 0 {
		add(keyword)
	}
	return out
}

func sortHitsByRelevance(hits []OnlineRecipeHit, keywords []string) {
	type scored struct {
		hit   OnlineRecipeHit
		score int
		idx   int
	}
	items := make([]scored, len(hits))
	for i, h := range hits {
		items[i] = scored{hit: h, score: scoreRecipeHit(h, keywords), idx: i}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score > items[j].score
		}
		return items[i].idx < items[j].idx
	})
	for i := range items {
		hits[i] = items[i].hit
	}
}

func needsContentEnrichment(hit OnlineRecipeHit, keywords []string) bool {
	if len(hit.Ingredients) > 0 || len(hit.ProcessSteps) > 0 {
		return false
	}
	return scoreRecipeHit(hit, keywords) < rankPartialTitle
}
