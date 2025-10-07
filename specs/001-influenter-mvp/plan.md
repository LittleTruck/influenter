# Influenter MVP - 實作計劃

> **目的**：詳細的開發實作步驟與時程規劃  
> **預估時程**：12 週（3 個月）  
> **最後更新**：2025-10-07

---

## 📅 整體時程規劃

| Phase | 內容 | 週數 | 累積週數 |
|-------|------|------|----------|
| Phase 0 | 專案初始化與環境設置 | 1 週 | 1 週 |
| Phase 1 | 後端基礎架構與認證 | 2 週 | 3 週 |
| Phase 2 | Gmail 整合與郵件同步 | 2 週 | 5 週 |
| Phase 3 | AI 分析與分類功能 | 2 週 | 7 週 |
| Phase 4 | 案件管理核心功能 | 2 週 | 9 週 |
| Phase 5 | 回覆生成與進階功能 | 2 週 | 11 週 |
| Phase 6 | 測試、優化與部署 | 1 週 | 12 週 |

---

## 🏗️ Phase 0: 專案初始化與環境設置（Week 1）

### 目標
建立開發環境、專案結構與基礎設定

### 任務清單

#### 後端 (Go)
- [ ] 初始化 Go 專案
  ```bash
  mkdir backend && cd backend
  go mod init github.com/yourusername/influenter-backend
  ```
- [ ] 安裝核心依賴
  ```bash
  go get github.com/gin-gonic/gin
  go get gorm.io/gorm
  go get gorm.io/driver/postgres
  go get github.com/golang-jwt/jwt/v5
  go get golang.org/x/oauth2
  go get google.golang.org/api/gmail/v1
  go get github.com/sashabaranov/go-openai
  go get github.com/hibiken/asynq
  ```
- [ ] 建立專案目錄結構
  ```
  backend/
  ├── cmd/
  │   ├── server/          # API server
  │   ├── worker/          # 背景任務 worker
  │   └── migrate/         # 資料庫遷移工具
  ├── internal/
  │   ├── api/             # API handlers
  │   │   ├── auth/
  │   │   ├── cases/
  │   │   ├── emails/
  │   │   └── replies/
  │   ├── models/          # GORM models
  │   ├── services/        # 業務邏輯
  │   │   ├── gmail/
  │   │   ├── openai/
  │   │   └── auth/
  │   ├── workers/         # 背景任務
  │   ├── middleware/      # Gin middleware
  │   ├── config/          # 設定載入
  │   └── utils/           # 工具函數
  ├── migrations/          # SQL 遷移檔案
  ├── .env.example
  ├── go.mod
  └── go.sum
  ```
- [ ] 建立 `.env.example`
  ```env
  # Database
  DATABASE_URL=postgresql://user:password@localhost:5432/influenter?sslmode=disable
  
  # Google OAuth
  GOOGLE_CLIENT_ID=your-client-id
  GOOGLE_CLIENT_SECRET=your-client-secret
  GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
  
  # JWT
  JWT_SECRET=your-secret-key
  
  # Encryption
  ENCRYPTION_KEY=32-byte-encryption-key
  
  # Redis
  REDIS_ADDR=localhost:6379
  
  # Server
  PORT=8080
  ENVIRONMENT=development
  ```

#### 前端 (Nuxt 3)
- [ ] 初始化 Nuxt 3 專案
  ```bash
  npx nuxi@latest init frontend
  cd frontend
  ```
- [ ] 安裝依賴
  ```bash
  npm install vuetify @mdi/font pinia @vite-pwa/nuxt
  ```
- [ ] 建立專案結構
  ```
  frontend/
  ├── assets/
  ├── components/
  │   ├── layout/
  │   ├── cases/
  │   ├── emails/
  │   └── common/
  ├── composables/
  │   └── useAPI.ts
  ├── layouts/
  │   └── default.vue
  ├── pages/
  │   ├── index.vue
  │   ├── login.vue
  │   ├── cases/
  │   ├── emails/
  │   └── settings/
  ├── plugins/
  │   └── vuetify.ts
  ├── stores/
  │   ├── auth.ts
  │   ├── cases.ts
  │   └── emails.ts
  ├── nuxt.config.ts
  └── package.json
  ```
- [ ] 設定 Vuetify 3（參考 research.md）
- [ ] 設定 Pinia store
- [ ] 建立 `.env.example`
  ```env
  NUXT_PUBLIC_API_BASE=http://localhost:8080/api/v1
  ```

#### 資料庫
- [ ] 安裝 PostgreSQL 15+
- [ ] 建立資料庫
  ```sql
  CREATE DATABASE influenter;
  CREATE USER influenter_user WITH PASSWORD 'your-password';
  GRANT ALL PRIVILEGES ON DATABASE influenter TO influenter_user;
  ```
- [ ] 安裝 Redis（用於 Asynq）

#### 開發工具
- [ ] 設定 Git repository
- [ ] 建立 `.gitignore`
- [ ] 設定 VSCode / IDE
- [ ] 安裝 Postman / Insomnia（測試 API）

### 驗收標準
- [x] 後端可以成功啟動（`go run cmd/server/main.go`）
- [x] 前端可以成功啟動（`npm run dev`）
- [x] 可以連線到 PostgreSQL
- [x] Redis 正常運作

---

## 🔐 Phase 1: 後端基礎架構與認證（Week 2-3）

### 目標
建立 API 基礎架構、資料庫連線、Google OAuth 認證

### 任務清單

#### 資料庫遷移
- [ ] 建立遷移工具（`cmd/migrate/main.go`）
- [ ] 撰寫第一個遷移檔案（users 表）
  ```sql
  -- migrations/001_create_users_table.up.sql
  CREATE TABLE users (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      email VARCHAR(255) NOT NULL UNIQUE,
      name VARCHAR(100),
      avatar_url TEXT,
      google_id VARCHAR(255) UNIQUE,
      ai_reply_tone VARCHAR(50) DEFAULT 'professional',
      timezone VARCHAR(50) DEFAULT 'Asia/Taipei',
      notification_prefs JSONB,
      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
      last_login_at TIMESTAMP
  );
  
  CREATE INDEX idx_users_email ON users(email);
  CREATE INDEX idx_users_google_id ON users(google_id);
  ```
- [ ] 建立其他核心表（參考 data-model.md）
- [ ] 測試遷移（up / down）

#### GORM Models
- [ ] 建立 `internal/models/user.go`
- [ ] 建立 `internal/models/gmail_account.go`
- [ ] 建立 `internal/models/email.go`
- [ ] 建立 `internal/models/case.go`
- [ ] 建立 `internal/models/task.go`
- [ ] 建立 `internal/models/reply.go`
- [ ] 建立基礎 model 方法（Create, Find, Update, Delete）

#### 設定管理
- [ ] 實作 `internal/config/config.go`
  ```go
  type Config struct {
      Database DatabaseConfig
      Google   GoogleOAuthConfig
      JWT      JWTConfig
      Server   ServerConfig
      Redis    RedisConfig
  }
  ```
- [ ] 從環境變數載入設定

#### 資料庫連線
- [ ] 實作 PostgreSQL 連線
  ```go
  func InitDB(config DatabaseConfig) (*gorm.DB, error)
  ```
- [ ] 實作連線池設定
- [ ] 實作 health check

#### JWT 認證
- [ ] 實作 JWT token 生成
  ```go
  func GenerateToken(userID string) (string, error)
  ```
- [ ] 實作 JWT token 驗證
- [ ] 實作 auth middleware
  ```go
  func AuthMiddleware() gin.HandlerFunc
  ```

#### Google OAuth
- [ ] 實作 OAuth 配置
- [ ] 實作 `/auth/google/login` handler
  ```go
  func HandleGoogleLogin(c *gin.Context)
  ```
- [ ] 實作 `/auth/google/callback` handler
  ```go
  func HandleGoogleCallback(c *gin.Context)
  ```
- [ ] 取得使用者資訊並建立/更新 user 記錄
- [ ] 返回 JWT token

#### API 路由設置
- [ ] 建立 router
  ```go
  func SetupRouter(db *gorm.DB) *gin.Engine
  ```
- [ ] 設置 CORS middleware
- [ ] 設置 logger middleware
- [ ] 設置 recovery middleware
- [ ] 分組路由（/api/v1）

#### 前端認證整合
- [ ] 建立 auth store（Pinia）
- [ ] 實作登入頁面
- [ ] 實作 OAuth 回調處理
- [ ] 儲存 JWT token（cookie）
- [ ] 實作 auth composable
  ```typescript
  const { user, isAuthenticated, login, logout } = useAuth()
  ```
- [ ] 實作 route guard（未登入跳轉）

### API 端點
- [ ] `GET /auth/google/login`
- [ ] `GET /auth/google/callback`
- [ ] `GET /auth/me`
- [ ] `POST /auth/logout`

### 驗收標準
- [x] 可以用 Google 帳號登入
- [x] 成功取得 JWT token
- [x] 前端可以顯示使用者資訊
- [x] 未登入時會重定向到登入頁
- [x] 資料庫正確記錄使用者資料

---

## 📧 Phase 2: Gmail 整合與郵件同步（Week 4-5）

### 目標
整合 Gmail API，實作郵件同步功能

### 任務清單

#### Gmail OAuth 權限
- [ ] 更新 Google OAuth scopes
  ```go
  gmail.GmailReadonlyScope,
  gmail.GmailSendScope,
  ```
- [ ] 在 callback 中儲存 Gmail tokens
- [ ] 建立 gmail_accounts 表遷移
- [ ] 實作 token 加密/解密
  ```go
  func Encrypt(plaintext string) (string, error)
  func Decrypt(ciphertext string) (string, error)
  ```

#### Gmail Service
- [ ] 實作 `internal/services/gmail/client.go`
  ```go
  type GmailService struct {
      service *gmail.Service
  }
  
  func NewGmailService(tokens *oauth2.Token) (*GmailService, error)
  ```
- [ ] 實作取得郵件列表
  ```go
  func (s *GmailService) ListMessages(query string) ([]*gmail.Message, error)
  ```
- [ ] 實作取得單封郵件詳情
  ```go
  func (s *GmailService) GetMessage(messageID string) (*gmail.Message, error)
  ```
- [ ] 實作寄送郵件
  ```go
  func (s *GmailService) SendMessage(to, subject, body string) error
  ```

#### 郵件解析
- [ ] 實作郵件解析器
  ```go
  func ParseGmailMessage(msg *gmail.Message) (*models.Email, error)
  ```
- [ ] 解析寄件者、收件者
- [ ] 解析主旨
- [ ] 解析純文字內容
- [ ] 解析 HTML 內容
- [ ] 處理 Gmail labels

#### 郵件同步邏輯
- [ ] 實作首次同步（最近 30 天）
  ```go
  func InitialSync(userID string) error
  ```
- [ ] 實作增量同步（只抓新郵件）
  ```go
  func IncrementalSync(userID string) error
  ```
- [ ] 實作去重邏輯（檢查 gmail_message_id）
- [ ] 實作錯誤處理與重試

#### 背景任務（Asynq）
- [ ] 建立 worker (`cmd/worker/main.go`)
- [ ] 實作郵件同步任務
  ```go
  func HandleEmailSyncTask(ctx context.Context, t *asynq.Task) error
  ```
- [ ] 設定 cron schedule（每 5 分鐘）
  ```go
  scheduler.Register("*/5 * * * *", NewEmailSyncTask())
  ```
- [ ] 實作任務監控

#### API Endpoints
- [ ] `GET /gmail/status` - 取得同步狀態
- [ ] `POST /gmail/sync` - 手動觸發同步
- [ ] `DELETE /gmail/disconnect` - 斷開連接
- [ ] `GET /emails` - 取得郵件列表（分頁）
- [ ] `GET /emails/:id` - 取得郵件詳情
- [ ] `PATCH /emails/:id` - 更新郵件（標記已讀）

#### 前端整合
- [ ] 建立 emails store
- [ ] 實作郵件列表頁面
  - 表格/卡片顯示
  - 分頁
  - 篩選（已讀/未讀）
  - 搜尋
- [ ] 實作郵件詳情頁面
  - 顯示完整內容
  - AI 分析結果（Phase 3 實作）
  - 關聯到案件
- [ ] 實作同步狀態顯示
  - 同步中動畫
  - 上次同步時間
  - 手動同步按鈕

### 驗收標準
- [x] 授權後可以看到 Gmail 信件
- [x] 每 5 分鐘自動同步
- [x] 手動同步正常運作
- [x] 郵件列表正確顯示
- [x] 可以查看郵件詳情
- [x] 同步狀態正確更新

---

## 🤖 Phase 3: AI 分析與分類功能（Week 6-7）

### 目標
整合 OpenAI API，實作郵件智慧分析

### 任務清單

#### OpenAI Service
- [ ] 實作 `internal/services/openai/client.go`
  ```go
  type OpenAIService struct {
      client *openai.Client
  }
  
  func NewOpenAIService(apiKey string) *OpenAIService
  ```
- [ ] 實作郵件分類功能
  ```go
  func (s *OpenAIService) ClassifyEmail(content string) (*EmailClassification, error)
  ```
- [ ] 設計 system prompt（參考 research.md）
- [ ] 實作結構化輸出（Function Calling）
- [ ] 實作錯誤處理與重試

#### AI 分析模型
- [ ] 建立 `ai_analysis` 表遷移
- [ ] 建立 GORM model
- [ ] 實作分析結果儲存
- [ ] 記錄 token 使用量與成本

#### 郵件分析流程
- [ ] 在郵件同步後觸發 AI 分析
- [ ] 實作分析背景任務
  ```go
  func HandleEmailAnalysisTask(ctx context.Context, t *asynq.Task) error
  ```
- [ ] 批次處理未分析的郵件
- [ ] 實作分析結果快取

#### 使用者修正機制
- [ ] 實作修正 API
  ```go
  PATCH /emails/:id/ai-analysis
  ```
- [ ] 記錄原始 AI 判斷與修正後結果
- [ ] 標記 `user_corrected` 為 true
- [ ] 為未來訓練保留資料

#### 自動建立案件
- [ ] 當 AI 判斷為「合作邀約」且信心指標 > 0.8 時
- [ ] 自動建立狀態為 `to_confirm` 的案件
- [ ] 填入抽取的資訊（品牌、金額、日期等）
- [ ] 關聯郵件到案件

#### API Endpoints
- [ ] `POST /emails/:id/analyze` - 手動觸發分析
- [ ] `PATCH /emails/:id/ai-analysis` - 修正 AI 分析結果

#### 前端整合
- [ ] 在郵件列表顯示 AI 分類標籤
- [ ] 顯示 AI 信心指標（進度條或百分比）
- [ ] 實作修正功能
  - 下拉選單修改分類
  - 編輯品牌名稱、金額等
  - 儲存修正
- [ ] 顯示「自動建立的案件」提示

### 驗收標準
- [x] 郵件同步後自動觸發 AI 分析
- [x] AI 分類準確率 > 85%（基於測試集）
- [x] 資訊抽取正確（品牌、金額、日期）
- [x] 信心指標合理
- [x] 使用者可以修正 AI 結果
- [x] 高信心的合作邀約自動建立案件

---

## 📊 Phase 4: 案件管理核心功能（Week 8-9）

### 目標
實作完整的案件 CRUD 與任務管理

### 任務清單

#### 案件 CRUD
- [ ] `POST /cases` - 手動建立案件
- [ ] `GET /cases` - 取得案件列表（分頁、篩選）
- [ ] `GET /cases/:id` - 取得案件詳情
- [ ] `PATCH /cases/:id` - 更新案件
- [ ] `DELETE /cases/:id` - 軟刪除案件

#### 案件與郵件關聯
- [ ] `POST /cases/:id/emails` - 關聯郵件到案件
- [ ] `GET /cases/:id/emails` - 取得案件相關郵件
- [ ] `DELETE /cases/:id/emails/:emailId` - 移除關聯

#### 案件狀態管理
- [ ] 實作狀態流程驗證
  ```
  to_confirm → in_progress → completed
       ↓              ↓
  cancelled      cancelled
  ```
- [ ] 實作狀態變更記錄（case_updates 表）
- [ ] 自動記錄變更歷史

#### 任務管理
- [ ] `POST /cases/:caseId/tasks` - 建立任務
- [ ] `GET /cases/:caseId/tasks` - 取得案件任務
- [ ] `PATCH /tasks/:id` - 更新任務
- [ ] `DELETE /tasks/:id` - 刪除任務
- [ ] `POST /tasks/:id/complete` - 標記完成

#### 自動任務生成
- [ ] 當案件狀態變更為 `in_progress` 且有 deadline_date 時
- [ ] 根據 deadline 自動建立任務
  - 前 7 天：腳本提交
  - 前 3 天：初稿完成
  - 當天：正式交稿

#### 任務提醒
- [ ] 實作提醒背景任務
  ```go
  func HandleTaskReminderTask(ctx context.Context, t *asynq.Task) error
  ```
- [ ] 每日檢查即將到期的任務（前 1 天）
- [ ] 建立通知記錄

#### 前端 - 案件列表頁
- [ ] 實作案件卡片/表格顯示
  - 品牌名稱
  - 案件標題
  - 狀態標籤（顏色區分）
  - 報價金額
  - 截止日期
  - 進度指示（任務完成度）
- [ ] 實作篩選
  - 按狀態篩選
  - 按品牌篩選
  - 按日期範圍篩選
- [ ] 實作搜尋
- [ ] 實作排序（最新、截止日、金額）

#### 前端 - 案件詳情頁
- [ ] 左側：基本資訊卡片
  - 編輯按鈕
  - 狀態下拉選單
  - 品牌、金額、日期等
- [ ] 中間：時程與任務
  - 任務列表（checkbox）
  - 新增任務按鈕
  - 拖曳排序
- [ ] 右側：郵件時間軸
  - 按時間排序
  - 點擊展開郵件內容
  - 快速回覆按鈕
- [ ] 底部：備註區域

#### 前端 - 建立/編輯案件表單
- [ ] 表單驗證（Vuelidate / Zod）
- [ ] 日期選擇器（Vuetify date picker）
- [ ] 自動完成（品牌名稱）
- [ ] 標籤輸入（chips）

#### 前端 - 日曆視圖
- [ ] 整合日曆元件（Vuetify / FullCalendar）
- [ ] 顯示所有案件的重要日期
- [ ] 點擊日期跳轉到案件
- [ ] 月/週視圖切換

### 驗收標準
- [x] 可以手動建立案件
- [x] 可以查看、編輯、刪除案件
- [x] 案件與郵件正確關聯
- [x] 可以建立、管理任務
- [x] 狀態變更為進行中時自動建立任務
- [x] 提醒功能正常運作
- [x] 前端 UI 完整且易用

---

## ✍️ Phase 5: 回覆生成與進階功能（Week 10-11）

### 目標
實作 AI 回覆生成、報價方案、通知系統

### 任務清單

#### 回覆生成
- [ ] 實作 `internal/services/openai/reply_generator.go`
  ```go
  func GenerateReply(email *models.Email, tone string, context string) (string, error)
  ```
- [ ] 設計回覆生成 prompt
  - 根據郵件內容
  - 考慮案件狀態
  - 應用使用者語氣偏好
  - 加入使用者補充資訊
- [ ] 實作 `POST /replies/generate` API
- [ ] 實作 `POST /replies/send` API（透過 Gmail API）
- [ ] 記錄回覆歷史

#### 報價方案管理
- [ ] 實作報價方案 CRUD
  - `GET /pricing-packages`
  - `POST /pricing-packages`
  - `PATCH /pricing-packages/:id`
  - `DELETE /pricing-packages/:id`
- [ ] 實作方案插入回覆功能
  ```go
  func FormatPricingPackage(pkg *models.PricingPackage) string
  ```

#### 通知系統
- [ ] 建立通知 service
  ```go
  func CreateNotification(userID, type, title, message string) error
  ```
- [ ] 實作通知觸發點
  - 新郵件（合作邀約）
  - 新案件建立
  - 任務即將到期
  - 任務已逾期
- [ ] 實作通知 API
  - `GET /notifications`
  - `PATCH /notifications/:id/read`
  - `POST /notifications/read-all`

#### 統計儀表板
- [ ] 實作 `GET /stats/dashboard` API
  - 案件統計（各狀態數量）
  - 收入統計（總額、已付、待付）
  - 郵件統計
  - 任務統計
  - Top 品牌
- [ ] 實作 `GET /stats/revenue` API
  - 月收入趨勢
  - 案件類型分布

#### 前端 - 回覆生成
- [ ] 建立回覆生成對話框（Dialog）
  - 選擇語氣（下拉選單）
  - 補充資訊（文字框）
  - 選擇報價方案（可選）
  - 生成按鈕
- [ ] 顯示 AI 生成結果
  - 可編輯文字框
  - 重新生成按鈕
  - 複製按鈕
  - 寄送按鈕
- [ ] 實作寄送確認
- [ ] 顯示寄送狀態（Loading、成功、失敗）

#### 前端 - 報價方案管理
- [ ] 建立方案列表頁面
  - 卡片顯示
  - 啟用/停用開關
  - 編輯/刪除按鈕
- [ ] 建立方案表單
  - 項目動態新增（Array input）
  - 預覽顯示

#### 前端 - 通知中心
- [ ] App bar 通知圖示（Badge 顯示未讀數）
- [ ] 通知下拉選單（Dropdown）
  - 最近 5 則通知
  - 標記已讀按鈕
  - 查看全部連結
- [ ] 通知列表頁面
  - 已讀/未讀區分
  - 點擊跳轉到相關頁面

#### 前端 - 儀表板
- [ ] 建立 Dashboard 頁面
  - 統計卡片（案件、收入、郵件）
  - 收入趨勢圖表（Chart.js / ApexCharts）
  - 最近案件列表
  - 即將到期任務
  - Top 品牌

#### 前端 - 設定頁面
- [ ] 個人資料編輯
- [ ] Gmail 連接狀態
- [ ] OpenAI API Key 設定
  - 輸入框（password type）
  - 測試連接按鈕
- [ ] AI 回覆語氣偏好
- [ ] 通知偏好設定（Checkboxes）
- [ ] 時區設定

### 驗收標準
- [x] 可以生成 AI 回覆草稿
- [x] 可以編輯並寄送回覆
- [x] 報價方案可以正常管理
- [x] 通知正確觸發並顯示
- [x] 儀表板統計資料正確
- [x] 設定頁面功能完整

---

## 🧪 Phase 6: 測試、優化與部署（Week 12）

### 目標
完整測試、效能優化、準備上線

### 任務清單

#### 後端測試
- [ ] 撰寫單元測試
  - Services 層測試
  - API handlers 測試
  - 目標覆蓋率 > 70%
- [ ] 撰寫整合測試
  - Gmail API 整合（使用測試帳號）
  - OpenAI API（使用 mock）
  - 資料庫操作
- [ ] API 端點測試（Postman Collection）

#### 前端測試
- [ ] 元件測試（Vitest）
- [ ] E2E 測試（Playwright）
  - 登入流程
  - 建立案件
  - 生成回覆

#### 效能優化
- [ ] 資料庫查詢優化
  - 檢查 N+1 問題
  - 新增必要索引
  - 使用 EXPLAIN ANALYZE
- [ ] API 回應時間優化
  - 實作快取（Redis）
  - 資料庫連線池調校
- [ ] 前端效能
  - 圖片優化
  - Lazy loading
  - Code splitting

#### 安全性檢查
- [ ] OWASP Top 10 檢查
- [ ] SQL Injection 防護（GORM 已內建）
- [ ] XSS 防護
- [ ] CSRF 防護
- [ ] Rate limiting
- [ ] API Key 加密確認

#### 文件撰寫
- [ ] API 文件（Swagger / OpenAPI）
- [ ] 使用者手冊
  - 如何取得 Google OAuth credentials
  - 如何取得 OpenAI API Key
  - 基本操作教學
- [ ] 開發者指南
  - 本地開發設置
  - 資料庫遷移
  - 部署流程

#### 部署準備
- [ ] Dockerfile（前後端分開）
  ```dockerfile
  # Backend Dockerfile
  FROM golang:1.21-alpine AS builder
  # ... build steps
  
  FROM alpine:latest
  # ... runtime
  ```
- [ ] docker-compose.yml（開發環境）
  ```yaml
  services:
    db:
      image: postgres:15-alpine
    redis:
      image: redis:7-alpine
    backend:
      build: ./backend
    frontend:
      build: ./frontend
  ```
- [ ] CI/CD 設定（GitHub Actions）
  - 自動測試
  - 自動部署（Render / Railway / Fly.io）
- [ ] 環境變數管理
- [ ] 資料庫備份策略

#### 上線檢查清單
- [ ] 環境變數全部設定
- [ ] HTTPS 啟用
- [ ] 資料庫遷移執行
- [ ] Redis 正常運作
- [ ] Worker 正常運作
- [ ] 日誌系統設定
- [ ] 監控系統設定（選用）
- [ ] 錯誤追蹤（Sentry 等，選用）

### 驗收標準
- [x] 所有測試通過
- [x] 效能符合需求（API < 200ms）
- [x] 安全性檢查通過
- [x] 文件完整
- [x] 可以成功部署到生產環境
- [x] 生產環境功能正常

---

## 📚 技術債務與未來改進

### 已知限制（MVP 階段）
- 單一 Gmail 帳號支援
- 不支援多人協作
- AI 無個性化學習
- 固定 5 分鐘同步頻率

### Phase 2 可能的功能
- 多 Gmail 帳號
- AI 風格學習
- 團隊協作功能
- 原生手機 App
- 自訂同步頻率
- 實時同步（Gmail Push Notifications）
- 更多 AI 功能（摘要、智慧排程）
- 整合其他郵件服務（Outlook 等）

---

## ✅ 完整功能檢查清單

### 認證與使用者
- [ ] Google OAuth 登入
- [ ] JWT 認證
- [ ] 使用者資料管理
- [ ] 登出功能

### Gmail 整合
- [ ] Gmail 授權
- [ ] 郵件同步（自動 + 手動）
- [ ] 郵件列表與詳情
- [ ] 同步狀態顯示

### AI 功能
- [ ] 郵件自動分類
- [ ] 資訊抽取
- [ ] AI 信心指標
- [ ] 使用者修正機制
- [ ] AI 回覆生成

### 案件管理
- [ ] 案件 CRUD
- [ ] 案件狀態管理
- [ ] 案件與郵件關聯
- [ ] 自動建立案件
- [ ] 案件搜尋與篩選

### 任務與時程
- [ ] 任務 CRUD
- [ ] 自動任務生成
- [ ] 任務提醒
- [ ] 日曆視圖

### 回覆與溝通
- [ ] AI 生成回覆
- [ ] 編輯回覆
- [ ] 寄送回覆
- [ ] 回覆歷史

### 進階功能
- [ ] 報價方案管理
- [ ] 通知系統
- [ ] 統計儀表板
- [ ] 設定管理

### UI/UX
- [ ] 響應式設計（RWD）
- [ ] 暗色模式
- [ ] PWA 功能
- [ ] 載入狀態
- [ ] 錯誤提示

---

**實作計劃完成！準備開始開發！** 🚀

