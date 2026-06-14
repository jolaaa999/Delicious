package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/delicious/delicious/pkg/model"
)

const spoonacularSource = "spoonacular"

type spoonacularProvider struct {
	client *http.Client
	apiKey string
}

func newSpoonacularProvider(apiKey string, client *http.Client) *spoonacularProvider {
	return &spoonacularProvider{client: client, apiKey: apiKey}
}

func (p *spoonacularProvider) Name() string { return spoonacularSource }

func (p *spoonacularProvider) Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error) {
	offset := (page - 1) * pageSize
	q := url.Values{}
	q.Set("query", keyword)
	q.Set("number", strconv.Itoa(pageSize))
	q.Set("offset", strconv.Itoa(offset))
	q.Set("addRecipeInformation", "false")
	q.Set("apiKey", p.apiKey)
	if containsChinese(keyword) {
		q.Set("cuisine", "Chinese")
	}
	body, err := p.get(ctx, "https://api.spoonacular.com/recipes/complexSearch?"+q.Encode())
	if err != nil {
		return nil, 0, err
	}
	var resp struct {
		Results      []spoonacularSearchItem `json:"results"`
		TotalResults int                     `json:"totalResults"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, err
	}
	if len(resp.Results) == 0 {
		return nil, resp.TotalResults, nil
	}

	hits := make([]OnlineRecipeHit, 0, len(resp.Results))
	for _, item := range resp.Results {
		detail, detailErr := p.Fetch(ctx, strconv.Itoa(item.ID))
		if detailErr != nil {
			hit := item.toBasicHit()
			hits = append(hits, hit)
			continue
		}
		hits = append(hits, *detail)
	}
	return hits, resp.TotalResults, nil
}

func (p *spoonacularProvider) Fetch(ctx context.Context, externalID string) (*OnlineRecipeHit, error) {
	rawURL := fmt.Sprintf(
		"https://api.spoonacular.com/recipes/%s/information?includeNutrition=false&apiKey=%s",
		url.PathEscape(externalID),
		url.QueryEscape(p.apiKey),
	)
	body, err := p.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	var detail spoonacularRecipeDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, err
	}
	hit := detail.toHit()
	return &hit, nil
}

func (p *spoonacularProvider) get(ctx context.Context, rawURL string) ([]byte, error) {
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
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("spoonacular: HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	return io.ReadAll(res.Body)
}

type spoonacularSearchItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Image string `json:"image"`
}

func (item spoonacularSearchItem) toBasicHit() OnlineRecipeHit {
	var cover *string
	if item.Image != "" {
		cover = &item.Image
	}
	return OnlineRecipeHit{
		Source:        spoonacularSource,
		ExternalID:    strconv.Itoa(item.ID),
		Name:          stripHTML(item.Title),
		CoverImageURL: cover,
	}
}

type spoonacularRecipeDetail struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Image       string `json:"image"`
	Cuisines    []string `json:"cuisines"`
	DishTypes   []string `json:"dishTypes"`
	Ingredients []struct {
		Name   string  `json:"name"`
		Amount float64 `json:"amount"`
		Unit   string  `json:"unit"`
	} `json:"extendedIngredients"`
	Instructions []struct {
		Steps []struct {
			Number int    `json:"number"`
			Step   string `json:"step"`
		} `json:"steps"`
	} `json:"analyzedInstructions"`
}

func (d spoonacularRecipeDetail) toHit() OnlineRecipeHit {
	ings := make([]model.Ingredient, 0, len(d.Ingredients))
	for _, ing := range d.Ingredients {
		name := strings.TrimSpace(ing.Name)
		if name == "" {
			continue
		}
		unit := strings.TrimSpace(ing.Unit)
		if unit == "" {
			unit = "份"
		}
		ings = append(ings, model.Ingredient{Name: name, Amount: ing.Amount, Unit: unit})
	}

	steps := make([]model.ProcessStep, 0)
	for _, block := range d.Instructions {
		for _, step := range block.Steps {
			content := stripHTML(strings.TrimSpace(step.Step))
			if content == "" {
				continue
			}
			order := step.Number
			if order <= 0 {
				order = len(steps) + 1
			}
			steps = append(steps, model.ProcessStep{Order: order, Content: content})
		}
	}
	if len(steps) == 0 {
		if summary := stripHTML(strings.TrimSpace(d.Summary)); summary != "" {
			steps = splitInstructions(summary)
		}
	}

	var category *string
	if len(d.Cuisines) > 0 {
		c := d.Cuisines[0]
		category = &c
	} else if len(d.DishTypes) > 0 {
		c := d.DishTypes[0]
		category = &c
	}

	var cover *string
	if d.Image != "" {
		cover = &d.Image
	}

	summary := stripHTML(strings.TrimSpace(d.Summary))
	var description *string
	if summary != "" {
		if len(summary) > 180 {
			short := summary[:180] + "…"
			description = &short
		} else {
			description = &summary
		}
	}

	tags := append([]string{}, d.Cuisines...)
	tags = append(tags, d.DishTypes...)

	return OnlineRecipeHit{
		Source:        spoonacularSource,
		ExternalID:    strconv.Itoa(d.ID),
		Name:          stripHTML(d.Title),
		Description:   description,
		CoverImageURL: cover,
		Category:      category,
		Tags:          tags,
		Ingredients:   ings,
		ProcessSteps:  steps,
	}
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func stripHTML(s string) string {
	s = htmlTagPattern.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return strings.TrimSpace(s)
}

func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}
