package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CaseHandler 案件處理器
type CaseHandler struct {
	db *gorm.DB
}

// NewCaseHandler 建立新的案件處理器
func NewCaseHandler(db *gorm.DB) *CaseHandler {
	return &CaseHandler{db: db}
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

	c.JSON(http.StatusOK, caseToResponse(&cs, int(emailCount), 0, 0))
}
