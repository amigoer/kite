package model

import "time"

// Album 相册模型，用于对文件进行分组管理。
// 每个用户可以创建多个相册，文件可以归属到某个相册中。
type Album struct {
	ID          string    `gorm:"column:id;primaryKey" json:"id"`
	UserID      string    `gorm:"column:user_id;index;not null" json:"user_id"`
	Name        string    `gorm:"column:name;not null" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	IsPublic    bool      `gorm:"column:is_public;default:false" json:"is_public"`
	CoverURL    *string   `gorm:"column:cover_url" json:"cover_url,omitempty"`
	FileCount   int64     `gorm:"-" json:"file_count,omitempty"` // 非数据库字段，查询时计算
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Album) TableName() string { return "albums" }
