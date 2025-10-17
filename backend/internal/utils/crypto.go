package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	// ErrInvalidKeySize 當加密金鑰長度不正確時返回
	ErrInvalidKeySize = errors.New("encryption key must be 32 bytes for AES-256")
	// ErrInvalidCiphertext 當密文無效時返回
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrEncryptionKeyNotSet 當加密金鑰未設定時返回
	ErrEncryptionKeyNotSet = errors.New("ENCRYPTION_KEY environment variable not set")
)

// encryptionKey 全域加密金鑰（從環境變數載入）
var encryptionKey []byte

// InitCrypto 初始化加密工具，從環境變數載入金鑰
// 開發環境下如果沒有設定，會生成臨時金鑰（⚠️ 僅限開發使用）
func InitCrypto() error {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	env := os.Getenv("ENV")

	// 開發環境下如果沒有設定，生成臨時金鑰
	if keyStr == "" {
		if env == "production" {
			return ErrEncryptionKeyNotSet
		}

		// ⚠️ 警告：使用臨時金鑰（開發環境）
		fmt.Println("⚠️  WARNING: Using temporary encryption key (development only)")
		key := make([]byte, 32)
		// 填充隨機數據
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("failed to generate temporary key: %w", err)
		}
		encryptionKey = key
		return nil
	}

	// 解碼 base64 編碼的金鑰
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		// 開發環境下，如果解碼失敗，生成臨時金鑰
		if env != "production" {
			fmt.Printf("⚠️  WARNING: Failed to decode encryption key, using temporary key (development only)\n")
			key = make([]byte, 32)
			if _, err := rand.Read(key); err != nil {
				return fmt.Errorf("failed to generate temporary key: %w", err)
			}
			encryptionKey = key
			return nil
		}
		return fmt.Errorf("failed to decode encryption key: %w", err)
	}

	// 檢查金鑰長度（AES-256 需要 32 bytes）
	if len(key) != 32 {
		// 開發環境下，如果長度不對，生成臨時金鑰
		if env != "production" {
			fmt.Printf("⚠️  WARNING: Invalid key length (%d bytes), using temporary key (development only)\n", len(key))
			key = make([]byte, 32)
			if _, err := rand.Read(key); err != nil {
				return fmt.Errorf("failed to generate temporary key: %w", err)
			}
			encryptionKey = key
			return nil
		}
		return ErrInvalidKeySize
	}

	encryptionKey = key
	return nil
}

// GenerateEncryptionKey 生成新的 AES-256 加密金鑰（32 bytes）並以 base64 編碼
// 這個函數用於生成新的 ENCRYPTION_KEY，應該只在設置時執行一次
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// Encrypt 使用 AES-256-GCM 加密資料
// 返回 base64 編碼的密文
func Encrypt(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", ErrEncryptionKeyNotSet
	}

	if plaintext == "" {
		return "", nil
	}

	// 創建 AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 使用 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 生成隨機 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 加密資料（nonce 會被附加在密文前面）
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回 base64 編碼的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 使用 AES-256-GCM 解密資料
// 輸入為 base64 編碼的密文
func Decrypt(ciphertextBase64 string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", ErrEncryptionKeyNotSet
	}

	if ciphertextBase64 == "" {
		return "", nil
	}

	// 解碼 base64
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// 創建 AES cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// 使用 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// 檢查密文長度
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	// 分離 nonce 和實際密文
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptTokens 加密 OAuth tokens
// 這是一個便利函數，用於同時加密 access token 和 refresh token
func EncryptTokens(accessToken, refreshToken string) (encryptedAccess, encryptedRefresh string, err error) {
	encryptedAccess, err = Encrypt(accessToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encryptedRefresh, err = Encrypt(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	return encryptedAccess, encryptedRefresh, nil
}

// DecryptTokens 解密 OAuth tokens
// 這是一個便利函數，用於同時解密 access token 和 refresh token
func DecryptTokens(encryptedAccess, encryptedRefresh string) (accessToken, refreshToken string, err error) {
	accessToken, err = Decrypt(encryptedAccess)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt access token: %w", err)
	}

	refreshToken, err = Decrypt(encryptedRefresh)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
