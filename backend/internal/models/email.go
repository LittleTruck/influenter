package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Email 郵件模型
// 用途：儲存從第三方帳號（如 Gmail）同步的郵件
type Email struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OAuthAccountID uuid.UUID `gorm:"type:uuid;not null;index" json:"oauth_account_id"`

	// 郵件提供商原始資訊
	ProviderMessageID string  `gorm:"type:varchar(255);not null;uniqueIndex" json:"provider_message_id"` // Gmail message ID 或其他提供商的 message ID
	ThreadID          *string `gorm:"type:varchar(255);index" json:"thread_id,omitempty"`                // 郵件串 ID

	// 郵件基本資訊
	FromEmail string  `gorm:"type:varchar(255);not null;index" json:"from_email"` // 寄件者 email
	FromName  *string `gorm:"type:varchar(255)" json:"from_name,omitempty"`       // 寄件者名稱
	ToEmail   *string `gorm:"type:varchar(255);index" json:"to_email,omitempty"`  // 收件者 email
	Subject   *string `gorm:"type:text" json:"subject,omitempty"`                 // 郵件主旨
	BodyText  *string `gorm:"type:text" json:"body_text,omitempty"`               // 純文字內容
	BodyHTML  *string `gorm:"type:text" json:"body_html,omitempty"`               // HTML 內容
	Snippet   *string `gorm:"type:text" json:"snippet,omitempty"`                 // 郵件摘要（前 150 字）

	// 郵件屬性
	ReceivedAt     time.Time      `gorm:"not null;index:idx_emails_received_at,sort:desc" json:"received_at"` // 收件時間
	IsRead         bool           `gorm:"default:false" json:"is_read"`                                       // 是否已讀
	HasAttachments bool           `gorm:"default:false" json:"has_attachments"`                               // 是否有附件
	Labels         pq.StringArray `gorm:"type:text[]" json:"labels,omitempty"`                                // 標籤（Gmail labels）

	// AI 分析狀態
	AIAnalyzed   bool       `gorm:"default:false;index:idx_emails_ai_analyzed,where:ai_analyzed = false" json:"ai_analyzed"` // 是否已 AI 分析
	AIAnalysisID *uuid.UUID `gorm:"type:uuid" json:"ai_analysis_id,omitempty"`                                               // AI 分析結果 ID

	// 案件關聯
	CaseID *uuid.UUID `gorm:"type:uuid;index" json:"case_id,omitempty"` // 關聯的案件 ID

	// 系統欄位
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 關聯
	OAuthAccount OAuthAccount `gorm:"foreignKey:OAuthAccountID;constraint:OnDelete:CASCADE" json:"-"`
	// Case         Case         `gorm:"foreignKey:CaseID;constraint:OnDelete:SET NULL" json:"-"` // 未來實作
	// AIAnalysis   AIAnalysis   `gorm:"foreignKey:AIAnalysisID" json:"-"` // 未來實作
}

// TableName 指定表名
func (Email) TableName() string {
	return "emails"
}

// BeforeCreate GORM hook - 在創建前執行
func (e *Email) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// MarkAsRead 標記郵件為已讀
func (e *Email) MarkAsRead(db *gorm.DB) error {
	e.IsRead = true
	return db.Model(e).Update("is_read", true).Error
}

// MarkAsUnread 標記郵件為未讀
func (e *Email) MarkAsUnread(db *gorm.DB) error {
	e.IsRead = false
	return db.Model(e).Update("is_read", false).Error
}

// HasLabel 檢查郵件是否有特定標籤
func (e *Email) HasLabel(label string) bool {
	for _, l := range e.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// IsImportant 檢查郵件是否重要（有 IMPORTANT 標籤）
func (e *Email) IsImportant() bool {
	return e.HasLabel("IMPORTANT")
}

// IsInInbox 檢查郵件是否在收件匣
func (e *Email) IsInInbox() bool {
	return e.HasLabel("INBOX")
}

// EmailListResponse 用於列表 API 回應的結構
type EmailListResponse struct {
	ID             uuid.UUID  `json:"id"`
	FromEmail      string     `json:"from_email"`
	FromName       *string    `json:"from_name,omitempty"`
	Subject        *string    `json:"subject,omitempty"`
	Snippet        *string    `json:"snippet,omitempty"`
	ReceivedAt     time.Time  `json:"received_at"`
	IsRead         bool       `json:"is_read"`
	HasAttachments bool       `json:"has_attachments"`
	Labels         []string   `json:"labels,omitempty"`
	CaseID         *uuid.UUID `json:"case_id,omitempty"`
	AIAnalyzed     bool       `json:"ai_analyzed"`
}

// ToListResponse 轉換為列表 API 回應格式
func (e *Email) ToListResponse() EmailListResponse {
	return EmailListResponse{
		ID:             e.ID,
		FromEmail:      e.FromEmail,
		FromName:       e.FromName,
		Subject:        e.Subject,
		Snippet:        e.Snippet,
		ReceivedAt:     e.ReceivedAt,
		IsRead:         e.IsRead,
		HasAttachments: e.HasAttachments,
		Labels:         e.Labels,
		CaseID:         e.CaseID,
		AIAnalyzed:     e.AIAnalyzed,
	}
}

// EmailDetailResponse 用於詳情 API 回應的結構
type EmailDetailResponse struct {
	ID                uuid.UUID  `json:"id"`
	OAuthAccountID    uuid.UUID  `json:"oauth_account_id"`
	ProviderMessageID string     `json:"provider_message_id"`
	ThreadID          *string    `json:"thread_id,omitempty"`
	FromEmail         string     `json:"from_email"`
	FromName          *string    `json:"from_name,omitempty"`
	ToEmail           *string    `json:"to_email,omitempty"`
	Subject           *string    `json:"subject,omitempty"`
	BodyText          *string    `json:"body_text,omitempty"`
	BodyHTML          *string    `json:"body_html,omitempty"`
	Snippet           *string    `json:"snippet,omitempty"`
	ReceivedAt        time.Time  `json:"received_at"`
	IsRead            bool       `json:"is_read"`
	HasAttachments    bool       `json:"has_attachments"`
	Labels            []string   `json:"labels,omitempty"`
	CaseID            *uuid.UUID `json:"case_id,omitempty"`
	AIAnalyzed        bool       `json:"ai_analyzed"`
	AIAnalysisID      *uuid.UUID `json:"ai_analysis_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ToDetailResponse 轉換為詳情 API 回應格式
func (e *Email) ToDetailResponse() EmailDetailResponse {
	return EmailDetailResponse{
		ID:                e.ID,
		OAuthAccountID:    e.OAuthAccountID,
		ProviderMessageID: e.ProviderMessageID,
		ThreadID:          e.ThreadID,
		FromEmail:         e.FromEmail,
		FromName:          e.FromName,
		ToEmail:           e.ToEmail,
		Subject:           e.Subject,
		BodyText:          e.BodyText,
		BodyHTML:          e.BodyHTML,
		Snippet:           e.Snippet,
		ReceivedAt:        e.ReceivedAt,
		IsRead:            e.IsRead,
		HasAttachments:    e.HasAttachments,
		Labels:            e.Labels,
		CaseID:            e.CaseID,
		AIAnalyzed:        e.AIAnalyzed,
		AIAnalysisID:      e.AIAnalysisID,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}

// EmailQueryParams 郵件查詢參數
type EmailQueryParams struct {
	OAuthAccountID *uuid.UUID `form:"oauth_account_id"`
	IsRead         *bool      `form:"is_read"`
	CaseID         *uuid.UUID `form:"case_id"`
	FromEmail      string     `form:"from_email"`
	Subject        string     `form:"subject"`
	StartDate      *time.Time `form:"start_date"`
	EndDate        *time.Time `form:"end_date"`
	Page           int        `form:"page" binding:"min=1"`
	PageSize       int        `form:"page_size" binding:"min=1,max=100"`
	SortBy         string     `form:"sort_by"`    // received_at, created_at
	SortOrder      string     `form:"sort_order"` // asc, desc
}

// SetDefaults 設定預設值
func (q *EmailQueryParams) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	if q.SortBy == "" {
		q.SortBy = "received_at"
	}
	if q.SortOrder == "" {
		q.SortOrder = "desc"
	}
}
