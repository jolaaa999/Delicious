package diff

import "github.com/delicious/delicious/pkg/model"

// Type 差异类型（与 proto DiffType 对应）
type Type int

const (
	TypeUnchanged Type = iota + 1
	TypeAdded
	TypeRemoved
	TypeModified
)

func (t Type) String() string {
	switch t {
	case TypeUnchanged:
		return "unchanged"
	case TypeAdded:
		return "added"
	case TypeRemoved:
		return "removed"
	case TypeModified:
		return "modified"
	default:
		return "unknown"
	}
}

// IngredientDiff 单条配料差异
type IngredientDiff struct {
	Type        Type             `json:"type"`
	Base        *model.Ingredient `json:"base,omitempty"`
	Target      *model.Ingredient `json:"target,omitempty"`
	AmountDelta *float64         `json:"amount_delta,omitempty"`
}

// ProcessStepDiff 单条步骤差异
type ProcessStepDiff struct {
	Type   Type              `json:"type"`
	Order  int               `json:"order"`
	Base   *model.ProcessStep `json:"base,omitempty"`
	Target *model.ProcessStep `json:"target,omitempty"`
}

// VersionDiffResult 完整对比结果
type VersionDiffResult struct {
	IngredientDiffs []IngredientDiff  `json:"ingredient_diffs"`
	ProcessDiffs    []ProcessStepDiff `json:"process_diffs"`
	Summary         string            `json:"summary"`
}

// VersionSnapshot 参与对比的版本快照（与 DB / RPC 解耦）
type VersionSnapshot struct {
	Ingredients  []model.Ingredient
	ProcessSteps []model.ProcessStep
}
