package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGmailGetStatus_Connected 測試已連接 Gmail
func TestGmailGetStatus_Connected(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	oauthAccount := createTestOAuthAccount(t, db, userID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/gmail/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["connected"])
	assert.Equal(t, oauthAccount.Email, response["email"])
}

// TestGmailGetStatus_NotConnected 測試未連接 Gmail
func TestGmailGetStatus_NotConnected(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	_, token, _ := createTestUser(t, db, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/gmail/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["connected"])
}

// TestGmailGetStatus_NoToken 測試無效用戶（無 token）
func TestGmailGetStatus_NoToken(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/gmail/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

// TestGmailTriggerSync_NotConnected 測試未連接 Gmail 時觸發同步
func TestGmailTriggerSync_NotConnected(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	_, token, _ := createTestUser(t, db, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/gmail/sync", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "not_connected", response.Error)
}

// TestGmailTriggerSync_Success 測試成功觸發同步
func TestGmailTriggerSync_Success(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	createTestOAuthAccount(t, db, userID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/gmail/sync", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Sync started", response["message"])
	assert.Equal(t, "syncing", response["status"])
}

// TestGmailTriggerSync_NoToken 測試未認證觸發同步
func TestGmailTriggerSync_NoToken(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/gmail/sync", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

// TestGmailDisconnect_Success 測試成功斷開連接
func TestGmailDisconnect_Success(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	userID, token, _ := createTestUser(t, db, cfg)
	createTestOAuthAccount(t, db, userID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/gmail/disconnect", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Gmail account disconnected successfully", response["message"])
}

// TestGmailDisconnect_NotConnected 測試未連接時斷開連接
func TestGmailDisconnect_NotConnected(t *testing.T) {
	db, router, cfg := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	_, token, _ := createTestUser(t, db, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/gmail/disconnect", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "not_connected", response.Error)
}

// TestGmailDisconnect_NoToken 測試未認證斷開連接
func TestGmailDisconnect_NoToken(t *testing.T) {
	db, router, _ := setupTestRouter(t)
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/gmail/disconnect", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}
