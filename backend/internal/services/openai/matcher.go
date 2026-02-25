package openai

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// MatchCollaborationItems 使用 AI 匹配合作項目
func (s *Service) MatchCollaborationItems(ctx context.Context, req MatchCollaborationItemsRequest) (*MatchCollaborationItemsResult, error) {
	if len(req.Items) == 0 {
		return &MatchCollaborationItemsResult{
			MatchedItemIDs: []string{},
			Confidence:     0,
			Reason:         "No collaboration items to match",
		}, nil
	}

	// 建立項目清單描述
	itemsJSON, _ := json.Marshal(req.Items)

	systemPrompt := `你是一位協助創作者管理合作案件的 AI 助手。
你的任務是分析郵件內容，並從使用者的合作項目清單中找出最可能相關的項目。

規則：
- 根據郵件中提到的合作類型、內容形式（影片、圖文、限時動態等）來比對
- 如果郵件明確提到某種合作形式，選擇最匹配的項目
- 如果不確定，可以回傳空的匹配結果
- confidence 範圍 0-1，0.5 以上才算有效匹配
- 可以匹配多個項目（例如郵件提到影片+圖文）`

	userPrompt := fmt.Sprintf(`## 郵件資訊
- 寄件者：%s
- 主旨：%s
- 內容：%s

## 使用者合作項目清單
%s

請分析郵件內容，找出匹配的合作項目。`, req.EmailFrom, req.EmailSubject, s.TruncateContent(req.EmailBody, 3000), string(itemsJSON))

	functions := []openai.FunctionDefinition{
		{
			Name:        "match_collaboration_items",
			Description: "從使用者的合作項目清單中匹配郵件相關的項目",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"matched_item_ids": {
						"type": "array",
						"items": {"type": "string"},
						"description": "匹配到的合作項目 ID 列表"
					},
					"confidence": {
						"type": "number",
						"minimum": 0,
						"maximum": 1,
						"description": "匹配信心度 (0-1)"
					},
					"reason": {
						"type": "string",
						"description": "匹配的理由說明"
					}
				},
				"required": ["matched_item_ids", "confidence", "reason"]
			}`),
		},
	}

	messages := s.buildPrompt(systemPrompt, userPrompt)
	resp, err := s.callAPI(ctx, messages, functions)
	if err != nil {
		return nil, fmt.Errorf("match collaboration items failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return &MatchCollaborationItemsResult{
			MatchedItemIDs: []string{},
			Confidence:     0,
			Reason:         "No response from AI",
		}, nil
	}

	args := safeString(&resp.Choices[0])
	var result MatchCollaborationItemsResult
	if err := json.Unmarshal([]byte(args), &result); err != nil {
		s.logger.Error().Err(err).Str("raw", args).Msg("Failed to parse match result")
		return &MatchCollaborationItemsResult{
			MatchedItemIDs: []string{},
			Confidence:     0,
			Reason:         "Failed to parse AI response",
		}, nil
	}

	return &result, nil
}
