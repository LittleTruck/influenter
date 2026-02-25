package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowTemplate 流程範本
type WorkflowTemplate struct {
	ID          uuid.UUID      `gorm:"primary_key" json:"id"`
	UserID      uuid.UUID      `gorm:"column:user_id;not null;index" json:"user_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	Color       string         `gorm:"type:varchar(50);not null;default:'primary'" json:"color"`
	Order       int            `gorm:"column:order;not null;default:0" json:"order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User   User            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Phases []WorkflowPhase `gorm:"foreignKey:WorkflowTemplateID" json:"phases"`
}

// TableName 指定表名
func (WorkflowTemplate) TableName() string {
	return "workflow_templates"
}

// BeforeCreate GORM hook
func (w *WorkflowTemplate) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
