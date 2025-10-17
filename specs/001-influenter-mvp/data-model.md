# Influenter MVP - 資料模型設計

> **目的**：定義資料庫 schema、關聯關係與索引策略  
> **資料庫**：PostgreSQL 15+  
> **ORM**：GORM  
> **最後更新**：2025-10-07

---

## 🗂️ 實體關聯圖（ERD）

```
┌─────────────────┐
│     users       │
└────────┬────────┘
         │
         │ 1:N
         ▼
┌─────────────────┐         ┌──────────────────┐
│ oauth_accounts  │────┬───▶│     emails       │
└─────────────────┘    │    └────────┬─────────┘
                       │             │
                       │             │ M:N
                       │             ▼
                       │    ┌──────────────────┐
                       └───▶│      cases       │
                            └────────┬─────────┘
                                     │
                         ┌───────────┼───────────┐
                         │           │           │
                         ▼           ▼           ▼
                ┌─────────────┐ ┌────────┐ ┌─────────────┐
                │    tasks    │ │replies │ │case_updates │
                └─────────────┘ └────────┘ └─────────────┘

┌──────────────────┐
│ pricing_packages │ (獨立表，供使用者管理報價方案)
└──────────────────┘

┌──────────────────┐
│  ai_analysis     │ (記錄 AI 分析結果與修正)
└──────────────────┘
```

---

## 📊 資料表設計

### 1. users（使用者）

**用途**：儲存使用者基本資料與認證資訊

```sql
CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email               VARCHAR(255) NOT NULL UNIQUE,
    name                VARCHAR(100),
    profile_picture_url TEXT,
    
    -- 設定
    ai_reply_tone       VARCHAR(50) DEFAULT 'professional',  -- 回覆語氣
    timezone            VARCHAR(50) DEFAULT 'Asia/Taipei',
    notification_prefs  JSONB,  -- 通知偏好設定
    
    -- 系統欄位
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at       TIMESTAMP WITH TIME ZONE,
    deleted_at          TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
```

**設計考量**：
- OAuth 相關資訊（Google ID、tokens 等）統一存放在 `oauth_accounts` 表
- 系統預設使用 Google 登入，但架構支援其他 OAuth 提供商
- 使用者可以連結多個第三方帳號（如同時連結 Google 和 Outlook）

**範例資料**：
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "name": "Alice Chen",
  "profile_picture_url": "https://...",
  "ai_reply_tone": "friendly",
  "notification_prefs": {
    "email_on_new_case": true,
    "email_on_deadline": true,
    "browser_notifications": true
  }
}
```

---

### 2. oauth_accounts（第三方 OAuth 帳號）

**用途**：儲存使用者連結的第三方 OAuth 帳號（如 Gmail、Outlook 等）的 OAuth tokens（加密）

```sql
CREATE TABLE oauth_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- OAuth 提供商資訊
    provider        VARCHAR(50) NOT NULL,           -- google, outlook, apple
    provider_id     VARCHAR(255),                   -- 提供商的使用者 ID
    email           VARCHAR(255) NOT NULL,
    
    -- OAuth tokens（加密儲存 - AES-256-GCM）
    access_token    TEXT NOT NULL,                  -- 加密的 access token
    refresh_token   TEXT NOT NULL,                  -- 加密的 refresh token
    token_expiry    TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- 同步狀態（主要用於郵件同步）
    last_sync_at    TIMESTAMP WITH TIME ZONE,
    last_history_id VARCHAR(100),                   -- Gmail API history ID 或其他提供商的同步 ID
    sync_status     VARCHAR(50) DEFAULT 'active',   -- active, paused, error
    sync_error      TEXT,
    
    -- 額外資訊（JSON 格式，可存放提供商特定資訊）
    metadata        JSONB,
    
    -- 系統欄位
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP WITH TIME ZONE,
    
    -- 唯一約束：一個使用者不能重複連結同一個 provider 的同一個 email
    CONSTRAINT unique_user_provider_email UNIQUE (user_id, provider, email)
);

CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);
CREATE INDEX idx_oauth_accounts_provider ON oauth_accounts(provider);
CREATE INDEX idx_oauth_accounts_sync_status ON oauth_accounts(sync_status);
CREATE INDEX idx_oauth_accounts_deleted_at ON oauth_accounts(deleted_at);
CREATE INDEX idx_oauth_accounts_token_expiry ON oauth_accounts(token_expiry);
```

**重要**：`access_token` 和 `refresh_token` 在儲存前必須使用 AES-256-GCM 加密。

**設計考量**：
- 使用通用的 `provider` 欄位支援多個 OAuth 提供商（Google、Outlook、Apple 等）
- `metadata` 欄位可儲存提供商特定的資訊（如 Gmail 的 history ID、Outlook 的 delta link 等）
- 支援軟刪除（deleted_at）
- 一個使用者可以連結多個不同提供商的帳號，但同一個 provider 的同一個 email 只能連結一次

---

### 3. emails（郵件）

**用途**：儲存從第三方 OAuth 帳號（如 Gmail、Outlook 等）同步的郵件

```sql
CREATE TABLE emails (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    oauth_account_id    UUID NOT NULL REFERENCES oauth_accounts(id) ON DELETE CASCADE,
    
    -- 郵件提供商原始資訊
    provider_message_id VARCHAR(255) NOT NULL UNIQUE,  -- Gmail message ID 或其他提供商的 message ID
    thread_id           VARCHAR(255),                  -- 郵件串 ID
    
    -- 郵件基本資訊
    from_email          VARCHAR(255) NOT NULL,
    from_name           VARCHAR(255),
    to_email            VARCHAR(255),
    subject             TEXT,
    body_text           TEXT,                          -- 純文字內容
    body_html           TEXT,                          -- HTML 內容
    snippet             TEXT,                          -- 郵件摘要（前 150 字）
    
    -- 郵件屬性
    received_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    is_read             BOOLEAN DEFAULT FALSE,
    has_attachments     BOOLEAN DEFAULT FALSE,
    labels              TEXT[],                        -- 標籤（Gmail labels 或其他提供商標籤）
    
    -- AI 分析狀態
    ai_analyzed         BOOLEAN DEFAULT FALSE,
    ai_analysis_id      UUID REFERENCES ai_analysis(id),
    
    -- 案件關聯
    case_id             UUID REFERENCES cases(id) ON DELETE SET NULL,
    
    -- 系統欄位
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_emails_oauth_account_id ON emails(oauth_account_id);
CREATE INDEX idx_emails_provider_message_id ON emails(provider_message_id);
CREATE INDEX idx_emails_thread_id ON emails(thread_id);
CREATE INDEX idx_emails_from_email ON emails(from_email);
CREATE INDEX idx_emails_to_email ON emails(to_email);
CREATE INDEX idx_emails_received_at ON emails(received_at DESC);
CREATE INDEX idx_emails_case_id ON emails(case_id);
CREATE INDEX idx_emails_deleted_at ON emails(deleted_at);

-- 部分索引：僅針對未分析的郵件（效能優化）
CREATE INDEX idx_emails_ai_analyzed ON emails(ai_analyzed) WHERE ai_analyzed = FALSE;

-- GIN 索引：用於主旨全文搜尋
CREATE INDEX idx_emails_subject_gin ON emails USING GIN (to_tsvector('english', COALESCE(subject, '')));
```

**設計考量**：
- 只儲存郵件基本資訊，不下載附件實體檔案
- `labels` 使用 PostgreSQL array 型別，方便查詢
- `ai_analyzed` 部分索引僅針對未分析的郵件，提升背景任務效率
- 使用 `provider_message_id` 而非 `gmail_message_id`，支援多種郵件提供商
- 支援軟刪除（deleted_at）
- GIN 索引用於主旨的全文搜尋功能

---

### 4. ai_analysis（AI 分析結果）

**用途**：記錄 AI 對郵件的分析結果

```sql
CREATE TABLE ai_analysis (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_id         UUID NOT NULL REFERENCES emails(id) ON DELETE CASCADE,
    
    -- AI 分析結果
    category         VARCHAR(50) NOT NULL,  -- 合作邀約、合約討論等
    brand_name       VARCHAR(255),
    amount           NUMERIC(12, 2),        -- 報價金額
    currency         VARCHAR(10) DEFAULT 'TWD',
    important_dates  JSONB,                 -- [{"date": "2024-03-15", "type": "deadline"}]
    contact_name     VARCHAR(100),
    contact_title    VARCHAR(100),
    collaboration_type VARCHAR(100),        -- 業配影片、開箱文等
    
    -- AI 信心指標
    confidence       FLOAT NOT NULL,        -- 0.0 - 1.0
    
    -- 使用者修正
    user_corrected   BOOLEAN DEFAULT FALSE,
    original_data    JSONB,                 -- 儲存 AI 原始判斷（供訓練使用）
    corrected_data   JSONB,                 -- 使用者修正後的資料
    
    -- OpenAI API 資訊
    model_used       VARCHAR(50),           -- gpt-4o-mini
    tokens_used      INTEGER,
    api_cost         NUMERIC(10, 6),        -- USD
    
    -- 系統欄位
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_analysis_email_id ON ai_analysis(email_id);
CREATE INDEX idx_ai_analysis_category ON ai_analysis(category);
CREATE INDEX idx_ai_analysis_user_corrected ON ai_analysis(user_corrected);
CREATE INDEX idx_ai_analysis_confidence ON ai_analysis(confidence);
```

**important_dates JSON 範例**：
```json
[
  {
    "date": "2024-03-20",
    "type": "deadline",
    "description": "腳本提交截止日"
  },
  {
    "date": "2024-03-25",
    "type": "delivery",
    "description": "影片上線日"
  }
]
```

---

### 5. cases（案件）

**用途**：合作案件的核心資料

```sql
CREATE TABLE cases (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 案件基本資訊
    title               VARCHAR(255) NOT NULL,
    brand_name          VARCHAR(255) NOT NULL,
    collaboration_type  VARCHAR(100),      -- 業配影片、貼文、Reels 等
    description         TEXT,
    
    -- 財務資訊
    quoted_amount       NUMERIC(12, 2),    -- 報價金額
    final_amount        NUMERIC(12, 2),    -- 最終成交金額
    currency            VARCHAR(10) DEFAULT 'TWD',
    payment_status      VARCHAR(50) DEFAULT 'pending',  -- pending, partial, completed
    
    -- 案件狀態
    status              VARCHAR(50) NOT NULL DEFAULT 'to_confirm',
    -- to_confirm: 待確認
    -- in_progress: 進行中
    -- completed: 已完成
    -- cancelled: 已取消
    
    -- 重要日期
    contract_date       DATE,
    deadline_date       DATE,
    delivery_date       DATE,
    publish_date        DATE,
    
    -- 聯絡資訊
    contact_name        VARCHAR(100),
    contact_email       VARCHAR(255),
    contact_phone       VARCHAR(50),
    
    -- 備註與附加資訊
    notes               TEXT,
    tags                TEXT[],            -- 自訂標籤
    
    -- 來源
    source              VARCHAR(50) DEFAULT 'email',  -- email, manual
    created_from_email_id UUID REFERENCES emails(id) ON DELETE SET NULL,
    
    -- 系統欄位
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMP          -- 軟刪除
);

CREATE INDEX idx_cases_user_id ON cases(user_id);
CREATE INDEX idx_cases_status ON cases(status);
CREATE INDEX idx_cases_brand_name ON cases(brand_name);
CREATE INDEX idx_cases_deadline_date ON cases(deadline_date);
CREATE INDEX idx_cases_deleted_at ON cases(deleted_at);
CREATE INDEX idx_cases_created_at ON cases(created_at DESC);
```

**status 流程**：
```
to_confirm → in_progress → completed
    ↓              ↓
cancelled      cancelled
```

---

### 6. case_emails（案件與郵件關聯）

**用途**：多對多關聯表（一個案件可能有多封郵件往來）

```sql
CREATE TABLE case_emails (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id     UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    email_id    UUID NOT NULL REFERENCES emails(id) ON DELETE CASCADE,
    
    -- 郵件在案件中的角色
    email_type  VARCHAR(50),  -- initial_inquiry, negotiation, contract, delivery, completion
    
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_case_email UNIQUE (case_id, email_id)
);

CREATE INDEX idx_case_emails_case_id ON case_emails(case_id);
CREATE INDEX idx_case_emails_email_id ON case_emails(email_id);
```

---

### 7. tasks（任務）

**用途**：案件相關的任務與時程

```sql
CREATE TABLE tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id         UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    
    -- 任務資訊
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    
    -- 時程
    due_date        DATE,
    due_time        TIME,
    
    -- 狀態
    status          VARCHAR(50) DEFAULT 'pending',  -- pending, completed, cancelled
    completed_at    TIMESTAMP,
    
    -- 提醒
    reminder_sent   BOOLEAN DEFAULT FALSE,
    reminder_days   INTEGER DEFAULT 1,  -- 提前幾天提醒
    
    -- 來源
    source          VARCHAR(50) DEFAULT 'manual',  -- manual, auto_generated
    
    -- 系統欄位
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_case_id ON tasks(case_id);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_reminder_pending ON tasks(due_date, reminder_sent) 
    WHERE status = 'pending' AND reminder_sent = FALSE;
```

---

### 8. replies（回覆）

**用途**：AI 生成與使用者編輯的回覆記錄

```sql
CREATE TABLE replies (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id                 UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    email_id                UUID REFERENCES emails(id) ON DELETE SET NULL,  -- 回覆哪封信
    
    -- 回覆內容
    ai_generated_content    TEXT,       -- AI 生成的原始內容
    user_final_content      TEXT,       -- 使用者最終版本
    
    -- 回覆設定
    reply_tone              VARCHAR(50),  -- professional, friendly, concise
    additional_context      TEXT,         -- 使用者補充的資訊
    
    -- 寄送狀態
    status                  VARCHAR(50) DEFAULT 'draft',  -- draft, sent, failed
    sent_at                 TIMESTAMP,
    gmail_message_id        VARCHAR(255), -- 寄出後的 Gmail message ID
    
    -- AI 資訊
    model_used              VARCHAR(50),
    tokens_used             INTEGER,
    api_cost                NUMERIC(10, 6),
    
    -- 系統欄位
    created_at              TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_replies_case_id ON replies(case_id);
CREATE INDEX idx_replies_email_id ON replies(email_id);
CREATE INDEX idx_replies_status ON replies(status);
```

---

### 9. pricing_packages（報價方案）

**用途**：使用者自訂的合作方案模板

```sql
CREATE TABLE pricing_packages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 方案資訊
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    price           NUMERIC(12, 2) NOT NULL,
    currency        VARCHAR(10) DEFAULT 'TWD',
    
    -- 包含項目
    items           JSONB,  -- [{"name": "60秒業配", "description": "..."}]
    
    -- 注意事項
    terms           TEXT,
    notes           TEXT,
    
    -- 預設時程（天數）
    default_duration_days INTEGER,
    
    -- 排序
    display_order   INTEGER DEFAULT 0,
    
    -- 啟用狀態
    is_active       BOOLEAN DEFAULT TRUE,
    
    -- 系統欄位
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pricing_packages_user_id ON pricing_packages(user_id);
CREATE INDEX idx_pricing_packages_active ON pricing_packages(is_active, display_order);
```

**items JSON 範例**：
```json
[
  {
    "name": "60-90 秒業配段落",
    "description": "在影片中置入 60-90 秒的產品介紹"
  },
  {
    "name": "影片說明欄連結",
    "description": "於說明欄放置品牌連結與折扣碼"
  },
  {
    "name": "保留影片 30 天",
    "description": "影片至少保留 30 天不下架"
  }
]
```

---

### 10. case_updates（案件更新記錄）

**用途**：追蹤案件的狀態變更歷史

```sql
CREATE TABLE case_updates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id     UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 變更類型
    update_type VARCHAR(50) NOT NULL,  -- status_change, amount_change, note_added
    
    -- 變更內容
    old_value   TEXT,
    new_value   TEXT,
    comment     TEXT,
    
    -- 系統欄位
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_case_updates_case_id ON case_updates(case_id, created_at DESC);
```

---

### 11. notifications（通知）

**用途**：站內通知中心

```sql
CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 通知內容
    type            VARCHAR(50) NOT NULL,  -- new_case, deadline, reply_needed
    title           VARCHAR(255) NOT NULL,
    message         TEXT,
    
    -- 關聯
    related_case_id UUID REFERENCES cases(id) ON DELETE SET NULL,
    related_email_id UUID REFERENCES emails(id) ON DELETE SET NULL,
    
    -- 狀態
    is_read         BOOLEAN DEFAULT FALSE,
    read_at         TIMESTAMP,
    
    -- 連結
    action_url      TEXT,
    
    -- 系統欄位
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) 
    WHERE is_read = FALSE;
```

---

## 🔑 索引策略總結

### 主要索引

1. **主鍵索引**：所有表使用 UUID 作為主鍵
2. **外鍵索引**：所有外鍵欄位都建立索引
3. **查詢索引**：依據常見查詢模式建立複合索引
4. **部分索引**：針對特定條件（如未讀通知）建立部分索引

### 效能考量

- **emails 表**：`received_at DESC` 索引支援「最新郵件」查詢
- **cases 表**：`deadline_date` 索引支援「即將到期」查詢
- **tasks 表**：複合索引支援「待提醒任務」查詢
- **notifications 表**：部分索引僅針對未讀通知

---

## 📈 資料增長預估

### 單一使用者（每月）
- **郵件**：~200 封（同步範圍內）
- **案件**：~10 個
- **任務**：~30 個
- **回覆**：~15 個
- **通知**：~50 則

### 1000 使用者（每月）
- **郵件**：~200,000 封
- **案件**：~10,000 個
- **任務**：~30,000 個

### 儲存空間預估（1000 使用者，1 年）
- **郵件（純文字）**：~2.4M × 5KB ≈ **12 GB**
- **其他資料**：≈ **3 GB**
- **總計**：~**15 GB**

PostgreSQL 完全可以應對 ✅

---

## 🔐 資料安全

### 敏感欄位加密

以下欄位必須加密儲存：
- `oauth_accounts.access_token`
- `oauth_accounts.refresh_token`

### 加密方式
- 演算法：**AES-256-GCM**
- 金鑰管理：環境變數 `ENCRYPTION_KEY`（32 bytes，base64 編碼）
- 生產環境：使用 AWS KMS / GCP KMS

---

## 🗃️ 資料備份策略

### 備份頻率
- **完整備份**：每日 00:00（UTC+8）
- **增量備份**：每 6 小時
- **保留期限**：30 天

### 備份內容
- 所有資料表
- 不包含：已軟刪除超過 30 天的資料

---

## ✅ GORM 模型範例

```go
// User 模型
type User struct {
    ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Email             string         `gorm:"uniqueIndex;not null"`
    Name              string         `gorm:"size:100"`
    ProfilePictureURL *string
    AIReplyTone       string         `gorm:"default:'professional'"`
    Timezone          string         `gorm:"default:'Asia/Taipei'"`
    NotificationPrefs datatypes.JSON `gorm:"type:jsonb"`
    CreatedAt         time.Time
    UpdatedAt         time.Time
    LastLoginAt       *time.Time
    DeletedAt         gorm.DeletedAt `gorm:"index"`
    
    // Relations
    OAuthAccounts []OAuthAccount
    Cases         []Case
}

// GetPrimaryOAuthAccount 取得主要的 OAuth 帳號（通常是用來登入的帳號）
// 優先返回 Google 帳號（系統預設登入方式）
func (u *User) GetPrimaryOAuthAccount() *OAuthAccount {
    for i := range u.OAuthAccounts {
        if u.OAuthAccounts[i].Provider == OAuthProviderGoogle {
            return &u.OAuthAccounts[i]
        }
    }
    if len(u.OAuthAccounts) > 0 {
        return &u.OAuthAccounts[0]
    }
    return nil
}

// Case 模型
type Case struct {
    ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    UserID              uuid.UUID      `gorm:"type:uuid;not null;index"`
    Title               string         `gorm:"not null"`
    BrandName           string         `gorm:"not null;index"`
    CollaborationType   string
    Description         string
    QuotedAmount        *float64       `gorm:"type:numeric(12,2)"`
    FinalAmount         *float64       `gorm:"type:numeric(12,2)"`
    Currency            string         `gorm:"default:'TWD'"`
    PaymentStatus       string         `gorm:"default:'pending'"`
    Status              string         `gorm:"not null;default:'to_confirm';index"`
    ContractDate        *time.Time     `gorm:"type:date"`
    DeadlineDate        *time.Time     `gorm:"type:date;index"`
    DeliveryDate        *time.Time     `gorm:"type:date"`
    PublishDate         *time.Time     `gorm:"type:date"`
    ContactName         string
    ContactEmail        string
    ContactPhone        string
    Notes               string
    Tags                pq.StringArray `gorm:"type:text[]"`
    Source              string         `gorm:"default:'email'"`
    CreatedFromEmailID  *uuid.UUID
    CreatedAt           time.Time
    UpdatedAt           time.Time
    DeletedAt           gorm.DeletedAt `gorm:"index"`
    
    // Relations
    User        User
    Emails      []Email      `gorm:"many2many:case_emails"`
    Tasks       []Task
    Replies     []Reply
    CaseUpdates []CaseUpdate
}
```

---

**下一步**：生成 API 規格文件（`contracts/api-spec.json`）與實作計劃（`plan.md`）

