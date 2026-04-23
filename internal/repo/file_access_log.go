package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/kite-plus/kite/internal/model"
	"gorm.io/gorm"
)

// FileAccessLogRepo is the data access layer for file access logs.
type FileAccessLogRepo struct {
	db *gorm.DB
}

func NewFileAccessLogRepo(db *gorm.DB) *FileAccessLogRepo {
	return &FileAccessLogRepo{db: db}
}

// Create records a single file access.
func (r *FileAccessLogRepo) Create(ctx context.Context, log *model.FileAccessLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("create file access log: %w", err)
	}
	return nil
}

// DailyAccessStat captures per-day access count and bandwidth.
type DailyAccessStat struct {
	Day         string `json:"day"`          // YYYY-MM-DD
	AccessCount int64  `json:"access_count"` // number of accesses that day
	BytesServed int64  `json:"bytes_served"` // bytes served that day
}

// HourlyWeekdayAccessStat aggregates access counts by weekday and hour.
// Weekday: 0=Sunday ... 6=Saturday (SQLite strftime('%w') convention).
// Hour: 0..23.
type HourlyWeekdayAccessStat struct {
	Weekday int   `json:"weekday"`
	Hour    int   `json:"hour"`
	Count   int64 `json:"count"`
}

// GetDailyAccessStats returns per-day access counts and bandwidth in [start, end).
// start/end are UTC day boundaries (start inclusive, end exclusive).
// An empty userID returns site-wide aggregates (admin view); otherwise results are scoped to that user's files.
func (r *FileAccessLogRepo) GetDailyAccessStats(ctx context.Context, userID string, start, end time.Time) ([]DailyAccessStat, error) {
	var rows []DailyAccessStat
	db := r.db.WithContext(ctx).
		Model(&model.FileAccessLog{}).
		Select("strftime('%Y-%m-%d', accessed_at) as day, COUNT(*) as access_count, COALESCE(SUM(bytes_served), 0) as bytes_served").
		Where("accessed_at >= ? AND accessed_at < ?", start, end)
	if userID != "" {
		db = db.Where("user_id = ?", userID)
	}
	if err := db.Group("day").Order("day ASC").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get daily access stats: %w", err)
	}
	return rows, nil
}

// GetHourlyAccessHeatmapStats returns access counts grouped by weekday and hour within [start, end).
// start/end are time boundaries (start inclusive, end exclusive).
// An empty userID returns site-wide aggregates (admin view); otherwise results are scoped to accesses of that user's files.
func (r *FileAccessLogRepo) GetHourlyAccessHeatmapStats(ctx context.Context, userID string, start, end time.Time) ([]HourlyWeekdayAccessStat, error) {
	var rows []HourlyWeekdayAccessStat
	db := r.db.WithContext(ctx).
		Model(&model.FileAccessLog{}).
		Select("CAST(strftime('%w', accessed_at) AS INTEGER) as weekday, CAST(strftime('%H', accessed_at) AS INTEGER) as hour, COUNT(*) as count").
		Where("accessed_at >= ? AND accessed_at < ?", start, end)
	if userID != "" {
		db = db.Where("user_id = ?", userID)
	}
	if err := db.Group("weekday, hour").Order("weekday ASC, hour ASC").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get hourly access heatmap stats: %w", err)
	}
	return rows, nil
}
