package handler

import (
	"net/http"
	"strconv"

	"github.com/delicious/delicious/internal/middleware"
	"github.com/delicious/delicious/internal/service"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// List 分类列表
// @Summary      分类列表
// @Description  获取全部分类（字典表）
// @Tags         分类管理
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "items"
// @Router       /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Create 创建分类
// @Summary      创建分类
// @Description  新增一个分类名称
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Param        body  body  map[string]string  true  "{\"name\":\"川菜\"}"
// @Success      201   {object}  map[string]interface{}  "category"
// @Failure      400   {object}  map[string]string  "名称不能为空"
// @Router       /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "分类名称不能为空")
		return
	}
	category, err := h.svc.Create(req.Name)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category": category})
}

// Update 修改分类
// @Summary      修改分类
// @Description  修改分类名称
// @Tags         分类管理
// @Accept       json
// @Produce      json
// @Param        id    path  int                 true  "分类 ID"
// @Param        body  body  map[string]string   true  "{\"name\":\"新名称\"}"
// @Success      200   {object}  map[string]interface{}  "category"
// @Failure      400   {object}  map[string]string  "无效 ID 或名称为空"
// @Router       /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的分类ID")
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "分类名称不能为空")
		return
	}
	category, err := h.svc.Update(id, req.Name)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": category})
}

// Delete 删除分类
// @Summary      删除分类
// @Description  删除一个分类
// @Tags         分类管理
// @Produce      json
// @Param        id   path  int  true  "分类 ID"
// @Success      200  {object}  map[string]interface{}  "{}"
// @Failure      404  {object}  map[string]string  "分类不存在"
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "无效的分类ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
