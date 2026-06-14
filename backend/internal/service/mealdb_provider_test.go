package service

import (
	"testing"

	"github.com/delicious/delicious/pkg/model"
)

func TestSplitInstructions(t *testing.T) {
	steps := splitInstructions("第一步\n\n第二步")
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0].Content != "第一步" || steps[1].Content != "第二步" {
		t.Fatalf("unexpected steps: %+v", steps)
	}
}

func TestParseMeasure(t *testing.T) {
	amount, unit := parseMeasure("500 g")
	if amount != 500 || unit != "g" {
		t.Fatalf("got %v %q", amount, unit)
	}
}

func TestMealDBMealToHit(t *testing.T) {
	hit := mealDBMeal{
		IDMeal:           "1",
		StrMeal:          "Test Dish",
		StrIngredient1:   "Salt",
		StrMeasure1:      "5 g",
		StrInstructions:  "Step one\nStep two",
	}.toHit()
	if hit.Name != "Test Dish" {
		t.Fatalf("name = %q", hit.Name)
	}
	if len(hit.Ingredients) != 1 || hit.Ingredients[0].Name != "Salt" {
		t.Fatalf("ingredients = %+v", hit.Ingredients)
	}
	if len(hit.ProcessSteps) != 2 {
		t.Fatalf("steps = %+v", hit.ProcessSteps)
	}
	if hit.Source != mealDBSource {
		t.Fatalf("source = %q", hit.Source)
	}
	_ = model.Ingredient{}
}
