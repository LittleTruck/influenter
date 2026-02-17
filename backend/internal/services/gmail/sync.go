package gmail

import (
	"context"
	"fmt"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SyncService 郵件同步服務
type SyncService struct {
	db           *gorm.DB
	gmailService *Service
	oauthAccount *models.OAuthAccount
}

// NewSyncService 建立新的同步服務
func NewSyncService(db *gorm.DB, oauthAccount *models.OAuthAccount) (*SyncService, error) {
	gmailService, err := NewService(db, oauthAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service: %w", err)
	}

	return &SyncService{
		db:           db,
		gmailService: gmailService,
		oauthAccount: oauthAccount,
	}, nil
}

// InitialSync 首次同步（抓取最近 100 封收件 + 100 封寄件）
func (s *SyncService) InitialSync(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		SyncedAt: time.Now(),
	}

	// 同步收件匣：最近 7 天，限制 100 封
	queryInbox := "in:inbox newer_than:7d"
	if _, err := s.syncWithQueryLimited(ctx, queryInbox, result, 100); err != nil {
		return nil, err
	}

	// 同步已寄出：最近 7 天，限制 100 封
	querySent := "in:sent newer_than:7d"
	res2, err := s.syncWithQueryLimited(ctx, querySent, result, 100)
	if err != nil {
		return result, nil // 收件已成功，寄件失敗不阻斷
	}
	result.TotalFetched += res2.TotalFetched
	result.NewEmails += res2.NewEmails
	result.UpdatedEmails += res2.UpdatedEmails
	if len(res2.Errors) > 0 {
		result.Errors = append(result.Errors, res2.Errors...)
	}

	return result, nil
}

// IncrementalSync 增量同步（收件匣 + 已寄出）
func (s *SyncService) IncrementalSync(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		SyncedAt: time.Now(),
	}

	afterClause := ""
	if s.oauthAccount.LastSyncAt != nil {
		afterDate := s.oauthAccount.LastSyncAt.Add(-1 * time.Minute)
		afterClause = " after:" + afterDate.Format("2006/01/02")
	} else {
		afterClause = " newer_than:30d"
	}

	// 同步收件匣
	queryInbox := "in:inbox" + afterClause
	if _, err := s.syncWithQuery(ctx, queryInbox, result); err != nil {
		return nil, err
	}

	// 同步已寄出
	querySent := "in:sent" + afterClause
	res2, err := s.syncWithQuery(ctx, querySent, result)
	if err != nil {
		return result, nil
	}
	result.TotalFetched += res2.TotalFetched
	result.NewEmails += res2.NewEmails
	result.UpdatedEmails += res2.UpdatedEmails
	if len(res2.Errors) > 0 {
		result.Errors = append(result.Errors, res2.Errors...)
	}

	return result, nil
}

// HistorySync 使用 Gmail History API 進行增量同步（更高效）
func (s *SyncService) HistorySync(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		SyncedAt: time.Now(),
	}

	// 需要有 last_history_id 才能使用 History API
	if s.oauthAccount.LastHistoryID == nil || *s.oauthAccount.LastHistoryID == "" {
		// 回退到一般增量同步
		return s.IncrementalSync(ctx)
	}

	// 將 string history ID 轉換為 uint64
	var historyID uint64
	if _, err := fmt.Sscanf(*s.oauthAccount.LastHistoryID, "%d", &historyID); err != nil {
		return nil, fmt.Errorf("invalid history id: %w", err)
	}

	// 取得歷史記錄
	histories, err := s.gmailService.GetHistory(historyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	// 處理歷史記錄中的變更
	messageIDs := make(map[string]bool)
	for _, history := range histories {
		// 新增的郵件
		for _, msg := range history.MessagesAdded {
			messageIDs[msg.Message.Id] = true
		}
		// 刪除的郵件
		for _, msg := range history.MessagesDeleted {
			// 從資料庫中標記刪除
			if err := s.markEmailAsDeleted(msg.Message.Id); err != nil {
				result.Errors = append(result.Errors, err)
			}
		}
		// 標籤變更的郵件
		for _, msg := range history.LabelsAdded {
			messageIDs[msg.Message.Id] = true
		}
		for _, msg := range history.LabelsRemoved {
			messageIDs[msg.Message.Id] = true
		}
	}

	// 批次取得並更新郵件
	for msgID := range messageIDs {
		if err := s.fetchAndSaveMessage(msgID); err != nil {
			result.Errors = append(result.Errors, err)
		} else {
			result.TotalFetched++
		}
	}

	return result, nil
}

// syncWithQuery 使用查詢同步郵件
func (s *SyncService) syncWithQuery(ctx context.Context, query string, result *SyncResult) (*SyncResult, error) {
	return s.syncWithQueryLimited(ctx, query, result, 0) // 0 表示無限制
}

// syncWithQueryLimited 使用查詢同步郵件（可限制數量）
func (s *SyncService) syncWithQueryLimited(ctx context.Context, query string, result *SyncResult, maxEmails int) (*SyncResult, error) {
	opts := &ListMessagesOptions{
		Query:      query,
		MaxResults: 100,
	}

	var allMessageIDs []string

	// 分頁獲取所有郵件 ID
	for {
		listResult, err := s.gmailService.ListMessages(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list messages: %w", err)
		}

		for _, msg := range listResult.Messages {
			allMessageIDs = append(allMessageIDs, msg.ID)
			// 如果設定了最大數量限制，達到後就停止
			if maxEmails > 0 && len(allMessageIDs) >= maxEmails {
				break
			}
		}

		// 如果達到限制或沒有下一頁，結束
		if (maxEmails > 0 && len(allMessageIDs) >= maxEmails) || listResult.NextPageToken == "" {
			break
		}

		opts.PageToken = listResult.NextPageToken
	}

	result.TotalFetched = len(allMessageIDs)

	// 批次處理郵件（每次處理 50 封，避免過載）
	batchSize := 50
	for i := 0; i < len(allMessageIDs); i += batchSize {
		end := i + batchSize
		if end > len(allMessageIDs) {
			end = len(allMessageIDs)
		}

		batch := allMessageIDs[i:end]

		for _, msgID := range batch {
			// 檢查是否已存在（去重）
			exists, err := s.emailExists(msgID)
			if err != nil {
				result.Errors = append(result.Errors, err)
				continue
			}

			if exists {
				// 已存在，更新
				if err := s.updateEmail(msgID); err != nil {
					result.Errors = append(result.Errors, err)
				} else {
					result.UpdatedEmails++
				}
			} else {
				// 不存在，創建
				if err := s.fetchAndSaveMessage(msgID); err != nil {
					result.Errors = append(result.Errors, err)
				} else {
					result.NewEmails++
				}
			}
		}
	}

	// 更新 oauth_account 的同步狀態
	if err := s.updateSyncStatus(result); err != nil {
		return nil, fmt.Errorf("failed to update sync status: %w", err)
	}

	return result, nil
}

// fetchAndSaveMessage 取得並儲存單封郵件
func (s *SyncService) fetchAndSaveMessage(messageID string) error {
	// 從 Gmail API 取得郵件
	gmailMsg, err := s.gmailService.GetMessage(messageID)
	if err != nil {
		return fmt.Errorf("failed to get message %s: %w", messageID, err)
	}

	// 解析為我們的 model
	email, err := ParseMessage(gmailMsg, s.oauthAccount.ID)
	if err != nil {
		return fmt.Errorf("failed to parse message %s: %w", messageID, err)
	}

	// 儲存到資料庫
	if err := s.db.Create(email).Error; err != nil {
		return fmt.Errorf("failed to save message %s: %w", messageID, err)
	}

	// 若為寄出信且同 thread 已有案件關聯，補上 case_id（寄出時寫入失敗的補救）
	if email.Direction == models.EmailDirectionOutgoing && email.ThreadID != nil && *email.ThreadID != "" {
		var caseIDs []uuid.UUID
		s.db.Model(&models.Email{}).
			Where("thread_id = ? AND oauth_account_id = ? AND case_id IS NOT NULL", *email.ThreadID, s.oauthAccount.ID).
			Limit(1).
			Pluck("case_id", &caseIDs)
		if len(caseIDs) > 0 {
			_ = s.db.Model(email).Update("case_id", caseIDs[0])
		}
	}

	return nil
}

// updateEmail 更新現有郵件（主要更新標籤狀態）
func (s *SyncService) updateEmail(messageID string) error {
	// 從 Gmail API 取得最新狀態
	gmailMsg, err := s.gmailService.GetMessage(messageID)
	if err != nil {
		return fmt.Errorf("failed to get message %s: %w", messageID, err)
	}

	// 查詢現有郵件
	var email models.Email
	if err := s.db.Where("provider_message_id = ?", messageID).First(&email).Error; err != nil {
		return fmt.Errorf("failed to find message %s: %w", messageID, err)
	}

	// 更新欄位
	email.Labels = gmailMsg.LabelIds // pq.StringArray 類型會自動處理
	email.IsRead = !contains(gmailMsg.LabelIds, LabelUnread)

	// 儲存更新
	if err := s.db.Save(&email).Error; err != nil {
		return fmt.Errorf("failed to update message %s: %w", messageID, err)
	}

	return nil
}

// emailExists 檢查郵件是否已存在資料庫
func (s *SyncService) emailExists(messageID string) (bool, error) {
	var count int64
	err := s.db.Model(&models.Email{}).
		Where("provider_message_id = ? AND oauth_account_id = ?", messageID, s.oauthAccount.ID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// markEmailAsDeleted 標記郵件為已刪除（軟刪除）
func (s *SyncService) markEmailAsDeleted(messageID string) error {
	err := s.db.Model(&models.Email{}).
		Where("provider_message_id = ?", messageID).
		Update("deleted_at", time.Now()).Error

	return err
}

// updateSyncStatus 更新同步狀態
func (s *SyncService) updateSyncStatus(result *SyncResult) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_sync_at": now,
		"sync_status":  models.SyncStatusActive,
		"sync_error":   nil,
	}

	// 如果有錯誤，記錄第一個錯誤訊息
	if len(result.Errors) > 0 {
		updates["sync_status"] = models.SyncStatusError
		updates["sync_error"] = result.Errors[0].Error()
	}

	// TODO: 更新 last_history_id（需要從最新的郵件取得）

	err := s.db.Model(&models.OAuthAccount{}).
		Where("id = ?", s.oauthAccount.ID).
		Updates(updates).Error

	return err
}

// SyncSpecificLabels 同步特定標籤的郵件
func (s *SyncService) SyncSpecificLabels(ctx context.Context, labels []string, maxDays int) (*SyncResult, error) {
	result := &SyncResult{
		SyncedAt: time.Now(),
	}

	// 建構查詢
	query := fmt.Sprintf("newer_than:%dd", maxDays)

	opts := &ListMessagesOptions{
		Query:      query,
		Labels:     labels,
		MaxResults: 100,
	}

	return s.syncWithListOptions(ctx, opts, result)
}

// syncWithListOptions 使用 ListOptions 同步
func (s *SyncService) syncWithListOptions(ctx context.Context, opts *ListMessagesOptions, result *SyncResult) (*SyncResult, error) {
	var allMessageIDs []string

	// 分頁獲取所有郵件 ID
	for {
		listResult, err := s.gmailService.ListMessages(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list messages: %w", err)
		}

		for _, msg := range listResult.Messages {
			allMessageIDs = append(allMessageIDs, msg.ID)
		}

		if listResult.NextPageToken == "" {
			break
		}

		opts.PageToken = listResult.NextPageToken
	}

	result.TotalFetched = len(allMessageIDs)

	// 批次處理
	for _, msgID := range allMessageIDs {
		exists, err := s.emailExists(msgID)
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		if !exists {
			if err := s.fetchAndSaveMessage(msgID); err != nil {
				result.Errors = append(result.Errors, err)
			} else {
				result.NewEmails++
			}
		}
	}

	// 更新同步狀態
	if err := s.updateSyncStatus(result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetSyncStats 取得同步統計資訊
func (s *SyncService) GetSyncStats() (*GmailStats, error) {
	stats := &GmailStats{
		CategoryCounts: make(map[string]int64),
		LabelCounts:    make(map[string]int64),
	}

	// 總郵件數
	if err := s.db.Model(&models.Email{}).
		Where("oauth_account_id = ?", s.oauthAccount.ID).
		Count(&stats.TotalMessages).Error; err != nil {
		return nil, err
	}

	// 未讀郵件數
	if err := s.db.Model(&models.Email{}).
		Where("oauth_account_id = ? AND is_read = ?", s.oauthAccount.ID, false).
		Count(&stats.UnreadMessages).Error; err != nil {
		return nil, err
	}

	// 為了跨資料庫相容，我們先取得所有郵件，然後在應用層進行統計
	// 這對於測試環境是可行的（因為數據量小），在生產環境可能需要優化
	var emails []models.Email
	if err := s.db.Model(&models.Email{}).
		Where("oauth_account_id = ?", s.oauthAccount.ID).
		Find(&emails).Error; err != nil {
		return nil, err
	}

	// 統計星號郵件和重要郵件
	categories := []string{
		LabelCategoryPersonal,
		LabelCategorySocial,
		LabelCategoryPromotions,
		LabelCategoryUpdates,
		LabelCategoryForums,
	}

	for _, email := range emails {
		if contains(email.Labels, LabelStarred) {
			stats.StarredMessages++
		}
		if contains(email.Labels, LabelImportant) {
			stats.ImportantMessages++
		}
		// Gmail 分類統計
		for _, category := range categories {
			if contains(email.Labels, category) {
				stats.CategoryCounts[category]++
			}
		}
	}

	return stats, nil
}

// CanSync 檢查是否可以同步（防止頻繁同步）
func (s *SyncService) CanSync(cooldownMinutes int) (bool, time.Duration, error) {
	if s.oauthAccount.LastSyncAt == nil {
		return true, 0, nil
	}

	elapsed := time.Since(*s.oauthAccount.LastSyncAt)
	cooldown := time.Duration(cooldownMinutes) * time.Minute

	if elapsed < cooldown {
		remaining := cooldown - elapsed
		return false, remaining, nil
	}

	return true, 0, nil
}

// UpdateEmailFromGmail 從 Gmail 更新單封郵件的狀態
func (s *SyncService) UpdateEmailFromGmail(emailID string) error {
	// 從資料庫取得郵件
	var email models.Email
	if err := s.db.First(&email, emailID).Error; err != nil {
		return fmt.Errorf("email not found: %w", err)
	}

	// 從 Gmail API 取得最新狀態
	gmailMsg, err := s.gmailService.GetMessage(email.ProviderMessageID)
	if err != nil {
		return fmt.Errorf("failed to get message from gmail: %w", err)
	}

	// 更新標籤和已讀狀態
	email.Labels = gmailMsg.LabelIds // pq.StringArray 類型會自動處理
	email.IsRead = !contains(gmailMsg.LabelIds, LabelUnread)

	// 儲存更新
	return s.db.Save(&email).Error
}
