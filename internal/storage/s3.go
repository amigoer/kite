package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Driver 基于 aws-sdk-go-v2 的 S3 兼容存储驱动。
// 同时支持阿里云 OSS、腾讯云 COS、Cloudflare R2、MinIO 等。
type S3Driver struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucket     string
	baseURL    string // 自定义 CDN 域名，为空时使用 S3 默认 URL
	endpoint   string
	pathStyle  bool
}

// NewS3Driver 根据 S3Config 创建一个新的 S3 存储驱动实例。
// endpoint 会自动补全 https:// 协议前缀（如果未指定）。
func NewS3Driver(cfg S3Config) (*S3Driver, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 driver: bucket is required")
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("s3 driver: endpoint is required")
	}
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("s3 driver: access_key_id and secret_access_key are required")
	}

	// 设置默认区域
	region := cfg.Region
	if region == "" {
		region = "auto"
	}

	// 确保 endpoint 包含协议前缀
	endpoint := cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	// 构建 S3 客户端配置
	client := s3.New(s3.Options{
		Region: region,
		BaseEndpoint: aws.String(endpoint),
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		// MinIO 等自建存储需要路径风格访问（如 http://host/bucket/key）
		// 阿里云 OSS、腾讯云 COS 等使用虚拟托管风格（如 http://bucket.host/key）
		UsePathStyle: cfg.ForcePathStyle,
	})

	// 处理 BaseURL：去除尾部斜杠，便于后续拼接
	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	return &S3Driver{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.Bucket,
		baseURL:   baseURL,
		endpoint:  endpoint,
		pathStyle: cfg.ForcePathStyle,
	}, nil
}

// Put 将文件上传到 S3 兼容存储。
// 使用 PutObject 进行上传，自动设置 Content-Type。
func (d *S3Driver) Put(ctx context.Context, key string, reader io.Reader, size int64, mimeType string) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(d.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(mimeType),
	}

	// 仅在已知文件大小时设置 ContentLength，
	// 避免某些 S3 兼容实现不支持 chunked transfer
	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}

	_, err := d.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("s3 put %q: %w", key, err)
	}
	return nil
}

// Get 从 S3 兼容存储下载文件。
// 返回文件内容的 ReadCloser 和文件大小，调用方必须负责关闭 ReadCloser。
func (d *S3Driver) Get(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	output, err := d.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("s3 get %q: %w", key, err)
	}

	var size int64
	if output.ContentLength != nil {
		size = *output.ContentLength
	}

	return output.Body, size, nil
}

// Delete 从 S3 兼容存储删除指定文件。
// 如果文件不存在，S3 API 默认不会返回错误（幂等操作）。
func (d *S3Driver) Delete(ctx context.Context, key string) error {
	_, err := d.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete %q: %w", key, err)
	}
	return nil
}

// Exists 检查 S3 兼容存储中指定 key 的文件是否存在。
// 使用 HeadObject 检查，不下载文件内容。
func (d *S3Driver) Exists(ctx context.Context, key string) (bool, error) {
	_, err := d.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// 检查是否为 NotFound 错误
		var notFound *s3types.NotFound
		if isNotFoundError(err, notFound) {
			return false, nil
		}
		return false, fmt.Errorf("s3 exists %q: %w", key, err)
	}
	return true, nil
}

// URL 根据存储 key 生成公开访问的 URL。
// 优先使用配置的 CDN 域名（BaseURL），否则使用 S3 默认域名拼接。
func (d *S3Driver) URL(key string) string {
	if d.baseURL != "" {
		return d.baseURL + "/" + key
	}

	// 无自定义域名时，根据访问风格拼接默认 URL
	if d.pathStyle {
		// 路径风格：{endpoint}/{bucket}/{key}
		return d.endpoint + "/" + d.bucket + "/" + key
	}
	// 虚拟托管风格：将 bucket 插入 endpoint 的域名前
	// 例如 https://oss-cn-hangzhou.aliyuncs.com -> https://bucket.oss-cn-hangzhou.aliyuncs.com/key
	return insertBucketIntoEndpoint(d.endpoint, d.bucket) + "/" + key
}

// SignedURL 生成带有效期的预签名 GET URL。
// 用于私有存储桶的临时文件访问，客户端可直接通过此 URL 下载文件。
func (d *S3Driver) SignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	output, err := d.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("s3 signed url %q: %w", key, err)
	}
	return output.URL, nil
}

// insertBucketIntoEndpoint 将 bucket 名称插入 endpoint URL 的域名前，
// 生成虚拟托管风格的 URL 前缀。
// 例如：https://oss-cn-hangzhou.aliyuncs.com + mybucket
//
//	-> https://mybucket.oss-cn-hangzhou.aliyuncs.com
func insertBucketIntoEndpoint(endpoint, bucket string) string {
	// 分离协议和主机部分
	scheme := "https://"
	host := endpoint
	if strings.HasPrefix(endpoint, "https://") {
		host = strings.TrimPrefix(endpoint, "https://")
	} else if strings.HasPrefix(endpoint, "http://") {
		scheme = "http://"
		host = strings.TrimPrefix(endpoint, "http://")
	}
	return scheme + bucket + "." + host
}

// isNotFoundError 判断 S3 错误是否为 NotFound 类型。
// 不同 S3 兼容实现返回的错误格式可能不同，
// 因此同时检查类型断言和错误消息中的关键字。
func isNotFoundError(err error, _ *s3types.NotFound) bool {
	// 优先检查 aws sdk 的标准错误类型
	var nf *s3types.NotFound
	if ok := containsError(err, &nf); ok {
		return true
	}
	// 兼容某些 S3 实现（如早期版本 MinIO）不返回标准错误类型的情况，
	// 通过错误消息中的关键字判断
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "nosuchkey") ||
		strings.Contains(msg, "404")
}

// containsError 使用 errors.As 的方式检查错误链中是否包含目标类型。
func containsError[T error](err error, target *T) bool {
	for err != nil {
		if v, ok := err.(T); ok {
			*target = v
			return true
		}
		// 尝试 unwrap
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
