package utils

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger 初始化全局 logger
func InitLogger(env string, logLevel string) zerolog.Logger {
	// 設定時間格式
	zerolog.TimeFieldFormat = time.RFC3339

	// 設定日誌級別
	level := parseLogLevel(logLevel)
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout

	// 開發環境使用彩色輸出，方便閱讀
	if env == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
	}

	// 建立 logger
	logger := zerolog.New(output).
		With().
		Timestamp().
		Str("service", "influenter-api").
		Str("env", env).
		Logger()

	// 設定為全局 logger
	log.Logger = logger

	return logger
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
