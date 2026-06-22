package utils

import (
	"testing"
	"time"
)

// 參考日曆（2026年6月）：
//   6/20 週六  6/21 週日  6/22 週一  6/23 週二  6/24 週三
//   6/25 週四  6/26 週五  6/27 週六  6/28 週日  6/29 週一  6/30 週二

// d 解析 "YYYY-MM-DD" 為 UTC 零點時間，測試輔助用。
func d(s string) time.Time {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}

// cleanConfig 只排除週末、無任何國定假日或補班日。
func cleanConfig() DayCalcConfig {
	return DefaultDayCalcConfig()
}

// testConfig 含一個國定假日（週二 6/23）與一個補班日（週六 6/27），以及跨年假日。
func testConfig() DayCalcConfig {
	cfg := DefaultDayCalcConfig()
	cfg.Holidays["2026-06-23"] = true       // 週二，假日
	cfg.MakeupWorkdays["2026-06-27"] = true // 週六，補班
	cfg.Holidays["2027-01-01"] = true       // 跨年假日
	return cfg
}

func TestIsWorkingDay(t *testing.T) {
	cfg := testConfig()
	cases := []struct {
		date string
		want bool
		desc string
	}{
		{"2026-06-22", true, "週一一般工作日"},
		{"2026-06-23", false, "週二國定假日"},
		{"2026-06-27", true, "週六補班日仍須上班"},
		{"2026-06-28", false, "週日"},
		{"2026-06-20", false, "週六（無補班）"},
		{"2027-01-01", false, "跨年假日"},
	}
	for _, c := range cases {
		if got := cfg.IsWorkingDay(d(c.date)); got != c.want {
			t.Errorf("IsWorkingDay(%s) = %v, want %v (%s)", c.date, got, c.want, c.desc)
		}
	}
}

func TestAddWorkingDaysInclusive(t *testing.T) {
	cfg := cleanConfig()
	cases := []struct {
		start string
		n     int
		want  string
		desc  string
	}{
		{"2026-06-22", 0, "2026-06-22", "工期 1 工作日（n=0）= 起始日"},
		{"2026-06-22", 2, "2026-06-24", "週一起算 3 工作日（含起始日）→ 週三"},
		{"2026-06-26", 2, "2026-06-30", "週五起算 3 工作日跨週末 → 下週二"},
		{"2026-06-20", 0, "2026-06-22", "週六起算先正規化到下週一"},
	}
	for _, c := range cases {
		got := cfg.AddWorkingDays(d(c.start), c.n).Format(dateLayout)
		if got != c.want {
			t.Errorf("AddWorkingDays(%s, %d) = %s, want %s (%s)", c.start, c.n, got, c.want, c.desc)
		}
	}
}

func TestAddWorkingDaysSkipsHoliday(t *testing.T) {
	cfg := testConfig()
	// 週一 6/22 起算 3 工作日，週二 6/23 為假日 → 6/22, 6/24, 6/25
	got := cfg.AddWorkingDays(d("2026-06-22"), 2).Format(dateLayout)
	if got != "2026-06-25" {
		t.Errorf("跨假日工期計算 = %s, want 2026-06-25", got)
	}
}

func TestAddWorkingDaysCountsMakeupDay(t *testing.T) {
	cfg := testConfig()
	// 週五 6/26 起算 2 工作日，6/27 為補班日（算工作日）→ 6/26, 6/27
	got := cfg.AddWorkingDays(d("2026-06-26"), 1).Format(dateLayout)
	if got != "2026-06-27" {
		t.Errorf("補班日工期計算 = %s, want 2026-06-27", got)
	}
}

func TestEnsureWorkingDay(t *testing.T) {
	clean := cleanConfig()
	for in, want := range map[string]string{
		"2026-06-22": "2026-06-22", // 週一不變
		"2026-06-20": "2026-06-22", // 週六 → 下週一
		"2026-06-21": "2026-06-22", // 週日 → 下週一
	} {
		if got := clean.EnsureWorkingDay(d(in)).Format(dateLayout); got != want {
			t.Errorf("EnsureWorkingDay(%s) = %s, want %s", in, got, want)
		}
	}

	cfg := testConfig()
	for in, want := range map[string]string{
		"2026-06-23": "2026-06-24", // 假日 → 隔日週三
		"2026-06-27": "2026-06-27", // 補班日不變
	} {
		if got := cfg.EnsureWorkingDay(d(in)).Format(dateLayout); got != want {
			t.Errorf("EnsureWorkingDay(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestNextWorkingDay(t *testing.T) {
	clean := cleanConfig()
	for in, want := range map[string]string{
		"2026-06-26": "2026-06-29", // 週五 → 下週一（跳週末）
		"2026-06-22": "2026-06-23", // 週一 → 週二
		"2026-06-19": "2026-06-22", // 週五 → 下週一
	} {
		if got := clean.NextWorkingDay(d(in)).Format(dateLayout); got != want {
			t.Errorf("NextWorkingDay(%s) = %s, want %s", in, got, want)
		}
	}

	cfg := testConfig()
	if got := cfg.NextWorkingDay(d("2026-06-22")).Format(dateLayout); got != "2026-06-24" {
		t.Errorf("NextWorkingDay(週一，跳週二假日) = %s, want 2026-06-24", got)
	}
}

func TestWorkingDaysBetween(t *testing.T) {
	clean := cleanConfig()
	cleanCases := []struct {
		a, b string
		want int
		desc string
	}{
		{"2026-06-22", "2026-06-22", 0, "同日"},
		{"2026-06-22", "2026-06-25", 3, "週一→週四 = 3 個工作日"},
		{"2026-06-26", "2026-06-29", 1, "週五→下週一，跳週末 = 1"},
		{"2026-06-25", "2026-06-22", -3, "反向為負"},
	}
	for _, c := range cleanCases {
		if got := clean.WorkingDaysBetween(d(c.a), d(c.b)); got != c.want {
			t.Errorf("WorkingDaysBetween(%s,%s) = %d, want %d (%s)", c.a, c.b, got, c.want, c.desc)
		}
	}

	cfg := testConfig()
	// 週一→週三，跳週二假日 = 1 個工作日
	if got := cfg.WorkingDaysBetween(d("2026-06-22"), d("2026-06-24")); got != 1 {
		t.Errorf("WorkingDaysBetween 跨假日 = %d, want 1", got)
	}
}

func TestCalendarModeDegradesToCalendarDays(t *testing.T) {
	cfg := DefaultDayCalcConfig()
	cfg.Mode = DayCalcModeCalendar
	cfg.Holidays["2026-06-23"] = true // calendar 模式應忽略假日

	if got := cfg.AddWorkingDays(d("2026-06-26"), 3).Format(dateLayout); got != "2026-06-29" {
		t.Errorf("calendar AddWorkingDays = %s, want 2026-06-29", got)
	}
	if got := cfg.WorkingDaysBetween(d("2026-06-22"), d("2026-06-25")); got != 3 {
		t.Errorf("calendar WorkingDaysBetween = %d, want 3", got)
	}
	if !cfg.IsWorkingDay(d("2026-06-21")) { // 週日在 calendar 模式仍算工作日
		t.Errorf("calendar IsWorkingDay(週日) = false, want true")
	}
}
