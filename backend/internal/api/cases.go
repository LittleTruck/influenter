package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/services/openai"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CaseHandler 案件處理器
type CaseHandler struct {
	db            *gorm.DB
	openaiService *openai.Service
}

// NewCaseHandler 建立新的案件處理器
func NewCaseHandler(db *gorm.DB, openaiSvc *openai.Service) *CaseHandler {
	return &CaseHandler{db: db, openaiService: openaiSvc}
}

// CreateCaseRequest 建立案件請求（與前端 CreateCaseRequest 對齊）
type CreateCaseRequest struct {
	Title             string   `json:"title" binding:"required"`
	BrandName         string   `json:"brand_name" binding:"required"`
	CollaborationType *string  `json:"collaboration_type"`
	Description       *string  `json:"description"`
	QuotedAmount     *float64 `json:"quoted_amount"`
	DeadlineDate     *string  `json:"deadline_date"` // ISO date string
	ContactName      *string  `json:"contact_name"`
	ContactEmail     *string  `json:"contact_email"`
	ContactPhone     *string  `json:"contact_phone"`
	Notes            *string  `json:"notes"`
	Tags             []string `json:"tags"`
	CollaborationItems []string `json:"collaboration_items"`
}

// CaseResponse 案件回應（與前端 Case 對齊）
type CaseResponse struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	BrandName         string   `json:"brand_name"`
	CollaborationType *string  `json:"collaboration_type,omitempty"`
	Status            string   `json:"status"`
	QuotedAmount      *float64 `json:"quoted_amount,omitempty"`
	FinalAmount       *float64 `json:"final_amount,omitempty"`
	Currency          *string  `json:"currency,omitempty"`
	DeadlineDate      *string  `json:"deadline_date,omitempty"`
	ContactName       *string  `json:"contact_name,omitempty"`
	ContactEmail      *string  `json:"contact_email,omitempty"`
	ContactPhone      *string  `json:"contact_phone,omitempty"`
	EmailCount        int      `json:"email_count"`
	TaskCount         int      `json:"task_count"`
	CompletedTaskCount int     `json:"completed_task_count"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// caseToResponse 將 Case 轉為 API 回應
func caseToResponse(c *models.Case, emailCount, taskCount, completedTaskCount int) CaseResponse {
	resp := CaseResponse{
		ID:                 c.ID.String(),
		Title:              c.Title,
		BrandName:          c.BrandName,
		CollaborationType:  c.CollaborationType,
		Status:             string(c.Status),
		QuotedAmount:       c.QuotedAmount,
		FinalAmount:        c.FinalAmount,
		Currency:           c.Currency,
		ContactName:        c.ContactName,
		ContactEmail:       c.ContactEmail,
		ContactPhone:       c.ContactPhone,
		EmailCount:         emailCount,
		TaskCount:          taskCount,
		CompletedTaskCount: completedTaskCount,
		CreatedAt:          c.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:          c.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
	if c.DeadlineDate != nil {
		s := c.DeadlineDate.Format("2006-01-02")
		resp.DeadlineDate = &s
	}
	return resp
}

// CreateCase 建立案件
func (h *CaseHandler) CreateCase(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "user_id required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_user_id", Message: "Invalid user ID"})
		return
	}

	var req CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	var deadlineDate *time.Time
	if req.DeadlineDate != nil && *req.DeadlineDate != "" {
		if t, err := time.Parse("2006-01-02", *req.DeadlineDate); err == nil {
			deadlineDate = &t
		}
	}

	cs := models.Case{
		UserID:            userID,
		Title:             req.Title,
		BrandName:         req.BrandName,
		Status:             models.CaseStatusToConfirm,
		CollaborationType:  req.CollaborationType,
		Description:       req.Description,
		QuotedAmount:      req.QuotedAmount,
		DeadlineDate:      deadlineDate,
		ContactName:       req.ContactName,
		ContactEmail:      req.ContactEmail,
		ContactPhone:      req.ContactPhone,
		Notes:             req.Notes,
	}
	if len(req.Tags) > 0 {
		cs.Tags = req.Tags
	}
	if len(req.CollaborationItems) > 0 {
		cs.CollaborationItems = req.CollaborationItems
	}

	if err := h.db.Create(&cs).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to create case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to create case"})
		return
	}

	logger.Info().Str("case_id", cs.ID.String()).Str("user_id", userIDStr).Msg("Case created")
	c.JSON(http.StatusCreated, caseToResponse(&cs, 0, 0, 0))
}

// ListCases 取得案件列表
func (h *CaseHandler) ListCases(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "user_id required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	status := c.Query("status")
	// sort: optional, e.g. updated_at_desc

	query := h.db.Model(&models.Case{}).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to count cases")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to list cases"})
		return
	}

	offset := (page - 1) * perPage
	query = query.Order("updated_at DESC").Offset(offset).Limit(perPage)

	var cases []models.Case
	if err := query.Find(&cases).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to list cases")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to list cases"})
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	data := make([]CaseResponse, 0, len(cases))
	for i := range cases {
		data = append(data, caseToResponse(&cases[i], 0, 0, 0))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"pagination": gin.H{
			"page":        page,
			"per_page":   perPage,
			"total":      total,
			"total_pages": totalPages,
		},
	})
}

// CaseFieldOption 屬性選項（select / multiselect）
type CaseFieldOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
	Color string      `json:"color,omitempty"`
}

// CaseFieldResponse 單一屬性（與前端 CaseField 對齊）
type CaseFieldResponse struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Label             string             `json:"label"`
	Type              string             `json:"type"`
	IsSystem          bool               `json:"is_system"`
	SystemColumnName  string             `json:"system_column_name,omitempty"`
	IsRequired        bool               `json:"is_required"`
	IsVisible         bool               `json:"is_visible"`
	Order             int                `json:"order"`
	Placeholder       string             `json:"placeholder,omitempty"`
	Options           []CaseFieldOption  `json:"options,omitempty"`
	CreatedAt         string             `json:"created_at,omitempty"`
	UpdatedAt         string             `json:"updated_at,omitempty"`
}

// CaseFieldsListResponse 屬性列表回應（與前端 FieldListResponse 對齊）
type CaseFieldsListResponse struct {
	SystemFields []CaseFieldResponse `json:"system_fields"`
	CustomFields []CaseFieldResponse `json:"custom_fields"`
}

// defaultSystemFields 與前端 defaultSystemFields 一致的預設系統屬性
var defaultSystemFields = []CaseFieldResponse{
	{ID: "system-title", Name: "title", Label: "案件標題", Type: "text", IsSystem: true, SystemColumnName: "title", IsRequired: true, IsVisible: true, Order: 1, Placeholder: "例如：Nike 球鞋業配"},
	{ID: "system-brand_name", Name: "brand_name", Label: "品牌名稱", Type: "text", IsSystem: true, SystemColumnName: "brand_name", IsRequired: true, IsVisible: true, Order: 2, Placeholder: "例如：Nike"},
	{ID: "system-status", Name: "status", Label: "案件狀態", Type: "select", IsSystem: true, SystemColumnName: "status", IsRequired: true, IsVisible: true, Order: 3, Options: []CaseFieldOption{
		{Label: "待確認", Value: "to_confirm"}, {Label: "進行中", Value: "in_progress"}, {Label: "已完成", Value: "completed"}, {Label: "已取消", Value: "cancelled"}, {Label: "非合作案件", Value: "other"},
	}},
	{ID: "system-deadline_date", Name: "deadline_date", Label: "截止日期", Type: "date", IsSystem: true, SystemColumnName: "deadline_date", IsRequired: false, IsVisible: true, Order: 4},
	{ID: "system-quoted_amount", Name: "quoted_amount", Label: "預估報價", Type: "number", IsSystem: true, SystemColumnName: "quoted_amount", IsRequired: false, IsVisible: true, Order: 5},
}

// ListCaseFields 取得案件屬性列表（系統屬性 + 自定義屬性）
// 目前僅回傳預設系統屬性，自定義屬性尚未持久化。
func (h *CaseHandler) ListCaseFields(c *gin.Context) {
	c.JSON(http.StatusOK, CaseFieldsListResponse{
		SystemFields: defaultSystemFields,
		CustomFields: []CaseFieldResponse{},
	})
}

// GetCase 取得案件詳情
func (h *CaseHandler) GetCase(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")

	id, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}

	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "case_not_found", Message: "Case not found"})
			return
		}
		logger.Error().Err(err).Str("case_id", caseID).Msg("Failed to fetch case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case"})
		return
	}

	var emailCount int64
	h.db.Model(&models.Email{}).Where("case_id = ?", id).Count(&emailCount)

	// 查詢案件階段
	var phases []models.CasePhase
	h.db.Where("case_id = ?", id).Order(`"order" ASC`).Find(&phases)

	resp := caseToResponse(&cs, int(emailCount), 0, 0)
	phaseList := make([]CasePhaseResponse, 0, len(phases))
	for _, p := range phases {
		phaseList = append(phaseList, casePhaseToResponse(&p))
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                  resp.ID,
		"title":               resp.Title,
		"brand_name":          resp.BrandName,
		"collaboration_type":  resp.CollaborationType,
		"status":              resp.Status,
		"quoted_amount":       resp.QuotedAmount,
		"final_amount":        resp.FinalAmount,
		"currency":            resp.Currency,
		"deadline_date":       resp.DeadlineDate,
		"contact_name":        resp.ContactName,
		"contact_email":       resp.ContactEmail,
		"contact_phone":       resp.ContactPhone,
		"email_count":         resp.EmailCount,
		"task_count":          resp.TaskCount,
		"completed_task_count": resp.CompletedTaskCount,
		"created_at":          resp.CreatedAt,
		"updated_at":          resp.UpdatedAt,
		"phases":              phaseList,
	})
}

// CaseEmailResponse 案件郵件列表項目（與前端 CaseEmail 對齊）
type CaseEmailResponse struct {
	ID         string  `json:"id"`
	Direction  string  `json:"direction"` // incoming | outgoing
	Subject    *string `json:"subject,omitempty"`
	FromEmail  string  `json:"from_email"`
	FromName   *string `json:"from_name,omitempty"`
	ToEmail    *string `json:"to_email,omitempty"` // 寄出時為收件者
	ReceivedAt string  `json:"received_at"`
	EmailType  string  `json:"email_type,omitempty"`
}

// DraftReplyRequest 擬回信 API 請求 body
type DraftReplyRequest struct {
	EmailID    string `json:"email_id"`    // 要回覆的郵件 ID，可選；未傳則用該案件最新一封
	Instruction string `json:"instruction"` // 使用者補充說明，可選
}

// ListCaseEmails 取得案件關聯的郵件列表
func (h *CaseHandler) ListCaseEmails(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")

	id, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}

	// 確認案件屬於當前使用者
	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "case_not_found", Message: "Case not found"})
			return
		}
		logger.Error().Err(err).Str("case_id", caseID).Msg("Failed to fetch case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case"})
		return
	}

	var emails []models.Email
	err = h.db.Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
		Where("emails.case_id = ? AND oauth_accounts.user_id = ?", id, userID).
		Order("emails.received_at ASC").
		Find(&emails).Error
	if err != nil {
		logger.Error().Err(err).Str("case_id", caseID).Msg("Failed to list case emails")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to list case emails"})
		return
	}

	data := make([]CaseEmailResponse, 0, len(emails))
	for i := range emails {
		e := &emails[i]
		dir := e.Direction
		if dir == "" {
			dir = "incoming"
		}
		emailType := ""
		if dir == "outgoing" {
			emailType = "sent"
		}
		data = append(data, CaseEmailResponse{
			ID:         e.ID.String(),
			Direction:  dir,
			Subject:    e.Subject,
			FromEmail:  e.FromEmail,
			FromName:   e.FromName,
			ToEmail:    e.ToEmail,
			ReceivedAt: e.ReceivedAt.Format("2006-01-02T15:04:05.000Z07:00"),
			EmailType:  emailType,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// DraftReply 產生 AI 擬回信草稿
func (h *CaseHandler) DraftReply(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")

	if h.openaiService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "openai_unavailable", Message: "AI 服務未設定"})
		return
	}

	id, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}

	var body DraftReplyRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "case_not_found", Message: "Case not found"})
			return
		}
		logger.Error().Err(err).Str("case_id", caseID).Msg("Failed to fetch case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case"})
		return
	}

	var email models.Email
	if body.EmailID != "" {
		emailUUID, err := uuid.Parse(body.EmailID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_email_id", Message: "Invalid email ID"})
			return
		}
		err = h.db.Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
			Where("emails.id = ? AND emails.case_id = ? AND oauth_accounts.user_id = ?", emailUUID, id, userID).
			First(&email).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "email_not_found", Message: "該郵件不存在或未關聯此案件"})
				return
			}
			logger.Error().Err(err).Str("email_id", body.EmailID).Msg("Failed to fetch email")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch email"})
			return
		}
	} else {
		err = h.db.Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
			Where("emails.case_id = ? AND oauth_accounts.user_id = ?", id, userID).
			Order("emails.received_at DESC").
			First(&email).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no_emails", Message: "此案件尚無關聯郵件，無法擬信"})
				return
			}
			logger.Error().Err(err).Str("case_id", caseID).Msg("Failed to fetch latest email")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch email"})
			return
		}
	}

	bodyText := ""
	if email.BodyText != nil && *email.BodyText != "" {
		bodyText = *email.BodyText
	} else if email.Snippet != nil {
		bodyText = *email.Snippet
	}
	subject := ""
	if email.Subject != nil {
		subject = *email.Subject
	}
	fromName := email.FromEmail
	if email.FromName != nil && *email.FromName != "" {
		fromName = *email.FromName + " <" + email.FromEmail + ">"
	}
	contactName := ""
	if cs.ContactName != nil {
		contactName = *cs.ContactName
	}
	contactEmail := ""
	if cs.ContactEmail != nil {
		contactEmail = *cs.ContactEmail
	}

	// 取得使用者 AI 注意事項
	userAIInstructions := ""
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err == nil && user.AIInstructions != nil {
		userAIInstructions = *user.AIInstructions
	}

	req := openai.DraftReplyRequest{
		CaseTitle:          cs.Title,
		BrandName:          cs.BrandName,
		ContactName:        contactName,
		ContactEmail:       contactEmail,
		EmailFrom:          fromName,
		EmailSubject:       subject,
		EmailBody:          bodyText,
		Instruction:        body.Instruction,
		UserAIInstructions: userAIInstructions,
	}

	result, err := h.openaiService.DraftReply(c.Request.Context(), req)
	if err != nil {
		logger.Error().Err(err).Str("case_id", caseID).Msg("DraftReply failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "draft_failed", Message: "產生草稿失敗，請稍後再試"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"draft": result.Draft})
}

// --- Case Phase types and handlers ---

// CasePhaseResponse 案件階段回應
type CasePhaseResponse struct {
	ID              string  `json:"id"`
	CaseID          string  `json:"case_id"`
	Name            string  `json:"name"`
	StartDate       *string `json:"start_date"`
	EndDate         *string `json:"end_date"`
	DurationDays    int     `json:"duration_days"`
	Order           int     `json:"order"`
	WorkflowPhaseID *string `json:"workflow_phase_id,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func casePhaseToResponse(p *models.CasePhase) CasePhaseResponse {
	resp := CasePhaseResponse{
		ID:           p.ID.String(),
		CaseID:       p.CaseID.String(),
		Name:         p.Name,
		DurationDays: p.DurationDays,
		Order:        p.Order,
		CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:    p.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
	if p.StartDate != nil {
		s := p.StartDate.Format("2006-01-02")
		resp.StartDate = &s
	}
	if p.EndDate != nil {
		s := p.EndDate.Format("2006-01-02")
		resp.EndDate = &s
	}
	if p.WorkflowPhaseID != nil {
		s := p.WorkflowPhaseID.String()
		resp.WorkflowPhaseID = &s
	}
	return resp
}

// CreateCasePhaseRequest 建立案件階段請求
type CreateCasePhaseRequest struct {
	Name         string `json:"name" binding:"required"`
	StartDate    string `json:"start_date"`
	DurationDays int    `json:"duration_days"`
	Order        *int   `json:"order"`
}

// UpdateCasePhaseRequest 更新案件階段請求
type UpdateCasePhaseRequest struct {
	Name         *string `json:"name"`
	StartDate    *string `json:"start_date"`
	EndDate      *string `json:"end_date"`
	DurationDays *int    `json:"duration_days"`
	Order        *int    `json:"order"`
}

// ApplyTemplateRequest 套用流程範本請求
type ApplyTemplateRequest struct {
	WorkflowID string `json:"workflow_id" binding:"required"`
	StartDate  string `json:"start_date" binding:"required"`
}

// ListCasePhases 取得案件階段列表
func (h *CaseHandler) ListCasePhases(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")

	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}

	// 確認案件屬於當前使用者
	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", caseUUID, userID).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "case_not_found", Message: "Case not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case"})
		return
	}

	var phases []models.CasePhase
	if err := h.db.Where("case_id = ?", caseUUID).Order(`"order" ASC`).Find(&phases).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to list case phases")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to list case phases"})
		return
	}

	data := make([]CasePhaseResponse, 0, len(phases))
	for _, p := range phases {
		data = append(data, casePhaseToResponse(&p))
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// CreateCasePhase 建立案件階段
func (h *CaseHandler) CreateCasePhase(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")

	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}

	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", caseUUID, userID).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "case_not_found", Message: "Case not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case"})
		return
	}

	var req CreateCasePhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	durationDays := req.DurationDays
	if durationDays <= 0 {
		durationDays = 1
	}

	phase := models.CasePhase{
		CaseID:       caseUUID,
		Name:         req.Name,
		DurationDays: durationDays,
	}

	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			phase.StartDate = &t
			endDate := t.AddDate(0, 0, durationDays)
			phase.EndDate = &endDate
		}
	}

	if req.Order != nil {
		phase.Order = *req.Order
	} else {
		var maxOrder int
		h.db.Model(&models.CasePhase{}).Where("case_id = ?", caseUUID).
			Select(`COALESCE(MAX("order"), -1)`).Scan(&maxOrder)
		phase.Order = maxOrder + 1
	}

	if err := h.db.Create(&phase).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to create case phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to create case phase"})
		return
	}

	c.JSON(http.StatusCreated, casePhaseToResponse(&phase))
}

// ApplyTemplate 套用流程範本到案件
func (h *CaseHandler) ApplyTemplate(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")

	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}

	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", caseUUID, userID).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "case_not_found", Message: "Case not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case"})
		return
	}

	var req ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	workflowUUID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_workflow_id", Message: "Invalid workflow ID"})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_start_date", Message: "Invalid start date format (YYYY-MM-DD)"})
		return
	}

	// 讀取流程範本及階段
	var tmpl models.WorkflowTemplate
	if err := h.db.Where("id = ? AND user_id = ?", workflowUUID, userID).
		Preload("Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "workflow_not_found", Message: "Workflow template not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch workflow template"})
		return
	}

	if len(tmpl.Phases) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no_phases", Message: "Workflow template has no phases"})
		return
	}

	tx := h.db.Begin()

	// 刪除既有的案件階段
	if err := tx.Where("case_id = ?", caseUUID).Delete(&models.CasePhase{}).Error; err != nil {
		tx.Rollback()
		logger.Error().Err(err).Msg("Failed to delete existing case phases")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to apply template"})
		return
	}

	// 從開始日期計算每個階段的日期
	currentDate := startDate
	var createdPhases []models.CasePhase

	for i, wp := range tmpl.Phases {
		endDate := currentDate.AddDate(0, 0, wp.DurationDays)
		wpID := wp.ID
		phase := models.CasePhase{
			CaseID:          caseUUID,
			Name:            wp.Name,
			StartDate:       &currentDate,
			EndDate:         &endDate,
			DurationDays:    wp.DurationDays,
			Order:           i,
			WorkflowPhaseID: &wpID,
		}
		if err := tx.Create(&phase).Error; err != nil {
			tx.Rollback()
			logger.Error().Err(err).Msg("Failed to create case phase")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: fmt.Sprintf("Failed to create phase: %s", wp.Name)})
			return
		}
		createdPhases = append(createdPhases, phase)
		currentDate = endDate
	}

	tx.Commit()

	data := make([]CasePhaseResponse, 0, len(createdPhases))
	for _, p := range createdPhases {
		data = append(data, casePhaseToResponse(&p))
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "message": fmt.Sprintf("Applied %d phases from template '%s'", len(createdPhases), tmpl.Name)})
}

// UpdateCasePhase 更新案件階段
func (h *CaseHandler) UpdateCasePhase(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")
	phaseID := c.Param("phaseId")

	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}
	phaseUUID, err := uuid.Parse(phaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid phase ID"})
		return
	}

	// 確認案件屬於當前使用者
	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", caseUUID, userID).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "case_not_found", Message: "Case not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case"})
		return
	}

	var phase models.CasePhase
	if err := h.db.Where("id = ? AND case_id = ?", phaseUUID, caseUUID).First(&phase).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "phase_not_found", Message: "Case phase not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch case phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case phase"})
		return
	}

	var req UpdateCasePhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.DurationDays != nil {
		updates["duration_days"] = *req.DurationDays
	}
	if req.Order != nil {
		updates["order"] = *req.Order
	}
	if req.StartDate != nil {
		if t, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			updates["start_date"] = t
			// 如果有 duration_days，自動計算 end_date
			dur := phase.DurationDays
			if req.DurationDays != nil {
				dur = *req.DurationDays
			}
			endDate := t.AddDate(0, 0, dur)
			updates["end_date"] = endDate
		}
	}
	if req.EndDate != nil {
		if t, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			updates["end_date"] = t
		}
	}

	if err := h.db.Model(&phase).Updates(updates).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to update case phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to update case phase"})
		return
	}

	h.db.First(&phase, "id = ?", phaseUUID)
	c.JSON(http.StatusOK, casePhaseToResponse(&phase))
}

// DeleteCasePhase 刪除案件階段
func (h *CaseHandler) DeleteCasePhase(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")
	phaseID := c.Param("phaseId")

	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}
	phaseUUID, err := uuid.Parse(phaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid phase ID"})
		return
	}

	// 確認案件屬於當前使用者
	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", caseUUID, userID).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "case_not_found", Message: "Case not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case"})
		return
	}

	var phase models.CasePhase
	if err := h.db.Where("id = ? AND case_id = ?", phaseUUID, caseUUID).First(&phase).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "phase_not_found", Message: "Case phase not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch case phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch case phase"})
		return
	}

	if err := h.db.Delete(&phase).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to delete case phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to delete case phase"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Case phase deleted"})
}
