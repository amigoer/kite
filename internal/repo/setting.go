package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amigoer/kite/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SettingRepo is the data access layer for system settings.
type SettingRepo struct {
	db *gorm.DB
}

func NewSettingRepo(db *gorm.DB) *SettingRepo {
	return &SettingRepo{db: db}
}

// Get returns the value stored under key.
func (r *SettingRepo) Get(ctx context.Context, key string) (string, error) {
	var setting model.Setting
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error; err != nil {
		return "", fmt.Errorf("get setting %q: %w", key, err)
	}
	return setting.Value, nil
}

// GetOrDefault returns the stored value or the provided fallback when the key
// does not exist yet.
func (r *SettingRepo) GetOrDefault(ctx context.Context, key, fallback string) (string, error) {
	val, err := r.Get(ctx, key)
	if err == nil {
		return val, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fallback, nil
	}
	return "", err
}

// GetBool returns the parsed boolean value stored under key. Missing keys fall
// back to the provided default.
func (r *SettingRepo) GetBool(ctx context.Context, key string, fallback bool) (bool, error) {
	def := "false"
	if fallback {
		def = "true"
	}
	val, err := r.GetOrDefault(ctx, key, def)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(val), "true"), nil
}

// Set upserts the value for key.
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

// GetAll returns every setting as a key/value map.
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

// SetBatch upserts multiple settings in a single transaction.
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

// Delete removes the setting identified by key.
func (r *SettingRepo) Delete(ctx context.Context, key string) error {
	if err := r.db.WithContext(ctx).Where("key = ?", key).Delete(&model.Setting{}).Error; err != nil {
		return fmt.Errorf("delete setting %q: %w", key, err)
	}
	return nil
}
