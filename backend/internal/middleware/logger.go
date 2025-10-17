package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoggerMiddleware 結構化日誌 middleware
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 記錄請求開始時間
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 處理請求
		c.Next()

		// 計算處理時間
		latency := time.Since(start)

		// 建立日誌事件
		logEvent := log.Info()

		// 如果有錯誤，使用 Error level
		if len(c.Errors) > 0 {
			// 將錯誤轉換為字串切片
			errMsgs := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errMsgs[i] = err.Error()
			}
			logEvent = log.Error().Strs("errors", errMsgs)
		} else if c.Writer.Status() >= 500 {
			logEvent = log.Error()
		} else if c.Writer.Status() >= 400 {
			logEvent = log.Warn()
		}

		// 記錄請求資訊
		logEvent.
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", raw).
			Int("status", c.Writer.Status()).
			Int("size", c.Writer.Size()).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP Request")
	}
}

// RequestIDMiddleware 為每個請求生成唯一 ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 header 取得或生成新的 request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 設定到 context
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// 更新 logger context
		logger := log.With().Str("request_id", requestID).Logger()
		c.Set("logger", &logger)

		c.Next()
	}
}

// GetLogger 從 context 取得 logger
func GetLogger(c *gin.Context) *zerolog.Logger {
	if logger, exists := c.Get("logger"); exists {
		return logger.(*zerolog.Logger)
	}
	return &log.Logger
}

// generateRequestID 生成請求 ID（簡單版本）
func generateRequestID() string {
	// 使用時間戳 + 隨機數生成簡單的 request ID
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成隨機字串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
