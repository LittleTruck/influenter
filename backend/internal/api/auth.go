package api

import (
	"net/http"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/middleware"
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
// @Summary      Google OAuth 登入
// @Description  使用 Google credential 登入系統，如果使用者不存在會自動建立帳號
// @Tags         認證
// @Accept       json
// @Produce      json
// @Param        request  body      GoogleLoginRequest  true  "Google 登入請求"
// @Success      200      {object}  services.LoginResponse  "登入成功，返回使用者資訊和 JWT token"
// @Failure      400      {object}  ErrorResponse  "請求格式錯誤"
// @Failure      401      {object}  ErrorResponse  "Google token 無效或驗證失敗"
// @Failure      500      {object}  ErrorResponse  "伺服器內部錯誤"
// @Router       /auth/google [post]
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
	logger := middleware.GetLogger(c)
	response, err := h.authService.GoogleLogin(req.Credential)
	if err != nil {
		logger.Error().
			Err(err).
			Str("client_id", req.ClientID).
			Msg("Google login failed")

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

	logger.Info().
		Str("user_id", response.User.ID.String()).
		Str("email", response.User.Email).
		Msg("User logged in successfully")

	c.JSON(http.StatusOK, response)
}

// GetCurrentUser 取得當前登入的使用者
// @Summary      取得當前使用者資訊
// @Description  根據 JWT token 取得當前登入的使用者完整資訊
// @Tags         認證
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  github_com_designcomb_influenter-backend_internal_models.User  "使用者資訊"
// @Failure      401  {object}  ErrorResponse  "未認證或 token 無效"
// @Failure      404  {object}  ErrorResponse  "使用者不存在"
// @Failure      500  {object}  ErrorResponse  "伺服器內部錯誤"
// @Router       /auth/me [get]
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
	logger := middleware.GetLogger(c)
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			logger.Warn().
				Str("user_id", userID.String()).
				Msg("User not found")
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "user_not_found",
				Message: "User not found",
			})
			return
		}

		logger.Error().
			Err(err).
			Str("user_id", userID.String()).
			Msg("Failed to get user information")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get user information",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Logout 登出
// @Summary      使用者登出
// @Description  登出當前使用者，客戶端需清除本地儲存的 token
// @Tags         認證
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "登出成功訊息"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 目前使用 JWT，登出只需要客戶端清除 token 即可
	// 如果未來需要實作 token blacklist，可以在這裡加入

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
