package api

import (
	"log"
	"net/http"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthHandler 認證處理器
type AuthHandler struct {
	authService *services.AuthService
	config      *config.Config
}

// NewAuthHandler 建立新的認證處理器
func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(db, cfg),
		config:      cfg,
	}
}

// GoogleLoginRequest Google 登入請求
type GoogleLoginRequest struct {
	Credential string `json:"credential" binding:"required"`
	ClientID   string `json:"clientId"`
}

// ErrorResponse 錯誤回應
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// GoogleLogin 處理 Google 登入
// @Summary Google OAuth 登入
// @Description 使用 Google credential 登入並返回 JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GoogleLoginRequest true "Google Login Request"
// @Success 200 {object} services.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// 呼叫認證服務
	response, err := h.authService.GoogleLogin(req.Credential)
	if err != nil {
		log.Printf("Google login error: %v", err)

		// 根據錯誤類型返回不同的 status code
		statusCode := http.StatusUnauthorized
		errorType := "authentication_failed"

		if err == services.ErrInvalidGoogleToken {
			errorType = "invalid_token"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	log.Printf("✅ User logged in: %s (%s)", response.User.Email, response.User.ID)

	c.JSON(http.StatusOK, response)
}

// GetCurrentUser 取得當前登入的使用者
// @Summary 取得當前使用者
// @Description 根據 JWT token 取得當前登入的使用者資訊
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// 從 context 中取得 user_id (由 auth middleware 設定)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	// 取得使用者資訊
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "user_not_found",
				Message: "User not found",
			})
			return
		}

		log.Printf("Error getting user: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get user information",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Logout 登出
// @Summary 登出
// @Description 登出當前使用者（客戶端需清除 token）
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 目前使用 JWT，登出只需要客戶端清除 token 即可
	// 如果未來需要實作 token blacklist，可以在這裡加入

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
