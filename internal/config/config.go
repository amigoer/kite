package config

import "time"

// Config 应用全局配置，从数据库 settings 表加载或使用默认值。
// 首次安装时通过 Web 向导写入数据库。
type Config struct {
	// 站点配置
	Site SiteConfig `json:"site"`

	// 上传配置
	Upload UploadConfig `json:"upload"`

	// 认证配置
	Auth AuthConfig `json:"auth"`

	// 服务器配置
	Server ServerConfig `json:"server"`

	// 数据库配置
	Database DatabaseConfig `json:"database"`
}

// SiteConfig 站点基础信息配置。
type SiteConfig struct {
	Name string `json:"name"` // 站点名称
	URL  string `json:"url"`  // 站点 URL（如 https://kite.plus），用于生成文件访问链接
}

// UploadConfig 文件上传相关配置。
type UploadConfig struct {
	MaxFileSize     int64    `json:"max_file_size"`     // 单文件最大字节数，默认 100MB
	AllowedTypes    []string `json:"allowed_types"`     // 允许的 MIME 类型前缀，如 ["image/", "video/", "audio/"]，空表示允许所有
	ForbiddenExts   []string `json:"forbidden_exts"`    // 禁止的文件扩展名，如 [".exe", ".bat"]
	AutoWebP        bool     `json:"auto_webp"`         // 是否自动将图片转换为 WebP
	ThumbWidth      int      `json:"thumb_width"`       // 缩略图宽度，默认 300px
	ThumbQuality    int      `json:"thumb_quality"`     // 缩略图质量 1-100，默认 80
	PathPattern     string   `json:"path_pattern"`      // 存储路径模板，默认 "{year}/{month}/{md5_8}/{uuid}.{ext}"
	AllowDuplicate  bool     `json:"allow_duplicate"`   // 是否允许重复文件（相同 MD5）
}

// AuthConfig 认证相关配置。
type AuthConfig struct {
	JWTSecret          string        `json:"jwt_secret"`           // JWT 签名密钥
	AccessTokenExpiry  time.Duration `json:"access_token_expiry"`  // Access Token 有效期，默认 2 小时
	RefreshTokenExpiry time.Duration `json:"refresh_token_expiry"` // Refresh Token 有效期，默认 7 天
	AllowRegistration  bool          `json:"allow_registration"`   // 是否允许用户自行注册
}

// ServerConfig HTTP 服务器配置。
type ServerConfig struct {
	Host string `json:"host"` // 监听地址，默认 "0.0.0.0"
	Port int    `json:"port"` // 监听端口，默认 8080
}

// DatabaseConfig 数据库配置。
type DatabaseConfig struct {
	Driver string `json:"driver"` // sqlite / postgres
	DSN    string `json:"dsn"`    // 数据库连接字符串，sqlite 默认 "data/kite.db"
}

// DefaultConfig 返回带有默认值的配置。
func DefaultConfig() Config {
	return Config{
		Site: SiteConfig{
			Name: "Kite",
			URL:  "http://localhost:8080",
		},
		Upload: UploadConfig{
			MaxFileSize:    100 * 1024 * 1024, // 100MB
			AllowedTypes:   nil,               // 允许所有类型
			ForbiddenExts:  []string{".exe", ".bat", ".cmd", ".sh", ".ps1"},
			AutoWebP:       false,
			ThumbWidth:     300,
			ThumbQuality:   80,
			PathPattern:    "{year}/{month}/{md5_8}/{uuid}.{ext}",
			AllowDuplicate: false,
		},
		Auth: AuthConfig{
			AccessTokenExpiry:  2 * time.Hour,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			AllowRegistration:  true,
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "data/kite.db",
		},
	}
}
