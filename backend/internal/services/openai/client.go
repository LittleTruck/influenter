package openai

import (
	"context"
	"fmt"
	"time"

	"github.com/designcomb/influenter-backend/internal/config"
	"github.com/rs/zerolog"
	openai "github.com/sashabaranov/go-openai"
)

// Service OpenAI 服務
type Service struct {
	client     *openai.Client
	config     config.OpenAIConfig
	userAPIKey string // 使用者自己設定的 API Key（可選）
	logger     *zerolog.Logger
}

// NewService 建立新的 OpenAI Service
// 如果提供了 userAPIKey，則使用使用者的 API Key，否則使用系統設定的
func NewService(cfg config.Config, logger *zerolog.Logger, userAPIKey string) *Service {
	apiKey := cfg.OpenAI.APIKey
	if userAPIKey != "" {
		apiKey = userAPIKey
	}

	client := openai.NewClient(apiKey)

	service := &Service{
		client:     client,
		config:     cfg.OpenAI,
		userAPIKey: userAPIKey,
		logger:     logger,
	}

	return service
}

// TestConnection 測試 OpenAI API 連線
func (s *Service) TestConnection(ctx context.Context) error {
	s.logger.Info().Msg("Testing OpenAI connection")

	req := openai.ChatCompletionRequest{
		Model:     s.config.Model,
		MaxTokens: 5,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "test",
			},
		},
	}

	_, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		s.logger.Error().Err(err).Msg("OpenAI connection test failed")
		return fmt.Errorf("failed to connect to OpenAI: %w", err)
	}

	s.logger.Info().Msg("OpenAI connection test succeeded")
	return nil
}

// GetModel 取得目前使用的模型
func (s *Service) GetModel() string {
	return s.config.Model
}

// GetMaxTokens 取得最大 token 限制
func (s *Service) GetMaxTokens() int {
	return s.config.MaxTokens
}

// buildPrompt 建立完整的 prompt
func (s *Service) buildPrompt(systemPrompt string, userPrompt string) []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}
}

// callAPI 呼叫 OpenAI API
func (s *Service) callAPI(ctx context.Context, messages []openai.ChatCompletionMessage, functions []openai.FunctionDefinition) (*openai.ChatCompletionResponse, error) {
	startTime := time.Now()

	req := openai.ChatCompletionRequest{
		Model:     s.config.Model,
		Messages:  messages,
		MaxTokens: s.config.MaxTokens,
		Functions: functions,
	}

	// 如果有定義 functions，啟用 function calling
	if len(functions) > 0 {
		req.FunctionCall = &openai.FunctionCall{
			Name: functions[0].Name,
		}
	}

	s.logger.Info().
		Str("model", s.config.Model).
		Int("messages", len(messages)).
		Msg("Calling OpenAI API")

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		s.logger.Error().
			Err(err).
			Dur("duration", time.Since(startTime)).
			Msg("OpenAI API call failed")
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	duration := time.Since(startTime)
	tokensUsed := resp.Usage.TotalTokens

	s.logger.Info().
		Dur("duration", duration).
		Int("tokens", tokensUsed).
		Msg("OpenAI API call succeeded")

	return &resp, nil
}

// CalculateCost 計算 API 成本
func (s *Service) CalculateCost(tokensUsed int, model string) float64 {
	// OpenAI 定價 (美元 per 1K tokens)
	// 注意：這些價格可能會變動，建議定期更新
	pricing := map[string]map[string]float64{
		"gpt-4o": {
			"prompt":     0.005,
			"completion": 0.015,
		},
		"gpt-4o-mini": {
			"prompt":     0.00015,
			"completion": 0.0006,
		},
		"gpt-4-turbo": {
			"prompt":     0.01,
			"completion": 0.03,
		},
		"gpt-4": {
			"prompt":     0.03,
			"completion": 0.06,
		},
		"gpt-3.5-turbo": {
			"prompt":     0.0005,
			"completion": 0.0015,
		},
	}

	modelPricing, exists := pricing[model]
	if !exists {
		// 預設使用 gpt-4o-mini 的價格
		modelPricing = pricing["gpt-4o-mini"]
	}

	// 假設平均分配 prompt 和 completion tokens
	avgPrice := (modelPricing["prompt"] + modelPricing["completion"]) / 2
	cost := float64(tokensUsed) / 1000.0 * avgPrice

	return cost
}

// RecordTokenUsage 記錄 token 使用情況
func (s *Service) RecordTokenUsage(usage TokenUsage) {
	s.logger.Info().
		Str("user_id", usage.UserID).
		Str("email_id", usage.EmailID).
		Str("model", usage.Model).
		Int("total_tokens", usage.TotalTokens).
		Float64("cost_usd", usage.CostUSD).
		Msg("Token usage recorded")

	// TODO: 將 usage 儲存到資料庫
}

// TruncateContent 截斷內容以避免超過 token 限制
func (s *Service) TruncateContent(content string, maxChars int) string {
	if len(content) <= maxChars {
		return content
	}

	// 粗略估算：假設 1 token ≈ 4 字元
	truncated := content[:maxChars]
	return truncated + "...\n[內容已截斷]"
}

// ValidateAPIKey 驗證 API Key 格式
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key is empty")
	}

	// OpenAI API Key 通常以 "sk-" keyword 開頭
	if len(apiKey) < 20 {
		return fmt.Errorf("API key seems too short")
	}

	return nil
}

// LogAnalysisStart 記錄分析開始
func (s *Service) logAnalysisStart(emailID string, operation string) {
	s.logger.Info().
		Str("email_id", emailID).
		Str("operation", operation).
		Str("model", s.config.Model).
		Msg("Starting AI analysis")
}

// LogAnalysisComplete 記錄分析完成
func (s *Service) logAnalysisComplete(emailID string, operation string, duration time.Duration, tokensUsed int) {
	s.logger.Info().
		Str("email_id", emailID).
		Str("operation", operation).
		Dur("duration", duration).
		Int("tokens", tokensUsed).
		Msg("AI analysis completed")
}

// LogAnalysisError 記錄分析錯誤
func (s *Service) logAnalysisError(emailID string, operation string, err error) {
	s.logger.Error().
		Str("email_id", emailID).
		Str("operation", operation).
		Err(err).
		Msg("AI analysis failed")
}

// Helper function to safely get string from response
func safeString(choice *openai.ChatCompletionChoice) string {
	if choice.Message.FunctionCall != nil {
		return choice.Message.FunctionCall.Arguments
	}
	return choice.Message.Content
}
