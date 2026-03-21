package api

import (
	"net/http"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReplyTemplateHandler 回覆範本處理器
type ReplyTemplateHandler struct {
	db *gorm.DB
}

// NewReplyTemplateHandler 建立新的回覆範本處理器
func NewReplyTemplateHandler(db *gorm.DB) *ReplyTemplateHandler {
	return &ReplyTemplateHandler{db: db}
}

// CreateReplyTemplateRequest 建立回覆範本請求
type CreateReplyTemplateRequest struct {
	Title  string `json:"title" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
}

// UpdateReplyTemplateRequest 更新回覆範本請求
type UpdateReplyTemplateRequest struct {
	Title  *string `json:"title"`
	Prompt *string `json:"prompt"`
}

// ListTemplates 取得使用者的所有回覆範本
func (h *ReplyTemplateHandler) ListTemplates(c *gin.Context) {
	userID := c.GetString("user_id")

	var templates []models.ReplyTemplate
	if err := h.db.Where("user_id = ?", userID).
		Order(`"order" ASC, created_at ASC`).
		Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list reply templates",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": templates})
}

// CreateTemplate 建立回覆範本
func (h *ReplyTemplateHandler) CreateTemplate(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_user_id", Message: "Invalid user ID"})
		return
	}

	var req CreateReplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	// 計算 order
	var maxOrder int
	h.db.Model(&models.ReplyTemplate{}).
		Where("user_id = ?", userID).
		Select(`COALESCE(MAX("order"), -1)`).
		Scan(&maxOrder)

	template := models.ReplyTemplate{
		UserID: userID,
		Title:  req.Title,
		Prompt: req.Prompt,
		Order:  maxOrder + 1,
	}

	if err := h.db.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to create reply template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// UpdateTemplate 更新回覆範本
func (h *ReplyTemplateHandler) UpdateTemplate(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid template ID"})
		return
	}

	var template models.ReplyTemplate
	if err := h.db.Where("id = ? AND user_id = ?", id, userIDStr).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to find template"})
		return
	}

	var req UpdateReplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Prompt != nil {
		updates["prompt"] = *req.Prompt
	}

	if len(updates) > 0 {
		if err := h.db.Model(&template).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to update template"})
			return
		}
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate 刪除回覆範本
func (h *ReplyTemplateHandler) DeleteTemplate(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid template ID"})
		return
	}

	var template models.ReplyTemplate
	if err := h.db.Where("id = ? AND user_id = ?", id, userIDStr).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to find template"})
		return
	}

	if err := h.db.Delete(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: "Failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}
