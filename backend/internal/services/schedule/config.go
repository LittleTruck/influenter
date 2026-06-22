// Package schedule 提供流程排程的「工作日」設定解析。
//
// 它把純函式的工作日引擎（internal/utils）與資料層（national_holidays）接起來，
// 是 handler 與 backfill 取得 DayCalcConfig 的單一入口。
package schedule

import (
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/utils"
	"gorm.io/gorm"
)

// ResolveDayCalcConfig 依使用者建立工作日計算設定。
//
// 目前一律回傳全域預設（工作日模式 + 載入內建國定假日）。
// userID 參數保留作為 per-account 覆寫的 drop-in 介面：未來可在此讀取帳號層級
// 設定以切回日曆日、不排除國定假日或載入自訂假日，而所有呼叫端皆無需更動。
func ResolveDayCalcConfig(db *gorm.DB, userID string) utils.DayCalcConfig {
	cfg := utils.DefaultDayCalcConfig()
	loadNationalHolidays(db, &cfg)
	// TODO(per-account): 依 userID 對應的帳號設定覆寫 cfg.Mode / 假日來源。
	return cfg
}

// loadNationalHolidays 將內建國定假日（source='national'）載入 cfg。
// 查詢失敗時靜默退化為「僅排除週末」，不阻斷排程。
func loadNationalHolidays(db *gorm.DB, cfg *utils.DayCalcConfig) {
	var rows []models.NationalHoliday
	if err := db.Where("source = ?", "national").Find(&rows).Error; err != nil {
		return
	}
	for _, h := range rows {
		key := h.Date.Format("2006-01-02")
		if h.IsWorkday {
			cfg.MakeupWorkdays[key] = true
		} else {
			cfg.Holidays[key] = true
		}
	}
}
