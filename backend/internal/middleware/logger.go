package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// responseWriter 自訂的 ResponseWriter 用於捕捉回應內容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggerMiddleware 結構化日誌 middleware（記錄請求和回應）
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 記錄請求開始時間
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 讀取請求 body（需要保存以便後續使用）
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// 恢復 body 以便後續 handler 使用
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// 截斷過長的 body（最多 1000 字元）
				requestBody = string(bodyBytes)
				if len(requestBody) > 1000 {
					requestBody = requestBody[:1000] + "...(truncated)"
				}

				// 過濾敏感資訊
				requestBody = maskSensitiveData(requestBody)
			}
		}

		// 使用自訂的 ResponseWriter 來捕捉回應
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 處理請求
		c.Next()

		// 計算處理時間
		latency := time.Since(start)

		// 取得回應 body
		responseBody := blw.body.String()
		if len(responseBody) > 1000 {
			responseBody = responseBody[:1000] + "...(truncated)"
		}
		responseBody = maskSensitiveData(responseBody)

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

		// 記錄詳細的請求和回應資訊
		logEvent.
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", raw).
			Int("status", c.Writer.Status()).
			Int("size", c.Writer.Size()).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Str("content_type", c.ContentType())

		// 記錄使用者資訊（如果已認證）
		if userID, exists := c.Get("user_id"); exists {
			logEvent.Str("user_id", userID.(string))
		}
		if userEmail, exists := c.Get("user_email"); exists {
			logEvent.Str("user_email", userEmail.(string))
		}

		// 只在 debug 模式記錄 body（避免日誌過大）
		if zerolog.GlobalLevel() <= zerolog.DebugLevel {
			if requestBody != "" {
				logEvent.Str("request_body", requestBody)
			}
			if responseBody != "" {
				logEvent.Str("response_body", responseBody)
			}
		}

		logEvent.Msg("HTTP Request")
	}
}

// maskSensitiveData 遮罩敏感資訊
func maskSensitiveData(data string) string {
	// 遮罩常見的敏感欄位
	sensitiveFields := []string{
		"password", "token", "secret", "credential",
		"authorization", "api_key", "apikey",
	}

	masked := data
	for _, field := range sensitiveFields {
		// 簡單的字串替換（生產環境應使用更強大的 regex）
		if strings.Contains(strings.ToLower(masked), field) {
			masked = strings.ReplaceAll(masked, field, field+":[MASKED]")
		}
	}

	return masked
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
