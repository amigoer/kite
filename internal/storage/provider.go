package storage

import (
	"encoding/json"
	"strings"
)

// Stable provider identifiers persisted in storage_configs.provider.
const (
	ProviderAWSS3        = "aws-s3"
	ProviderAliyunOSS    = "aliyun-oss"
	ProviderTencentCOS   = "tencent-cos"
	ProviderHuaweiOBS    = "huawei-obs"
	ProviderBaiduBOS     = "baidu-bos"
	ProviderCloudflareR2 = "cloudflare-r2"
	ProviderGoogleGCS    = "google-gcs"
	ProviderBackblazeB2  = "backblaze-b2"
	ProviderWasabi       = "wasabi"
	ProviderDOSpaces     = "do-spaces"
	ProviderMinIO        = "minio"
	ProviderScaleway     = "scaleway"
	ProviderUCloudUS3    = "ucloud-us3"
	ProviderJDCloudOSS   = "jdcloud-oss"
	ProviderCustomS3     = "custom-s3"
)

// NormalizeProvider folds aliases into the canonical provider identifier stored in the database.
func NormalizeProvider(provider string) string {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "", "none", "null":
		return ""
	case "aws", "amazon", "amazon-s3", "aws-s3":
		return ProviderAWSS3
	case "aliyun", "alibaba", "aliyun-oss", "oss":
		return ProviderAliyunOSS
	case "tencent", "tencent-cos", "cos":
		return ProviderTencentCOS
	case "huawei", "huawei-obs", "obs":
		return ProviderHuaweiOBS
	case "baidu", "baidu-bos", "bos":
		return ProviderBaiduBOS
	case "cloudflare", "r2", "cloudflare-r2":
		return ProviderCloudflareR2
	case "gcp", "google", "google-gcs", "gcs":
		return ProviderGoogleGCS
	case "backblaze", "b2", "backblaze-b2":
		return ProviderBackblazeB2
	case "wasabi":
		return ProviderWasabi
	case "do", "do-spaces", "digitalocean", "digitalocean-spaces":
		return ProviderDOSpaces
	case "minio":
		return ProviderMinIO
	case "scaleway":
		return ProviderScaleway
	case "ucloud", "ucloud-us3", "us3":
		return ProviderUCloudUS3
	case "jdcloud", "jdcloud-oss":
		return ProviderJDCloudOSS
	case "custom", "custom-s3", "s3":
		return ProviderCustomS3
	default:
		return strings.TrimSpace(strings.ToLower(provider))
	}
}

// DetectProvider returns the effective provider identifier used by the frontend for logos and labels.
// Explicit database values win; when absent, legacy drivers and S3 endpoints are used as fallbacks.
func DetectProvider(driver string, provider *string, rawConfig string) string {
	if provider != nil {
		if normalized := NormalizeProvider(*provider); normalized != "" {
			return normalized
		}
	}

	switch driver {
	case DriverLocal, DriverFTP:
		return ""
	case DriverOSS:
		return ProviderAliyunOSS
	case DriverCOS:
		return ProviderTencentCOS
	case DriverOBS:
		return ProviderHuaweiOBS
	case DriverBOS:
		return ProviderBaiduBOS
	case DriverS3:
		return detectS3Provider(rawConfig)
	default:
		return ""
	}
}

// CanonicalDriverAndProvider normalises legacy rows into the new storage domain model.
func CanonicalDriverAndProvider(driver string, provider *string, rawConfig string) (string, string) {
	switch strings.TrimSpace(strings.ToLower(driver)) {
	case DriverLocal:
		return DriverLocal, ""
	case DriverFTP:
		return DriverFTP, ""
	case DriverOSS:
		return DriverS3, ProviderAliyunOSS
	case DriverCOS:
		return DriverS3, ProviderTencentCOS
	case DriverOBS:
		return DriverS3, ProviderHuaweiOBS
	case DriverBOS:
		return DriverS3, ProviderBaiduBOS
	case DriverS3:
		p := DetectProvider(DriverS3, provider, rawConfig)
		if p == "" {
			p = ProviderCustomS3
		}
		return DriverS3, p
	default:
		p := DetectProvider(driver, provider, rawConfig)
		return strings.TrimSpace(strings.ToLower(driver)), p
	}
}

func detectS3Provider(rawConfig string) string {
	if rawConfig == "" {
		return ProviderCustomS3
	}

	var c struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.Unmarshal([]byte(rawConfig), &c); err != nil {
		return ProviderCustomS3
	}

	ep := strings.ToLower(c.Endpoint)
	switch {
	case strings.Contains(ep, "r2.cloudflarestorage.com"):
		return ProviderCloudflareR2
	case strings.Contains(ep, "amazonaws.com"):
		return ProviderAWSS3
	case strings.Contains(ep, "backblazeb2.com"):
		return ProviderBackblazeB2
	case strings.Contains(ep, "wasabisys.com"):
		return ProviderWasabi
	case strings.Contains(ep, "digitaloceanspaces.com"):
		return ProviderDOSpaces
	case strings.Contains(ep, "scw.cloud"):
		return ProviderScaleway
	case strings.Contains(ep, "aliyuncs.com"):
		return ProviderAliyunOSS
	case strings.Contains(ep, "myqcloud.com"):
		return ProviderTencentCOS
	case strings.Contains(ep, "myhuaweicloud.com"):
		return ProviderHuaweiOBS
	case strings.Contains(ep, "bcebos.com"):
		return ProviderBaiduBOS
	case strings.Contains(ep, "googleapis.com"):
		return ProviderGoogleGCS
	case strings.Contains(ep, "ufileos.com") || strings.Contains(ep, "us3.ucloud"):
		return ProviderUCloudUS3
	case strings.Contains(ep, "jdcloud-oss.com"):
		return ProviderJDCloudOSS
	case strings.Contains(ep, "min.io") || strings.Contains(ep, "minio"):
		return ProviderMinIO
	default:
		return ProviderCustomS3
	}
}
