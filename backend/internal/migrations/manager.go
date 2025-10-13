package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Migration 代表一個資料庫遷移
type Migration struct {
	Version   string
	Name      string
	UpSQL     string
	DownSQL   string
	Timestamp time.Time
}

// MigrationRecord 記錄已執行的遷移
type MigrationRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Version   string    `gorm:"uniqueIndex;size:255"`
	Name      string    `gorm:"size:255"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// Manager 遷移管理器
type Manager struct {
	db             *gorm.DB
	migrationsPath string
	migrations     []Migration
}

// NewManager 建立遷移管理器
func NewManager(db *gorm.DB, migrationsPath string) (*Manager, error) {
	m := &Manager{
		db:             db,
		migrationsPath: migrationsPath,
	}

	// 確保 migrations 資料夾存在
	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// 建立 migrations 記錄表
	if err := m.ensureMigrationsTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	// 載入所有遷移檔案
	if err := m.loadMigrations(); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	return m, nil
}

// ensureMigrationsTable 確保遷移記錄表存在
func (m *Manager) ensureMigrationsTable() error {
	return m.db.AutoMigrate(&MigrationRecord{})
}

// loadMigrations 從檔案系統載入所有遷移
func (m *Manager) loadMigrations() error {
	m.migrations = []Migration{}

	// 讀取 migrations 目錄
	entries, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// 收集所有 .up.sql 檔案
	upFiles := make(map[string]string)
	downFiles := make(map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			version := strings.TrimSuffix(name, ".up.sql")
			upFiles[version] = filepath.Join(m.migrationsPath, name)
		} else if strings.HasSuffix(name, ".down.sql") {
			version := strings.TrimSuffix(name, ".down.sql")
			downFiles[version] = filepath.Join(m.migrationsPath, name)
		}
	}

	// 建立 Migration 物件
	for version, upPath := range upFiles {
		// 讀取 up SQL
		upSQL, err := os.ReadFile(upPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", upPath, err)
		}

		// 讀取 down SQL (可選)
		var downSQL []byte
		if downPath, exists := downFiles[version]; exists {
			downSQL, err = os.ReadFile(downPath)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", downPath, err)
			}
		}

		// 解析版本和名稱
		parts := strings.SplitN(version, "_", 2)
		timestamp := parts[0]
		name := ""
		if len(parts) > 1 {
			name = parts[1]
		}

		// 解析時間戳
		ts, err := time.Parse("20060102150405", timestamp)
		if err != nil {
			// 如果無法解析，使用當前時間
			ts = time.Now()
		}

		m.migrations = append(m.migrations, Migration{
			Version:   version,
			Name:      name,
			UpSQL:     string(upSQL),
			DownSQL:   string(downSQL),
			Timestamp: ts,
		})
	}

	// 按版本排序
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// Up 執行所有待執行的遷移
func (m *Manager) Up() error {
	// 取得已執行的遷移
	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, record := range appliedMigrations {
		appliedMap[record.Version] = true
	}

	// 執行未執行的遷移
	executed := 0
	for _, migration := range m.migrations {
		if appliedMap[migration.Version] {
			continue
		}

		fmt.Printf("⬆️  Applying migration: %s - %s\n", migration.Version, migration.Name)

		// 執行遷移
		if err := m.executeMigration(migration.UpSQL); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}

		// 記錄遷移
		if err := m.recordMigration(migration); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		fmt.Printf("   ✅ Applied: %s\n", migration.Version)
		executed++
	}

	if executed == 0 {
		fmt.Println("✨ No migrations to apply")
	} else {
		fmt.Printf("✅ Applied %d migration(s)\n", executed)
	}

	return nil
}

// Down 回滾最後一個遷移
func (m *Manager) Down() error {
	// 取得已執行的遷移
	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(appliedMigrations) == 0 {
		fmt.Println("✨ No migrations to rollback")
		return nil
	}

	// 取得最後一個遷移
	lastRecord := appliedMigrations[len(appliedMigrations)-1]

	// 找到對應的遷移
	var migration *Migration
	for i := range m.migrations {
		if m.migrations[i].Version == lastRecord.Version {
			migration = &m.migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found in files", lastRecord.Version)
	}

	if migration.DownSQL == "" {
		return fmt.Errorf("migration %s has no down SQL", migration.Version)
	}

	fmt.Printf("⬇️  Rolling back migration: %s - %s\n", migration.Version, migration.Name)

	// 執行回滾
	if err := m.executeMigration(migration.DownSQL); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", migration.Version, err)
	}

	// 刪除記錄
	if err := m.db.Where("version = ?", migration.Version).Delete(&MigrationRecord{}).Error; err != nil {
		return fmt.Errorf("failed to delete migration record: %w", err)
	}

	fmt.Printf("   ✅ Rolled back: %s\n", migration.Version)

	return nil
}

// Status 顯示遷移狀態
func (m *Manager) Status() error {
	// 取得已執行的遷移
	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, record := range appliedMigrations {
		appliedMap[record.Version] = true
	}

	fmt.Println("\n📋 Migration Status:")
	fmt.Println("-------------------")

	if len(m.migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	for _, migration := range m.migrations {
		status := "❌ Pending"
		if appliedMap[migration.Version] {
			status = "✅ Applied"
		}

		name := migration.Name
		if name == "" {
			name = "(no name)"
		}

		fmt.Printf("%s  %s - %s\n", status, migration.Version, name)
	}

	fmt.Printf("\nTotal: %d migrations, %d applied, %d pending\n",
		len(m.migrations), len(appliedMigrations), len(m.migrations)-len(appliedMigrations))

	return nil
}

// getAppliedMigrations 取得已執行的遷移記錄
func (m *Manager) getAppliedMigrations() ([]MigrationRecord, error) {
	var records []MigrationRecord
	if err := m.db.Order("version").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// executeMigration 執行 SQL 語句
func (m *Manager) executeMigration(sql string) error {
	// 分割多個 SQL 語句
	statements := strings.Split(sql, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if err := m.db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	return nil
}

// recordMigration 記錄已執行的遷移
func (m *Manager) recordMigration(migration Migration) error {
	record := MigrationRecord{
		Version: migration.Version,
		Name:    migration.Name,
	}
	return m.db.Create(&record).Error
}

// CreateMigration 建立新的遷移檔案
func (m *Manager) CreateMigration(name string) error {
	// 生成版本號（時間戳）
	version := time.Now().Format("20060102150405")

	// 清理名稱（移除空格，轉小寫）
	name = strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// 檔案名稱
	filename := fmt.Sprintf("%s_%s", version, name)
	upFile := filepath.Join(m.migrationsPath, filename+".up.sql")
	downFile := filepath.Join(m.migrationsPath, filename+".down.sql")

	// 建立 up 檔案
	upTemplate := fmt.Sprintf(`-- Migration: %s
-- Created at: %s

-- Write your UP migration here

`, name, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(upFile, []byte(upTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// 建立 down 檔案
	downTemplate := fmt.Sprintf(`-- Migration: %s (Rollback)
-- Created at: %s

-- Write your DOWN migration here (to undo the UP migration)

`, name, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(downFile, []byte(downTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("✅ Created migration files:\n")
	fmt.Printf("   📄 %s\n", upFile)
	fmt.Printf("   📄 %s\n", downFile)

	return nil
}
