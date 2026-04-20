package model

import "time"

// APIToken is the API-token model.
// Users mint API tokens to let third-party tools (e.g. PicGo) call the upload endpoints.
// For safety only the SHA256 digest is stored; the plaintext is shown to the user exactly once.
type APIToken struct {
	ID        string     `gorm:"column:id;primaryKey" json:"id"`
	UserID    string     `gorm:"column:user_id;index;not null" json:"user_id"`
	Name      string     `gorm:"column:name;not null" json:"name"`         // human-readable token label
	TokenHash string     `gorm:"column:token_hash;uniqueIndex;not null" json:"-"` // SHA256(token)
	LastUsed  *time.Time `gorm:"column:last_used" json:"last_used,omitempty"`
	ExpiresAt *time.Time `gorm:"column:expires_at" json:"expires_at,omitempty"`   // NULL means never expires
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (APIToken) TableName() string { return "api_tokens" }

// IsExpired reports whether the token has passed its expiry time.
func (t *APIToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}
