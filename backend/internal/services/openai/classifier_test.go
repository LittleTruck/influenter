package openai

import (
	"encoding/json"
	"testing"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/rs/zerolog"
)

// TestClassifyEmail 測試郵件分類功能
// 注意：這是單元測試，不實際調用 OpenAI API

func TestGetCategoryDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		category EmailCategory
		expected string
	}{
		{"Collaboration", CategoryCollaboration, "合作邀約"},
		{"Payment", CategoryPayment, "付款相關"},
		{"Confirmation", CategoryConfirmation, "確認郵件"},
		{"Inquiry", CategoryInquiry, "詢問"},
		{"Social", CategorySocial, "社交"},
		{"Newsletter", CategoryNewsletter, "訂閱/電子報"},
		{"Notification", CategoryNotification, "通知"},
		{"Spam", CategorySpam, "垃圾郵件"},
		{"Other", CategoryOther, "其他"},
		{"Unknown", EmailCategory("unknown"), "未知"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCategoryDisplayName(tt.category)
			if result != tt.expected {
				t.Errorf("GetCategoryDisplayName(%s) = %s, want %s", tt.category, result, tt.expected)
			}
		})
	}
}

func TestGetCategoryColor(t *testing.T) {
	tests := []struct {
		name     string
		category EmailCategory
		expected string
	}{
		{"Collaboration", CategoryCollaboration, "primary"},
		{"Payment", CategoryPayment, "success"},
		{"Confirmation", CategoryConfirmation, "info"},
		{"Inquiry", CategoryInquiry, "warning"},
		{"Social", CategorySocial, "secondary"},
		{"Newsletter", CategoryNewsletter, "info"},
		{"Notification", CategoryNotification, "warning"},
		{"Spam", CategorySpam, "error"},
		{"Other", CategoryOther, "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCategoryColor(tt.category)
			if result != tt.expected {
				t.Errorf("GetCategoryColor(%s) = %s, want %s", tt.category, result, tt.expected)
			}
		})
	}
}

func TestIsHighPriorityCategory(t *testing.T) {
	tests := []struct {
		name     string
		category EmailCategory
		expected bool
	}{
		{"Collaboration", CategoryCollaboration, true},
		{"Payment", CategoryPayment, true},
		{"Confirmation", CategoryConfirmation, false},
		{"Inquiry", CategoryInquiry, true},
		{"Social", CategorySocial, false},
		{"Newsletter", CategoryNewsletter, false},
		{"Notification", CategoryNotification, false},
		{"Spam", CategorySpam, false},
		{"Other", CategoryOther, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHighPriorityCategory(tt.category)
			if result != tt.expected {
				t.Errorf("IsHighPriorityCategory(%s) = %v, want %v", tt.category, result, tt.expected)
			}
		})
	}
}

func TestParseClassificationFromContent(t *testing.T) {
	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	tests := []struct {
		name         string
		content      string
		expectedCat  EmailCategory
		expectedConf float64
	}{
		{
			name:         "contains collaboration",
			content:      "This is a collaboration email about brand partnership",
			expectedCat:  CategoryCollaboration,
			expectedConf: 0.7,
		},
		{
			name:         "contains payment",
			content:      "This is about polemics ", // Note: 使用包含 " polemics " 的文字
			expectedCat:  CategoryPayment,
			expectedConf: 0.7,
		},
		{
			name:         "contains spam",
			content:      "This is spam content",
			expectedCat:  CategorySpam,
			expectedConf: 0.7,
		},
		{
			name:         "no match",
			content:      "This is a regular email without any keywords",
			expectedCat:  CategoryOther,
			expectedConf: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.parseClassificationFromContent(tt.content)

			if err != nil {
				t.Fatalf("parseClassificationFromContent() error = %v", err)
			}

			if result.Category != tt.expectedCat {
				t.Errorf("Category = %v, want %v", result.Category, tt.expectedCat)
			}

			if result.Confidence != tt.expectedConf {
				t.Errorf("Confidence = %v, want %v", result.Confidence, tt.expectedConf)
			}
		})
	}
}

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		substr string
		want   bool
	}{
		{"exact match", "hello world", "hello world", true},
		{"contains substring", "hello world", "world", true},
		{"start match", "hello world", "hello", true},
		{"end match", "hello world", "world", true},
		{"no match", "hello world", "xyz", false},
		{"case insensitive", "Hello World", "hello", true},
		{"case insensitive 2", "HELLO WORLD", "world", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.text, tt.substr)
			if result != tt.want {
				t.Errorf("contains(%s, %s) = %v, want %v", tt.text, tt.substr, result, tt.want)
			}
		})
	}
}

// TestParseClassificationResult 測試解析分類結果
func TestParseClassificationResult(t *testing.T) {
	tests := []struct {
		name         string
		jsonResponse string
		wantCategory EmailCategory
		wantConf     float64
		wantErr      bool
	}{
		{
			name:         "valid collaboration response",
			jsonResponse: `{"category":"collaboration","confidence":0.95,"reason":"Brand partnership invitation"}`,
			wantCategory: CategoryCollaboration,
			wantConf:     0.95,
			wantErr:      false,
		},
		{
			name:         "valid payment response",
			jsonResponse: `{"category":"payment","confidence":0.8,"reason":"Invoice notification"}`,
			wantCategory: CategoryPayment,
			wantConf:     0.8,
			wantErr:      false,
		},
		{
			name:         "invalid JSON",
			jsonResponse: `{invalid json}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result EmailClassification
			err := json.Unmarshal([]byte(tt.jsonResponse), &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Category != tt.wantCategory {
					t.Errorf("Category = %v, want %v", result.Category, tt.wantCategory)
				}
				if result.Confidence != tt.wantConf {
					t.Errorf("Confidence = %v, want %v", result.Confidence, tt.wantConf)
				}
			}
		})
	}
}

// Helper functions for classifier tests
func getClassifierTestConfig() config.Config {
	return config.Config{
		OpenAI: config.OpenAIConfig{
			APIKey:    "test-api-key",
			Model:     "gpt-4o-mini",
			MaxTokens: 2000,
		},
	}
}

func getClassifierMockLogger() *zerolog.Logger {
	logger := zerolog.Nop()
	return &logger
}

// Benchmark tests
func BenchmarkContains(b *testing.B) {
	text := "This is a long email content about collaboration and partnership opportunities"
	substr := "collaboration"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contains(text, substr)
	}
}

func BenchmarkGetCategoryDisplayName(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCategoryDisplayName(CategoryCollaboration)
	}
}
