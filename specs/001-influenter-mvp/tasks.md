# Influenter MVP - 任務分解

> **目的**：將實作計劃分解為可執行的具體任務  
> **格式**：每個任務都可以在 4 小時內完成  
> **總預估**：約 480 小時（12 週 × 40 小時）

---

## 📋 任務編號規則

格式：`PHASE-MODULE-NUMBER`

範例：
- `P0-SETUP-001`：Phase 0 的設置任務
- `P1-AUTH-001`：Phase 1 的認證任務

---

## 🎯 Phase 0: 專案初始化（預估: 40h）

### 後端初始化 (16h)

**P0-BACKEND-001**: 建立 Go 專案結構（2h）
- [ ] 初始化 `go.mod`
- [ ] 建立目錄結構（cmd, internal, migrations）
- [ ] 設定 `.gitignore`
- [ ] 建立 `.env.example`

**P0-BACKEND-002**: 安裝核心依賴（1h）
- [ ] 安裝 Gin
- [ ] 安裝 GORM + PostgreSQL driver
- [ ] 安裝 OAuth2 相關套件
- [ ] 安裝 OpenAI SDK
- [ ] 安裝 Asynq

**P0-BACKEND-003**: 實作設定載入（3h）
- [ ] 建立 `internal/config/config.go`
- [ ] 定義 Config struct
- [ ] 從環境變數載入
- [ ] 驗證必要設定
- [ ] 單元測試

**P0-BACKEND-004**: 實作資料庫連線（3h）
- [ ] 建立 `internal/database/db.go`
- [ ] PostgreSQL 連線邏輯
- [ ] 連線池設定
- [ ] Health check 功能
- [ ] 單元測試

**P0-BACKEND-005**: 建立基礎 API server（3h）
- [ ] 建立 `cmd/server/main.go`
- [ ] 初始化 Gin router
- [ ] 設定基礎 middleware（logger, recovery）
- [ ] 設定 CORS
- [ ] Health check endpoint

**P0-BACKEND-006**: 建立遷移工具（4h）
- [ ] 建立 `cmd/migrate/main.go`
- [ ] 實作 up/down 遷移邏輯
- [ ] 載入 migrations 資料夾
- [ ] 測試遷移功能

### 前端初始化 (16h)

**P0-FRONTEND-001**: 建立 Nuxt 3 專案（2h）
- [ ] 初始化 Nuxt 專案
- [ ] 安裝核心依賴
- [ ] 建立目錄結構
- [ ] 設定 `.gitignore`

**P0-FRONTEND-002**: 設定 Vuetify 3（3h）
- [ ] 安裝 Vuetify 3
- [ ] 建立 `plugins/vuetify.ts`
- [ ] 設定主題（light/dark）
- [ ] 設定圖示（MDI）
- [ ] 測試元件顯示

**P0-FRONTEND-003**: 設定 Pinia（2h）
- [ ] 安裝 Pinia
- [ ] 在 `nuxt.config.ts` 啟用
- [ ] 建立範例 store
- [ ] 測試 store 運作

**P0-FRONTEND-004**: 建立基礎 Layout（3h）
- [ ] 建立 `layouts/default.vue`
- [ ] App Bar
- [ ] Navigation Drawer（側邊選單）
- [ ] Footer
- [ ] 響應式設計

**P0-FRONTEND-005**: 建立 API composable（3h）
- [ ] 建立 `composables/useAPI.ts`
- [ ] 設定 base URL
- [ ] 自動加上 JWT token
- [ ] 統一錯誤處理
- [ ] 測試

**P0-FRONTEND-006**: 建立基礎頁面（3h）
- [ ] 首頁 (`pages/index.vue`)
- [ ] 登入頁 (`pages/login.vue`)
- [ ] 404 頁面
- [ ] Loading 元件

### 基礎設施 (8h)

**P0-INFRA-001**: 設置 PostgreSQL（2h）
- [ ] 安裝 PostgreSQL
- [ ] 建立資料庫
- [ ] 建立使用者
- [ ] 授予權限
- [ ] 測試連線

**P0-INFRA-002**: 設置 Redis（1h）
- [ ] 安裝 Redis
- [ ] 啟動服務
- [ ] 測試連線

**P0-INFRA-003**: 設定開發工具（2h）
- [ ] 安裝 Postman/Insomnia
- [ ] 建立 API collection
- [ ] 設定 VSCode extensions
- [ ] 設定 code formatter

**P0-INFRA-004**: 設定 Git repository（2h）
- [ ] 建立 GitHub repository
- [ ] 設定 branch protection
- [ ] 撰寫 README
- [ ] 第一次 commit

**P0-INFRA-005**: 建立 Docker Compose（1h）
- [ ] 撰寫 `docker-compose.yml`
- [ ] PostgreSQL service
- [ ] Redis service
- [ ] 測試啟動

---

## 🔐 Phase 1: 認證系統（預估: 80h）

### 資料庫遷移 (12h)

**P1-DB-001**: 建立 users 表遷移（2h）
- [ ] 撰寫 `001_create_users_table.up.sql`
- [ ] 撰寫 down 遷移
- [ ] 執行遷移
- [ ] 驗證 schema

**P1-DB-002**: 建立 gmail_accounts 表遷移（2h）
- [ ] 撰寫遷移檔案
- [ ] 外鍵設定
- [ ] 索引設定
- [ ] 測試

**P1-DB-003**: 建立其他核心表（8h）
- [ ] emails 表
- [ ] cases 表
- [ ] tasks 表
- [ ] replies 表
- [ ] 其他表（見 data-model.md）

### GORM Models (16h)

**P1-MODEL-001**: 建立 User model（2h）
- [ ] 定義 struct
- [ ] GORM tags
- [ ] Relations
- [ ] 基礎方法（Create, Find）
- [ ] 單元測試

**P1-MODEL-002**: 建立 GmailAccount model（2h）
- [ ] 定義 struct
- [ ] Relations
- [ ] 加密/解密方法
- [ ] 單元測試

**P1-MODEL-003**: 建立 Email model（3h）
- [ ] 定義 struct
- [ ] Relations
- [ ] 查詢方法
- [ ] 單元測試

**P1-MODEL-004**: 建立 Case model（3h）
- [ ] 定義 struct
- [ ] Relations
- [ ] 狀態管理方法
- [ ] 單元測試

**P1-MODEL-005**: 建立其他 models（6h）
- [ ] Task model
- [ ] Reply model
- [ ] Notification model
- [ ] 其他 models

### JWT 認證 (12h)

**P1-AUTH-001**: 實作 JWT token 生成（3h）
- [ ] 建立 `internal/services/auth/jwt.go`
- [ ] GenerateToken 函數
- [ ] 設定過期時間
- [ ] 加入 user claims
- [ ] 單元測試

**P1-AUTH-002**: 實作 JWT token 驗證（3h）
- [ ] ParseToken 函數
- [ ] 驗證簽章
- [ ] 檢查過期
- [ ] 抽取 user ID
- [ ] 單元測試

**P1-AUTH-003**: 實作 Auth Middleware（3h）
- [ ] 建立 `internal/middleware/auth.go`
- [ ] 從 Header 取得 token
- [ ] 驗證 token
- [ ] 設定 user context
- [ ] 錯誤處理

**P1-AUTH-004**: 實作 GetCurrentUser helper（3h）
- [ ] 從 context 取得 user ID
- [ ] 查詢資料庫
- [ ] 快取處理
- [ ] 單元測試

### Google OAuth (20h)

**P1-OAUTH-001**: 設定 OAuth 配置（2h）
- [ ] 在 Google Console 建立專案
- [ ] 設定 OAuth consent screen
- [ ] 建立 credentials
- [ ] 設定 redirect URI

**P1-OAUTH-002**: 實作 OAuth service（4h）
- [ ] 建立 `internal/services/auth/google_oauth.go`
- [ ] 初始化 OAuth config
- [ ] 設定 scopes
- [ ] GenerateAuthURL 函數

**P1-OAUTH-003**: 實作登入 handler（3h）
- [ ] 建立 `internal/api/auth/google.go`
- [ ] HandleLogin 函數
- [ ] 生成 state token（CSRF 防護）
- [ ] 重定向到 Google

**P1-OAUTH-004**: 實作 callback handler（5h）
- [ ] HandleCallback 函數
- [ ] 驗證 state token
- [ ] 交換 code 取得 token
- [ ] 取得使用者資訊
- [ ] 建立或更新 user 記錄
- [ ] 生成 JWT token

**P1-OAUTH-005**: 實作 /auth/me endpoint（2h）
- [ ] HandleGetMe 函數
- [ ] 取得當前使用者
- [ ] 返回使用者資訊
- [ ] 錯誤處理

**P1-OAUTH-006**: 實作登出（2h）
- [ ] HandleLogout 函數
- [ ] 清除 token（前端處理）
- [ ] 記錄登出時間

**P1-OAUTH-007**: 單元測試與整合測試（2h）
- [ ] Mock Google API
- [ ] 測試各種情境
- [ ] 測試錯誤處理

### 前端認證整合 (20h)

**P1-FE-001**: 建立 auth store（4h）
- [ ] 建立 `stores/auth.ts`
- [ ] State 定義（user, token）
- [ ] loginWithGoogle action
- [ ] logout action
- [ ] fetchUserProfile action
- [ ] Persistence（cookie）

**P1-FE-002**: 實作登入頁面（4h）
- [ ] UI 設計（Vuetify Card）
- [ ] Logo 與標語
- [ ] Google 登入按鈕
- [ ] Loading 狀態
- [ ] 錯誤提示

**P1-FE-003**: 實作 OAuth callback 處理（3h）
- [ ] 建立 callback 頁面
- [ ] 從 URL 取得 code
- [ ] 呼叫後端 API
- [ ] 儲存 token
- [ ] 重定向到首頁

**P1-FE-004**: 實作 auth composable（3h）
- [ ] 建立 `composables/useAuth.ts`
- [ ] isAuthenticated computed
- [ ] currentUser computed
- [ ] login 函數
- [ ] logout 函數

**P1-FE-005**: 實作 route guard（3h）
- [ ] 建立 middleware
- [ ] 檢查認證狀態
- [ ] 未登入重定向到登入頁
- [ ] 已登入避免訪問登入頁

**P1-FE-006**: 更新 Layout（3h）
- [ ] 顯示使用者頭像
- [ ] 使用者選單（Dropdown）
- [ ] 登出按鈕
- [ ] 條件顯示（登入/未登入）

---

## 📧 Phase 2: Gmail 整合（預估: 80h）

### Gmail OAuth 與 Token 管理 (16h)

**P2-GMAIL-001**: 更新 OAuth scopes（1h）
- [ ] 加入 Gmail scopes
- [ ] 測試授權流程

**P2-GMAIL-002**: 實作 token 加密（4h）
- [ ] 建立 `internal/utils/crypto.go`
- [ ] Encrypt 函數（AES-256-GCM）
- [ ] Decrypt 函數
- [ ] 金鑰管理（環境變數）
- [ ] 單元測試

**P2-GMAIL-003**: 儲存 Gmail tokens（3h）
- [ ] 在 OAuth callback 中儲存 tokens
- [ ] 加密後寫入資料庫
- [ ] 關聯到使用者
- [ ] 錯誤處理

**P2-GMAIL-004**: 實作 token refresh（4h）
- [ ] 建立 RefreshToken 函數
- [ ] 檢查 token 過期
- [ ] 自動 refresh
- [ ] 更新資料庫
- [ ] 錯誤處理

**P2-GMAIL-005**: 實作 /gmail/status endpoint（2h）
- [ ] 查詢 gmail_account
- [ ] 返回連接狀態
- [ ] 返回同步資訊

**P2-GMAIL-006**: 實作 /gmail/disconnect endpoint（2h）
- [ ] 刪除 gmail_account 記錄
- [ ] 撤銷 Google token（選用）
- [ ] 返回成功訊息

### Gmail Service (24h)

**P2-SERVICE-001**: 建立 Gmail client（3h）
- [ ] 建立 `internal/services/gmail/client.go`
- [ ] NewGmailService 函數
- [ ] 從 token 建立 service
- [ ] 錯誤處理

**P2-SERVICE-002**: 實作 ListMessages（4h）
- [ ] ListMessages 函數
- [ ] 支援 query 參數
- [ ] 支援分頁
- [ ] 錯誤處理與重試
- [ ] 單元測試（mock）

**P2-SERVICE-003**: 實作 GetMessage（4h）
- [ ] GetMessage 函數
- [ ] 取得完整郵件資訊
- [ ] 錯誤處理
- [ ] 單元測試

**P2-SERVICE-004**: 實作郵件解析器（6h）
- [ ] 建立 `internal/services/gmail/parser.go`
- [ ] ParseMessage 函數
- [ ] 解析 headers（From, To, Subject）
- [ ] 解析 body（text/html）
- [ ] 處理編碼
- [ ] 單元測試

**P2-SERVICE-005**: 實作 SendMessage（4h）
- [ ] SendMessage 函數
- [ ] 建立 MIME 訊息
- [ ] 呼叫 Gmail API
- [ ] 錯誤處理
- [ ] 單元測試

**P2-SERVICE-006**: 實作 batch 操作（選用）（3h）
- [ ] BatchGetMessages
- [ ] 減少 API 呼叫次數
- [ ] 效能優化

### 郵件同步邏輯 (20h)

**P2-SYNC-001**: 實作首次同步（6h）
- [ ] 建立 `internal/services/gmail/sync.go`
- [ ] InitialSync 函數
- [ ] 建立 query（最近 30 天）
- [ ] 批次處理郵件
- [ ] 儲存到資料庫
- [ ] 去重邏輯

**P2-SYNC-002**: 實作增量同步（6h）
- [ ] IncrementalSync 函數
- [ ] 使用 last_sync_at
- [ ] 只抓取新郵件
- [ ] 更新 last_sync_at
- [ ] 錯誤處理

**P2-SYNC-003**: 實作同步狀態管理（3h）
- [ ] 更新 sync_status
- [ ] 記錄 sync_error
- [ ] 更新 last_history_id
- [ ] 單元測試

**P2-SYNC-004**: 實作手動同步（3h）
- [ ] /gmail/sync endpoint
- [ ] 冷卻時間檢查（1 分鐘）
- [ ] 觸發同步任務
- [ ] 返回預估時間

**P2-SYNC-005**: 單元測試與整合測試（2h）
- [ ] Mock Gmail API
- [ ] 測試各種情境
- [ ] 測試錯誤處理

### 背景任務（Asynq）(20h)

**P2-WORKER-001**: 建立 worker 基礎架構（4h）
- [ ] 建立 `cmd/worker/main.go`
- [ ] 初始化 Asynq server
- [ ] 設定 Redis 連線
- [ ] 設定 concurrency

**P2-WORKER-002**: 實作郵件同步任務（4h）
- [ ] 建立 `internal/workers/email_sync.go`
- [ ] 定義 task payload
- [ ] NewEmailSyncTask 函數
- [ ] HandleEmailSyncTask 函數

**P2-WORKER-003**: 設定 cron schedule（2h）
- [ ] 建立 scheduler
- [ ] 註冊每 5 分鐘執行
- [ ] 針對所有使用者
- [ ] 測試

**P2-WORKER-004**: 實作任務監控（選用）（3h）
- [ ] 設定 Asynq web UI
- [ ] 監控任務狀態
- [ ] 查看失敗任務

**P2-WORKER-005**: 錯誤處理與重試（3h）
- [ ] 設定 retry 策略
- [ ] Exponential backoff
- [ ] 記錄錯誤日誌
- [ ] 通知管理員（選用）

**P2-WORKER-006**: 整合測試（4h）
- [ ] 測試任務排程
- [ ] 測試任務執行
- [ ] 測試錯誤重試

### 前端整合 (0h → 將在 Phase 4 完成郵件 UI)

---

## 🤖 Phase 3: AI 分析（預估: 80h）

### OpenAI Service (24h)

**P3-AI-001**: 設定 OpenAI client（2h）
- [ ] 建立 `internal/services/openai/client.go`
- [ ] NewOpenAIService 函數
- [ ] 從使用者設定取得 API Key
- [ ] 錯誤處理

**P3-AI-002**: 設計郵件分類 prompt（4h）
- [ ] 研究最佳 prompt 結構
- [ ] 撰寫 system prompt
- [ ] 定義分類類別
- [ ] 定義輸出格式（JSON schema）
- [ ] 測試與調整

**P3-AI-003**: 實作郵件分類功能（6h）
- [ ] ClassifyEmail 函數
- [ ] 使用 Function Calling
- [ ] 解析結構化輸出
- [ ] 計算信心指標
- [ ] 錯誤處理與重試
- [ ] 單元測試（mock）

**P3-AI-004**: 實作資訊抽取（6h）
- [ ] ExtractInfo 函數
- [ ] 抽取品牌名稱
- [ ] 抽取金額（正則 + AI）
- [ ] 抽取日期
- [ ] 抽取聯絡人資訊
- [ ] 單元測試

**P3-AI-005**: 整合分類與抽取（3h）
- [ ] AnalyzeEmail 函數
- [ ] 同時執行分類與抽取
- [ ] 合併結果
- [ ] 錯誤處理

**P3-AI-006**: 實作成本追蹤（3h）
- [ ] 記錄 token 使用量
- [ ] 計算 API 成本
- [ ] 儲存到資料庫
- [ ] 統計報表（選用）

### AI 分析流程 (20h)

**P3-ANALYSIS-001**: 實作分析 API（4h）
- [ ] POST /emails/:id/analyze
- [ ] 取得郵件
- [ ] 呼叫 AI 分析
- [ ] 儲存結果
- [ ] 返回結果

**P3-ANALYSIS-002**: 實作自動分析（6h）
- [ ] 在郵件同步後觸發
- [ ] 建立分析背景任務
- [ ] HandleEmailAnalysisTask
- [ ] 批次處理未分析郵件

**P3-ANALYSIS-003**: 實作分析結果儲存（4h）
- [ ] 建立 ai_analysis 記錄
- [ ] 更新 email.ai_analyzed
- [ ] 關聯 analysis 到 email
- [ ] 錯誤處理

**P3-ANALYSIS-004**: 實作快取機制（選用）（3h）
- [ ] 相同郵件內容不重複分析
- [ ] 使用 Redis 快取
- [ ] 設定過期時間

**P3-ANALYSIS-005**: 單元測試與整合測試（3h）
- [ ] Mock OpenAI API
- [ ] 測試各種郵件類型
- [ ] 測試錯誤處理

### 使用者修正機制 (12h)

**P3-CORRECT-001**: 實作修正 API（4h）
- [ ] PATCH /emails/:id/ai-analysis
- [ ] 接收修正資料
- [ ] 儲存原始判斷到 original_data
- [ ] 儲存修正到 corrected_data
- [ ] 標記 user_corrected

**P3-CORRECT-002**: 實作修正歷史（3h）
- [ ] 查詢被修正的分析
- [ ] 統計修正率
- [ ] 用於改進 prompt

**P3-CORRECT-003**: 單元測試（2h）
- [ ] 測試修正流程
- [ ] 測試資料儲存

**P3-CORRECT-004**: 實作準確度追蹤（選用）（3h）
- [ ] 計算分類準確率
- [ ] 計算抽取準確率
- [ ] 產生報表

### 自動建立案件 (12h)

**P3-AUTO-001**: 實作自動建立案件邏輯（6h）
- [ ] 判斷條件（類別 + 信心指標）
- [ ] 建立 Case 記錄
- [ ] 填入 AI 抽取的資訊
- [ ] 設定狀態為 to_confirm
- [ ] 關聯郵件

**P3-AUTO-002**: 實作案件建議通知（3h）
- [ ] 建立通知
- [ ] 通知使用者確認
- [ ] 提供編輯連結

**P3-AUTO-003**: 單元測試（3h）
- [ ] 測試自動建立
- [ ] 測試不同情境

### API Endpoints (8h)

**P3-API-001**: 實作 /emails/:id/analyze（2h）
- [ ] Handler 實作
- [ ] 錯誤處理
- [ ] 測試

**P3-API-002**: 實作 /emails/:id/ai-analysis（PATCH）（2h）
- [ ] Handler 實作
- [ ] 驗證輸入
- [ ] 測試

**P3-API-003**: 整合測試（4h）
- [ ] 測試完整流程
- [ ] 測試錯誤情境

### 前端整合（將在 Phase 4 完成）

---

## 📊 Phase 4: 案件管理（預估: 80h）

### 案件 CRUD API (24h)

**P4-CASE-001**: 實作 POST /cases（4h）
- [ ] Handler 實作
- [ ] 輸入驗證
- [ ] 建立案件
- [ ] 返回結果
- [ ] 單元測試

**P4-CASE-002**: 實作 GET /cases（列表）（6h）
- [ ] Handler 實作
- [ ] 分頁處理
- [ ] 篩選（status, brand）
- [ ] 排序
- [ ] 搜尋功能
- [ ] 測試

**P4-CASE-003**: 實作 GET /cases/:id（詳情）（4h）
- [ ] Handler 實作
- [ ] 預載入關聯（emails, tasks）
- [ ] 格式化輸出
- [ ] 測試

**P4-CASE-004**: 實作 PATCH /cases/:id（4h）
- [ ] Handler 實作
- [ ] 部分更新邏輯
- [ ] 驗證輸入
- [ ] 記錄變更歷史
- [ ] 測試

**P4-CASE-005**: 實作 DELETE /cases/:id（軟刪除）（3h）
- [ ] Handler 實作
- [ ] 設定 deleted_at
- [ ] 測試

**P4-CASE-006**: 實作狀態變更記錄（3h）
- [ ] 在更新時自動記錄
- [ ] 建立 case_update 記錄
- [ ] 記錄舊值與新值

### 案件與郵件關聯 (12h)

**P4-RELATION-001**: 實作 POST /cases/:id/emails（4h）
- [ ] Handler 實作
- [ ] 建立關聯記錄
- [ ] 設定 email_type
- [ ] 測試

**P4-RELATION-002**: 實作 GET /cases/:id/emails（4h）
- [ ] Handler 實作
- [ ] 取得關聯郵件
- [ ] 按時間排序
- [ ] 測試

**P4-RELATION-003**: 實作 DELETE /cases/:id/emails/:emailId（2h）
- [ ] Handler 實作
- [ ] 移除關聯
- [ ] 測試

**P4-RELATION-004**: 整合測試（2h）

### 任務管理 API (16h)

**P4-TASK-001**: 實作 POST /cases/:caseId/tasks（4h）
- [ ] Handler 實作
- [ ] 建立任務
- [ ] 驗證輸入
- [ ] 測試

**P4-TASK-002**: 實作 GET /cases/:caseId/tasks（3h）
- [ ] Handler 實作
- [ ] 按日期排序
- [ ] 測試

**P4-TASK-003**: 實作 PATCH /tasks/:id（3h）
- [ ] Handler 實作
- [ ] 更新任務
- [ ] 測試

**P4-TASK-004**: 實作 DELETE /tasks/:id（2h）
- [ ] Handler 實作
- [ ] 刪除任務
- [ ] 測試

**P4-TASK-005**: 實作 POST /tasks/:id/complete（2h）
- [ ] Handler 實作
- [ ] 標記完成
- [ ] 設定 completed_at
- [ ] 測試

**P4-TASK-006**: 整合測試（2h）

### 自動任務生成 (8h)

**P4-AUTO-TASK-001**: 實作自動生成邏輯（5h）
- [ ] 監聽案件狀態變更
- [ ] 當變更為 in_progress
- [ ] 根據 deadline_date 計算任務日期
- [ ] 建立預設任務
- [ ] 測試

**P4-AUTO-TASK-002**: 實作任務模板（選用）（3h）
- [ ] 定義常用任務模板
- [ ] 使用者可自訂模板
- [ ] 套用模板建立任務

### 任務提醒 (12h)

**P4-REMINDER-001**: 實作提醒背景任務（6h）
- [ ] 建立 HandleTaskReminderTask
- [ ] 每日執行（cron）
- [ ] 查詢即將到期的任務（前 N 天）
- [ ] 建立通知
- [ ] 標記 reminder_sent

**P4-REMINDER-002**: 實作通知邏輯（4h）
- [ ] 建立 notification 記錄
- [ ] 設定類型為 task_reminder
- [ ] 關聯到案件與任務

**P4-REMINDER-003**: 測試（2h）
- [ ] 單元測試
- [ ] 整合測試

### 前端 - 案件列表頁 (20h)

**P4-FE-LIST-001**: 建立案件 store（3h）
- [ ] 建立 `stores/cases.ts`
- [ ] fetchCases action
- [ ] createCase action
- [ ] updateCase action
- [ ] deleteCase action
- [ ] 狀態管理

**P4-FE-LIST-002**: 建立案件列表頁面（8h）
- [ ] UI 設計（Vuetify Data Table / Cards）
- [ ] 顯示案件資訊
- [ ] 狀態標籤（chips with colors）
- [ ] 分頁元件
- [ ] 篩選 UI（下拉選單）
- [ ] 搜尋框
- [ ] 排序功能
- [ ] Loading 狀態

**P4-FE-LIST-003**: 實作篩選與搜尋邏輯（4h）
- [ ] 監聽篩選變更
- [ ] 呼叫 API 重新載入
- [ ] 更新 URL query parameters
- [ ] 測試

**P4-FE-LIST-004**: 實作建立案件按鈕與對話框（5h）
- [ ] 新增案件按鈕（FAB）
- [ ] 建立對話框（Dialog）
- [ ] 表單欄位
- [ ] 表單驗證
- [ ] 提交邏輯

### 前端 - 案件詳情頁 (24h)

**P4-FE-DETAIL-001**: 建立詳情頁面結構（4h）
- [ ] 建立 `pages/cases/[id].vue`
- [ ] 三欄 layout（左中右）
- [ ] 響應式設計

**P4-FE-DETAIL-002**: 實作左側基本資訊卡片（6h）
- [ ] 顯示所有欄位
- [ ] 編輯按鈕
- [ ] 編輯模式（Inline editing）
- [ ] 狀態下拉選單
- [ ] 儲存邏輯

**P4-FE-DETAIL-003**: 實作中間任務區域（8h）
- [ ] 任務列表（Vuetify List）
- [ ] Checkbox 顯示狀態
- [ ] 點擊標記完成
- [ ] 新增任務按鈕
- [ ] 新增任務對話框
- [ ] 編輯任務
- [ ] 刪除任務

**P4-FE-DETAIL-004**: 實作右側郵件時間軸（6h）
- [ ] 郵件列表（Timeline style）
- [ ] 顯示寄件者、時間、主旨
- [ ] 點擊展開/收合
- [ ] 顯示完整內容
- [ ] 快速回覆按鈕（Phase 5 實作）

### 前端 - 日曆視圖 (8h)

**P4-FE-CAL-001**: 整合日曆元件（5h）
- [ ] 安裝日曆套件（FullCalendar / Vuetify）
- [ ] 建立 `pages/calendar.vue`
- [ ] 載入所有案件的日期
- [ ] 顯示在日曆上
- [ ] 點擊事件跳轉

**P4-FE-CAL-002**: 實作月/週視圖切換（2h）
- [ ] 視圖切換按鈕
- [ ] 儲存偏好

**P4-FE-CAL-003**: 測試與優化（1h）

---

## ✍️ Phase 5: 回覆生成與進階功能（預估: 80h）

### 回覆生成 (24h)

**P5-REPLY-001**: 設計回覆生成 prompt（4h）
- [ ] 研究最佳 prompt
- [ ] 考慮郵件內容
- [ ] 考慮案件狀態
- [ ] 考慮語氣偏好
- [ ] 測試與調整

**P5-REPLY-002**: 實作 GenerateReply 函數（6h）
- [ ] 建立 `internal/services/openai/reply.go`
- [ ] 建構 prompt
- [ ] 插入郵件內容
- [ ] 插入補充資訊
- [ ] 呼叫 OpenAI API
- [ ] 錯誤處理

**P5-REPLY-003**: 實作報價方案插入（4h）
- [ ] FormatPricingPackage 函數
- [ ] 格式化方案資訊
- [ ] 插入到回覆中
- [ ] 測試

**P5-REPLY-004**: 實作 POST /replies/generate API（4h）
- [ ] Handler 實作
- [ ] 驗證輸入
- [ ] 呼叫生成函數
- [ ] 儲存到資料庫
- [ ] 返回結果

**P5-REPLY-005**: 實作 POST /replies/send API（4h）
- [ ] Handler 實作
- [ ] 取得 reply 記錄
- [ ] 透過 Gmail API 寄送
- [ ] 更新狀態為 sent
- [ ] 記錄 gmail_message_id

**P5-REPLY-006**: 單元測試（2h）

### 報價方案管理 (16h)

**P5-PACKAGE-001**: 實作 POST /pricing-packages（4h）
- [ ] Handler 實作
- [ ] 建立方案
- [ ] 驗證輸入
- [ ] 測試

**P5-PACKAGE-002**: 實作 GET /pricing-packages（3h）
- [ ] Handler 實作
- [ ] 篩選（is_active）
- [ ] 排序（display_order）
- [ ] 測試

**P5-PACKAGE-003**: 實作 PATCH /pricing-packages/:id（3h）
- [ ] Handler 實作
- [ ] 更新方案
- [ ] 測試

**P5-PACKAGE-004**: 實作 DELETE /pricing-packages/:id（2h）
- [ ] Handler 實作
- [ ] 刪除方案
- [ ] 測試

**P5-PACKAGE-005**: 實作方案排序（選用）（2h）
- [ ] 拖曳排序 API
- [ ] 更新 display_order

**P5-PACKAGE-006**: 整合測試（2h）

### 通知系統 (16h)

**P5-NOTIF-001**: 實作通知 service（4h）
- [ ] 建立 `internal/services/notification/service.go`
- [ ] CreateNotification 函數
- [ ] 定義通知類型
- [ ] 儲存到資料庫

**P5-NOTIF-002**: 實作通知觸發點（6h）
- [ ] 新郵件（合作邀約）
- [ ] 新案件建立
- [ ] 任務即將到期
- [ ] 任務已逾期
- [ ] 在相應流程中呼叫

**P5-NOTIF-003**: 實作通知 API（4h）
- [ ] GET /notifications（列表）
- [ ] PATCH /notifications/:id/read
- [ ] POST /notifications/read-all
- [ ] 測試

**P5-NOTIF-004**: 整合測試（2h）

### 統計儀表板 (12h)

**P5-STATS-001**: 實作 GET /stats/dashboard API（8h）
- [ ] 查詢案件統計（各狀態）
- [ ] 查詢收入統計
- [ ] 查詢郵件統計
- [ ] 查詢任務統計
- [ ] 查詢 Top 品牌
- [ ] 組合結果
- [ ] 測試

**P5-STATS-002**: 實作 GET /stats/revenue API（選用）（4h）
- [ ] 月收入趨勢
- [ ] 案件類型分布
- [ ] 圖表資料格式

### 前端 - 回覆生成 (20h)

**P5-FE-REPLY-001**: 建立回覆生成對話框（8h）
- [ ] UI 設計（Dialog）
- [ ] 語氣選擇（Select）
- [ ] 補充資訊（Textarea）
- [ ] 報價方案選擇（Select）
- [ ] 生成按鈕
- [ ] Loading 狀態

**P5-FE-REPLY-002**: 實作生成邏輯（4h）
- [ ] 呼叫 API
- [ ] 顯示結果
- [ ] 可編輯文字框
- [ ] 重新生成按鈕

**P5-FE-REPLY-003**: 實作寄送功能（4h）
- [ ] 寄送按鈕
- [ ] 確認對話框
- [ ] 呼叫 API
- [ ] 顯示成功/失敗訊息

**P5-FE-REPLY-004**: 實作回覆歷史（選用）（4h）
- [ ] 顯示案件的回覆歷史
- [ ] 查看過往回覆

### 前端 - 報價方案管理 (16h)

**P5-FE-PACKAGE-001**: 建立方案列表頁面（6h）
- [ ] 建立 `pages/pricing-packages/index.vue`
- [ ] 卡片顯示（Vuetify Cards）
- [ ] 顯示方案資訊
- [ ] 啟用/停用開關
- [ ] 編輯/刪除按鈕

**P5-FE-PACKAGE-002**: 建立方案表單（8h）
- [ ] 新增/編輯對話框
- [ ] 表單欄位
- [ ] 項目動態新增（Array input）
- [ ] 表單驗證
- [ ] 提交邏輯

**P5-FE-PACKAGE-003**: 實作預覽功能（選用）（2h）
- [ ] 預覽方案顯示效果
- [ ] 在回覆中的樣式

### 前端 - 通知中心 (16h)

**P5-FE-NOTIF-001**: 建立通知 store（3h）
- [ ] 建立 `stores/notifications.ts`
- [ ] fetchNotifications action
- [ ] markAsRead action
- [ ] unreadCount computed

**P5-FE-NOTIF-002**: 實作 App Bar 通知圖示（5h）
- [ ] 通知 icon（Bell）
- [ ] Badge 顯示未讀數
- [ ] 點擊展開下拉選單
- [ ] 顯示最近 5 則
- [ ] 標記已讀按鈕
- [ ] 查看全部連結

**P5-FE-NOTIF-003**: 建立通知列表頁面（6h）
- [ ] 建立 `pages/notifications.vue`
- [ ] 列表顯示（已讀/未讀區分）
- [ ] 點擊跳轉到相關頁面
- [ ] 全部標記已讀按鈕

**P5-FE-NOTIF-004**: 測試與優化（2h）

### 前端 - 儀表板 (16h)

**P5-FE-DASH-001**: 建立 Dashboard 頁面（12h）
- [ ] 建立 `pages/index.vue`（或 dashboard.vue）
- [ ] 統計卡片（Cards）
  - 案件統計
  - 收入統計
  - 郵件統計
  - 任務統計
- [ ] 收入趨勢圖表（安裝 Chart.js / ApexCharts）
- [ ] 最近案件列表（Table）
- [ ] 即將到期任務（List）
- [ ] Top 品牌（List）
- [ ] 響應式設計

**P5-FE-DASH-002**: 實作圖表（選用）（4h）
- [ ] 收入趨勢折線圖
- [ ] 案件類型圓餅圖
- [ ] 互動效果

### 前端 - 設定頁面 (20h)

**P5-FE-SETTINGS-001**: 建立設定頁面結構（4h）
- [ ] 建立 `pages/settings.vue`
- [ ] Tabs 導覽（Profile, Gmail, AI, Notifications）

**P5-FE-SETTINGS-002**: 實作個人資料 Tab（4h）
- [ ] 顯示當前資料
- [ ] 編輯表單（Name, Avatar）
- [ ] 上傳頭像（選用）
- [ ] 儲存邏輯

**P5-FE-SETTINGS-003**: 實作 Gmail Tab（4h）
- [ ] 顯示連接狀態
- [ ] 連接的 Email
- [ ] 重新授權按鈕
- [ ] 斷開連接按鈕

**P5-FE-SETTINGS-004**: 實作 AI Tab（4h）
- [ ] OpenAI API Key 輸入（password type）
- [ ] 測試連接按鈕
- [ ] AI 回覆語氣偏好（Select）
- [ ] 儲存邏輯

**P5-FE-SETTINGS-005**: 實作 Notifications Tab（4h）
- [ ] 通知偏好設定（Checkboxes）
  - Email 通知
  - 瀏覽器通知
  - 各類型通知開關
- [ ] 時區設定（Select）
- [ ] 儲存邏輯

---

## 🧪 Phase 6: 測試、優化與部署（預估: 40h）

### 後端測試 (16h)

**P6-TEST-BE-001**: 撰寫 Services 單元測試（6h）
- [ ] Gmail service tests
- [ ] OpenAI service tests
- [ ] Auth service tests
- [ ] Mock 外部 API

**P6-TEST-BE-002**: 撰寫 API 端點測試（6h）
- [ ] 測試所有 CRUD 端點
- [ ] 測試認證流程
- [ ] 測試錯誤處理
- [ ] 測試邊界條件

**P6-TEST-BE-003**: 整合測試（4h）
- [ ] 測試完整流程
- [ ] 測試資料庫操作
- [ ] 測試背景任務

### 前端測試 (8h)

**P6-TEST-FE-001**: 元件測試（4h）
- [ ] 測試關鍵元件
- [ ] 使用 Vitest

**P6-TEST-FE-002**: E2E 測試（4h）
- [ ] 安裝 Playwright
- [ ] 測試登入流程
- [ ] 測試建立案件
- [ ] 測試生成回覆

### 效能優化 (12h)

**P6-OPT-001**: 資料庫優化（6h）
- [ ] 檢查 N+1 問題
- [ ] 新增必要索引
- [ ] 使用 EXPLAIN ANALYZE
- [ ] 查詢優化

**P6-OPT-002**: API 效能優化（4h）
- [ ] 實作 Redis 快取
- [ ] 資料庫連線池調校
- [ ] 壓縮回應（gzip）

**P6-OPT-003**: 前端效能優化（2h）
- [ ] 圖片優化（WebP）
- [ ] Lazy loading
- [ ] Code splitting

### 安全性與文件 (8h)

**P6-SEC-001**: 安全性檢查（4h）
- [ ] OWASP Top 10 檢查
- [ ] Rate limiting 設定
- [ ] HTTPS 設定
- [ ] 加密驗證

**P6-DOC-001**: 撰寫文件（4h）
- [ ] API 文件（Swagger）
- [ ] 使用者手冊
- [ ] 開發者指南

### 部署 (12h)

**P6-DEPLOY-001**: 建立 Dockerfile（3h）
- [ ] 後端 Dockerfile
- [ ] 前端 Dockerfile
- [ ] Multi-stage build

**P6-DEPLOY-002**: 設定 CI/CD（4h）
- [ ] GitHub Actions workflow
- [ ] 自動測試
- [ ] 自動部署

**P6-DEPLOY-003**: 部署到生產環境（5h）
- [ ] 選擇平台（Render / Railway / Fly.io）
- [ ] 設定環境變數
- [ ] 設定資料庫
- [ ] 設定 Redis
- [ ] 啟動 worker
- [ ] 驗證功能

---

## ✅ 總結

### 時程統計
- **Phase 0**: 40 hours
- **Phase 1**: 80 hours
- **Phase 2**: 80 hours
- **Phase 3**: 80 hours
- **Phase 4**: 80 hours
- **Phase 5**: 80 hours
- **Phase 6**: 40 hours
- **總計**: **480 hours**（約 12 週）

### 關鍵里程碑
- ✅ **Week 1**: 專案初始化完成
- ✅ **Week 3**: 可以用 Google 登入
- ✅ **Week 5**: Gmail 郵件自動同步
- ✅ **Week 7**: AI 自動分析並建立案件
- ✅ **Week 9**: 案件管理功能完整
- ✅ **Week 11**: AI 回覆生成與進階功能
- ✅ **Week 12**: 測試完成，上線！

### 建議
- 每週進行一次 sprint review
- 遇到阻礙時及時調整計劃
- 優先完成核心功能，進階功能可後續迭代
- 保持程式碼品質，避免技術債累積

---

**任務分解完成！準備開始實作！** 🎉

