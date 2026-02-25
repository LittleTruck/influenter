package api

import (
	"net/http"

	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CollaborationItemHandler 合作項目處理器
type CollaborationItemHandler struct {
	db *gorm.DB
}

// NewCollaborationItemHandler 建立合作項目處理器
func NewCollaborationItemHandler(db *gorm.DB) *CollaborationItemHandler {
	return &CollaborationItemHandler{db: db}
}

// CreateCollaborationItemRequest 建立合作項目請求
type CreateCollaborationItemRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	ParentID    *string `json:"parent_id"`
	WorkflowID  *string `json:"workflow_id"`
}

// UpdateCollaborationItemRequest 更新合作項目請求
type UpdateCollaborationItemRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	ParentID    *string  `json:"parent_id"`
	WorkflowID  *string  `json:"workflow_id"`
}

// ReorderItemsRequest 重新排序請求
type ReorderItemsRequest struct {
	ItemIDs  []string `json:"item_ids" binding:"required"`
	ParentID *string  `json:"parent_id"`
}

// ListItems 取得合作項目列表
func (h *CollaborationItemHandler) ListItems(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "user_id required"})
		return
	}

	var items []models.CollaborationItem
	err := h.db.Where("user_id = ?", userID).
		Preload("Workflow").
		Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Order(`"order" ASC`).
		Find(&items).Error
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list collaboration items")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to list collaboration items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateItem 建立合作項目
func (h *CollaborationItemHandler) CreateItem(c *gin.Context) {
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

	var req CreateCollaborationItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	item := models.CollaborationItem{
		UserID: userID,
		Title:  req.Title,
		Price:  req.Price,
	}
	if req.Description != nil {
		item.Description = req.Description
	}
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err == nil {
			item.ParentID = &pid
		}
	}
	if req.WorkflowID != nil && *req.WorkflowID != "" {
		wid, err := uuid.Parse(*req.WorkflowID)
		if err == nil {
			item.WorkflowID = &wid
		}
	}

	// 計算 order：同層級的最大 order + 1
	var maxOrder int
	q := h.db.Model(&models.CollaborationItem{}).Where("user_id = ?", userID)
	if item.ParentID != nil {
		q = q.Where("parent_id = ?", item.ParentID)
	} else {
		q = q.Where("parent_id IS NULL")
	}
	q.Select("COALESCE(MAX(\"order\"), -1)").Scan(&maxOrder)
	item.Order = maxOrder + 1

	if err := h.db.Create(&item).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to create collaboration item")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to create collaboration item"})
		return
	}

	// Reload with workflow
	h.db.Preload("Workflow").Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
		return db.Order(`"order" ASC`)
	}).First(&item, "id = ?", item.ID)

	c.JSON(http.StatusCreated, item)
}

// UpdateItem 更新合作項目
func (h *CollaborationItemHandler) UpdateItem(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	itemID := c.Param("id")

	id, err := uuid.Parse(itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid item ID"})
		return
	}

	var item models.CollaborationItem
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Collaboration item not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch collaboration item")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch collaboration item"})
		return
	}

	var req UpdateCollaborationItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.ParentID != nil {
		if *req.ParentID == "" {
			updates["parent_id"] = nil
		} else if pid, err := uuid.Parse(*req.ParentID); err == nil {
			updates["parent_id"] = pid
		}
	}
	if req.WorkflowID != nil {
		if *req.WorkflowID == "" {
			updates["workflow_id"] = nil
		} else if wid, err := uuid.Parse(*req.WorkflowID); err == nil {
			updates["workflow_id"] = wid
		}
	}

	if err := h.db.Model(&item).Updates(updates).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to update collaboration item")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to update collaboration item"})
		return
	}

	// Reload with workflow
	h.db.Preload("Workflow").Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
		return db.Order(`"order" ASC`)
	}).First(&item, "id = ?", item.ID)

	c.JSON(http.StatusOK, item)
}

// DeleteItem 刪除合作項目
func (h *CollaborationItemHandler) DeleteItem(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	itemID := c.Param("id")

	id, err := uuid.Parse(itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid item ID"})
		return
	}

	var item models.CollaborationItem
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "Collaboration item not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to fetch collaboration item")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch collaboration item"})
		return
	}

	if err := h.db.Delete(&item).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to delete collaboration item")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to delete collaboration item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collaboration item deleted"})
}

// ReorderItems 重新排序合作項目
func (h *CollaborationItemHandler) ReorderItems(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "user_id required"})
		return
	}

	var req ReorderItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	tx := h.db.Begin()
	for i, idStr := range req.ItemIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "Invalid item ID: " + idStr})
			return
		}
		if err := tx.Model(&models.CollaborationItem{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("order", i).Error; err != nil {
			tx.Rollback()
			logger.Error().Err(err).Msg("Failed to reorder collaboration items")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to reorder"})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Reordered successfully"})
}
