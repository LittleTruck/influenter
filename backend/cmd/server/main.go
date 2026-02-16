package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/designcomb/influenter-backend/internal/api"
	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/database"
	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/services/openai"
	"github.com/designcomb/influenter-backend/internal/utils"

	_ "github.com/designcomb/influenter-backend/docs" // Swagger docs
)

// @title           Influenter API
// @version         1.0
// @description     AI 驅動的網紅案件管理系統 API
// @description     提供 Google OAuth 認證、郵件管理、案件管理、AI 分析等功能
// @termsOfService  http://influenter.example.com/terms/

// @contact.name   API Support
// @contact.url    http://influenter.example.com/support
// @contact.email  support@influenter.example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. 載入配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// 2. 初始化結構化日誌
	logger := utils.InitLogger(cfg.Env, cfg.LogLevel)
	logger.Info().
		Str("env", cfg.Env).
		Str("log_level", cfg.LogLevel).
		Msg("Config loaded successfully")

	// 3. 初始化加密工具
	if err := utils.InitCrypto(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize crypto")
	}
	logger.Info().Msg("Crypto initialized successfully")

	// 4. 連接資料庫
	db, err := database.New(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	logger.Info().
		Str("host", cfg.Database.Host).
		Str("database", cfg.Database.Database).
		Msg("Database connected successfully")

	// 5. 設定 Gin 模式
	gin.SetMode(cfg.GinMode)

	// 6. 建立路由
	router := setupRouter(cfg, db, &logger)

	// 7. 啟動伺服器
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info().
		Str("addr", addr).
		Str("env", cfg.Env).
		Str("frontend_url", cfg.FrontendURL).
		Msg("Starting HTTP server")

	logger.Info().Msg("📡 Available endpoints:")
	logger.Info().Msg("   GET  /health                    - Health check")
	logger.Info().Msg("   GET  /swagger/index.html        - API Documentation (Swagger UI)")
	logger.Info().Msg("   GET  /api/v1/ping               - Ping test")
	logger.Info().Msg("   POST /api/v1/auth/google        - Google OAuth login")
	logger.Info().Msg("   GET  /api/v1/auth/me            - Get current user (protected)")
	logger.Info().Msg("   POST /api/v1/auth/logout        - Logout (protected)")
	logger.Info().Msg("   GET  /api/v1/emails             - List emails (protected)")
	logger.Info().Msg("   GET  /api/v1/emails/:id         - Get email (protected)")
	logger.Info().Msg("   PATCH /api/v1/emails/:id        - Update email (protected)")
	logger.Info().Msg("   GET  /api/v1/gmail/status       - Gmail sync status (protected)")
	logger.Info().Msg("   POST /api/v1/gmail/sync         - Trigger sync (protected)")
	logger.Info().Msg("   DELETE /api/v1/gmail/disconnect - Disconnect Gmail (protected)")
	logger.Info().Msg("   GET  /api/v1/cases/fields       - List case fields (protected)")

	if err := router.Run(addr); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}

// setupRouter 設定並返回 Gin router
func setupRouter(cfg *config.Config, db *database.DB, logger *zerolog.Logger) *gin.Engine {
	// 建立 router（不使用預設的 logger）
	router := gin.New()

	// 使用自訂的結構化日誌 middleware
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(corsMiddleware(cfg))

	// Health check endpoint
	router.GET("/health", healthCheckHandler(db))

	// Swagger 文檔
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 建立 handlers
	authHandler := api.NewAuthHandler(db.DB, cfg)
	openaiSvc := openai.NewService(*cfg, logger, "")
	emailHandler := api.NewEmailHandler(db.DB, openaiSvc)
	gmailHandler := api.NewGmailHandler(db.DB)
	caseHandler := api.NewCaseHandler(db.DB, openaiSvc)

	// API v1 路由群組
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", pingHandler)

		// Auth routes
		auth := v1.Group("/auth")
		{
			// 公開路由
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
				emails.POST("/:id/create-case", emailHandler.CreateCaseFromEmail)
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

			// Case routes（/fields 必須在 /:id 之前，否則 "fields" 會被當成 id）
			casesGroup := protected.Group("/cases")
			{
				casesGroup.POST("", caseHandler.CreateCase)
				casesGroup.GET("", caseHandler.ListCases)
				casesGroup.GET("/fields", caseHandler.ListCaseFields)
				casesGroup.GET("/:id", caseHandler.GetCase)
				casesGroup.GET("/:id/emails", caseHandler.ListCaseEmails)
				casesGroup.POST("/:id/draft-reply", caseHandler.DraftReply)
			}
		}
	}

	return router
}

// corsMiddleware 設定 CORS
func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins: cfg.CORS.AllowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// 開發環境允許所有來源（方便測試）
	if cfg.IsDevelopment() {
		corsConfig.AllowAllOrigins = false
		corsConfig.AllowOriginFunc = func(origin string) bool {
			// 允許 localhost 的任何埠號
			return true
		}
	}

	return cors.New(corsConfig)
}

// healthCheckHandler 健康檢查處理器
func healthCheckHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// 檢查資料庫連線
		if err := db.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database is unavailable",
				"error":   err.Error(),
			})
			return
		}

		// 取得資料庫統計資訊
		stats, _ := db.GetStats()

		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"message":   "Influenter API is running",
			"database":  stats,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// pingHandler 測試用的 ping 處理器
// @Summary      Ping 測試
// @Description  測試 API 是否正常運作
// @Tags         系統
// @Produce      json
// @Success      200  {object}  map[string]string  "Pong 回應"
// @Router       /ping [get]
func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
