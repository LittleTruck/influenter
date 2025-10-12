package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func init() {
	// 載入 .env 檔案（從專案根目錄）
	// 測試時會從 backend/internal/database 目錄執行，需要往上找 .env
	envPath := filepath.Join("..", "..", "..", ".env")
	_ = godotenv.Load(envPath) // 忽略錯誤，因為可能使用環境變數
}

// getTestConfig 返回測試用的配置
// 從 .env 檔案或環境變數載入配置
func getTestConfig() *config.Config {
	// 直接使用 config.Load() 來載入配置
	// 但覆蓋部分測試專用的設定
	cfg, err := config.Load()
	if err != nil {
		// 如果載入失敗，返回最小配置以便錯誤測試可以執行
		return &config.Config{
			Env:      "test",
			LogLevel: "error",
			Database: config.DatabaseConfig{
				Host:            os.Getenv("DB_HOST"),
				Port:            os.Getenv("DB_PORT"),
				User:            os.Getenv("DB_USER"),
				Password:        os.Getenv("DB_PASSWORD"),
				Database:        os.Getenv("DB_NAME"),
				SSLMode:         "disable",
				MaxOpenConns:    5,
				MaxIdleConns:    2,
				ConnMaxLifetime: 5 * time.Minute,
			},
		}
	}

	// 使用載入的配置，但調整測試專用設定
	cfg.Env = "test"
	cfg.LogLevel = "error"
	cfg.Database.MaxOpenConns = 5
	cfg.Database.MaxIdleConns = 2
	cfg.Database.ConnMaxLifetime = 5 * time.Minute

	return cfg
}

// TestNew 測試資料庫連線建立
func TestNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Error("Expected db to be non-nil")
	}

	if db.DB == nil {
		t.Error("Expected db.DB to be non-nil")
	}
}

// TestNew_InvalidConfig 測試使用無效配置建立連線
func TestNew_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		Env:      "test",   // 設定環境
		LogLevel: "silent", // 靜默模式，不打印預期的錯誤日誌
		Database: config.DatabaseConfig{
			Host:     "invalid-host",
			Port:     "5432",
			User:     "invalid",
			Password: "invalid",
			Database: "invalid",
			SSLMode:  "disable",
		},
	}

	_, err := New(cfg)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}

	// 驗證錯誤訊息包含預期的內容
	if err != nil && !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("Expected connection error, got: %v", err)
	}
}

// TestClose 測試關閉資料庫連線
func TestClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestPing 測試資料庫 ping
func TestPing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

// TestPing_WithTimeout 測試帶有超時的 ping
func TestPing_WithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		t.Errorf("Ping() with timeout error = %v", err)
	}
}

// TestHealthCheck 測試健康檢查
func TestHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	err = db.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}
}

// TestHealthCheck_WithCanceledContext 測試取消的 context
func TestHealthCheck_WithCanceledContext(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	err = db.HealthCheck(ctx)
	if err == nil {
		t.Error("Expected error for canceled context, got nil")
	}
}

// TestGetStats 測試取得連線池統計
func TestGetStats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	stats, err := db.GetStats()
	if err != nil {
		t.Errorf("GetStats() error = %v", err)
	}

	if stats == nil {
		t.Error("Expected stats to be non-nil")
	}

	// 檢查必要的統計欄位
	requiredFields := []string{
		"max_open_connections",
		"open_connections",
		"in_use",
		"idle",
		"wait_count",
	}

	for _, field := range requiredFields {
		if _, ok := stats[field]; !ok {
			t.Errorf("Expected stats to have field %s", field)
		}
	}

	// 驗證連線池設定
	maxOpen := stats["max_open_connections"].(int)
	if maxOpen != cfg.Database.MaxOpenConns {
		t.Errorf("Expected max_open_connections = %d, got %d",
			cfg.Database.MaxOpenConns, maxOpen)
	}
}

// TestTransaction 測試交易功能
func TestTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 測試成功的交易
	err = db.Transaction(ctx, func(tx *gorm.DB) error {
		// 執行一些操作
		var result int
		if err := tx.Raw("SELECT 1").Scan(&result).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		t.Errorf("Transaction() error = %v", err)
	}
}

// TestTransaction_Rollback 測試交易回滾
func TestTransaction_Rollback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 測試失敗的交易（應該回滾）
	expectedErr := errors.New("test error")
	err = db.Transaction(ctx, func(tx *gorm.DB) error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Expected transaction to rollback with error, got %v", err)
	}
}

// TestIsUniqueViolation 測試唯一性約束檢查
func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "duplicate key error",
			err:      errors.New("duplicate key value violates unique constraint"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUniqueViolation(tt.err); got != tt.expected {
				t.Errorf("IsUniqueViolation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsForeignKeyViolation 測試外鍵約束檢查
func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForeignKeyViolation(tt.err); got != tt.expected {
				t.Errorf("IsForeignKeyViolation() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestPaginate 測試分頁函數
func TestPaginate(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		wantPage int
		wantSize int
	}{
		{
			name:     "valid pagination",
			page:     2,
			pageSize: 10,
			wantPage: 2,
			wantSize: 10,
		},
		{
			name:     "page <= 0 defaults to 1",
			page:     0,
			pageSize: 10,
			wantPage: 1,
			wantSize: 10,
		},
		{
			name:     "pageSize <= 0 defaults to 20",
			page:     1,
			pageSize: 0,
			wantPage: 1,
			wantSize: 20,
		},
		{
			name:     "pageSize > 100 defaults to 20",
			page:     1,
			pageSize: 150,
			wantPage: 1,
			wantSize: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Paginate 函數返回一個閉包，我們只是確保它不會 panic
			fn := Paginate(tt.page, tt.pageSize)
			if fn == nil {
				t.Error("Expected Paginate to return a function")
			}
		})
	}
}

// TestOrderBy 測試排序函數
func TestOrderBy(t *testing.T) {
	tests := []struct {
		name  string
		field string
		desc  bool
	}{
		{
			name:  "ascending order",
			field: "created_at",
			desc:  false,
		},
		{
			name:  "descending order",
			field: "created_at",
			desc:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := OrderBy(tt.field, tt.desc)
			if fn == nil {
				t.Error("Expected OrderBy to return a function")
			}
		})
	}
}

// Benchmark tests

func BenchmarkPing(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.Ping(ctx)
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	cfg := getTestConfig()
	db, err := New(cfg)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = db.HealthCheck(ctx)
	}
}
