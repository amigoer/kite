package model

// 通知类型
const (
	NotificationTypeComment = "comment" // 新评论
	NotificationTypeSystem  = "system"  // 系统通知
)

// Notification 通知模型
type Notification struct {
	BaseModel
	Type    string `gorm:"size:32;not null;index" json:"type"`
	Title   string `gorm:"size:255;not null" json:"title"`
	Content string `gorm:"size:1024" json:"content"`
	Link    string `gorm:"size:512" json:"link"`    // 点击跳转路径，如 /admin/comments
	IsRead  bool   `gorm:"not null;default:false;index" json:"is_read"`
}
