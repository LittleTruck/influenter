package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CaseTotalAdjustment 案件總價調整歷史記錄
type CaseTotalAdjustment struct {
	ID            uuid.UUID `gorm:"primary_key" json:"id"`
	CaseID        uuid.UUID `gorm:"column:case_id;not null;index" json:"case_id"`
	UserID        uuid.UUID `gorm:"column:user_id;not null;index" json:"user_id"`
	OriginalTotal float64   `gorm:"column:original_total;type:numeric(12,2);not null" json:"original_total"`
	AdjustedTotal float64   `gorm:"column:adjusted_total;type:numeric(12,2);not null" json:"adjusted_total"`
	Reason        string    `gorm:"column:reason;type:text;not null" json:"reason"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	Case Case `gorm:"foreignKey:CaseID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (CaseTotalAdjustment) TableName() string {
	return "case_total_adjustments"
}

// BeforeCreate GORM hook
func (a *CaseTotalAdjustment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
