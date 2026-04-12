package model

import "time"

// APIToken API 令牌模型。
// 用户可创建 API Token 用于第三方工具（如 PicGo）调用上传接口。
// 为安全起见，只存储 token 的 SHA256 哈希，不存储明文。
type APIToken struct {
	ID        string     `gorm:"column:id;primaryKey" json:"id"`
	UserID    string     `gorm:"column:user_id;index;not null" json:"user_id"`
	Name      string     `gorm:"column:name;not null" json:"name"`         // Token 名称，用于用户识别
	TokenHash string     `gorm:"column:token_hash;uniqueIndex;not null" json:"-"` // SHA256(token)
	LastUsed  *time.Time `gorm:"column:last_used" json:"last_used,omitempty"`
	ExpiresAt *time.Time `gorm:"column:expires_at" json:"expires_at,omitempty"`   // NULL 表示永不过期
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (APIToken) TableName() string { return "api_tokens" }

// IsExpired 检查 Token 是否已过期。
func (t *APIToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}
