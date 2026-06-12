package diff

import (
	"fmt"
	"math"
	"strings"

	"github.com/delicious/delicious/pkg/model"
)

const amountEpsilon = 1e-6

// Compare 对比 base（左/旧/百科）与 target（右/新/我的），返回结构化 Diff。
// 配料按 name 归一化匹配；步骤按 order 匹配。时间复杂度 O(n+m)。
func Compare(base, target VersionSnapshot) VersionDiffResult {
	ingDiffs := diffIngredients(base.Ingredients, target.Ingredients)
	procDiffs := diffProcessSteps(base.ProcessSteps, target.ProcessSteps)

	return VersionDiffResult{
		IngredientDiffs: ingDiffs,
		ProcessDiffs:    procDiffs,
		Summary:         buildSummary(ingDiffs, procDiffs),
	}
}

func diffIngredients(base, target []model.Ingredient) []IngredientDiff {
	baseMap := indexIngredients(base)
	targetMap := indexIngredients(target)

	seen := make(map[string]struct{}, len(baseMap)+len(targetMap))
	var result []IngredientDiff

	// 按 target 顺序输出，便于前端右栏对齐
	for _, t := range target {
		key := ingredientKey(t.Name)
		seen[key] = struct{}{}
		b, ok := baseMap[key]
		if !ok {
			tCopy := t
			result = append(result, IngredientDiff{
				Type:   TypeAdded,
				Target: &tCopy,
			})
			continue
		}
		result = append(result, compareIngredientPair(b, t))
	}

	// base 中剩余项为 REMOVED，保持 base 原顺序
	for _, b := range base {
		key := ingredientKey(b.Name)
		if _, ok := seen[key]; ok {
			continue
		}
		bCopy := b
		result = append(result, IngredientDiff{
			Type: TypeRemoved,
			Base: &bCopy,
		})
	}

	return result
}

func compareIngredientPair(base, target model.Ingredient) IngredientDiff {
	if ingredientsEqual(base, target) {
		bCopy, tCopy := base, target
		return IngredientDiff{
			Type:   TypeUnchanged,
			Base:   &bCopy,
			Target: &tCopy,
		}
	}
	delta := target.Amount - base.Amount
	bCopy, tCopy := base, target
	return IngredientDiff{
		Type:        TypeModified,
		Base:        &bCopy,
		Target:      &tCopy,
		AmountDelta: &delta,
	}
}

func ingredientsEqual(a, b model.Ingredient) bool {
	return ingredientKey(a.Name) == ingredientKey(b.Name) &&
		floatEqual(a.Amount, b.Amount) &&
		normalizeUnit(a.Unit) == normalizeUnit(b.Unit) &&
		strings.TrimSpace(a.Note) == strings.TrimSpace(b.Note)
}

func diffProcessSteps(base, target []model.ProcessStep) []ProcessStepDiff {
	baseMap := indexProcessSteps(base)
	targetMap := indexProcessSteps(target)

	seen := make(map[int]struct{}, len(baseMap)+len(targetMap))
	var result []ProcessStepDiff

	for _, t := range target {
		seen[t.Order] = struct{}{}
		b, ok := baseMap[t.Order]
		if !ok {
			tCopy := t
			result = append(result, ProcessStepDiff{
				Type:   TypeAdded,
				Order:  t.Order,
				Target: &tCopy,
			})
			continue
		}
		result = append(result, compareProcessPair(b, t))
	}

	for _, b := range base {
		if _, ok := seen[b.Order]; ok {
			continue
		}
		bCopy := b
		result = append(result, ProcessStepDiff{
			Type:  TypeRemoved,
			Order: b.Order,
			Base:  &bCopy,
		})
	}

	return result
}

func compareProcessPair(base, target model.ProcessStep) ProcessStepDiff {
	if processStepsEqual(base, target) {
		bCopy, tCopy := base, target
		return ProcessStepDiff{
			Type:   TypeUnchanged,
			Order:  base.Order,
			Base:   &bCopy,
			Target: &tCopy,
		}
	}
	bCopy, tCopy := base, target
	return ProcessStepDiff{
		Type:   TypeModified,
		Order:  base.Order,
		Base:   &bCopy,
		Target: &tCopy,
	}
}

func processStepsEqual(a, b model.ProcessStep) bool {
	if a.Order != b.Order {
		return false
	}
	if strings.TrimSpace(a.Content) != strings.TrimSpace(b.Content) {
		return false
	}
	if !intPtrEqual(a.DurationMinutes, b.DurationMinutes) {
		return false
	}
	return strPtrEqual(a.ImageURL, b.ImageURL)
}

// indexIngredients 按归一化 name 建索引；同名重复时保留首次出现。
func indexIngredients(items []model.Ingredient) map[string]model.Ingredient {
	m := make(map[string]model.Ingredient, len(items))
	for _, item := range items {
		key := ingredientKey(item.Name)
		if _, exists := m[key]; !exists {
			m[key] = item
		}
	}
	return m
}

func indexProcessSteps(items []model.ProcessStep) map[int]model.ProcessStep {
	m := make(map[int]model.ProcessStep, len(items))
	for _, item := range items {
		if _, exists := m[item.Order]; !exists {
			m[item.Order] = item
		}
	}
	return m
}

func ingredientKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeUnit(unit string) string {
	u := strings.TrimSpace(unit)
	// 常见单位归一化，减少「克」vs「g」误判为修改
	switch strings.ToLower(u) {
	case "g", "克":
		return "g"
	case "kg", "千克", "公斤":
		return "kg"
	case "ml", "毫升":
		return "ml"
	case "l", "升":
		return "l"
	default:
		return u
	}
}

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < amountEpsilon
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func strPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func buildSummary(ingDiffs []IngredientDiff, procDiffs []ProcessStepDiff) string {
	var addedIng, removedIng, modifiedIng int
	var addedProc, removedProc, modifiedProc int

	for _, d := range ingDiffs {
		switch d.Type {
		case TypeAdded:
			addedIng++
		case TypeRemoved:
			removedIng++
		case TypeModified:
			modifiedIng++
		}
	}
	for _, d := range procDiffs {
		switch d.Type {
		case TypeAdded:
			addedProc++
		case TypeRemoved:
			removedProc++
		case TypeModified:
			modifiedProc++
		}
	}

	if addedIng+removedIng+modifiedIng+addedProc+removedProc+modifiedProc == 0 {
		return "两个版本完全一致"
	}

	var parts []string
	if addedIng > 0 {
		parts = append(parts, fmt.Sprintf("新增 %d 项配料", addedIng))
	}
	if removedIng > 0 {
		parts = append(parts, fmt.Sprintf("删除 %d 项配料", removedIng))
	}
	if modifiedIng > 0 {
		parts = append(parts, fmt.Sprintf("修改 %d 项配料", modifiedIng))
	}
	if addedProc > 0 {
		parts = append(parts, fmt.Sprintf("新增 %d 个步骤", addedProc))
	}
	if removedProc > 0 {
		parts = append(parts, fmt.Sprintf("删除 %d 个步骤", removedProc))
	}
	if modifiedProc > 0 {
		parts = append(parts, fmt.Sprintf("修改 %d 个步骤", modifiedProc))
	}

	return strings.Join(parts, "，")
}

// CountIngredientByType 统计配料差异数量
func CountIngredientByType(diffs []IngredientDiff, t Type) int {
	n := 0
	for _, d := range diffs {
		if d.Type == t {
			n++
		}
	}
	return n
}

// CountProcessByType 统计步骤差异数量
func CountProcessByType(diffs []ProcessStepDiff, t Type) int {
	n := 0
	for _, d := range diffs {
		if d.Type == t {
			n++
		}
	}
	return n
}
