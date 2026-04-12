package model

// Setting 系统设置模型，key-value 形式存储全局配置。
// 用于持久化站点配置、上传限制、认证策略等运行时配置项。
type Setting struct {
	Key   string `gorm:"column:key;primaryKey" json:"key"`
	Value string `gorm:"column:value;not null" json:"value"`
}

func (Setting) TableName() string { return "settings" }
