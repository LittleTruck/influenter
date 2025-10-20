package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/services/gmail"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const (
	// TypeEmailSync 郵件同步任務類型
	TypeEmailSync = "email:sync"

	// TypeEmailSyncAll 同步所有使用者的郵件
	TypeEmailSyncAll = "email:sync:all"
)

// EmailSyncPayload 郵件同步任務的 payload
type EmailSyncPayload struct {
	OAuthAccountID string `json:"oauth_account_id"`
	SyncType       string `json:"sync_type"` // initial, incremental, history
}

// EmailSyncAllPayload 同步所有使用者的 payload
type EmailSyncAllPayload struct {
	MaxAccounts int `json:"max_accounts"` // 每次最多同步幾個帳號（避免過載）
}

// NewEmailSyncTask 建立郵件同步任務
func NewEmailSyncTask(oauthAccountID string, syncType string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailSyncPayload{
		OAuthAccountID: oauthAccountID,
		SyncType:       syncType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 設定任務選項
	opts := []asynq.Option{
		asynq.MaxRetry(3),               // 最多重試 3 次
		asynq.Timeout(10 * time.Minute), // 超時時間 10 分鐘
		asynq.Retention(24 * time.Hour), // 保留成功任務 24 小時
	}

	return asynq.NewTask(TypeEmailSync, payload, opts...), nil
}

// NewEmailSyncAllTask 建立同步所有使用者的任務
func NewEmailSyncAllTask(maxAccounts int) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailSyncAllPayload{
		MaxAccounts: maxAccounts,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	opts := []asynq.Option{
		asynq.MaxRetry(2),
		asynq.Timeout(30 * time.Minute),
	}

	return asynq.NewTask(TypeEmailSyncAll, payload, opts...), nil
}

// HandleEmailSyncTask 處理單個帳號的郵件同步任務
func HandleEmailSyncTask(ctx context.Context, t *asynq.Task, db *gorm.DB) error {
	var payload EmailSyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Info().
		Str("oauth_account_id", payload.OAuthAccountID).
		Str("sync_type", payload.SyncType).
		Msg("Starting email sync task")

	// 查詢 OAuth 帳號
	var oauthAccount models.OAuthAccount
	if err := db.First(&oauthAccount, "id = ?", payload.OAuthAccountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("oauth_account_id", payload.OAuthAccountID).Msg("OAuth account not found")
			return nil // 不重試
		}
		return fmt.Errorf("failed to query oauth account: %w", err)
	}

	// 檢查是否為 Gmail 帳號
	if !oauthAccount.IsGmail() {
		log.Warn().
			Str("oauth_account_id", payload.OAuthAccountID).
			Str("provider", string(oauthAccount.Provider)).
			Msg("Not a Gmail account")
		return nil // 不重試
	}

	// 檢查帳號狀態（只檢查 sync_status，不檢查 token 過期）
	// Token 過期時會由 OAuth2 client 自動刷新
	if oauthAccount.SyncStatus != models.SyncStatusActive {
		log.Warn().
			Str("oauth_account_id", payload.OAuthAccountID).
			Str("sync_status", string(oauthAccount.SyncStatus)).
			Msg("Account sync is not active")
		return nil // 不重試
	}

	// 如果 token 已過期，記錄但繼續執行（讓 OAuth2 client 嘗試刷新）
	if oauthAccount.IsTokenExpired() {
		log.Info().
			Str("oauth_account_id", payload.OAuthAccountID).
			Msg("Token expired, will attempt to refresh during sync")
	}

	// 建立同步服務
	syncService, err := gmail.NewSyncService(db, &oauthAccount)
	if err != nil {
		return fmt.Errorf("failed to create sync service: %w", err)
	}

	// 根據同步類型執行同步
	var result *gmail.SyncResult
	switch payload.SyncType {
	case "initial":
		result, err = syncService.InitialSync(ctx)
	case "incremental":
		result, err = syncService.IncrementalSync(ctx)
	case "history":
		result, err = syncService.HistorySync(ctx)
	default:
		// 預設使用增量同步
		result, err = syncService.IncrementalSync(ctx)
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("oauth_account_id", payload.OAuthAccountID).
			Msg("Email sync failed")
		return fmt.Errorf("sync failed: %w", err)
	}

	// 記錄同步結果
	log.Info().
		Str("oauth_account_id", payload.OAuthAccountID).
		Int("total_fetched", result.TotalFetched).
		Int("new_emails", result.NewEmails).
		Int("updated_emails", result.UpdatedEmails).
		Int("errors", len(result.Errors)).
		Msg("Email sync completed successfully")

	// 如果有錯誤，記錄詳細資訊
	if len(result.Errors) > 0 {
		for i, err := range result.Errors {
			if i >= 5 { // 最多記錄 5 個錯誤
				break
			}
			log.Warn().Err(err).Msg("Sync error")
		}
	}

	return nil
}

// HandleEmailSyncAllTask 處理同步所有使用者的任務
func HandleEmailSyncAllTask(ctx context.Context, t *asynq.Task, db *gorm.DB, client *asynq.Client) error {
	var payload EmailSyncAllPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.MaxAccounts == 0 {
		payload.MaxAccounts = 100 // 預設一次處理 100 個帳號
	}

	log.Info().
		Int("max_accounts", payload.MaxAccounts).
		Msg("Starting sync all users task")

	// 查詢所有可同步的 Gmail 帳號
	var oauthAccounts []models.OAuthAccount
	err := db.Where("provider = ? AND sync_status = ? AND deleted_at IS NULL",
		models.OAuthProviderGoogle,
		models.SyncStatusActive).
		Limit(payload.MaxAccounts).
		Find(&oauthAccounts).Error

	if err != nil {
		return fmt.Errorf("failed to query oauth accounts: %w", err)
	}

	log.Info().
		Int("accounts_found", len(oauthAccounts)).
		Msg("Found Gmail accounts to sync")

	// 為每個帳號建立同步任務
	successCount := 0
	errorCount := 0

	for _, account := range oauthAccounts {
		// 檢查是否可以同步（避免頻繁同步）
		syncService, err := gmail.NewSyncService(db, &account)
		if err != nil {
			log.Error().
				Err(err).
				Str("oauth_account_id", account.ID.String()).
				Msg("Failed to create sync service")
			errorCount++
			continue
		}

		canSync, remaining, err := syncService.CanSync(5) // 5 分鐘冷卻
		if err != nil {
			log.Error().Err(err).Msg("Failed to check sync cooldown")
			errorCount++
			continue
		}

		if !canSync {
			log.Debug().
				Str("oauth_account_id", account.ID.String()).
				Float64("remaining_seconds", remaining.Seconds()).
				Msg("Skipping account (in cooldown)")
			continue
		}

		// 建立同步任務
		syncType := "incremental"
		if account.LastSyncAt == nil {
			syncType = "initial"
		}

		task, err := NewEmailSyncTask(account.ID.String(), syncType)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create sync task")
			errorCount++
			continue
		}

		// 加入任務佇列
		if _, err := client.Enqueue(task); err != nil {
			log.Error().
				Err(err).
				Str("oauth_account_id", account.ID.String()).
				Msg("Failed to enqueue sync task")
			errorCount++
			continue
		}

		successCount++
	}

	log.Info().
		Int("success", successCount).
		Int("errors", errorCount).
		Int("total", len(oauthAccounts)).
		Msg("Sync all task completed")

	return nil
}
