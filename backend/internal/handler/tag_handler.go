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

// List 标签列表
// @Summary      标签列表
// @Description  获取全部标签（字典表）
// @Tags         标签管理
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "items"
// @Router       /tags [get]
func (h *TagHandler) List(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Create 创建标签
// @Summary      创建标签
// @Description  新增一个标签名称
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        body  body  map[string]string  true  "{\"name\":\"快手菜\"}"
// @Success      201   {object}  map[string]interface{}  "tag"
// @Failure      400   {object}  map[string]string  "名称不能为空"
// @Router       /tags [post]
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
// @Summary      删除标签
// @Description  删除一个标签（同时清除关联关系）
// @Tags         标签管理
// @Produce      json
// @Param        id   path  int  true  "标签 ID"
// @Success      200  {object}  map[string]interface{}  "{}"
// @Failure      404  {object}  map[string]string  "标签不存在"
// @Router       /tags/{id} [delete]
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
// @Summary      获取菜谱关联的标签
// @Description  查看某个百科菜谱的所有标签
// @Tags         标签管理
// @Produce      json
// @Param        id   path  int  true  "百科菜谱 ID"
// @Success      200  {object}  map[string]interface{}  "items"
// @Router       /encyclopedia/{id}/tags [get]
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
// @Summary      给菜谱添加标签
// @Description  为百科菜谱关联一个标签
// @Tags         标签管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                true  "百科菜谱 ID"
// @Param        body  body  map[string]uint64   true  "{\"tag_id\":1}"
// @Success      201   {object}  map[string]interface{}  "{}"
// @Failure      400   {object}  map[string]string  "tag_id 无效"
// @Router       /encyclopedia/{id}/tags [post]
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
// @Summary      移除菜谱的标签
// @Description  取消百科菜谱与标签的关联
// @Tags         标签管理
// @Produce      json
// @Param        id      path  int  true  "百科菜谱 ID"
// @Param        tag_id  path  int  true  "标签 ID"
// @Success      200  {object}  map[string]interface{}  "{}"
// @Failure      404  {object}  map[string]string  "关联不存在"
// @Router       /encyclopedia/{id}/tags/{tag_id} [delete]
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
