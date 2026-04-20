package model

import "time"

// FileReplica is a replica record for a file.
// Rows exist only when the upload policy is mirror (dual- or multi-copy writes); the primary
// location is still tracked on files.storage_config_id. The replica shares files.storage_key,
// so every backend uses the same object path.
type FileReplica struct {
	ID              string    `gorm:"column:id;primaryKey" json:"id"`
	FileID          string    `gorm:"column:file_id;index;not null" json:"file_id"`
	StorageConfigID string    `gorm:"column:storage_config_id;index;not null" json:"storage_config_id"`
	Status          string    `gorm:"column:status;not null;default:'pending'" json:"status"` // pending / ok / failed
	ErrorMsg        string    `gorm:"column:error_msg" json:"error_msg,omitempty"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (FileReplica) TableName() string { return "file_replicas" }

// Replica-status constants.
const (
	ReplicaStatusPending = "pending"
	ReplicaStatusOK      = "ok"
	ReplicaStatusFailed  = "failed"
)
