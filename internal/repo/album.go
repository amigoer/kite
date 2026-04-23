package repo

import (
	"context"
	"fmt"

	"github.com/kite-plus/kite/internal/model"
	"gorm.io/gorm"
)

// AlbumRepo is the data access layer for albums (folders).
type AlbumRepo struct {
	db *gorm.DB
}

func NewAlbumRepo(db *gorm.DB) *AlbumRepo {
	return &AlbumRepo{db: db}
}

// Create inserts a new album.
func (r *AlbumRepo) Create(ctx context.Context, album *model.Album) error {
	if err := r.db.WithContext(ctx).Create(album).Error; err != nil {
		return fmt.Errorf("create album: %w", err)
	}
	return nil
}

// GetByID fetches an album by its ID.
func (r *AlbumRepo) GetByID(ctx context.Context, id string) (*model.Album, error) {
	var album model.Album
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&album).Error; err != nil {
		return nil, fmt.Errorf("get album by id: %w", err)
	}
	return &album, nil
}

// Update persists changes to an album.
func (r *AlbumRepo) Update(ctx context.Context, album *model.Album) error {
	if err := r.db.WithContext(ctx).Save(album).Error; err != nil {
		return fmt.Errorf("update album: %w", err)
	}
	return nil
}

// Delete removes a folder recursively (including subfolders); contained files are retained with their
// folder_id/album_id cleared.
func (r *AlbumRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ids, err := r.collectDescendantIDs(ctx, tx, id)
		if err != nil {
			return err
		}

		// Clear album_id on every file whose folder is part of the subtree.
		if err := tx.Model(&model.File{}).
			Where("album_id IN ?", ids).
			Update("album_id", nil).Error; err != nil {
			return fmt.Errorf("clear folder files: %w", err)
		}
		if err := tx.Where("id IN ?", ids).Delete(&model.Album{}).Error; err != nil {
			return fmt.Errorf("delete folders: %w", err)
		}
		return nil
	})
}

// ListByUser returns the folders owned by userID under the given parent.
func (r *AlbumRepo) ListByUser(ctx context.Context, userID string, parentID *string, page, pageSize int) ([]model.Album, int64, error) {
	var albums []model.Album
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Album{}).Where("user_id = ?", userID)
	if parentID == nil {
		db = db.Where("parent_id IS NULL")
	} else {
		db = db.Where("parent_id = ?", *parentID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count albums: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&albums).Error; err != nil {
		return nil, 0, fmt.Errorf("list albums: %w", err)
	}

	return albums, total, nil
}

func (r *AlbumRepo) CountChildren(ctx context.Context, id string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Album{}).
		Where("parent_id = ?", id).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count child folders: %w", err)
	}
	return count, nil
}

func (r *AlbumRepo) ListAncestors(ctx context.Context, userID string, id string) ([]model.Album, error) {
	ancestors := make([]model.Album, 0)
	currentID := id
	for currentID != "" {
		folder, err := r.GetByID(ctx, currentID)
		if err != nil {
			return nil, err
		}
		if folder.UserID != userID {
			return nil, gorm.ErrRecordNotFound
		}
		ancestors = append([]model.Album{*folder}, ancestors...)
		if folder.ParentID == nil {
			break
		}
		currentID = *folder.ParentID
	}
	return ancestors, nil
}

func (r *AlbumRepo) collectDescendantIDs(ctx context.Context, tx *gorm.DB, rootID string) ([]string, error) {
	ids := []string{rootID}
	frontier := []string{rootID}

	for len(frontier) > 0 {
		var childIDs []string
		if err := tx.WithContext(ctx).
			Model(&model.Album{}).
			Where("parent_id IN ?", frontier).
			Pluck("id", &childIDs).Error; err != nil {
			return nil, fmt.Errorf("list child folders: %w", err)
		}
		if len(childIDs) == 0 {
			break
		}
		ids = append(ids, childIDs...)
		frontier = childIDs
	}

	return ids, nil
}

// ListPublic returns all public albums.
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
