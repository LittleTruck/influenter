package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 包含應用程式所有配置
type Config struct {
	// 環境設定
	Env      string
	Port     string
	LogLevel string
	GinMode  string

	// 資料庫設定
	Database DatabaseConfig

	// Redis 設定
	Redis RedisConfig

	// Google OAuth 設定
	Google GoogleOAuthConfig

	// JWT 設定
	JWT JWTConfig

	// 加密設定
	EncryptionKey string

	// OpenAI 設定
	OpenAI OpenAIConfig

	// 前端 URL
	FrontendURL string

	// CORS 設定
	CORS CORSConfig

	// Asynq 設定
	Asynq AsynqConfig

	// 郵件同步設定
	EmailSync EmailSyncConfig

	// AI 分析設定
	AI AIConfig

	// 通知設定
	Notification NotificationConfig

	// 安全設定
	Security SecurityConfig
}

// DatabaseConfig 資料庫配置
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// GoogleOAuthConfig Google OAuth 配置
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins []string
}

// AsynqConfig Asynq 背景任務配置
type AsynqConfig struct {
	Concurrency  int
	SyncSchedule string
	MaxRetry     int
}

// EmailSyncConfig 郵件同步配置
type EmailSyncConfig struct {
	InitialSyncDays  int
	MaxEmailsPerSync int
	SyncCooldownSec  int
}

// AIConfig AI 分析配置
type AIConfig struct {
	AutoAnalyze             bool
	ConfidenceThreshold     float64
	AutoCreateCaseThreshold float64
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	TaskReminderDays []int
	SMTPHost         string
	SMTPPort         int
	SMTPUser         string
	SMTPPassword     string
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RateLimitPerMinute    int
	SessionCookieName     string
	SessionCookieSecure   bool
	SessionCookieHTTPOnly bool
	SessionMaxAge         int
}

// Load 從環境變數載入配置
func Load() (*Config, error) {
	cfg := &Config{
		// 基本設定
		Env:      getEnv("ENV", "development"),
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		GinMode:  getEnv("GIN_MODE", "debug"),

		// 資料庫設定
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", ""),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_NAME", ""),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
		},

		// Redis 設定
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},

		// Google OAuth 設定
		Google: GoogleOAuthConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		},

		// JWT 設定
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Expiry: getEnvAsDuration("JWT_EXPIRY", "168h"),
		},

		// 加密金鑰
		EncryptionKey: getEnv("ENCRYPTION_KEY", ""),

		// OpenAI 設定
		OpenAI: OpenAIConfig{
			APIKey:    getEnv("OPENAI_API_KEY", ""),
			Model:     getEnv("OPENAI_MODEL", "gpt-4o-mini"),
			MaxTokens: getEnvAsInt("OPENAI_MAX_TOKENS", 2000),
		},

		// 前端 URL
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		// CORS 設定
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),
		},

		// Asynq 設定
		Asynq: AsynqConfig{
			Concurrency:  getEnvAsInt("ASYNQ_CONCURRENCY", 10),
			SyncSchedule: getEnv("SYNC_SCHEDULE", "*/5 * * * *"),
			MaxRetry:     getEnvAsInt("TASK_MAX_RETRY", 3),
		},

		// 郵件同步設定
		EmailSync: EmailSyncConfig{
			InitialSyncDays:  getEnvAsInt("INITIAL_SYNC_DAYS", 30),
			MaxEmailsPerSync: getEnvAsInt("MAX_EMAILS_PER_SYNC", 100),
			SyncCooldownSec:  getEnvAsInt("SYNC_COOLDOWN", 60),
		},

		// AI 分析設定
		AI: AIConfig{
			AutoAnalyze:             getEnvAsBool("AUTO_ANALYZE_EMAILS", true),
			ConfidenceThreshold:     getEnvAsFloat("AI_CONFIDENCE_THRESHOLD", 0.7),
			AutoCreateCaseThreshold: getEnvAsFloat("AUTO_CREATE_CASE_THRESHOLD", 0.85),
		},

		// 通知設定
		Notification: NotificationConfig{
			TaskReminderDays: getEnvAsIntSlice("TASK_REMINDER_DAYS", []int{3, 7}),
			SMTPHost:         getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:         getEnvAsInt("SMTP_PORT", 587),
			SMTPUser:         getEnv("SMTP_USER", ""),
			SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		},

		// 安全設定
		Security: SecurityConfig{
			RateLimitPerMinute:    getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
			SessionCookieName:     getEnv("SESSION_COOKIE_NAME", "influenter_session"),
			SessionCookieSecure:   getEnvAsBool("SESSION_COOKIE_SECURE", false),
			SessionCookieHTTPOnly: getEnvAsBool("SESSION_COOKIE_HTTP_ONLY", true),
			SessionMaxAge:         getEnvAsInt("SESSION_MAX_AGE", 86400),
		},
	}

	// 驗證必要設定
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate 驗證配置的必要欄位
func (c *Config) Validate() error {
	// 資料庫必要欄位
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	// Google OAuth 必要欄位
	if c.Google.ClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	if c.Google.ClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	if c.Google.RedirectURL == "" {
		return fmt.Errorf("GOOGLE_REDIRECT_URL is required")
	}

	// JWT 必要欄位
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	// 加密金鑰必要欄位（開發環境可選，但建議設定）
	if c.IsProduction() && c.EncryptionKey == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required in production")
	}
	// 如果設定了 ENCRYPTION_KEY，驗證其長度（應該是 base64 編碼）
	// 如果是 base64 編碼，解碼後應為 32 bytes
	if c.EncryptionKey != "" {
		// 嘗試解碼 base64
		keyBytes, err := base64.StdEncoding.DecodeString(c.EncryptionKey)
		if err == nil {
			// 成功解碼，檢查解碼後的長度應為 32 bytes
			if len(keyBytes) != 32 {
				return fmt.Errorf("ENCRYPTION_KEY decoded length must be exactly 32 bytes (got %d bytes). To generate a new key, run: go run cmd/generate-key/main.go", len(keyBytes))
			}
		} else {
			// 不是 base64，檢查原始字串長度應為 32 characters
			if len(c.EncryptionKey) != 32 {
				return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters or a valid base64 encoded 32-byte key (got %d characters). To generate a new key, run: go run cmd/generate-key/main.go", len(c.EncryptionKey))
			}
		}
	}

	// OpenAI API Key (在開發環境可選，但生產環境建議要有)
	if c.Env == "production" && c.OpenAI.APIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required in production")
	}

	return nil
}

// IsDevelopment 檢查是否為開發環境
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction 檢查是否為生產環境
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// GetDSN 取得資料庫連線字串
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// Helper functions

// getEnv 取得環境變數，如果不存在則返回預設值
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt 取得環境變數並轉換為整數
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsBool 取得環境變數並轉換為布林值
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsFloat 取得環境變數並轉換為浮點數
func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsDuration 取得環境變數並轉換為時間長度
func getEnvAsDuration(key string, defaultValue string) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		valueStr = defaultValue
	}
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		// 如果解析失敗，返回預設值
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

// getEnvAsSlice 取得環境變數並轉換為字串切片（以逗號分隔）
func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}

// getEnvAsIntSlice 取得環境變數並轉換為整數切片（以逗號分隔）
func getEnvAsIntSlice(key string, defaultValue []int) []int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	parts := strings.Split(valueStr, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if num, err := strconv.Atoi(part); err == nil {
			result = append(result, num)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}

	return result
}
