package model

import "time"

// UploadSession is the tus resumable-upload session model.
// It tracks the progress of large uploads so clients can resume after an interruption.
// Sessions live for 24 hours and are cleaned up by a background task once expired.
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

// Upload-session status constants.
const (
	UploadStatusPending   = "pending"
	UploadStatusUploading = "uploading"
	UploadStatusComplete  = "complete"
	UploadStatusFailed    = "failed"
)
