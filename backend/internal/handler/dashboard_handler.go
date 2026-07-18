package handler

import (
	"net/http"

	"github.com/delicious/delicious/internal/middleware"
	"github.com/delicious/delicious/internal/service"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	recipeSvc *service.RecipeService
	dashSvc   *service.DashboardService
}

func NewDashboardHandler(recipeSvc *service.RecipeService, dashSvc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{recipeSvc: recipeSvc, dashSvc: dashSvc}
}

// Stats 数据总览
// @Summary      数据总览
// @Description  返回当前用户的统计概览：总菜谱数、平均评分、总版本数、评分分布
// @Tags         数据面板
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "total_recipes, average_rating, total_versions, rating_distribution[], latest_recipe_at"
// @Router       /dashboard/stats [get]
func (h *DashboardHandler) Stats(c *gin.Context) {
	stats, err := h.dashSvc.Stats(middleware.GetUserID(c))
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Timeline 菜谱时间线
// @Summary      菜谱时间线
// @Description  获取菜谱完整时间线：基本信息 + 所有版本节点 + 当前版本标记
// @Tags         数据面板
// @Produce      json
// @Param        id   path  int  true  "菜谱 ID"
// @Success      200  {object}  map[string]interface{}  "recipe, timeline[]"
// @Failure      404  {object}  map[string]string  "菜谱不存在"
// @Router       /dashboard/recipes/{id}/timeline [get]
func (h *DashboardHandler) Timeline(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	recipe, timeline, err := h.recipeSvc.Timeline(id, middleware.GetUserID(c))
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"recipe": recipe, "timeline": timeline})
}
