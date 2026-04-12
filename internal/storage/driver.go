package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// StorageDriver 是所有存储后端必须实现的接口。
// 本地文件系统和 S3 兼容存储均通过此接口抽象，
// 上层业务代码不感知底层存储细节。
type StorageDriver interface {
	// Put 将文件内容写入存储后端。
	// key 为存储路径（如 "2024/01/abc12345/uuid.jpg"），由上层生成。
	// size 为文件字节数，mimeType 为文件的 MIME 类型。
	Put(ctx context.Context, key string, reader io.Reader, size int64, mimeType string) error

	// Get 从存储后端读取文件内容。
	// 返回的 io.ReadCloser 必须由调用方负责关闭。
	// 第二个返回值为文件大小（字节数）。
	Get(ctx context.Context, key string) (io.ReadCloser, int64, error)

	// Delete 从存储后端删除指定文件。
	// 如果文件不存在，应返回 nil 而非错误（幂等删除）。
	Delete(ctx context.Context, key string) error

	// Exists 检查指定 key 的文件是否存在于存储后端。
	Exists(ctx context.Context, key string) (bool, error)

	// URL 根据存储 key 生成公开访问的 URL。
	// 对于配置了 CDN 域名的存储，返回 CDN URL。
	URL(key string) string

	// SignedURL 生成带有效期的预签名访问 URL。
	// 主要用于私有存储桶的临时文件访问。
	SignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
}

// StorageConfig 存储配置，从数据库 storage_configs 表的 config 字段反序列化。
type StorageConfig struct {
	Driver string       `json:"driver"`          // 存储驱动类型：local / s3
	Local  *LocalConfig `json:"local,omitempty"` // 本地存储配置，driver 为 local 时必填
	S3     *S3Config    `json:"s3,omitempty"`    // S3 存储配置，driver 为 s3 时必填
}

// LocalConfig 本地文件系统存储的配置。
type LocalConfig struct {
	BasePath string `json:"base_path"` // 文件存储根目录（绝对路径）
	BaseURL  string `json:"base_url"`  // 访问 URL 前缀（如 "https://kite.plus/files"）
}

// S3Config S3 兼容存储的配置。
// 通过不同的 Endpoint 和选项支持阿里云 OSS、腾讯云 COS、Cloudflare R2、MinIO 等。
type S3Config struct {
	Endpoint        string `json:"endpoint"`          // 服务端点（如 "oss-cn-hangzhou.aliyuncs.com"）
	Region          string `json:"region"`            // 区域（如 "cn-hangzhou"、"auto"）
	Bucket          string `json:"bucket"`            // 存储桶名称
	AccessKeyID     string `json:"access_key_id"`     // 访问密钥 ID
	SecretAccessKey string `json:"secret_access_key"` // 访问密钥 Secret
	BaseURL         string `json:"base_url"`          // CDN 域名，可选（如 "https://cdn.example.com"）
	ForcePathStyle  bool   `json:"force_path_style"`  // 是否使用路径风格访问，MinIO 需设为 true
}

// NewDriver 根据配置创建对应的存储驱动实例。
func NewDriver(cfg StorageConfig) (StorageDriver, error) {
	switch cfg.Driver {
	case "s3":
		if cfg.S3 == nil {
			return nil, fmt.Errorf("storage config: driver is s3 but s3 config is nil")
		}
		return NewS3Driver(*cfg.S3)
	case "local":
		if cfg.Local == nil {
			return nil, fmt.Errorf("storage config: driver is local but local config is nil")
		}
		return NewLocalDriver(*cfg.Local)
	default:
		return nil, fmt.Errorf("storage config: unknown driver %q", cfg.Driver)
	}
}
