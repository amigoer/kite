package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/kite-plus/kite/internal/model"
	"gorm.io/gorm"
)

// FileReplicaRepo is the data access layer for file replicas (mirror mode).
type FileReplicaRepo struct {
	db *gorm.DB
}

func NewFileReplicaRepo(db *gorm.DB) *FileReplicaRepo {
	return &FileReplicaRepo{db: db}
}

// Create inserts a replica record, typically pre-written as pending during the upload flow.
func (r *FileReplicaRepo) Create(ctx context.Context, replica *model.FileReplica) error {
	if err := r.db.WithContext(ctx).Create(replica).Error; err != nil {
		return fmt.Errorf("create file replica: %w", err)
	}
	return nil
}

// ListByFile returns all replica records for the given file.
func (r *FileReplicaRepo) ListByFile(ctx context.Context, fileID string) ([]model.FileReplica, error) {
	var replicas []model.FileReplica
	if err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Find(&replicas).Error; err != nil {
		return nil, fmt.Errorf("list file replicas: %w", err)
	}
	return replicas, nil
}

// UpdateStatus records the outcome of a replica write.
func (r *FileReplicaRepo) UpdateStatus(ctx context.Context, id, status, errMsg string) error {
	if err := r.db.WithContext(ctx).
		Model(&model.FileReplica{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":    status,
			"error_msg": errMsg,
		}).Error; err != nil {
		return fmt.Errorf("update file replica status: %w", err)
	}
	return nil
}

// DeleteByFile removes all replica records for the given file.
func (r *FileReplicaRepo) DeleteByFile(ctx context.Context, fileID string) error {
	if err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Delete(&model.FileReplica{}).Error; err != nil {
		return fmt.Errorf("delete file replicas: %w", err)
	}
	return nil
}

// MarkStalePending flips replica rows that have been stuck at "pending" past
// the given staleness threshold to "failed". Background replication happens
// in goroutines that do not survive process restarts, so if the server is
// killed between the successful Put and the status update, the row would
// otherwise be stuck forever. Callers run this on startup to surface
// orphans to operators. Returns the number of rows flipped.
func (r *FileReplicaRepo) MarkStalePending(ctx context.Context, olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		return 0, fmt.Errorf("mark stale replicas: olderThan must be positive, got %s", olderThan)
	}
	cutoff := time.Now().Add(-olderThan)
	res := r.db.WithContext(ctx).
		Model(&model.FileReplica{}).
		Where("status = ? AND updated_at < ?", model.ReplicaStatusPending, cutoff).
		Updates(map[string]any{
			"status":    model.ReplicaStatusFailed,
			"error_msg": "startup reconciliation: replica was stuck at pending past staleness threshold",
		})
	if res.Error != nil {
		return 0, fmt.Errorf("mark stale replicas: %w", res.Error)
	}
	return res.RowsAffected, nil
}
