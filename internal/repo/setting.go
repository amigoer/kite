package repo

import (
	"context"
	"fmt"

	"github.com/amigoer/kite/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SettingRepo 系统设置数据访问层。
type SettingRepo struct {
	db *gorm.DB
}

func NewSettingRepo(db *gorm.DB) *SettingRepo {
	return &SettingRepo{db: db}
}

// Get 获取指定 key 的配置值。
func (r *SettingRepo) Get(ctx context.Context, key string) (string, error) {
	var setting model.Setting
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error; err != nil {
		return "", fmt.Errorf("get setting %q: %w", key, err)
	}
	return setting.Value, nil
}

// Set 设置指定 key 的值，不存在则创建。
func (r *SettingRepo) Set(ctx context.Context, key, value string) error {
	setting := model.Setting{Key: key, Value: value}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(&setting).Error; err != nil {
		return fmt.Errorf("set setting %q: %w", key, err)
	}
	return nil
}

// GetAll 获取所有配置。
func (r *SettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	var settings []model.Setting
	if err := r.db.WithContext(ctx).Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	result := make(map[string]string, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

// SetBatch 批量设置配置。
func (r *SettingRepo) SetBatch(ctx context.Context, settings map[string]string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range settings {
			setting := model.Setting{Key: key, Value: value}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(&setting).Error; err != nil {
				return fmt.Errorf("set setting %q: %w", key, err)
			}
		}
		return nil
	})
}

// Delete 删除指定 key 的配置。
func (r *SettingRepo) Delete(ctx context.Context, key string) error {
	if err := r.db.WithContext(ctx).Where("key = ?", key).Delete(&model.Setting{}).Error; err != nil {
		return fmt.Errorf("delete setting %q: %w", key, err)
	}
	return nil
}
