package repo

import (
	"errors"
	"fmt"

	"github.com/amigoer/kite-blog/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotificationNotFound = errors.New("notification not found")

// NotificationRepository 通知数据仓库
type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// List 查询通知列表（按时间倒序）
func (r *NotificationRepository) List(page, pageSize int, unreadOnly bool) ([]model.Notification, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("notification repository is unavailable")
	}

	query := r.db.Model(&model.Notification{})
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var notifications []model.Notification
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}

	return notifications, total, nil
}

// CountUnread 统计未读通知数量
func (r *NotificationRepository) CountUnread() (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("notification repository is unavailable")
	}

	var count int64
	if err := r.db.Model(&model.Notification{}).Where("is_read = ?", false).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}

// MarkRead 标记单条通知为已读
func (r *NotificationRepository) MarkRead(id uuid.UUID) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("notification repository is unavailable")
	}

	result := r.db.Model(&model.Notification{}).Where("id = ?", id).Update("is_read", true)
	if result.Error != nil {
		return fmt.Errorf("mark notification read: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

// MarkAllRead 标记所有通知为已读
func (r *NotificationRepository) MarkAllRead() error {
	if r == nil || r.db == nil {
		return fmt.Errorf("notification repository is unavailable")
	}

	if err := r.db.Model(&model.Notification{}).Where("is_read = ?", false).Update("is_read", true).Error; err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

// Create 创建通知
func (r *NotificationRepository) Create(notification *model.Notification) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("notification repository is unavailable")
	}

	if err := r.db.Create(notification).Error; err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}
