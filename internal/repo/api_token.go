package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/kite-plus/kite/internal/model"
	"gorm.io/gorm"
)

// APITokenRepo is the data access layer for API tokens.
type APITokenRepo struct {
	db *gorm.DB
}

func NewAPITokenRepo(db *gorm.DB) *APITokenRepo {
	return &APITokenRepo{db: db}
}

// Create inserts a new API-token record.
func (r *APITokenRepo) Create(ctx context.Context, token *model.APIToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("create api token: %w", err)
	}
	return nil
}

// GetByTokenHash looks up a token by its hash; used to validate inbound API requests.
func (r *APITokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*model.APIToken, error) {
	var token model.APIToken
	if err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&token).Error; err != nil {
		return nil, fmt.Errorf("get api token by hash: %w", err)
	}
	return &token, nil
}

// ListByUser returns every API token owned by the user.
func (r *APITokenRepo) ListByUser(ctx context.Context, userID string) ([]model.APIToken, error) {
	var tokens []model.APIToken
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("list api tokens: %w", err)
	}
	return tokens, nil
}

// Delete removes an API token scoped to the owning user.
func (r *APITokenRepo) Delete(ctx context.Context, id, userID string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.APIToken{})
	if result.Error != nil {
		return fmt.Errorf("delete api token: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateLastUsed refreshes the token's last-used timestamp.
func (r *APITokenRepo) UpdateLastUsed(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).
		Model(&model.APIToken{}).
		Where("id = ?", id).
		Update("last_used", time.Now()).Error; err != nil {
		return fmt.Errorf("update api token last used: %w", err)
	}
	return nil
}
