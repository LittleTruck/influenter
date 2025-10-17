package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 初始化全局 logger
func InitLogger(env string, logLevel string) zerolog.Logger {
	// 設定時間格式
	zerolog.TimeFieldFormat = time.RFC3339

	// 設定日誌級別
	level := parseLogLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	var writers []io.Writer

	// 開發環境使用彩色輸出到 console
	if env == "development" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
		writers = append(writers, consoleWriter)
	} else {
		// 生產環境輸出 JSON 到 console
		writers = append(writers, os.Stdout)
	}

	// 添加檔案輸出（所有環境都啟用）
	fileWriter := getFileWriter(env)
	writers = append(writers, fileWriter)

	// 組合多個 writer（同時輸出到 console 和檔案）
	multiWriter := io.MultiWriter(writers...)

	// 建立 logger
	logger := zerolog.New(multiWriter).
		With().
		Timestamp().
		Str("service", "influenter-api").
		Str("env", env).
		Logger()

	// 設定為全局 logger
	log.Logger = logger

	return logger
}

// getFileWriter 取得檔案 writer（支援自動輪轉）
func getFileWriter(env string) io.Writer {
	// 確保 logs 目錄存在
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Warn().Err(err).Msg("Failed to create log directory, using stdout only")
		return os.Stdout
	}

	// 使用日期作為檔案名稱：app-2025-10-17.log
	today := time.Now().Format("2006-01-02")
	var filename string

	if env == "development" {
		filename = filepath.Join(logDir, fmt.Sprintf("app-%s.log", today))
	} else {
		filename = filepath.Join(logDir, fmt.Sprintf("app-%s.json.log", today))
	}

	// 使用 lumberjack 實現日誌輪轉
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    500,  // 每個日誌檔案最大 500 MB（避免單日超過限制）
		MaxBackups: 90,   // 保留最多 90 個舊檔案（約 90 天）
		MaxAge:     90,   // 保留最多 90 天
		Compress:   true, // 壓縮舊檔案
		LocalTime:  true, // 使用本地時間
	}
}

// parseLogLevel 解析日誌級別字串
func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// GetLogger 取得全局 logger
func GetLogger() *zerolog.Logger {
	return &log.Logger
}
