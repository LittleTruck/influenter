package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CaseCollaborationItem 案件與合作項目的多對多關聯
type CaseCollaborationItem struct {
	ID                  uuid.UUID `gorm:"primary_key" json:"id"`
	CaseID              uuid.UUID `gorm:"column:case_id;not null;index" json:"case_id"`
	CollaborationItemID uuid.UUID `gorm:"column:collaboration_item_id;not null;index" json:"collaboration_item_id"`
	Price               *float64  `gorm:"column:price;type:numeric(12,2)" json:"price,omitempty"` // 案件個別價格，NULL 表示使用合作項目預設價格
	Order               int       `gorm:"column:order;not null;default:0" json:"order"`
	CreatedAt           time.Time `json:"created_at"`

	// Relationships
	Case              Case              `gorm:"foreignKey:CaseID;constraint:OnDelete:CASCADE" json:"-"`
	CollaborationItem CollaborationItem `gorm:"foreignKey:CollaborationItemID;constraint:OnDelete:CASCADE" json:"collaboration_item,omitempty"`
}

// TableName 指定表名
func (CaseCollaborationItem) TableName() string {
	return "case_collaboration_items"
}

// BeforeCreate GORM hook
func (cci *CaseCollaborationItem) BeforeCreate(tx *gorm.DB) error {
	if cci.ID == uuid.Nil {
		cci.ID = uuid.New()
	}
	return nil
}
