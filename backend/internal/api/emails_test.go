package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/designcomb/influenter-backend/internal/models"
	"github.com/designcomb/influenter-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// createTestOAuthAccount 創建測試用的 OAuth Account
func createTestOAuthAccount(t *testing.T, db *gorm.DB, userID uuid.UUID) *models.OAuthAccount {
	now := time.Now()
	// 以可解密格式存入（使用 utils.Encrypt）
	encAT, err := utils.Encrypt("test-access-token")
	if err != nil {
		t.Fatalf("Failed to encrypt access token: %v", err)
	}
	encRT, err := utils.Encrypt("test-refresh-token")
	if err != nil {
		t.Fatalf("Failed to encrypt refresh token: %v", err)
	}
	account := &models.OAuthAccount{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     models.OAuthProviderGoogle,
		ProviderID:   "google-id-123",
		Email:        "gmail@example.com",
		AccessToken:  encAT,
		RefreshToken: encRT,
		TokenExpiry:  now.Add(24 * time.Hour),
		SyncStatus:   models.SyncStatusActive,
	}

	if err := db.Create(account).Error; err != nil {
		t.Fatalf("Failed to create oauth account: %v", err)
	}

	return account
}

// createTestEmail 創建測試用的郵件
func createTestEmail(t *testing.T, db *gorm.DB, oauthAccountID uuid.UUID) *models.Email {
	subject := "Test Subject"
	snippet := "Test snippet"
	email := &models.Email{
		OAuthAccountID:    oauthAccountID,
		ProviderMessageID: uuid.New().String(),
		FromEmail:         "sender@example.com",
		Subject:           &subject,
		Snippet:           &snippet,
		ReceivedAt:        time.Now(),
		IsRead:            false,
		Labels:            []string{"INBOX", "UNREAD"},
	}

	if err := db.Create(email).Error; err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	return email
}

// TestListEmails_Success 測試成功列出郵件
func TestListEmails_Success(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	oauthAccount := createTestOAuthAccount(t, db, userID)

	// 創建兩封郵件
	createTestEmail(t, db, oauthAccount.ID)
	createTestEmail(t, db, oauthAccount.ID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "emails")
	assert.Contains(t, response, "pagination")
}

// TestListEmails_NoToken 測試未認證的情況
func TestListEmails_NoToken(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

// TestListEmails_FilterByReadStatus 測試篩選已讀/未讀郵件
func TestListEmails_FilterByReadStatus(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	oauthAccount := createTestOAuthAccount(t, db, userID)

	// 創建一封已讀郵件
	email1 := createTestEmail(t, db, oauthAccount.ID)
	email1.IsRead = true
	db.Save(email1)

	// 創建一封未讀郵件
	createTestEmail(t, db, oauthAccount.ID)

	// 測試篩選已讀郵件
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails?page=1&page_size=20&is_read=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "emails")

	// 驗證只返回已讀郵件
	emails, ok := response["emails"].([]interface{})
	if ok && len(emails) > 0 {
		email := emails[0].(map[string]interface{})
		assert.Equal(t, true, email["is_read"])
	}
}

// TestListEmails_Pagination 測試分頁功能
func TestListEmails_Pagination(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	oauthAccount := createTestOAuthAccount(t, db, userID)

	// 創建 3 封郵件
	createTestEmail(t, db, oauthAccount.ID)
	createTestEmail(t, db, oauthAccount.ID)
	createTestEmail(t, db, oauthAccount.ID)

	// 測試第一頁，每頁 2 條
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails?page=1&page_size=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	emails := response["emails"].([]interface{})
	assert.LessOrEqual(t, len(emails), 2)

	pagination := response["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(2), pagination["page_size"])
}

// TestListEmails_EmptyResult 測試空結果
func TestListEmails_EmptyResult(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	_, token, _ := createTestUser(t, db, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	emails := response["emails"].([]interface{})
	assert.Equal(t, 0, len(emails))
}

// TestGetEmail_Success 測試成功獲取郵件詳情
func TestGetEmail_Success(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	oauthAccount := createTestOAuthAccount(t, db, userID)
	email := createTestEmail(t, db, oauthAccount.ID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails/"+email.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var emailResponse models.EmailDetailResponse
	err := json.Unmarshal(w.Body.Bytes(), &emailResponse)
	assert.NoError(t, err)
	assert.Equal(t, email.ID, emailResponse.ID)
}

// TestGetEmail_InvalidID 測試無效 ID
func TestGetEmail_InvalidID(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	_, token, _ := createTestUser(t, db, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails/invalid-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// TestGetEmail_NotFound 測試郵件不存在
func TestGetEmail_NotFound(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	_, token, _ := createTestUser(t, db, cfg)

	// 使用一個不存在的 UUID
	nonExistentID := uuid.New()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails/"+nonExistentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

// TestGetEmail_NotOwned 測試不屬於用戶的郵件
func TestGetEmail_NotOwned(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 創建兩個用戶
	user1ID, _, _ := createTestUser(t, db, cfg)
	oauthAccount1 := createTestOAuthAccount(t, db, user1ID)
	email := createTestEmail(t, db, oauthAccount1.ID)

	// 創建第二個用戶和其 token
	user2ID, token2, _ := createTestUser(t, db, cfg)
	createTestOAuthAccount(t, db, user2ID)

	// 使用第二個用戶的 token 嘗試獲取第一個用戶的郵件
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/emails/"+email.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token2)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

// TestUpdateEmail_MarkAsRead 測試標記已讀
func TestUpdateEmail_MarkAsRead(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	oauthAccount := createTestOAuthAccount(t, db, userID)
	email := createTestEmail(t, db, oauthAccount.ID)

	// 更新為已讀
	body := `{"is_read": true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/v1/emails/"+email.ID.String(), bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var emailResponse models.EmailDetailResponse
	err := json.Unmarshal(w.Body.Bytes(), &emailResponse)
	assert.NoError(t, err)
	assert.Equal(t, true, emailResponse.IsRead)
}

// TestUpdateEmail_MarkAsUnread 測試標記未讀
func TestUpdateEmail_MarkAsUnread(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	oauthAccount := createTestOAuthAccount(t, db, userID)
	email := createTestEmail(t, db, oauthAccount.ID)

	// 先標記為已讀
	email.IsRead = true
	db.Save(email)

	// 更新為未讀
	body := `{"is_read": false}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/v1/emails/"+email.ID.String(), bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var emailResponse models.EmailDetailResponse
	err := json.Unmarshal(w.Body.Bytes(), &emailResponse)
	assert.NoError(t, err)
	assert.Equal(t, false, emailResponse.IsRead)
}

// TestUpdateEmail_InvalidID 測試無效 ID
func TestUpdateEmail_InvalidID(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	_, token, _ := createTestUser(t, db, cfg)

	body := `{"is_read": true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/v1/emails/invalid-id", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// TestUpdateEmail_NotOwned 測試不屬於用戶的郵件
func TestUpdateEmail_NotOwned(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 創建兩個用戶
	user1ID, _, _ := createTestUser(t, db, cfg)
	oauthAccount1 := createTestOAuthAccount(t, db, user1ID)
	email := createTestEmail(t, db, oauthAccount1.ID)

	_, token2, _ := createTestUser(t, db, cfg)

	// 使用第二個用戶的 token 嘗試更新第一個用戶的郵件
	body := `{"is_read": true}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/api/v1/emails/"+email.ID.String(), bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+token2)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}
