package utils

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestGenerateEncryptionKey(t *testing.T) {
	key, err := GenerateEncryptionKey()
	if err != nil {
		t.Fatalf("GenerateEncryptionKey failed: %v", err)
	}

	// 檢查是否為有效的 base64
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		t.Fatalf("Generated key is not valid base64: %v", err)
	}

	// 檢查解碼後長度是否為 32 bytes (AES-256)
	if len(decoded) != 32 {
		t.Fatalf("Expected key length 32, got %d", len(decoded))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// 生成測試用金鑰
	testKey, _ := GenerateEncryptionKey()
	os.Setenv("ENCRYPTION_KEY", testKey)
	defer os.Unsetenv("ENCRYPTION_KEY")

	// 初始化加密工具
	if err := InitCrypto(); err != nil {
		t.Fatalf("InitCrypto failed: %v", err)
	}

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"Simple text", "hello world"},
		{"Empty string", ""},
		{"Long text", "This is a much longer text that should still be encrypted and decrypted correctly without any issues."},
		{"Special characters", "!@#$%^&*()_+-={}[]|\\:;\"'<>,.?/~`"},
		{"Unicode", "你好世界 🌍 مرحبا العالم"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 加密
			encrypted, err := Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// 空字串應該返回空字串
			if tc.plaintext == "" && encrypted != "" {
				t.Fatalf("Expected empty string for empty input, got %s", encrypted)
			}

			if tc.plaintext == "" {
				return
			}

			// 解密
			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// 比較
			if decrypted != tc.plaintext {
				t.Fatalf("Expected %s, got %s", tc.plaintext, decrypted)
			}
		})
	}
}

func TestEncryptDecryptTokens(t *testing.T) {
	// 生成測試用金鑰
	testKey, _ := GenerateEncryptionKey()
	os.Setenv("ENCRYPTION_KEY", testKey)
	defer os.Unsetenv("ENCRYPTION_KEY")

	// 初始化加密工具
	if err := InitCrypto(); err != nil {
		t.Fatalf("InitCrypto failed: %v", err)
	}

	accessToken := "access_token_123456789"
	refreshToken := "refresh_token_987654321"

	// 加密
	encryptedAccess, encryptedRefresh, err := EncryptTokens(accessToken, refreshToken)
	if err != nil {
		t.Fatalf("EncryptTokens failed: %v", err)
	}

	// 解密
	decryptedAccess, decryptedRefresh, err := DecryptTokens(encryptedAccess, encryptedRefresh)
	if err != nil {
		t.Fatalf("DecryptTokens failed: %v", err)
	}

	// 比較
	if decryptedAccess != accessToken {
		t.Fatalf("Access token mismatch: expected %s, got %s", accessToken, decryptedAccess)
	}

	if decryptedRefresh != refreshToken {
		t.Fatalf("Refresh token mismatch: expected %s, got %s", refreshToken, decryptedRefresh)
	}
}

func TestEncryptWithoutInit(t *testing.T) {
	// 重置 encryptionKey
	encryptionKey = nil

	_, err := Encrypt("test")
	if err != ErrEncryptionKeyNotSet {
		t.Fatalf("Expected ErrEncryptionKeyNotSet, got %v", err)
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	// 生成測試用金鑰
	testKey, _ := GenerateEncryptionKey()
	os.Setenv("ENCRYPTION_KEY", testKey)
	defer os.Unsetenv("ENCRYPTION_KEY")

	// 初始化加密工具
	if err := InitCrypto(); err != nil {
		t.Fatalf("InitCrypto failed: %v", err)
	}

	// 無效的 base64
	_, err := Decrypt("invalid base64!!!")
	if err == nil {
		t.Fatal("Expected error for invalid base64")
	}

	// 太短的密文
	_, err = Decrypt(base64.StdEncoding.EncodeToString([]byte("short")))
	if err != ErrInvalidCiphertext {
		t.Fatalf("Expected ErrInvalidCiphertext, got %v", err)
	}
}

func TestInitCryptoErrors(t *testing.T) {
	// 儲存原始的 ENV 值
	originalEnv := os.Getenv("ENV")

	// 設定為 production 環境以觸發嚴格的驗證
	os.Setenv("ENV", "production")
	defer func() {
		// 恢復原始的 ENV 值
		if originalEnv != "" {
			os.Setenv("ENV", originalEnv)
		} else {
			os.Unsetenv("ENV")
		}
	}()

	// 重置 encryptionKey
	encryptionKey = nil

	// 測試未設定環境變數
	os.Unsetenv("ENCRYPTION_KEY")
	err := InitCrypto()
	if err != ErrEncryptionKeyNotSet {
		t.Fatalf("Expected ErrEncryptionKeyNotSet, got %v", err)
	}

	// 測試無效的 base64
	os.Setenv("ENCRYPTION_KEY", "invalid base64!!!")
	err = InitCrypto()
	if err == nil {
		t.Fatal("Expected error for invalid base64")
	}

	// 測試錯誤的金鑰長度
	shortKey := base64.StdEncoding.EncodeToString([]byte("short"))
	os.Setenv("ENCRYPTION_KEY", shortKey)
	err = InitCrypto()
	if err != ErrInvalidKeySize {
		t.Fatalf("Expected ErrInvalidKeySize, got %v", err)
	}
}
