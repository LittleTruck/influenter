package middleware

import (
	"net/http"
	"strings"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 認證 middleware
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 Authorization header 取得 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 檢查格式：Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header format must be Bearer {token}",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 驗證 token
		claims, err := utils.ValidateJWT(token, cfg.JWT.Secret)
		if err != nil {
			if err == utils.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "token_expired",
					"message": "Token has expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "invalid_token",
					"message": "Invalid token",
				})
			}
			c.Abort()
			return
		}

		// 將使用者資訊設定到 context
		c.Set("user_id", claims.UserID.String())
		c.Set("user_email", claims.Email)

		c.Next()
	}
}

// OptionalAuthMiddleware 可選的認證 middleware（不強制要求登入）
func OptionalAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 沒有 token，繼續執行但不設定使用者資訊
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Token 格式錯誤，繼續執行但不設定使用者資訊
			c.Next()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateJWT(token, cfg.JWT.Secret)
		if err != nil {
			// Token 無效，繼續執行但不設定使用者資訊
			c.Next()
			return
		}

		// Token 有效，設定使用者資訊
		c.Set("user_id", claims.UserID.String())
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
