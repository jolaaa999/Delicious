package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/delicious/delicious/pkg/model"
)

const howToCookSource = "howtocook"

type howToCookProvider struct {
	client  *http.Client
	baseURL string
}

func newHowToCookProvider(client *http.Client) *howToCookProvider {
	return &howToCookProvider{
		client:  client,
		baseURL: "https://api.howtocook.cn",
	}
}

func (p *howToCookProvider) Name() string { return howToCookSource }

func (p *howToCookProvider) Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	}
	// API 仅支持 limit，多取后本地分页
	limit := page * pageSize
	if limit > 80 {
		limit = 80
	}
	u := fmt.Sprintf("%s/search?q=%s&limit=%d&language=zh",
		p.baseURL, url.QueryEscape(keyword), limit)
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, 0, err
	}
	var resp struct {
		Results []howToCookListItem `json:"results"`
		Total   int                 `json:"total"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, err
	}
	all := make([]OnlineRecipeHit, 0, len(resp.Results))
	for _, item := range resp.Results {
		all = append(all, item.toHit())
	}
	start := (page - 1) * pageSize
	if start >= len(all) {
		return nil, maxInt(resp.Total, len(all)), nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	total := resp.Total
	if total == 0 {
		total = len(all)
	}
	return all[start:end], total, nil
}

func (p *howToCookProvider) Fetch(ctx context.Context, externalID string) (*OnlineRecipeHit, error) {
	u := p.detailURL(externalID)
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var detail howToCookDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, err
	}
	if strings.TrimSpace(detail.ID) == "" && strings.TrimSpace(detail.Name) == "" {
		return nil, fmt.Errorf("howtocook: recipe %s not found", externalID)
	}
	hit := detail.toHit()
	return &hit, nil
}

func (p *howToCookProvider) detailURL(id string) string {
	segs := strings.Split(id, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return p.baseURL + "/recipe/" + strings.Join(segs, "/")
}

func (p *howToCookProvider) get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("howtocook: HTTP %d", res.StatusCode)
	}
	return io.ReadAll(res.Body)
}

type howToCookListItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	Cuisine       string `json:"cuisine"`
	CookingMethod string `json:"cooking_method"`
	ImageURL      string `json:"image_url"`
	Difficulty    int    `json:"difficulty"`
}

func (item howToCookListItem) toHit() OnlineRecipeHit {
	var cover *string
	if img := strings.TrimSpace(item.ImageURL); img != "" {
		cover = &img
	}
	var category *string
	if c := howToCookCategoryLabel(item.Category, item.Cuisine); c != "" {
		category = &c
	}
	var desc *string
	parts := make([]string, 0, 2)
	if item.Cuisine != "" {
		parts = append(parts, item.Cuisine)
	}
	if item.CookingMethod != "" {
		parts = append(parts, item.CookingMethod)
	}
	if len(parts) > 0 {
		s := strings.Join(parts, " · ")
		desc = &s
	}
	tags := []string{}
	if item.Cuisine != "" {
		tags = append(tags, item.Cuisine)
	}
	if item.CookingMethod != "" {
		tags = append(tags, item.CookingMethod)
	}
	return OnlineRecipeHit{
		Source:        howToCookSource,
		ExternalID:    item.ID,
		Name:          strings.TrimSpace(item.Name),
		Description:   desc,
		CoverImageURL: cover,
		Category:      category,
		Tags:          tags,
	}
}

type howToCookDetail struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Cuisine         string   `json:"cuisine"`
	CookingMethod   string   `json:"cooking_method"`
	ImageURL        string   `json:"image_url"`
	Introduction    string   `json:"introduction"`
	Ingredients     []string `json:"ingredients"`
	MainIngredients []string `json:"main_ingredients"`
	Steps           []string `json:"steps"`
	Difficulty      int      `json:"difficulty"`
}

func (d howToCookDetail) toHit() OnlineRecipeHit {
	ings := make([]model.Ingredient, 0, len(d.Ingredients))
	for _, name := range d.Ingredients {
		name = cleanHowToCookText(name)
		if name == "" {
			continue
		}
		ings = append(ings, model.Ingredient{Name: name, Amount: 0, Unit: "适量"})
	}
	steps := make([]model.ProcessStep, 0, len(d.Steps))
	for i, text := range d.Steps {
		text = cleanHowToCookText(text)
		if text == "" {
			continue
		}
		steps = append(steps, model.ProcessStep{Order: i + 1, Content: text})
	}
	var cover *string
	if img := strings.TrimSpace(d.ImageURL); img != "" {
		cover = &img
	}
	var category *string
	if c := howToCookCategoryLabel(d.Category, d.Cuisine); c != "" {
		category = &c
	}
	var desc *string
	if intro := strings.TrimSpace(d.Introduction); intro != "" {
		if len([]rune(intro)) > 180 {
			short := string([]rune(intro)[:180]) + "…"
			desc = &short
		} else {
			desc = &intro
		}
	}
	tags := append([]string{}, d.MainIngredients...)
	if d.Cuisine != "" {
		tags = append(tags, d.Cuisine)
	}
	return OnlineRecipeHit{
		Source:        howToCookSource,
		ExternalID:    firstNonEmpty(d.ID, d.Name),
		Name:          strings.TrimSpace(d.Name),
		Description:   desc,
		CoverImageURL: cover,
		Category:      category,
		Tags:          tags,
		Ingredients:   ings,
		ProcessSteps:  steps,
	}
}

func howToCookCategoryLabel(category, cuisine string) string {
	labels := map[string]string{
		"meat_dish":     "荤菜",
		"vegetable_dish": "素菜",
		"staple":        "主食",
		"soup":          "汤粥",
		"breakfast":     "早餐",
		"aquatic":       "水产",
		"dessert":       "甜品",
		"drink":         "饮品",
		"condiment":     "酱料",
		"semi-finished": "半成品",
	}
	if v, ok := labels[category]; ok {
		return v
	}
	if cuisine != "" {
		return cuisine
	}
	return category
}

func cleanHowToCookText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "**", "")
	return strings.TrimSpace(s)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
