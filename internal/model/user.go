package model

import "time"

// User 用户模型。
// 支持管理员和普通用户两种角色，普通用户可通过注册页面自行创建账号。
type User struct {
	ID                 string    `gorm:"column:id;primaryKey" json:"id"`
	Username           string    `gorm:"column:username;uniqueIndex;not null" json:"username"`
	Email              string    `gorm:"column:email;uniqueIndex;not null" json:"email"`
	PasswordHash       string    `gorm:"column:password_hash;not null" json:"-"`
	Role               string    `gorm:"column:role;default:user" json:"role"`                            // admin / user
	StorageLimit       int64     `gorm:"column:storage_limit;default:10737418240" json:"storage_limit"`   // 默认 10GB，-1 无限
	StorageUsed        int64     `gorm:"column:storage_used;default:0" json:"storage_used"`
	IsActive           bool      `gorm:"column:is_active;default:true" json:"is_active"`
	PasswordMustChange bool      `gorm:"column:password_must_change;default:false" json:"password_must_change"` // 首次登录必须重置账号密码
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "users" }

// HasStorageSpace 检查用户是否有足够的存储空间上传指定大小的文件。
func (u *User) HasStorageSpace(fileSize int64) bool {
	if u.StorageLimit == -1 {
		return true
	}
	return u.StorageUsed+fileSize <= u.StorageLimit
}
