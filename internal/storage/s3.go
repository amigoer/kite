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

// S3Driver is the S3-compatible storage driver, built on aws-sdk-go-v2.
// It handles Aliyun OSS, Tencent COS, Cloudflare R2, MinIO, and any other S3-compatible backend.
type S3Driver struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucket     string
	baseURL    string // custom CDN domain; falls back to the S3 default URL when empty
	endpoint   string
	pathStyle  bool
}

// NewS3Driver builds an S3Driver from the given S3Config.
// The endpoint is prefixed with https:// automatically when a scheme is missing.
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

	// Default the region when unset.
	region := cfg.Region
	if region == "" {
		region = "auto"
	}

	// Ensure the endpoint carries a scheme.
	endpoint := cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	// Build the S3 client.
	client := s3.New(s3.Options{
		Region: region,
		BaseEndpoint: aws.String(endpoint),
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		// MinIO and self-hosted stores expect path-style access (http://host/bucket/key).
		// Aliyun OSS, Tencent COS, etc. use virtual-hosted style (http://bucket.host/key).
		UsePathStyle: cfg.ForcePathStyle,
	})

	// Trim trailing slashes on BaseURL for cleaner concatenation.
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

// Put uploads a file to the S3-compatible backend via PutObject.
// Content-Type is set automatically.
func (d *S3Driver) Put(ctx context.Context, key string, reader io.Reader, size int64, mimeType string) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(d.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(mimeType),
	}

	// Only set ContentLength when the size is known; some S3-compatible backends do not
	// support chunked transfer encoding.
	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}

	_, err := d.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("s3 put %q: %w", key, err)
	}
	return nil
}

// Get downloads a file from the S3-compatible backend.
// Returns a ReadCloser and the object size; the caller must close the ReadCloser.
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

// Delete removes the object from the S3-compatible backend.
// S3's delete is idempotent, so a missing key does not produce an error.
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

// Exists reports whether the given key exists in the S3-compatible backend.
// Uses HeadObject, so the body is never downloaded.
func (d *S3Driver) Exists(ctx context.Context, key string) (bool, error) {
	_, err := d.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Unwrap to detect the NotFound error returned by the SDK.
		var notFound *s3types.NotFound
		if isNotFoundError(err, notFound) {
			return false, nil
		}
		return false, fmt.Errorf("s3 exists %q: %w", key, err)
	}
	return true, nil
}

// URL returns the public access URL for the object.
// When BaseURL is configured it is used directly; otherwise the URL is built from the endpoint.
func (d *S3Driver) URL(key string) string {
	if d.baseURL != "" {
		return d.baseURL + "/" + key
	}

	// No custom domain: build the URL according to access style.
	if d.pathStyle {
		// Path style: {endpoint}/{bucket}/{key}.
		return d.endpoint + "/" + d.bucket + "/" + key
	}
	// Virtual-hosted style: prepend the bucket to the endpoint's host.
	// e.g. https://oss-cn-hangzhou.aliyuncs.com -> https://bucket.oss-cn-hangzhou.aliyuncs.com/key
	return insertBucketIntoEndpoint(d.endpoint, d.bucket) + "/" + key
}

// SignedURL returns a time-limited pre-signed GET URL.
// Used to grant temporary access to objects inside a private bucket.
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

// insertBucketIntoEndpoint prepends the bucket name to the endpoint's host, producing the
// virtual-hosted URL prefix.
// e.g. https://oss-cn-hangzhou.aliyuncs.com + mybucket
//
//	-> https://mybucket.oss-cn-hangzhou.aliyuncs.com
func insertBucketIntoEndpoint(endpoint, bucket string) string {
	// Split the scheme from the host.
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

// isNotFoundError reports whether err represents an S3 NotFound response.
// Different S3-compatible backends surface "not found" inconsistently, so both the typed
// error and keywords in the error message are checked.
func isNotFoundError(err error, _ *s3types.NotFound) bool {
	// Prefer the typed error from the AWS SDK.
	var nf *s3types.NotFound
	if ok := containsError(err, &nf); ok {
		return true
	}
	// Fall back to message matching for backends (e.g. older MinIO) that do not return the
	// typed error.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "nosuchkey") ||
		strings.Contains(msg, "404")
}

// containsError walks the err chain (errors.As-style) looking for a target type.
func containsError[T error](err error, target *T) bool {
	for err != nil {
		if v, ok := err.(T); ok {
			*target = v
			return true
		}
		// Attempt to unwrap.
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
