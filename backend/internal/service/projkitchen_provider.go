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
	"sync"
	"time"

	"github.com/delicious/delicious/pkg/model"
)

const projKitchenSource = "projkitchen"

type projKitchenProvider struct {
	client  *http.Client
	baseURL string

	mu        sync.Mutex
	index     []projKitchenIndexItem
	indexAt   time.Time
	indexTTL  time.Duration
}

func newProjKitchenProvider(client *http.Client) *projKitchenProvider {
	return &projKitchenProvider{
		client:   client,
		baseURL:  "https://proj.kitchen",
		indexTTL: time.Hour,
	}
}

func (p *projKitchenProvider) Name() string { return projKitchenSource }

func (p *projKitchenProvider) Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error) {
	index, err := p.loadIndex(ctx)
	if err != nil {
		return nil, 0, err
	}
	kw := strings.ToLower(strings.TrimSpace(keyword))
	matched := make([]projKitchenIndexItem, 0)
	for _, item := range index {
		name := strings.ToLower(item.Name)
		cat := strings.ToLower(item.Category)
		if kw == "" || strings.Contains(name, kw) || strings.Contains(cat, kw) {
			matched = append(matched, item)
		}
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	}
	start := (page - 1) * pageSize
	if start >= len(matched) {
		return nil, len(matched), nil
	}
	end := start + pageSize
	if end > len(matched) {
		end = len(matched)
	}
	hits := make([]OnlineRecipeHit, 0, end-start)
	for _, item := range matched[start:end] {
		hits = append(hits, item.toHit())
	}
	return hits, len(matched), nil
}

func (p *projKitchenProvider) Fetch(ctx context.Context, externalID string) (*OnlineRecipeHit, error) {
	u := fmt.Sprintf("%s/api/recipes/%s", p.baseURL, url.PathEscape(externalID))
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var detail projKitchenDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, err
	}
	if strings.TrimSpace(detail.ID) == "" && strings.TrimSpace(detail.Name) == "" {
		return nil, fmt.Errorf("projkitchen: recipe %s not found", externalID)
	}
	hit := detail.toHit()
	return &hit, nil
}

func (p *projKitchenProvider) loadIndex(ctx context.Context) ([]projKitchenIndexItem, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.index) > 0 && time.Since(p.indexAt) < p.indexTTL {
		return p.index, nil
	}
	body, err := p.get(ctx, p.baseURL+"/api/recipes")
	if err != nil {
		if len(p.index) > 0 {
			return p.index, nil
		}
		return nil, err
	}
	var items []projKitchenIndexItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}
	p.index = items
	p.indexAt = time.Now()
	return p.index, nil
}

func (p *projKitchenProvider) get(ctx context.Context, rawURL string) ([]byte, error) {
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
		return nil, fmt.Errorf("projkitchen: HTTP %d", res.StatusCode)
	}
	return io.ReadAll(res.Body)
}

type projKitchenIndexItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	Difficulty string `json:"difficulty"`
}

func (item projKitchenIndexItem) toHit() OnlineRecipeHit {
	var category *string
	if c := strings.TrimSpace(item.Category); c != "" {
		category = &c
	}
	var desc *string
	if d := strings.TrimSpace(item.Difficulty); d != "" {
		desc = &d
	}
	return OnlineRecipeHit{
		Source:      projKitchenSource,
		ExternalID:  item.ID,
		Name:        strings.TrimSpace(item.Name),
		Description: desc,
		Category:    category,
		Tags:        []string{item.Category, item.Difficulty},
	}
}

type projKitchenDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Difficulty  string `json:"difficulty"`
	Tips        string `json:"tips"`
	Ingredients []struct {
		Name   string `json:"name"`
		Amount string `json:"amount"`
	} `json:"ingredients"`
	Steps []string `json:"steps"`
}

func (d projKitchenDetail) toHit() OnlineRecipeHit {
	ings := make([]model.Ingredient, 0, len(d.Ingredients))
	for _, ing := range d.Ingredients {
		name := strings.TrimSpace(ing.Name)
		if name == "" {
			continue
		}
		amount, unit := parseChineseAmount(ing.Amount)
		ings = append(ings, model.Ingredient{Name: name, Amount: amount, Unit: unit})
	}
	steps := make([]model.ProcessStep, 0, len(d.Steps))
	for i, text := range d.Steps {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		steps = append(steps, model.ProcessStep{Order: i + 1, Content: text})
	}
	var category *string
	if c := strings.TrimSpace(d.Category); c != "" {
		category = &c
	}
	var desc *string
	if tip := strings.TrimSpace(d.Tips); tip != "" {
		desc = &tip
	} else if d.Difficulty != "" {
		desc = &d.Difficulty
	}
	return OnlineRecipeHit{
		Source:       projKitchenSource,
		ExternalID:   firstNonEmpty(d.ID, d.Name),
		Name:         strings.TrimSpace(d.Name),
		Description:  desc,
		Category:     category,
		Tags:         []string{d.Category, d.Difficulty},
		Ingredients:  ings,
		ProcessSteps: steps,
	}
}

var chineseAmountPattern = regexp.MustCompile(`^([\d./]+)\s*(.*)$`)

func parseChineseAmount(raw string) (float64, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, "份"
	}
	m := chineseAmountPattern.FindStringSubmatch(raw)
	if len(m) != 3 {
		return 0, raw
	}
	amount, err := strconv.ParseFloat(strings.ReplaceAll(m[1], "/", "."), 64)
	if err != nil {
		return 0, raw
	}
	unit := strings.TrimSpace(m[2])
	if unit == "" {
		unit = "份"
	}
	return amount, unit
}
