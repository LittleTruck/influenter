package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CollaborationItemType 合作項目類型
type CollaborationItemType string

const (
	CollaborationItemTypeIndividual CollaborationItemType = "individual"
	CollaborationItemTypeBundle     CollaborationItemType = "bundle"
)

// CollaborationItem 合作項目
type CollaborationItem struct {
	ID          uuid.UUID             `gorm:"primary_key" json:"id"`
	UserID      uuid.UUID             `gorm:"column:user_id;not null;index" json:"user_id"`
	Title       string                `gorm:"type:varchar(500);not null" json:"title"`
	Description *string               `gorm:"type:text" json:"description,omitempty"`
	Notes       *string               `gorm:"type:text" json:"notes,omitempty"`
	Price       float64               `gorm:"column:price;type:numeric(12,2);not null;default:0" json:"price"`
	Type        CollaborationItemType `gorm:"column:type;type:varchar(20);not null;default:'individual'" json:"type"`
	ParentID    *uuid.UUID            `gorm:"column:parent_id;index" json:"-"` // Deprecated: kept for migration rollback
	WorkflowID  *uuid.UUID            `gorm:"column:workflow_id" json:"workflow_id"`
	Order       int                   `gorm:"column:order;not null;default:0" json:"order"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   gorm.DeletedAt        `gorm:"index" json:"-"`

	// Relationships
	User        User              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Workflow    *WorkflowTemplate `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	BundleItems []BundleItem      `gorm:"foreignKey:BundleID" json:"bundle_items,omitempty"`
}

// TableName 指定表名
func (CollaborationItem) TableName() string {
	return "collaboration_items"
}

// BeforeCreate GORM hook
func (ci *CollaborationItem) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return nil
}
