package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sort"

	"github.com/google/uuid"
)

// ImageInfo 图片信息
type ImageInfo struct {
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ImageListResult 图片列表结果
type ImageListResult struct {
	Items      []ImageInfo `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

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

// ListImages 列出上传目录中的所有图片
func (s *UploadService) ListImages(page, pageSize int) (*ImageListResult, error) {
	var images []ImageInfo

	err := filepath.Walk(s.uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if !allowedImageExts[ext] {
			return nil
		}

		relPath, _ := filepath.Rel(s.uploadDir, path)
		url := "/uploads/" + strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		images = append(images, ImageInfo{
			URL:       url,
			Filename:  strings.ReplaceAll(relPath, string(os.PathSeparator), "/"),
			Size:      info.Size(),
			UpdatedAt: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return &ImageListResult{Items: []ImageInfo{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}

	// 按时间倒序
	sort.Slice(images, func(i, j int) bool {
		return images[i].UpdatedAt.After(images[j].UpdatedAt)
	})

	total := len(images)

	// 分页
	start := (page - 1) * pageSize
	if start >= total {
		return &ImageListResult{Items: []ImageInfo{}, Total: total, Page: page, PageSize: pageSize}, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return &ImageListResult{
		Items:    images[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// DeleteImage 删除指定图片
func (s *UploadService) DeleteImage(relPath string) error {
	// 防止路径穿越
	clean := filepath.Clean(relPath)
	if strings.Contains(clean, "..") {
		return fmt.Errorf("invalid file path")
	}

	fullPath := filepath.Join(s.uploadDir, clean)

	// 确保在 uploads 目录内
	abs, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("invalid file path")
	}
	baseAbs, _ := filepath.Abs(s.uploadDir)
	if !strings.HasPrefix(abs, baseAbs) {
		return fmt.Errorf("invalid file path")
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	return os.Remove(fullPath)
}
