package openai

import (
	"context"
	"fmt"
)

// DraftReply 根據案件與對方郵件產生回信草稿（純文字）
func (s *Service) DraftReply(ctx context.Context, req DraftReplyRequest) (*DraftReplyResult, error) {
	s.logger.Info().
		Str("case_title", req.CaseTitle).
		Str("email_from", req.EmailFrom).
		Msg("Starting draft reply")

	systemPrompt := `你是一位協助影響者（influencer）回覆合作邀約的專業助手。
你的任務是根據案件資訊與對方來信，撰寫一封禮貌、專業的回信草稿。
請直接產出回信「內文」純文字，不要包含主旨或稱謂以外的多餘說明。
語氣要專業且友善，適合商業合作往來。`

	if req.UserAIInstructions != "" {
		systemPrompt += fmt.Sprintf("\n\n## 使用者常規注意事項（請務必遵守）\n%s", req.UserAIInstructions)
	}

	userPrompt := fmt.Sprintf(`## 案件摘要
- 標題：%s
- 品牌：%s
- 聯絡人：%s
- 聯絡信箱：%s

## 要回覆的來信
- 寄件者：%s
- 主旨：%s

內文：
%s
`,
		req.CaseTitle,
		req.BrandName,
		req.ContactName,
		req.ContactEmail,
		req.EmailFrom,
		req.EmailSubject,
		s.TruncateContent(req.EmailBody, 3000),
	)

	if req.Instruction != "" {
		userPrompt += fmt.Sprintf("\n## 使用者補充說明\n%s\n\n請在草稿中適當反映以上說明。", req.Instruction)
	}

	messages := s.buildPrompt(systemPrompt, userPrompt)
	// 不使用 function calling，直接取得 assistant 回覆內容
	resp, err := s.callAPI(ctx, messages, nil)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return nil, fmt.Errorf("empty draft from OpenAI")
	}

	return &DraftReplyResult{Draft: content}, nil
}
