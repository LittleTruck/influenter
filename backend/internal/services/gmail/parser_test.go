package gmail

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/gmail/v1"
)

// getTestGmailMessage 返回測試用的 Gmail Message
func getTestGmailMessage() *gmail.Message {
	return &gmail.Message{
		Id:           "test-message-id-123",
		ThreadId:     "test-thread-id-456",
		HistoryId:    12345,
		LabelIds:     []string{"INBOX", "UNREAD"},
		Snippet:      "This is a test email snippet",
		SizeEstimate: 1024,
		InternalDate: time.Now().UnixMilli(),
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "From", Value: "Test Sender <sender@example.com>"},
				{Name: "To", Value: "receiver@example.com"},
				{Name: "Cc", Value: "cc@example.com"},
				{Name: "Subject", Value: "Test Subject"},
				{Name: "Date", Value: time.Now().Format(time.RFC1123Z)},
				{Name: "Message-ID", Value: "<message-id@example.com>"},
			},
			Body: &gmail.MessagePartBody{
				Data: base64.URLEncoding.EncodeToString([]byte("Test email body text")),
			},
			MimeType: "text/plain",
		},
	}
}

func TestParseMessage(t *testing.T) {
	oauthAccountID := uuid.New()
	gmailMsg := getTestGmailMessage()

	email, err := ParseMessage(gmailMsg, oauthAccountID)
	var _ *models.Email = email // 確保 models 包被使用
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if email == nil {
		t.Fatal("Expected email to be non-nil")
	}

	if email.ProviderMessageID != "test-message-id-123" {
		t.Errorf("Expected ProviderMessageID 'test-message-id-123', got %s", email.ProviderMessageID)
	}

	if email.FromEmail != "sender@example.com" {
		t.Errorf("Expected FromEmail 'sender@example.com', got %s", email.FromEmail)
	}

	if email.FromName == nil || *email.FromName != "Test Sender" {
		t.Errorf("Expected FromName 'Test Sender', got %v", email.FromName)
	}

	if email.Subject == nil || *email.Subject != "Test Subject" {
		t.Errorf("Expected Subject 'Test Subject', got %v", email.Subject)
	}

	if email.IsRead {
		t.Error("Expected email to be unread (has UNREAD label)")
	}
}

func TestParseMessage_HTMLBody(t *testing.T) {
	oauthAccountID := uuid.New()

	htmlContent := base64.URLEncoding.EncodeToString([]byte("<html><body>HTML Content</body></html>"))
	textContent := base64.URLEncoding.EncodeToString([]byte("Plain text content"))

	gmailMsg := &gmail.Message{
		Id:       "test-html-id",
		ThreadId: "test-thread",
		LabelIds: []string{"INBOX"},
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "From", Value: "sender@example.com"},
				{Name: "To", Value: "receiver@example.com"},
				{Name: "Subject", Value: "Test HTML"},
			},
			MimeType: "multipart/alternative",
			Parts: []*gmail.MessagePart{
				{
					MimeType: "text/plain",
					Body: &gmail.MessagePartBody{
						Data: textContent,
					},
				},
				{
					MimeType: "text/html",
					Body: &gmail.MessagePartBody{
						Data: htmlContent,
					},
				},
			},
		},
	}

	email, err := ParseMessage(gmailMsg, oauthAccountID)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if email.BodyText == nil || *email.BodyText != "Plain text content" {
		t.Errorf("Expected BodyText 'Plain text content', got %v", email.BodyText)
	}

	if email.BodyHTML == nil || *email.BodyHTML != "<html><body>HTML Content</body></html>" {
		t.Errorf("Expected BodyHTML '<html><body>HTML Content</body></html>', got %v", email.BodyHTML)
	}
}

func TestParseMessage_WithAttachment(t *testing.T) {
	oauthAccountID := uuid.New()

	gmailMsg := &gmail.Message{
		Id:       "test-attachment-id",
		ThreadId: "test-thread",
		LabelIds: []string{"INBOX"},
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "From", Value: "sender@example.com"},
				{Name: "To", Value: "receiver@example.com"},
				{Name: "Subject", Value: "Test with Attachment"},
			},
			MimeType: "multipart/mixed",
			Parts: []*gmail.MessagePart{
				{
					MimeType: "text/plain",
					Body: &gmail.MessagePartBody{
						Data: base64.URLEncoding.EncodeToString([]byte("Body text")),
					},
				},
				{
					Filename: "test.pdf",
					MimeType: "application/pdf",
					Body: &gmail.MessagePartBody{
						AttachmentId: "attachment-id-123",
						Size:         4096,
					},
				},
			},
		},
	}

	email, err := ParseMessage(gmailMsg, oauthAccountID)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if !email.HasAttachments {
		t.Error("Expected email to have attachments")
	}
}

func TestParseMessage_EmptyPayload(t *testing.T) {
	oauthAccountID := uuid.New()

	gmailMsg := &gmail.Message{
		Id:       "test-empty-id",
		ThreadId: "test-thread",
		LabelIds: []string{"INBOX"},
		Payload:  nil,
	}

	email, err := ParseMessage(gmailMsg, oauthAccountID)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}

	if email == nil {
		t.Fatal("Expected email to be non-nil even with empty payload")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		slice  []string
		item   string
		expect bool
	}{
		{
			name:   "contains item",
			slice:  []string{"a", "b", "c"},
			item:   "b",
			expect: true,
		},
		{
			name:   "does not contain item",
			slice:  []string{"a", "b", "c"},
			item:   "d",
			expect: false,
		},
		{
			name:   "empty slice",
			slice:  []string{},
			item:   "a",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expect {
				t.Errorf("contains() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect *string
	}{
		{
			name:   "non-empty string",
			input:  "test",
			expect: stringPtr("test"),
		},
		{
			name:   "empty string should return nil",
			input:  "",
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringPtr(tt.input)
			if (result == nil) != (tt.expect == nil) {
				t.Errorf("stringPtr() = %v, want %v", result, tt.expect)
			}
			if result != nil && *result != *tt.expect {
				t.Errorf("stringPtr() = %v, want %v", *result, *tt.expect)
			}
		})
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name    string
		labels  []string
		wantCat string
	}{
		{
			name:    "has personal category",
			labels:  []string{"INBOX", "CATEGORY_PERSONAL"},
			wantCat: "CATEGORY_PERSONAL",
		},
		{
			name:    "has social category",
			labels:  []string{"INBOX", "CATEGORY_SOCIAL"},
			wantCat: "CATEGORY_SOCIAL",
		},
		{
			name:    "no category, should return personal",
			labels:  []string{"INBOX"},
			wantCat: LabelCategoryPersonal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat := GetCategory(tt.labels)
			if cat != tt.wantCat {
				t.Errorf("GetCategory() = %v, want %v", cat, tt.wantCat)
			}
		})
	}
}

func TestIsInCategory(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
		cat    string
		expect bool
	}{
		{
			name:   "in category",
			labels: []string{"INBOX", "CATEGORY_PERSONAL"},
			cat:    "CATEGORY_PERSONAL",
			expect: true,
		},
		{
			name:   "not in category",
			labels: []string{"INBOX"},
			cat:    "CATEGORY_PERSONAL",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInCategory(tt.labels, tt.cat)
			if result != tt.expect {
				t.Errorf("IsInCategory() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestExtractPlainText(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "simple html",
			html: "<html><body>Hello World</body></html>",
			want: "Hello World",
		},
		{
			name: "with script tags",
			html: "<html><body><script>alert('test')</script>Hello</body></html>",
			want: "Hello",
		},
		{
			name: "with style tags",
			html: "<html><body><style>body{color:red}</style>Hello</body></html>",
			want: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPlainText(tt.html)
			if result != tt.want {
				t.Errorf("ExtractPlainText() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestRemoveTagAndContent(t *testing.T) {
	tests := []struct {
		name string
		html string
		tag  string
		want string
	}{
		{
			name: "remove script tag",
			html: "<script>alert('test')</script><div>Content</div>",
			tag:  "script",
			want: "<div>Content</div>",
		},
		{
			name: "remove style tag",
			html: "<style>body{color:red}</style><p>Text</p>",
			tag:  "style",
			want: "<p>Text</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeTagAndContent(tt.html, tt.tag)
			if result != tt.want {
				t.Errorf("removeTagAndContent() = %v, want %v", result, tt.want)
			}
		})
	}
}
