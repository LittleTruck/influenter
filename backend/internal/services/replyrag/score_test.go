package replyrag

import (
	"math"
	"testing"
)

func TestScoreCandidate(t *testing.T) {
	const cos = 0.5
	tests := []struct {
		name       string
		brandMatch bool
		typeMatch  bool
		isGood     bool
		want       float64
	}{
		{"base only", false, false, false, cos},
		{"brand match", true, false, false, cos + brandBoost},
		{"type match", false, true, false, cos + typeBoost},
		{"good example", false, false, true, cos + goodBoost},
		{"all boosts", true, true, true, cos + brandBoost + typeBoost + goodBoost},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreCandidate(cos, tt.brandMatch, tt.typeMatch, tt.isGood)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("scoreCandidate = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScoreCandidateOrdering(t *testing.T) {
	// 優質範例（低相似度）應勝過普通候選（中相似度），體現「有標記時優先」
	good := scoreCandidate(0.30, false, false, true)
	normal := scoreCandidate(0.55, false, false, false)
	if good <= normal {
		t.Errorf("good example (%v) should outrank normal candidate (%v)", good, normal)
	}
	// 同品牌應勝過跨品牌但相似度稍高者
	sameBrand := scoreCandidate(0.40, true, false, false)
	crossBrand := scoreCandidate(0.50, false, false, false)
	if sameBrand <= crossBrand {
		t.Errorf("same-brand (%v) should outrank cross-brand (%v)", sameBrand, crossBrand)
	}
}

func TestShouldInclude(t *testing.T) {
	tests := []struct {
		name       string
		cos        float64
		brandMatch bool
		isGood     bool
		want       bool
	}{
		{"below threshold cross-brand", 0.05, false, false, false},
		{"above threshold cross-brand", 0.20, false, false, true},
		{"low cos but brand match", 0.01, true, false, true},
		{"low cos but good example", 0.01, false, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldInclude(tt.cos, tt.brandMatch, tt.isGood); got != tt.want {
				t.Errorf("shouldInclude(%v, %v, %v) = %v, want %v", tt.cos, tt.brandMatch, tt.isGood, got, tt.want)
			}
		})
	}
}

func TestEqualFoldTrim(t *testing.T) {
	if !equalFoldTrim(" Nike ", "nike") {
		t.Error("expected case/space-insensitive match")
	}
	if equalFoldTrim("", "nike") {
		t.Error("empty should not match")
	}
	if equalFoldTrim("a", "") {
		t.Error("empty should not match")
	}
}

func TestTruncateRunes(t *testing.T) {
	// 中文字應以 rune 計數，不切斷多位元組字元
	s := "你好世界測試"
	got := truncateRunes(s, 3)
	if got != "你好世…" {
		t.Errorf("truncateRunes = %q, want %q", got, "你好世…")
	}
	if truncateRunes("abc", 10) != "abc" {
		t.Error("should not truncate when under limit")
	}
	if truncateRunes("abc", 0) != "" {
		t.Error("zero limit should return empty")
	}
}
