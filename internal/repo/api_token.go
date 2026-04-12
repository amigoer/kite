package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/amigoer/kite/internal/model"
	"gorm.io/gorm"
)

// APITokenRepo API Token 数据访问层。
type APITokenRepo struct {
	db *gorm.DB
}

func NewAPITokenRepo(db *gorm.DB) *APITokenRepo {
	return &APITokenRepo{db: db}
}

// Create 创建 API Token 记录。
func (r *APITokenRepo) Create(ctx context.Context, token *model.APIToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("create api token: %w", err)
	}
	return nil
}

// GetByTokenHash 通过 token hash 查询（用于认证验证）。
func (r *APITokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*model.APIToken, error) {
	var token model.APIToken
	if err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&token).Error; err != nil {
		return nil, fmt.Errorf("get api token by hash: %w", err)
	}
	return &token, nil
}

// ListByUser 查询用户的所有 API Token。
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

// Delete 删除 API Token。
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

// UpdateLastUsed 更新 Token 最后使用时间。
func (r *APITokenRepo) UpdateLastUsed(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).
		Model(&model.APIToken{}).
		Where("id = ?", id).
		Update("last_used", time.Now()).Error; err != nil {
		return fmt.Errorf("update api token last used: %w", err)
	}
	return nil
}
