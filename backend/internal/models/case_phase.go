package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CasePhase 案件階段實例
type CasePhase struct {
	ID              uuid.UUID      `gorm:"primary_key" json:"id"`
	CaseID          uuid.UUID      `gorm:"column:case_id;not null;index" json:"case_id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	StartDate       *time.Time     `gorm:"column:start_date;type:date" json:"start_date"`
	EndDate         *time.Time     `gorm:"column:end_date;type:date" json:"end_date"`
	DurationDays    int            `gorm:"column:duration_days;not null;default:1" json:"duration_days"`
	Order           int            `gorm:"column:order;not null;default:0" json:"order"`
	WorkflowPhaseID *uuid.UUID     `gorm:"column:workflow_phase_id" json:"workflow_phase_id,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Case          Case           `gorm:"foreignKey:CaseID;constraint:OnDelete:CASCADE" json:"-"`
	WorkflowPhase *WorkflowPhase `gorm:"foreignKey:WorkflowPhaseID" json:"-"`
}

// TableName 指定表名
func (CasePhase) TableName() string {
	return "case_phases"
}

// BeforeCreate GORM hook
func (cp *CasePhase) BeforeCreate(tx *gorm.DB) error {
	if cp.ID == uuid.Nil {
		cp.ID = uuid.New()
	}
	return nil
}
