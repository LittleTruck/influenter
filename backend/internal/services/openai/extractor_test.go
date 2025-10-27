package openai

import (
	"testing"
	"time"
)

func TestExtractBrandName(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "email with name",
			text:     "John Doe <john@example.com>",
			expected: "John Doe",
		},
		{
			name:     "email without name",
			text:     "john@example.com",
			expected: "john@example.com",
		},
		{
			name:     "email with company",
			text:     "Brand Inc <contact@brand.com>",
			expected: "Brand Inc",
		},
		{
			name:     "plain name",
			text:     "Alice",
			expected: "Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractBrandName(tt.text)
			if result != tt.expected {
				t.Errorf("ExtractBrandName(%s) = %s, want %s", tt.text, result, tt.expected)
			}
		})
	}
}

func TestExtractAmount(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    float64
		wantErr bool
	}{
		{
			name:    "NTD with symbol",
			text:    "NT$5,000",
			want:    5000.0,
			wantErr: false,
		},
		{
			name:    "USD with symbol",
			text:    "USD 1000",
			want:    1000.0,
			wantErr: false,
		},
		{
			name:    "with comma",
			text:    "10,000",
			want:    10000.0,
			wantErr: false,
		},
		{
			name:    "with 萬",
			text:    "5萬",
			want:    50000.0,
			wantErr: false,
		},
		{
			name:    "invalid",
			text:    "not a number",
			want:    0.0,
			wantErr: true,
		},
		{
			name:    "empty",
			text:    "",
			want:    0.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractAmount(tt.text)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && *got != tt.want {
				t.Errorf("ExtractAmount() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func TestExtractPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Taiwan mobile",
			text:     "請聯絡 0912345678",
			expected: "0912345678",
		},
		{
			name:     "with spaces",
			text:     "電話: 0912345678",
			expected: "0912345678",
		},
		{
			name:     "US phone",
			text:     "Call us at 1234567890",
			expected: "1234567890",
		},
		{
			name:     "no phone",
			text:     "This is just regular text",
			expected: "",
		},
		{
			name:     "too short",
			text:     "123",
			expected: "",
		},
		{
			name:     "too long",
			text:     "123456789012345678901",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPhoneNumber(tt.text)
			if result != tt.expected {
				t.Errorf("ExtractPhoneNumber(%s) = %s, want %s", tt.text, result, tt.expected)
			}
		})
	}
}

func TestValidateExtractedInfo(t *testing.T) {
	tests := []struct {
		name    string
		info    ExtractedInfo
		wantLen int // 預期的錯誤數量
	}{
		{
			name: "valid info",
			info: ExtractedInfo{
				BrandName: "Brand Inc",
				Amount:    func() *float64 { v := 10000.0; return &v }(),
				Currency:  "TWD",
				DueDate:   func() *time.Time { t := time.Now().AddDate(0, 0, 7); return &t }(),
			},
			wantLen: 0,
		},
		{
			name: "missing brand name",
			info: ExtractedInfo{
				BrandName: "",
				Amount:    func() *float64 { v := 5000.0; return &v }(),
			},
			wantLen: 1,
		},
		{
			name: "negative amount",
			info: ExtractedInfo{
				BrandName: "Brand",
				Amount:    func() *float64 { v := -100.0; return &v }(),
			},
			wantLen: 1,
		},
		{
			name: "past due date",
			info: ExtractedInfo{
				BrandName: "Brand",
				DueDate:   func() *time.Time { t := time.Now().AddDate(0, 0, -7); return &t }(),
			},
			wantLen: 1,
		},
		{
			name: "multiple errors",
			info: ExtractedInfo{
				BrandName: "",
				Amount:    func() *float64 { v := -500.0; return &v }(),
				DueDate:   func() *time.Time { t := time.Now().AddDate(0, 0, -1); return &t }(),
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateExtractedInfo(&tt.info)
			if len(errors) != tt.wantLen {
				t.Errorf("ValidateExtractedInfo() errors length = %d, want %d, errors: %v", len(errors), tt.wantLen, errors)
			}
		})
	}
}

func TestSummarizeExtractedInfo(t *testing.T) {
	tests := []struct {
		name     string
		info     ExtractedInfo
		contains []string // 應該包含的字串
	}{
		{
			name: "complete info",
			info: ExtractedInfo{
				BrandName:   "Brand Inc",
				Amount:      func() *float64 { v := 10000.0; return &v }(),
				Currency:    "TWD",
				DueDate:     func() *time.Time { t := time.Now().AddDate(0, 0, 7); return &t }(),
				ContentType: "影片",
			},
			contains: []string{"品牌：Brand Inc", "金額：10000 TWD", "類型：影片"},
		},
		{
			name: "minimal info",
			info: ExtractedInfo{
				BrandName: "Brand",
			},
			contains: []string{"品牌：Brand"},
		},
		{
			name:     "empty info",
			info:     ExtractedInfo{},
			contains: []string{"無明顯資訊"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SummarizeExtractedInfo(&tt.info)

			for _, contain := range tt.contains {
				if !containsString(result, contain) {
					t.Errorf("SummarizeExtractedInfo() result %s should contain %s", result, contain)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOfSubstring(s, substr) >= 0))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Benchmark tests
func BenchmarkExtractBrandName(b *testing.B) {
	text := "Brand Company <contact@brand.com>"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractBrandName(text)
	}
}

func BenchmarkExtractAmount(b *testing.B) {
	text := "NT$50,000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractAmount(text)
	}
}

func BenchmarkExtractPhoneNumber(b *testing.B) {
	text := "請聯絡 0912345678 或 02-12345678"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractPhoneNumber(text)
	}
}

func BenchmarkValidateExtractedInfo(b *testing.B) {
	info := ExtractedInfo{
		BrandName: "Brand Inc",
		Amount:    func() *float64 { v := 10000.0; return &v }(),
		Currency:  "TWD",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateExtractedInfo(&info)
	}
}

func BenchmarkSummarizeExtractedInfo(b *testing.B) {
	info := ExtractedInfo{
		BrandName: "Brand Inc",
		Amount:    func() *float64 { v := 10000.0; return &v }(),
		Currency:  "TWD",
		DueDate:   func() *time.Time { t := time.Now().AddDate(0, 0, 7); return &t }(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SummarizeExtractedInfo(&info)
	}
}
