package main

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/kite-plus/kite/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// initDatabase opens a database connection using the driver named in cfg and
// applies driver-specific tuning (WAL mode and foreign keys for SQLite, a
// standard connection pool for MySQL and Postgres). Supported drivers are
// sqlite, mysql and postgres.
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
		// WAL mode significantly improves read concurrency for SQLite.
		if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return nil, fmt.Errorf("set WAL mode: %w", err)
		}
		if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
			return nil, fmt.Errorf("enable foreign keys: %w", err)
		}
		// SQLite has a single-writer model; capping the pool at one
		// connection avoids "database is locked" errors under contention.
		sqlDB.SetMaxOpenConns(1)

	case "mysql", "postgres":
		// Typical defaults for networked databases. 50 open connections
		// leave headroom under the usual server-side connection limit.
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return db, nil
}

// newDialector returns the gorm dialector matching cfg.Driver. An empty DSN
// or an unrecognised driver produces an error so startup fails fast rather
// than hanging on a malformed connection attempt.
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
