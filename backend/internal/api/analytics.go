package api

import (
	"net/http"
	"sort"
	"time"

	"github.com/designcomb/influenter-backend/internal/middleware"
	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// taipeiZone 台北時區（UTC+8，無日光節約），用於將案件建立時間分桶到月份
var taipeiZone = time.FixedZone("Asia/Taipei", 8*60*60)

// uncategorizedSeriesID 未綁定任何合作項目的案件所歸入的虛擬系列 ID
const uncategorizedSeriesID = "uncategorized"

// AnalyticsSeries 堆疊系列（對應一個合作項目，或「未分類」）
type AnalyticsSeries struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// AnalyticsMonthPoint 單一月份的聚合資料
type AnalyticsMonthPoint struct {
	Month       string             `json:"month"`        // YYYY-MM
	Amounts     map[string]float64 `json:"amounts"`      // seriesID -> 金額
	Counts      map[string]int     `json:"counts"`       // seriesID -> 專案數（項目出現次數）
	TotalAmount float64            `json:"total_amount"` // 當月金額合計
	TotalCount  int                `json:"total_count"`  // 當月專案數合計
}

// CollaborationItemAnalyticsResponse 合作項目數據分析回應
type CollaborationItemAnalyticsResponse struct {
	Series []AnalyticsSeries     `json:"series"`
	Months []AnalyticsMonthPoint `json:"months"`
}

// caseDisplayTotalValue 計算案件顯示總額：adjusted_total 優先，否則項目原始加總，否則 quoted_amount。
// 與前端 caseCalculations.ts 的 getCaseDisplayTotal 對齊。
func caseDisplayTotalValue(adjusted *float64, rawSum float64, quoted *float64) float64 {
	if adjusted != nil {
		return *adjusted
	}
	if rawSum > 0 {
		return rawSum
	}
	if quoted != nil {
		return *quoted
	}
	return 0
}

// CollaborationItemAnalytics 依「建立月份」聚合、依「合作項目」堆疊的金額與專案數。
// X 軸＝案件建立月份（Asia/Taipei）；堆疊系列＝合作項目；
// 納入狀態＝排除「已取消(cancelled)」與「非合作(other)」。
//
// 金額分配：每個案件以其顯示總額（adjusted_total ?? 項目加總 ?? quoted_amount）按各項目原始價格等比分配，
// 因此手動調整總價時仍保留各項目比例，且每月金額合計等於各案件顯示總額之和。
// 無合作項目（或項目皆為 0 元）的案件，其金額歸入「未分類」系列。
// 專案數：每個案件對它的每個合作項目各計 1；無項目的案件對「未分類」計 1。
func (h *CaseHandler) CollaborationItemAnalytics(c *gin.Context) {
	logger := middleware.GetLogger(c)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: "user_id required"})
		return
	}

	var cases []models.Case
	if err := h.db.
		Where("user_id = ? AND status NOT IN ?", userID, []string{
			string(models.CaseStatusCancelled),
			string(models.CaseStatusOther),
		}).
		Find(&cases).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to load cases for analytics")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Failed to load analytics"})
		return
	}

	empty := CollaborationItemAnalyticsResponse{Series: []AnalyticsSeries{}, Months: []AnalyticsMonthPoint{}}
	if len(cases) == 0 {
		c.JSON(http.StatusOK, empty)
		return
	}

	// 載入這些案件關聯的合作項目（含項目本身的標題、預設價、排序）
	caseIDs := make([]uuid.UUID, 0, len(cases))
	for i := range cases {
		caseIDs = append(caseIDs, cases[i].ID)
	}
	var ccis []models.CaseCollaborationItem
	// 以 Unscoped 載入合作項目，讓「已軟刪除」的項目仍能解析出標題/價格/排序，
	// 避免歷史案件引用到被刪除的項目時出現空白系列名稱或排序錯亂。
	h.db.Where("case_id IN ?", caseIDs).
		Preload("CollaborationItem", func(tx *gorm.DB) *gorm.DB { return tx.Unscoped() }).
		Find(&ccis)

	itemsByCase := make(map[uuid.UUID][]models.CaseCollaborationItem, len(cases))
	for _, cci := range ccis {
		itemsByCase[cci.CaseID] = append(itemsByCase[cci.CaseID], cci)
	}

	// 系列中繼資料（排序用）
	type seriesMeta struct {
		title string
		order int
	}
	seriesInfo := make(map[string]seriesMeta)
	hasUncategorized := false

	// 月份 -> 系列ID -> 值
	monthAmounts := make(map[string]map[string]float64)
	monthCounts := make(map[string]map[string]int)
	ensureMonth := func(m string) {
		if monthAmounts[m] == nil {
			monthAmounts[m] = make(map[string]float64)
			monthCounts[m] = make(map[string]int)
		}
	}

	var minMonth, maxMonth string
	trackRange := func(m string) {
		if minMonth == "" || m < minMonth {
			minMonth = m
		}
		if maxMonth == "" || m > maxMonth {
			maxMonth = m
		}
	}

	for i := range cases {
		cs := &cases[i]
		month := cs.CreatedAt.In(taipeiZone).Format("2006-01")
		ensureMonth(month)
		trackRange(month)

		items := itemsByCase[cs.ID]

		// 各項目原始價格（案件個別價優先，否則項目預設價）與加總
		type itemAmt struct {
			id    string
			title string
			order int
			price float64
		}
		perItem := make([]itemAmt, 0, len(items))
		var rawSum float64
		for _, cci := range items {
			price := cci.CollaborationItem.Price
			if cci.Price != nil {
				price = *cci.Price
			}
			title := cci.CollaborationItem.Title
			if title == "" {
				title = "未命名項目"
			}
			perItem = append(perItem, itemAmt{
				id:    cci.CollaborationItemID.String(),
				title: title,
				order: cci.CollaborationItem.Order,
				price: price,
			})
			rawSum += price
		}

		displayTotal := caseDisplayTotalValue(cs.AdjustedTotal, rawSum, cs.QuotedAmount)

		if len(perItem) == 0 {
			// 無合作項目：金額與專案數皆歸入未分類
			hasUncategorized = true
			monthAmounts[month][uncategorizedSeriesID] += displayTotal
			monthCounts[month][uncategorizedSeriesID]++
			continue
		}

		// 專案數：案件對它的每個合作項目各計 1（與金額一致地歸屬到實際項目）
		for _, it := range perItem {
			if _, ok := seriesInfo[it.id]; !ok {
				seriesInfo[it.id] = seriesMeta{title: it.title, order: it.order}
			}
			monthCounts[month][it.id]++
		}

		// 金額：項目有價格時按原始價格等比分配（手動調整總價時仍保留各項目比例）；
		// 若項目皆為 0 元但案件仍有總額，則平均分攤到各項目，避免與專案數的歸類不一致。
		if rawSum > 0 {
			for _, it := range perItem {
				monthAmounts[month][it.id] += it.price * displayTotal / rawSum
			}
		} else if displayTotal > 0 {
			share := displayTotal / float64(len(perItem))
			for _, it := range perItem {
				monthAmounts[month][it.id] += share
			}
		}
	}

	// 系列清單：先依合作項目 order、再依標題排序；未分類置於最後
	series := make([]AnalyticsSeries, 0, len(seriesInfo)+1)
	for id, meta := range seriesInfo {
		series = append(series, AnalyticsSeries{ID: id, Title: meta.title})
	}
	sort.Slice(series, func(i, j int) bool {
		mi, mj := seriesInfo[series[i].ID], seriesInfo[series[j].ID]
		if mi.order != mj.order {
			return mi.order < mj.order
		}
		return mi.title < mj.title
	})
	if hasUncategorized {
		series = append(series, AnalyticsSeries{ID: uncategorizedSeriesID, Title: "未分類"})
	}

	// 連續月份範圍（含無資料的空月份，讓 X 軸不跳號）
	start, _ := time.ParseInLocation("2006-01", minMonth, taipeiZone)
	end, _ := time.ParseInLocation("2006-01", maxMonth, taipeiZone)
	months := make([]AnalyticsMonthPoint, 0)
	for d := start; !d.After(end); d = d.AddDate(0, 1, 0) {
		key := d.Format("2006-01")
		amts := monthAmounts[key]
		if amts == nil {
			amts = map[string]float64{}
		}
		cnts := monthCounts[key]
		if cnts == nil {
			cnts = map[string]int{}
		}
		var totalAmount float64
		for _, v := range amts {
			totalAmount += v
		}
		var totalCount int
		for _, v := range cnts {
			totalCount += v
		}
		months = append(months, AnalyticsMonthPoint{
			Month:       key,
			Amounts:     amts,
			Counts:      cnts,
			TotalAmount: totalAmount,
			TotalCount:  totalCount,
		})
	}

	c.JSON(http.StatusOK, CollaborationItemAnalyticsResponse{Series: series, Months: months})
}
