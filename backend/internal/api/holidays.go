package api

import (
	"net/http"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HolidayHandler 國定假日處理器
type HolidayHandler struct {
	db *gorm.DB
}

// NewHolidayHandler 建立國定假日處理器
func NewHolidayHandler(db *gorm.DB) *HolidayHandler {
	return &HolidayHandler{db: db}
}

// HolidayResponse 假日 API 回應。
// date 以 "YYYY-MM-DD" 字串回傳，避免前端因時區位移誤判日期。
type HolidayResponse struct {
	Date      string `json:"date"`
	Name      string `json:"name"`
	IsWorkday bool   `json:"is_workday"` // true=補班日（即使週末仍須上班）
	Source    string `json:"source"`
}

// ListHolidays 回傳國定假日 / 補班日清單，供前端計算工作日（工期預覽與期限倒數）。
//
// 目前僅回傳 source='national' 的內建資料；未來可依登入帳號合併 source='custom'
// 的自訂假日，前端無需更動。
func (h *HolidayHandler) ListHolidays(c *gin.Context) {
	var rows []models.NationalHoliday
	if err := h.db.Where("source = ?", "national").Order("date ASC").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to fetch holidays"})
		return
	}

	data := make([]HolidayResponse, 0, len(rows))
	for _, r := range rows {
		data = append(data, HolidayResponse{
			Date:      r.Date.Format("2006-01-02"),
			Name:      r.Name,
			IsWorkday: r.IsWorkday,
			Source:    r.Source,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}
