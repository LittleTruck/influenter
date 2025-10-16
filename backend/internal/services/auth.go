package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// findOrCreateUser 查找或建立使用者
func (s *AuthService) findOrCreateUser(tokenInfo *GoogleTokenInfo) (*models.User, error) {
	var user models.User

	// 先嘗試用 Google ID 查找
	result := s.db.Where("google_id = ?", tokenInfo.Sub).First(&user)
	if result.Error == nil {
		// 找到使用者，更新資訊
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

	// 使用者不存在，建立新使用者
	user = models.User{
		ID:                uuid.New(),
		Email:             tokenInfo.Email,
		Name:              tokenInfo.Name,
		GoogleID:          &tokenInfo.Sub,
		ProfilePictureURL: &tokenInfo.Picture,
	}

	if err := s.db.Create(&user).Error; err != nil {
		// 檢查是否是唯一性衝突（email 已存在）
		if strings.Contains(err.Error(), "duplicate key") {
			// Email 已存在，嘗試更新該使用者的 Google ID
			result := s.db.Where("email = ?", tokenInfo.Email).First(&user)
			if result.Error != nil {
				return nil, result.Error
			}

			// 更新 Google ID 和其他資訊
			updates := map[string]interface{}{
				"google_id":           tokenInfo.Sub,
				"name":                tokenInfo.Name,
				"profile_picture_url": tokenInfo.Picture,
			}
			if err := s.db.Model(&user).Updates(updates).Error; err != nil {
				return nil, fmt.Errorf("failed to link google account: %w", err)
			}
			return &user, nil
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
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
