package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/delicious/delicious/pkg/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type RecipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

type ListRecipesFilter struct {
	UserID        uint64
	Page          int
	PageSize      int
	MinRating     *uint8
	MaxRating     *uint8
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Keyword       string
	OrderBy       string
	Desc          bool
}

func (r *RecipeRepository) CreateWithVersion(recipe *model.MyRecipe, version *model.RecipeVersion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(recipe).Error; err != nil {
			return err
		}
		version.RecipeID = recipe.ID
		version.VersionNumber = 1
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		cover := pickCover(version.Images, recipe.CoverImageURL)
		return tx.Model(recipe).Updates(map[string]interface{}{
			"current_version_id": version.ID,
			"cover_image_url":    cover,
		}).Error
	})
}

func (r *RecipeRepository) AddVersion(recipe *model.MyRecipe, version *model.RecipeVersion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var maxVer uint32
		if err := tx.Model(&model.RecipeVersion{}).
			Where("recipe_id = ?", recipe.ID).
			Select("COALESCE(MAX(version_number), 0)").
			Scan(&maxVer).Error; err != nil {
			return err
		}
		version.RecipeID = recipe.ID
		version.VersionNumber = maxVer + 1
		if err := tx.Create(version).Error; err != nil {
			return err
		}
		cover := pickCover(version.Images, recipe.CoverImageURL)
		updates := map[string]interface{}{
			"name":                 recipe.Name,
			"current_version_id":   version.ID,
			"cover_image_url":      cover,
			"encyclopedia_recipe_id": recipe.EncyclopediaRecipeID,
		}
		if recipe.UserRating != nil {
			updates["user_rating"] = *recipe.UserRating
		}
		return tx.Model(recipe).Updates(updates).Error
	})
}

func pickCover(images model.StringSlice, existing *string) *string {
	if len(images) > 0 {
		s := images[0]
		return &s
	}
	return existing
}

func (r *RecipeRepository) GetByID(id, userID uint64) (*model.MyRecipe, error) {
	var recipe model.MyRecipe
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&recipe).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &recipe, err
}

func (r *RecipeRepository) GetVersion(id uint64) (*model.RecipeVersion, error) {
	var ver model.RecipeVersion
	err := r.db.First(&ver, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &ver, err
}

// GetVersionsByIDs 批量获取版本，仅查询 List 所需的 version_number 字段。
// 返回 map[version_id]*RecipeVersion，避免 N+1 查询。
func (r *RecipeRepository) GetVersionsByIDs(ids []uint64) (map[uint64]*model.RecipeVersion, error) {
	if len(ids) == 0 {
		return map[uint64]*model.RecipeVersion{}, nil
	}
	var versions []model.RecipeVersion
	if err := r.db.Where("id IN ?", ids).Find(&versions).Error; err != nil {
		return nil, err
	}
	m := make(map[uint64]*model.RecipeVersion, len(versions))
	for i := range versions {
		m[versions[i].ID] = &versions[i]
	}
	return m, nil
}

func (r *RecipeRepository) GetVersionByRecipe(recipeID, versionID, userID uint64) (*model.RecipeVersion, error) {
	if _, err := r.GetByID(recipeID, userID); err != nil {
		return nil, err
	}
	var ver model.RecipeVersion
	err := r.db.Where("id = ? AND recipe_id = ?", versionID, recipeID).First(&ver).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &ver, err
}

func (r *RecipeRepository) ListVersions(recipeID, userID uint64) ([]model.RecipeVersion, error) {
	if _, err := r.GetByID(recipeID, userID); err != nil {
		return nil, err
	}
	var versions []model.RecipeVersion
	err := r.db.Where("recipe_id = ?", recipeID).
		Order("version_number DESC").
		Find(&versions).Error
	return versions, err
}

func (r *RecipeRepository) List(filter ListRecipesFilter) ([]model.MyRecipe, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	q := r.db.Model(&model.MyRecipe{}).Where("user_id = ?", filter.UserID)
	if filter.MinRating != nil {
		q = q.Where("user_rating >= ?", *filter.MinRating)
	}
	if filter.MaxRating != nil {
		q = q.Where("user_rating <= ?", *filter.MaxRating)
	}
	if filter.CreatedAfter != nil {
		q = q.Where("created_at >= ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		q = q.Where("created_at <= ?", *filter.CreatedBefore)
	}
	if filter.Keyword != "" {
		q = q.Where("name LIKE ?", fmt.Sprintf("%%%s%%", filter.Keyword))
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderCol := "updated_at"
	switch filter.OrderBy {
	case "created_at", "updated_at", "user_rating":
		orderCol = filter.OrderBy
	}
	order := orderCol + " DESC"
	if !filter.Desc {
		order = orderCol + " ASC"
	}

	var items []model.MyRecipe
	err := q.Order(order).
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&items).Error
	return items, total, err
}

func (r *RecipeRepository) SoftDelete(id, userID uint64) error {
	res := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.MyRecipe{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *RecipeRepository) CountVersionsByUser(userID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.RecipeVersion{}).
		Joins("JOIN my_recipes ON my_recipes.id = recipe_versions.recipe_id").
		Where("my_recipes.user_id = ? AND my_recipes.deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (r *RecipeRepository) RatingDistribution(userID uint64) (map[uint8]int64, error) {
	type row struct {
		Rating uint8
		Count  int64
	}
	var rows []row
	err := r.db.Model(&model.MyRecipe{}).
		Select("user_rating as rating, COUNT(*) as count").
		Where("user_id = ? AND user_rating IS NOT NULL", userID).
		Group("user_rating").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint8]int64)
	for _, row := range rows {
		m[row.Rating] = row.Count
	}
	return m, nil
}

func (r *RecipeRepository) AverageRating(userID uint64) (float64, error) {
	var avg *float64
	err := r.db.Model(&model.MyRecipe{}).
		Where("user_id = ? AND user_rating IS NOT NULL", userID).
		Select("AVG(user_rating)").
		Scan(&avg).Error
	if err != nil || avg == nil {
		return 0, err
	}
	return *avg, nil
}

func (r *RecipeRepository) LatestRecipeAt(userID uint64) (*time.Time, error) {
	var t time.Time
	err := r.db.Model(&model.MyRecipe{}).
		Where("user_id = ?", userID).
		Select("MAX(created_at)").
		Scan(&t).Error
	if err != nil || t.IsZero() {
		return nil, err
	}
	return &t, nil
}

func (r *RecipeRepository) CountByUser(userID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.MyRecipe{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// AllReferencedImagePaths 收集所有被引用的图片相对路径（从 my_recipes 和 recipe_versions 中）
func (r *RecipeRepository) AllReferencedImagePaths(userID uint64) (map[string]bool, error) {
	refs := make(map[string]bool)

	// 从 my_recipes 收集封面图
	var covers []*string
	if err := r.db.Model(&model.MyRecipe{}).
		Where("user_id = ? AND cover_image_url IS NOT NULL AND cover_image_url != ''", userID).
		Pluck("cover_image_url", &covers).Error; err != nil {
		return nil, err
	}
	for _, c := range covers {
		if c != nil {
			refs[*c] = true
		}
	}

	// 从 recipe_versions 收集图片数组（JSON 字段，用 LIKE 匹配 /uploads/）
	var imageJsons []string
	if err := r.db.Model(&model.RecipeVersion{}).
		Joins("JOIN my_recipes ON my_recipes.id = recipe_versions.recipe_id").
		Where("my_recipes.user_id = ? AND recipe_versions.images IS NOT NULL", userID).
		Pluck("recipe_versions.images::text", &imageJsons).Error; err != nil {
		return nil, err
	}
	for _, raw := range imageJsons {
		// JSON 数组字符串，简单提取 /uploads/... URL
		for _, u := range extractUploadURLs(raw) {
			if u != "" {
				refs[u] = true
			}
		}
	}
	return refs, nil
}

// extractUploadURLs 从 JSON 数组字符串中提取 /uploads/ 开头的 URL
func extractUploadURLs(raw string) []string {
	var urls []string
	// 简单扫描引号内的 /uploads/ 路径
	inQuote := false
	start := -1
	for i := 0; i < len(raw); i++ {
		if raw[i] == '"' {
			if inQuote && start >= 0 {
				val := raw[start:i]
				if len(val) > 0 && (val[0] == '/' || strings.HasPrefix(val, "http")) {
					urls = append(urls, val)
				}
				start = -1
			} else {
				start = i + 1
			}
			inQuote = !inQuote
		}
	}
	return urls
}

// ── 回收站 ──

// ListDeleted 列出被软删除的菜谱（回收站）
func (r *RecipeRepository) ListDeleted(userID uint64, page, pageSize int) ([]model.MyRecipe, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := r.db.Unscoped().Model(&model.MyRecipe{}).
		Where("user_id = ? AND deleted_at IS NOT NULL", userID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.MyRecipe
	err := q.Order("deleted_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	return items, total, err
}

// Restore 恢复软删除的菜谱
func (r *RecipeRepository) Restore(id, userID uint64) error {
	res := r.db.Unscoped().Model(&model.MyRecipe{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Update("deleted_at", nil)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// PermanentDelete 物理删除（不可恢复）
func (r *RecipeRepository) PermanentDelete(id, userID uint64) error {
	res := r.db.Unscoped().Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Delete(&model.MyRecipe{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ── 导出/导入 ──

// GetAllWithCurrentVersion 获取用户所有菜谱及其当前版本（不分页，供导出用）
func (r *RecipeRepository) GetAllWithCurrentVersion(userID uint64) ([]model.MyRecipe, error) {
	var recipes []model.MyRecipe
	err := r.db.Where("user_id = ?", userID).
		Preload("CurrentVersion").
		Order("created_at DESC").
		Find(&recipes).Error
	return recipes, err
}

// ExistsByName 按名称检查菜谱是否已存在（用于导入去重）
func (r *RecipeRepository) ExistsByName(userID uint64, name string) (*model.MyRecipe, error) {
	var recipe model.MyRecipe
	err := r.db.Where("user_id = ? AND name = ?", userID, name).First(&recipe).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &recipe, err
}
