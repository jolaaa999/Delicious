package service

import (
	"testing"

	"github.com/delicious/delicious/pkg/model"
)

func TestScoreRecipeHitRanking(t *testing.T) {
	keywords := []string{"番茄炒蛋", "tomato scrambled eggs"}

	exact := OnlineRecipeHit{Name: "Tomato Scrambled Eggs"}
	partialTitle := OnlineRecipeHit{Name: "Tomato Egg Stir Fry"}
	contentOnly := OnlineRecipeHit{
		Name: "Home Style Breakfast",
		Ingredients: []model.Ingredient{
			{Name: "tomato", Amount: 2, Unit: "个"},
			{Name: "egg", Amount: 3, Unit: "个"},
		},
	}
	unrelated := OnlineRecipeHit{Name: "Beef Stew"}

	if scoreRecipeHit(exact, keywords) != rankExactTitle {
		t.Fatalf("exact score = %d", scoreRecipeHit(exact, keywords))
	}
	if scoreRecipeHit(partialTitle, keywords) != rankPartialTitle {
		t.Fatalf("partial title score = %d", scoreRecipeHit(partialTitle, keywords))
	}
	if scoreRecipeHit(contentOnly, keywords) != rankPartialContent {
		t.Fatalf("content score = %d", scoreRecipeHit(contentOnly, keywords))
	}
	if scoreRecipeHit(unrelated, keywords) != 0 {
		t.Fatalf("unrelated score = %d", scoreRecipeHit(unrelated, keywords))
	}
}

func TestSortHitsByRelevance(t *testing.T) {
	keywords := []string{"pasta"}
	hits := []OnlineRecipeHit{
		{Name: "Chicken Soup", ProcessSteps: []model.ProcessStep{{Order: 1, Content: "boil pasta water"}}},
		{Name: "Creamy Pasta Bake"},
		{Name: "Pasta"},
		{Name: "Steak"},
	}
	sortHitsByRelevance(hits, keywords)
	if hits[0].Name != "Pasta" {
		t.Fatalf("want exact first, got %q", hits[0].Name)
	}
	if hits[1].Name != "Creamy Pasta Bake" {
		t.Fatalf("want partial title second, got %q", hits[1].Name)
	}
	if hits[2].Name != "Chicken Soup" {
		t.Fatalf("want content third, got %q", hits[2].Name)
	}
}
