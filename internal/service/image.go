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

// ImageService handles image processing such as thumbnail generation and dimension inspection.
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

// ImageDimensions holds an image's width and height.
type ImageDimensions struct {
	Width  int
	Height int
}

// GetDimensions returns the image's width and height.
func (s *ImageService) GetDimensions(reader io.Reader) (*ImageDimensions, error) {
	cfg, _, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, fmt.Errorf("decode image config: %w", err)
	}
	return &ImageDimensions{Width: cfg.Width, Height: cfg.Height}, nil
}

// GenerateThumbnail produces a thumbnail scaled proportionally to thumbWidth.
// The output format depends on the source:
//   - PNG / WebP / GIF -> PNG, preserving the alpha channel.
//   - Everything else (JPEG / BMP / TIFF, ...) -> JPEG for smaller files.
//
// It returns (buf, actual output mime). The caller should use the returned mime for the storage
// Content-Type and serve the same Content-Type on access.
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

// ThumbnailMimeFor returns the thumbnail output MIME for a given source image MIME.
// Formats that support transparency (PNG / WebP / GIF) map to image/png to preserve alpha;
// everything else (including unknown inputs) maps to image/jpeg.
// Exported so both the upload pipeline and the /t/:hash handler can agree on the Content-Type.
func ThumbnailMimeFor(srcMime string) string {
	switch srcMime {
	case "image/png", "image/webp", "image/gif":
		return "image/png"
	default:
		return "image/jpeg"
	}
}
