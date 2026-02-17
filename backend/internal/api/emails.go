package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/services/gmail"
	"github.com/designcomb/influenter-backend/internal/services/openai"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// EmailHandler 郵件處理器
type EmailHandler struct {
	db            *gorm.DB
	openaiService *openai.Service
}

// NewEmailHandler 建立新的郵件處理器
func NewEmailHandler(db *gorm.DB, openaiService *openai.Service) *EmailHandler {
	return &EmailHandler{
		db:            db,
		openaiService: openaiService,
	}
}

// ListEmails 取得郵件列表
// @Summary      取得郵件列表
// @Description  取得使用者的郵件列表，支援分頁、篩選、搜尋
// @Tags         郵件
// @Produce      json
// @Security     BearerAuth
// @Param        oauth_account_id  query     string  false  "OAuth 帳號 ID"
// @Param        is_read           query     bool    false  "是否已讀"
// @Param        case_id           query     string  false  "案件 ID"
// @Param        from_email        query     string  false  "寄件者 email"
// @Param        subject           query     string  false  "主旨關鍵字"
// @Param        start_date        query     string  false  "開始日期 (RFC3339)"
// @Param        end_date          query     string  false  "結束日期 (RFC3339)"
// @Param        page              query     int     false  "頁數" default(1)
// @Param        page_size         query     int     false  "每頁數量" default(20)
// @Param        sort_by           query     string  false  "排序欄位" default(received_at)
// @Param        sort_order        query     string  false  "排序方向 (asc/desc)" default(desc)
// @Success      200  {object}  map[string]interface{}  "郵件列表和分頁資訊"
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /emails [get]
func (h *EmailHandler) ListEmails(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")

	// 解析查詢參數
	var params models.EmailQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_params",
			Message: err.Error(),
		})
		return
	}

	// 設定預設值
	params.SetDefaults()

	// 建立查詢
	query := h.db.Model(&models.Email{}).
		Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
		Where("oauth_accounts.user_id = ?", userID)

	// 應用篩選條件
	if params.OAuthAccountID != nil {
		query = query.Where("emails.oauth_account_id = ?", *params.OAuthAccountID)
	}

	if params.IsRead != nil {
		query = query.Where("emails.is_read = ?", *params.IsRead)
	}

	if params.CaseID != nil {
		query = query.Where("emails.case_id = ?", *params.CaseID)
	}

	if params.FromEmail != "" {
		query = query.Where("emails.from_email ILIKE ?", "%"+params.FromEmail+"%")
	}

	if params.Subject != "" {
		query = query.Where("emails.subject ILIKE ?", "%"+params.Subject+"%")
	}

	if params.StartDate != nil {
		query = query.Where("emails.received_at >= ?", params.StartDate)
	}

	if params.EndDate != nil {
		query = query.Where("emails.received_at <= ?", params.EndDate)
	}

	// 計算總數
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to count emails")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to count emails",
		})
		return
	}

	// 排序
	orderClause := params.SortBy + " " + params.SortOrder
	query = query.Order(orderClause)

	// 分頁
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	// 執行查詢
	var emails []models.Email
	if err := query.Find(&emails).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to fetch emails")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch emails",
		})
		return
	}

	// 轉換為列表回應格式
	emailsResponse := make([]models.EmailListResponse, 0, len(emails))
	for _, email := range emails {
		emailsResponse = append(emailsResponse, email.ToListResponse())
	}

	// 計算分頁資訊
	totalPages := (int(total) + params.PageSize - 1) / params.PageSize

	c.JSON(http.StatusOK, gin.H{
		"emails": emailsResponse,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetEmail 取得郵件詳情
// @Summary      取得郵件詳情
// @Description  取得單封郵件的完整內容
// @Tags         郵件
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "郵件 ID"
// @Success      200  {object}  models.EmailDetailResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /emails/{id} [get]
func (h *EmailHandler) GetEmail(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	emailID := c.Param("id")

	// 解析 UUID
	id, err := uuid.Parse(emailID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid email ID",
		})
		return
	}

	// 查詢郵件（確保屬於當前使用者）
	var email models.Email
	err = h.db.Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
		Where("emails.id = ? AND oauth_accounts.user_id = ?", id, userID).
		First(&email).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "email_not_found",
				Message: "Email not found",
			})
			return
		}

		logger.Error().Err(err).Str("email_id", emailID).Msg("Failed to fetch email")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch email",
		})
		return
	}

	c.JSON(http.StatusOK, email.ToDetailResponse())
}

// UpdateEmail 更新郵件
// @Summary      更新郵件
// @Description  更新郵件狀態（標記已讀、關聯案件等）
// @Tags         郵件
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                    true  "郵件 ID"
// @Param        updates body      UpdateEmailRequest        true  "更新內容"
// @Success      200     {object}  models.EmailDetailResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      401     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /emails/{id} [patch]
func (h *EmailHandler) UpdateEmail(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	emailID := c.Param("id")

	// 解析 UUID
	id, err := uuid.Parse(emailID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid email ID",
		})
		return
	}

	// 解析請求
	var req UpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// 查詢郵件（確保屬於當前使用者）
	var email models.Email
	err = h.db.Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
		Where("emails.id = ? AND oauth_accounts.user_id = ?", id, userID).
		First(&email).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "email_not_found",
				Message: "Email not found",
			})
			return
		}

		logger.Error().Err(err).Str("email_id", emailID).Msg("Failed to fetch email")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch email",
		})
		return
	}

	// 建立更新 map
	updates := make(map[string]interface{})

	if req.IsRead != nil {
		updates["is_read"] = *req.IsRead
	}

	if req.CaseID != nil {
		updates["case_id"] = *req.CaseID
	}

	// 執行更新
	if len(updates) > 0 {
		if err := h.db.Model(&email).Updates(updates).Error; err != nil {
			logger.Error().Err(err).Str("email_id", emailID).Msg("Failed to update email")
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to update email",
			})
			return
		}
	}

	// 重新查詢以取得最新資料
	if err := h.db.First(&email, id).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to fetch updated email")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch updated email",
		})
		return
	}

	logger.Info().
		Str("email_id", emailID).
		Interface("updates", updates).
		Msg("Email updated successfully")

	c.JSON(http.StatusOK, email.ToDetailResponse())
}

// SendReplyRequest 寄出回信請求
type SendReplyRequest struct {
	Body string `json:"body" binding:"required"`
}

// SendReply 寄出回信（透過 Gmail API）
func (h *EmailHandler) SendReply(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	emailID := c.Param("id")

	id, err := uuid.Parse(emailID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid email ID"})
		return
	}

	var body SendReplyRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "請提供回信內容 (body)"})
		return
	}

	var email models.Email
	err = h.db.Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
		Where("emails.id = ? AND oauth_accounts.user_id = ?", id, userID).
		First(&email).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "email_not_found", Message: "Email not found"})
			return
		}
		logger.Error().Err(err).Str("email_id", emailID).Msg("Failed to fetch email")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch email"})
		return
	}

	var oauthAccount models.OAuthAccount
	if err := h.db.Where("id = ?", email.OAuthAccountID).First(&oauthAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "oauth_not_found", Message: "Gmail 帳號不存在"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch oauth account")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to send reply"})
		return
	}

	gmailSvc, err := gmail.NewService(h.db, &oauthAccount)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create Gmail service")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "gmail_error", Message: "無法連接 Gmail，請確認已授權"})
		return
	}

	subject := "Re: "
	if email.Subject != nil {
		subj := strings.TrimSpace(*email.Subject)
		if !strings.HasPrefix(strings.ToUpper(subj), "RE:") {
			subject += subj
		} else {
			subject = subj
		}
	}

	threadID := ""
	if email.ThreadID != nil {
		threadID = *email.ThreadID
	}

	req := &gmail.SendMessageRequest{
		To:       []string{email.FromEmail},
		Subject:  subject,
		TextBody: body.Body,
		ThreadID: threadID,
	}

	sentID, err := gmailSvc.SendMessage(req)
	if err != nil {
		logger.Error().Err(err).Str("email_id", emailID).Msg("Failed to send reply")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "send_failed", Message: "寄出失敗，請稍後再試"})
		return
	}

	logger.Info().Str("email_id", emailID).Str("sent_id", sentID).Msg("Reply sent successfully")
	c.JSON(http.StatusOK, gin.H{"message_id": sentID, "message": "回信已寄出"})
}

// UpdateEmailRequest 更新郵件請求
type UpdateEmailRequest struct {
	IsRead *bool      `json:"is_read"`
	CaseID *uuid.UUID `json:"case_id"`
}

// debugLogWrite 寫入一行 NDJSON 到 .cursor/debug.log（僅除錯用）
func debugLogWrite(hypothesisId, location, message string, data map[string]interface{}) {
	dir, _ := os.Getwd()
	logPath := filepath.Join(dir, ".cursor", "debug.log")
	if filepath.Base(dir) == "backend" {
		logPath = filepath.Join(dir, "..", ".cursor", "debug.log")
	}
	line, _ := json.Marshal(map[string]interface{}{
		"hypothesisId": hypothesisId, "location": location, "message": message, "data": data, "timestamp": time.Now().UnixMilli(),
	})
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	f.Write(append(line, '\n'))
	f.Close()
}

// CreateCaseFromEmail 由郵件經 AI 分析後建立案件並關聯
// 立即回傳 202 Accepted，在背景執行 AI 分析與建立，避免連線逾時導致前端收到 ERR_EMPTY_RESPONSE
func (h *EmailHandler) CreateCaseFromEmail(c *gin.Context) {
	// #region agent log
	debugLogWrite("H1", "emails.go:CreateCaseFromEmail", "handler entered", map[string]interface{}{"email_id": c.Param("id"), "method": c.Request.Method})
	// #endregion
	logger := middleware.GetLogger(c)
	defer func() {
		if r := recover(); r != nil {
			logger.Error().Interface("panic", r).Msg("CreateCaseFromEmail panicked")
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Request failed unexpectedly",
			})
		}
	}()

	userID := c.GetString("user_id")
	emailID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "user_id required"})
		return
	}
	if h.openaiService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "ai_unavailable", Message: "AI analysis is not configured"})
		return
	}

	id, err := uuid.Parse(emailID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid email ID"})
		return
	}

	var email models.Email
	err = h.db.Where("emails.id = ?", id).
		Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
		Where("oauth_accounts.user_id = ?", userID).
		First(&email).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "email_not_found", Message: "Email not found"})
			return
		}
		logger.Error().Err(err).Str("email_id", emailID).Msg("Failed to fetch email")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch email"})
		return
	}

	// #region agent log
	debugLogWrite("H2", "emails.go:CreateCaseFromEmail", "about to send 202", map[string]interface{}{"email_id": emailID})
	// #endregion
	// 立即回傳 202，在背景執行，避免長時間等待導致連線被關閉
	c.JSON(http.StatusAccepted, gin.H{
		"status":   "processing",
		"email_id": emailID,
		"message":  "Case creation started. Poll GET /emails/:id for case_id.",
	})
	// 強制送出回應，避免 proxy/緩衝導致前端收到空回應
	if f, ok := c.Writer.(interface{ Flush() }); ok {
		f.Flush()
	}
	// #region agent log
	debugLogWrite("H2", "emails.go:CreateCaseFromEmail", "202 sent and flush done", map[string]interface{}{"email_id": emailID})
	// #endregion

	go h.runCreateCaseFromEmail(context.Background(), logger, userID, emailID, &email)
}

// runCreateCaseFromEmail 在背景執行 AI 分析並建立案件（由 CreateCaseFromEmail 呼叫）
func (h *EmailHandler) runCreateCaseFromEmail(ctx context.Context, logger *zerolog.Logger, userID, emailID string, email *models.Email) {
	subject := ""
	if email.Subject != nil {
		subject = *email.Subject
	}
	body := emailBodyForAnalysis(email)
	from := email.FromEmail

	req := openai.AnalyzeEmailRequest{
		Subject: subject,
		Body:    body,
		From:    from,
		To:      nil,
		Date:    email.ReceivedAt,
		Options: openai.AnalysisOptions{DetailLevel: "standard"},
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	result, err := h.openaiService.AnalyzeEmail(ctx, req)
	if err != nil {
		logger.Error().Err(err).Str("email_id", emailID).Msg("OpenAI analysis failed")
		return
	}

	userUUID, _ := uuid.Parse(userID)
	cs := analysisResultToCase(userUUID, subject, result)

	if err := h.db.Create(cs).Error; err != nil {
		logger.Error().Err(err).Str("email_id", emailID).Msg("Failed to create case")
		return
	}

	updates := map[string]interface{}{
		"case_id":     cs.ID,
		"ai_analyzed": true,
	}
	if err := h.db.Model(email).Updates(updates).Error; err != nil {
		logger.Error().Err(err).Str("email_id", emailID).Msg("Failed to link email to case")
		return
	}

	logger.Info().
		Str("email_id", emailID).
		Str("case_id", cs.ID.String()).
		Msg("Case created from email")
}

// emailBodyForAnalysis 取得用於 AI 分析的郵件內文（純文字）
func emailBodyForAnalysis(e *models.Email) string {
	if e.BodyText != nil && strings.TrimSpace(*e.BodyText) != "" {
		return *e.BodyText
	}
	if e.Snippet != nil && *e.Snippet != "" {
		return *e.Snippet
	}
	if e.BodyHTML != nil && *e.BodyHTML != "" {
		return stripHTML(*e.BodyHTML)
	}
	return ""
}

var htmlTagRe = regexp.MustCompile(`(?s)<[^>]*>`)

func stripHTML(s string) string {
	return strings.TrimSpace(htmlTagRe.ReplaceAllString(s, " "))
}

// analysisResultToCase 將 AI 分析結果轉為 Case 模型
func analysisResultToCase(userID uuid.UUID, subject string, result *openai.EmailAnalysisResult) *models.Case {
	// 判斷是否為合作相關案件
	isCollaboration := result.Classification.Category != "" && openai.IsCollaborationRelated(result.Classification.Category)

	var title, brandName string
	var status models.CaseStatus

	if isCollaboration {
		title = subject
		if title == "" {
			title = result.Summary
		}
		if title == "" {
			title = "未命名案件"
		}
		brandName = result.ExtractedInfo.BrandName
		if brandName == "" {
			brandName = "未知品牌"
		}
		status = models.CaseStatusToConfirm
	} else {
		// 非合作案件：標題用郵件主旨，不填品牌/報價/截止日等
		title = subject
		if title == "" {
			title = "未命名案件"
		}
		brandName = "—" // DB 必填，用佔位符
		status = models.CaseStatusOther
	}

	cs := &models.Case{
		UserID:    userID,
		Title:     title,
		BrandName: brandName,
		Status:    status,
	}

	if isCollaboration {
		if result.ExtractedInfo.ContentType != "" {
			cs.CollaborationType = &result.ExtractedInfo.ContentType
		}
		cs.QuotedAmount = result.ExtractedInfo.Amount
		if result.ExtractedInfo.Currency != "" {
			cs.Currency = &result.ExtractedInfo.Currency
		}
		cs.DeadlineDate = result.ExtractedInfo.DueDate
		if result.ExtractedInfo.ContactName != "" {
			cs.ContactName = &result.ExtractedInfo.ContactName
		}
		if result.ExtractedInfo.ContactEmail != "" {
			cs.ContactEmail = &result.ExtractedInfo.ContactEmail
		}
		if result.ExtractedInfo.ContactPhone != "" {
			cs.ContactPhone = &result.ExtractedInfo.ContactPhone
		}
		if result.Summary != "" {
			cs.Description = &result.Summary
		} else if result.ExtractedInfo.ProjectDetails != "" {
			cs.Description = &result.ExtractedInfo.ProjectDetails
		}
	} else if result.Summary != "" {
		// 非合作案件僅保留摘要作為說明
		cs.Description = &result.Summary
	}
	return cs
}
