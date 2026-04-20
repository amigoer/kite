package storage

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Scheme describes one user-facing storage option. The UI renders schemes directly,
// while the backend resolves them into driver + provider + normalized config.
type Scheme struct {
	Key         string         `json:"key"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Driver      string         `json:"driver"`
	Provider    string         `json:"provider,omitempty"`
	Defaults    map[string]any `json:"defaults"`
}

var schemeCatalog = []Scheme{
	{
		Key:         DriverLocal,
		Name:        "本地存储",
		Description: "服务器本地磁盘，适合单机部署和开发环境。",
		Driver:      DriverLocal,
		Defaults:    map[string]any{"base_path": "./uploads", "base_url": ""},
	},
	{
		Key:         DriverFTP,
		Name:        "FTP",
		Description: "传统 FTP 服务器，适合接入已有文件服务。",
		Driver:      DriverFTP,
		Defaults:    map[string]any{"host": "", "port": 21, "username": "", "password": "", "base_path": "/", "base_url": ""},
	},
	{
		Key:         ProviderAWSS3,
		Name:        "Amazon S3",
		Description: "AWS 官方对象存储。",
		Driver:      DriverS3,
		Provider:    ProviderAWSS3,
		Defaults:    map[string]any{"endpoint": "s3.amazonaws.com", "region": "us-east-1", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderAliyunOSS,
		Name:        "阿里云 OSS",
		Description: "阿里云对象存储服务。",
		Driver:      DriverS3,
		Provider:    ProviderAliyunOSS,
		Defaults:    map[string]any{"endpoint": "oss-cn-hangzhou.aliyuncs.com", "region": "cn-hangzhou", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderTencentCOS,
		Name:        "腾讯云 COS",
		Description: "腾讯云对象存储服务。",
		Driver:      DriverS3,
		Provider:    ProviderTencentCOS,
		Defaults:    map[string]any{"endpoint": "cos.ap-guangzhou.myqcloud.com", "region": "ap-guangzhou", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderHuaweiOBS,
		Name:        "华为云 OBS",
		Description: "华为云对象存储服务。",
		Driver:      DriverS3,
		Provider:    ProviderHuaweiOBS,
		Defaults:    map[string]any{"endpoint": "obs.cn-north-4.myhuaweicloud.com", "region": "cn-north-4", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderBaiduBOS,
		Name:        "百度云 BOS",
		Description: "百度智能云对象存储服务。",
		Driver:      DriverS3,
		Provider:    ProviderBaiduBOS,
		Defaults:    map[string]any{"endpoint": "s3.bj.bcebos.com", "region": "bj", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderCloudflareR2,
		Name:        "Cloudflare R2",
		Description: "Cloudflare 的 S3 兼容对象存储。",
		Driver:      DriverS3,
		Provider:    ProviderCloudflareR2,
		Defaults:    map[string]any{"endpoint": "<accountid>.r2.cloudflarestorage.com", "region": "auto", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderMinIO,
		Name:        "MinIO",
		Description: "自托管 S3 兼容对象存储，默认启用路径式访问。",
		Driver:      DriverS3,
		Provider:    ProviderMinIO,
		Defaults:    map[string]any{"endpoint": "minio.example.com", "region": "us-east-1", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": true},
	},
	{
		Key:         ProviderGoogleGCS,
		Name:        "Google Cloud Storage",
		Description: "Google Cloud Storage 的 XML / HMAC 兼容模式。",
		Driver:      DriverS3,
		Provider:    ProviderGoogleGCS,
		Defaults:    map[string]any{"endpoint": "storage.googleapis.com", "region": "auto", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderBackblazeB2,
		Name:        "Backblaze B2",
		Description: "Backblaze B2 的 S3 Compatible API。",
		Driver:      DriverS3,
		Provider:    ProviderBackblazeB2,
		Defaults:    map[string]any{"endpoint": "s3.us-west-000.backblazeb2.com", "region": "us-west-000", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderWasabi,
		Name:        "Wasabi",
		Description: "Wasabi 对象存储。",
		Driver:      DriverS3,
		Provider:    ProviderWasabi,
		Defaults:    map[string]any{"endpoint": "s3.wasabisys.com", "region": "us-east-1", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderDOSpaces,
		Name:        "DigitalOcean Spaces",
		Description: "DigitalOcean 的 S3 兼容对象存储。",
		Driver:      DriverS3,
		Provider:    ProviderDOSpaces,
		Defaults:    map[string]any{"endpoint": "nyc3.digitaloceanspaces.com", "region": "nyc3", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
	{
		Key:         ProviderCustomS3,
		Name:        "自定义 S3 兼容",
		Description: "适配任意 S3 兼容对象存储服务。",
		Driver:      DriverS3,
		Provider:    ProviderCustomS3,
		Defaults:    map[string]any{"endpoint": "", "region": "auto", "bucket": "", "access_key_id": "", "secret_access_key": "", "base_url": "", "force_path_style": false},
	},
}

// Catalog returns a stable copy of the user-facing storage scheme list.
func Catalog() []Scheme {
	out := make([]Scheme, 0, len(schemeCatalog))
	for _, scheme := range schemeCatalog {
		out = append(out, scheme.clone())
	}
	return out
}

// SchemeByKey looks up a scheme by key and accepts a few legacy aliases for compatibility.
func SchemeByKey(key string) (Scheme, bool) {
	normalized := normalizeSchemeKey(key)
	for _, scheme := range schemeCatalog {
		if scheme.Key == normalized {
			return scheme.clone(), true
		}
	}
	return Scheme{}, false
}

// SchemeKeyForStoredConfig maps a persisted storage row to the UI scheme key.
func SchemeKeyForStoredConfig(driver string, provider *string, rawConfig string) string {
	canonicalDriver, canonicalProvider := CanonicalDriverAndProvider(driver, provider, rawConfig)
	switch canonicalDriver {
	case DriverLocal:
		return DriverLocal
	case DriverFTP:
		return DriverFTP
	case DriverS3:
		if canonicalProvider == "" {
			return ProviderCustomS3
		}
		if _, ok := SchemeByKey(canonicalProvider); ok {
			return canonicalProvider
		}
		return ProviderCustomS3
	default:
		return canonicalDriver
	}
}

// ResolveSchemeConfig merges a user payload with scheme defaults, returning the canonical
// driver/provider pair plus the normalized JSON persisted in the database.
func ResolveSchemeConfig(key string, raw json.RawMessage) (string, *string, json.RawMessage, StorageConfig, error) {
	scheme, ok := SchemeByKey(key)
	if !ok {
		return "", nil, nil, StorageConfig{}, fmt.Errorf("unknown storage scheme %q", key)
	}

	normalizedRaw, err := mergeConfig(scheme.Defaults, raw)
	if err != nil {
		return "", nil, nil, StorageConfig{}, err
	}

	scfg, err := ParseConfig(scheme.Driver, normalizedRaw)
	if err != nil {
		return "", nil, nil, StorageConfig{}, err
	}

	var provider *string
	if scheme.Provider != "" {
		p := scheme.Provider
		provider = &p
	}

	return scheme.Driver, provider, normalizedRaw, scfg, nil
}

func normalizeSchemeKey(key string) string {
	switch strings.TrimSpace(strings.ToLower(key)) {
	case DriverLocal, "local-storage":
		return DriverLocal
	case DriverFTP:
		return DriverFTP
	case DriverS3, ProviderCustomS3, "custom":
		return ProviderCustomS3
	case DriverOSS, ProviderAliyunOSS:
		return ProviderAliyunOSS
	case DriverCOS, ProviderTencentCOS:
		return ProviderTencentCOS
	case DriverOBS, ProviderHuaweiOBS:
		return ProviderHuaweiOBS
	case DriverBOS, ProviderBaiduBOS:
		return ProviderBaiduBOS
	default:
		return strings.TrimSpace(strings.ToLower(key))
	}
}

func mergeConfig(defaults map[string]any, raw json.RawMessage) (json.RawMessage, error) {
	merged := make(map[string]any, len(defaults))
	for k, v := range defaults {
		merged[k] = v
	}

	if len(raw) == 0 || string(raw) == "null" {
		out, err := json.Marshal(merged)
		if err != nil {
			return nil, fmt.Errorf("marshal default scheme config: %w", err)
		}
		return out, nil
	}

	var user map[string]any
	if err := json.Unmarshal(raw, &user); err != nil {
		return nil, fmt.Errorf("parse scheme config: %w", err)
	}
	for k, v := range user {
		merged[k] = v
	}

	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal scheme config: %w", err)
	}
	return out, nil
}

func (s Scheme) clone() Scheme {
	copyDefaults := make(map[string]any, len(s.Defaults))
	for k, v := range s.Defaults {
		copyDefaults[k] = v
	}
	s.Defaults = copyDefaults
	return s
}

// KnownS3Providers returns the persisted provider set accepted for driver=s3.
func KnownS3Providers() []string {
	out := make([]string, 0, len(schemeCatalog))
	for _, scheme := range schemeCatalog {
		if scheme.Driver == DriverS3 && scheme.Provider != "" {
			out = append(out, scheme.Provider)
		}
	}
	slices.Sort(out)
	return out
}
