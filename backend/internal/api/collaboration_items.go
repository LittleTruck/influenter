package api

import (
	"fmt"
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
	Title         string  `json:"title" binding:"required"`
	Description   *string `json:"description"`
	Notes         *string `json:"notes"`
	Price         float64 `json:"price"`
	Type          string  `json:"type"`           // "individual" or "bundle", defaults to "individual"
	BundleItemIDs []string `json:"bundle_item_ids"` // Only for bundle type
	WorkflowID    *string `json:"workflow_id"`     // Only for individual type
}

// UpdateCollaborationItemRequest 更新合作項目請求
type UpdateCollaborationItemRequest struct {
	Title         *string  `json:"title"`
	Description   *string  `json:"description"`
	Notes         *string  `json:"notes"`
	Price         *float64 `json:"price"`
	BundleItemIDs *[]string `json:"bundle_item_ids"` // Only for bundle type
	WorkflowID    *string  `json:"workflow_id"`      // Only for individual type
}

// ReorderItemsRequest 重新排序請求
type ReorderItemsRequest struct {
	ItemIDs []string `json:"item_ids" binding:"required"`
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
	query := h.db.Where("user_id = ?", userID).
		Preload("Workflow").
		Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems.Item").
		Order(`"order" ASC`)

	// Optional type filter
	if itemType := c.Query("type"); itemType != "" {
		query = query.Where("type = ?", itemType)
	}

	if err := query.Find(&items).Error; err != nil {
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

	// Default type to individual
	itemType := models.CollaborationItemType(req.Type)
	if itemType == "" {
		itemType = models.CollaborationItemTypeIndividual
	}
	if itemType != models.CollaborationItemTypeIndividual && itemType != models.CollaborationItemTypeBundle {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_type", Message: "Type must be 'individual' or 'bundle'"})
		return
	}

	// Bundle cannot have workflow
	if itemType == models.CollaborationItemTypeBundle && req.WorkflowID != nil && *req.WorkflowID != "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: "Bundle items cannot have a workflow"})
		return
	}

	item := models.CollaborationItem{
		UserID: userID,
		Title:  req.Title,
		Price:  req.Price,
		Type:   itemType,
	}
	if req.Description != nil {
		item.Description = req.Description
	}
	if req.Notes != nil {
		item.Notes = req.Notes
	}
	if itemType == models.CollaborationItemTypeIndividual && req.WorkflowID != nil && *req.WorkflowID != "" {
		wid, err := uuid.Parse(*req.WorkflowID)
		if err == nil {
			item.WorkflowID = &wid
		}
	}

	// Calculate order: max order + 1
	var maxOrder int
	h.db.Model(&models.CollaborationItem{}).
		Where("user_id = ?", userID).
		Select(`COALESCE(MAX("order"), -1)`).
		Scan(&maxOrder)
	item.Order = maxOrder + 1

	tx := h.db.Begin()

	if err := tx.Create(&item).Error; err != nil {
		tx.Rollback()
		logger.Error().Err(err).Msg("Failed to create collaboration item")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to create collaboration item"})
		return
	}

	// Create bundle_items if bundle type
	if itemType == models.CollaborationItemTypeBundle && len(req.BundleItemIDs) > 0 {
		if err := h.createBundleItems(tx, userID, item.ID, req.BundleItemIDs); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_bundle_items", Message: err.Error()})
			return
		}
	}

	tx.Commit()

	// Reload with relationships
	h.db.Preload("Workflow").
		Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems.Item").
		First(&item, "id = ?", item.ID)

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

	// Bundle cannot have workflow
	if item.Type == models.CollaborationItemTypeBundle && req.WorkflowID != nil && *req.WorkflowID != "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: "Bundle items cannot have a workflow"})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if item.Type == models.CollaborationItemTypeIndividual && req.WorkflowID != nil {
		if *req.WorkflowID == "" {
			updates["workflow_id"] = nil
		} else if wid, err := uuid.Parse(*req.WorkflowID); err == nil {
			updates["workflow_id"] = wid
		}
	}

	tx := h.db.Begin()

	if len(updates) > 0 {
		if err := tx.Model(&item).Updates(updates).Error; err != nil {
			tx.Rollback()
			logger.Error().Err(err).Msg("Failed to update collaboration item")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to update collaboration item"})
			return
		}
	}

	// Update bundle items if provided
	if item.Type == models.CollaborationItemTypeBundle && req.BundleItemIDs != nil {
		userUUID, _ := uuid.Parse(userID)
		// Delete existing bundle items
		if err := tx.Where("bundle_id = ?", item.ID).Delete(&models.BundleItem{}).Error; err != nil {
			tx.Rollback()
			logger.Error().Err(err).Msg("Failed to delete existing bundle items")
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to update bundle items"})
			return
		}
		// Create new bundle items
		if len(*req.BundleItemIDs) > 0 {
			if err := h.createBundleItems(tx, userUUID, item.ID, *req.BundleItemIDs); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_bundle_items", Message: err.Error()})
				return
			}
		}
	}

	tx.Commit()

	// Reload with relationships
	h.db.Preload("Workflow").
		Preload("Workflow.Phases", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems", func(db *gorm.DB) *gorm.DB {
			return db.Order(`"order" ASC`)
		}).
		Preload("BundleItems.Item").
		First(&item, "id = ?", item.ID)

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

// createBundleItems validates and creates bundle item associations
func (h *CollaborationItemHandler) createBundleItems(tx *gorm.DB, userID uuid.UUID, bundleID uuid.UUID, itemIDs []string) error {
	for i, idStr := range itemIDs {
		itemUUID, err := uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("invalid item ID: %s", idStr)
		}

		// Verify item exists, belongs to user, and is individual type
		var target models.CollaborationItem
		if err := tx.Where("id = ? AND user_id = ? AND type = ?", itemUUID, userID, models.CollaborationItemTypeIndividual).
			First(&target).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("item %s not found or is not an individual item", idStr)
			}
			return fmt.Errorf("failed to verify item %s: %w", idStr, err)
		}

		bundleItem := models.BundleItem{
			BundleID: bundleID,
			ItemID:   itemUUID,
			Order:    i,
		}
		if err := tx.Create(&bundleItem).Error; err != nil {
			return fmt.Errorf("failed to create bundle item association: %w", err)
		}
	}
	return nil
}
