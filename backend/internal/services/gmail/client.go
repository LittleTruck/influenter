package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/utils"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Service Gmail 服務
type Service struct {
	client       *gmail.Service
	oauthAccount *models.OAuthAccount
	userEmail    string
}

// NewService 建立新的 Gmail Service
// 從 OAuthAccount 創建 Gmail API client
func NewService(oauthAccount *models.OAuthAccount) (*Service, error) {
	if !oauthAccount.IsGmail() {
		return nil, fmt.Errorf("oauth account is not a Gmail account")
	}

	// 解密 tokens
	accessToken, refreshToken, err := utils.DecryptTokens(
		oauthAccount.AccessToken,
		oauthAccount.RefreshToken,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tokens: %w", err)
	}

	// 創建 OAuth2 token
	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       oauthAccount.TokenExpiry,
		TokenType:    "Bearer",
	}

	// 創建 HTTP client
	// 注意：如果沒有 client config，token 過期時需要手動處理
	ctx := context.Background()
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	// 創建 Gmail service
	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service: %w", err)
	}

	return &Service{
		client:       gmailService,
		oauthAccount: oauthAccount,
		userEmail:    oauthAccount.Email,
	}, nil
}

// ListMessages 列出郵件
func (s *Service) ListMessages(opts *ListMessagesOptions) (*MessageListResult, error) {
	if opts.MaxResults == 0 {
		opts.MaxResults = 100
	}

	call := s.client.Users.Messages.List("me").
		MaxResults(opts.MaxResults).
		Q(opts.Query)

	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}

	if len(opts.Labels) > 0 {
		call = call.LabelIds(opts.Labels...)
	}

	if !opts.IncludeSpam {
		call = call.IncludeSpamTrash(false)
	}

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	// 轉換為簡化格式
	messages := make([]*SimplifiedMessage, 0, len(response.Messages))
	for _, msg := range response.Messages {
		messages = append(messages, &SimplifiedMessage{
			ID:       msg.Id,
			ThreadID: msg.ThreadId,
		})
	}

	return &MessageListResult{
		Messages:      messages,
		NextPageToken: response.NextPageToken,
		TotalEstimate: uint32(response.ResultSizeEstimate),
	}, nil
}

// GetMessage 取得單封郵件的完整內容
func (s *Service) GetMessage(messageID string) (*gmail.Message, error) {
	message, err := s.client.Users.Messages.Get("me", messageID).
		Format("full").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return message, nil
}

// SendMessage 寄送郵件
func (s *Service) SendMessage(req *SendMessageRequest) (string, error) {
	// 建構 RFC 2822 格式的郵件
	message := s.buildRFC2822Message(req)

	// Base64 URL encode
	raw := base64.URLEncoding.EncodeToString([]byte(message))

	gmailMessage := &gmail.Message{
		Raw:      raw,
		ThreadId: req.ThreadID, // 如果是回覆，設定 ThreadID
	}

	sent, err := s.client.Users.Messages.Send("me", gmailMessage).Do()
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return sent.Id, nil
}

// ModifyLabels 修改郵件標籤（批次操作）
func (s *Service) ModifyLabels(messageID string, req *ModifyLabelsRequest) error {
	modifyReq := &gmail.ModifyMessageRequest{
		AddLabelIds:    req.AddLabels,
		RemoveLabelIds: req.RemoveLabels,
	}

	_, err := s.client.Users.Messages.Modify("me", messageID, modifyReq).Do()
	if err != nil {
		return fmt.Errorf("failed to modify labels: %w", err)
	}

	return nil
}

// BatchModifyLabels 批次修改多封郵件的標籤
func (s *Service) BatchModifyLabels(op *BatchOperation) error {
	batchReq := &gmail.BatchModifyMessagesRequest{
		Ids:            op.MessageIDs,
		AddLabelIds:    op.AddLabels,
		RemoveLabelIds: op.RemoveLabels,
	}

	err := s.client.Users.Messages.BatchModify("me", batchReq).Do()
	if err != nil {
		return fmt.Errorf("failed to batch modify: %w", err)
	}

	return nil
}

// MarkAsRead 標記郵件為已讀
func (s *Service) MarkAsRead(messageID string) error {
	return s.ModifyLabels(messageID, &ModifyLabelsRequest{
		RemoveLabels: []string{LabelUnread},
	})
}

// MarkAsUnread 標記郵件為未讀
func (s *Service) MarkAsUnread(messageID string) error {
	return s.ModifyLabels(messageID, &ModifyLabelsRequest{
		AddLabels: []string{LabelUnread},
	})
}

// AddStar 加星號
func (s *Service) AddStar(messageID string) error {
	return s.ModifyLabels(messageID, &ModifyLabelsRequest{
		AddLabels: []string{LabelStarred},
	})
}

// RemoveStar 移除星號
func (s *Service) RemoveStar(messageID string) error {
	return s.ModifyLabels(messageID, &ModifyLabelsRequest{
		RemoveLabels: []string{LabelStarred},
	})
}

// Archive 歸檔郵件（移除 INBOX 標籤）
func (s *Service) Archive(messageID string) error {
	return s.ModifyLabels(messageID, &ModifyLabelsRequest{
		RemoveLabels: []string{LabelInbox},
	})
}

// MoveToTrash 移到垃圾桶
func (s *Service) MoveToTrash(messageID string) error {
	_, err := s.client.Users.Messages.Trash("me", messageID).Do()
	if err != nil {
		return fmt.Errorf("failed to trash message: %w", err)
	}
	return nil
}

// Untrash 從垃圾桶還原
func (s *Service) Untrash(messageID string) error {
	_, err := s.client.Users.Messages.Untrash("me", messageID).Do()
	if err != nil {
		return fmt.Errorf("failed to untrash message: %w", err)
	}
	return nil
}

// Delete 永久刪除郵件
func (s *Service) Delete(messageID string) error {
	err := s.client.Users.Messages.Delete("me", messageID).Do()
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// GetLabels 取得所有可用的標籤
func (s *Service) GetLabels() ([]*gmail.Label, error) {
	response, err := s.client.Users.Labels.List("me").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}
	return response.Labels, nil
}

// GetProfile 取得使用者的 Gmail profile
func (s *Service) GetProfile() (*gmail.Profile, error) {
	profile, err := s.client.Users.GetProfile("me").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return profile, nil
}

// GetHistory 取得歷史記錄（用於增量同步）
func (s *Service) GetHistory(startHistoryID uint64) ([]*gmail.History, error) {
	response, err := s.client.Users.History.List("me").
		StartHistoryId(startHistoryID).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	return response.History, nil
}

// buildRFC2822Message 建構 RFC 2822 格式的郵件
func (s *Service) buildRFC2822Message(req *SendMessageRequest) string {
	message := ""

	// To
	if len(req.To) > 0 {
		message += "To: "
		for i, to := range req.To {
			if i > 0 {
				message += ", "
			}
			message += to
		}
		message += "\r\n"
	}

	// Cc
	if len(req.Cc) > 0 {
		message += "Cc: "
		for i, cc := range req.Cc {
			if i > 0 {
				message += ", "
			}
			message += cc
		}
		message += "\r\n"
	}

	// Bcc
	if len(req.Bcc) > 0 {
		message += "Bcc: "
		for i, bcc := range req.Bcc {
			if i > 0 {
				message += ", "
			}
			message += bcc
		}
		message += "\r\n"
	}

	// Subject
	message += "Subject: " + req.Subject + "\r\n"

	// In-Reply-To (用於回覆)
	if req.InReplyTo != "" {
		message += "In-Reply-To: " + req.InReplyTo + "\r\n"
	}

	// References (用於郵件串)
	if req.References != "" {
		message += "References: " + req.References + "\r\n"
	}

	// MIME version
	message += "MIME-Version: 1.0\r\n"

	// Content-Type
	if req.HTMLBody != "" {
		// 同時包含 HTML 和 plain text
		boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
		message += "Content-Type: multipart/alternative; boundary=" + boundary + "\r\n"
		message += "\r\n"

		// Plain text part
		message += "--" + boundary + "\r\n"
		message += "Content-Type: text/plain; charset=UTF-8\r\n"
		message += "\r\n"
		message += req.TextBody + "\r\n"
		message += "\r\n"

		// HTML part
		message += "--" + boundary + "\r\n"
		message += "Content-Type: text/html; charset=UTF-8\r\n"
		message += "\r\n"
		message += req.HTMLBody + "\r\n"
		message += "\r\n"
		message += "--" + boundary + "--"
	} else {
		// 只有 plain text
		message += "Content-Type: text/plain; charset=UTF-8\r\n"
		message += "\r\n"
		message += req.TextBody
	}

	return message
}
