package repo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amigoer/kite-blog/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrFriendLinkNotFound = errors.New("friend link not found")

type FriendLinkListParams struct {
	Page       int
	PageSize   int
	Keyword    string
	IsActive   *bool
	PublicOnly bool
}

type FriendLinkRepository struct {
	db *gorm.DB
}

func NewFriendLinkRepository(db *gorm.DB) *FriendLinkRepository {
	return &FriendLinkRepository{db: db}
}

func (r *FriendLinkRepository) List(params FriendLinkListParams) ([]model.FriendLink, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("friend link repository is unavailable")
	}

	query := r.db.Model(&model.FriendLink{})
	if params.PublicOnly {
		query = query.Where("is_active = ?", true)
	}
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}
	if params.Keyword != "" {
		keyword := "%" + strings.TrimSpace(params.Keyword) + "%"
		query = query.Where("name LIKE ? OR description LIKE ? OR url LIKE ?", keyword, keyword, keyword)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count friend links: %w", err)
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	var links []model.FriendLink
	if err := query.Order("sort ASC, created_at DESC").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&links).Error; err != nil {
		return nil, 0, fmt.Errorf("list friend links: %w", err)
	}

	return links, total, nil
}

func (r *FriendLinkRepository) GetByID(id uuid.UUID) (*model.FriendLink, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("friend link repository is unavailable")
	}

	var link model.FriendLink
	if err := r.db.First(&link, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFriendLinkNotFound
		}
		return nil, fmt.Errorf("get friend link by id: %w", err)
	}

	return &link, nil
}

func (r *FriendLinkRepository) GetPublicByID(id uuid.UUID) (*model.FriendLink, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("friend link repository is unavailable")
	}

	var link model.FriendLink
	if err := r.db.First(&link, "id = ? AND is_active = ?", id, true).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFriendLinkNotFound
		}
		return nil, fmt.Errorf("get public friend link by id: %w", err)
	}

	return &link, nil
}

func (r *FriendLinkRepository) GetByURL(url string) (*model.FriendLink, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("friend link repository is unavailable")
	}

	var link model.FriendLink
	if err := r.db.First(&link, "url = ?", url).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFriendLinkNotFound
		}
		return nil, fmt.Errorf("get friend link by url: %w", err)
	}

	return &link, nil
}

func (r *FriendLinkRepository) Create(link *model.FriendLink) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("friend link repository is unavailable")
	}
	if err := r.db.Create(link).Error; err != nil {
		return fmt.Errorf("create friend link: %w", err)
	}
	return nil
}

func (r *FriendLinkRepository) Update(link *model.FriendLink) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("friend link repository is unavailable")
	}
	result := r.db.Model(link).Select("name", "url", "description", "logo", "sort", "is_active").Updates(link)
	if result.Error != nil {
		return fmt.Errorf("update friend link: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFriendLinkNotFound
	}
	return nil
}

func (r *FriendLinkRepository) Delete(id uuid.UUID) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("friend link repository is unavailable")
	}
	result := r.db.Delete(&model.FriendLink{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("delete friend link: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFriendLinkNotFound
	}
	return nil
}
