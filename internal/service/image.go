package service

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
)

// ImageService 图片处理服务，负责缩略图生成和图片尺寸获取。
type ImageService struct {
	thumbWidth   int
	thumbQuality int
}

func NewImageService(thumbWidth, thumbQuality int) *ImageService {
	return &ImageService{
		thumbWidth:   thumbWidth,
		thumbQuality: thumbQuality,
	}
}

// ImageDimensions 图片尺寸信息。
type ImageDimensions struct {
	Width  int
	Height int
}

// GetDimensions 获取图片的宽高。
func (s *ImageService) GetDimensions(reader io.Reader) (*ImageDimensions, error) {
	cfg, _, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, fmt.Errorf("decode image config: %w", err)
	}
	return &ImageDimensions{Width: cfg.Width, Height: cfg.Height}, nil
}

// GenerateThumbnail 生成缩略图。按 thumbWidth 等比缩放。
// 根据源图片类型选择输出格式：
//   - PNG / WebP / GIF → 输出 PNG，保留 alpha 透明通道；
//   - 其它（JPEG / BMP / TIFF 等）→ 输出 JPEG，体积更小。
//
// 返回 (buf, 实际输出的 mime)。调用方应使用返回的 mime 写入存储 Content-Type
// 并在访问时返回相同的 Content-Type。
func (s *ImageService) GenerateThumbnail(reader io.Reader, srcMime string) (*bytes.Buffer, string, error) {
	src, err := imaging.Decode(reader, imaging.AutoOrientation(true))
	if err != nil {
		return nil, "", fmt.Errorf("decode image for thumbnail: %w", err)
	}

	thumb := imaging.Resize(src, s.thumbWidth, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)
	outMime := ThumbnailMimeFor(srcMime)
	switch outMime {
	case "image/png":
		if err := imaging.Encode(buf, thumb, imaging.PNG); err != nil {
			return nil, "", fmt.Errorf("encode png thumbnail: %w", err)
		}
	default:
		if err := imaging.Encode(buf, thumb, imaging.JPEG, imaging.JPEGQuality(s.thumbQuality)); err != nil {
			return nil, "", fmt.Errorf("encode jpeg thumbnail: %w", err)
		}
	}

	return buf, outMime, nil
}

// ThumbnailMimeFor 返回给定源图 MIME 类型对应的缩略图输出 MIME。
// 透明类格式（PNG / WebP / GIF）映射到 image/png 以保留 alpha；
// 其它格式（含未知）映射到 image/jpeg。
// 导出供上传写入 Content-Type 以及 /t/:hash 访问处理程序读取 Content-Type 复用。
func ThumbnailMimeFor(srcMime string) string {
	switch srcMime {
	case "image/png", "image/webp", "image/gif":
		return "image/png"
	default:
		return "image/jpeg"
	}
}
