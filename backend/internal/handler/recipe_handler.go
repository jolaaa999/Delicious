package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/delicious/delicious/internal/dto"
	"github.com/delicious/delicious/internal/middleware"
	"github.com/delicious/delicious/internal/repository"
	"github.com/delicious/delicious/internal/service"
	"github.com/gin-gonic/gin"
)

type RecipeHandler struct {
	svc *service.RecipeService
}

func NewRecipeHandler(svc *service.RecipeService) *RecipeHandler {
	return &RecipeHandler{svc: svc}
}

// Create 创建菜谱
// @Summary      创建菜谱
// @Description  新建菜谱并自动创建第一个版本（version 1）
// @Tags         菜谱管理
// @Accept       json
// @Produce      json
// @Param        body  body  object  true  "菜谱信息"
// @Success      201  {object}  map[string]interface{}  "recipe"
// @Failure      400  {object}  map[string]string  "参数错误"
// @Failure      500  {object}  map[string]string  "服务端错误"
// @Router       /recipes [post]
func (h *RecipeHandler) Create(c *gin.Context) {
	var req dto.CreateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	recipe, err := h.svc.Create(uid, req)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"recipe": recipe})
}

// Get 获取菜谱详情
// @Summary      获取菜谱详情
// @Description  根据 ID 获取菜谱及当前版本完整信息
// @Tags         菜谱管理
// @Produce      json
// @Param        id   path      int  true  "菜谱 ID"
// @Success      200  {object}  map[string]interface{}  "recipe"
// @Failure      404  {object}  map[string]string  "菜谱不存在"
// @Router       /recipes/{id} [get]
func (h *RecipeHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	recipe, err := h.svc.Get(id, middleware.GetUserID(c))
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"recipe": recipe})
}

// List 菜谱列表
// @Summary      菜谱列表
// @Description  分页获取当前用户的菜谱列表，支持关键词搜索、评分筛选、排序
// @Tags         菜谱管理
// @Produce      json
// @Param        page          query     int     false  "页码"           default(1)
// @Param        page_size     query     int     false  "每页数量"        default(20)
// @Param        keyword       query     string  false  "菜名搜索关键词"
// @Param        min_rating    query     int     false  "最低评分 (1-5)"
// @Param        max_rating    query     int     false  "最高评分 (1-5)"
// @Param        order_by      query     string  false  "排序字段"        Enums(created_at, updated_at, user_rating)  default(updated_at)
// @Param        desc          query     bool    false  "是否降序"        default(true)
// @Param        created_after query     string  false  "创建时间起点 (RFC3339)"
// @Param        created_before query    string  false  "创建时间终点 (RFC3339)"
// @Success      200  {object}  map[string]interface{}  "items[], page_info"
// @Router       /recipes [get]
func (h *RecipeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	filter := repository.ListRecipesFilter{
		UserID:   middleware.GetUserID(c),
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
		OrderBy:  c.DefaultQuery("order_by", "updated_at"),
		Desc:     c.DefaultQuery("desc", "true") != "false",
	}
	if v := c.Query("min_rating"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 8); err == nil {
			u := uint8(n)
			filter.MinRating = &u
		}
	}
	if v := c.Query("max_rating"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 8); err == nil {
			u := uint8(n)
			filter.MaxRating = &u
		}
	}
	if v := c.Query("created_after"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.CreatedAfter = &t
		}
	}
	if v := c.Query("created_before"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.CreatedBefore = &t
		}
	}
	items, pageInfo, err := h.svc.List(filter)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page_info": pageInfo})
}

// Update 编辑菜谱
// @Summary      编辑菜谱
// @Description  编辑菜谱会新增一个不可变版本（版本号+1），不会修改旧版本
// @Tags         菜谱管理
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "菜谱 ID"
// @Param        body  body  object   true  "更新内容"
// @Success      200   {object}  map[string]interface{}  "recipe"
// @Failure      400   {object}  map[string]string  "参数错误"
// @Failure      404   {object}  map[string]string  "菜谱不存在"
// @Router       /recipes/{id} [put]
func (h *RecipeHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	var req dto.UpdateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	recipe, err := h.svc.Update(id, middleware.GetUserID(c), req)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"recipe": recipe})
}

// Delete 删除菜谱
// @Summary      删除菜谱（软删除）
// @Description  软删除菜谱，移入回收站。30 天内可通过 /restore 恢复
// @Tags         菜谱管理
// @Produce      json
// @Param        id   path      int  true  "菜谱 ID"
// @Success      200  {object}  map[string]interface{}  "{}"
// @Failure      404  {object}  map[string]string       "菜谱不存在"
// @Router       /recipes/{id} [delete]
func (h *RecipeHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Delete(id, middleware.GetUserID(c)); err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// ListVersions 版本历史列表
// @Summary      版本历史列表
// @Description  获取某个菜谱的所有历史版本，按版本号降序
// @Tags         版本管理
// @Produce      json
// @Param        id   path      int  true  "菜谱 ID"
// @Success      200  {object}  map[string]interface{}  "versions"
// @Failure      404  {object}  map[string]string  "菜谱不存在"
// @Router       /recipes/{id}/versions [get]
func (h *RecipeHandler) ListVersions(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	versions, err := h.svc.ListVersions(id, middleware.GetUserID(c))
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"versions": versions})
}

// GetVersion 获取某个版本详情
// @Summary      获取某个版本详情
// @Description  查看某个历史版本的完整配料和步骤
// @Tags         版本管理
// @Produce      json
// @Param        id          path  int  true  "菜谱 ID"
// @Param        version_id  path  int  true  "版本 ID"
// @Success      200  {object}  map[string]interface{}  "version"
// @Failure      404  {object}  map[string]string  "版本不存在"
// @Router       /recipes/{id}/versions/{version_id} [get]
func (h *RecipeHandler) GetVersion(c *gin.Context) {
	recipeID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	versionID, err := parseUintParam(c, "version_id")
	if err != nil {
		return
	}
	ver, err := h.svc.GetVersion(recipeID, versionID, middleware.GetUserID(c))
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"version": ver})
}

// CompareVersions 对比两个版本
// @Summary      对比两个版本
// @Description  O(n+m) 算法对比两个版本的配料和步骤差异，返回中文摘要
// @Tags         版本管理
// @Produce      json
// @Param        id               path      int    true   "菜谱 ID"
// @Param        base_version_id  query     int    true   "基准版本 ID"
// @Param        target_version_id query    int    true   "对比版本 ID"
// @Success      200  {object}  map[string]interface{}  "base_version, target_version, diff"
// @Router       /recipes/{id}/diff [get]
func (h *RecipeHandler) CompareVersions(c *gin.Context) {
	recipeID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	baseID, err := strconv.ParseUint(c.Query("base_version_id"), 10, 64)
	if err != nil || baseID == 0 {
		middleware.BadRequest(c, "base_version_id required")
		return
	}
	targetID, err := strconv.ParseUint(c.Query("target_version_id"), 10, 64)
	if err != nil || targetID == 0 {
		middleware.BadRequest(c, "target_version_id required")
		return
	}
	base, target, diffResult, err := h.svc.CompareVersions(recipeID, middleware.GetUserID(c), baseID, targetID)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"base_version":   base,
		"target_version": target,
		"diff":           diffResult,
	})
}

// CompareEncyclopedia 与百科基准对比
// @Summary      与百科基准对比
// @Description  将菜谱当前版本与关联的百科菜谱进行对比
// @Tags         版本管理
// @Produce      json
// @Param        id                       path      int    true   "菜谱 ID"
// @Param        encyclopedia_recipe_id   query     int    false  "指定百科菜谱 ID（可选，默认用菜谱已关联的）"
// @Success      200  {object}  map[string]interface{}  "encyclopedia_recipe_id, encyclopedia_name, encyclopedia_ingredients, encyclopedia_process_steps, my_version, diff"
// @Router       /recipes/{id}/diff/encyclopedia [get]
func (h *RecipeHandler) CompareEncyclopedia(c *gin.Context) {
	recipeID, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	var encyID *uint64
	if v := c.Query("encyclopedia_recipe_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			encyID = &n
		}
	}
	ency, myVer, diffResult, err := h.svc.CompareWithEncyclopedia(recipeID, middleware.GetUserID(c), encyID)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"encyclopedia_recipe_id":     ency.ID,
		"encyclopedia_name":          ency.Name,
		"encyclopedia_ingredients":   ency.Ingredients,
		"encyclopedia_process_steps": ency.ProcessSteps,
		"my_version":                 myVer,
		"diff":                       diffResult,
	})
}

// ── 回收站 ──

// ListTrash 回收站列表
// @Summary      回收站列表
// @Description  查看已删除的菜谱
// @Tags         回收站
// @Produce      json
// @Param        page       query  int  false  "页码"        default(1)
// @Param        page_size  query  int  false  "每页数量"     default(20)
// @Success      200  {object}  map[string]interface{}  "items[], page_info"
// @Router       /recipes/trash [get]
func (h *RecipeHandler) ListTrash(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, pageInfo, err := h.svc.ListTrash(middleware.GetUserID(c), page, pageSize)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page_info": pageInfo})
}

// Restore 恢复菜谱
// @Summary      恢复菜谱
// @Description  从回收站恢复已删除的菜谱
// @Tags         回收站
// @Produce      json
// @Param        id   path  int  true  "菜谱 ID"
// @Success      200  {object}  map[string]string  "message: 已恢复"
// @Failure      404  {object}  map[string]string  "菜谱不在回收站中"
// @Router       /recipes/{id}/restore [post]
func (h *RecipeHandler) Restore(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.Restore(id, middleware.GetUserID(c)); err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已恢复"})
}

// PermanentDelete 物理删除
// @Summary      物理删除
// @Description  彻底删除菜谱，不可恢复
// @Tags         回收站
// @Produce      json
// @Param        id   path  int  true  "菜谱 ID"
// @Success      200  {object}  map[string]string  "message: 已彻底删除"
// @Failure      404  {object}  map[string]string  "菜谱不在回收站中"
// @Router       /recipes/{id}/permanent [delete]
func (h *RecipeHandler) PermanentDelete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	if err := h.svc.PermanentDelete(id, middleware.GetUserID(c)); err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已彻底删除"})
}

// ── 导出/导入 ──

// Export 导出菜谱
// @Summary      导出菜谱
// @Description  导出当前用户所有菜谱（含当前版本完整数据）为 JSON
// @Tags         导入导出
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "recipes"
// @Router       /recipes/export [get]
func (h *RecipeHandler) Export(c *gin.Context) {
	recipes, err := h.svc.Export(middleware.GetUserID(c))
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"recipes": recipes})
}

// Import 导入菜谱
// @Summary      导入菜谱
// @Description  从 JSON 导入菜谱。按 name 去重：已存在则追加新版本，不存在则新建
// @Tags         导入导出
// @Accept       json
// @Produce      json
// @Param        body  body  object  true  "{\"recipes\": [...]}"
// @Success      200   {object}  map[string]interface{}  "result"
// @Failure      400   {object}  map[string]string  "参数错误"
// @Router       /recipes/import [post]
func (h *RecipeHandler) Import(c *gin.Context) {
	var req struct {
		Recipes []dto.ExportRecipeDTO `json:"recipes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "请提供 recipes 数组")
		return
	}
	result, err := h.svc.Import(middleware.GetUserID(c), req.Recipes)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func parseUintParam(c *gin.Context, name string) (uint64, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || v == 0 {
		middleware.BadRequest(c, "invalid "+name)
		return 0, err
	}
	return v, nil
}

func handleRepoErr(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		middleware.NotFound(c, "资源不存在")
		return
	}
	middleware.InternalError(c, err)
}
