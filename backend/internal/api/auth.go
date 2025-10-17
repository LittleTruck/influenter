package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

// GoogleOAuthCallbackRequest Google OAuth callback 請求
type GoogleOAuthCallbackRequest struct {
	Code        string `json:"code" binding:"required"`
	RedirectURI string `json:"redirect_uri" binding:"required"`
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

// GoogleOAuthCallback 處理 Google OAuth callback
// @Summary      Google OAuth Callback
// @Description  處理 Google OAuth 授權回調，換取 access token 並建立/更新使用者
// @Tags         認證
// @Accept       json
// @Produce      json
// @Param        request  body      GoogleOAuthCallbackRequest  true  "OAuth callback 請求"
// @Success      200      {object}  services.LoginResponse  "登入成功，返回使用者資訊和 JWT token"
// @Failure      400      {object}  ErrorResponse  "請求格式錯誤"
// @Failure      401      {object}  ErrorResponse  "OAuth 授權失敗"
// @Failure      500      {object}  ErrorResponse  "伺服器內部錯誤"
// @Router       /auth/google/callback [post]
func (h *AuthHandler) GoogleOAuthCallback(c *gin.Context) {
	var req GoogleOAuthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	logger := middleware.GetLogger(c)

	// 1. 建立 OAuth config
	oauth2Config := &oauth2.Config{
		ClientID:     h.config.Google.ClientID,
		ClientSecret: h.config.Google.ClientSecret,
		RedirectURL:  req.RedirectURI,
		Scopes: []string{
			"openid",
			"email",
			"profile",
			"https://www.googleapis.com/auth/gmail.readonly",
			"https://www.googleapis.com/auth/gmail.modify",
			"https://www.googleapis.com/auth/gmail.labels",
		},
		Endpoint: google.Endpoint,
	}

	// 2. 用 authorization code 換取 tokens
	token, err := oauth2Config.Exchange(context.Background(), req.Code)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to exchange code for token")
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "oauth_exchange_failed",
			Message: "Failed to exchange authorization code: " + err.Error(),
		})
		return
	}

	// 3. 用 access token 取得使用者資訊
	client := oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get user info")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "user_info_failed",
			Message: "Failed to get user information",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error().Int("status", resp.StatusCode).Msg("Failed to get user info")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "user_info_failed",
			Message: "Failed to get user information from Google",
		})
		return
	}

	// 4. 解析使用者資訊
	var userInfo struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Sub     string `json:"id"` // Google User ID
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logger.Error().Err(err).Msg("Failed to parse user info")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "parse_failed",
			Message: "Failed to parse user information",
		})
		return
	}

	// 5. 呼叫認證服務處理登入和 token 儲存
	response, err := h.authService.GoogleOAuthLogin(&services.GoogleOAuthData{
		GoogleID:     userInfo.Sub,
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		Picture:      userInfo.Picture,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  token.Expiry,
	})

	if err != nil {
		logger.Error().Err(err).Msg("Failed to complete OAuth login")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "login_failed",
			Message: "Failed to complete login: " + err.Error(),
		})
		return
	}

	logger.Info().
		Str("user_id", response.User.ID.String()).
		Str("email", response.User.Email).
		Bool("has_refresh_token", token.RefreshToken != "").
		Msg("User logged in successfully via OAuth")

	c.JSON(http.StatusOK, response)
}

// Logout 登出
// @Summary      使用者登出
// @Description  登出當前使用者，客戶端需清除本地儲存的 token
// @Tags         認證
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "登出成功訊息"
// @Failure      401  {object}  ErrorResponse  "未認證"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// 取得使用者 ID（由 auth middleware 設定）
	logger := middleware.GetLogger(c)

	userIDStr, exists := c.Get("user_id")
	if exists {
		logger.Info().
			Str("user_id", userIDStr.(string)).
			Msg("User logged out successfully")
	}

	// 目前使用 JWT，登出只需要客戶端清除 token 即可
	// 如果未來需要實作 token blacklist，可以在這裡加入

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
