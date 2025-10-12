package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// 設定測試環境變數
	setupTestEnv(t)
	defer clearTestEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 驗證基本設定
	if cfg.Env != "test" {
		t.Errorf("Expected ENV=test, got %s", cfg.Env)
	}
	if cfg.Port != "8080" {
		t.Errorf("Expected Port=8080, got %s", cfg.Port)
	}

	// 驗證資料庫設定
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected DB_HOST=localhost, got %s", cfg.Database.Host)
	}
	if cfg.Database.User != "test_user" {
		t.Errorf("Expected DB_USER=test_user, got %s", cfg.Database.User)
	}
	if cfg.Database.Password != "test_pass" {
		t.Errorf("Expected DB_PASSWORD=test_pass, got %s", cfg.Database.Password)
	}
	if cfg.Database.Database != "test_db" {
		t.Errorf("Expected DB_NAME=test_db, got %s", cfg.Database.Database)
	}

	// 驗證 Redis 設定
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("Expected REDIS_ADDR=localhost:6379, got %s", cfg.Redis.Addr)
	}

	// 驗證 Google OAuth 設定
	if cfg.Google.ClientID != "test-client-id" {
		t.Errorf("Expected GOOGLE_CLIENT_ID=test-client-id, got %s", cfg.Google.ClientID)
	}
	if cfg.Google.ClientSecret != "test-client-secret" {
		t.Errorf("Expected GOOGLE_CLIENT_SECRET=test-client-secret, got %s", cfg.Google.ClientSecret)
	}

	// 驗證 JWT 設定
	if cfg.JWT.Secret != "test-jwt-secret-with-minimum-32-chars-required" {
		t.Errorf("Expected JWT_SECRET, got %s", cfg.JWT.Secret)
	}
	if cfg.JWT.Expiry != 168*time.Hour {
		t.Errorf("Expected JWT_EXPIRY=168h, got %v", cfg.JWT.Expiry)
	}

	// 驗證加密金鑰
	if len(cfg.EncryptionKey) != 32 {
		t.Errorf("Expected ENCRYPTION_KEY length=32, got %d", len(cfg.EncryptionKey))
	}

	// 驗證 OpenAI 設定
	if cfg.OpenAI.Model != "gpt-4o-mini" {
		t.Errorf("Expected OPENAI_MODEL=gpt-4o-mini, got %s", cfg.OpenAI.Model)
	}
	if cfg.OpenAI.MaxTokens != 2000 {
		t.Errorf("Expected OPENAI_MAX_TOKENS=2000, got %d", cfg.OpenAI.MaxTokens)
	}
}

func TestValidate_Success(t *testing.T) {
	setupTestEnv(t)
	defer clearTestEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidate_MissingDBUser(t *testing.T) {
	setupTestEnv(t)
	defer clearTestEnv()

	os.Unsetenv("DB_USER")

	cfg, err := Load()
	if err == nil {
		t.Error("Expected error for missing DB_USER, got nil")
	}
	if cfg != nil {
		t.Error("Expected cfg to be nil when validation fails")
	}
}

func TestValidate_MissingDBPassword(t *testing.T) {
	setupTestEnv(t)
	defer clearTestEnv()

	os.Unsetenv("DB_PASSWORD")

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing DB_PASSWORD, got nil")
	}
}

func TestValidate_MissingGoogleClientID(t *testing.T) {
	setupTestEnv(t)
	defer clearTestEnv()

	os.Unsetenv("GOOGLE_CLIENT_ID")

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing GOOGLE_CLIENT_ID, got nil")
	}
}

func TestValidate_ShortJWTSecret(t *testing.T) {
	setupTestEnv(t)
	defer clearTestEnv()

	os.Setenv("JWT_SECRET", "short")

	_, err := Load()
	if err == nil {
		t.Error("Expected error for JWT_SECRET less than 32 characters, got nil")
	}
}

func TestValidate_WrongEncryptionKeyLength(t *testing.T) {
	setupTestEnv(t)
	defer clearTestEnv()

	os.Setenv("ENCRYPTION_KEY", "tooshort")

	_, err := Load()
	if err == nil {
		t.Error("Expected error for ENCRYPTION_KEY not 32 characters, got nil")
	}
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"development", true},
		{"production", false},
		{"staging", false},
		{"test", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			if got := cfg.IsDevelopment(); got != tt.expected {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"production", true},
		{"development", false},
		{"staging", false},
		{"test", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			if got := cfg.IsProduction(); got != tt.expected {
				t.Errorf("IsProduction() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "disable",
		},
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	if got := cfg.GetDSN(); got != expected {
		t.Errorf("GetDSN() = %v, want %v", got, expected)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue int
		expected     int
	}{
		{"Valid integer", "TEST_INT", "42", 0, 42},
		{"Empty value", "TEST_INT", "", 10, 10},
		{"Invalid integer", "TEST_INT", "abc", 20, 20},
		{"Negative integer", "TEST_INT", "-5", 0, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			if got := getEnvAsInt(tt.key, tt.defaultValue); got != tt.expected {
				t.Errorf("getEnvAsInt() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue bool
		expected     bool
	}{
		{"True value", "TEST_BOOL", "true", false, true},
		{"False value", "TEST_BOOL", "false", true, false},
		{"1 as true", "TEST_BOOL", "1", false, true},
		{"0 as false", "TEST_BOOL", "0", true, false},
		{"Empty value", "TEST_BOOL", "", true, true},
		{"Invalid value", "TEST_BOOL", "abc", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			if got := getEnvAsBool(tt.key, tt.defaultValue); got != tt.expected {
				t.Errorf("getEnvAsBool() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetEnvAsFloat(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue float64
		expected     float64
	}{
		{"Valid float", "TEST_FLOAT", "3.14", 0.0, 3.14},
		{"Integer as float", "TEST_FLOAT", "42", 0.0, 42.0},
		{"Empty value", "TEST_FLOAT", "", 1.5, 1.5},
		{"Invalid float", "TEST_FLOAT", "abc", 2.5, 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			if got := getEnvAsFloat(tt.key, tt.defaultValue); got != tt.expected {
				t.Errorf("getEnvAsFloat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     time.Duration
	}{
		{"Valid duration", "TEST_DURATION", "5m", "1m", 5 * time.Minute},
		{"Hours", "TEST_DURATION", "2h", "1h", 2 * time.Hour},
		{"Empty value", "TEST_DURATION", "", "30s", 30 * time.Second},
		{"Invalid duration", "TEST_DURATION", "invalid", "15s", 15 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			if got := getEnvAsDuration(tt.key, tt.defaultValue); got != tt.expected {
				t.Errorf("getEnvAsDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetEnvAsSlice(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue []string
		expected     []string
	}{
		{"Multiple values", "TEST_SLICE", "a,b,c", []string{}, []string{"a", "b", "c"}},
		{"Single value", "TEST_SLICE", "single", []string{}, []string{"single"}},
		{"Empty value", "TEST_SLICE", "", []string{"default"}, []string{"default"}},
		{"With spaces", "TEST_SLICE", "a, b, c", []string{}, []string{"a", " b", " c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsSlice(tt.key, tt.defaultValue)
			if len(got) != len(tt.expected) {
				t.Errorf("getEnvAsSlice() length = %v, want %v", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("getEnvAsSlice()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestGetEnvAsIntSlice(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue []int
		expected     []int
	}{
		{"Multiple integers", "TEST_INT_SLICE", "1,2,3", []int{}, []int{1, 2, 3}},
		{"Single integer", "TEST_INT_SLICE", "42", []int{}, []int{42}},
		{"Empty value", "TEST_INT_SLICE", "", []int{5, 10}, []int{5, 10}},
		{"With spaces", "TEST_INT_SLICE", "1, 2, 3", []int{}, []int{1, 2, 3}},
		{"Mixed valid/invalid", "TEST_INT_SLICE", "1,abc,3", []int{}, []int{1, 3}},
		{"All invalid", "TEST_INT_SLICE", "a,b,c", []int{7}, []int{7}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			got := getEnvAsIntSlice(tt.key, tt.defaultValue)
			if len(got) != len(tt.expected) {
				t.Errorf("getEnvAsIntSlice() length = %v, want %v", len(got), len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("getEnvAsIntSlice()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

// Helper functions for tests

func setupTestEnv(t *testing.T) {
	t.Helper()
	
	// 基本設定
	os.Setenv("ENV", "test")
	os.Setenv("PORT", "8080")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("GIN_MODE", "test")

	// 資料庫設定
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_pass")
	os.Setenv("DB_NAME", "test_db")
	os.Setenv("DB_SSLMODE", "disable")

	// Redis 設定
	os.Setenv("REDIS_ADDR", "localhost:6379")

	// Google OAuth 設定
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GOOGLE_REDIRECT_URL", "http://localhost:8080/callback")

	// JWT 設定
	os.Setenv("JWT_SECRET", "test-jwt-secret-with-minimum-32-chars-required")
	os.Setenv("JWT_EXPIRY", "168h")

	// 加密金鑰 (必須是 32 個字元)
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")

	// OpenAI 設定
	os.Setenv("OPENAI_API_KEY", "test-api-key")
	os.Setenv("OPENAI_MODEL", "gpt-4o-mini")
	os.Setenv("OPENAI_MAX_TOKENS", "2000")

	// 前端 URL
	os.Setenv("FRONTEND_URL", "http://localhost:3000")
}

func clearTestEnv() {
	envVars := []string{
		"ENV", "PORT", "LOG_LEVEL", "GIN_MODE",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"REDIS_ADDR", "REDIS_PASSWORD", "REDIS_DB",
		"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "GOOGLE_REDIRECT_URL",
		"JWT_SECRET", "JWT_EXPIRY",
		"ENCRYPTION_KEY",
		"OPENAI_API_KEY", "OPENAI_MODEL", "OPENAI_MAX_TOKENS",
		"FRONTEND_URL",
	}

	for _, key := range envVars {
		os.Unsetenv(key)
	}
}

