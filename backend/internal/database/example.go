package database

// Example 展示如何使用 database 套件
//
// 在你的 main.go 中使用:
//
//	import (
//		"context"
//		"log"
//		"github.com/designcomb/influenter-backend/internal/config"
//		"github.com/designcomb/influenter-backend/internal/database"
//	)
//
//	func main() {
//		// 1. 載入配置
//		cfg, err := config.Load()
//		if err != nil {
//			log.Fatalf("Failed to load config: %v", err)
//		}
//
//		// 2. 建立資料庫連線
//		db, err := database.New(cfg)
//		if err != nil {
//			log.Fatalf("Failed to connect to database: %v", err)
//		}
//		defer db.Close()
//
//		// 3. 檢查連線健康狀態
//		ctx := context.Background()
//		if err := db.HealthCheck(ctx); err != nil {
//			log.Printf("Database health check failed: %v", err)
//		}
//
//		// 4. 取得連線池統計資訊
//		stats, err := db.GetStats()
//		if err != nil {
//			log.Printf("Failed to get stats: %v", err)
//		}
//		log.Printf("Database stats: %+v", stats)
//
//		// 5. 執行查詢
//		var users []User
//		result := db.Find(&users)
//		if result.Error != nil {
//			log.Printf("Query failed: %v", result.Error)
//		}
//
//		// 6. 使用分頁
//		var paginatedUsers []User
//		db.Scopes(database.Paginate(1, 20)).Find(&paginatedUsers)
//
//		// 7. 使用排序
//		var sortedUsers []User
//		db.Scopes(database.OrderBy("created_at", true)).Find(&sortedUsers)
//
//		// 8. 使用交易
//		err = db.Transaction(ctx, func(tx *gorm.DB) error {
//			// 在交易中執行多個操作
//			if err := tx.Create(&user).Error; err != nil {
//				return err
//			}
//			if err := tx.Create(&profile).Error; err != nil {
//				return err
//			}
//			return nil
//		})
//		if err != nil {
//			log.Printf("Transaction failed: %v", err)
//		}
//	}
//
// Health Check Endpoint 範例:
//
//	func healthHandler(db *database.DB) gin.HandlerFunc {
//		return func(c *gin.Context) {
//			ctx := c.Request.Context()
//
//			if err := db.HealthCheck(ctx); err != nil {
//				c.JSON(http.StatusServiceUnavailable, gin.H{
//					"status": "error",
//					"error":  "database unavailable",
//				})
//				return
//			}
//
//			stats, _ := db.GetStats()
//			c.JSON(http.StatusOK, gin.H{
//				"status":   "ok",
//				"database": stats,
//			})
//		}
//	}
