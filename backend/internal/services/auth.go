package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrInvalidGoogleToken Google token 無效
	ErrInvalidGoogleToken = errors.New("invalid google token")
	// ErrUserNotFound 使用者不存在
	ErrUserNotFound = errors.New("user not found")
)

// AuthService 認證服務
type AuthService struct {
	db     *gorm.DB
	config *config.Config
}

// NewAuthService 建立新的認證服務
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:     db,
		config: cfg,
	}
}

// GoogleTokenInfo Google token 資訊
type GoogleTokenInfo struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`            // Google User ID
	EmailVerified string `json:"email_verified"` // "true" 或 "false"
}

// LoginResponse 登入回應
type LoginResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

// GoogleOAuthData Google OAuth 資料
type GoogleOAuthData struct {
	GoogleID     string
	Email        string
	Name         string
	Picture      string
	AccessToken  string
	RefreshToken string
	TokenExpiry  time.Time
}

// VerifyGoogleToken 驗證 Google ID token
func (s *AuthService) VerifyGoogleToken(idToken string) (*GoogleTokenInfo, error) {
	// 使用 Google 的 tokeninfo endpoint 驗證 token
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrInvalidGoogleToken, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenInfo GoogleTokenInfo
	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse token info: %w", err)
	}

	// 驗證 email 是否已驗證
	if tokenInfo.EmailVerified != "true" {
		return nil, errors.New("email not verified")
	}

	return &tokenInfo, nil
}

// GoogleLogin 處理 Google 登入
func (s *AuthService) GoogleLogin(credential string) (*LoginResponse, error) {
	// 1. 驗證 Google token
	tokenInfo, err := s.VerifyGoogleToken(credential)
	if err != nil {
		return nil, err
	}

	// 2. 查找或建立使用者
	user, err := s.findOrCreateUser(tokenInfo)
	if err != nil {
		return nil, err
	}

	// 3. 生成 JWT token
	jwtToken, err := utils.GenerateJWT(
		user.ID,
		user.Email,
		s.config.JWT.Secret,
		s.config.JWT.Expiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		User:  user,
		Token: jwtToken,
	}, nil
}

// findOrCreateUser 查找或建立使用者（新架構：使用 oauth_accounts）
func (s *AuthService) findOrCreateUser(tokenInfo *GoogleTokenInfo) (*models.User, error) {
	var user models.User
	var oauthAccount models.OAuthAccount

	// 1. 先嘗試透過 oauth_accounts 查找使用者
	result := s.db.Joins("JOIN oauth_accounts ON oauth_accounts.user_id = users.id").
		Where("oauth_accounts.provider = ? AND oauth_accounts.provider_id = ?",
			models.OAuthProviderGoogle, tokenInfo.Sub).
		First(&user)

	if result.Error == nil {
		// 找到使用者，更新基本資訊
		updates := map[string]interface{}{
			"name":                tokenInfo.Name,
			"email":               tokenInfo.Email,
			"profile_picture_url": tokenInfo.Picture,
		}
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return &user, nil
	}

	// 如果是 "record not found" 以外的錯誤，返回錯誤
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 2. 使用者不存在，檢查是否有相同 email 的使用者
	result = s.db.Where("email = ?", tokenInfo.Email).First(&user)
	if result.Error == nil {
		// Email 已存在，為該使用者創建 Google OAuth 帳號記錄
		oauthAccount = models.OAuthAccount{
			ID:         uuid.New(),
			UserID:     user.ID,
			Provider:   models.OAuthProviderGoogle,
			ProviderID: tokenInfo.Sub,
			Email:      tokenInfo.Email,
			// Note: AccessToken 和 RefreshToken 需要在 OAuth callback 中設定
		}
		if err := s.db.Create(&oauthAccount).Error; err != nil {
			return nil, fmt.Errorf("failed to create oauth account: %w", err)
		}

		// 更新使用者資訊
		updates := map[string]interface{}{
			"name":                tokenInfo.Name,
			"profile_picture_url": tokenInfo.Picture,
		}
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return &user, nil
	}

	// 如果是 "record not found" 以外的錯誤，返回錯誤
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 3. 完全新的使用者，創建 User 和 OAuthAccount
	// 使用 Transaction 確保資料一致性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 創建使用者
		user = models.User{
			ID:                uuid.New(),
			Email:             tokenInfo.Email,
			Name:              tokenInfo.Name,
			ProfilePictureURL: &tokenInfo.Picture,
		}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// 創建 OAuth 帳號記錄
		oauthAccount = models.OAuthAccount{
			ID:         uuid.New(),
			UserID:     user.ID,
			Provider:   models.OAuthProviderGoogle,
			ProviderID: tokenInfo.Sub,
			Email:      tokenInfo.Email,
			// Note: AccessToken 和 RefreshToken 需要在 OAuth callback 中設定
		}
		if err := tx.Create(&oauthAccount).Error; err != nil {
			return fmt.Errorf("failed to create oauth account: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GoogleOAuthLogin 處理 Google OAuth 登入並儲存 tokens
func (s *AuthService) GoogleOAuthLogin(oauthData *GoogleOAuthData) (*LoginResponse, error) {
	var user models.User
	var oauthAccount models.OAuthAccount

	// 1. 先嘗試透過 oauth_accounts 查找使用者
	result := s.db.Joins("JOIN oauth_accounts ON oauth_accounts.user_id = users.id").
		Where("oauth_accounts.provider = ? AND oauth_accounts.provider_id = ?",
			models.OAuthProviderGoogle, oauthData.GoogleID).
		First(&user)

	if result.Error == nil {
		// 找到使用者，更新基本資訊和 tokens
		updates := map[string]interface{}{
			"name":                oauthData.Name,
			"email":               oauthData.Email,
			"profile_picture_url": oauthData.Picture,
		}
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		// 更新 OAuth Account tokens
		if err := s.updateOAuthTokens(user.ID, oauthData); err != nil {
			return nil, err
		}

		// 生成 JWT token
		return s.generateLoginResponse(&user)
	}

	// 如果是 "record not found" 以外的錯誤，返回錯誤
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 2. 使用者不存在，檢查是否有相同 email 的使用者
	result = s.db.Where("email = ?", oauthData.Email).First(&user)
	if result.Error == nil {
		// Email 已存在，為該使用者創建 Google OAuth 帳號記錄
		if err := s.createOAuthAccount(user.ID, oauthData); err != nil {
			return nil, err
		}

		// 更新使用者資訊
		updates := map[string]interface{}{
			"name":                oauthData.Name,
			"profile_picture_url": oauthData.Picture,
		}
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return s.generateLoginResponse(&user)
	}

	// 如果是 "record not found" 以外的錯誤，返回錯誤
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// 3. 完全新的使用者，創建 User 和 OAuthAccount
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 創建使用者
		user = models.User{
			ID:                uuid.New(),
			Email:             oauthData.Email,
			Name:              oauthData.Name,
			ProfilePictureURL: &oauthData.Picture,
		}
		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// 創建 OAuth 帳號記錄（含 tokens）
		encryptedAccessToken, err := utils.Encrypt(oauthData.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt access token: %w", err)
		}

		encryptedRefreshToken := ""
		if oauthData.RefreshToken != "" {
			encryptedRefreshToken, err = utils.Encrypt(oauthData.RefreshToken)
			if err != nil {
				return fmt.Errorf("failed to encrypt refresh token: %w", err)
			}
		}

		oauthAccount = models.OAuthAccount{
			ID:           uuid.New(),
			UserID:       user.ID,
			Provider:     models.OAuthProviderGoogle,
			ProviderID:   oauthData.GoogleID,
			Email:        oauthData.Email,
			AccessToken:  encryptedAccessToken,
			RefreshToken: encryptedRefreshToken,
			TokenExpiry:  oauthData.TokenExpiry,
			SyncStatus:   "active",
		}
		if err := tx.Create(&oauthAccount).Error; err != nil {
			return fmt.Errorf("failed to create oauth account: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.generateLoginResponse(&user)
}

// updateOAuthTokens 更新 OAuth tokens
func (s *AuthService) updateOAuthTokens(userID uuid.UUID, oauthData *GoogleOAuthData) error {
	var oauthAccount models.OAuthAccount
	if err := s.db.Where("user_id = ? AND provider = ?", userID, models.OAuthProviderGoogle).
		First(&oauthAccount).Error; err != nil {
		return fmt.Errorf("failed to find oauth account: %w", err)
	}

	// 加密 tokens
	encryptedAccessToken, err := utils.Encrypt(oauthData.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	updates := map[string]interface{}{
		"access_token": encryptedAccessToken,
		"token_expiry": oauthData.TokenExpiry,
		"email":        oauthData.Email,
		"sync_status":  "active",
		"sync_error":   nil,
	}

	// 如果有 refresh token，也更新
	if oauthData.RefreshToken != "" {
		encryptedRefreshToken, err := utils.Encrypt(oauthData.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		updates["refresh_token"] = encryptedRefreshToken
	}

	return s.db.Model(&oauthAccount).Updates(updates).Error
}

// createOAuthAccount 創建 OAuth 帳號記錄
func (s *AuthService) createOAuthAccount(userID uuid.UUID, oauthData *GoogleOAuthData) error {
	encryptedAccessToken, err := utils.Encrypt(oauthData.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encryptedRefreshToken := ""
	if oauthData.RefreshToken != "" {
		encryptedRefreshToken, err = utils.Encrypt(oauthData.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	oauthAccount := models.OAuthAccount{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     models.OAuthProviderGoogle,
		ProviderID:   oauthData.GoogleID,
		Email:        oauthData.Email,
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
		TokenExpiry:  oauthData.TokenExpiry,
		SyncStatus:   "active",
	}

	return s.db.Create(&oauthAccount).Error
}

// generateLoginResponse 生成登入回應
func (s *AuthService) generateLoginResponse(user *models.User) (*LoginResponse, error) {
	jwtToken, err := utils.GenerateJWT(
		user.ID,
		user.Email,
		s.config.JWT.Secret,
		s.config.JWT.Expiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		User:  user,
		Token: jwtToken,
	}, nil
}

// UpdateAIInstructions 更新使用者的 AI 注意事項
func (s *AuthService) UpdateAIInstructions(userID uuid.UUID, instructions *string) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("ai_instructions", instructions).Error
}

// GetUserByID 根據 ID 取得使用者
func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
