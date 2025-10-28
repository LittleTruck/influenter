package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	// 設置測試環境
	os.Setenv("ENCRYPTION_KEY", "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")
	os.Setenv("ENV", "test")
	gin.SetMode(gin.TestMode)

	// 初始化加密工具
	_ = utils.InitCrypto()
}

// setupTestDB 設置測試用的資料庫
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Skipf("Skipping test: SQLite not available: %v", err)
	}

	// Auto migrate
	err = db.AutoMigrate(&models.User{}, &models.OAuthAccount{}, &models.Email{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// getTestConfig 獲取測試用的配置
func getTestConfig() *config.Config {
	return &config.Config{
		Env:      "test",
		Port:     "8080",
		LogLevel: "debug",
		GinMode:  gin.TestMode,
		Google: config.GoogleOAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/callback",
		},
		JWT: config.JWTConfig{
			Secret: "test-jwt-secret-with-minimum-32-chars-required",
			Expiry: 168 * time.Hour,
		},
		EncryptionKey: "12345678901234567890123456789012",
		FrontendURL:   "http://localhost:3000",
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}
}

// setupTestRouter 設置測試用的完整 router
func setupTestRouter(t *testing.T) (*gorm.DB, *gin.Engine, *config.Config) {
	db := setupTestDB(t)
	cfg := getTestConfig()

	// 設置 Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 使用必要的 middleware
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())

	// 建立 handlers
	authHandler := NewAuthHandler(db, cfg)
	emailHandler := NewEmailHandler(db)
	gmailHandler := NewGmailHandler(db)

	// 設置路由
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/google", authHandler.GoogleLogin)
			auth.POST("/google/callback", authHandler.GoogleOAuthCallback)

			// 需要認證的路由
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware(cfg))
			{
				authProtected.GET("/me", authHandler.GetCurrentUser)
				authProtected.POST("/logout", authHandler.Logout)
			}
		}

		// 需要認證的路由群組
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			// Email routes
			emails := protected.Group("/emails")
			{
				emails.GET("", emailHandler.ListEmails)
				emails.GET("/:id", emailHandler.GetEmail)
				emails.PATCH("/:id", emailHandler.UpdateEmail)
			}

			// Gmail integration routes
			gmailGroup := protected.Group("/gmail")
			{
				gmailGroup.GET("/status", gmailHandler.GetStatus)
				gmailGroup.POST("/sync", gmailHandler.TriggerSync)
				gmailGroup.DELETE("/disconnect", gmailHandler.DisconnectGmail)
			}
		}
	}

	return db, router, cfg
}

// createTestUser 創建測試用戶並返回 token
func createTestUser(t *testing.T, db *gorm.DB, cfg *config.Config) (uuid.UUID, string, string) {
	// 產生唯一 email，避免 UNIQUE 衝突
	uniqueEmail := uuid.New().String() + "@example.com"
	user := models.User{
		ID:    uuid.New(),
		Email: uniqueEmail,
		Name:  "Test User",
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// 生成 JWT token（使用測試配置中的有效期限，避免立即過期）
	token, err := utils.GenerateJWT(user.ID, user.Email, cfg.JWT.Secret, cfg.JWT.Expiry)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	return user.ID, token, uniqueEmail
}

// TestGoogleLogin_MissingCredential 測試缺少 credential 的情況
func TestGoogleLogin_MissingCredential(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 發送請求（沒有 credential）
	body := `{}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/google", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

// TestGoogleLogin_InvalidCredential 測試無效的 credential
func TestGoogleLogin_InvalidCredential(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 發送請求（無效的 credential）
	body := `{"credential": "invalid-token"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/google", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, []string{"invalid_token", "authentication_failed"}, response.Error)
}

// TestGetCurrentUser_NoToken 測試未提供 token
func TestGetCurrentUser_NoToken(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", response.Error)
}

// TestGetCurrentUser_InvalidToken 測試無效的 token
func TestGetCurrentUser_InvalidToken(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_token", response.Error)
}

// TestGetCurrentUser_UserNotFound 測試用戶不存在
func TestGetCurrentUser_UserNotFound(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 生成一個有效但用戶不存在的 token（不可立即過期）
	nonExistentUserID := uuid.New()
	token, err := utils.GenerateJWT(nonExistentUserID, "nonexistent@example.com", cfg.JWT.Secret, cfg.JWT.Expiry)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user_not_found", response.Error)
}

// TestGetCurrentUser_Success 測試成功獲取用戶資訊
func TestGetCurrentUser_Success(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 創建測試用戶並獲取 token
	userID, token, email := createTestUser(t, db, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)
}

// TestLogout_NoToken 測試未認證的登出
func TestLogout_NoToken(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

// TestLogout_Success 測試成功登出
func TestLogout_Success(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 創建測試用戶並獲取 token
	_, token, _ := createTestUser(t, db, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])
}
