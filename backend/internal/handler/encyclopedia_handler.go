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

// Search 搜索百科菜谱
// @Summary      搜索百科菜谱
// @Description  搜索百科菜谱库。优先从外部 API（Spoonacular + MealDB）搜索，无结果或禁用时回退本地数据库。支持中英翻译
// @Tags         百科菜谱
// @Produce      json
// @Param        keyword    query  string  false  "搜索关键词（中英文均可）"
// @Param        category   query  string  false  "分类筛选"
// @Param        lang       query  string  false  "返回语言 zh/en"  default(en)
// @Param        page       query  int     false  "页码"            default(1)
// @Param        page_size  query  int     false  "每页数量"         default(20)
// @Success      200  {object}  map[string]interface{}  "items[], page_info"
// @Failure      500  {object}  map[string]string  "服务端错误"
// @Router       /encyclopedia/search [get]
func (h *EncyclopediaHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, pageInfo, err := h.svc.Search(c.Query("keyword"), c.Query("category"), c.Query("lang"), page, pageSize)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page_info": pageInfo})
}

// Get 百科菜谱详情
// @Summary      百科菜谱详情
// @Description  获取百科菜谱的完整信息，包括配料和制作步骤。浏览量自动 +1。支持翻译
// @Tags         百科菜谱
// @Produce      json
// @Param        id    path  int     true   "百科菜谱 ID"
// @Param        lang  query string  false  "返回语言 zh/en"
// @Success      200   {object}  map[string]interface{}  "recipe"
// @Failure      404   {object}  map[string]string  "菜谱不存在"
// @Router       /encyclopedia/{id} [get]
func (h *EncyclopediaHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return
	}
	recipe, err := h.svc.Get(id, c.Query("lang"))
	if err != nil {
		handleRepoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"recipe": recipe})
}

// ListByCategory 按分类浏览百科
// @Summary      按分类浏览百科
// @Description  获取某个分类下的所有百科菜谱
// @Tags         百科菜谱
// @Produce      json
// @Param        category   path   string  true   "分类名称，如 川菜、西餐"
// @Param        lang       query  string  false  "返回语言 zh/en"
// @Param        page       query  int     false  "页码"        default(1)
// @Param        page_size  query  int     false  "每页数量"     default(20)
// @Success      200  {object}  map[string]interface{}  "items[], page_info"
// @Router       /encyclopedia/category/{category} [get]
func (h *EncyclopediaHandler) ListByCategory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, pageInfo, err := h.svc.ListByCategory(c.Param("category"), c.Query("lang"), page, pageSize)
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "page_info": pageInfo})
}
