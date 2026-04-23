package repo

import (
	"context"
	"fmt"

	"github.com/kite-plus/kite/internal/model"
	"github.com/kite-plus/kite/internal/storage"
	"gorm.io/gorm"
)

// StorageConfigRepo is the data access layer for storage configurations.
type StorageConfigRepo struct {
	db *gorm.DB
}

func NewStorageConfigRepo(db *gorm.DB) *StorageConfigRepo {
	return &StorageConfigRepo{db: db}
}

// Create inserts a new storage configuration.
func (r *StorageConfigRepo) Create(ctx context.Context, cfg *model.StorageConfig) error {
	if err := r.db.WithContext(ctx).Create(cfg).Error; err != nil {
		return fmt.Errorf("create storage config: %w", err)
	}
	return nil
}

// GetByID fetches a storage configuration by ID.
func (r *StorageConfigRepo) GetByID(ctx context.Context, id string) (*model.StorageConfig, error) {
	if err := r.normalizeLegacyConfigs(ctx); err != nil {
		return nil, err
	}
	var cfg model.StorageConfig
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("get storage config: %w", err)
	}
	return &cfg, nil
}

// GetDefault returns the default storage configuration.
func (r *StorageConfigRepo) GetDefault(ctx context.Context) (*model.StorageConfig, error) {
	if err := r.normalizeLegacyConfigs(ctx); err != nil {
		return nil, err
	}
	var cfg model.StorageConfig
	if err := r.db.WithContext(ctx).
		Where("is_default = ? AND is_active = ?", true, true).
		First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("get default storage config: %w", err)
	}
	return &cfg, nil
}

// Update persists changes to a storage configuration.
func (r *StorageConfigRepo) Update(ctx context.Context, cfg *model.StorageConfig) error {
	if err := r.db.WithContext(ctx).Save(cfg).Error; err != nil {
		return fmt.Errorf("update storage config: %w", err)
	}
	return nil
}

// SetDefault marks the given configuration as the default and clears the default flag on all others.
func (r *StorageConfigRepo) SetDefault(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Clear any existing default.
		if err := tx.Model(&model.StorageConfig{}).
			Where("is_default = ?", true).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("clear default storage config: %w", err)
		}
		// Set the new default.
		if err := tx.Model(&model.StorageConfig{}).
			Where("id = ?", id).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("set default storage config: %w", err)
		}
		return nil
	})
}

// Delete removes a storage configuration.
func (r *StorageConfigRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.StorageConfig{}).Error; err != nil {
		return fmt.Errorf("delete storage config: %w", err)
	}
	return nil
}

// List returns all storage configurations, ordered by priority ascending with created_at as a tiebreaker.
func (r *StorageConfigRepo) List(ctx context.Context) ([]model.StorageConfig, error) {
	if err := r.normalizeLegacyConfigs(ctx); err != nil {
		return nil, err
	}
	var configs []model.StorageConfig
	if err := r.db.WithContext(ctx).
		Order("priority ASC, created_at ASC").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("list storage configs: %w", err)
	}
	return configs, nil
}

// ListActive returns all active storage configurations, ordered by priority ascending with created_at as a tiebreaker.
func (r *StorageConfigRepo) ListActive(ctx context.Context) ([]model.StorageConfig, error) {
	if err := r.normalizeLegacyConfigs(ctx); err != nil {
		return nil, err
	}
	var configs []model.StorageConfig
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("priority ASC, created_at ASC").
		Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("list active storage configs: %w", err)
	}
	return configs, nil
}

// Reorder rewrites priorities to match the given ID order: the first entry gets 100 and each subsequent
// entry gets +100. The wide step leaves room for future ad-hoc priority adjustments without a full re-index.
func (r *StorageConfigRepo) Reorder(ctx context.Context, orderedIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range orderedIDs {
			priority := (i + 1) * 100
			if err := tx.Model(&model.StorageConfig{}).
				Where("id = ?", id).
				Update("priority", priority).Error; err != nil {
				return fmt.Errorf("reorder storage %q: %w", id, err)
			}
		}
		return nil
	})
}

// BuildRawConfigs converts every storage configuration to storage.RawConfig values for Manager.Reload.
// Inactive entries are included; Reload filters them internally so toggling active state only requires
// another Reload call.
func (r *StorageConfigRepo) BuildRawConfigs(ctx context.Context) ([]storage.RawConfig, error) {
	configs, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]storage.RawConfig, 0, len(configs))
	for _, c := range configs {
		out = append(out, storage.RawConfig{
			ID:                 c.ID,
			Name:               c.Name,
			Driver:             c.Driver,
			Provider:           providerValue(c.Provider),
			ConfigJSON:         c.Config,
			Priority:           c.Priority,
			CapacityLimitBytes: c.CapacityLimitBytes,
			IsDefault:          c.IsDefault,
			IsActive:           c.IsActive,
		})
	}
	return out, nil
}

// normalizeLegacyConfigs rewrites legacy S3 vendor driver values into the new
// driver=s3 + provider=* shape. It is safe to call repeatedly.
func (r *StorageConfigRepo) normalizeLegacyConfigs(ctx context.Context) error {
	var configs []model.StorageConfig
	if err := r.db.WithContext(ctx).Find(&configs).Error; err != nil {
		return fmt.Errorf("list storage configs for normalization: %w", err)
	}

	for _, cfg := range configs {
		driver, provider := storage.CanonicalDriverAndProvider(cfg.Driver, cfg.Provider, cfg.Config)
		if cfg.Driver == driver && providerValue(cfg.Provider) == provider {
			continue
		}

		updates := map[string]any{
			"driver": driver,
		}
		if provider == "" {
			updates["provider"] = nil
		} else {
			updates["provider"] = provider
		}

		if err := r.db.WithContext(ctx).
			Model(&model.StorageConfig{}).
			Where("id = ?", cfg.ID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("normalize storage config %q: %w", cfg.ID, err)
		}
	}

	return nil
}

func providerValue(provider *string) string {
	if provider == nil {
		return ""
	}
	return *provider
}
