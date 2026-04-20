package model

import "time"

// StorageConfig is the storage configuration model.
// The system persists a small, stable domain model:
//   - Driver: the runtime implementation family (local / s3 / ftp)
//   - Provider: an optional vendor preset layered on top of the driver
//   - Config: the driver-specific payload as JSON
//
// Only S3-compatible object storage uses Provider; local and FTP leave it NULL.
type StorageConfig struct {
	ID                 string    `gorm:"column:id;primaryKey" json:"id"`
	Name               string    `gorm:"column:name;not null" json:"name"`
	Driver             string    `gorm:"column:driver;not null" json:"driver"`            // local / s3 / ftp
	Provider           *string   `gorm:"column:provider;index" json:"provider,omitempty"` // optional vendor preset; only used by s3
	Config             string    `gorm:"column:config;not null" json:"-"`
	CapacityLimitBytes int64     `gorm:"column:capacity_limit_bytes;default:0" json:"capacity_limit_bytes"` // capacity cap in bytes; 0 means unlimited
	Priority           int       `gorm:"column:priority;default:100;index" json:"priority"`                 // lower number = higher priority; drives fallback / round_robin / mirror
	IsDefault          bool      `gorm:"column:is_default;default:false" json:"is_default"`
	IsActive           bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (StorageConfig) TableName() string { return "storage_configs" }
