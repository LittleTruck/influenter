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

// ItemPrice 單一合作項目的價格
type ItemPrice struct {
	ItemName string  `json:"item_name"` // 項目名稱（如：IG 貼文、YouTube 影片）
	Price    float64 `json:"price"`     // 該項目的價格
}

// ExtractedInfo 從郵件中抽取的資訊
type ExtractedInfo struct {
	BrandName      string     `json:"brand_name"`      // 品牌名稱
	ContactName    string     `json:"contact_name"`    // 聯絡人姓名
	ContactEmail   string     `json:"contact_email"`   // 聯絡人郵件
	ContactPhone   string     `json:"contact_phone"`   // 聯絡電話
	Amount         *float64   `json:"amount"`          // 金額（總金額）
	Currency       string     `json:"currency"`        // 幣別
	ItemPrices     []ItemPrice `json:"item_prices"`    // 各合作項目的個別價格
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

// DraftReplyExample 用於 few-shot 的單筆過往回信範例（依相似度檢索而來）
type DraftReplyExample struct {
	BrandName         string  // 品牌名稱
	CollaborationType string  // 合作類型
	IncomingSubject   string  // 當時來信主旨（可能為空）
	IncomingBody      string  // 當時來信摘要（已截斷，可能為空）
	ReplyBody         string  // 使用者當時寫的回信（純文字，已截斷）
	Similarity        float64 // 與本次來信情境的相似度分數（除錯/排序用）
}

// DraftReplyRequest 擬回信請求（案件摘要 + 要回覆的郵件 + 可選補充說明）
type DraftReplyRequest struct {
	CaseTitle    string // 案件標題
	BrandName    string // 品牌名稱
	ContactName  string // 聯絡人姓名
	ContactEmail string // 聯絡人 email
	// 要回覆的那封郵件
	EmailFrom    string // 寄件者
	EmailSubject string // 主旨
	EmailBody    string // 內文（純文字）
	Instruction         string // 使用者補充說明（選填）
	UserAIInstructions  string // 使用者 AI 注意事項（全域設定）
	UserAIReplyHeader   string // 使用者信件標頭
	UserAIReplyFooter   string // 使用者信件標尾
	TemplatePrompt      string // 回覆範本提示詞
	// FewShotExamples 依品牌、案件類型、語氣檢索出的過往優質回信，作為 few-shot 範例
	FewShotExamples []DraftReplyExample
}

// 來信分類代碼（用於判斷回信策略與是否需人工介入）
const (
	EmailTypeFirstInviteFull = "first_invite_full"         // 首次邀約（資訊完整）
	EmailTypeFirstInviteLack = "first_invite_insufficient" // 首次邀約（資訊不足）
	EmailTypeBarter          = "barter"                    // 互惠／公關品邀約
	EmailTypeFollowUp        = "follow_up"                 // 追蹤回覆
	EmailTypeNegotiation     = "negotiation"               // 議價討論
	EmailTypeExecution       = "execution"                 // 執行階段（合約、腳本、時程）
	EmailTypeOther           = "other"                     // 其他
)

// EmailTypeLabel 將來信分類代碼轉為正體中文標籤
func EmailTypeLabel(code string) string {
	switch code {
	case EmailTypeFirstInviteFull:
		return "首次邀約（資訊完整）"
	case EmailTypeFirstInviteLack:
		return "首次邀約（資訊不足）"
	case EmailTypeBarter:
		return "互惠／公關品邀約"
	case EmailTypeFollowUp:
		return "追蹤回覆"
	case EmailTypeNegotiation:
		return "議價討論"
	case EmailTypeExecution:
		return "執行階段"
	case EmailTypeOther:
		return "其他"
	default:
		return ""
	}
}

// DraftReplyResult 擬回信結果
type DraftReplyResult struct {
	Draft            string `json:"draft"`              // 回信草稿內文（HTML）
	EmailType        string `json:"email_type"`         // 來信分類代碼
	EmailTypeLabel   string `json:"email_type_label"`   // 來信分類（正體中文）
	NeedsHumanReview bool   `json:"needs_human_review"` // 是否建議人工介入
	ReviewReason     string `json:"review_reason"`      // 建議人工介入的原因
	UsedExamples     int    `json:"used_examples"`      // 實際注入的 few-shot 範例數量
}

// ReplyCaseUpdateRequest 回信後 AI 分析案件更新請求
type ReplyCaseUpdateRequest struct {
	ReplyBody         string // 寄出的回信內容
	EmailSubject      string // 原始郵件主旨
	EmailBody         string // 原始郵件內文
	EmailFrom         string // 原始寄件者
	CaseTitle         string // 案件標題
	CaseStatus        string // 目前案件狀態
	CaseDescription   string // 案件描述
	CaseNotes         string // 案件備註
	CaseQuotedAmount  string // 預估報價（顯示用）
	CaseDeadline      string // 截止日期（顯示用）
}

// ReplyCaseUpdateResult AI 分析後建議的案件更新
type ReplyCaseUpdateResult struct {
	ShouldUpdate      bool        `json:"should_update"`       // 是否有建議更新
	Status            string      `json:"status"`              // 建議的新狀態（to_confirm/in_progress/completed/cancelled/other）
	NotesProgress     string      `json:"notes_progress"`      // 進度說明（附加到 notes）
	DescriptionUpdate string      `json:"description_update"`  // 描述更新
	QuotedAmount      *float64    `json:"quoted_amount"`       // 預估報價
	FinalAmount       *float64    `json:"final_amount"`        // 最終金額
	ItemPrices        []ItemPrice `json:"item_prices"`         // 各合作項目的個別價格
	DeadlineDate      string      `json:"deadline_date"`       // 截止日期 ISO YYYY-MM-DD
	Reason            string      `json:"reason"`              // 更新理由
}

// MatchCollaborationItemsRequest 匹配合作項目請求
type MatchCollaborationItemsRequest struct {
	EmailSubject string                  `json:"email_subject"`
	EmailBody    string                  `json:"email_body"`
	EmailFrom    string                  `json:"email_from"`
	Items        []CollaborationItemInfo `json:"items"`
}

// CollaborationItemInfo 合作項目摘要（用於 AI 匹配）
type CollaborationItemInfo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Type        string   `json:"type"`                   // "individual" or "bundle"
	ItemNames   []string `json:"item_names,omitempty"`   // For bundles: names of contained items
}

// MatchCollaborationItemsResult 合作項目匹配結果
type MatchCollaborationItemsResult struct {
	MatchedItemIDs []string `json:"matched_item_ids"`
	Confidence     float64  `json:"confidence"`
	Reason         string   `json:"reason"`
}

// MatchWorkflowTemplateRequest AI 自動選擇流程範本請求
type MatchWorkflowTemplateRequest struct {
	CaseTitle       string                 `json:"case_title"`
	CaseBrandName   string                 `json:"case_brand_name"`
	CaseDescription string                 `json:"case_description"`
	EmailSubject    string                 `json:"email_subject"`
	EmailBody       string                 `json:"email_body"`
	Templates       []WorkflowTemplateInfo `json:"templates"`
}

// WorkflowTemplateInfo 流程範本摘要（用於 AI 匹配）
type WorkflowTemplateInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Phases      []string `json:"phases"` // 階段名稱列表
}

// MatchWorkflowTemplateResult AI 選擇流程範本結果
type MatchWorkflowTemplateResult struct {
	TemplateID string  `json:"template_id"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
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
