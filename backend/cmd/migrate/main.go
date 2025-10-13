package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/database"
	"github.com/designcomb/influenter-backend/internal/migrations"
)

func main() {
	// 檢查命令列參數
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// 連接資料庫
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 取得 migrations 資料夾路徑
	migrationsPath := getMigrationsPath()

	// 建立遷移管理器
	manager, err := migrations.NewManager(db.DB, migrationsPath)
	if err != nil {
		log.Fatalf("❌ Failed to create migration manager: %v", err)
	}

	// 執行命令
	switch command {
	case "up":
		if err := manager.Up(); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}

	case "down":
		if err := manager.Down(); err != nil {
			log.Fatalf("❌ Rollback failed: %v", err)
		}

	case "status":
		if err := manager.Status(); err != nil {
			log.Fatalf("❌ Failed to get status: %v", err)
		}

	case "create":
		if len(os.Args) < 3 {
			fmt.Println("❌ Error: migration name required")
			fmt.Println("Usage: migrate create <migration_name>")
			os.Exit(1)
		}
		migrationName := os.Args[2]
		if err := manager.CreateMigration(migrationName); err != nil {
			log.Fatalf("❌ Failed to create migration: %v", err)
		}

	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// printUsage 印出使用說明
func printUsage() {
	fmt.Println("Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate up              Run all pending migrations")
	fmt.Println("  migrate down            Rollback the last migration")
	fmt.Println("  migrate status          Show migration status")
	fmt.Println("  migrate create <name>   Create a new migration")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  migrate create create_users_table")
	fmt.Println("  migrate up")
	fmt.Println("  migrate status")
	fmt.Println("  migrate down")
}

// getMigrationsPath 取得 migrations 資料夾的絕對路徑
func getMigrationsPath() string {
	// 嘗試從環境變數取得
	if path := os.Getenv("MIGRATIONS_PATH"); path != "" {
		return path
	}

	// 取得當前執行檔的目錄
	exe, err := os.Executable()
	if err != nil {
		// 如果失敗，使用相對路徑
		return "./migrations"
	}

	// 回到專案根目錄的 migrations 資料夾
	// 通常執行檔在 backend/cmd/migrate 或 backend/tmp
	dir := filepath.Dir(exe)

	// 先檢查是否在 backend/tmp (Air hot reload)
	if filepath.Base(dir) == "tmp" {
		// backend/tmp -> backend -> backend/migrations
		return filepath.Join(filepath.Dir(dir), "migrations")
	}

	// 檢查是否在 backend/cmd/migrate
	if filepath.Base(dir) == "migrate" {
		// backend/cmd/migrate -> backend/cmd -> backend -> backend/migrations
		return filepath.Join(filepath.Dir(filepath.Dir(dir)), "migrations")
	}

	// 預設使用相對路徑
	return "./migrations"
}
