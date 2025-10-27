package openai

import (
	"testing"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/rs/zerolog"
)

// MockOpenAIClient 是 OpenAI 客戶端的 mock 實作
type MockOpenAIClient struct {
	Responses map[string]interface{}
	Errors    map[string]error
}

// getMockLogger 返回測試用的 logger
func getMockLogger() *zerolog.Logger {
	logger := zerolog.Nop()
	return &logger
}

// getTestConfig 返回測試用的配置
func getTestConfig() config.Config {
	return config.Config{
		OpenAI: config.OpenAIConfig{
			APIKey:    "test-api-key",
			Model:     "gpt-4o-mini",
			MaxTokens: 2000,
		},
	}
}

func TestNewService(t *testing.T) {
	cfg := getTestConfig()
	logger := getMockLogger()

	service := NewService(cfg, logger, "")

	if service == nil {
		t.Fatal("Expected service to be non-nil")
	}

	if service.client == nil {
		t.Error("Expected client to be non-nil")
	}

	if service.config.Model != "gpt-4o-mini" {
		t.Errorf("Expected model gpt-4o-mini, got %s", service.config.Model)
	}
}

func TestNewService_WithUserAPIKey(t *testing.T) {
	cfg := getTestConfig()
	logger := getMockLogger()
	userKey := "user-api-key"

	service := NewService(cfg, logger, userKey)

	if service == nil {
		t.Fatal("Expected service to be non-nil")
	}

	if service.userAPIKey != userKey {
		t.Errorf("Expected userAPIKey to be %s, got %s", userKey, service.userAPIKey)
	}
}

func TestGetModel(t *testing.T) {
	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	model := service.GetModel()

	if model != "gpt-4o-mini" {
		t.Errorf("Expected model gpt-4o-mini, got %s", model)
	}
}

func TestGetMaxTokens(t *testing.T) {
	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	maxTokens := service.GetMaxTokens()

	if maxTokens != 2000 {
		t.Errorf("Expected maxTokens 2000, got %d", maxTokens)
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name        string
		model       string
		tokensUsed  int
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "gpt-4o-mini low tokens",
			model:       "gpt-4o-mini",
			tokensUsed:  1000,
			expectedMin: 0.0001,
			expectedMax: 0.001,
		},
		{
			name:        "gpt-4o low tokens",
			model:       "gpt-4o",
			tokensUsed:  1000,
			expectedMin: 0.001,
			expectedMax: 0.05,
		},
		{
			name:        "gpt-4o-mini high tokens",
			model:       "gpt-4o-mini",
			tokensUsed:  10000,
			expectedMin: 0.001,
			expectedMax: 0.01,
		},
	}

	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := service.CalculateCost(tt.tokensUsed, tt.model)

			if cost < tt.expectedMin || cost > tt.expectedMax {
				t.Errorf("Expected cost between %f and %f, got %f", tt.expectedMin, tt.expectedMax, cost)
			}
		})
	}
}

func TestTruncateContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxChars  int
		shouldEnd string
	}{
		{
			name:      "content shorter than max",
			content:   "Short content",
			maxChars:  100,
			shouldEnd: "Short content",
		},
		{
			name:      "content exactly at max",
			content:   "This is exactly 50 characters long, so it passes!",
			maxChars:  50,
			shouldEnd: "",
		},
		{
			name:      "content longer than max",
			content:   "This is a very long content that exceeds the maximum character limit and should be truncated",
			maxChars:  50,
			shouldEnd: "...\n[內容已截斷]",
		},
		{
			name:      "empty content",
			content:   "",
			maxChars:  100,
			shouldEnd: "",
		},
	}

	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.TruncateContent(tt.content, tt.maxChars)

			if tt.content == "" {
				if result != "" {
					t.Errorf("Expected empty string for empty content")
				}
				return
			}

			if len(tt.content) <= tt.maxChars {
				if result != tt.content {
					t.Errorf("Expected %s, got %s", tt.content, result)
				}
			} else {
				if len(result) > tt.maxChars+len(tt.shouldEnd) {
					t.Errorf("Result too long: expected max %d, got %d", tt.maxChars+len(tt.shouldEnd), len(result))
				}
				if !hasSuffix(result, tt.shouldEnd) {
					t.Errorf("Expected result to end with %s, got %s", tt.shouldEnd, result)
				}
			}
		})
	}
}

func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid API key",
			apiKey:  "sk-proj-1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "too short API key",
			apiKey:  "sk-short",
			wantErr: true,
		},
		{
			name:    "valid length",
			apiKey:  "sk-proj-123456789012345",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.apiKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	systemPrompt := "You are a helpful assistant"
	userPrompt := "Hello, world!"

	messages := service.buildPrompt(systemPrompt, userPrompt)

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("Expected first message role to be 'system', got %s", messages[0].Role)
	}

	if messages[0].Content != systemPrompt {
		t.Errorf("Expected system prompt to be %s, got %s", systemPrompt, messages[0].Content)
	}

	if messages[1].Role != "user" {
		t.Errorf("Expected second message role to be 'user', got %s", messages[1].Role)
	}

	if messages[1].Content != userPrompt {
		t.Errorf("Expected user prompt to be %s, got %s", userPrompt, messages[1].Content)
	}
}

// Benchmark tests
func BenchmarkTruncateContent_Short(b *testing.B) {
	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	content := "This is a short content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.TruncateContent(content, 100)
	}
}

func BenchmarkTruncateContent_Long(b *testing.B) {
	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	content := "This is a very long content that will be truncated. " +
		"It contains many words and characters to test the performance of the truncation function. " +
		"We want to make sure it's efficient even with long strings."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.TruncateContent(content, 50)
	}
}

func BenchmarkCalculateCost(b *testing.B) {
	cfg := getTestConfig()
	logger := getMockLogger()
	service := NewService(cfg, logger, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateCost(1000, "gpt-4o-mini")
	}
}
