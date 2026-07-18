package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

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

// Upload 上传单张图片
// @Summary      上传单张图片
// @Description  上传一张图片（JPG/PNG/WebP/GIF），最大 10MB。Vercel 环境自动使用 Blob 存储
// @Tags         图片上传
// @Accept       mpfd
// @Produce      json
// @Param        file  formData  file  true  "图片文件"
// @Success      200   {object}  map[string]interface{}  "url, filename, size"
// @Failure      400   {object}  map[string]string  "文件格式不支持或超过大小限制"
// @Router       /upload [post]
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

// ProxyMedia 代理 private Blob，使前端 <img> 可直接展示
func (h *UploadHandler) ProxyMedia(c *gin.Context) {
	raw := c.Query("url")
	if raw == "" {
		middleware.BadRequest(c, "缺少 url 参数")
		return
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		middleware.BadRequest(c, "无效的图片地址")
		return
	}
	host := strings.ToLower(u.Hostname())
	if !strings.HasSuffix(host, ".blob.vercel-storage.com") {
		middleware.BadRequest(c, "仅允许代理 Vercel Blob 地址")
		return
	}

	body, contentType, err := h.svc.FetchBlob(u.String())
	if err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	defer body.Close()

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "private, max-age=86400")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, body)
}

// UploadMultiple 批量上传图片
// @Summary      批量上传图片
// @Description  一次上传多张图片，返回全部上传结果
// @Tags         图片上传
// @Accept       mpfd
// @Produce      json
// @Param        files  formData  []file  true  "图片文件数组"
// @Success      200    {object}  map[string]interface{}  "files"
// @Failure      400    {object}  map[string]string  "文件格式不支持"
// @Router       /upload/batch [post]
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

// CleanupScan 扫描孤立图片
// @Summary      扫描孤立图片
// @Description  对比数据库引用和磁盘文件，列出未被引用的孤立图片。仅本地存储模式可用
// @Tags         系统管理
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "result: total_files, orphan_files, freed_bytes"
// @Failure      400  {object}  map[string]string  "Blob 存储不支持此功能"
// @Router       /admin/cleanup-images [post]
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

// CleanupExecute 执行图片清理
// @Summary      执行图片清理
// @Description  删除所有未被引用的孤立图片。仅本地存储模式可用
// @Tags         系统管理
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "result: total_files, orphan_files(已删除), freed_bytes"
// @Failure      400  {object}  map[string]string  "Blob 存储不支持此功能"
// @Router       /admin/cleanup-images/execute [post]
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
