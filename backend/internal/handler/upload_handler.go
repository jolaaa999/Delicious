package handler

import (
	"net/http"

	"github.com/delicious/delicious/internal/middleware"
	"github.com/delicious/delicious/internal/service"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
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
