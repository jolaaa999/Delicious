package service

import (
	"errors"
	"math"

	"github.com/delicious/delicious/internal/dto"
	"github.com/delicious/delicious/internal/repository"
	"github.com/delicious/delicious/pkg/diff"
	"github.com/delicious/delicious/pkg/model"
)

type RecipeService struct {
	repo *repository.RecipeRepository
	ency *repository.EncyclopediaRepository
}

func NewRecipeService(repo *repository.RecipeRepository, ency *repository.EncyclopediaRepository) *RecipeService {
	return &RecipeService{repo: repo, ency: ency}
}

func (s *RecipeService) Create(userID uint64, req dto.CreateRecipeRequest) (*dto.MyRecipeDTO, error) {
	recipe := &model.MyRecipe{
		UserID:               userID,
		Name:                 req.Name,
		UserRating:           req.UserRating,
		EncyclopediaRecipeID: req.EncyclopediaRecipeID,
	}
	msg := req.CommitMsg
	if msg == "" {
		msg = "初次记录"
	}
	version := &model.RecipeVersion{
		Ingredients:  model.JSONSlice[model.Ingredient](req.Ingredients),
		ProcessSteps: model.JSONSlice[model.ProcessStep](req.ProcessSteps),
		ProcessText:  req.ProcessText,
		Images:       model.StringSlice(req.Images),
		CommitMsg:    msg,
	}
	if err := s.repo.CreateWithVersion(recipe, version); err != nil {
		return nil, err
	}
	out := dto.ToRecipeDTO(recipe, version)
	return &out, nil
}

func (s *RecipeService) Get(id, userID uint64) (*dto.MyRecipeDTO, error) {
	recipe, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	var ver *model.RecipeVersion
	if recipe.CurrentVersionID != nil {
		ver, err = s.repo.GetVersion(*recipe.CurrentVersionID)
		if err != nil {
			return nil, err
		}
	}
	out := dto.ToRecipeDTO(recipe, ver)
	return &out, nil
}

func (s *RecipeService) List(filter repository.ListRecipesFilter) ([]dto.RecipeListItemDTO, dto.PageInfo, error) {
	items, total, err := s.repo.List(filter)
	if err != nil {
		return nil, dto.PageInfo{}, err
	}

	// 批量获取版本号，避免 N+1 查询
	var versionIDs []uint64
	for _, item := range items {
		if item.CurrentVersionID != nil {
			versionIDs = append(versionIDs, *item.CurrentVersionID)
		}
	}
	versionMap, err := s.repo.GetVersionsByIDs(versionIDs)
	if err != nil {
		return nil, dto.PageInfo{}, err
	}

	result := make([]dto.RecipeListItemDTO, 0, len(items))
	for _, item := range items {
		d := dto.RecipeListItemDTO{
			ID:            item.ID,
			Name:          item.Name,
			CoverImageURL: item.CoverImageURL,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		}
		if item.UserRating != nil {
			d.UserRating = *item.UserRating
		}
		if item.CurrentVersionID != nil {
			if ver, ok := versionMap[*item.CurrentVersionID]; ok {
				d.CurrentVersionNumber = ver.VersionNumber
			}
		}
		result = append(result, d)
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	return result, dto.PageInfo{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: pages,
	}, nil
}

func (s *RecipeService) Update(id, userID uint64, req dto.UpdateRecipeRequest) (*dto.MyRecipeDTO, error) {
	recipe, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	recipe.Name = req.Name
	recipe.UserRating = req.UserRating
	recipe.EncyclopediaRecipeID = req.EncyclopediaRecipeID

	version := &model.RecipeVersion{
		Ingredients:  model.JSONSlice[model.Ingredient](req.Ingredients),
		ProcessSteps: model.JSONSlice[model.ProcessStep](req.ProcessSteps),
		ProcessText:  req.ProcessText,
		Images:       model.StringSlice(req.Images),
		CommitMsg:    req.CommitMsg,
	}
	if err := s.repo.AddVersion(recipe, version); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	out := dto.ToRecipeDTO(updated, version)
	return &out, nil
}

func (s *RecipeService) Delete(id, userID uint64) error {
	return s.repo.SoftDelete(id, userID)
}

func (s *RecipeService) ListVersions(recipeID, userID uint64) ([]dto.VersionListItemDTO, error) {
	versions, err := s.repo.ListVersions(recipeID, userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.VersionListItemDTO, 0, len(versions))
	for _, v := range versions {
		result = append(result, dto.VersionListItemDTO{
			ID:            v.ID,
			VersionNumber: v.VersionNumber,
			CommitMsg:     v.CommitMsg,
			CreatedAt:     v.CreatedAt,
		})
	}
	return result, nil
}

func (s *RecipeService) GetVersion(recipeID, versionID, userID uint64) (*dto.RecipeVersionDTO, error) {
	ver, err := s.repo.GetVersionByRecipe(recipeID, versionID, userID)
	if err != nil {
		return nil, err
	}
	out := dto.ToVersionDTO(ver)
	return &out, nil
}

func (s *RecipeService) CompareVersions(recipeID, userID, baseID, targetID uint64) (
	*dto.RecipeVersionDTO, *dto.RecipeVersionDTO, *dto.VersionDiffResultDTO, error,
) {
	if _, err := s.repo.GetByID(recipeID, userID); err != nil {
		return nil, nil, nil, err
	}
	base, err := s.repo.GetVersion(baseID)
	if err != nil || base.RecipeID != recipeID {
		return nil, nil, nil, repository.ErrNotFound
	}
	target, err := s.repo.GetVersion(targetID)
	if err != nil || target.RecipeID != recipeID {
		return nil, nil, nil, repository.ErrNotFound
	}

	result := diff.Compare(diff.FromRecipeVersion(*base), diff.FromRecipeVersion(*target))
	bDTO := dto.ToVersionDTO(base)
	tDTO := dto.ToVersionDTO(target)
	dDTO := toDiffDTO(result)
	return &bDTO, &tDTO, &dDTO, nil
}

func (s *RecipeService) CompareWithEncyclopedia(recipeID, userID uint64, encyclopediaID *uint64) (
	*dto.EncyclopediaRecipeDTO, *dto.RecipeVersionDTO, *dto.VersionDiffResultDTO, error,
) {
	recipe, err := s.repo.GetByID(recipeID, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	if recipe.CurrentVersionID == nil {
		return nil, nil, nil, errors.New("recipe has no current version")
	}
	myVer, err := s.repo.GetVersion(*recipe.CurrentVersionID)
	if err != nil {
		return nil, nil, nil, err
	}

	var ency *model.EncyclopediaRecipe
	if encyclopediaID != nil && *encyclopediaID > 0 {
		ency, err = s.ency.GetByID(*encyclopediaID)
	} else if recipe.EncyclopediaRecipeID != nil {
		ency, err = s.ency.GetByID(*recipe.EncyclopediaRecipeID)
	} else {
		ency, err = s.ency.FindByName(recipe.Name)
	}
	if err != nil {
		return nil, nil, nil, err
	}

	result := diff.Compare(diff.FromEncyclopedia(*ency), diff.FromRecipeVersion(*myVer))
	eDTO := toEncyclopediaDTO(ency)
	mDTO := dto.ToVersionDTO(myVer)
	dDTO := toDiffDTO(result)
	return &eDTO, &mDTO, &dDTO, nil
}

func (s *RecipeService) Timeline(recipeID, userID uint64) (*dto.MyRecipeDTO, []dto.TimelineNodeDTO, error) {
	recipe, err := s.repo.GetByID(recipeID, userID)
	if err != nil {
		return nil, nil, err
	}
	var ver *model.RecipeVersion
	if recipe.CurrentVersionID != nil {
		ver, _ = s.repo.GetVersion(*recipe.CurrentVersionID)
	}
	versions, err := s.repo.ListVersions(recipeID, userID)
	if err != nil {
		return nil, nil, err
	}
	nodes := make([]dto.TimelineNodeDTO, 0, len(versions))
	for _, v := range versions {
		nodes = append(nodes, dto.TimelineNodeDTO{
			VersionID:     v.ID,
			VersionNumber: v.VersionNumber,
			CommitMsg:     v.CommitMsg,
			CreatedAt:     v.CreatedAt,
			IsCurrent:     recipe.CurrentVersionID != nil && v.ID == *recipe.CurrentVersionID,
		})
	}
	out := dto.ToRecipeDTO(recipe, ver)
	return &out, nodes, nil
}

// ── 回收站 ──

func (s *RecipeService) ListTrash(userID uint64, page, pageSize int) ([]dto.RecipeListItemDTO, dto.PageInfo, error) {
	items, total, err := s.repo.ListDeleted(userID, page, pageSize)
	if err != nil {
		return nil, dto.PageInfo{}, err
	}
	result := make([]dto.RecipeListItemDTO, 0, len(items))
	for _, item := range items {
		d := dto.RecipeListItemDTO{
			ID:            item.ID,
			Name:          item.Name,
			CoverImageURL: item.CoverImageURL,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		}
		if item.UserRating != nil {
			d.UserRating = *item.UserRating
		}
		result = append(result, d)
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	return result, dto.PageInfo{Page: page, PageSize: pageSize, Total: total, TotalPages: pages}, nil
}

func (s *RecipeService) Restore(id, userID uint64) error {
	return s.repo.Restore(id, userID)
}

func (s *RecipeService) PermanentDelete(id, userID uint64) error {
	return s.repo.PermanentDelete(id, userID)
}

// ── 导出/导入 ──

func (s *RecipeService) Export(userID uint64) ([]dto.ExportRecipeDTO, error) {
	recipes, err := s.repo.GetAllWithCurrentVersion(userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ExportRecipeDTO, 0, len(recipes))
	for _, r := range recipes {
		exp := dto.ExportRecipeDTO{
			Name:                 r.Name,
			UserRating:           r.UserRating,
			EncyclopediaRecipeID: r.EncyclopediaRecipeID,
		}
		if r.CurrentVersion != nil {
			exp.Ingredients = []dto.Ingredient(r.CurrentVersion.Ingredients)
			exp.ProcessSteps = []dto.ProcessStep(r.CurrentVersion.ProcessSteps)
			exp.ProcessText = r.CurrentVersion.ProcessText
			exp.Images = []string(r.CurrentVersion.Images)
			exp.CommitMsg = r.CurrentVersion.CommitMsg
		}
		result = append(result, exp)
	}
	return result, nil
}

func (s *RecipeService) Import(userID uint64, recipes []dto.ExportRecipeDTO) (*dto.ImportResultDTO, error) {
	res := &dto.ImportResultDTO{Total: len(recipes)}
	for _, exp := range recipes {
		existing, err := s.repo.ExistsByName(userID, exp.Name)
		if err != nil {
			res.Errors = append(res.Errors, exp.Name+": "+err.Error())
			continue
		}
		if existing != nil {
			// 已存在则添加新版本
			version := &model.RecipeVersion{
				Ingredients:  model.JSONSlice[model.Ingredient](exp.Ingredients),
				ProcessSteps: model.JSONSlice[model.ProcessStep](exp.ProcessSteps),
				ProcessText:  exp.ProcessText,
				Images:       model.StringSlice(exp.Images),
				CommitMsg:    impMsg(exp.CommitMsg),
			}
			if err := s.repo.AddVersion(existing, version); err != nil {
				res.Errors = append(res.Errors, exp.Name+": "+err.Error())
				continue
			}
			res.Updated++
		} else {
			recipe := &model.MyRecipe{
				UserID:               userID,
				Name:                 exp.Name,
				UserRating:           exp.UserRating,
				EncyclopediaRecipeID: exp.EncyclopediaRecipeID,
			}
			version := &model.RecipeVersion{
				Ingredients:  model.JSONSlice[model.Ingredient](exp.Ingredients),
				ProcessSteps: model.JSONSlice[model.ProcessStep](exp.ProcessSteps),
				ProcessText:  exp.ProcessText,
				Images:       model.StringSlice(exp.Images),
				CommitMsg:    impMsg(exp.CommitMsg),
			}
			if err := s.repo.CreateWithVersion(recipe, version); err != nil {
				res.Errors = append(res.Errors, exp.Name+": "+err.Error())
				continue
			}
			res.Created++
		}
	}
	return res, nil
}

func impMsg(msg string) string {
	if msg == "" {
		return "导入"
	}
	return msg
}

func toDiffDTO(r diff.VersionDiffResult) dto.VersionDiffResultDTO {
	ings := make([]dto.IngredientDiffDTO, 0, len(r.IngredientDiffs))
	for _, d := range r.IngredientDiffs {
		item := dto.IngredientDiffDTO{Type: d.Type.String(), AmountDelta: d.AmountDelta}
		if d.Base != nil {
			b := dto.Ingredient(*d.Base)
			item.Base = &b
		}
		if d.Target != nil {
			t := dto.Ingredient(*d.Target)
			item.Target = &t
		}
		ings = append(ings, item)
	}
	procs := make([]dto.ProcessStepDiffDTO, 0, len(r.ProcessDiffs))
	for _, d := range r.ProcessDiffs {
		item := dto.ProcessStepDiffDTO{Type: d.Type.String(), Order: d.Order}
		if d.Base != nil {
			b := dto.ProcessStep(*d.Base)
			item.Base = &b
		}
		if d.Target != nil {
			t := dto.ProcessStep(*d.Target)
			item.Target = &t
		}
		procs = append(procs, item)
	}
	return dto.VersionDiffResultDTO{
		IngredientDiffs: ings,
		ProcessDiffs:    procs,
		Summary:         r.Summary,
	}
}

func toEncyclopediaDTO(e *model.EncyclopediaRecipe) dto.EncyclopediaRecipeDTO {
	tags := []string(e.Tags)
	if tags == nil {
		tags = []string{}
	}
	return dto.EncyclopediaRecipeDTO{
		ID:            e.ID,
		Name:          e.Name,
		Description:   e.Description,
		CoverImageURL: e.CoverImageURL,
		Category:      e.Category,
		Tags:          tags,
		Ingredients:   []dto.Ingredient(e.Ingredients),
		ProcessSteps:  []dto.ProcessStep(e.ProcessSteps),
		Source:        e.Source,
		ViewCount:     e.ViewCount,
		CreatedAt:     e.CreatedAt,
	}
}
