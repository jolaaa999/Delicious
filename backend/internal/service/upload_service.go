package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/delicious/delicious/internal/config"
	"github.com/google/uuid"
)

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type UploadService struct {
	cfg config.Config
}

func NewUploadService(cfg config.Config) *UploadService {
	return &UploadService{cfg: cfg}
}

func (s *UploadService) EnsureDir() error {
	if s.cfg.UseBlob || s.cfg.UploadDir == "" {
		return nil
	}
	return os.MkdirAll(s.cfg.UploadDir, 0o755)
}

type UploadResult struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

func (s *UploadService) Save(fileHeader *multipart.FileHeader) (*UploadResult, error) {
	if fileHeader.Size > s.maxBytes() {
		return nil, fmt.Errorf("文件大小超过限制（最大 %d MB）", s.maxBytes()/1024/1024)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	mime := detectMIME(buf[:n], fileHeader.Filename)
	ext, ok := allowedMIME[mime]
	if !ok {
		return nil, fmt.Errorf("不支持的图片格式，请上传 JPG/PNG/WebP/GIF")
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(io.LimitReader(src, s.maxBytes()+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > s.maxBytes() {
		return nil, fmt.Errorf("文件大小超过限制")
	}

	if s.cfg.UseBlob {
		return s.saveToBlob(data, mime, ext, fileHeader.Filename, int64(len(data)))
	}
	return s.saveToLocal(data, ext, fileHeader.Filename, int64(len(data)))
}

func (s *UploadService) maxBytes() int64 {
	maxMB := s.cfg.MaxUploadMB
	if maxMB <= 0 {
		maxMB = 10
	}
	return maxMB * 1024 * 1024
}

func (s *UploadService) saveToLocal(data []byte, ext, originalName string, size int64) (*UploadResult, error) {
	subDir := time.Now().Format("2006/01")
	targetDir := filepath.Join(s.cfg.UploadDir, subDir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, err
	}

	name := uuid.New().String() + ext
	targetPath := filepath.Join(targetDir, name)
	if err := os.WriteFile(targetPath, data, 0o644); err != nil {
		return nil, err
	}

	relPath := "/uploads/" + filepath.ToSlash(filepath.Join(subDir, name))
	url := relPath
	if s.cfg.PublicBaseURL != "" {
		url = strings.TrimRight(s.cfg.PublicBaseURL, "/") + relPath
	}

	return &UploadResult{
		URL:      url,
		Filename: originalName,
		Size:     size,
	}, nil
}

func (s *UploadService) saveToBlob(data []byte, mime, ext, originalName string, size int64) (*UploadResult, error) {
	subDir := time.Now().Format("2006/01")
	pathname := fmt.Sprintf("uploads/%s/%s%s", subDir, uuid.New().String(), ext)

	req, err := http.NewRequest(http.MethodPut, "https://blob.vercel-storage.com/"+pathname, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.BlobToken)
	req.Header.Set("x-api-version", "7")
	req.Header.Set("x-content-type", mime)
	req.Header.Set("x-add-random-suffix", "0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上传至 Blob 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("上传至 Blob 失败 (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 Blob 响应失败: %w", err)
	}
	if result.URL == "" {
		return nil, fmt.Errorf("Blob 未返回 URL")
	}

	return &UploadResult{
		URL:      result.URL,
		Filename: originalName,
		Size:     size,
	}, nil
}

func detectMIME(header []byte, filename string) string {
	if len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "image/jpeg"
	}
	if len(header) >= 8 && string(header[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(header) >= 12 && string(header[:4]) == "RIFF" && string(header[8:12]) == "WEBP" {
		return "image/webp"
	}
	if len(header) >= 6 && (string(header[:6]) == "GIF87a" || string(header[:6]) == "GIF89a") {
		return "image/gif"
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	}
	return "application/octet-stream"
}
