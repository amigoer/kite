package service

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

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

// GenerateThumbnail 生成缩略图。
// 按 thumbWidth 等比缩放，输出为 JPEG 格式。
func (s *ImageService) GenerateThumbnail(reader io.Reader) (*bytes.Buffer, error) {
	src, err := imaging.Decode(reader, imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decode image for thumbnail: %w", err)
	}

	thumb := imaging.Resize(src, s.thumbWidth, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, thumb, imaging.JPEG, imaging.JPEGQuality(s.thumbQuality)); err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return buf, nil
}
