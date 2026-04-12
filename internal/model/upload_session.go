package model

import "time"

// UploadSession tus 断点续传会话模型。
// 记录大文件上传的进度，支持客户端断点续传。
// 会话有 24 小时有效期，过期后由后台任务清理。
type UploadSession struct {
	ID         string    `gorm:"column:id;primaryKey" json:"id"`           // tus upload ID
	UserID     string    `gorm:"column:user_id;index;not null" json:"user_id"`
	Filename   string    `gorm:"column:filename;not null" json:"filename"`
	SizeBytes  int64     `gorm:"column:size_bytes;not null" json:"size_bytes"`
	Offset     int64     `gorm:"column:offset;default:0" json:"offset"`
	MimeType   string    `gorm:"column:mime_type" json:"mime_type"`
	Status     string    `gorm:"column:status;default:pending" json:"status"` // pending / uploading / complete / failed
	StorageKey string    `gorm:"column:storage_key" json:"storage_key"`
	ExpiresAt  time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (UploadSession) TableName() string { return "upload_sessions" }

// Upload session 状态常量。
const (
	UploadStatusPending   = "pending"
	UploadStatusUploading = "uploading"
	UploadStatusComplete  = "complete"
	UploadStatusFailed    = "failed"
)
