package diff

import "github.com/delicious/delicious/pkg/model"

// FromRecipeVersion 从 DB 版本记录构建快照
func FromRecipeVersion(v model.RecipeVersion) VersionSnapshot {
	ings := make([]model.Ingredient, len(v.Ingredients))
	copy(ings, v.Ingredients)
	steps := make([]model.ProcessStep, len(v.ProcessSteps))
	copy(steps, v.ProcessSteps)
	return VersionSnapshot{
		Ingredients:  ings,
		ProcessSteps: steps,
	}
}

// FromEncyclopedia 从百科记录构建快照（基准对比）
func FromEncyclopedia(e model.EncyclopediaRecipe) VersionSnapshot {
	ings := make([]model.Ingredient, len(e.Ingredients))
	copy(ings, e.Ingredients)
	steps := make([]model.ProcessStep, len(e.ProcessSteps))
	copy(steps, e.ProcessSteps)
	return VersionSnapshot{
		Ingredients:  ings,
		ProcessSteps: steps,
	}
}
