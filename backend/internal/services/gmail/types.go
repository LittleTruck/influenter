package gmail

import (
	"time"
)

// Gmail 常用標籤和分類
const (
	// Gmail 系統標籤
	LabelInbox     = "INBOX"
	LabelSent      = "SENT"
	LabelDraft     = "DRAFT"
	LabelSpam      = "SPAM"
	LabelTrash     = "TRASH"
	LabelUnread    = "UNREAD"
	LabelStarred   = "STARRED"
	LabelImportant = "IMPORTANT"

	// Gmail 分類標籤
	LabelCategoryPersonal   = "CATEGORY_PERSONAL"
	LabelCategorySocial     = "CATEGORY_SOCIAL"
	LabelCategoryPromotions = "CATEGORY_PROMOTIONS"
	LabelCategoryUpdates    = "CATEGORY_UPDATES"
	LabelCategoryForums     = "CATEGORY_FORUMS"
)

// ListMessagesOptions 列出郵件的選項
type ListMessagesOptions struct {
	Query        string   // Gmail 查詢語法 (例如: "in:inbox newer_than:7d")
	Labels       []string // 標籤篩選
	MaxResults   int64    // 最大結果數量 (預設 100)
	PageToken    string   // 分頁 token
	IncludeSpam  bool     // 是否包含垃圾郵件
	IncludeTrash bool     // 是否包含已刪除郵件
}

// MessageListResult 郵件列表結果
type MessageListResult struct {
	Messages      []*SimplifiedMessage
	NextPageToken string
	TotalEstimate uint32
}

// SimplifiedMessage 簡化的郵件資訊（列表用）
type SimplifiedMessage struct {
	ID           string
	ThreadID     string
	Snippet      string
	Labels       []string
	InternalDate time.Time
	SizeEstimate int32
}

// ParsedMessage 解析後的完整郵件
type ParsedMessage struct {
	ID           string
	ThreadID     string
	LabelIDs     []string
	Snippet      string
	InternalDate time.Time

	// Headers
	MessageID string
	From      EmailAddress
	To        []EmailAddress
	Cc        []EmailAddress
	Bcc       []EmailAddress
	Subject   string
	Date      time.Time

	// Body
	TextBody string
	HTMLBody string

	// Properties
	HasAttachments bool
	Attachments    []Attachment

	// Gmail specific
	HistoryID    string
	SizeEstimate int32
}

// EmailAddress 郵件地址
type EmailAddress struct {
	Name    string
	Address string
}

// Attachment 附件資訊
type Attachment struct {
	PartID   string
	Filename string
	MimeType string
	Size     int32
}

// SendMessageRequest 寄送郵件請求
type SendMessageRequest struct {
	To         []string
	Cc         []string
	Bcc        []string
	Subject    string
	TextBody   string
	HTMLBody   string
	InReplyTo  string // 回覆郵件的 Message-ID
	References string // 郵件串參考
	ThreadID   string // Gmail thread ID（用於回覆）
}

// SyncResult 同步結果
type SyncResult struct {
	TotalFetched  int
	NewEmails     int
	UpdatedEmails int
	Errors        []error
	LastHistoryID string
	SyncedAt      time.Time
}

// GmailStats Gmail 統計資訊
type GmailStats struct {
	TotalMessages     int64
	UnreadMessages    int64
	StarredMessages   int64
	ImportantMessages int64
	CategoryCounts    map[string]int64 // 各分類的郵件數量
	LabelCounts       map[string]int64 // 各標籤的郵件數量
}

// BatchOperation 批次操作
type BatchOperation struct {
	MessageIDs   []string
	AddLabels    []string // 要新增的標籤
	RemoveLabels []string // 要移除的標籤
}

// ModifyLabelsRequest 修改標籤請求
type ModifyLabelsRequest struct {
	AddLabels    []string
	RemoveLabels []string
}

// SearchOptions 搜尋選項
type SearchOptions struct {
	From          string    // 寄件者
	To            string    // 收件者
	Subject       string    // 主旨關鍵字
	HasWords      string    // 包含的字詞
	DoesntHave    string    // 不包含的字詞
	After         time.Time // 之後
	Before        time.Time // 之前
	HasAttachment bool      // 有附件
	IsUnread      bool      // 未讀
	IsStarred     bool      // 已加星號
	Labels        []string  // 標籤
}

// BuildQuery 建構 Gmail 查詢字串
func (opts *SearchOptions) BuildQuery() string {
	query := ""

	if opts.From != "" {
		query += "from:" + opts.From + " "
	}
	if opts.To != "" {
		query += "to:" + opts.To + " "
	}
	if opts.Subject != "" {
		query += "subject:" + opts.Subject + " "
	}
	if opts.HasWords != "" {
		query += opts.HasWords + " "
	}
	if opts.DoesntHave != "" {
		query += "-" + opts.DoesntHave + " "
	}
	if !opts.After.IsZero() {
		query += "after:" + opts.After.Format("2006/01/02") + " "
	}
	if !opts.Before.IsZero() {
		query += "before:" + opts.Before.Format("2006/01/02") + " "
	}
	if opts.HasAttachment {
		query += "has:attachment "
	}
	if opts.IsUnread {
		query += "is:unread "
	}
	if opts.IsStarred {
		query += "is:starred "
	}
	for _, label := range opts.Labels {
		query += "label:" + label + " "
	}

	return query
}
