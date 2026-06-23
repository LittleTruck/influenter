package openai

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    []float64
		b    []float64
		want float64
	}{
		{"identical", []float64{1, 2, 3}, []float64{1, 2, 3}, 1},
		{"orthogonal", []float64{1, 0}, []float64{0, 1}, 0},
		{"opposite", []float64{1, 0}, []float64{-1, 0}, -1},
		{"scaled same direction", []float64{1, 1}, []float64{2, 2}, 1},
		{"length mismatch", []float64{1, 2}, []float64{1, 2, 3}, 0},
		{"empty", []float64{}, []float64{}, 0},
		{"zero vector", []float64{0, 0}, []float64{1, 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarity(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("CosineSimilarity(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestEmailTypeLabel(t *testing.T) {
	cases := map[string]string{
		EmailTypeFirstInviteFull: "首次邀約（資訊完整）",
		EmailTypeBarter:          "互惠／公關品邀約",
		EmailTypeNegotiation:     "議價討論",
		"unknown_code":           "",
		"":                       "",
	}
	for code, want := range cases {
		if got := EmailTypeLabel(code); got != want {
			t.Errorf("EmailTypeLabel(%q) = %q, want %q", code, got, want)
		}
	}
}
