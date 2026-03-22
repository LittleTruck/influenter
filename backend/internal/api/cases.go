package api

import (
	"context"
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
	CollaborationItems []string `json:"collaboration_items"` // Item IDs to link
	FlowLayout       *string  `json:"flow_layout"`          // "parallel" or "sequential"
}

// CaseResponse 案件回應（與前端 Case 對齊）
type CaseResponse struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	BrandName         string   `json:"brand_name"`
	CollaborationType *string  `json:"collaboration_type,omitempty"`
	Status            string   `json:"status"`
	FlowLayout        string   `json:"flow_layout"`
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
	flowLayout := c.FlowLayout
	if flowLayout == "" {
		flowLayout = "parallel"
	}
	resp := CaseResponse{
		ID:                 c.ID.String(),
		Title:              c.Title,
		BrandName:          c.BrandName,
		CollaborationType:  c.CollaborationType,
		Status:             string(c.Status),
		FlowLayout:         flowLayout,
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

	flowLayout := "parallel"
	if req.FlowLayout != nil && (*req.FlowLayout == "parallel" || *req.FlowLayout == "sequential") {
		flowLayout = *req.FlowLayout
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
		FlowLayout:        flowLayout,
	}
	if len(req.Tags) > 0 {
		cs.Tags = req.Tags
	}

	tx := h.db.Begin()

	if err := tx.Create(&cs).Error; err != nil {
		tx.Rollback()
		logger.Error().Err(err).Msg("Failed to create case")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to create case"})
		return
	}

	// Create case_collaboration_items associations
	if len(req.CollaborationItems) > 0 {
		for i, itemIDStr := range req.CollaborationItems {
			itemUUID, err := uuid.Parse(itemIDStr)
			if err != nil {
				continue
			}
			cci := models.CaseCollaborationItem{
				CaseID:              cs.ID,
				CollaborationItemID: itemUUID,
				Order:               i,
			}
			if err := tx.Create(&cci).Error; err != nil {
				logger.Warn().Err(err).Str("item_id", itemIDStr).Msg("Failed to link collaboration item")
			}
		}
	}

	tx.Commit()

	logger.Info().Str("case_id", cs.ID.String()).Str("user_id", userIDStr).Msg("Case created")

	// 背景觸發 AI 自動匹配合作項目（如果案件有關聯郵件）
	go h.autoMatchCollaborationItemsForCase(cs.ID, userID)

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

	// 查詢案件關聯的合作項目（多對多）
	var caseCollabItems []models.CaseCollaborationItem
	h.db.Where("case_id = ?", id).
		Preload("CollaborationItem").
		Preload("CollaborationItem.BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("CollaborationItem.BundleItems.Item").
		Preload("CollaborationItem.Workflow").
		Order(`"order" ASC`).
		Find(&caseCollabItems)

	resp := caseToResponse(&cs, int(emailCount), 0, 0)
	phaseList := make([]CasePhaseResponse, 0, len(phases))
	for _, p := range phases {
		phaseList = append(phaseList, casePhaseToResponse(&p))
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                      resp.ID,
		"title":                   resp.Title,
		"brand_name":              resp.BrandName,
		"collaboration_type":      resp.CollaborationType,
		"status":                  resp.Status,
		"flow_layout":             resp.FlowLayout,
		"quoted_amount":           resp.QuotedAmount,
		"final_amount":            resp.FinalAmount,
		"currency":                resp.Currency,
		"deadline_date":           resp.DeadlineDate,
		"contact_name":            resp.ContactName,
		"contact_email":           resp.ContactEmail,
		"contact_phone":           resp.ContactPhone,
		"email_count":             resp.EmailCount,
		"task_count":              resp.TaskCount,
		"completed_task_count":    resp.CompletedTaskCount,
		"created_at":              resp.CreatedAt,
		"updated_at":              resp.UpdatedAt,
		"phases":                  phaseList,
		"case_collaboration_items": caseCollabItems,
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
	EmailID      string `json:"email_id"`      // 要回覆的郵件 ID，可選；未傳則用該案件最新一封
	Instruction  string `json:"instruction"`   // 使用者補充說明，可選
	TemplateID   string `json:"template_id"`   // 回覆範本 ID，可選
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

	// 取得使用者 AI 助理設定
	userAIInstructions := ""
	userAIReplyHeader := ""
	userAIReplyFooter := ""
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err == nil {
		if user.AIInstructions != nil {
			userAIInstructions = *user.AIInstructions
		}
		if user.AIReplyHeader != nil {
			userAIReplyHeader = *user.AIReplyHeader
		}
		if user.AIReplyFooter != nil {
			userAIReplyFooter = *user.AIReplyFooter
		}
	}

	// 取得回覆範本（如有選擇）
	templatePrompt := ""
	if body.TemplateID != "" {
		tplID, err := uuid.Parse(body.TemplateID)
		if err == nil {
			var tpl models.ReplyTemplate
			if err := h.db.Where("id = ? AND user_id = ?", tplID, userID).First(&tpl).Error; err == nil {
				templatePrompt = tpl.Prompt
			}
		}
	}

	req := openai.DraftReplyRequest{
		CaseTitle:           cs.Title,
		BrandName:           cs.BrandName,
		ContactName:         contactName,
		ContactEmail:        contactEmail,
		EmailFrom:           fromName,
		EmailSubject:        subject,
		EmailBody:           bodyText,
		Instruction:         body.Instruction,
		UserAIInstructions:  userAIInstructions,
		UserAIReplyHeader:   userAIReplyHeader,
		UserAIReplyFooter:   userAIReplyFooter,
		TemplatePrompt:      templatePrompt,
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
	ID                  string  `json:"id"`
	CaseID              string  `json:"case_id"`
	Name                string  `json:"name"`
	StartDate           *string `json:"start_date"`
	EndDate             *string `json:"end_date"`
	DurationDays        int     `json:"duration_days"`
	Order               int     `json:"order"`
	WorkflowPhaseID     *string `json:"workflow_phase_id,omitempty"`
	CollaborationItemID *string `json:"collaboration_item_id,omitempty"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
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
	if p.CollaborationItemID != nil {
		s := p.CollaborationItemID.String()
		resp.CollaborationItemID = &s
	}
	return resp
}

// CreateCasePhaseRequest 建立案件階段請求
type CreateCasePhaseRequest struct {
	Name                string  `json:"name" binding:"required"`
	StartDate           string  `json:"start_date"`
	DurationDays        int     `json:"duration_days"`
	Order               *int    `json:"order"`
	CollaborationItemID *string `json:"collaboration_item_id"`
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

	if req.CollaborationItemID != nil && *req.CollaborationItemID != "" {
		if itemUUID, err := uuid.Parse(*req.CollaborationItemID); err == nil {
			phase.CollaborationItemID = &itemUUID
		}
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

	// 取得目前最大 order，新階段 append 到最後
	var maxOrder int
	h.db.Model(&models.CasePhase{}).Where("case_id = ?", caseUUID).
		Select(`COALESCE(MAX("order"), -1)`).Scan(&maxOrder)

	tx := h.db.Begin()

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
			Order:           maxOrder + 1 + i,
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

	// 回傳所有階段（包含既有的）
	var allPhases []models.CasePhase
	h.db.Where("case_id = ?", caseUUID).Order(`"order" ASC`).Find(&allPhases)

	data := make([]CasePhaseResponse, 0, len(allPhases))
	for _, p := range allPhases {
		data = append(data, casePhaseToResponse(&p))
	}

	c.JSON(http.StatusOK, gin.H{"data": data, "message": fmt.Sprintf("已新增「%s」流程（%d 個階段）", tmpl.Name, len(createdPhases))})
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

// ClearCasePhases 清空案件所有流程階段
func (h *CaseHandler) ClearCasePhases(c *gin.Context) {
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

	result := h.db.Where("case_id = ?", caseUUID).Delete(&models.CasePhase{})
	if result.Error != nil {
		logger.Error().Err(result.Error).Msg("Failed to clear case phases")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to clear phases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "所有流程階段已清空", "deleted_count": result.RowsAffected})
}

// AutoApplyTemplate 根據案件關聯的合作項目自動套用流程範本
// 每個 individual 合作項目如果綁定了 workflow，就建立對應的流程階段
func (h *CaseHandler) AutoApplyTemplate(c *gin.Context) {
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

	// 取得案件關聯的合作項目（含 workflow）
	var cciList []models.CaseCollaborationItem
	if err := h.db.Where("case_id = ?", caseUUID).
		Preload("CollaborationItem").
		Preload("CollaborationItem.Workflow").
		Preload("CollaborationItem.Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("CollaborationItem.BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("CollaborationItem.BundleItems.Item").
		Preload("CollaborationItem.BundleItems.Item.Workflow").
		Preload("CollaborationItem.BundleItems.Item.Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Order(`"order" ASC`).
		Find(&cciList).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to fetch case collaboration items")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch collaboration items"})
		return
	}

	if len(cciList) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"matched": false,
			"reason":  "案件尚未關聯任何合作項目，請先新增合作項目",
			"message": "案件尚未關聯任何合作項目",
		})
		return
	}

	// 清除舊的流程階段
	tx := h.db.Begin()
	if err := tx.Where("case_id = ?", caseUUID).Delete(&models.CasePhase{}).Error; err != nil {
		tx.Rollback()
		logger.Error().Err(err).Msg("Failed to delete existing case phases")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to clear phases"})
		return
	}

	// 根據每個合作項目的 workflow 建立流程階段
	appliedItems := []string{}
	skippedItems := []string{}
	totalPhases := 0

	for _, cci := range cciList {
		item := cci.CollaborationItem
		if item.Type == models.CollaborationItemTypeBundle {
			// Bundle：為每個內含的 individual 項目套用流程
			for _, bi := range item.BundleItems {
				if bi.Item.Workflow != nil && len(bi.Item.Workflow.Phases) > 0 {
					h.createPhasesFromWorkflow(tx, caseUUID, bi.ItemID, bi.Item.Workflow)
					appliedItems = append(appliedItems, bi.Item.Title)
					totalPhases += len(bi.Item.Workflow.Phases)
				} else {
					skippedItems = append(skippedItems, bi.Item.Title)
				}
			}
		} else {
			// Individual：直接套用
			if item.Workflow != nil && len(item.Workflow.Phases) > 0 {
				h.createPhasesFromWorkflow(tx, caseUUID, item.ID, item.Workflow)
				appliedItems = append(appliedItems, item.Title)
				totalPhases += len(item.Workflow.Phases)
			} else {
				skippedItems = append(skippedItems, item.Title)
			}
		}
	}

	tx.Commit()

	if totalPhases == 0 {
		reason := "關聯的合作項目都沒有設定流程範本，請先到設定頁為項目綁定流程"
		c.JSON(http.StatusOK, gin.H{
			"matched":       false,
			"reason":        reason,
			"message":       reason,
			"skipped_items": skippedItems,
		})
		return
	}

	// 根據 flow_layout 計算日期
	startDate := time.Now().Truncate(24 * time.Hour)
	if cs.FlowLayout == "parallel" {
		h.recalculateParallelDates(caseUUID, startDate)
	} else {
		h.recalculateSequentialDates(caseUUID)
	}

	// 回傳更新後的階段
	var phases []models.CasePhase
	h.db.Where("case_id = ?", caseUUID).Order(`"order" ASC`).Find(&phases)

	data := make([]CasePhaseResponse, 0, len(phases))
	for _, p := range phases {
		data = append(data, casePhaseToResponse(&p))
	}

	msg := fmt.Sprintf("已套用 %d 個項目的流程（共 %d 個階段）", len(appliedItems), totalPhases)
	if len(skippedItems) > 0 {
		msg += fmt.Sprintf("，%d 個項目未設定流程範本", len(skippedItems))
	}

	c.JSON(http.StatusOK, gin.H{
		"matched":        true,
		"applied_items":  appliedItems,
		"skipped_items":  skippedItems,
		"data":           data,
		"message":        msg,
	})
}

// --- AI Auto-Match Collaboration Items ---

// AutoMatchCollaborationItems 手動觸發 AI 自動匹配合作項目
func (h *CaseHandler) AutoMatchCollaborationItems(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")

	if h.openaiService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "openai_unavailable", Message: "AI 服務未設定"})
		return
	}

	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}
	userUUID, _ := uuid.Parse(userID)

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

	// 取得案件關聯的郵件內容（用於 AI 分析）
	emailSubject, emailBody, emailFrom := h.getLatestEmailContent(caseUUID, userUUID)

	// 如果沒有郵件，用案件自身資訊
	if emailSubject == "" && emailBody == "" {
		emailSubject = cs.Title
		if cs.Description != nil {
			emailBody = *cs.Description
		}
		emailFrom = cs.BrandName
	}

	// 讀取使用者合作項目
	var items []models.CollaborationItem
	if err := h.db.Where("user_id = ?", userID).
		Preload("Workflow").
		Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems.Item").
		Preload("BundleItems.Item.Workflow").
		Preload("BundleItems.Item.Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Find(&items).Error; err != nil || len(items) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"matched":    false,
			"reason":     "沒有可用的合作項目",
			"matched_ids": []string{},
		})
		return
	}

	// 準備 AI 匹配請求
	itemInfos := make([]openai.CollaborationItemInfo, 0, len(items))
	for _, item := range items {
		desc := ""
		if item.Description != nil {
			desc = *item.Description
		}
		info := openai.CollaborationItemInfo{
			ID:          item.ID.String(),
			Title:       item.Title,
			Description: desc,
			Price:       item.Price,
			Type:        string(item.Type),
		}
		if item.Type == models.CollaborationItemTypeBundle {
			for _, bi := range item.BundleItems {
				info.ItemNames = append(info.ItemNames, bi.Item.Title)
			}
		}
		itemInfos = append(itemInfos, info)
	}

	matchResult, err := h.openaiService.MatchCollaborationItems(c.Request.Context(), openai.MatchCollaborationItemsRequest{
		EmailSubject: emailSubject,
		EmailBody:    emailBody,
		EmailFrom:    emailFrom,
		Items:        itemInfos,
	})
	if err != nil {
		logger.Error().Err(err).Msg("AI collaboration item matching failed")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "ai_error", Message: "AI 分析失敗"})
		return
	}

	if matchResult.Confidence < 0.5 || len(matchResult.MatchedItemIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"matched":     false,
			"confidence":  matchResult.Confidence,
			"reason":      matchResult.Reason,
			"matched_ids": []string{},
		})
		return
	}

	// 建立關聯 + 套用流程
	matchedNames := []string{}
	for i, matchedID := range matchResult.MatchedItemIDs {
		itemUUID, err := uuid.Parse(matchedID)
		if err != nil {
			continue
		}

		// 建立關聯（忽略已存在的）
		cci := models.CaseCollaborationItem{
			CaseID:              caseUUID,
			CollaborationItemID: itemUUID,
			Order:               i,
		}
		h.db.Create(&cci) // ignore duplicate errors

		// 找到對應的項目並套用流程
		for idx := range items {
			if items[idx].ID.String() == matchedID {
				matchedNames = append(matchedNames, items[idx].Title)
				// 建立流程階段
				h.createPhasesForItem(h.db, caseUUID, &items[idx], cs.FlowLayout)
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"matched":       true,
		"confidence":    matchResult.Confidence,
		"reason":        matchResult.Reason,
		"matched_ids":   matchResult.MatchedItemIDs,
		"matched_names": matchedNames,
		"message":       fmt.Sprintf("AI 自動匹配了 %d 個合作項目", len(matchResult.MatchedItemIDs)),
	})
}

// autoMatchCollaborationItemsForCase 背景自動匹配（案件建立後觸發）
func (h *CaseHandler) autoMatchCollaborationItemsForCase(caseID, userID uuid.UUID) {
	if h.openaiService == nil {
		return
	}

	// 檢查是否已有合作項目關聯
	var count int64
	h.db.Model(&models.CaseCollaborationItem{}).Where("case_id = ?", caseID).Count(&count)
	if count > 0 {
		return // 已有關聯，不自動匹配
	}

	var cs models.Case
	if err := h.db.Where("id = ? AND user_id = ?", caseID, userID).First(&cs).Error; err != nil {
		return
	}

	emailSubject, emailBody, emailFrom := h.getLatestEmailContent(caseID, userID)
	if emailSubject == "" && emailBody == "" {
		emailSubject = cs.Title
		if cs.Description != nil {
			emailBody = *cs.Description
		}
		emailFrom = cs.BrandName
	}

	// 如果完全沒有內容可供分析，跳過
	if emailSubject == "" && emailBody == "" {
		return
	}

	var items []models.CollaborationItem
	if err := h.db.Where("user_id = ?", userID).
		Preload("Workflow").
		Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems.Item").
		Preload("BundleItems.Item.Workflow").
		Preload("BundleItems.Item.Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Find(&items).Error; err != nil || len(items) == 0 {
		return
	}

	itemInfos := make([]openai.CollaborationItemInfo, 0, len(items))
	for _, item := range items {
		desc := ""
		if item.Description != nil {
			desc = *item.Description
		}
		info := openai.CollaborationItemInfo{
			ID:          item.ID.String(),
			Title:       item.Title,
			Description: desc,
			Price:       item.Price,
			Type:        string(item.Type),
		}
		if item.Type == models.CollaborationItemTypeBundle {
			for _, bi := range item.BundleItems {
				info.ItemNames = append(info.ItemNames, bi.Item.Title)
			}
		}
		itemInfos = append(itemInfos, info)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	matchResult, err := h.openaiService.MatchCollaborationItems(ctx, openai.MatchCollaborationItemsRequest{
		EmailSubject: emailSubject,
		EmailBody:    emailBody,
		EmailFrom:    emailFrom,
		Items:        itemInfos,
	})
	if err != nil || matchResult.Confidence < 0.5 || len(matchResult.MatchedItemIDs) == 0 {
		return
	}

	for i, matchedID := range matchResult.MatchedItemIDs {
		itemUUID, err := uuid.Parse(matchedID)
		if err != nil {
			continue
		}
		cci := models.CaseCollaborationItem{
			CaseID:              caseID,
			CollaborationItemID: itemUUID,
			Order:               i,
		}
		h.db.Create(&cci)

		for idx := range items {
			if items[idx].ID.String() == matchedID {
				h.createPhasesForItem(h.db, caseID, &items[idx], cs.FlowLayout)
				break
			}
		}
	}
}

// getLatestEmailContent 取得案件最新郵件內容
func (h *CaseHandler) getLatestEmailContent(caseID, userID uuid.UUID) (subject, body, from string) {
	var email models.Email
	err := h.db.Joins("JOIN oauth_accounts ON oauth_accounts.id = emails.oauth_account_id").
		Where("emails.case_id = ? AND oauth_accounts.user_id = ?", caseID, userID).
		Order("emails.received_at DESC").
		First(&email).Error
	if err != nil {
		return "", "", ""
	}
	if email.Subject != nil {
		subject = *email.Subject
	}
	if email.BodyText != nil {
		body = *email.BodyText
	} else if email.Snippet != nil {
		body = *email.Snippet
	}
	from = email.FromEmail
	if email.FromName != nil && *email.FromName != "" {
		from = *email.FromName
	}
	return
}

// --- Case Collaboration Items (many-to-many) ---

// AddCaseCollaborationItemRequest 新增案件合作項目關聯
type AddCaseCollaborationItemRequest struct {
	CollaborationItemID string `json:"collaboration_item_id" binding:"required"`
}

// ReorderCaseCollaborationItemsRequest 重新排序案件合作項目
type ReorderCaseCollaborationItemsRequest struct {
	ItemIDs []string `json:"item_ids" binding:"required"` // collaboration_item_id 順序
}

// ListCaseCollaborationItems 取得案件關聯的合作項目
func (h *CaseHandler) ListCaseCollaborationItems(c *gin.Context) {
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

	var items []models.CaseCollaborationItem
	if err := h.db.Where("case_id = ?", caseUUID).
		Preload("CollaborationItem").
		Preload("CollaborationItem.BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("CollaborationItem.BundleItems.Item").
		Preload("CollaborationItem.Workflow").
		Order(`"order" ASC`).
		Find(&items).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to list case collaboration items")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to list items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// AddCaseCollaborationItem 新增案件合作項目關聯 + 自動建立流程階段
func (h *CaseHandler) AddCaseCollaborationItem(c *gin.Context) {
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

	var req AddCaseCollaborationItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	itemUUID, err := uuid.Parse(req.CollaborationItemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_item_id", Message: "Invalid collaboration item ID"})
		return
	}

	// Verify item exists and belongs to user
	var item models.CollaborationItem
	if err := h.db.Where("id = ? AND user_id = ?", itemUUID, userID).
		Preload("Workflow").
		Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems.Item").
		Preload("BundleItems.Item.Workflow").
		Preload("BundleItems.Item.Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "item_not_found", Message: "Collaboration item not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch collaboration item")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch item"})
		return
	}

	tx := h.db.Begin()

	// Calculate order
	var maxOrder int
	h.db.Model(&models.CaseCollaborationItem{}).Where("case_id = ?", caseUUID).
		Select(`COALESCE(MAX("order"), -1)`).Scan(&maxOrder)

	cci := models.CaseCollaborationItem{
		CaseID:              caseUUID,
		CollaborationItemID: itemUUID,
		Order:               maxOrder + 1,
	}
	if err := tx.Create(&cci).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusConflict, ErrorResponse{Error: "already_linked", Message: "此合作項目已關聯到案件"})
		return
	}

	// Auto-create flow phases from the item's workflow
	h.createPhasesForItem(tx, caseUUID, &item, cs.FlowLayout)

	tx.Commit()

	// Reload
	h.db.Where("id = ?", cci.ID).
		Preload("CollaborationItem").
		Preload("CollaborationItem.BundleItems.Item").
		Preload("CollaborationItem.Workflow").
		First(&cci)

	c.JSON(http.StatusCreated, cci)
}

// RemoveCaseCollaborationItem 移除案件合作項目關聯 + 刪除對應流程階段
func (h *CaseHandler) RemoveCaseCollaborationItem(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	caseID := c.Param("id")
	itemID := c.Param("itemId")

	caseUUID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid case ID"})
		return
	}
	itemUUID, err := uuid.Parse(itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid item ID"})
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

	tx := h.db.Begin()

	// Delete the association
	result := tx.Where("case_id = ? AND collaboration_item_id = ?", caseUUID, itemUUID).
		Delete(&models.CaseCollaborationItem{})
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_linked", Message: "此合作項目未關聯到案件"})
		return
	}

	// Delete associated phases (but check if it's the last phases)
	var totalPhases int64
	tx.Model(&models.CasePhase{}).Where("case_id = ?", caseUUID).Count(&totalPhases)

	var itemPhases int64
	tx.Model(&models.CasePhase{}).Where("case_id = ? AND collaboration_item_id = ?", caseUUID, itemUUID).Count(&itemPhases)

	if totalPhases > itemPhases {
		// Safe to delete — there will be remaining phases
		tx.Where("case_id = ? AND collaboration_item_id = ?", caseUUID, itemUUID).Delete(&models.CasePhase{})
	}
	// If all phases belong to this item, keep them but set collaboration_item_id to NULL

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "合作項目已從案件移除"})
}

// ReorderCaseCollaborationItems 重新排序案件合作項目
func (h *CaseHandler) ReorderCaseCollaborationItems(c *gin.Context) {
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

	var req ReorderCaseCollaborationItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	tx := h.db.Begin()
	for i, idStr := range req.ItemIDs {
		itemUUID, err := uuid.Parse(idStr)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid item ID: " + idStr})
			return
		}
		if err := tx.Model(&models.CaseCollaborationItem{}).
			Where("case_id = ? AND collaboration_item_id = ?", caseUUID, itemUUID).
			Update("order", i).Error; err != nil {
			tx.Rollback()
			logger.Error().Err(err).Msg("Failed to reorder case collaboration items")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to reorder"})
			return
		}
	}
	tx.Commit()

	// If sequential layout, recalculate phase dates
	if cs.FlowLayout == "sequential" {
		h.recalculateSequentialDates(caseUUID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reordered successfully"})
}

// UpdateFlowLayoutRequest 更新流程排列模式
type UpdateFlowLayoutRequest struct {
	FlowLayout string `json:"flow_layout" binding:"required"` // "parallel" or "sequential"
	StartDate  string `json:"start_date"`                     // Optional start date for recalculation
}

// UpdateFlowLayout 切換並聯/串聯模式
func (h *CaseHandler) UpdateFlowLayout(c *gin.Context) {
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

	var req UpdateFlowLayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	if req.FlowLayout != "parallel" && req.FlowLayout != "sequential" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_layout", Message: "flow_layout must be 'parallel' or 'sequential'"})
		return
	}

	// Update the flow layout
	if err := h.db.Model(&cs).Update("flow_layout", req.FlowLayout).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to update flow layout")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to update flow layout"})
		return
	}

	// Recalculate dates based on new layout
	startDate := time.Now().Truncate(24 * time.Hour)
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = t
		}
	} else {
		// Use earliest existing phase start date
		var firstPhase models.CasePhase
		if err := h.db.Where("case_id = ? AND start_date IS NOT NULL", caseUUID).
			Order("start_date ASC").First(&firstPhase).Error; err == nil && firstPhase.StartDate != nil {
			startDate = *firstPhase.StartDate
		}
	}

	if req.FlowLayout == "parallel" {
		h.recalculateParallelDates(caseUUID, startDate)
	} else {
		h.recalculateSequentialDates(caseUUID)
	}

	// Return updated phases
	var phases []models.CasePhase
	h.db.Where("case_id = ?", caseUUID).Order(`"order" ASC`).Find(&phases)

	data := make([]CasePhaseResponse, 0, len(phases))
	for _, p := range phases {
		data = append(data, casePhaseToResponse(&p))
	}

	c.JSON(http.StatusOK, gin.H{"flow_layout": req.FlowLayout, "phases": data})
}

// createPhasesForItem 為合作項目建立流程階段
func (h *CaseHandler) createPhasesForItem(tx *gorm.DB, caseID uuid.UUID, item *models.CollaborationItem, _ string) {
	if item.Type == models.CollaborationItemTypeBundle {
		// For bundles, create phases for each individual item inside
		for _, bi := range item.BundleItems {
			if bi.Item.Workflow != nil && len(bi.Item.Workflow.Phases) > 0 {
				h.createPhasesFromWorkflow(tx, caseID, bi.ItemID, bi.Item.Workflow)
			}
		}
	} else if item.Workflow != nil && len(item.Workflow.Phases) > 0 {
		h.createPhasesFromWorkflow(tx, caseID, item.ID, item.Workflow)
	}
}

// createPhasesFromWorkflow 從流程範本建立案件階段
func (h *CaseHandler) createPhasesFromWorkflow(tx *gorm.DB, caseID uuid.UUID, itemID uuid.UUID, wf *models.WorkflowTemplate) {
	var maxOrder int
	tx.Model(&models.CasePhase{}).Where("case_id = ?", caseID).
		Select(`COALESCE(MAX("order"), -1)`).Scan(&maxOrder)

	startDate := time.Now().Truncate(24 * time.Hour)
	currentDate := startDate

	for i, wp := range wf.Phases {
		endDate := currentDate.AddDate(0, 0, wp.DurationDays)
		wpID := wp.ID
		phase := models.CasePhase{
			CaseID:              caseID,
			Name:                wp.Name,
			StartDate:           &currentDate,
			EndDate:             &endDate,
			DurationDays:        wp.DurationDays,
			Order:               maxOrder + 1 + i,
			WorkflowPhaseID:     &wpID,
			CollaborationItemID: &itemID,
		}
		tx.Create(&phase)
		currentDate = endDate
	}
}

// recalculateParallelDates 重新計算並聯模式日期（所有流程從同一天開始）
func (h *CaseHandler) recalculateParallelDates(caseID uuid.UUID, startDate time.Time) {
	// Group phases by collaboration_item_id
	var phases []models.CasePhase
	h.db.Where("case_id = ?", caseID).Order(`collaboration_item_id, "order" ASC`).Find(&phases)

	currentDate := startDate
	var currentItemID *uuid.UUID

	for i := range phases {
		p := &phases[i]
		// Reset date for each new item group
		if currentItemID == nil || p.CollaborationItemID == nil || *currentItemID != *p.CollaborationItemID {
			currentDate = startDate
			currentItemID = p.CollaborationItemID
		}

		endDate := currentDate.AddDate(0, 0, p.DurationDays)
		h.db.Model(p).Updates(map[string]any{
			"start_date": currentDate,
			"end_date":   endDate,
		})
		currentDate = endDate
	}
}

// recalculateSequentialDates 重新計算串聯模式日期（按 case_collaboration_items order 順序串接）
func (h *CaseHandler) recalculateSequentialDates(caseID uuid.UUID) {
	// Get items in order
	var cciList []models.CaseCollaborationItem
	h.db.Where("case_id = ?", caseID).Order(`"order" ASC`).Find(&cciList)

	// Find earliest start date
	startDate := time.Now().Truncate(24 * time.Hour)
	var firstPhase models.CasePhase
	if err := h.db.Where("case_id = ? AND start_date IS NOT NULL", caseID).
		Order("start_date ASC").First(&firstPhase).Error; err == nil && firstPhase.StartDate != nil {
		startDate = *firstPhase.StartDate
	}

	currentDate := startDate

	for _, cci := range cciList {
		var phases []models.CasePhase
		h.db.Where("case_id = ? AND collaboration_item_id = ?", caseID, cci.CollaborationItemID).
			Order(`"order" ASC`).Find(&phases)

		for i := range phases {
			p := &phases[i]
			endDate := currentDate.AddDate(0, 0, p.DurationDays)
			h.db.Model(p).Updates(map[string]any{
				"start_date": currentDate,
				"end_date":   endDate,
			})
			currentDate = endDate
		}
	}

	// Handle phases without collaboration_item_id
	var orphanPhases []models.CasePhase
	h.db.Where("case_id = ? AND collaboration_item_id IS NULL", caseID).
		Order(`"order" ASC`).Find(&orphanPhases)
	for i := range orphanPhases {
		p := &orphanPhases[i]
		endDate := currentDate.AddDate(0, 0, p.DurationDays)
		h.db.Model(p).Updates(map[string]any{
			"start_date": currentDate,
			"end_date":   endDate,
		})
		currentDate = endDate
	}
}
