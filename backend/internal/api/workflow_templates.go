package api

import (
	"net/http"

	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowTemplateHandler 流程範本處理器
type WorkflowTemplateHandler struct {
	db *gorm.DB
}

// NewWorkflowTemplateHandler 建立流程範本處理器
func NewWorkflowTemplateHandler(db *gorm.DB) *WorkflowTemplateHandler {
	return &WorkflowTemplateHandler{db: db}
}

// CreateWorkflowTemplateRequest 建立流程範本請求
type CreateWorkflowTemplateRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Color       string  `json:"color"`
}

// UpdateWorkflowTemplateRequest 更新流程範本請求
type UpdateWorkflowTemplateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}

// CreatePhaseRequest 建立流程階段請求
type CreatePhaseRequest struct {
	Name         string `json:"name" binding:"required"`
	DurationDays int    `json:"duration_days"`
	Order        *int   `json:"order"`
}

// UpdatePhaseRequest 更新流程階段請求
type UpdatePhaseRequest struct {
	Name         *string `json:"name"`
	DurationDays *int    `json:"duration_days"`
	Order        *int    `json:"order"`
}

// ListTemplates 取得流程範本列表
func (h *WorkflowTemplateHandler) ListTemplates(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "user_id required"})
		return
	}

	var templates []models.WorkflowTemplate
	err := h.db.Where("user_id = ?", userID).
		Preload("Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Order(`"order" ASC`).
		Find(&templates).Error
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list workflow templates")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to list workflow templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": templates})
}

// CreateTemplate 建立流程範本
func (h *WorkflowTemplateHandler) CreateTemplate(c *gin.Context) {
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

	var req CreateWorkflowTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	color := req.Color
	if color == "" {
		color = "primary"
	}

	tmpl := models.WorkflowTemplate{
		UserID: userID,
		Name:   req.Name,
		Color:  color,
	}
	if req.Description != nil {
		tmpl.Description = req.Description
	}

	// 計算 order
	var maxOrder int
	h.db.Model(&models.WorkflowTemplate{}).Where("user_id = ?", userID).
		Select(`COALESCE(MAX("order"), -1)`).Scan(&maxOrder)
	tmpl.Order = maxOrder + 1

	if err := h.db.Create(&tmpl).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to create workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to create workflow template"})
		return
	}

	// Reload with empty phases
	tmpl.Phases = []models.WorkflowPhase{}

	c.JSON(http.StatusCreated, tmpl)
}

// UpdateTemplate 更新流程範本
func (h *WorkflowTemplateHandler) UpdateTemplate(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	tmplID := c.Param("id")

	id, err := uuid.Parse(tmplID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid template ID"})
		return
	}

	var tmpl models.WorkflowTemplate
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Workflow template not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch workflow template"})
		return
	}

	var req UpdateWorkflowTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}

	if err := h.db.Model(&tmpl).Updates(updates).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to update workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to update workflow template"})
		return
	}

	// Reload with phases
	h.db.Preload("Phases", func(db *gorm.DB) *gorm.DB {
		return db.Order(`"order" ASC`)
	}).First(&tmpl, "id = ?", tmpl.ID)

	c.JSON(http.StatusOK, tmpl)
}

// DeleteTemplate 刪除流程範本
func (h *WorkflowTemplateHandler) DeleteTemplate(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	tmplID := c.Param("id")

	id, err := uuid.Parse(tmplID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid template ID"})
		return
	}

	var tmpl models.WorkflowTemplate
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Workflow template not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch workflow template"})
		return
	}

	if err := h.db.Delete(&tmpl).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to delete workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to delete workflow template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow template deleted"})
}

// CreatePhase 建立流程階段
func (h *WorkflowTemplateHandler) CreatePhase(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	tmplID := c.Param("id")

	tmplUUID, err := uuid.Parse(tmplID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid template ID"})
		return
	}

	// 確認範本屬於當前使用者
	var tmpl models.WorkflowTemplate
	if err := h.db.Where("id = ? AND user_id = ?", tmplUUID, userID).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Workflow template not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch workflow template"})
		return
	}

	var req CreatePhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	durationDays := req.DurationDays
	if durationDays <= 0 {
		durationDays = 1
	}

	phase := models.WorkflowPhase{
		WorkflowTemplateID: tmplUUID,
		Name:               req.Name,
		DurationDays:       durationDays,
	}

	if req.Order != nil {
		phase.Order = *req.Order
	} else {
		var maxOrder int
		h.db.Model(&models.WorkflowPhase{}).Where("workflow_template_id = ?", tmplUUID).
			Select(`COALESCE(MAX("order"), -1)`).Scan(&maxOrder)
		phase.Order = maxOrder + 1
	}

	if err := h.db.Create(&phase).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to create workflow phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to create workflow phase"})
		return
	}

	c.JSON(http.StatusCreated, phase)
}

// UpdatePhase 更新流程階段
func (h *WorkflowTemplateHandler) UpdatePhase(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	tmplID := c.Param("id")
	phaseID := c.Param("phaseId")

	tmplUUID, err := uuid.Parse(tmplID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid template ID"})
		return
	}
	phaseUUID, err := uuid.Parse(phaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid phase ID"})
		return
	}

	// 確認範本屬於當前使用者
	var tmpl models.WorkflowTemplate
	if err := h.db.Where("id = ? AND user_id = ?", tmplUUID, userID).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Workflow template not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch workflow template"})
		return
	}

	var phase models.WorkflowPhase
	if err := h.db.Where("id = ? AND workflow_template_id = ?", phaseUUID, tmplUUID).First(&phase).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Workflow phase not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch workflow phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch workflow phase"})
		return
	}

	var req UpdatePhaseRequest
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

	if err := h.db.Model(&phase).Updates(updates).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to update workflow phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to update workflow phase"})
		return
	}

	h.db.First(&phase, "id = ?", phaseUUID)
	c.JSON(http.StatusOK, phase)
}

// DeletePhase 刪除流程階段
func (h *WorkflowTemplateHandler) DeletePhase(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	tmplID := c.Param("id")
	phaseID := c.Param("phaseId")

	tmplUUID, err := uuid.Parse(tmplID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid template ID"})
		return
	}
	phaseUUID, err := uuid.Parse(phaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid phase ID"})
		return
	}

	// 確認範本屬於當前使用者
	var tmpl models.WorkflowTemplate
	if err := h.db.Where("id = ? AND user_id = ?", tmplUUID, userID).First(&tmpl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Workflow template not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch workflow template")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch workflow template"})
		return
	}

	var phase models.WorkflowPhase
	if err := h.db.Where("id = ? AND workflow_template_id = ?", phaseUUID, tmplUUID).First(&phase).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Workflow phase not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch workflow phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch workflow phase"})
		return
	}

	if err := h.db.Delete(&phase).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to delete workflow phase")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to delete workflow phase"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow phase deleted"})
}
