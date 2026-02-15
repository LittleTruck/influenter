package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// ExtractInfo 從郵件中抽取結構化資訊
func (s *Service) ExtractInfo(ctx context.Context, req AnalyzeEmailRequest) (*ExtractedInfo, error) {
	s.logger.Info().
		Str("from", req.From).
		Str("subject", req.Subject).
		Msg("Starting info extraction")

	// 建立 system prompt
	systemPrompt := `你是一個專業的資訊抽取助手，專門協助影響者（influencer）從合作邀約郵件中抽取重要的結構化資訊。

你的任務是從郵件內容中識別並抽取以下資訊：
1. **品牌名稱** (brand_name) - 發送合作邀約的品牌或公司名稱
2. **聯絡人姓名** (contact_name) - 發送郵件的人的姓名
3. **聯絡人郵件** (contact_email) - 發送郵件的電子郵件地址
4. **聯絡電話** (contact_phone) - 聯絡電話號碼（如果有的話）
5. **金額** (amount) - 合作的預算或報酬金額
6. **幣別** (currency) - 金額使用的貨幣（函式預設為 "TWD"）
7. **截止日期** (due_date) - 專案的截止日期或回覆期限（ISO 8601 格式: YYYY-MM-DD）
8. **內容類型** (content_type) - 要求的內容類型（例如：影片、圖文、直播、貼文等）
9. **粉絲數** (follower_count) - 提及的粉絲數或影響力要求（例如："10萬以上"）
10. **預算範圍** (budget) - 如果提到預算範圍（例如："5萬-10萬"）
11. **專案詳情** (project_details) - 專案的詳細說明

注意事項：
- 如果某個欄位在郵件中沒有提到，請填 null 或空字串
- 日期請使用 ISO 8601 格式 (YYYY-MM-DD)
- 金額只需要數字部分，不需要包含貨幣符號或單位
- 幣別請使用標準的 ISO 4217 代碼（如 TWD, USD, EUR 等）
- 如果只有金額範圍，請將 budget 欄位填入範圍，amount 欄位填 null
- 專案詳情請簡要摘要（建議 100 字以內）`

	// 建立 user prompt
	userPrompt := fmt.Sprintf(`請從以下郵件中抽取所有相關資訊：

**寄件者**: %s
**收件者**: %s
**主旨**: %s
**日期**: %s
**內容**: %s

請仔細分析郵件，盡可能填寫所有能找到的資訊。如果某些欄位在郵件中沒有明確提到，請填 null 或空字串。`,
		req.From,
		strings.Join(req.To, ", "),
		req.Subject,
		req.Date.Format("2006-01-02 15:04:05"),
		s.TruncateContent(req.Body, 4000), // 較長一點以包含更多資訊
	)

	// 定義 function calling
	functions := []openai.FunctionDefinition{
		{
			Name:        "extract_info",
			Description: "從郵件中抽取結構化的合作資訊",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"brand_name": {
						"type": "string",
						"description": "品牌或公司名稱"
					},
					"contact_name": {
						"type": "string",
						"description": "聯絡人姓名"
					},
					"contact_email": {
						"type": "string",
						"description": "聯絡人電子郵件地址"
					},
					"contact_phone": {
						"type": "string",
						"description": "聯絡電話號碼"
					},
					"amount": {
						"type": "number",
						"description": "合作的預算或報酬金額（純數字）"
					},
					"currency": {
						"type": "string",
						"default": "TWD",
						"description": "金額使用的貨幣（ISO 4217 代碼）"
					},
					"due_date": {
						"type": "string",
						"description": "截止日期（ISO 8601 格式: YYYY-MM-DD）"
					},
					"content_type": {
						"type": "string",
						"description": "內容類型（例如：影片、圖文、直播等）"
					},
					"follower_count": {
						"type": "string",
						"description": "粉絲數或影響力要求"
					},
					"budget": {
						"type": "string",
						"description": "預算範圍（例如：'5萬-10萬'）"
					},
					"project_details": {
						"type": "string",
						"description": "專案詳情摘要"
					}
				}
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
			Msg("Failed to extract info")
		return nil, fmt.Errorf("failed to extract info: %w", err)
	}

	// 解析結果
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := resp.Choices[0]
	if choice.Message.FunctionCall == nil {
		s.logger.Warn().
			Str("response", choice.Message.Content).
			Msg("No function call in response, returning empty result")
		return &ExtractedInfo{}, nil
	}

	// 解析 function call arguments
	args := choice.Message.FunctionCall.Arguments
	var result ExtractedInfo

	if err := json.Unmarshal([]byte(args), &result); err != nil {
		s.logger.Error().
			Err(err).
			Str("arguments", args).
			Msg("Failed to parse function call arguments")
		return nil, fmt.Errorf("failed to parse extraction result: %w", err)
	}

	// 解析日期
	if result.DueDate != nil && (*result.DueDate == time.Time{}) {
		// 嘗試解析字串日期
		// 注意：JSON unmarshal 可能會失敗，我們需要在 unmarshal 之後手動處理
		result.DueDate = nil
	}

	amountStr := "N/A"
	if result.Amount != nil {
		amountStr = fmt.Sprintf("%.2f", *result.Amount)
	}
	s.logger.Info().
		Str("brand", result.BrandName).
		Str("amount", amountStr).
		Int("tokens", resp.Usage.TotalTokens).
		Dur("duration", time.Since(startTime)).
		Msg("Info extraction completed")

	return &result, nil
}

// UnmarshalJSON 自訂 Unmarshal 以正確解析 ExtractedInfo
func (e *ExtractedInfo) UnmarshalJSON(data []byte) error {
	type Alias ExtractedInfo
	aux := &struct {
		DueDate string `json:"due_date"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 解析日期字串
	if aux.DueDate != "" {
		parsedDate, err := time.Parse("2006-01-02", aux.DueDate)
		if err != nil {
			// 如果解析失敗，嘗試其他格式
			parsedDate, err = time.Parse("2006/01/02", aux.DueDate)
			if err != nil {
				parsedDate, err = time.Parse("2006-01-02 15:04:05", aux.DueDate)
				if err != nil {
					// 解析失敗，設為 nil
					e.DueDate = nil
					return nil
				}
			}
		}
		e.DueDate = &parsedDate
	}

	return nil
}

// ExtractBrandName 使用簡單的正則表達式提取品牌名稱
func ExtractBrandName(text string) string {
	// 這個方法可以作為輔助，但主要還是依靠 AI 抽取
	// 預設從 "from" 欄位提取
	parts := strings.Split(text, "<")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[0])
	}
	return text
}

// ExtractAmount 從文字中抽取金額（輔助函數）
func ExtractAmount(text string) (*float64, error) {
	// 移除常見的貨幣符號和單位
	cleaned := strings.ReplaceAll(text, "NT$", "")
	cleaned = strings.ReplaceAll(cleaned, "TWD", "")
	cleaned = strings.ReplaceAll(cleaned, "USD", "")
	cleaned = strings.ReplaceAll(cleaned, "$", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "萬", "0000")
	cleaned = strings.TrimSpace(cleaned)

	// 嘗試解析數字
	amount, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return nil, err
	}

	return &amount, nil
}

// ExtractPhoneNumber 從文字中抽取電話號碼
func ExtractPhoneNumber(text string) string {
	// 簡單的正則匹配（可以改進）
	// 尋找連續的數字
	var result strings.Builder
	inNumber := false

	for _, char := range text {
		if char >= '0' && char <= '9' {
			inNumber = true
			result.WriteRune(char)
		} else if inNumber {
			// 遇到非數字字元，但如果已經有一定的長度就返回
			if result.Len() >= 8 && result.Len() <= 15 {
				return result.String()
			}
			result.Reset()
			inNumber = false
		}
	}

	phone := result.String()
	if len(phone) >= 8 && len(phone) <= 15 {
		return phone
	}

	return ""
}

// ValidateExtractedInfo 驗證抽取的資訊
func ValidateExtractedInfo(info *ExtractedInfo) []string {
	var errors []string

	if info.BrandName == "" {
		errors = append(errors, "品牌名稱為空")
	}

	if info.Amount != nil && *info.Amount < 0 {
		errors = append(errors, "金額不能為負數")
	}

	if info.DueDate != nil && info.DueDate.Before(time.Now()) {
		errors = append(errors, "截止日期不能是過去")
	}

	return errors
}

// SummarizeExtractedInfo 摘要抽取的資訊（用於顯示）
func SummarizeExtractedInfo(info *ExtractedInfo) string {
	var parts []string

	if info.BrandName != "" {
		parts = append(parts, fmt.Sprintf("品牌：%s", info.BrandName))
	}

	if info.Amount != nil {
		parts = append(parts, fmt.Sprintf("金額：%.0f %s", *info.Amount, info.Currency))
	}

	if info.DueDate != nil {
		parts = append(parts, fmt.Sprintf("截止：%s", info.DueDate.Format("2006-01-02")))
	}

	if info.ContentType != "" {
		parts = append(parts, fmt.Sprintf("類型：%s", info.ContentType))
	}

	if len(parts) == 0 {
		return "無明顯資訊"
	}

	return strings.Join(parts, " / ")
}
