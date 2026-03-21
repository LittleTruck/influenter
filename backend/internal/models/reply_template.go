package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReplyTemplate 回覆範本
type ReplyTemplate struct {
	ID        uuid.UUID      `gorm:"primary_key" json:"id"`
	UserID    uuid.UUID      `gorm:"column:user_id;not null;index" json:"user_id"`
	Title     string         `gorm:"type:varchar(200);not null" json:"title"`
	Prompt    string         `gorm:"type:text;not null" json:"prompt"`
	Order     int            `gorm:"column:order;not null;default:0" json:"order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (ReplyTemplate) TableName() string {
	return "reply_templates"
}

// BeforeCreate GORM hook
func (rt *ReplyTemplate) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}
