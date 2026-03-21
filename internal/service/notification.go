package service

import (
	"fmt"

	"github.com/amigoer/kite-blog/internal/model"
	"github.com/amigoer/kite-blog/internal/repo"
	"github.com/google/uuid"
)

// NotificationService 通知业务服务
type NotificationService struct {
	notificationRepo *repo.NotificationRepository
}

func NewNotificationService(notificationRepo *repo.NotificationRepository) *NotificationService {
	return &NotificationService{notificationRepo: notificationRepo}
}

// NotificationListResult 通知列表结果
type NotificationListResult struct {
	Items      []model.Notification `json:"items"`
	Pagination Pagination           `json:"pagination"`
}

// List 查询通知列表
func (s *NotificationService) List(page, pageSize int, unreadOnly bool) (*NotificationListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	notifications, total, err := s.notificationRepo.List(page, pageSize, unreadOnly)
	if err != nil {
		return nil, err
	}

	return &NotificationListResult{
		Items: notifications,
		Pagination: Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

// CountUnread 统计未读通知数量
func (s *NotificationService) CountUnread() (int64, error) {
	return s.notificationRepo.CountUnread()
}

// MarkRead 标记单条通知为已读
func (s *NotificationService) MarkRead(idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("invalid notification id")
	}
	return s.notificationRepo.MarkRead(id)
}

// MarkAllRead 标记所有通知为已读
func (s *NotificationService) MarkAllRead() error {
	return s.notificationRepo.MarkAllRead()
}

// CreateFromComment 根据新评论创建通知
func (s *NotificationService) CreateFromComment(comment *model.Comment, postTitle string) error {
	title := fmt.Sprintf("%s 发表了新评论", comment.Author)
	content := comment.Content
	if len(content) > 200 {
		content = content[:200] + "…"
	}

	notification := &model.Notification{
		Type:    model.NotificationTypeComment,
		Title:   title,
		Content: content,
		Link:    "/admin/comments",
		IsRead:  false,
	}
	return s.notificationRepo.Create(notification)
}
