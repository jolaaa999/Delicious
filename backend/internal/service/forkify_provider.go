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

const forkifySource = "forkify"

type forkifyProvider struct {
	client  *http.Client
	baseURL string
}

func newForkifyProvider(client *http.Client) *forkifyProvider {
	return &forkifyProvider{
		client:  client,
		baseURL: "https://forkify-api.herokuapp.com/api/v2",
	}
}

func (p *forkifyProvider) Name() string { return forkifySource }

func (p *forkifyProvider) Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error) {
	u := fmt.Sprintf("%s/recipes?search=%s", p.baseURL, url.QueryEscape(keyword))
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, 0, err
	}
	var resp struct {
		Results int `json:"results"`
		Data    struct {
			Recipes []forkifyListItem `json:"recipes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, err
	}
	if len(resp.Data.Recipes) == 0 {
		return nil, 0, nil
	}
	all := make([]OnlineRecipeHit, 0, len(resp.Data.Recipes))
	for _, item := range resp.Data.Recipes {
		all = append(all, item.toHit())
	}
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= len(all) {
		return nil, len(all), nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	total := resp.Results
	if total == 0 {
		total = len(all)
	}
	return all[start:end], total, nil
}

func (p *forkifyProvider) Fetch(ctx context.Context, externalID string) (*OnlineRecipeHit, error) {
	u := fmt.Sprintf("%s/recipes/%s", p.baseURL, url.PathEscape(externalID))
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data struct {
			Recipe forkifyDetail `json:"recipe"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.Data.Recipe.ID) == "" {
		return nil, fmt.Errorf("forkify: recipe %s not found", externalID)
	}
	hit := resp.Data.Recipe.toHit()
	return &hit, nil
}

func (p *forkifyProvider) get(ctx context.Context, rawURL string) ([]byte, error) {
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
		return nil, fmt.Errorf("forkify: HTTP %d", res.StatusCode)
	}
	return io.ReadAll(res.Body)
}

type forkifyListItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
	ImageURL  string `json:"image_url"`
}

func (item forkifyListItem) toHit() OnlineRecipeHit {
	var cover *string
	if img := httpsURL(strings.TrimSpace(item.ImageURL)); img != "" {
		cover = &img
	}
	var desc *string
	if pub := strings.TrimSpace(item.Publisher); pub != "" {
		d := "来自 " + pub
		desc = &d
	}
	return OnlineRecipeHit{
		Source:        forkifySource,
		ExternalID:    item.ID,
		Name:          strings.TrimSpace(item.Title),
		Description:   desc,
		CoverImageURL: cover,
		Tags:          []string{strings.TrimSpace(item.Publisher)},
	}
}

type forkifyDetail struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Publisher    string `json:"publisher"`
	ImageURL     string `json:"image_url"`
	CookingTime  int    `json:"cooking_time"`
	Servings     int    `json:"servings"`
	Ingredients  []struct {
		Quantity    *float64 `json:"quantity"`
		Unit        string   `json:"unit"`
		Description string   `json:"description"`
	} `json:"ingredients"`
}

func (d forkifyDetail) toHit() OnlineRecipeHit {
	ings := make([]model.Ingredient, 0, len(d.Ingredients))
	for _, ing := range d.Ingredients {
		name := strings.TrimSpace(ing.Description)
		if name == "" {
			continue
		}
		amount := 0.0
		if ing.Quantity != nil {
			amount = *ing.Quantity
		}
		unit := strings.TrimSpace(ing.Unit)
		if unit == "" {
			if amount == 0 {
				unit = "适量"
			} else {
				unit = "份"
			}
		}
		ings = append(ings, model.Ingredient{Name: name, Amount: amount, Unit: unit})
	}

	var cover *string
	if img := httpsURL(strings.TrimSpace(d.ImageURL)); img != "" {
		cover = &img
	}
	var desc *string
	parts := make([]string, 0, 2)
	if pub := strings.TrimSpace(d.Publisher); pub != "" {
		parts = append(parts, "来自 "+pub)
	}
	if d.CookingTime > 0 {
		parts = append(parts, "约 "+strconv.Itoa(d.CookingTime)+" 分钟")
	}
	if len(parts) > 0 {
		s := strings.Join(parts, " · ")
		desc = &s
	}

	steps := []model.ProcessStep{}
	if d.Servings > 0 {
		steps = append(steps, model.ProcessStep{
			Order:   1,
			Content: fmt.Sprintf("建议份量：%d 人份。按配料准备食材后，参考原菜谱完成烹饪。", d.Servings),
		})
	}

	return OnlineRecipeHit{
		Source:        forkifySource,
		ExternalID:    d.ID,
		Name:          strings.TrimSpace(d.Title),
		Description:   desc,
		CoverImageURL: cover,
		Tags:          []string{strings.TrimSpace(d.Publisher)},
		Ingredients:   ings,
		ProcessSteps:  steps,
	}
}

func httpsURL(raw string) string {
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") {
		return "https://" + strings.TrimPrefix(raw, "http://")
	}
	return raw
}
