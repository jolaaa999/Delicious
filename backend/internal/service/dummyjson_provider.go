package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/delicious/delicious/pkg/model"
)

const dummyJSONSource = "dummyjson"

type dummyJSONProvider struct {
	client  *http.Client
	baseURL string
}

func newDummyJSONProvider(client *http.Client) *dummyJSONProvider {
	return &dummyJSONProvider{
		client:  client,
		baseURL: "https://dummyjson.com",
	}
}

func (p *dummyJSONProvider) Name() string { return dummyJSONSource }

func (p *dummyJSONProvider) Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error) {
	skip := (page - 1) * pageSize
	if skip < 0 {
		skip = 0
	}
	u := fmt.Sprintf("%s/recipes/search?q=%s&limit=%d&skip=%d",
		p.baseURL, url.QueryEscape(keyword), pageSize, skip)
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, 0, err
	}
	var resp struct {
		Recipes []dummyJSONRecipe `json:"recipes"`
		Total   int               `json:"total"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, err
	}
	hits := make([]OnlineRecipeHit, 0, len(resp.Recipes))
	for _, r := range resp.Recipes {
		hits = append(hits, r.toHit())
	}
	return hits, resp.Total, nil
}

func (p *dummyJSONProvider) Fetch(ctx context.Context, externalID string) (*OnlineRecipeHit, error) {
	u := fmt.Sprintf("%s/recipes/%s", p.baseURL, url.PathEscape(externalID))
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var recipe dummyJSONRecipe
	if err := json.Unmarshal(body, &recipe); err != nil {
		return nil, err
	}
	if recipe.ID == 0 {
		return nil, fmt.Errorf("dummyjson: recipe %s not found", externalID)
	}
	hit := recipe.toHit()
	return &hit, nil
}

func (p *dummyJSONProvider) get(ctx context.Context, rawURL string) ([]byte, error) {
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
		return nil, fmt.Errorf("dummyjson: HTTP %d", res.StatusCode)
	}
	return io.ReadAll(res.Body)
}

type dummyJSONRecipe struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
	Cuisine      string   `json:"cuisine"`
	Tags         []string `json:"tags"`
	Image        string   `json:"image"`
	Difficulty   string   `json:"difficulty"`
	MealType     []string `json:"mealType"`
}

func (r dummyJSONRecipe) toHit() OnlineRecipeHit {
	ings := make([]model.Ingredient, 0, len(r.Ingredients))
	for _, raw := range r.Ingredients {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		ings = append(ings, model.Ingredient{Name: raw, Amount: 0, Unit: "份"})
	}
	steps := make([]model.ProcessStep, 0, len(r.Instructions))
	for i, text := range r.Instructions {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		steps = append(steps, model.ProcessStep{Order: i + 1, Content: text})
	}

	var cover *string
	if img := strings.TrimSpace(r.Image); img != "" {
		cover = &img
	}
	var category *string
	if c := strings.TrimSpace(r.Cuisine); c != "" {
		category = &c
	}
	var desc *string
	parts := make([]string, 0, 2)
	if d := strings.TrimSpace(r.Difficulty); d != "" {
		parts = append(parts, d)
	}
	if len(r.MealType) > 0 {
		parts = append(parts, strings.Join(r.MealType, "/"))
	}
	if len(parts) > 0 {
		s := strings.Join(parts, " · ")
		desc = &s
	}

	tags := append([]string{}, r.Tags...)
	if category != nil {
		tags = append(tags, *category)
	}

	return OnlineRecipeHit{
		Source:        dummyJSONSource,
		ExternalID:    strconv.Itoa(r.ID),
		Name:          strings.TrimSpace(r.Name),
		Description:   desc,
		CoverImageURL: cover,
		Category:      category,
		Tags:          tags,
		Ingredients:   ings,
		ProcessSteps:  steps,
	}
}
