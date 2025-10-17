package api

import (
	"net/http"

	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailHandler 郵件處理器
type EmailHandler struct {
	db *gorm.DB
}

// NewEmailHandler 建立新的郵件處理器
func NewEmailHandler(db *gorm.DB) *EmailHandler {
	return &EmailHandler{
		db: db,
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

// UpdateEmailRequest 更新郵件請求
type UpdateEmailRequest struct {
	IsRead *bool      `json:"is_read"`
	CaseID *uuid.UUID `json:"case_id"`
}
