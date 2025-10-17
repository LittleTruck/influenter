package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/designcomb/influenter-backend/internal/api"
	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/database"
	"github.com/designcomb/influenter-backend/internal/middleware"

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
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	log.Printf("✅ Config loaded (env=%s)", cfg.Env)

	// 2. 連接資料庫
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Database connected")

	// 3. 設定 Gin 模式
	gin.SetMode(cfg.GinMode)

	// 4. 建立路由
	router := setupRouter(cfg, db)

	// 5. 啟動伺服器
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Server is starting on %s", addr)
	log.Printf("📝 Environment: %s", cfg.Env)
	log.Printf("🌐 Frontend URL: %s", cfg.FrontendURL)
	log.Println("📡 Available endpoints:")
	log.Println("   GET  /health              - Health check")
	log.Println("   GET  /swagger/index.html  - API Documentation (Swagger UI)")
	log.Println("   GET  /api/v1/ping         - Ping test")
	log.Println("   POST /api/v1/auth/google  - Google OAuth login")
	log.Println("   GET  /api/v1/auth/me      - Get current user (protected)")
	log.Println("   POST /api/v1/auth/logout  - Logout (protected)")

	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// setupRouter 設定並返回 Gin router
func setupRouter(cfg *config.Config, db *database.DB) *gin.Engine {
	// 建立 router（包含 logger 和 recovery middleware）
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware(cfg))

	// Health check endpoint
	router.GET("/health", healthCheckHandler(db))

	// Swagger 文檔
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 建立 auth handler
	authHandler := api.NewAuthHandler(db.DB, cfg)

	// API v1 路由群組
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", pingHandler)

		// Auth routes (公開)
		auth := v1.Group("/auth")
		{
			auth.POST("/google", authHandler.GoogleLogin)
			auth.POST("/logout", authHandler.Logout)

			// 需要認證的路由
			auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.GetCurrentUser)
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
