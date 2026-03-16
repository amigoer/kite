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

const (
	// 默认上传目录
	DefaultUploadDir = "uploads"
	// 最大文件大小 10MB
	MaxUploadSize = 10 << 20
)

// 允许的图片扩展名
var allowedImageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".webp": true, ".svg": true, ".ico": true, ".avif": true,
}

// UploadResult 上传结果
type UploadResult struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// UploadService 文件上传服务
type UploadService struct {
	uploadDir string
}

func NewUploadService(uploadDir string) *UploadService {
	if uploadDir == "" {
		uploadDir = DefaultUploadDir
	}
	return &UploadService{uploadDir: uploadDir}
}

// SaveImage 保存上传的图片文件
func (s *UploadService) SaveImage(file *multipart.FileHeader) (*UploadResult, error) {
	// 检查文件大小
	if file.Size > MaxUploadSize {
		return nil, fmt.Errorf("文件大小超过限制 (最大 %dMB)", MaxUploadSize>>20)
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 按日期分目录：uploads/2026/03/
	now := time.Now()
	subDir := filepath.Join(s.uploadDir, now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(subDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 生成唯一文件名
	newFilename := uuid.New().String() + ext
	dstPath := filepath.Join(subDir, newFilename)

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	// 写入
	written, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	// 构建访问 URL（相对路径，由前端补全域名）
	relPath := filepath.Join(now.Format("2006"), now.Format("01"), newFilename)
	url := "/uploads/" + strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

	return &UploadResult{
		URL:      url,
		Filename: newFilename,
		Size:     written,
	}, nil
}
