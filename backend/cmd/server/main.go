package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/database"
)

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
	log.Println("   GET  /health       - Health check")
	log.Println("   GET  /api/v1/ping  - Ping test")

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

	// API v1 路由群組
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", pingHandler)
		// 之後會在這裡加入其他 API endpoints
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
func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
