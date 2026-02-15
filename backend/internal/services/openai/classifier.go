package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	openai "github.com/sashabaranov/go-openai"
)

// ClassifyEmail 對郵件進行分類
func (s *Service) ClassifyEmail(ctx context.Context, req ClassifyEmailRequest) (*EmailClassification, error) {
	s.logger.Info().
		Str("from", req.From).
		Str("subject", req.Subject).
		Msg("Starting email classification")

	// 建立 system prompt
	systemPrompt := `你是一個專業的郵件分類助手，專門協助影響者（influencer）分類合作邀約相關的郵件。

你的任務是分析郵件內容，判斷郵件的主要類別。

分類類別說明：
1. **collaboration** (合作邀約) - 品牌邀請進行內容合作、代言、廣告等商業合作
2. **payment** (付款相關) - 收到款項、發票、付款提醒等
3. **confirmation** (確認郵件) - 確認合作細節、會議時間、檔期等
4. **inquiry** (詢問) - 詢問報價、檔期、合作可能性等
5. **social** (社交) - 粉絲來信、日常交流、非商業性郵件
6. **newsletter** (訂閱/電子報) - 品牌新聞、促銷資訊、一般訂閱郵件
7. **notification** (通知) - 系統通知、平台通知等
8. **spam** (垃圾郵件) - 明顯的垃圾郵件、詐騙郵件等
9. **other** (其他) - 無法明確歸類的郵件

請仔細分析郵件內容，並提供一個信心指標（0-1之間的小數），表示你對分類結果的把握程度。`

	// 建立 user prompt
	userPrompt := fmt.Sprintf(`請分析以下郵件：

**寄件者**: %s
**主旨**: %s
**內容**: %s

請判斷這個郵件屬於哪個類別，並提供你的信心指標和分類理由。`,
		req.From,
		req.Subject,
		s.TruncateContent(req.Body, 2000), // 限制長度避免超出 token 限制
	)

	// 定義 function calling
	functions := []openai.FunctionDefinition{
		{
			Name:        "classify_email",
			Description: "分類郵件並返回分類結果",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"category": {
						"type": "string",
						"enum": ["collaboration", "payment", "confirmation", "inquiry", "social", "newsletter", "notification", "spam", "other"],
						"description": "郵件的主要類別"
					},
					"confidence": {
						"type": "number",
						"minimum": 0,
						"maximum": 1,
						"description": "對分類結果的信心指標，0表示完全不確定，1表示非常確定"
					},
					"reason": {
						"type": "string",
						"description": "分類的理由，簡短說明為什麼歸類到這個類別"
					}
				},
				"required": ["category", "confidence", "reason"]
			}`),
		},
	}

	// 建立訊息
	messages := s.buildPrompt(systemPrompt, userPrompt)

	// 呼叫 API
	startTime := time.Now()
	resp, err := s.callAPI(ctx, messages, functions)
	if err != nil {
		s.logger.Error().
			Err(err).
			Dur("duration", time.Since(startTime)).
			Msg("Failed to classify email")
		return nil, fmt.Errorf("failed to classify email: %w", err)
	}

	// 解析結果
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := resp.Choices[0]
	if choice.Message.FunctionCall == nil {
		s.logger.Warn().
			Str("response", choice.Message.Content).
			Msg("No function call in response, trying to parse content")

		// 嘗試解析內容
		return s.parseClassificationFromContent(choice.Message.Content)
	}

	// 解析 function call arguments
	args := choice.Message.FunctionCall.Arguments
	var result EmailClassification

	if err := json.Unmarshal([]byte(args), &result); err != nil {
		s.logger.Error().
			Err(err).
			Str("arguments", args).
			Msg("Failed to parse function call arguments")
		return nil, fmt.Errorf("failed to parse classification result: %w", err)
	}

	s.logger.Info().
		Str("category", string(result.Category)).
		Float64("confidence", result.Confidence).
		Int("tokens", resp.Usage.TotalTokens).
		Dur("duration", time.Since(startTime)).
		Msg("Email classification completed")

	return &result, nil
}

// parseClassificationFromContent 從內容中解析分類結果（fallback 機制）
func (s *Service) parseClassificationFromContent(content string) (*EmailClassification, error) {
	// 這是一個簡單的解析邏輯，如果 AI 沒有返回 function call
	// 嘗試從文本中提取資訊
	result := EmailClassification{
		Category:   CategoryOther,
		Confidence: 0.5,
		Reason:     content,
	}

	// 嘗試找到類別關鍵字
	categoryMap := map[string]EmailCategory{
		"collaboration": CategoryCollaboration,
		" polemics ":    CategoryPayment,
		"confirmation":  CategoryConfirmation,
		"inquiry":       CategoryInquiry,
		"social":        CategorySocial,
		"newsletter":    CategoryNewsletter,
		"notification":  CategoryNotification,
		"spam":          CategorySpam,
		"other":         CategoryOther,
	}

	for keyword, category := range categoryMap {
		if contains(content, keyword) {
			result.Category = category
			result.Confidence = 0.7
			break
		}
	}

	return &result, nil
}

// contains 檢查字串是否包含子字串（不分大小寫）
func contains(text, substr string) bool {
	textLower := toLowerCase(text)
	substrLower := toLowerCase(substr)
	return len(textLower) >= len(substrLower) &&
		(textLower == substrLower ||
			len(textLower) > len(substrLower) &&
				(textLower[:len(substrLower)] == substrLower ||
					textLower[len(textLower)-len(substrLower):] == substrLower ||
					indexOf(textLower, substrLower) >= 0))
}

func toLowerCase(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetCategoryDisplayName 取得分類的中文顯示名稱
func GetCategoryDisplayName(category EmailCategory) string {
	names := map[EmailCategory]string{
		CategoryCollaboration: "合作邀約",
		CategoryPayment:       "付款相關",
		CategoryConfirmation:  "確認郵件",
		CategoryInquiry:       "詢問",
		CategorySocial:        "社交",
		CategoryNewsletter:    "訂閱/電子報",
		CategoryNotification:  "通知",
		CategorySpam:          "垃圾郵件",
		CategoryOther:         "其他",
	}

	if name, exists := names[category]; exists {
		return name
	}
	return "未知"
}

// GetCategoryColor 取得分類的顏色（用於 UI 顯示）
func GetCategoryColor(category EmailCategory) string {
	colors := map[EmailCategory]string{
		CategoryCollaboration: "primary",   // 藍色
		CategoryPayment:       "success",   // 綠色
		CategoryConfirmation:  "info",      // 淺藍
		CategoryInquiry:       "warning",   // 橘色
		CategorySocial:        "secondary", // 灰色
		CategoryNewsletter:    "info",      // 淺藍
		CategoryNotification:  "warning",   // 橘色
		CategorySpam:          "error",     // 紅色
		CategoryOther:         "default",   // 預設
	}

	if color, exists := colors[category]; exists {
		return color
	}
	return "default"
}

// IsHighPriorityCategory 判斷是否為高優先級分類
func IsHighPriorityCategory(category EmailCategory) bool {
	highPriority := []EmailCategory{
		CategoryCollaboration,
		CategoryPayment,
		CategoryInquiry,
	}

	for _, cp := range highPriority {
		if category == cp {
			return true
		}
	}
	return false
}

// IsCollaborationRelated 判斷郵件分類是否與工作/合作案件相關
// 合作相關：collaboration, payment, confirmation, inquiry
// 非合作：social, newsletter, notification, spam, other
func IsCollaborationRelated(category EmailCategory) bool {
	switch category {
	case CategoryCollaboration, CategoryPayment, CategoryConfirmation, CategoryInquiry:
		return true
	default:
		return false
	}
}

// logClassificationResult 記錄分類結果（用於除錯和統計）
func (s *Service) logClassificationResult(logger *zerolog.Logger, result *EmailClassification, tokensUsed int) {
	logger.Info().
		Str("category", string(result.Category)).
		Float64("confidence", result.Confidence).
		Str("reason", result.Reason).
		Int("tokens", tokensUsed).
		Msg("Classification result")
}
