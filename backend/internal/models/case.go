package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// CaseStatus 案件狀態
type CaseStatus string

const (
	CaseStatusToConfirm   CaseStatus = "to_confirm"
	CaseStatusInProgress  CaseStatus = "in_progress"
	CaseStatusCompleted   CaseStatus = "completed"
	CaseStatusCancelled   CaseStatus = "cancelled"
	CaseStatusOther       CaseStatus = "other" // 非合作案件（郵件非工作相關）
)

// Case 案件模型
type Case struct {
	ID   uuid.UUID `gorm:"primary_key" json:"id"`
	UserID uuid.UUID `gorm:"column:user_id;not null;index" json:"user_id"`

	Title             string   `gorm:"type:varchar(500);not null" json:"title"`
	BrandName         string   `gorm:"column:brand_name;type:varchar(255);not null" json:"brand_name"`
	Status            CaseStatus `gorm:"type:varchar(50);not null;default:'to_confirm';index" json:"status"`
	CollaborationType *string  `gorm:"column:collaboration_type;type:varchar(255)" json:"collaboration_type,omitempty"`
	Description       *string  `gorm:"type:text" json:"description,omitempty"`

	QuotedAmount *float64 `gorm:"column:quoted_amount" json:"quoted_amount,omitempty"`
	FinalAmount  *float64 `gorm:"column:final_amount" json:"final_amount,omitempty"`
	Currency     *string  `gorm:"type:varchar(10)" json:"currency,omitempty"`

	DeadlineDate *time.Time `gorm:"column:deadline_date;type:date" json:"deadline_date,omitempty"`

	ContactName  *string `gorm:"column:contact_name;type:varchar(255)" json:"contact_name,omitempty"`
	ContactEmail *string `gorm:"column:contact_email;type:varchar(255)" json:"contact_email,omitempty"`
	ContactPhone *string `gorm:"column:contact_phone;type:varchar(100)" json:"contact_phone,omitempty"`

	Notes              *string         `gorm:"type:text" json:"notes,omitempty"`
	Tags               pq.StringArray  `gorm:"type:text[]" json:"tags,omitempty"`
	CollaborationItems pq.StringArray  `gorm:"column:collaboration_items;type:text[]" json:"collaboration_items,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (Case) TableName() string {
	return "cases"
}

// BeforeCreate GORM hook
func (c *Case) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
