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

const mealDBSource = "themealdb"

type mealDBProvider struct {
	client  *http.Client
	baseURL string
}

func newMealDBProvider(client *http.Client) *mealDBProvider {
	return &mealDBProvider{
		client:  client,
		baseURL: "https://www.themealdb.com/api/json/v1/1",
	}
}

func (p *mealDBProvider) Name() string { return mealDBSource }

func (p *mealDBProvider) Search(ctx context.Context, keyword string, page, pageSize int) ([]OnlineRecipeHit, int, error) {
	u := fmt.Sprintf("%s/search.php?s=%s", p.baseURL, url.QueryEscape(keyword))
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, 0, err
	}
	var resp struct {
		Meals []mealDBMeal `json:"meals"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, err
	}
	if len(resp.Meals) == 0 {
		return nil, 0, nil
	}
	all := make([]OnlineRecipeHit, 0, len(resp.Meals))
	for _, meal := range resp.Meals {
		all = append(all, meal.toHit())
	}
	start := (page - 1) * pageSize
	if start >= len(all) {
		return nil, len(all), nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], len(all), nil
}

func (p *mealDBProvider) Fetch(ctx context.Context, externalID string) (*OnlineRecipeHit, error) {
	u := fmt.Sprintf("%s/lookup.php?i=%s", p.baseURL, url.QueryEscape(externalID))
	body, err := p.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Meals []mealDBMeal `json:"meals"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Meals) == 0 {
		return nil, fmt.Errorf("mealdb: recipe %s not found", externalID)
	}
	hit := resp.Meals[0].toHit()
	return &hit, nil
}

func (p *mealDBProvider) get(ctx context.Context, rawURL string) ([]byte, error) {
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
		return nil, fmt.Errorf("mealdb: HTTP %d", res.StatusCode)
	}
	return io.ReadAll(res.Body)
}

type mealDBMeal struct {
	IDMeal          string `json:"idMeal"`
	StrMeal         string `json:"strMeal"`
	StrMealThumb    string `json:"strMealThumb"`
	StrCategory     string `json:"strCategory"`
	StrArea         string `json:"strArea"`
	StrTags         string `json:"strTags"`
	StrInstructions string `json:"strInstructions"`
	StrIngredient1  string `json:"strIngredient1"`
	StrIngredient2  string `json:"strIngredient2"`
	StrIngredient3  string `json:"strIngredient3"`
	StrIngredient4  string `json:"strIngredient4"`
	StrIngredient5  string `json:"strIngredient5"`
	StrIngredient6  string `json:"strIngredient6"`
	StrIngredient7  string `json:"strIngredient7"`
	StrIngredient8  string `json:"strIngredient8"`
	StrIngredient9  string `json:"strIngredient9"`
	StrIngredient10 string `json:"strIngredient10"`
	StrIngredient11 string `json:"strIngredient11"`
	StrIngredient12 string `json:"strIngredient12"`
	StrIngredient13 string `json:"strIngredient13"`
	StrIngredient14 string `json:"strIngredient14"`
	StrIngredient15 string `json:"strIngredient15"`
	StrIngredient16 string `json:"strIngredient16"`
	StrIngredient17 string `json:"strIngredient17"`
	StrIngredient18 string `json:"strIngredient18"`
	StrIngredient19 string `json:"strIngredient19"`
	StrIngredient20 string `json:"strIngredient20"`
	StrMeasure1     string `json:"strMeasure1"`
	StrMeasure2     string `json:"strMeasure2"`
	StrMeasure3     string `json:"strMeasure3"`
	StrMeasure4     string `json:"strMeasure4"`
	StrMeasure5     string `json:"strMeasure5"`
	StrMeasure6     string `json:"strMeasure6"`
	StrMeasure7     string `json:"strMeasure7"`
	StrMeasure8     string `json:"strMeasure8"`
	StrMeasure9     string `json:"strMeasure9"`
	StrMeasure10    string `json:"strMeasure10"`
	StrMeasure11    string `json:"strMeasure11"`
	StrMeasure12    string `json:"strMeasure12"`
	StrMeasure13    string `json:"strMeasure13"`
	StrMeasure14    string `json:"strMeasure14"`
	StrMeasure15    string `json:"strMeasure15"`
	StrMeasure16    string `json:"strMeasure16"`
	StrMeasure17    string `json:"strMeasure17"`
	StrMeasure18    string `json:"strMeasure18"`
	StrMeasure19    string `json:"strMeasure19"`
	StrMeasure20    string `json:"strMeasure20"`
}

func (m mealDBMeal) toHit() OnlineRecipeHit {
	ings := make([]model.Ingredient, 0, 8)
	measures := []string{
		m.StrMeasure1, m.StrMeasure2, m.StrMeasure3, m.StrMeasure4, m.StrMeasure5,
		m.StrMeasure6, m.StrMeasure7, m.StrMeasure8, m.StrMeasure9, m.StrMeasure10,
		m.StrMeasure11, m.StrMeasure12, m.StrMeasure13, m.StrMeasure14, m.StrMeasure15,
		m.StrMeasure16, m.StrMeasure17, m.StrMeasure18, m.StrMeasure19, m.StrMeasure20,
	}
	names := []string{
		m.StrIngredient1, m.StrIngredient2, m.StrIngredient3, m.StrIngredient4, m.StrIngredient5,
		m.StrIngredient6, m.StrIngredient7, m.StrIngredient8, m.StrIngredient9, m.StrIngredient10,
		m.StrIngredient11, m.StrIngredient12, m.StrIngredient13, m.StrIngredient14, m.StrIngredient15,
		m.StrIngredient16, m.StrIngredient17, m.StrIngredient18, m.StrIngredient19, m.StrIngredient20,
	}
	for i, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		amount, unit := parseMeasure(measures[i])
		ings = append(ings, model.Ingredient{Name: name, Amount: amount, Unit: unit})
	}

	var category *string
	if cat := strings.TrimSpace(m.StrCategory); cat != "" {
		category = &cat
	}
	var cover *string
	if thumb := strings.TrimSpace(m.StrMealThumb); thumb != "" {
		cover = &thumb
	}
	desc := strings.TrimSpace(m.StrInstructions)
	var description *string
	if desc != "" {
		if len(desc) > 180 {
			short := desc[:180] + "…"
			description = &short
		} else {
			description = &desc
		}
	}
	tags := []string{}
	if area := strings.TrimSpace(m.StrArea); area != "" {
		tags = append(tags, area)
	}
	if rawTags := strings.TrimSpace(m.StrTags); rawTags != "" {
		for _, tag := range strings.Split(rawTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	return OnlineRecipeHit{
		Source:        mealDBSource,
		ExternalID:    m.IDMeal,
		Name:          strings.TrimSpace(m.StrMeal),
		Description:   description,
		CoverImageURL: cover,
		Category:      category,
		Tags:          tags,
		Ingredients:   ings,
		ProcessSteps:  splitInstructions(m.StrInstructions),
	}
}

var measurePattern = regexp.MustCompile(`^([\d./]+)\s*(.*)$`)

func parseMeasure(raw string) (float64, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, "份"
	}
	m := measurePattern.FindStringSubmatch(raw)
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

func splitInstructions(text string) []model.ProcessStep {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	parts := regexp.MustCompile(`\r?\n+`).Split(text, -1)
	steps := make([]model.ProcessStep, 0, len(parts))
	order := 1
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		steps = append(steps, model.ProcessStep{Order: order, Content: part})
		order++
	}
	if len(steps) == 0 {
		return []model.ProcessStep{{Order: 1, Content: text}}
	}
	return steps
}
