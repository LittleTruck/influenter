package gmail

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/gmail/v1"
)

// ParseMessage 解析 Gmail API 回傳的郵件為我們的格式
func ParseMessage(gmailMsg *gmail.Message, oauthAccountID uuid.UUID) (*models.Email, error) {
	parsed := &ParsedMessage{
		ID:           gmailMsg.Id,
		ThreadID:     gmailMsg.ThreadId,
		LabelIDs:     gmailMsg.LabelIds,
		Snippet:      gmailMsg.Snippet,
		HistoryID:    fmt.Sprintf("%d", gmailMsg.HistoryId),
		SizeEstimate: int32(gmailMsg.SizeEstimate),
	}

	// 解析 InternalDate (milliseconds since epoch)
	if gmailMsg.InternalDate > 0 {
		parsed.InternalDate = time.Unix(0, gmailMsg.InternalDate*int64(time.Millisecond))
	}

	// 解析 Headers
	if gmailMsg.Payload != nil {
		parseHeaders(gmailMsg.Payload.Headers, parsed)

		// 解析 Body
		parseBody(gmailMsg.Payload, parsed)
	}

	// 轉換為我們的 Email model
	email := &models.Email{
		ID:                uuid.New(),
		OAuthAccountID:    oauthAccountID,
		ProviderMessageID: parsed.ID,
		ThreadID:          &parsed.ThreadID,
		FromEmail:         parsed.From.Address,
		FromName:          stringPtr(parsed.From.Name),
		Subject:           stringPtr(parsed.Subject),
		BodyText:          stringPtr(parsed.TextBody),
		BodyHTML:          stringPtr(parsed.HTMLBody),
		Snippet:           stringPtr(parsed.Snippet),
		ReceivedAt:        parsed.InternalDate,
		Labels:            parsed.LabelIDs,
		HasAttachments:    parsed.HasAttachments,
		IsRead:            !contains(parsed.LabelIDs, LabelUnread),
	}

	// 設定 To email（取第一個）
	if len(parsed.To) > 0 {
		email.ToEmail = stringPtr(parsed.To[0].Address)
	}

	return email, nil
}

// parseHeaders 解析郵件 headers
func parseHeaders(headers []*gmail.MessagePartHeader, parsed *ParsedMessage) {
	for _, header := range headers {
		switch header.Name {
		case "From":
			parsed.From = parseEmailAddress(header.Value)
		case "To":
			parsed.To = parseEmailAddresses(header.Value)
		case "Cc":
			parsed.Cc = parseEmailAddresses(header.Value)
		case "Bcc":
			parsed.Bcc = parseEmailAddresses(header.Value)
		case "Subject":
			// 使用 mail.ParseAddress 的內部解碼機制
			// MIME encoded words 會在 parseEmailAddress 中處理
			parsed.Subject = header.Value
		case "Date":
			if t, err := mail.ParseDate(header.Value); err == nil {
				parsed.Date = t
			}
		case "Message-ID":
			parsed.MessageID = header.Value
		}
	}
}

// parseBody 解析郵件內容
func parseBody(payload *gmail.MessagePart, parsed *ParsedMessage) {
	// 檢查是否有附件
	if payload.Filename != "" && payload.Body != nil && payload.Body.AttachmentId != "" {
		parsed.HasAttachments = true
		attachmentSize := int32(payload.Body.Size)
		if payload.Body.Size > 2147483647 { // int32 max
			attachmentSize = 2147483647
		}
		parsed.Attachments = append(parsed.Attachments, Attachment{
			PartID:   payload.PartId,
			Filename: payload.Filename,
			MimeType: payload.MimeType,
			Size:     attachmentSize,
		})
	}

	// 處理 multipart
	if len(payload.Parts) > 0 {
		for _, part := range payload.Parts {
			parseBody(part, parsed)
		}
		return
	}

	// 處理單一 part
	if payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			return
		}

		body := string(decoded)

		switch payload.MimeType {
		case "text/plain":
			if parsed.TextBody == "" {
				parsed.TextBody = body
			}
		case "text/html":
			if parsed.HTMLBody == "" {
				parsed.HTMLBody = body
			}
		}
	}
}

// parseEmailAddress 解析單個郵件地址
func parseEmailAddress(str string) EmailAddress {
	addr, err := mail.ParseAddress(str)
	if err != nil {
		// 如果解析失敗，嘗試簡單分割
		parts := strings.Split(str, "<")
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			email := strings.Trim(parts[1], ">")
			return EmailAddress{
				Name:    name,
				Address: email,
			}
		}
		return EmailAddress{
			Address: str,
		}
	}

	return EmailAddress{
		Name:    addr.Name,
		Address: addr.Address,
	}
}

// parseEmailAddresses 解析多個郵件地址
func parseEmailAddresses(str string) []EmailAddress {
	addresses, err := mail.ParseAddressList(str)
	if err != nil {
		// 簡單分割
		parts := strings.Split(str, ",")
		result := make([]EmailAddress, 0, len(parts))
		for _, part := range parts {
			result = append(result, parseEmailAddress(strings.TrimSpace(part)))
		}
		return result
	}

	result := make([]EmailAddress, 0, len(addresses))
	for _, addr := range addresses {
		result = append(result, EmailAddress{
			Name:    addr.Name,
			Address: addr.Address,
		})
	}
	return result
}

// contains 檢查 slice 是否包含某個字串
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// stringPtr 返回字串指標（如果不為空）
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ParseMessageToModel 直接將 Gmail Message 轉換為 Email Model（便利函數）
func ParseMessageToModel(gmailMsg *gmail.Message, oauthAccountID uuid.UUID) (*models.Email, error) {
	return ParseMessage(gmailMsg, oauthAccountID)
}

// GetCategory 取得郵件的 Gmail 分類
func GetCategory(labels []string) string {
	for _, label := range labels {
		if strings.HasPrefix(label, "CATEGORY_") {
			return label
		}
	}
	return LabelCategoryPersonal // 預設為 Personal
}

// IsInCategory 檢查郵件是否在特定分類中
func IsInCategory(labels []string, category string) bool {
	return contains(labels, category)
}

// ExtractPlainText 從 HTML 中提取純文字（簡單版本）
func ExtractPlainText(html string) string {
	// 移除 HTML 標籤（簡單實作）
	// 未來可以使用更強大的 HTML parser
	text := html

	// 移除 <script> 和 <style> 標籤及其內容
	text = removeTagAndContent(text, "script")
	text = removeTagAndContent(text, "style")

	// 移除所有 HTML 標籤
	for {
		start := strings.Index(text, "<")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + " " + text[start+end+1:]
	}

	// 清理多餘空白
	text = strings.TrimSpace(text)

	return text
}

// removeTagAndContent 移除特定 HTML 標籤及其內容
func removeTagAndContent(html, tag string) string {
	result := html
	openTag := "<" + tag
	closeTag := "</" + tag + ">"

	for {
		start := strings.Index(strings.ToLower(result), strings.ToLower(openTag))
		if start == -1 {
			break
		}

		// 找到 > 的位置
		tagEnd := strings.Index(result[start:], ">")
		if tagEnd == -1 {
			break
		}

		// 找到結束標籤
		end := strings.Index(strings.ToLower(result[start:]), strings.ToLower(closeTag))
		if end == -1 {
			break
		}

		result = result[:start] + result[start+end+len(closeTag):]
	}

	return result
}
