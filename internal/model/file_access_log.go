package model

import "time"

// FileAccessLog records each file access or download for bandwidth and traffic statistics.
// UserID stores the file owner (not the visitor) so stats can be segmented per owner; empty for anonymous uploads.
type FileAccessLog struct {
	ID          string    `gorm:"column:id;primaryKey" json:"id"`
	FileID      string    `gorm:"column:file_id;index;not null" json:"file_id"`
	UserID      string    `gorm:"column:user_id;index" json:"user_id"`
	BytesServed int64     `gorm:"column:bytes_served;not null" json:"bytes_served"`
	AccessedAt  time.Time `gorm:"column:accessed_at;autoCreateTime;index" json:"accessed_at"`
}

func (FileAccessLog) TableName() string { return "file_access_logs" }
