package model

import "time"

// File is the model for every uploaded asset (images, videos, audio, and arbitrary files).
type File struct {
	ID              string    `gorm:"column:id;primaryKey" json:"id"`
	UserID          string    `gorm:"column:user_id;index;not null" json:"user_id"`
	AlbumID         *string   `gorm:"column:album_id;index" json:"album_id"`                 // optional album/folder
	StorageConfigID string    `gorm:"column:storage_config_id;not null" json:"storage_config_id"`
	OriginalName    string    `gorm:"column:original_name;not null" json:"original_name"`     // filename at upload time
	StorageKey      string    `gorm:"column:storage_key;not null" json:"storage_key"`         // object key in the storage backend
	HashMD5         string    `gorm:"column:hash_md5;index;not null" json:"hash_md5"`         // MD5 digest; used for dedupe
	SizeBytes       int64     `gorm:"column:size_bytes;not null" json:"size_bytes"`
	MimeType        string    `gorm:"column:mime_type;not null" json:"mime_type"`              // detected MIME type
	FileType        string    `gorm:"column:file_type;not null" json:"file_type"`              // image / video / audio / file
	Width           *int      `gorm:"column:width" json:"width,omitempty"`                    // images only
	Height          *int      `gorm:"column:height" json:"height,omitempty"`                  // images only
	Duration        *int      `gorm:"column:duration" json:"duration,omitempty"`              // video/audio only, seconds
	URL             string    `gorm:"column:url;not null" json:"url"`                         // access URL
	ThumbURL        *string   `gorm:"column:thumb_url" json:"thumb_url,omitempty"`            // thumbnail URL, images only
	IsDeleted       bool      `gorm:"column:is_deleted;default:false" json:"is_deleted"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (File) TableName() string { return "files" }

// File-type constants.
const (
	FileTypeImage = "image"
	FileTypeVideo = "video"
	FileTypeAudio = "audio"
	FileTypeFile  = "file"
)
