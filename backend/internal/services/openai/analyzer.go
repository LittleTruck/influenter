package openai

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AnalyzeEmail 分析郵件（分類 + 資訊抽取）
func (s *Service) AnalyzeEmail(ctx context.Context, req AnalyzeEmailRequest) (*EmailAnalysisResult, error) {
	s.logger.Info().
		Str("from", req.From).
		Str("subject", req.Subject).
		Msg("Starting full email analysis")

	startTime := time.Now()

	// 同時執行分類和資訊抽取
	classificationReq := ClassifyEmailRequest{
		Subject: req.Subject,
		Body:    req.Body,
		From:    req.From,
		Options: req.Options,
	}

	classificationChan := make(chan *ClassificationResult, 1)
	extractionChan := make(chan *ExtractionResult, 1)

	// 並行執行分類
	go func() {
		classification, err := s.ClassifyEmail(ctx, classificationReq)
		classificationChan <- &ClassificationResult{
			Classification: classification,
			Err:            err,
		}
	}()

	// 並行執行資訊抽取
	go func() {
		info, err := s.ExtractInfo(ctx, req)
		extractionChan <- &ExtractionResult{
			Info: info,
			Err:  err,
		}
	}()

	// 收集結果
	var classification *EmailClassification
	var extractedInfo *ExtractedInfo

	// 等待分類結果
	select {
	case result := <-classificationChan:
		if result.Err != nil {
			s.logger.Warn().
				Err(result.Err).
				Msg("Classification failed")
		} else {
			classification = result.Classification
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled during classification")
	}

	// 等待抽取結果
	select {
	case result := <-extractionChan:
		if result.Err != nil {
			s.logger.Warn().
				Err(result.Err).
				Msg("Info extraction failed")
		} else {
			extractedInfo = result.Info
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled during extraction")
	}

	// 如果兩個都失敗，返回錯誤
	if classification == nil && extractedInfo == nil {
		return nil, fmt.Errorf("both classification and extraction failed")
	}

	// 建立分析結果
	result := &EmailAnalysisResult{
		Classification: EmailClassification{},
		ExtractedInfo:  ExtractedInfo{},
		Summary:        "",
		KeyPoints:      []string{},
		ActionRequired: false,
		Priority:       "medium",
		TokensUsed:     0,
		Model:          s.config.Model,
		AnalyzedAt:     time.Now(),
	}

	// 填入分類結果
	if classification != nil {
		result.Classification = *classification
		// 根據分類判斷是否需要行動
		result.ActionRequired = IsHighPriorityCategory(classification.Category) || classification.Confidence > 0.8
		// 根據分類設定優先級
		if classification.Category == CategoryCollaboration && classification.Confidence > 0.8 {
			result.Priority = "high"
		} else if classification.Category == CategorySpam || classification.Confidence < 0.5 {
			result.Priority = "low"
		}
	}

	// 填入抽取資訊
	if extractedInfo != nil {
		result.ExtractedInfo = *extractedInfo
	}

	// 生成摘要
	result.Summary = s.generateSummary(classification, extractedInfo)

	// 提取關鍵要點
	result.KeyPoints = s.extractKeyPoints(classification, extractedInfo)

	s.logger.Info().
		Str("category", string(result.Classification.Category)).
		Float64("confidence", result.Classification.Confidence).
		Str("priority", result.Priority).
		Dur("duration", time.Since(startTime)).
		Msg("Email analysis completed")

	return result, nil
}

// ClassificationResult 分類結果包裝
type ClassificationResult struct {
	Classification *EmailClassification
	Err            error
}

// ExtractionResult 抽取結果包裝
type ExtractionResult struct {
	Info *ExtractedInfo
	Err  error
}

// generateSummary 生成郵件摘要
func (s *Service) generateSummary(classification *EmailClassification, info *ExtractedInfo) string {
	if classification == nil {
		return "無法生成摘要"
	}

	var summary string
	switch classification.Category {
	case CategoryCollaboration:
		if info != nil && info.BrandName != "" {
			summary = fmt.Sprintf("收到來自 %s 的合作邀約", info.BrandName)
			if info.Amount != nil {
				summary += fmt.Sprintf("，預算約 %.0f %s", *info.Amount, info.Currency)
			}
		} else {
			summary = "收到合作邀約郵件"
		}
	case CategoryPayment:
		summary = "付款相關郵件"
		if info != nil && info.Amount != nil {
			summary += fmt.Sprintf("，金額：%.0f %s", *info.Amount, info.Currency)
		}
	case CategoryConfirmation:
		summary = "確認郵件"
	case CategoryInquiry:
		summary = "詢問郵件"
		if info != nil && info.BrandName != "" {
			summary += fmt.Sprintf("，來自 %s", info.BrandName)
		}
	default:
		summary = fmt.Sprintf("分類為：%s", GetCategoryDisplayName(classification.Category))
	}

	return summary
}

// extractKeyPoints 提取關鍵要點
func (s *Service) extractKeyPoints(classification *EmailClassification, info *ExtractedInfo) []string {
	var points []string

	if classification == nil {
		return points
	}

	// 根據分類添加關鍵要點
	switch classification.Category {
	case CategoryCollaboration:
		if info != nil {
			if info.BrandName != "" {
				points = append(points, fmt.Sprintf("品牌：%s", info.BrandName))
			}
			if info.Amount != nil {
				points = append(points, fmt.Sprintf("預算：%.0f %s", *info.Amount, info.Currency))
			}
			if info.DueDate != nil {
				points = append(points, fmt.Sprintf("截止日期：%s", info.DueDate.Format("2006-01-02")))
			}
			if info.ContentType != "" {
				points = append(points, fmt.Sprintf("內容類型：%s", info.ContentType))
			}
			if info.FollowerCount != "" {
				points = append(points, fmt.Sprintf("粉絲要求：%s", info.FollowerCount))
			}
		}
	case CategoryPayment:
		if info != nil && info.Amount != nil {
			points = append(points, fmt.Sprintf("金額：%.0f %s", *info.Amount, info.Currency))
			if info.DueDate != nil {
				points = append(points, fmt.Sprintf("繳費期限：%s", info.DueDate.Format("2006-01-02")))
			}
		}
	case CategoryInquiry:
		if info != nil && info.BrandName != "" {
			points = append(points, fmt.Sprintf("詢問方：%s", info.BrandName))
		}
	}

	return points
}

// AnalyzeEmailSimple 簡化的郵件分析（只做分類）
func (s *Service) AnalyzeEmailSimple(ctx context.Context, subject, body, from string) (*EmailAnalysisResult, error) {
	req := AnalyzeEmailRequest{
		Subject: subject,
		Body:    body,
		From:    from,
		To:      []string{},
		Date:    time.Now(),
		Options: AnalysisOptions{
			DetailLevel: "basic",
		},
	}

	return s.AnalyzeEmail(ctx, req)
}

// GetAnalysisCost 計算分析的預估成本
func (s *Service) GetAnalysisCost(avgTokens int) float64 {
	return s.CalculateCost(avgTokens, s.config.Model)
}

// IsWorthAnalyzing 判斷郵件是否值得分析
func (s *Service) IsWorthAnalyzing(subject, body string) bool {
	// 簡單的啟發式規則
	subjectLower := strings.ToLower(subject)
	bodyLower := strings.ToLower(body)

	// 可能的垃圾郵件指標
	spamIndicators := []string{
		"unsubscribe",
		"click here",
		"limited time",
		"act now",
		"guarantee",
		"no risk",
		"free",
		"winner",
	}

	for _, indicator := range spamIndicators {
		if strings.Contains(subjectLower, indicator) || strings.Contains(bodyLower, indicator) {
			return false
		}
	}

	// 如果主題或內容太短，可能不值得分析
	if len(subject) < 5 || len(body) < 50 {
		return false
	}

	return true
}

// BatchAnalyzeEmails 批次分析郵件
func (s *Service) BatchAnalyzeEmails(ctx context.Context, emails []AnalyzeEmailRequest) ([]*EmailAnalysisResult, []error) {
	results := make([]*EmailAnalysisResult, 0, len(emails))
	errors := make([]error, 0)

	for _, email := range emails {
		if !s.IsWorthAnalyzing(email.Subject, email.Body) {
			s.logger.Debug().
				Str("subject", email.Subject).
				Msg("Skipping email analysis")
			continue
		}

		result, err := s.AnalyzeEmail(ctx, email)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to analyze email %s: %w", email.Subject, err))
			continue
		}

		results = append(results, result)
	}

	return results, errors
}
