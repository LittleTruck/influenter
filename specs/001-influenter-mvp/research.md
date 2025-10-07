# Influenter MVP - 技術研究文件

> **目的**：研究技術選型、最佳實踐與實作細節  
> **最後更新**：2025-10-07

---

## 🎯 研究目標

基於需求澄清的結果，需要研究以下技術領域：

1. **Go Web 框架**選擇（Gin vs Fiber vs Echo）
2. **Gmail API** 整合最佳實踐
3. **OpenAI API** 結構化輸出實作
4. **Nuxt 3 + Vuetify 3** 整合方式
5. **背景任務處理**方案（定期同步郵件）
6. **資料加密**方案（API Key 儲存）
7. **PWA** 實作細節

---

## 1️⃣ Go Web 框架選擇

### 候選框架對比

#### Option 1: Gin 🏆 **推薦**
**優勢**：
- ⭐ **成熟穩定**：GitHub 75k+ stars，社群活躍
- ⚡ **效能優異**：基於 httprouter，路由效能極佳
- 📚 **文件完整**：中文文件豐富，學習曲線平緩
- 🔧 **中介軟體豐富**：CORS、JWT、限流等現成方案
- 🎯 **輕量但完整**：功能足夠，不過度複雜

**劣勢**：
- 相較 Fiber，效能略遜一籌（但差距不大）
- API 設計較傳統（非 Express-like）

**適用場景**：
- 需要穩定性與社群支援
- 團隊對 Go 標準庫熟悉
- 長期維護的專案

**範例程式碼**：
```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.Default()
    
    // 中介軟體
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    
    // 路由群組
    api := r.Group("/api/v1")
    {
        api.GET("/cases", getCases)
        api.POST("/cases", createCase)
        api.GET("/cases/:id", getCaseByID)
    }
    
    r.Run(":8080")
}

func getCases(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "data": []string{"case1", "case2"},
    })
}
```

#### Option 2: Fiber
**優勢**：
- ⚡ **效能最佳**：基於 Fasthttp，benchmark 領先
- 💡 **Express-like API**：前端工程師容易上手
- 📦 **功能完整**：內建許多實用功能

**劣勢**：
- ❌ **不基於 net/http**：不相容標準庫生態系
- ⚠️ **社群較小**：相較 Gin 資源較少
- 🔄 **版本迭代快**：Breaking changes 較多

**適用場景**：
- 極度追求效能
- 團隊熟悉 Express.js
- 不依賴 net/http 生態系

#### Option 3: Echo
**優勢**：
- 🎯 **平衡效能與易用性**
- 📝 **文件清晰**
- 🔧 **中介軟體豐富**

**劣勢**：
- 社群規模介於 Gin 和 Fiber 之間
- 沒有特別突出的優勢

### ✅ 最終決策：**Gin**

**理由**：
1. **成熟度**：經過大量生產環境驗證
2. **社群支援**：遇到問題容易找到解決方案
3. **穩定性**：API 穩定，不常有 breaking changes
4. **效能足夠**：對於 MVP 階段，Gin 的效能完全足夠
5. **學習資源**：中文教學豐富，降低學習成本

**效能基準測試**（參考 TechEmpower Benchmark）：
- Gin: ~600k req/sec
- Fiber: ~800k req/sec
- Echo: ~550k req/sec

> 對於 Influenter 的使用情境，Gin 的效能綽綽有餘（預估 QPS < 1000）

---

## 2️⃣ Gmail API 整合最佳實踐

### 認證流程

#### Google OAuth 2.0 流程
```
1. 前端觸發「使用 Google 登入」
   ↓
2. 前端重定向到 Google OAuth 授權頁面
   Scopes: 
   - https://www.googleapis.com/auth/gmail.readonly
   - https://www.googleapis.com/auth/gmail.send
   ↓
3. 使用者授權後，Google 回調 callback URL
   帶上 authorization code
   ↓
4. 後端用 code 交換 access_token 與 refresh_token
   ↓
5. 儲存 tokens（加密）到資料庫
   ↓
6. 使用 access_token 呼叫 Gmail API
```

#### Go 實作範例
```go
import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/gmail/v1"
)

// OAuth2 設定
var googleOAuthConfig = &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  "http://localhost:8080/auth/google/callback",
    Scopes: []string{
        gmail.GmailReadonlyScope,
        gmail.GmailSendScope,
    },
    Endpoint: google.Endpoint,
}

// 處理登入
func handleGoogleLogin(c *gin.Context) {
    url := googleOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

// 處理回調
func handleGoogleCallback(c *gin.Context) {
    code := c.Query("code")
    token, err := googleOAuthConfig.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 儲存 token（加密）
    saveEncryptedToken(token)
}
```

### 郵件同步策略

#### 增量同步實作
```go
type GmailSyncer struct {
    service *gmail.Service
    userID  string
}

// 首次同步：抓取最近 30 天
func (s *GmailSyncer) InitialSync() error {
    query := "after:" + time.Now().AddDate(0, 0, -30).Format("2006/01/02")
    
    return s.syncMessages(query)
}

// 增量同步：只抓取新郵件
func (s *GmailSyncer) IncrementalSync(lastSyncTime time.Time) error {
    query := "after:" + lastSyncTime.Format("2006/01/02")
    
    return s.syncMessages(query)
}

func (s *GmailSyncer) syncMessages(query string) error {
    req := s.service.Users.Messages.List(s.userID).Q(query).MaxResults(100)
    
    for {
        res, err := req.Do()
        if err != nil {
            return err
        }
        
        for _, msg := range res.Messages {
            // 獲取完整郵件內容
            fullMsg, err := s.service.Users.Messages.Get(s.userID, msg.Id).
                Format("full").Do()
            if err != nil {
                log.Printf("Error getting message %s: %v", msg.Id, err)
                continue
            }
            
            // 解析並儲存郵件
            s.parseAndSaveMessage(fullMsg)
        }
        
        if res.NextPageToken == "" {
            break
        }
        req.PageToken(res.NextPageToken)
    }
    
    return nil
}
```

### API 配額管理

**Gmail API 配額限制**：
- 每日配額：1,000,000,000 單位
- `users.messages.list`：5 單位/次
- `users.messages.get`：5 單位/次
- `users.messages.send`：100 單位/次

**最佳實踐**：
1. **批次請求**：使用 batch API 減少請求次數
2. **Exponential Backoff**：遇到限流時自動重試
3. **快取**：快取不常變動的資料
4. **監控**：記錄 API 使用量

```go
// Exponential backoff 重試
func retryWithBackoff(fn func() error) error {
    backoff := time.Second
    maxRetries := 5
    
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if isRateLimitError(err) {
            time.Sleep(backoff)
            backoff *= 2
            continue
        }
        
        return err
    }
    
    return fmt.Errorf("max retries exceeded")
}
```

---

## 3️⃣ OpenAI API 結構化輸出

### GPT-4 結構化輸出（Function Calling）

OpenAI 提供 **Function Calling** 功能，可確保輸出符合 JSON schema。

#### 郵件分類範例
```go
type EmailClassification struct {
    Category    string   `json:"category"`     // 合作邀約、合約討論等
    Brand       string   `json:"brand"`        // 品牌名稱
    Amount      *float64 `json:"amount"`       // 報價金額
    Dates       []string `json:"dates"`        // 重要日期
    ContactName string   `json:"contact_name"` // 聯絡人
    Confidence  float64  `json:"confidence"`   // 信心指標 0-1
}

// OpenAI API 請求
func classifyEmail(emailContent string) (*EmailClassification, error) {
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
    
    resp, err := client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: openai.GPT4oMini,  // 較便宜的模型
            Messages: []openai.ChatCompletionMessage{
                {
                    Role:    openai.ChatMessageRoleSystem,
                    Content: systemPrompt,
                },
                {
                    Role:    openai.ChatMessageRoleUser,
                    Content: emailContent,
                },
            },
            Functions: []openai.FunctionDefinition{
                {
                    Name:        "classify_email",
                    Description: "分類並抽取郵件資訊",
                    Parameters: jsonschema.Definition{
                        Type: jsonschema.Object,
                        Properties: map[string]jsonschema.Definition{
                            "category": {
                                Type: jsonschema.String,
                                Enum: []string{
                                    "合作邀約", "合約討論", "執行通知",
                                    "交稿相關", "結案感謝", "其他",
                                },
                            },
                            "brand": {
                                Type:        jsonschema.String,
                                Description: "品牌或公司名稱",
                            },
                            "amount": {
                                Type:        jsonschema.Number,
                                Description: "提及的金額（新台幣）",
                            },
                            // ... 其他欄位
                        },
                        Required: []string{"category", "confidence"},
                    },
                },
            },
            FunctionCall: "auto",
        },
    )
    
    if err != nil {
        return nil, err
    }
    
    // 解析 function call 結果
    var result EmailClassification
    json.Unmarshal(
        []byte(resp.Choices[0].Message.FunctionCall.Arguments),
        &result,
    )
    
    return &result, nil
}
```

### System Prompt 設計

針對台灣創作者情境的 prompt：

```text
你是一個專為台灣創作者設計的郵件分析助手。你的任務是分析品牌合作相關的郵件，並抽取關鍵資訊。

## 分類標準
- **合作邀約**：品牌主動詢問合作意願、介紹產品、詢問檔期
- **合約討論**：討論報價、合約條款、合作細節
- **執行通知**：確認合作開始、提供素材、確認腳本
- **交稿相關**：提交初稿、修改意見、正式交稿
- **結案感謝**：感謝合作、確認款項、索取發票
- **其他**：不屬於以上分類

## 資訊抽取注意事項
- 品牌名稱：找出寄件者代表的品牌或公司
- 金額：只抽取明確提及的新台幣金額（不要猜測）
- 日期：抽取截止日期、交稿日、上線日等重要日期
- 聯絡人：找出品牌窗口的姓名與職稱

## 信心指標
- 1.0：非常確定
- 0.8：相當確定
- 0.6：有些不確定
- 0.4：很不確定
- 0.2：幾乎是猜測

請根據郵件內容的清晰程度給出合理的信心指標。
```

### 成本預估

**GPT-4o-mini 定價**（2024 價格）：
- Input: $0.150 / 1M tokens
- Output: $0.600 / 1M tokens

**單封郵件分析成本**：
- 假設 system prompt: 300 tokens
- 平均郵件: 500 tokens
- 輸出: 200 tokens
- 總計：1000 tokens ≈ **$0.0003 USD**（約 NT$ 0.01）

**月成本預估**（單一使用者）：
- 假設每月同步 200 封新郵件
- 200 × $0.0003 = **$0.06 USD**（約 NT$ 2）

非常便宜！✅

---

## 4️⃣ Nuxt 3 + Vuetify 3 整合

### 專案設置

```bash
# 建立 Nuxt 3 專案
npx nuxi@latest init influenter-frontend
cd influenter-frontend

# 安裝 Vuetify 3
npm install vuetify @mdi/font

# 安裝 Nuxt UI（可選）
npm install @nuxt/ui
```

### Vuetify 3 設定

```typescript
// plugins/vuetify.ts
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { mdi } from 'vuetify/iconsets/mdi'
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

export default defineNuxtPlugin((nuxtApp) => {
  const vuetify = createVuetify({
    components,
    directives,
    icons: {
      defaultSet: 'mdi',
      sets: { mdi },
    },
    theme: {
      defaultTheme: 'light',
      themes: {
        light: {
          colors: {
            primary: '#1976D2',
            secondary: '#424242',
            accent: '#82B1FF',
            error: '#FF5252',
            info: '#2196F3',
            success: '#4CAF50',
            warning: '#FB8C00',
          },
        },
        dark: {
          colors: {
            primary: '#2196F3',
            secondary: '#616161',
            // ... 其他顏色
          },
        },
      },
    },
  })

  nuxtApp.vueApp.use(vuetify)
})
```

### 狀態管理（Pinia）

```typescript
// stores/auth.ts
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    gmailConnected: false,
  }),
  
  getters: {
    isAuthenticated: (state) => !!state.user,
  },
  
  actions: {
    async loginWithGoogle() {
      // 觸發 OAuth 流程
      window.location.href = 'http://localhost:8080/auth/google/login'
    },
    
    async fetchUserProfile() {
      const data = await $fetch('/api/v1/user/profile')
      this.user = data
    },
  },
})
```

### API 層設計

```typescript
// composables/useAPI.ts
export const useAPI = () => {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || 'http://localhost:8080/api/v1'
  
  const apiFetch = $fetch.create({
    baseURL,
    onRequest({ options }) {
      // 自動加上 JWT token
      const token = useCookie('auth_token')
      if (token.value) {
        options.headers = {
          ...options.headers,
          Authorization: `Bearer ${token.value}`,
        }
      }
    },
    onResponseError({ response }) {
      // 統一錯誤處理
      if (response.status === 401) {
        // 登出並重定向
        navigateTo('/login')
      }
    },
  })
  
  return { apiFetch }
}

// 使用範例
const { apiFetch } = useAPI()
const cases = await apiFetch('/cases')
```

---

## 5️⃣ 背景任務處理

### 方案選擇

#### Option 1: Asynq 🏆 **推薦**
**特色**：
- Redis-backed 的分散式任務佇列
- 支援定時任務（cron）
- 支援重試與錯誤處理
- Web UI 可視化監控

**安裝**：
```bash
go get github.com/hibiken/asynq
```

**實作範例**：
```go
// 定義任務
type EmailSyncTask struct {
    UserID string
}

func NewEmailSyncTask(userID string) (*asynq.Task, error) {
    payload, err := json.Marshal(EmailSyncTask{UserID: userID})
    if err != nil {
        return nil, err
    }
    return asynq.NewTask("email:sync", payload), nil
}

// 處理任務
func HandleEmailSyncTask(ctx context.Context, t *asynq.Task) error {
    var p EmailSyncTask
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }
    
    log.Printf("Syncing emails for user %s", p.UserID)
    // 執行同步邏輯
    return syncEmailsForUser(p.UserID)
}

// 啟動 worker
func main() {
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: "localhost:6379"},
        asynq.Config{Concurrency: 10},
    )
    
    mux := asynq.NewServeMux()
    mux.HandleFunc("email:sync", HandleEmailSyncTask)
    
    srv.Run(mux)
}

// 定時任務
scheduler := asynq.NewScheduler(
    asynq.RedisClientOpt{Addr: "localhost:6379"},
    nil,
)

// 每 5 分鐘執行
scheduler.Register("*/5 * * * *", NewEmailSyncTask("all"))
```

#### Option 2: 自建 Goroutine + Ticker
**適用場景**：不想引入 Redis 依賴

```go
func startEmailSyncWorker() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            syncAllUsersEmails()
        }
    }
}

func syncAllUsersEmails() {
    users := getAllUsers()
    
    for _, user := range users {
        go func(u User) {
            if err := syncEmailsForUser(u.ID); err != nil {
                log.Printf("Error syncing for user %s: %v", u.ID, err)
            }
        }(user)
    }
}
```

### ✅ 最終決策：**Asynq**（需要 Redis）

**理由**：
1. **可靠性**：任務持久化，重啟不丟失
2. **可擴展**：支援多個 worker 分散處理
3. **監控**：內建 Web UI
4. **重試機制**：自動處理失敗任務

---

## 6️⃣ 資料加密方案

### AES-256 加密 API Key

```go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

// 加密
func Encrypt(plaintext string, key []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密
func Decrypt(ciphertext string, key []byte) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonceSize := gcm.NonceSize()
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}
```

**密鑰管理**：
- 使用環境變數 `ENCRYPTION_KEY`（32 bytes）
- 不要 hard-code 在程式碼中
- 生產環境使用 AWS KMS / GCP KMS

---

## 7️⃣ PWA 實作

### Nuxt 3 PWA 模組

```bash
npm install @vite-pwa/nuxt
```

```typescript
// nuxt.config.ts
export default defineNuxtConfig({
  modules: ['@vite-pwa/nuxt'],
  
  pwa: {
    manifest: {
      name: 'Influenter',
      short_name: 'Influenter',
      description: 'AI 驅動的創作者合作案件管理系統',
      theme_color: '#1976D2',
      icons: [
        {
          src: '/icon-192.png',
          sizes: '192x192',
          type: 'image/png',
        },
        {
          src: '/icon-512.png',
          sizes: '512x512',
          type: 'image/png',
        },
      ],
    },
    workbox: {
      navigateFallback: '/',
      globPatterns: ['**/*.{js,css,html,png,svg,ico}'],
    },
  },
})
```

---

## ✅ 技術研究總結

### 已確認技術選型

| 領域 | 選擇 | 理由 |
|------|------|------|
| **後端框架** | Gin | 成熟、社群大、穩定 |
| **ORM** | GORM | Go 生態系最流行 |
| **背景任務** | Asynq + Redis | 可靠、可擴展、有監控 |
| **前端框架** | Nuxt 3 | SSR、SEO 友善 |
| **UI 框架** | Vuetify 3 | 元件完整、RWD 優秀 |
| **狀態管理** | Pinia | Nuxt 3 官方推薦 |
| **資料庫** | PostgreSQL 15+ | 關聯式資料、JSONB 支援 |
| **加密** | AES-256-GCM | 業界標準 |
| **AI 模型** | GPT-4o-mini | 成本低、速度快 |

### 核心依賴套件

**後端（Go）**：
```go
require (
    github.com/gin-gonic/gin v1.9.1
    github.com/hibiken/asynq v0.24.1
    golang.org/x/oauth2 v0.15.0
    google.golang.org/api v0.153.0
    github.com/sashabaranov/go-openai v1.17.9
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    github.com/golang-jwt/jwt/v5 v5.2.0
)
```

**前端（Node.js）**：
```json
{
  "dependencies": {
    "nuxt": "^3.8.2",
    "vue": "^3.3.11",
    "vuetify": "^3.4.9",
    "@mdi/font": "^7.3.67",
    "pinia": "^2.1.7",
    "@vite-pwa/nuxt": "^0.4.0"
  }
}
```

---

**下一步**：生成詳細的實作計劃（`plan.md`）與資料模型設計（`data-model.md`）

