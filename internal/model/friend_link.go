package model

type FriendLink struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	URL         string `gorm:"size:1024;not null;uniqueIndex" json:"url"`
	Description string `gorm:"type:text" json:"description"`
	Logo        string `gorm:"size:1024" json:"logo"`
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`
	IsActive    bool   `gorm:"not null;default:true;index" json:"is_active"`
}
