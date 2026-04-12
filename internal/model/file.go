package model

import "time"

// File 文件模型，支持图片、视频、音频和任意静态文件。
type File struct {
	ID              string    `gorm:"column:id;primaryKey" json:"id"`
	UserID          string    `gorm:"column:user_id;index;not null" json:"user_id"`
	AlbumID         *string   `gorm:"column:album_id;index" json:"album_id"`                 // 可选，归属相册
	StorageConfigID string    `gorm:"column:storage_config_id;not null" json:"storage_config_id"`
	OriginalName    string    `gorm:"column:original_name;not null" json:"original_name"`     // 用户上传时的原始文件名
	StorageKey      string    `gorm:"column:storage_key;not null" json:"storage_key"`         // 存储后端的 key 路径
	HashMD5         string    `gorm:"column:hash_md5;index;not null" json:"hash_md5"`         // 文件 MD5，用于去重
	SizeBytes       int64     `gorm:"column:size_bytes;not null" json:"size_bytes"`
	MimeType        string    `gorm:"column:mime_type;not null" json:"mime_type"`              // 真实 MIME 类型
	FileType        string    `gorm:"column:file_type;not null" json:"file_type"`              // image / video / audio / file
	Width           *int      `gorm:"column:width" json:"width,omitempty"`                    // 仅图片
	Height          *int      `gorm:"column:height" json:"height,omitempty"`                  // 仅图片
	Duration        *int      `gorm:"column:duration" json:"duration,omitempty"`              // 仅视频/音频，单位秒
	URL             string    `gorm:"column:url;not null" json:"url"`                         // 访问 URL
	ThumbURL        *string   `gorm:"column:thumb_url" json:"thumb_url,omitempty"`            // 缩略图 URL，仅图片
	IsDeleted       bool      `gorm:"column:is_deleted;default:false" json:"is_deleted"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (File) TableName() string { return "files" }

// FileType 常量定义。
const (
	FileTypeImage = "image"
	FileTypeVideo = "video"
	FileTypeAudio = "audio"
	FileTypeFile  = "file"
)
