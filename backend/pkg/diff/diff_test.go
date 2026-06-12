package diff

import (
	"testing"

	"github.com/delicious/delicious/pkg/model"
)

func ptrInt(v int) *int       { return &v }
func ptrStr(v string) *string { return &v }

func TestCompare_Unchanged(t *testing.T) {
	ing := []model.Ingredient{{Name: "生抽", Amount: 2, Unit: "勺"}}
	steps := []model.ProcessStep{{Order: 1, Content: "切肉"}}

	result := Compare(
		VersionSnapshot{Ingredients: ing, ProcessSteps: steps},
		VersionSnapshot{Ingredients: ing, ProcessSteps: steps},
	)

	if result.Summary != "两个版本完全一致" {
		t.Fatalf("summary = %q", result.Summary)
	}
	if len(result.IngredientDiffs) != 1 || result.IngredientDiffs[0].Type != TypeUnchanged {
		t.Fatalf("ingredient diff: %+v", result.IngredientDiffs)
	}
}

func TestCompare_IngredientAddedRemovedModified(t *testing.T) {
	base := []model.Ingredient{
		{Name: "猪肉", Amount: 500, Unit: "g"},
		{Name: "盐", Amount: 3, Unit: "g"},
	}
	target := []model.Ingredient{
		{Name: "猪肉", Amount: 400, Unit: "g"}, // modified
		{Name: "生抽", Amount: 2, Unit: "勺"},    // added
	}

	result := Compare(
		VersionSnapshot{Ingredients: base},
		VersionSnapshot{Ingredients: target},
	)

	byType := map[Type]int{}
	for _, d := range result.IngredientDiffs {
		byType[d.Type]++
	}
	if byType[TypeModified] != 1 || byType[TypeAdded] != 1 || byType[TypeRemoved] != 1 {
		t.Fatalf("counts = %+v, diffs = %+v", byType, result.IngredientDiffs)
	}

	var porkDiff *IngredientDiff
	for i := range result.IngredientDiffs {
		if result.IngredientDiffs[i].Type == TypeModified {
			porkDiff = &result.IngredientDiffs[i]
			break
		}
	}
	if porkDiff == nil || porkDiff.AmountDelta == nil || *porkDiff.AmountDelta != -100 {
		t.Fatalf("pork delta: %+v", porkDiff)
	}
}

func TestCompare_UnitNormalization(t *testing.T) {
	base := []model.Ingredient{{Name: "面粉", Amount: 200, Unit: "克"}}
	target := []model.Ingredient{{Name: "面粉", Amount: 200, Unit: "g"}}

	result := Compare(
		VersionSnapshot{Ingredients: base},
		VersionSnapshot{Ingredients: target},
	)
	if result.IngredientDiffs[0].Type != TypeUnchanged {
		t.Fatalf("expected unchanged for 克/g, got %v", result.IngredientDiffs[0].Type)
	}
}

func TestCompare_ProcessSteps(t *testing.T) {
	base := []model.ProcessStep{
		{Order: 1, Content: "焯水"},
		{Order: 2, Content: "炒制"},
	}
	target := []model.ProcessStep{
		{Order: 1, Content: "焯水"},
		{Order: 2, Content: "大火炒制", DurationMinutes: ptrInt(10)},
		{Order: 3, Content: "出锅"},
	}

	result := Compare(
		VersionSnapshot{ProcessSteps: base},
		VersionSnapshot{ProcessSteps: target},
	)

	byType := map[Type]int{}
	for _, d := range result.ProcessDiffs {
		byType[d.Type]++
	}
	if byType[TypeUnchanged] != 1 || byType[TypeModified] != 1 ||
		byType[TypeAdded] != 1 || byType[TypeRemoved] != 0 {
		t.Fatalf("process counts = %+v, diffs = %+v", byType, result.ProcessDiffs)
	}
}

func TestCompare_EncyclopediaScenario(t *testing.T) {
	encyclopedia := VersionSnapshot{
		Ingredients: []model.Ingredient{
			{Name: "五花肉", Amount: 500, Unit: "g"},
			{Name: "冰糖", Amount: 30, Unit: "g"},
		},
		ProcessSteps: []model.ProcessStep{
			{Order: 1, Content: "切块焯水"},
		},
	}
	mine := VersionSnapshot{
		Ingredients: []model.Ingredient{
			{Name: "五花肉", Amount: 600, Unit: "g"},
			{Name: "冰糖", Amount: 20, Unit: "g"},
		},
		ProcessSteps: []model.ProcessStep{
			{Order: 1, Content: "切块焯水"},
			{Order: 2, Content: "炒糖色"},
		},
	}

	result := Compare(encyclopedia, mine)
	if result.Summary == "两个版本完全一致" {
		t.Fatal("expected differences")
	}
	if CountIngredientByType(result.IngredientDiffs, TypeModified) != 2 {
		t.Fatalf("ingredient diffs: %+v", result.IngredientDiffs)
	}
	if CountProcessByType(result.ProcessDiffs, TypeAdded) != 1 {
		t.Fatalf("process diffs: %+v", result.ProcessDiffs)
	}
}

func TestIngredientKey_CaseInsensitive(t *testing.T) {
	base := []model.Ingredient{{Name: "生抽", Amount: 1, Unit: "勺"}}
	target := []model.Ingredient{{Name: " 生抽 ", Amount: 1, Unit: "勺"}}

	result := Compare(
		VersionSnapshot{Ingredients: base},
		VersionSnapshot{Ingredients: target},
	)
	if result.IngredientDiffs[0].Type != TypeUnchanged {
		t.Fatalf("expected unchanged, got %v", result.IngredientDiffs[0].Type)
	}
}
