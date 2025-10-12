package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/designcomb/influenter-backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 封裝資料庫連線
type DB struct {
	*gorm.DB
}

// New 建立新的資料庫連線
func New(cfg *config.Config) (*DB, error) {
	// 建立 GORM 配置
	gormConfig := &gorm.Config{
		Logger: getLogger(cfg),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// 禁用外鍵約束（由應用層管理）
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 連接資料庫
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 取得底層的 *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 設定連線池
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// 驗證連線
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✅ Database connected successfully (host=%s, db=%s)",
		cfg.Database.Host, cfg.Database.Database)

	return &DB{DB: db}, nil
}

// Close 關閉資料庫連線
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("🔌 Database connection closed")
	return nil
}

// Ping 檢查資料庫連線是否正常
func (db *DB) Ping(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// HealthCheck 執行完整的健康檢查
func (db *DB) HealthCheck(ctx context.Context) error {
	// 1. 基本連線檢查
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// 2. 執行簡單查詢測試
	var result int
	if err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: expected 1, got %d", result)
	}

	return nil
}

// GetStats 取得資料庫連線池統計資訊
func (db *DB) GetStats() (map[string]interface{}, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// Transaction 執行資料庫交易
func (db *DB) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}

// IsUniqueViolation 檢查錯誤是否為唯一性約束違反
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// PostgreSQL unique violation error code is 23505
	return errStr == "23505" ||
		errStr == "duplicate key value violates unique constraint" ||
		err == gorm.ErrDuplicatedKey
}

// IsForeignKeyViolation 檢查錯誤是否為外鍵約束違反
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// PostgreSQL foreign key violation error code is 23503
	return errStr == "23503" ||
		errStr == "foreign key constraint"
}

// getLogger 根據配置返回適當的 logger
func getLogger(cfg *config.Config) logger.Interface {
	logLevel := logger.Info

	switch cfg.LogLevel {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Warn
	case "warn":
		logLevel = logger.Warn
	case "error":
		logLevel = logger.Error
	case "silent":
		logLevel = logger.Silent
	default:
		logLevel = logger.Info
	}

	// 生產環境使用較少的日誌
	if cfg.IsProduction() {
		logLevel = logger.Error
	}

	return logger.Default.LogMode(logLevel)
}

// Paginate 分頁輔助函數
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// OrderBy 排序輔助函數
func OrderBy(field string, desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		order := field
		if desc {
			order += " DESC"
		}
		return db.Order(order)
	}
}

// AutoMigrate 執行資料庫遷移
func (db *DB) AutoMigrate(models ...interface{}) error {
	if err := db.DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}
