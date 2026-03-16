package model

const (
	FriendLinkStatusActive  = "active"
	FriendLinkStatusPending = "pending"
	FriendLinkStatusDown    = "down"
)

type FriendLink struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	URL         string `gorm:"size:1024;not null;uniqueIndex" json:"url"`
	Description string `gorm:"type:text" json:"description"`
	Logo        string `gorm:"size:1024" json:"logo"`
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`
	Status      string `gorm:"size:32;not null;default:active;index" json:"status"` // active / pending / down
}
