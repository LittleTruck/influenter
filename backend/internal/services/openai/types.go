package openai

import "time"

// EmailCategory 郵件分類
type EmailCategory string

const (
	CategoryCollaboration EmailCategory = "collaboration" // 合作邀約
	CategoryPayment       EmailCategory = "payment"       // 付款相關
	CategoryConfirmation  EmailCategory = "confirmation"  // 確認郵件
	CategoryInquiry       EmailCategory = "inquiry"       // 詢問
	CategorySocial        EmailCategory = "social"        // 社交
	CategoryNewsletter    EmailCategory = "newsletter"    // 訂閱/電子報
	CategoryNotification  EmailCategory = "notification"  // 通知
	CategorySpam          EmailCategory = "spam"          // 垃圾郵件
	CategoryOther         EmailCategory = "other"         // 其他
)

// EmailClassification 郵件分類結果
type EmailClassification struct {
	Category   EmailCategory `json:"category"`   // 分類
	Confidence float64       `json:"confidence"` // 信心指標 (0-1)
	Reason     string        `json:"reason"`     // 分類原因
}

// ExtractedInfo 從郵件中抽取的資訊
type ExtractedInfo struct {
	BrandName      string     `json:"brand_name"`      // 品牌名稱
	ContactName    string     `json:"contact_name"`    // 聯絡人姓名
	ContactEmail   string     `json:"contact_email"`   // 聯絡人郵件
	ContactPhone   string     `json:"contact_phone"`   // 聯絡電話
	Amount         *float64   `json:"amount"`          // 金額
	Currency       string     `json:"currency"`        // 幣別
	DueDate        *time.Time `json:"due_date"`        // 截止日期
	ContentType    string     `json:"content_type"`    // 內容類型（如：影片、圖文等）
	FollowerCount  string     `json:"follower_count"`  // 粉絲數
	Budget         string     `json:"budget"`          // 預算範圍
	ProjectDetails string     `json:"project_details"` // 專案詳情
}

// EmailAnalysisResult AI 分析結果
type EmailAnalysisResult struct {
	Classification EmailClassification `json:"classification"`
	ExtractedInfo  ExtractedInfo       `json:"extracted_info"`
	Summary        string              `json:"summary"`         // 郵件摘要
	KeyPoints      []string            `json:"key_points"`      // 關鍵要點
	ActionRequired bool                `json:"action_required"` // 是否需要行動
	Priority       string              `json:"priority"`        // 優先級: low, medium, high
	TokensUsed     int                 `json:"tokens_used"`     // 使用的 token 數量
	Model          string              `json:"model"`           // 使用的模型
	AnalyzedAt     time.Time           `json:"analyzed_at"`     // 分析時間
}

// AnalysisOptions 分析選項
type AnalysisOptions struct {
	DetailLevel   string // basic, standard, detailed
	IncludeReason bool   // 是否包含分類原因
}

// ClassifyEmailRequest 郵件分類請求
type ClassifyEmailRequest struct {
	Subject string
	Body    string
	From    string
	Options AnalysisOptions
}

// AnalyzeEmailRequest AI 分析郵件請求
type AnalyzeEmailRequest struct {
	Subject string
	Body    string
	From    string
	To      []string
	Date    time.Time
	Options AnalysisOptions
}

// TokenUsage 記錄 token 使用情況
type TokenUsage struct {
	UserID           string
	EmailID          string
	Model            string
	PromptTokens     int       // Prompt 使用的 tokens
	CompletionTokens int       // Completion 使用的 tokens
	TotalTokens      int       // 總 tokens
	CostUSD          float64   // 成本 (USD)
	AnalyzedAt       time.Time // 分析時間
}
