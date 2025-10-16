package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 使用者模型
type User struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email             string         `gorm:"uniqueIndex;not null" json:"email"`
	Name              string         `gorm:"not null" json:"name"`
	GoogleID          *string        `gorm:"uniqueIndex" json:"google_id,omitempty"`
	ProfilePictureURL *string        `json:"profile_picture_url,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
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
