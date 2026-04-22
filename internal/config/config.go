package config

import "time"

// Config is the application-wide configuration, loaded from the settings table or defaults.
// The setup wizard writes the initial values to the database on first install.
type Config struct {
	// Site settings.
	Site SiteConfig `json:"site"`

	// Upload settings.
	Upload UploadConfig `json:"upload"`

	// Authentication settings.
	Auth AuthConfig `json:"auth"`

	// HTTP server settings.
	Server ServerConfig `json:"server"`

	// Database settings.
	Database DatabaseConfig `json:"database"`
}

// SiteConfig holds basic site metadata.
type SiteConfig struct {
	Name string `json:"name"` // site display name
	URL  string `json:"url"`  // site URL (e.g. https://kite.plus), used to build file access links
}

// UploadConfig holds the upload-pipeline configuration.
type UploadConfig struct {
	MaxFileSize    int64    `json:"max_file_size"`   // max bytes per file, default 100MB
	AllowedTypes   []string `json:"allowed_types"`   // allowed MIME prefixes (e.g. ["image/", "video/"]); empty allows all
	ForbiddenExts  []string `json:"forbidden_exts"`  // blocked file extensions (e.g. [".exe", ".bat"])
	AutoWebP       bool     `json:"auto_webp"`       // convert uploaded images to WebP
	ThumbWidth     int      `json:"thumb_width"`     // thumbnail width, default 300px
	ThumbQuality   int      `json:"thumb_quality"`   // thumbnail quality 1-100, default 80
	PathPattern    string   `json:"path_pattern"`    // storage key template, default "{year}/{month}/{md5_8}/{uuid}.{ext}"
	AllowDuplicate bool     `json:"allow_duplicate"` // allow duplicate files with identical MD5
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	JWTSecret          string        `json:"jwt_secret"`           // JWT signing secret
	AccessTokenExpiry  time.Duration `json:"access_token_expiry"`  // access-token lifetime, default 2h
	RefreshTokenExpiry time.Duration `json:"refresh_token_expiry"` // refresh-token lifetime, default 7d
	AllowRegistration  bool          `json:"allow_registration"`   // allow self-service user registration
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string `json:"host"` // listen address, default "0.0.0.0"
	Port int    `json:"port"` // listen port, default 8080
}

// DatabaseConfig holds database settings.
type DatabaseConfig struct {
	Driver string `json:"driver"` // sqlite / mysql / postgres
	// DSN is the database connection string; the format depends on the driver:
	//   sqlite:   "data/kite.db"
	//   mysql:    "user:pass@tcp(host:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	//   postgres: "host=localhost user=postgres password=pass dbname=kite port=5432 sslmode=disable"
	DSN string `json:"dsn"`
}

// DefaultConfig returns a Config populated with default values.
func DefaultConfig() Config {
	return Config{
		Site: SiteConfig{
			Name: "Kite",
			URL:  "http://localhost:8080",
		},
		Upload: UploadConfig{
			MaxFileSize:    100 * 1024 * 1024, // 100MB
			AllowedTypes:   nil,               // allow all types
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
			// New installs default to closed registration. An operator who
			// stands up a fresh instance hasn't had a chance to configure
			// rate limiting, captchas, or email verification yet, so leaving
			// the front door open risks abuse from day one. Admins can flip
			// this on from the settings page once they're ready for public
			// signups.
			AllowRegistration: false,
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
