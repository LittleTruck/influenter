package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BundleItem 組合項目與單項的關聯
type BundleItem struct {
	ID        uuid.UUID `gorm:"primary_key" json:"id"`
	BundleID  uuid.UUID `gorm:"column:bundle_id;not null;index" json:"bundle_id"`
	ItemID    uuid.UUID `gorm:"column:item_id;not null;index" json:"item_id"`
	Order     int       `gorm:"column:order;not null;default:0" json:"order"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Bundle CollaborationItem `gorm:"foreignKey:BundleID;constraint:OnDelete:CASCADE" json:"-"`
	Item   CollaborationItem `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE" json:"item,omitempty"`
}

// TableName 指定表名
func (BundleItem) TableName() string {
	return "bundle_items"
}

// BeforeCreate GORM hook
func (bi *BundleItem) BeforeCreate(tx *gorm.DB) error {
	if bi.ID == uuid.Nil {
		bi.ID = uuid.New()
	}
	return nil
}
