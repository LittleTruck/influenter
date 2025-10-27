package gmail

import (
	"testing"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/google/uuid"
)

func TestOAuthAccount_ToResponse(t *testing.T) {
	now := time.Now()
	account := &models.OAuthAccount{
		ID:          uuid.New(),
		Provider:    models.OAuthProviderGoogle,
		Email:       "test@example.com",
		TokenExpiry: now.Add(24 * time.Hour),
		SyncStatus:  models.SyncStatusActive,
		CreatedAt:   now,
	}

	response := account.ToResponse()

	if response.ID != account.ID {
		t.Errorf("Expected ID %v, got %v", account.ID, response.ID)
	}

	if response.Provider != "google" {
		t.Errorf("Expected Provider 'google', got %s", response.Provider)
	}

	if response.Email != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got %s", response.Email)
	}

	if response.SyncStatus != "active" {
		t.Errorf("Expected SyncStatus 'active', got %s", response.SyncStatus)
	}

	if response.IsTokenExpired {
		t.Error("Expected token not to be expired")
	}
}

func TestEmail_ToListResponse(t *testing.T) {
	subject := "Test Subject"
	snippet := "Test snippet"
	fromName := "Test Sender"

	email := &models.Email{
		ID:             uuid.New(),
		FromEmail:      "sender@example.com",
		FromName:       &fromName,
		Subject:        &subject,
		Snippet:        &snippet,
		ReceivedAt:     time.Now(),
		IsRead:         false,
		HasAttachments: true,
		Labels:         []string{"INBOX", "UNREAD"},
	}

	response := email.ToListResponse()

	if response.ID != email.ID {
		t.Errorf("Expected ID %v, got %v", email.ID, response.ID)
	}

	if response.FromEmail != "sender@example.com" {
		t.Errorf("Expected FromEmail 'sender@example.com', got %s", response.FromEmail)
	}

	if response.Subject == nil || *response.Subject != "Test Subject" {
		t.Errorf("Expected Subject 'Test Subject', got %v", response.Subject)
	}

	if response.HasAttachments != true {
		t.Error("Expected HasAttachments to be true")
	}

	if len(response.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(response.Labels))
	}
}

func TestEmail_ToDetailResponse(t *testing.T) {
	subject := "Test Subject"
	snippet := "Test snippet"
	body := "Test body"
	html := "<html>Test</html>"
	fromName := "Test Sender"
	toEmail := "recipient@example.com"
	threadID := "thread-id-123"

	email := &models.Email{
		ID:                uuid.New(),
		OAuthAccountID:    uuid.New(),
		ProviderMessageID: "msg-id-123",
		ThreadID:          &threadID,
		FromEmail:         "sender@example.com",
		FromName:          &fromName,
		ToEmail:           &toEmail,
		Subject:           &subject,
		BodyText:          &body,
		BodyHTML:          &html,
		Snippet:           &snippet,
		ReceivedAt:        time.Now(),
		IsRead:            false,
		HasAttachments:    true,
		Labels:            []string{"INBOX"},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	response := email.ToDetailResponse()

	if response.ID != email.ID {
		t.Errorf("Expected ID %v, got %v", email.ID, response.ID)
	}

	if response.ProviderMessageID != "msg-id-123" {
		t.Errorf("Expected ProviderMessageID 'msg-id-123', got %s", response.ProviderMessageID)
	}

	if response.ThreadID == nil || *response.ThreadID != "thread-id-123" {
		t.Errorf("Expected ThreadID 'thread-id-123', got %v", response.ThreadID)
	}

	if response.BodyHTML == nil || *response.BodyHTML != "<html>Test</html>" {
		t.Errorf("Expected BodyHTML '<html>Test</html>', got %v", response.BodyHTML)
	}
}

func TestEmail_HasLabel(t *testing.T) {
	tests := []struct {
		name   string
		email  *models.Email
		label  string
		expect bool
	}{
		{
			name: "has label",
			email: &models.Email{
				Labels: []string{"INBOX", "UNREAD"},
			},
			label:  "INBOX",
			expect: true,
		},
		{
			name: "does not have label",
			email: &models.Email{
				Labels: []string{"INBOX"},
			},
			label:  "STARRED",
			expect: false,
		},
		{
			name: "empty labels",
			email: &models.Email{
				Labels: []string{},
			},
			label:  "INBOX",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.email.HasLabel(tt.label)
			if result != tt.expect {
				t.Errorf("HasLabel() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestEmail_IsImportant(t *testing.T) {
	tests := []struct {
		name   string
		email  *models.Email
		expect bool
	}{
		{
			name: "is important",
			email: &models.Email{
				Labels: []string{"INBOX", "IMPORTANT"},
			},
			expect: true,
		},
		{
			name: "not important",
			email: &models.Email{
				Labels: []string{"INBOX"},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.email.IsImportant()
			if result != tt.expect {
				t.Errorf("IsImportant() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestEmail_IsInInbox(t *testing.T) {
	tests := []struct {
		name   string
		email  *models.Email
		expect bool
	}{
		{
			name: "in inbox",
			email: &models.Email{
				Labels: []string{"INBOX", "UNREAD"},
			},
			expect: true,
		},
		{
			name: "not in inbox",
			email: &models.Email{
				Labels: []string{"SENT"},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.email.IsInInbox()
			if result != tt.expect {
				t.Errorf("IsInInbox() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestSyncStatus_String(t *testing.T) {
	tests := []struct {
		status models.SyncStatus
		want   string
	}{
		{models.SyncStatusActive, "active"},
		{models.SyncStatusPaused, "paused"},
		{models.SyncStatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := string(tt.status)
			if result != tt.want {
				t.Errorf("SyncStatus.String() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestOAuthProvider_String(t *testing.T) {
	tests := []struct {
		provider models.OAuthProvider
		want     string
	}{
		{models.OAuthProviderGoogle, "google"},
		{models.OAuthProviderOutlook, "outlook"},
		{models.OAuthProviderApple, "apple"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := string(tt.provider)
			if result != tt.want {
				t.Errorf("OAuthProvider.String() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestEmailQueryParams_SetDefaults 測試 EmailQueryParams 的預設值設定
func TestEmailQueryParams_SetDefaults(t *testing.T) {
	params := &models.EmailQueryParams{}

	params.SetDefaults()

	if params.Page != 1 {
		t.Errorf("Expected Page to be 1, got %d", params.Page)
	}

	if params.PageSize != 20 {
		t.Errorf("Expected PageSize to be 20, got %d", params.PageSize)
	}

	if params.SortBy != "received_at" {
		t.Errorf("Expected SortBy to be 'received_at', got %s", params.SortBy)
	}

	if params.SortOrder != "desc" {
		t.Errorf("Expected SortOrder to be 'desc', got %s", params.SortOrder)
	}
}
