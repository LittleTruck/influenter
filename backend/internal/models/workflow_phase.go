package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowPhase 流程階段
type WorkflowPhase struct {
	ID                 uuid.UUID      `gorm:"primary_key" json:"id"`
	WorkflowTemplateID uuid.UUID      `gorm:"column:workflow_template_id;not null;index" json:"workflow_template_id"`
	Name               string         `gorm:"type:varchar(255);not null" json:"name"`
	DurationDays       int            `gorm:"column:duration_days;not null;default:1" json:"duration_days"`
	Order              int            `gorm:"column:order;not null;default:0" json:"order"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	WorkflowTemplate WorkflowTemplate `gorm:"foreignKey:WorkflowTemplateID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (WorkflowPhase) TableName() string {
	return "workflow_phases"
}

// BeforeCreate GORM hook
func (w *WorkflowPhase) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
