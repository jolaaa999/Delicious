package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type UploadService struct {
	dir         string
	publicBase  string
	maxBytes    int64
}

func NewUploadService(dir, publicBase string, maxMB int64) *UploadService {
	if maxMB <= 0 {
		maxMB = 10
	}
	return &UploadService{
		dir:        dir,
		publicBase: strings.TrimRight(publicBase, "/"),
		maxBytes:   maxMB * 1024 * 1024,
	}
}

func (s *UploadService) EnsureDir() error {
	return os.MkdirAll(s.dir, 0o755)
}

type UploadResult struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

func (s *UploadService) Save(fileHeader *multipart.FileHeader) (*UploadResult, error) {
	if fileHeader.Size > s.maxBytes {
		return nil, fmt.Errorf("文件大小超过限制（最大 %d MB）", s.maxBytes/1024/1024)
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

	subDir := time.Now().Format("2006/01")
	targetDir := filepath.Join(s.dir, subDir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, err
	}

	name := uuid.New().String() + ext
	targetPath := filepath.Join(targetDir, name)
	dst, err := os.Create(targetPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(src, s.maxBytes+1))
	if err != nil {
		os.Remove(targetPath)
		return nil, err
	}
	if written > s.maxBytes {
		os.Remove(targetPath)
		return nil, fmt.Errorf("文件大小超过限制")
	}

	relPath := "/uploads/" + filepath.ToSlash(filepath.Join(subDir, name))
	url := relPath
	if s.publicBase != "" {
		url = s.publicBase + relPath
	}

	return &UploadResult{
		URL:      url,
		Filename: fileHeader.Filename,
		Size:     written,
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
