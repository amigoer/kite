package main

import (
	"fmt"
	"time"

	"github.com/amigoer/kite/internal/config"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// initDatabase 根据配置初始化数据库连接。
// 支持的驱动：sqlite / mysql / postgres。
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dialector, err := newDialector(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying db: %w", err)
	}

	switch cfg.Driver {
	case "sqlite":
		// WAL 模式提升并发读性能
		if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return nil, fmt.Errorf("set WAL mode: %w", err)
		}
		// 启用外键约束
		if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
			return nil, fmt.Errorf("enable foreign keys: %w", err)
		}
		// SQLite 单写者模型，保留 1 个连接避免 database is locked
		sqlDB.SetMaxOpenConns(1)

	case "mysql", "postgres":
		// 网络数据库使用标准连接池，留出余量避免压满服务端连接数
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return db, nil
}

// newDialector 构造对应驱动的 gorm.Dialector。
func newDialector(cfg config.DatabaseConfig) (gorm.Dialector, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}
	switch cfg.Driver {
	case "sqlite":
		return sqlite.Open(cfg.DSN), nil
	case "mysql":
		return mysql.Open(cfg.DSN), nil
	case "postgres":
		return postgres.Open(cfg.DSN), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %q (expected sqlite, mysql or postgres)", cfg.Driver)
	}
}
