package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// AnalyzeReplyForCaseUpdate 根據寄出的回信內容，分析並建議案件狀態與進度更新
func (s *Service) AnalyzeReplyForCaseUpdate(ctx context.Context, req ReplyCaseUpdateRequest) (*ReplyCaseUpdateResult, error) {
	s.logger.Info().
		Str("case_title", req.CaseTitle).
		Str("case_status", req.CaseStatus).
		Msg("Starting reply analysis for case update")

	systemPrompt := `你是一個專業的合作案件管理助手，協助影響者（influencer）管理與品牌的合作案件。

任務：根據使用者剛剛寄出的回信內容，判斷是否需要更新關聯的案件狀態與進度。

案件狀態說明：
1. **to_confirm** - 待確認：剛收到邀約或尚未確認合作意向
2. **in_progress** - 進行中：已確認合作、正在溝通細節或執行中
3. **completed** - 已完成：合作結案
4. **cancelled** - 已取消：婉拒或取消合作
5. **other** - 其他：非合作相關

請分析回信內容，判斷：
- 是否應更新案件狀態（例如：回信確認合作 → in_progress；婉拒 → cancelled；結案 → completed）
- 是否需要新增進度說明（notes_progress）：簡短描述此次回信的重點或後續
- 是否可從回信抽取出新的報價、截止日等資訊

若回信內容與案件進度無關（如純禮貌性回覆），請設 should_update 為 false。`

	userPrompt := fmt.Sprintf(`## 原始來信
**寄件者**: %s
**主旨**: %s
**內文**:
%s

## 目前案件
- 標題: %s
- 狀態: %s
- 描述: %s
- 備註: %s
- 預估報價: %s
- 截止日: %s

## 剛寄出的回信
%s

請根據回信內容，判斷是否應更新案件，並填寫建議的更新項目。`,
		req.EmailFrom,
		req.EmailSubject,
		s.TruncateContent(req.EmailBody, 1500),
		req.CaseTitle,
		req.CaseStatus,
		req.CaseDescription,
		req.CaseNotes,
		req.CaseQuotedAmount,
		req.CaseDeadline,
		s.TruncateContent(req.ReplyBody, 1500),
	)

	functions := []openai.FunctionDefinition{
		{
			Name:        "suggest_case_update",
			Description: "根據回信內容建議案件狀態與進度更新",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"should_update": {
						"type": "boolean",
						"description": "是否有建議的更新項目（若回信與案件進度無關則填 false）"
					},
					"status": {
						"type": "string",
						"enum": ["to_confirm", "in_progress", "completed", "cancelled", "other"],
						"description": "建議的新案件狀態"
					},
					"notes_progress": {
						"type": "string",
						"description": "進度說明，簡短描述此次回信重點或後續（可附加到 notes）"
					},
					"description_update": {
						"type": "string",
						"description": "案件描述的更新或補充"
					},
					"quoted_amount": {
						"type": "number",
						"description": "預估報價（若回信中提及）"
					},
					"final_amount": {
						"type": "number",
						"description": "最終金額（若回信中已確定）"
					},
					"deadline_date": {
						"type": "string",
						"description": "截止日期，ISO 8601 格式 YYYY-MM-DD（若回信中提及）"
					},
					"reason": {
						"type": "string",
						"description": "更新建議的理由"
					}
				},
				"required": ["should_update", "reason"]
			}`),
		},
	}

	messages := s.buildPrompt(systemPrompt, userPrompt)

	resp, err := s.callAPI(ctx, messages, functions)
	if err != nil {
		s.logger.Error().Err(err).Msg("AnalyzeReplyForCaseUpdate API failed")
		return nil, fmt.Errorf("analyze reply failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := resp.Choices[0]
	args := ""
	if choice.Message.FunctionCall != nil {
		args = choice.Message.FunctionCall.Arguments
	} else if choice.Message.Content != "" {
		args = choice.Message.Content
	}

	var result ReplyCaseUpdateResult
	if err := json.Unmarshal([]byte(args), &result); err != nil {
		s.logger.Error().Err(err).Str("args", args).Msg("Failed to parse reply update result")
		return nil, fmt.Errorf("parse reply update result: %w", err)
	}

	// 若 should_update 為 false，仍回傳結果但不做欄位更新
	if !result.ShouldUpdate {
		s.logger.Info().
			Str("reason", result.Reason).
			Msg("Reply analysis: no update suggested")
		return &result, nil
	}

	// 驗證 status 為合法值
	validStatus := map[string]bool{
		"to_confirm": true, "in_progress": true, "completed": true,
		"cancelled": true, "other": true,
	}
	if result.Status != "" && !validStatus[result.Status] {
		result.Status = ""
	}

	// 驗證 deadline_date 格式
	if result.DeadlineDate != "" {
		if _, err := time.Parse("2006-01-02", result.DeadlineDate); err != nil {
			result.DeadlineDate = ""
		}
	}

	s.logger.Info().
		Bool("should_update", result.ShouldUpdate).
		Str("status", result.Status).
		Str("notes_progress", truncateStr(result.NotesProgress, 50)).
		Msg("Reply analysis completed")

	return &result, nil
}

func truncateStr(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
