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
	AgencyName        *string  `gorm:"column:agency_name;type:varchar(255)" json:"agency_name,omitempty"`
	Status            CaseStatus `gorm:"type:varchar(50);not null;default:'to_confirm';index" json:"status"`
	CollaborationType *string  `gorm:"column:collaboration_type;type:varchar(255)" json:"collaboration_type,omitempty"`
	Description       *string  `gorm:"type:text" json:"description,omitempty"`

	QuotedAmount *float64 `gorm:"column:quoted_amount" json:"quoted_amount,omitempty"`
	FinalAmount  *float64 `gorm:"column:final_amount" json:"final_amount,omitempty"`
	// AdjustedTotal 手動調整後的案件總價（廠商殺價時微調）。為 nil 時使用合作項目價格加總。
	AdjustedTotal *float64 `gorm:"column:adjusted_total" json:"adjusted_total,omitempty"`
	Currency      *string  `gorm:"type:varchar(10)" json:"currency,omitempty"`

	DeadlineDate *time.Time `gorm:"column:deadline_date;type:date" json:"deadline_date,omitempty"`

	ContactName  *string `gorm:"column:contact_name;type:varchar(255)" json:"contact_name,omitempty"`
	ContactEmail *string `gorm:"column:contact_email;type:varchar(255)" json:"contact_email,omitempty"`
	ContactPhone *string `gorm:"column:contact_phone;type:varchar(100)" json:"contact_phone,omitempty"`

	Alias              *string         `gorm:"column:alias;type:varchar(100)" json:"alias,omitempty"`
	Notes              *string         `gorm:"type:text" json:"notes,omitempty"`
	Tags               pq.StringArray  `gorm:"type:text[]" json:"tags,omitempty"`
	CollaborationItems pq.StringArray  `gorm:"column:collaboration_items;type:text[]" json:"-"` // Deprecated: use CaseCollaborationItems
	FlowLayout         string          `gorm:"column:flow_layout;type:varchar(20);not null;default:'parallel'" json:"flow_layout"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User                     User                    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	CaseCollaborationItems   []CaseCollaborationItem `gorm:"foreignKey:CaseID" json:"case_collaboration_items,omitempty"`
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
