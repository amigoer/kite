package repo

import (
	"context"
	"fmt"

	"github.com/amigoer/kite/internal/model"
	"gorm.io/gorm"
)

// UserRepo 用户数据访问层。
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create 创建新用户。
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetByID 通过 ID 查询用户。
func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

// GetByUsername 通过用户名查询用户。
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &user, nil
}

// GetByEmail 通过邮箱查询用户。
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

// Update 更新用户信息。
func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// UpdateStorageUsed 增减用户已用存储量。delta 可为负值。
func (r *UserRepo) UpdateStorageUsed(ctx context.Context, userID string, delta int64) error {
	if err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("storage_used", gorm.Expr("storage_used + ?", delta)).Error; err != nil {
		return fmt.Errorf("update storage used: %w", err)
	}
	return nil
}

// Delete 删除用户（软删除：设置 is_active = false）。
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// List 分页查询用户列表。
func (r *UserRepo) List(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := r.db.WithContext(ctx).Model(&model.User{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	return users, total, nil
}

// ExistsByUsernameOrEmail 检查用户名或邮箱是否已存在。
func (r *UserRepo) ExistsByUsernameOrEmail(ctx context.Context, username, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return count > 0, nil
}

// ExistsByUsernameOrEmailExcept 检查除指定用户外是否有其他用户名或邮箱冲突。
// 用于用户更新自身资料时的冲突检测。
func (r *UserRepo) ExistsByUsernameOrEmailExcept(ctx context.Context, username, email, excludeID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("(username = ? OR email = ?) AND id <> ?", username, email, excludeID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check user exists except: %w", err)
	}
	return count > 0, nil
}

// Count 返回用户总数（用于判断是否需要初始化安装）。
func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}
