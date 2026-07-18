package service

import "github.com/delicious/delicious/pkg/model"

// OnlineRecipeHit 联网搜索得到的菜谱快照。
type OnlineRecipeHit struct {
	Source        string
	ExternalID    string
	Name          string
	Description   *string
	CoverImageURL *string
	Category      *string
	Tags          []string
	Ingredients   []model.Ingredient
	ProcessSteps  []model.ProcessStep
}

func (h OnlineRecipeHit) toModel() model.EncyclopediaRecipe {
	src := h.Source
	extID := h.ExternalID
	sourceLabel := h.Source
	tags := model.StringSlice(h.Tags)
	if tags == nil {
		tags = model.StringSlice{}
	}
	ings := model.JSONSlice[model.Ingredient](h.Ingredients)
	if ings == nil {
		ings = model.JSONSlice[model.Ingredient]{}
	}
	steps := model.JSONSlice[model.ProcessStep](h.ProcessSteps)
	if steps == nil {
		steps = model.JSONSlice[model.ProcessStep]{}
	}
	return model.EncyclopediaRecipe{
		Name: h.Name,
		Description:    h.Description,
		CoverImageURL:  h.CoverImageURL,
		Category:       h.Category,
		Tags:           tags,
		Ingredients:    ings,
		ProcessSteps:   steps,
		Source:         &sourceLabel,
		ExternalSource: &src,
		ExternalID:     &extID,
	}
}
