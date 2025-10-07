# Influenter - 環境變數設定指南

## 建立 .env 檔案

請在專案根目錄建立 `.env` 檔案，並填入以下內容：

```env
# ======================
# Google OAuth 設定
# ======================
# 取得方式：
# 1. 前往 https://console.cloud.google.com/
# 2. 建立新專案或選擇現有專案
# 3. 啟用 Gmail API
# 4. 建立 OAuth 2.0 憑證（桌面應用程式）
# 5. 設定授權重定向 URI: http://localhost:8080/api/v1/auth/google/callback

GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# ======================
# JWT 設定
# ======================
# 生成方式: openssl rand -base64 32
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# ======================
# 資料加密金鑰 (32 bytes)
# ======================
# 用於加密 Gmail tokens 和 OpenAI API keys
# 生成方式: openssl rand -hex 32
ENCRYPTION_KEY=your-32-byte-encryption-key-here-change-this

# ======================
# 前端設定
# ======================
NUXT_PUBLIC_API_BASE=http://localhost:8080/api/v1
```

## 前端環境變數

另外在 `frontend/` 目錄建立 `.env` 檔案：

```env
NUXT_PUBLIC_API_BASE=http://localhost:8080/api/v1
```

## 生成密鑰的命令

```bash
# 生成 JWT Secret
openssl rand -base64 32

# 生成 Encryption Key (32 bytes hex)
openssl rand -hex 32
```

## 取得 Google OAuth 憑證的步驟

### 1. 前往 Google Cloud Console
https://console.cloud.google.com/

### 2. 建立專案
- 點擊專案下拉選單
- 選擇「新增專案」
- 輸入專案名稱：`Influenter`
- 點擊「建立」

### 3. 啟用 Gmail API
- 在搜尋框搜尋「Gmail API」
- 點擊「啟用」

### 4. 設定 OAuth 同意畫面
- 前往「API 和服務」>「OAuth 同意畫面」
- 選擇「外部」（測試階段）
- 填寫應用程式名稱：`Influenter`
- 填寫使用者支援電子郵件（您的 Email）
- 填寫開發人員聯絡資訊
- 點擊「儲存並繼續」

### 5. 新增測試使用者
- 在「測試使用者」區段
- 點擊「+ ADD USERS」
- 輸入您的 Gmail 帳號
- 點擊「儲存」

### 6. 建立 OAuth 2.0 憑證
- 前往「API 和服務」>「憑證」
- 點擊「+ 建立憑證」>「OAuth 用戶端 ID」
- 應用程式類型：「網路應用程式」
- 名稱：`Influenter Web Client`
- 已授權的重新導向 URI：
  - `http://localhost:8080/api/v1/auth/google/callback`
  - `http://localhost:3000/auth/callback` (如需)
- 點擊「建立」

### 7. 複製憑證
- 從彈出視窗複製「用戶端 ID」和「用戶端密鑰」
- 貼到 `.env` 檔案中的對應欄位

### 8. 設定 Gmail API Scopes
在程式碼中已設定以下 scopes：
- `https://www.googleapis.com/auth/gmail.readonly` - 讀取郵件
- `https://www.googleapis.com/auth/gmail.send` - 寄送郵件

這些會在使用者授權時自動請求。

## 安全性注意事項

⚠️ **重要**：
- `.env` 檔案已加入 `.gitignore`，請勿 commit 到 Git
- 生產環境請使用更強的密鑰
- 定期更換密鑰
- 不要在程式碼中 hard-code 任何密鑰

## 驗證設定

建立好 `.env` 後，執行以下命令驗證：

```bash
# 檢查檔案是否存在
ls -la .env

# 檢查 Docker Compose 是否能讀取環境變數
docker-compose config
```

