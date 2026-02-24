package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 使用者模型
type User struct {
	ID                uuid.UUID      `gorm:"primary_key" json:"id"`
	Email             string         `gorm:"uniqueIndex;not null" json:"email"`
	Name              string         `gorm:"not null" json:"name"`
	ProfilePictureURL *string        `json:"profile_picture_url,omitempty"`
	AIInstructions    *string        `json:"ai_instructions,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// 關聯
	OAuthAccounts []OAuthAccount `gorm:"foreignKey:UserID" json:"-"`
}

// GetPrimaryOAuthAccount 取得主要的 OAuth 帳號（通常是用來登入的帳號）
// 優先返回 Google 帳號
func (u *User) GetPrimaryOAuthAccount() *OAuthAccount {
	for i := range u.OAuthAccounts {
		if u.OAuthAccounts[i].Provider == OAuthProviderGoogle {
			return &u.OAuthAccounts[i]
		}
	}
	// 如果沒有 Google 帳號，返回第一個
	if len(u.OAuthAccounts) > 0 {
		return &u.OAuthAccounts[0]
	}
	return nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM hook - 在創建前執行
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
