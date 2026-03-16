package model

type Category struct {
	BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	Slug        string `gorm:"size:255;not null;uniqueIndex" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	// PostCount 不存数据库，通过 SQL 聚合查询填充
	PostCount int64 `gorm:"-" json:"post_count"`
}
