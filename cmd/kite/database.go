package main

import (
	"fmt"

	"github.com/amigoer/kite/internal/config"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// initDatabase 根据配置初始化数据库连接。
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite 性能优化
	if cfg.Driver == "sqlite" {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("get underlying db: %w", err)
		}
		// WAL 模式提升并发读性能
		if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return nil, fmt.Errorf("set WAL mode: %w", err)
		}
		// 启用外键约束
		if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
			return nil, fmt.Errorf("enable foreign keys: %w", err)
		}
		sqlDB.SetMaxOpenConns(1) // SQLite 单写
	}

	return db, nil
}
