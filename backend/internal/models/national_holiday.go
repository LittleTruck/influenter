package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NationalHoliday 國定假日 / 補班日資料，作為工作日計算的單一真相來源。
//
// 設計上預留 per-account 擴充：
//   - Source='national'：系統內建的國定假日（全域共用，OAuthAccountID 為 nil）。
//   - Source='custom' + OAuthAccountID：未來的帳號自訂假日（drop-in，無需改 schema）。
//   - IsWorkday=true：補班日（即使落在週末仍須上班）。台灣自 2025 下半年起取消補班，
//     目前無補班列，欄位保留以備未來或其他地區使用。
type NationalHoliday struct {
	ID             uuid.UUID  `gorm:"primary_key" json:"id"`
	Date           time.Time  `gorm:"column:date;type:date;not null;index" json:"date"`
	Name           string     `gorm:"type:varchar(255);not null" json:"name"`
	IsWorkday      bool       `gorm:"column:is_workday;not null;default:false" json:"is_workday"`
	Source         string     `gorm:"type:varchar(50);not null;default:'national';index" json:"source"`
	OAuthAccountID *uuid.UUID `gorm:"column:oauth_account_id;index" json:"oauth_account_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (NationalHoliday) TableName() string {
	return "national_holidays"
}

// BeforeCreate GORM hook
func (h *NationalHoliday) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}
