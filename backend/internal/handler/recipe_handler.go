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
