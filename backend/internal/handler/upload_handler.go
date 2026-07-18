package handler

import (
	"net/http"

	"github.com/delicious/delicious/internal/middleware"
	"github.com/delicious/delicious/internal/repository"
	"github.com/delicious/delicious/internal/service"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	svc        *service.UploadService
	recipeRepo *repository.RecipeRepository
}

func NewUploadHandler(svc *service.UploadService, recipeRepo *repository.RecipeRepository) *UploadHandler {
	return &UploadHandler{svc: svc, recipeRepo: recipeRepo}
}

func (h *UploadHandler) CleanupScan(c *gin.Context) {
	refs, err := h.recipeRepo.AllReferencedImagePaths(middleware.GetUserID(c))
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	result, err := h.svc.ScanOrphans(refs)
	if err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *UploadHandler) CleanupExecute(c *gin.Context) {
	refs, err := h.recipeRepo.AllReferencedImagePaths(middleware.GetUserID(c))
	if err != nil {
		middleware.InternalError(c, err)
		return
	}
	result, err := h.svc.DeleteOrphans(refs)
	if err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		middleware.BadRequest(c, "请选择要上传的图片")
		return
	}
	result, err := h.svc.Save(file)
	if err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		middleware.BadRequest(c, "无效的表单数据")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		middleware.BadRequest(c, "请选择要上传的图片")
		return
	}
	results := make([]interface{}, 0, len(files))
	for _, f := range files {
		result, err := h.svc.Save(f)
		if err != nil {
			middleware.BadRequest(c, err.Error())
			return
		}
		results = append(results, result)
	}
	c.JSON(http.StatusOK, gin.H{"files": results})
}
