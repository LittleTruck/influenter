package openai

import (
	"context"
	"fmt"
)

// DraftReply 根據案件與對方郵件產生回信草稿（HTML 格式）
func (s *Service) DraftReply(ctx context.Context, req DraftReplyRequest) (*DraftReplyResult, error) {
	s.logger.Info().
		Str("case_title", req.CaseTitle).
		Str("email_from", req.EmailFrom).
		Msg("Starting draft reply")

	systemPrompt := `你是一位協助影響者（influencer）回覆合作邀約的專業助手。
你的任務是根據案件資訊與對方來信，撰寫一封禮貌、專業的回信草稿。
請直接產出回信「內文」的 HTML 格式，不要包含主旨或稱謂以外的多餘說明。
語氣要專業且友善，適合商業合作往來。

## 格式要求
- 請使用 HTML 標籤來格式化回信內容（如 <p>、<strong>、<em>、<ul>、<ol>、<li>、<br> 等）
- 每個段落請用 <p> 標籤包裹
- 不要包含 <html>、<head>、<body> 等外層標籤，只需要內文的 HTML 片段
- 如果信件標頭或標尾包含 HTML 標籤，請保留其原始格式直接嵌入`

	if req.UserAIInstructions != "" {
		systemPrompt += fmt.Sprintf("\n\n## 使用者常規注意事項（請務必遵守）\n%s", req.UserAIInstructions)
	}

	if req.UserAIReplyHeader != "" {
		systemPrompt += fmt.Sprintf("\n\n## 信件標頭（請在回信開頭加上以下內容）\n%s", req.UserAIReplyHeader)
	}

	if req.UserAIReplyFooter != "" {
		systemPrompt += fmt.Sprintf("\n\n## 信件標尾（請在回信結尾加上以下內容）\n%s", req.UserAIReplyFooter)
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

	if req.TemplatePrompt != "" {
		userPrompt += fmt.Sprintf("\n## 回覆範本（請參考以下範本的風格與格式撰寫回信）\n%s", req.TemplatePrompt)
	}

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
