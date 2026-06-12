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

func (h *DashboardHandler) Stats(c *gin.Context) {
	stats, err := h.dashSvc.Stats(middleware.GetUserID(c))
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

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
