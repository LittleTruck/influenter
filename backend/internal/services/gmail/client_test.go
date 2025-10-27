package gmail

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/google/uuid"
)

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		name string
		opts *SearchOptions
		want string
	}{
		{
			name: "from only",
			opts: &SearchOptions{
				From: "sender@example.com",
			},
			want: "from:sender@example.com ",
		},
		{
			name: "subject only",
			opts: &SearchOptions{
				Subject: "test",
			},
			want: "subject:test ",
		},
		{
			name: "has attachment",
			opts: &SearchOptions{
				HasAttachment: true,
			},
			want: "has:attachment ",
		},
		{
			name: "is unread",
			opts: &SearchOptions{
				IsUnread: true,
			},
			want: "is:unread ",
		},
		{
			name: "complex query",
			opts: &SearchOptions{
				From:          "sender@example.com",
				Subject:       "test",
				HasWords:      "important",
				HasAttachment: true,
				IsUnread:      true,
			},
			want: "from:sender@example.com subject:test important has:attachment is:unread ",
		},
		{
			name: "with after date",
			opts: &SearchOptions{
				After: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: "after:2024/01/01 ",
		},
		{
			name: "with labels",
			opts: &SearchOptions{
				Labels: []string{"INBOX", "IMPORTANT"},
			},
			want: "label:INBOX label:IMPORTANT ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.BuildQuery()
			if got != tt.want {
				t.Errorf("BuildQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildRFC2822Message_PlainText(t *testing.T) {
	// 創建一個 mock service（不需要真正的 Gmail API）
	service := &Service{
		userEmail: "test@example.com",
	}

	req := &SendMessageRequest{
		To:       []string{"recipient@example.com"},
		Subject:  "Test Subject",
		TextBody: "Test body text",
	}

	message := service.buildRFC2822Message(req)

	if message == "" {
		t.Fatal("Expected message to be non-empty")
	}

	// 檢查是否包含必要欄位
	if !containsString(message, "To: recipient@example.com") {
		t.Error("Expected message to contain To field")
	}

	if !containsString(message, "Subject: Test Subject") {
		t.Error("Expected message to contain Subject field")
	}

	if !containsString(message, "Test body text") {
		t.Error("Expected message to contain body text")
	}

	if !containsString(message, "MIME-Version: 1.0") {
		t.Error("Expected message to contain MIME-Version")
	}
}

func TestBuildRFC2822Message_HTMLAndText(t *testing.T) {
	service := &Service{
		userEmail: "test@example.com",
	}

	req := &SendMessageRequest{
		To:       []string{"recipient@example.com"},
		Cc:       []string{"cc@example.com"},
		Bcc:      []string{"bcc@example.com"},
		Subject:  "Test Subject",
		TextBody: "Plain text body",
		HTMLBody: "<html><body>HTML body</body></html>",
	}

	message := service.buildRFC2822Message(req)

	if !containsString(message, "Content-Type: multipart/alternative") {
		t.Error("Expected message to be multipart/alternative")
	}

	if !containsString(message, "Plain text body") {
		t.Error("Expected message to contain plain text")
	}

	if !containsString(message, "<html><body>HTML body</body></html>") {
		t.Error("Expected message to contain HTML")
	}

	if !containsString(message, "Cc: cc@example.com") {
		t.Error("Expected message to contain Cc field")
	}

	if !containsString(message, "Bcc: bcc@example.com") {
		t.Error("Expected message to contain Bcc field")
	}
}

func TestBuildRFC2822Message_Reply(t *testing.T) {
	service := &Service{
		userEmail: "test@example.com",
	}

	req := &SendMessageRequest{
		To:         []string{"recipient@example.com"},
		Subject:    "Re: Test Subject",
		TextBody:   "Reply text",
		InReplyTo:  "<original-message-id@example.com>",
		References: "<original-message-id@example.com>",
	}

	message := service.buildRFC2822Message(req)

	if !containsString(message, "In-Reply-To: <original-message-id@example.com>") {
		t.Error("Expected message to contain In-Reply-To field")
	}

	if !containsString(message, "References: <original-message-id@example.com>") {
		t.Error("Expected message to contain References field")
	}
}

func TestBuildRFC2822Message_MultipleRecipients(t *testing.T) {
	service := &Service{
		userEmail: "test@example.com",
	}

	req := &SendMessageRequest{
		To:       []string{"recipient1@example.com", "recipient2@example.com"},
		Subject:  "Test",
		TextBody: "Body",
	}

	message := service.buildRFC2822Message(req)

	if !containsString(message, "recipient1@example.com, recipient2@example.com") {
		t.Error("Expected message to contain multiple recipients")
	}
}

// TestBase64Encoding 測試 base64 encoding 是否正確
func TestBase64Encoding(t *testing.T) {
	testString := "Test email body text"
	encoded := base64.URLEncoding.EncodeToString([]byte(testString))

	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if string(decoded) != testString {
		t.Errorf("Expected decoded string '%s', got '%s'", testString, string(decoded))
	}
}

// containsString 檢查字串是否包含子字串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findInMiddle(s, substr))))
}

func findInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// MockGmailMessage 用於測試的 Gmail Message
func MockGmailMessage(id string) *models.Email {
	subject := "Test Subject"
	body := "Test body"
	fromName := "Test Sender"
	toEmail := "recipient@example.com"
	snippet := "Test snippet"

	return &models.Email{
		ID:                uuid.New(),
		ProviderMessageID: id,
		FromEmail:         "sender@example.com",
		FromName:          &fromName,
		ToEmail:           &toEmail,
		Subject:           &subject,
		BodyText:          &body,
		Snippet:           &snippet,
		ReceivedAt:        time.Now(),
		IsRead:            false,
		HasAttachments:    false,
		Labels:            []string{"INBOX", "UNREAD"},
	}
}

// TestIsTokenExpired 測試 token 過期檢查
func TestIsTokenExpired(t *testing.T) {
	tests := []struct {
		name        string
		expiry      time.Time
		wantExpired bool
	}{
		{
			name:        "expired token",
			expiry:      time.Now().Add(-1 * time.Hour),
			wantExpired: true,
		},
		{
			name:        "valid token",
			expiry:      time.Now().Add(1 * time.Hour),
			wantExpired: false,
		},
		{
			name:        "token expiring soon",
			expiry:      time.Now().Add(5 * time.Minute),
			wantExpired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &models.OAuthAccount{
				TokenExpiry: tt.expiry,
			}

			expired := account.IsTokenExpired()
			if expired != tt.wantExpired {
				t.Errorf("IsTokenExpired() = %v, want %v", expired, tt.wantExpired)
			}
		})
	}
}

// TestCanSync 測試是否可以同步
func TestCanSync(t *testing.T) {
	tests := []struct {
		name    string
		account *models.OAuthAccount
		wantCan bool
	}{
		{
			name: "can sync - active and valid token",
			account: &models.OAuthAccount{
				SyncStatus:  models.SyncStatusActive,
				TokenExpiry: time.Now().Add(1 * time.Hour),
			},
			wantCan: true,
		},
		{
			name: "cannot sync - expired token",
			account: &models.OAuthAccount{
				SyncStatus:  models.SyncStatusActive,
				TokenExpiry: time.Now().Add(-1 * time.Hour),
			},
			wantCan: false,
		},
		{
			name: "cannot sync - error status",
			account: &models.OAuthAccount{
				SyncStatus:  models.SyncStatusError,
				TokenExpiry: time.Now().Add(1 * time.Hour),
			},
			wantCan: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			can := tt.account.CanSync()
			if can != tt.wantCan {
				t.Errorf("CanSync() = %v, want %v", can, tt.wantCan)
			}
		})
	}
}

// TestOAuthProvider 測試 OAuth 提供商檢查
func TestOAuthProvider(t *testing.T) {
	googleAccount := &models.OAuthAccount{
		Provider: models.OAuthProviderGoogle,
	}

	if !googleAccount.IsGmail() {
		t.Error("Expected account to be Gmail")
	}

	if googleAccount.IsOutlook() {
		t.Error("Expected account not to be Outlook")
	}

	outlookAccount := &models.OAuthAccount{
		Provider: models.OAuthProviderOutlook,
	}

	if !outlookAccount.IsOutlook() {
		t.Error("Expected account to be Outlook")
	}

	if outlookAccount.IsGmail() {
		t.Error("Expected account not to be Gmail")
	}
}

// TestMessageIDGeneration 測試 message ID 生成
func TestMessageIDGeneration(t *testing.T) {
	msg := MockGmailMessage("test-id-123")

	if msg.ProviderMessageID != "test-id-123" {
		t.Errorf("Expected ProviderMessageID 'test-id-123', got %s", msg.ProviderMessageID)
	}

	if msg.ID == uuid.Nil {
		t.Error("Expected ID to be generated")
	}
}

// 輔助函數：檢查 slice 是否包含字串
func containsStringInSlice(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func TestSliceContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !containsStringInSlice(slice, "a") {
		t.Error("Expected slice to contain 'a'")
	}

	if containsStringInSlice(slice, "d") {
		t.Error("Expected slice not to contain 'd'")
	}
}
