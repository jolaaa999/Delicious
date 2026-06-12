package handler

import (
	"net/http"
	"strconv"

	"github.com/delicious/delicious/internal/middleware"
	"github.com/delicious/delicious/internal/service"
	"github.com/gin-gonic/gin"
)

type EncyclopediaHandler struct {
	svc *service.EncyclopediaService
}

func NewEncyclopediaHandler(svc *service.EncyclopediaService) *EncyclopediaHandler {
	return &EncyclopediaHandler{svc: svc}
}

func (h *EncyclopediaHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, pageInfo, err := h.svc.Search(c.Query("keyword"), c.Query("category"), page, pageSize)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page_info": pageInfo})
}

func (h *EncyclopediaHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	recipe, err := h.svc.Get(id)
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"recipe": recipe})
}

func (h *EncyclopediaHandler) ListByCategory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, pageInfo, err := h.svc.ListByCategory(c.Param("category"), page, pageSize)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page_info": pageInfo})
}
