// Command recalc-phases 是一次性 backfill 工具：
// 將所有既有案件的流程階段日期，依「工作日」規則（排除週末與國定假日）重算。
//
// 背景：本專案把階段排程從「日曆天」改為「工作日」後，既有案件仍是用舊日曆天算出的日期。
// 本工具重跑與 API 相同的重算邏輯（CaseHandler.RecalculateCaseDates），以既有最早 start_date
// 作為錨點，僅修正工作日數學，不把時間軸整體推到今天。
//
// 用法：
//
//	go run ./cmd/recalc-phases
package main

import (
	"log"

	"github.com/designcomb/influenter-backend/internal/api"
	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/designcomb/influenter-backend/internal/database"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/services/schedule"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	// openaiService 在重算路徑用不到，傳 nil。
	handler := api.NewCaseHandler(db.DB, nil)

	var cases []models.Case
	if err := db.DB.Select("id, user_id, flow_layout").Find(&cases).Error; err != nil {
		log.Fatalf("❌ Failed to load cases: %v", err)
	}

	log.Printf("▶️  掃描 %d 個案件，重算含階段者的工作日時程…", len(cases))

	updated := 0
	for _, cs := range cases {
		var phaseCount int64
		db.DB.Model(&models.CasePhase{}).Where("case_id = ?", cs.ID).Count(&phaseCount)
		if phaseCount == 0 {
			continue
		}
		dayCfg := schedule.ResolveDayCalcConfig(db.DB, cs.UserID.String())
		handler.RecalculateCaseDates(cs.ID, cs.FlowLayout, dayCfg)
		updated++
	}

	log.Printf("✅ 完成：已重算 %d 個含階段的案件", updated)
}
