package config

// Example 展示如何使用 config 套件
//
// 在你的 main.go 中使用:
//
//	import "github.com/designcomb/influenter-backend/internal/config"
//
//	func main() {
//		// 載入配置
//		cfg, err := config.Load()
//		if err != nil {
//			log.Fatalf("Failed to load config: %v", err)
//		}
//
//		// 使用配置
//		fmt.Printf("Environment: %s\n", cfg.Env)
//		fmt.Printf("Server Port: %s\n", cfg.Port)
//		fmt.Printf("Database DSN: %s\n", cfg.GetDSN())
//
//		// 檢查環境
//		if cfg.IsDevelopment() {
//			log.Println("Running in development mode")
//		}
//
//		// 存取巢狀配置
//		fmt.Printf("JWT Expiry: %v\n", cfg.JWT.Expiry)
//		fmt.Printf("OpenAI Model: %s\n", cfg.OpenAI.Model)
//		fmt.Printf("Redis Address: %s\n", cfg.Redis.Addr)
//	}

