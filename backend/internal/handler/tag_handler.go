package handler

import (
	"net/http"
	"strconv"

	"github.com/delicious/delicious/internal/middleware"
	"github.com/delicious/delicious/internal/service"
	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

// List 获取所有标签
func (h *TagHandler) List(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Create 创建新标签
func (h *TagHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "标签名称不能为空")
		return
	}
	tag, err := h.svc.Create(req.Name)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"tag": tag})
}

// Delete 删除标签
func (h *TagHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的标签ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// ListByRecipe 获取菜谱关联的标签
func (h *TagHandler) ListByRecipe(c *gin.Context) {
	recipeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的菜谱ID")
		return
	}
	items, err := h.svc.ListByRecipe(recipeID)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// AddToRecipe 给菜谱添加标签
func (h *TagHandler) AddToRecipe(c *gin.Context) {
	recipeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的菜谱ID")
		return
	}
	var req struct {
		TagID uint64 `json:"tag_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "请提供 tag_id")
		return
	}
	if err := h.svc.AddToRecipe(recipeID, req.TagID); err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

// RemoveFromRecipe 移除菜谱的标签
func (h *TagHandler) RemoveFromRecipe(c *gin.Context) {
	recipeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的菜谱ID")
		return
	}
	tagID, err := strconv.ParseUint(c.Param("tag_id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的标签ID")
		return
	}
	if err := h.svc.RemoveFromRecipe(recipeID, tagID); err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
