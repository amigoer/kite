package model

import "time"

// Album is a folder record stored in the albums table; the table is reused to represent folder hierarchies.
type Album struct {
	ID          string    `gorm:"column:id;primaryKey" json:"id"`
	UserID      string    `gorm:"column:user_id;index;not null" json:"user_id"`
	ParentID    *string   `gorm:"column:parent_id;index" json:"parent_id,omitempty"`
	Name        string    `gorm:"column:name;not null" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	IsPublic    bool      `gorm:"column:is_public;default:false" json:"is_public"`
	CoverURL    *string   `gorm:"column:cover_url" json:"cover_url,omitempty"`
	FileCount   int64     `gorm:"-" json:"file_count,omitempty"` // not persisted; computed at query time
	FolderCount int64     `gorm:"-" json:"folder_count,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Album) TableName() string { return "albums" }
