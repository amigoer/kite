package model

// Setting is the key/value model backing the runtime configuration store.
// It persists site config, upload limits, authentication policy, and other runtime options.
type Setting struct {
	Key   string `gorm:"column:key;primaryKey" json:"key"`
	Value string `gorm:"column:value;not null" json:"value"`
}

func (Setting) TableName() string { return "settings" }
