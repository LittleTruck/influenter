package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OAuthProvider OAuth 提供商類型
type OAuthProvider string

const (
	OAuthProviderGoogle  OAuthProvider = "google"
	OAuthProviderOutlook OAuthProvider = "outlook"
	OAuthProviderApple   OAuthProvider = "apple"
)

// SyncStatus 同步狀態
type SyncStatus string

const (
	SyncStatusActive SyncStatus = "active"
	SyncStatusPaused SyncStatus = "paused"
	SyncStatusError  SyncStatus = "error"
)

// OAuthAccount 第三方 OAuth 帳號模型
// 用途：儲存使用者連結的第三方帳號（如 Gmail、Outlook 等）的 OAuth tokens（加密）
type OAuthAccount struct {
	ID     uuid.UUID `gorm:"primary_key" json:"id"`
	UserID uuid.UUID `gorm:"not null;index" json:"user_id"`

	// OAuth 提供商資訊
	Provider   OAuthProvider `gorm:"type:varchar(50);not null;index" json:"provider"` // google, outlook, apple
	ProviderID string        `gorm:"type:varchar(255)" json:"provider_id,omitempty"`  // 提供商的使用者 ID
	Email      string        `gorm:"type:varchar(255);not null" json:"email"`         // 帳號 email

	// OAuth tokens（加密儲存 - 使用 AES-256-GCM）
	AccessToken  string    `gorm:"type:text;not null" json:"-"`  // 加密的 access token
	RefreshToken string    `gorm:"type:text;not null" json:"-"`  // 加密的 refresh token
	TokenExpiry  time.Time `gorm:"not null" json:"token_expiry"` // Token 過期時間

	// 同步狀態（主要用於郵件同步）
	LastSyncAt    *time.Time `json:"last_sync_at,omitempty"`                                     // 最後同步時間
	LastHistoryID *string    `gorm:"type:varchar(100)" json:"last_history_id,omitempty"`         // Gmail API history ID 或其他提供商的同步 ID
	SyncStatus    SyncStatus `gorm:"type:varchar(50);default:'active';index" json:"sync_status"` // active, paused, error
	SyncError     *string    `gorm:"type:text" json:"sync_error,omitempty"`                      // 同步錯誤訊息

	// 額外資訊（JSON 格式，可存放提供商特定資訊）
	Metadata datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`

	// 系統欄位
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 關聯
	User   User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Emails []Email `gorm:"foreignKey:OAuthAccountID" json:"-"`
}

// TableName 指定表名
func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}

// BeforeCreate GORM hook - 在創建前執行
func (oa *OAuthAccount) BeforeCreate(tx *gorm.DB) error {
	if oa.ID == uuid.Nil {
		oa.ID = uuid.New()
	}
	return nil
}

// IsTokenExpired 檢查 token 是否過期
func (oa *OAuthAccount) IsTokenExpired() bool {
	return time.Now().After(oa.TokenExpiry)
}

// IsGmail 檢查是否為 Gmail 帳號
func (oa *OAuthAccount) IsGmail() bool {
	return oa.Provider == OAuthProviderGoogle
}

// IsOutlook 檢查是否為 Outlook 帳號
func (oa *OAuthAccount) IsOutlook() bool {
	return oa.Provider == OAuthProviderOutlook
}

// CanSync 檢查是否可以同步（包含 token 狀態檢查）
// 注意：這個方法主要用於前端顯示狀態
// 實際同步時即使 token 過期也會嘗試（OAuth2 會自動刷新）
func (oa *OAuthAccount) CanSync() bool {
	return oa.SyncStatus == SyncStatusActive && !oa.IsTokenExpired()
}

// OAuthAccountResponse 用於 API 回應的結構（不包含敏感資訊）
type OAuthAccountResponse struct {
	ID             uuid.UUID  `json:"id"`
	Provider       string     `json:"provider"`
	Email          string     `json:"email"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	SyncStatus     string     `json:"sync_status"`
	TokenExpiry    time.Time  `json:"token_expiry"`
	IsTokenExpired bool       `json:"is_token_expired"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ToResponse 轉換為 API 回應格式
func (oa *OAuthAccount) ToResponse() OAuthAccountResponse {
	return OAuthAccountResponse{
		ID:             oa.ID,
		Provider:       string(oa.Provider),
		Email:          oa.Email,
		LastSyncAt:     oa.LastSyncAt,
		SyncStatus:     string(oa.SyncStatus),
		TokenExpiry:    oa.TokenExpiry,
		IsTokenExpired: oa.IsTokenExpired(),
		CreatedAt:      oa.CreatedAt,
	}
}
