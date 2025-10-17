package api

import (
	"context"
	"net/http"

	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/services/gmail"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GmailHandler Gmail 整合處理器
type GmailHandler struct {
	db *gorm.DB
}

// NewGmailHandler 建立新的 Gmail 處理器
func NewGmailHandler(db *gorm.DB) *GmailHandler {
	return &GmailHandler{
		db: db,
	}
}

// GetStatus 取得 Gmail 同步狀態
// @Summary      取得 Gmail 同步狀態
// @Description  取得使用者的 Gmail 帳號連接和同步狀態
// @Tags         Gmail
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /gmail/status [get]
func (h *GmailHandler) GetStatus(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")

	// 查詢使用者的 Google OAuth 帳號
	var oauthAccount models.OAuthAccount
	err := h.db.Where("user_id = ? AND provider = ?", userID, models.OAuthProviderGoogle).
		First(&oauthAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"connected": false,
				"message":   "Gmail account not connected",
			})
			return
		}

		logger.Error().Err(err).Msg("Failed to query oauth account")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to query gmail status",
		})
		return
	}

	// 查詢郵件統計
	syncService, err := gmail.NewSyncService(h.db, &oauthAccount)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create sync service")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "service_error",
			Message: "Failed to create sync service",
		})
		return
	}

	stats, err := syncService.GetSyncStats()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get sync stats")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to get statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":     true,
		"email":         oauthAccount.Email,
		"last_sync_at":  oauthAccount.LastSyncAt,
		"sync_status":   oauthAccount.SyncStatus,
		"sync_error":    oauthAccount.SyncError,
		"token_expired": oauthAccount.IsTokenExpired(),
		"can_sync":      oauthAccount.CanSync(),
		"stats":         stats,
	})
}

// TriggerSync 手動觸發同步
// @Summary      手動觸發 Gmail 同步
// @Description  手動觸發郵件同步（有冷卻時間限制）
// @Tags         Gmail
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      429  {object}  ErrorResponse  "同步冷卻中"
// @Failure      500  {object}  ErrorResponse
// @Router       /gmail/sync [post]
func (h *GmailHandler) TriggerSync(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")

	// 查詢使用者的 Google OAuth 帳號
	var oauthAccount models.OAuthAccount
	err := h.db.Where("user_id = ? AND provider = ?", userID, models.OAuthProviderGoogle).
		First(&oauthAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_connected",
				Message: "Gmail account not connected",
			})
			return
		}

		logger.Error().Err(err).Msg("Failed to query oauth account")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to query gmail account",
		})
		return
	}

	// 建立同步服務
	syncService, err := gmail.NewSyncService(h.db, &oauthAccount)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create sync service")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "service_error",
			Message: "Failed to create sync service",
		})
		return
	}

	// 檢查是否可以同步（1 分鐘冷卻時間）
	canSync, remaining, err := syncService.CanSync(1)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to check sync cooldown")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "service_error",
			Message: "Failed to check sync status",
		})
		return
	}

	if !canSync {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":     "sync_cooldown",
			"message":   "Please wait before syncing again",
			"remaining": remaining.Seconds(),
		})
		return
	}

	// 執行同步（使用 goroutine 以免阻塞 API）
	go func() {
		ctx := context.Background()

		// 首次同步或增量同步
		var syncResult *gmail.SyncResult
		var syncErr error

		if oauthAccount.LastSyncAt == nil {
			syncResult, syncErr = syncService.InitialSync(ctx)
		} else {
			syncResult, syncErr = syncService.IncrementalSync(ctx)
		}

		if syncErr != nil {
			logger.Error().Err(syncErr).Msg("Sync failed")
			return
		}

		logger.Info().
			Int("total", syncResult.TotalFetched).
			Int("new", syncResult.NewEmails).
			Int("updated", syncResult.UpdatedEmails).
			Int("errors", len(syncResult.Errors)).
			Msg("Sync completed")
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Sync started",
		"status":  "syncing",
	})
}

// DisconnectGmail 斷開 Gmail 連接
// @Summary      斷開 Gmail 連接
// @Description  刪除 Gmail OAuth 帳號連接（保留已同步的郵件）
// @Tags         Gmail
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /gmail/disconnect [delete]
func (h *GmailHandler) DisconnectGmail(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")

	// 查詢使用者的 Google OAuth 帳號
	var oauthAccount models.OAuthAccount
	err := h.db.Where("user_id = ? AND provider = ?", userID, models.OAuthProviderGoogle).
		First(&oauthAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_connected",
				Message: "Gmail account not connected",
			})
			return
		}

		logger.Error().Err(err).Msg("Failed to query oauth account")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to query gmail account",
		})
		return
	}

	// 軟刪除 OAuth 帳號（郵件會因為 CASCADE 規則一起刪除）
	// 注意：如果要保留郵件，需要修改外鍵為 SET NULL
	if err := h.db.Delete(&oauthAccount).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to disconnect gmail")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to disconnect gmail account",
		})
		return
	}

	logger.Info().
		Str("email", oauthAccount.Email).
		Msg("Gmail account disconnected")

	c.JSON(http.StatusOK, gin.H{
		"message": "Gmail account disconnected successfully",
	})
}
