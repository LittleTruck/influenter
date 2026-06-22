package utils

import "time"

// dateLayout 是工作日計算統一使用的日期鍵格式。
const dateLayout = "2006-01-02"

// DayCalcMode 控制天數如何換算為實際日期。
type DayCalcMode string

const (
	// DayCalcModeWorking 工作日模式：排除週末與國定假日（補班日除外）。
	DayCalcModeWorking DayCalcMode = "working"
	// DayCalcModeCalendar 日曆日模式：所有日期一律視為可用（行為等同舊邏輯）。
	DayCalcModeCalendar DayCalcMode = "calendar"
)

// DayCalcConfig 描述「天數 → 日期」的換算規則。
//
// 這是整個流程排程（案件階段、流程範本）天數計算的單一真相來源；
// 前端 utils/workdays.ts 鏡像相同語意。
//
// 目前由 DefaultDayCalcConfig() 提供全域預設（工作日 + 國定假日），
// 並預留 per-account 覆寫的 drop-in 介面：未來 resolver 可依帳號改寫
// Mode（切回日曆日）、清空 Holidays（不排除國定假日）或加入自訂假日，
// 而所有呼叫端（AddWorkingDays / NextWorkingDay …）完全不需更動。
type DayCalcConfig struct {
	Mode DayCalcMode
	// Weekends 視為休假的星期（預設週六、週日）。
	Weekends map[time.Weekday]bool
	// Holidays "YYYY-MM-DD" → 國定假日（休假，不算工作日）。
	Holidays map[string]bool
	// MakeupWorkdays "YYYY-MM-DD" → 補班日（台灣特有：即使落在週末仍須上班）。
	MakeupWorkdays map[string]bool
}

// DefaultWeekends 回傳預設週末集合（週六、週日）。
func DefaultWeekends() map[time.Weekday]bool {
	return map[time.Weekday]bool{time.Saturday: true, time.Sunday: true}
}

// DefaultDayCalcConfig 回傳全域預設設定：工作日模式、週末為六日、
// 假日集合為空（需由 resolver 從 national_holidays 載入）。
func DefaultDayCalcConfig() DayCalcConfig {
	return DayCalcConfig{
		Mode:           DayCalcModeWorking,
		Weekends:       DefaultWeekends(),
		Holidays:       map[string]bool{},
		MakeupWorkdays: map[string]bool{},
	}
}

// toDay 取 t 的「當地日曆日」並正規化為 UTC 零點，避免時分秒與時區位移
// 影響日期比較與步進。
func toDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// dateKey 取 t 的 YYYY-MM-DD 日期鍵。
func dateKey(t time.Time) string {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Format(dateLayout)
}

// IsWorkingDay 判斷 t 是否為工作日。
//
// 規則優先序：補班日 > 國定假日 > 週末。
// calendar 模式下所有日期皆為工作日。
func (c DayCalcConfig) IsWorkingDay(t time.Time) bool {
	if c.Mode == DayCalcModeCalendar {
		return true
	}
	key := dateKey(t)
	if c.MakeupWorkdays[key] {
		return true
	}
	if c.Holidays[key] {
		return false
	}
	if c.Weekends[t.Weekday()] {
		return false
	}
	return true
}

// EnsureWorkingDay 將 t 正規化為日期；若不是工作日則順延到下一個工作日。
// calendar 模式僅做日期正規化。
func (c DayCalcConfig) EnsureWorkingDay(t time.Time) time.Time {
	d := toDay(t)
	if c.Mode == DayCalcModeCalendar {
		return d
	}
	for !c.IsWorkingDay(d) {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

// NextWorkingDay 回傳嚴格晚於 t 的第一個工作日。
// 用於串聯排程「下一階段從前一階段結束日之後開始」。
func (c DayCalcConfig) NextWorkingDay(t time.Time) time.Time {
	d := toDay(t).AddDate(0, 0, 1)
	if c.Mode == DayCalcModeCalendar {
		return d
	}
	for !c.IsWorkingDay(d) {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

// AddWorkingDays 從 start 起算往後加 n 個工作日。
//
// start 會先正規化到工作日（EnsureWorkingDay）。採「含起始日」語意：
// 工期 N 個工作日的結束日 = AddWorkingDays(start, N-1)，故 n=0 時回傳正規化後的 start。
// n 應為非負；calendar 模式退化為 AddDate(0, 0, n)。
func (c DayCalcConfig) AddWorkingDays(start time.Time, n int) time.Time {
	if c.Mode == DayCalcModeCalendar {
		return toDay(start).AddDate(0, 0, n)
	}
	d := c.EnsureWorkingDay(start)
	for remaining := n; remaining > 0; {
		d = d.AddDate(0, 0, 1)
		if c.IsWorkingDay(d) {
			remaining--
		}
	}
	return d
}

// WorkingDaysBetween 回傳 a 到 b 之間的工作日數（不含 a 當天、含 b 當天）。
// b 早於 a 時回傳負值。用於「距離截止還有幾個工作日」的倒數計算。
// calendar 模式等同日曆天差。
func (c DayCalcConfig) WorkingDaysBetween(a, b time.Time) int {
	da, db := toDay(a), toDay(b)
	if da.Equal(db) {
		return 0
	}
	sign := 1
	if db.Before(da) {
		da, db = db, da
		sign = -1
	}
	count := 0
	for d := da; d.Before(db); {
		d = d.AddDate(0, 0, 1)
		if c.IsWorkingDay(d) {
			count++
		}
	}
	return sign * count
}
