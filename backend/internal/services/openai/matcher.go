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

// MatchWorkflowTemplate 使用 AI 選擇最適合的流程範本
func (s *Service) MatchWorkflowTemplate(ctx context.Context, req MatchWorkflowTemplateRequest) (*MatchWorkflowTemplateResult, error) {
	if len(req.Templates) == 0 {
		return &MatchWorkflowTemplateResult{
			Confidence: 0,
			Reason:     "No workflow templates available",
		}, nil
	}

	templatesJSON, _ := json.Marshal(req.Templates)

	systemPrompt := `你是一位協助創作者管理合作案件的 AI 助手。
你的任務是根據案件資訊和郵件內容，從使用者的流程範本清單中選出最適合套用的範本。

規則：
- 根據案件的合作類型（影片、圖文、限時動態等）來比對範本
- 考慮範本的名稱、描述和階段內容是否與案件匹配
- 如果找不到適合的範本，回傳空的 template_id 和低信心度
- confidence 範圍 0-1，0.5 以上才算有效匹配
- 只能選擇一個最適合的範本`

	emailInfo := ""
	if req.EmailSubject != "" || req.EmailBody != "" {
		emailInfo = fmt.Sprintf("\n\n## 相關郵件\n- 主旨：%s\n- 內容：%s",
			req.EmailSubject, s.TruncateContent(req.EmailBody, 2000))
	}

	caseDesc := ""
	if req.CaseDescription != "" {
		caseDesc = fmt.Sprintf("\n- 描述：%s", req.CaseDescription)
	}

	userPrompt := fmt.Sprintf(`## 案件資訊
- 標題：%s
- 品牌：%s%s%s

## 可用的流程範本
%s

請選出最適合此案件的流程範本。`, req.CaseTitle, req.CaseBrandName, caseDesc, emailInfo, string(templatesJSON))

	functions := []openai.FunctionDefinition{
		{
			Name:        "match_workflow_template",
			Description: "從流程範本清單中選出最適合案件的範本",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"template_id": {
						"type": "string",
						"description": "選中的流程範本 ID，如果沒有適合的則為空字串"
					},
					"confidence": {
						"type": "number",
						"minimum": 0,
						"maximum": 1,
						"description": "匹配信心度 (0-1)"
					},
					"reason": {
						"type": "string",
						"description": "選擇的理由說明"
					}
				},
				"required": ["template_id", "confidence", "reason"]
			}`),
		},
	}

	messages := s.buildPrompt(systemPrompt, userPrompt)
	resp, err := s.callAPI(ctx, messages, functions)
	if err != nil {
		return nil, fmt.Errorf("match workflow template failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return &MatchWorkflowTemplateResult{
			Confidence: 0,
			Reason:     "No response from AI",
		}, nil
	}

	args := safeString(&resp.Choices[0])
	var result MatchWorkflowTemplateResult
	if err := json.Unmarshal([]byte(args), &result); err != nil {
		s.logger.Error().Err(err).Str("raw", args).Msg("Failed to parse workflow match result")
		return &MatchWorkflowTemplateResult{
			Confidence: 0,
			Reason:     "Failed to parse AI response",
		}, nil
	}

	return &result, nil
}
