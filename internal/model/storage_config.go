package model

import "time"

// StorageConfig 存储配置模型。
// 系统支持多个存储配置，管理员可以配置本地或 S3 兼容存储。
// config 字段为 JSON 序列化的配置详情。
type StorageConfig struct {
	ID        string    `gorm:"column:id;primaryKey" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`       // 配置名称（如"阿里云 OSS"、"本地存储"）
	Driver    string    `gorm:"column:driver;not null" json:"driver"`   // local / s3
	Config    string    `gorm:"column:config;not null" json:"-"`        // JSON 格式的配置详情，不直接暴露给前端
	IsDefault bool      `gorm:"column:is_default;default:false" json:"is_default"`
	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (StorageConfig) TableName() string { return "storage_configs" }
