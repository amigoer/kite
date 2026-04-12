package repo

import (
	"context"
	"fmt"

	"github.com/amigoer/kite/internal/model"
	"gorm.io/gorm"
)

// AlbumRepo 相册数据访问层。
type AlbumRepo struct {
	db *gorm.DB
}

func NewAlbumRepo(db *gorm.DB) *AlbumRepo {
	return &AlbumRepo{db: db}
}

// Create 创建相册。
func (r *AlbumRepo) Create(ctx context.Context, album *model.Album) error {
	if err := r.db.WithContext(ctx).Create(album).Error; err != nil {
		return fmt.Errorf("create album: %w", err)
	}
	return nil
}

// GetByID 通过 ID 查询相册。
func (r *AlbumRepo) GetByID(ctx context.Context, id string) (*model.Album, error) {
	var album model.Album
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&album).Error; err != nil {
		return nil, fmt.Errorf("get album by id: %w", err)
	}
	return &album, nil
}

// Update 更新相册。
func (r *AlbumRepo) Update(ctx context.Context, album *model.Album) error {
	if err := r.db.WithContext(ctx).Save(album).Error; err != nil {
		return fmt.Errorf("update album: %w", err)
	}
	return nil
}

// Delete 删除相册（硬删除，相册中的文件不会被删除，album_id 置空）。
func (r *AlbumRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 将相册下所有文件的 album_id 置空
		if err := tx.Model(&model.File{}).
			Where("album_id = ?", id).
			Update("album_id", nil).Error; err != nil {
			return fmt.Errorf("clear album files: %w", err)
		}
		// 删除相册
		if err := tx.Where("id = ?", id).Delete(&model.Album{}).Error; err != nil {
			return fmt.Errorf("delete album: %w", err)
		}
		return nil
	})
}

// ListByUser 查询用户的所有相册。
func (r *AlbumRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]model.Album, int64, error) {
	var albums []model.Album
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Album{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count albums: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&albums).Error; err != nil {
		return nil, 0, fmt.Errorf("list albums: %w", err)
	}

	return albums, total, nil
}

// ListPublic 查询所有公开相册。
func (r *AlbumRepo) ListPublic(ctx context.Context, page, pageSize int) ([]model.Album, int64, error) {
	var albums []model.Album
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Album{}).Where("is_public = ?", true)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count public albums: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&albums).Error; err != nil {
		return nil, 0, fmt.Errorf("list public albums: %w", err)
	}

	return albums, total, nil
}
