# Influenter API 規格文件

> **版本**: v1.0  
> **Base URL**: `http://localhost:8080/api/v1`  
> **認證方式**: JWT Bearer Token  
> **最後更新**: 2025-10-07

---

## 🔐 認證

所有 API 端點（除了認證相關）都需要在 Header 中帶上 JWT token：

```
Authorization: Bearer <token>
```

---

## 📑 API 端點總覽

### 認證相關
- `GET /auth/google/login` - 觸發 Google OAuth
- `GET /auth/google/callback` - OAuth 回調
- `POST /auth/logout` - 登出
- `GET /auth/me` - 取得當前使用者資訊

### Gmail 整合
- `POST /gmail/connect` - 連接 Gmail 帳號
- `GET /gmail/status` - 取得同步狀態
- `POST /gmail/sync` - 手動觸發同步
- `DELETE /gmail/disconnect` - 斷開 Gmail 連接

### 郵件管理
- `GET /emails` - 取得郵件列表
- `GET /emails/:id` - 取得郵件詳情
- `PATCH /emails/:id` - 更新郵件（標記已讀等）

### 案件管理
- `GET /cases` - 取得案件列表
- `POST /cases` - 建立新案件
- `GET /cases/:id` - 取得案件詳情
- `PATCH /cases/:id` - 更新案件
- `DELETE /cases/:id` - 刪除案件（軟刪除）
- `GET /cases/:id/emails` - 取得案件相關郵件
- `POST /cases/:id/emails` - 關聯郵件到案件

### 任務管理
- `GET /cases/:caseId/tasks` - 取得案件任務列表
- `POST /cases/:caseId/tasks` - 建立新任務
- `PATCH /tasks/:id` - 更新任務
- `DELETE /tasks/:id` - 刪除任務
- `POST /tasks/:id/complete` - 標記任務完成

### 回覆生成
- `POST /replies/generate` - AI 生成回覆
- `POST /replies/send` - 寄送回覆
- `GET /cases/:caseId/replies` - 取得案件回覆歷史

### 報價方案
- `GET /pricing-packages` - 取得方案列表
- `POST /pricing-packages` - 建立新方案
- `PATCH /pricing-packages/:id` - 更新方案
- `DELETE /pricing-packages/:id` - 刪除方案

### 通知
- `GET /notifications` - 取得通知列表
- `PATCH /notifications/:id/read` - 標記通知已讀
- `POST /notifications/read-all` - 全部標記已讀

### 統計資料
- `GET /stats/dashboard` - 儀表板統計
- `GET /stats/revenue` - 收入統計

---

## 📖 API 端點詳細規格

### 1. 認證相關

#### `GET /auth/google/login`
**描述**：重定向到 Google OAuth 授權頁面

**回應**：
- HTTP 302 重定向到 Google

---

#### `GET /auth/google/callback`
**描述**：處理 Google OAuth 回調

**Query Parameters**：
- `code` (string, required): Google 提供的 authorization code
- `state` (string, required): CSRF token

**回應**：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "alice@example.com",
    "name": "Alice Chen",
    "avatar_url": "https://..."
  }
}
```

**錯誤回應**：
- `400 Bad Request`: code 無效或已過期
- `500 Internal Server Error`: 無法取得使用者資訊

---

#### `GET /auth/me`
**描述**：取得當前登入使用者資訊

**認證**：Required

**回應**：
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "name": "Alice Chen",
  "avatar_url": "https://...",
  "gmail_connected": true,
  "ai_reply_tone": "friendly",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### 2. Gmail 整合

#### `GET /gmail/status`
**描述**：取得 Gmail 同步狀態

**認證**：Required

**回應**：
```json
{
  "connected": true,
  "email": "alice@example.com",
  "last_sync_at": "2024-03-15T14:30:00Z",
  "sync_status": "active",
  "total_emails": 1250,
  "unread_emails": 15,
  "next_sync_in": 180
}
```

**欄位說明**：
- `next_sync_in`: 下次自動同步的秒數

---

#### `POST /gmail/sync`
**描述**：手動觸發 Gmail 同步

**認證**：Required

**回應**：
```json
{
  "message": "Sync started",
  "estimated_duration": 30
}
```

**錯誤回應**：
- `429 Too Many Requests`: 同步冷卻時間未到（1 分鐘內只能同步一次）

---

### 3. 郵件管理

#### `GET /emails`
**描述**：取得郵件列表（分頁）

**認證**：Required

**Query Parameters**：
- `page` (integer, default: 1): 頁碼
- `per_page` (integer, default: 20, max: 100): 每頁筆數
- `is_read` (boolean, optional): 過濾已讀/未讀
- `has_case` (boolean, optional): 是否已關聯案件
- `search` (string, optional): 搜尋關鍵字（主旨、寄件者）
- `sort` (string, default: "received_at_desc"): 排序方式

**回應**：
```json
{
  "data": [
    {
      "id": "email-uuid-1",
      "from_email": "brand@example.com",
      "from_name": "Brand Manager",
      "subject": "合作邀約：新產品開箱影片",
      "snippet": "您好，我們是 XX 品牌...",
      "received_at": "2024-03-15T10:00:00Z",
      "is_read": false,
      "has_attachments": false,
      "case_id": null,
      "ai_category": "合作邀約",
      "ai_confidence": 0.95
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

#### `GET /emails/:id`
**描述**：取得郵件詳細內容

**認證**：Required

**回應**：
```json
{
  "id": "email-uuid-1",
  "from_email": "brand@example.com",
  "from_name": "Brand Manager",
  "to_email": "alice@example.com",
  "subject": "合作邀約：新產品開箱影片",
  "body_text": "完整郵件內容...",
  "body_html": "<p>完整郵件內容...</p>",
  "received_at": "2024-03-15T10:00:00Z",
  "is_read": false,
  "has_attachments": false,
  "labels": ["INBOX", "IMPORTANT"],
  "case_id": null,
  "ai_analysis": {
    "category": "合作邀約",
    "brand_name": "XX 品牌",
    "amount": 30000,
    "currency": "TWD",
    "important_dates": [
      {
        "date": "2024-03-25",
        "type": "deadline",
        "description": "回覆截止日"
      }
    ],
    "contact_name": "王小明",
    "collaboration_type": "開箱影片",
    "confidence": 0.95
  }
}
```

---

### 4. 案件管理

#### `GET /cases`
**描述**：取得案件列表

**認證**：Required

**Query Parameters**：
- `page` (integer, default: 1)
- `per_page` (integer, default: 20)
- `status` (string, optional): to_confirm, in_progress, completed, cancelled
- `brand` (string, optional): 品牌名稱篩選
- `sort` (string, default: "updated_at_desc")

**回應**：
```json
{
  "data": [
    {
      "id": "case-uuid-1",
      "title": "XX 品牌 - 新產品開箱影片",
      "brand_name": "XX 品牌",
      "collaboration_type": "開箱影片",
      "status": "in_progress",
      "quoted_amount": 30000,
      "final_amount": 32000,
      "currency": "TWD",
      "deadline_date": "2024-03-30",
      "contact_name": "王小明",
      "email_count": 5,
      "task_count": 3,
      "completed_task_count": 1,
      "created_at": "2024-03-15T10:00:00Z",
      "updated_at": "2024-03-16T14:20:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

---

#### `POST /cases`
**描述**：手動建立新案件

**認證**：Required

**Request Body**：
```json
{
  "title": "XX 品牌 - 新產品開箱影片",
  "brand_name": "XX 品牌",
  "collaboration_type": "開箱影片",
  "description": "合作細節...",
  "quoted_amount": 30000,
  "deadline_date": "2024-03-30",
  "contact_name": "王小明",
  "contact_email": "wang@example.com",
  "contact_phone": "0912-345-678",
  "notes": "備註...",
  "tags": ["業配", "美妝"]
}
```

**回應**：
```json
{
  "id": "case-uuid-1",
  "title": "XX 品牌 - 新產品開箱影片",
  "status": "to_confirm",
  "created_at": "2024-03-15T10:00:00Z"
}
```

**驗證規則**：
- `title`: 必填，最多 255 字元
- `brand_name`: 必填，最多 255 字元
- `quoted_amount`: 選填，必須 > 0
- `deadline_date`: 選填，必須是未來日期

---

#### `PATCH /cases/:id`
**描述**：更新案件資訊

**認證**：Required

**Request Body** (所有欄位皆選填)：
```json
{
  "title": "新標題",
  "status": "in_progress",
  "final_amount": 32000,
  "payment_status": "completed",
  "notes": "更新備註"
}
```

**回應**：
```json
{
  "id": "case-uuid-1",
  "message": "Case updated successfully",
  "updated_at": "2024-03-16T14:20:00Z"
}
```

---

#### `GET /cases/:id`
**描述**：取得案件完整詳情

**認證**：Required

**回應**：
```json
{
  "id": "case-uuid-1",
  "title": "XX 品牌 - 新產品開箱影片",
  "brand_name": "XX 品牌",
  "collaboration_type": "開箱影片",
  "description": "合作細節...",
  "status": "in_progress",
  "quoted_amount": 30000,
  "final_amount": 32000,
  "currency": "TWD",
  "payment_status": "pending",
  "contract_date": "2024-03-16",
  "deadline_date": "2024-03-30",
  "delivery_date": null,
  "publish_date": null,
  "contact_name": "王小明",
  "contact_email": "wang@example.com",
  "contact_phone": "0912-345-678",
  "notes": "備註...",
  "tags": ["業配", "美妝"],
  "source": "email",
  "emails": [
    {
      "id": "email-uuid-1",
      "subject": "合作邀約",
      "from_email": "wang@example.com",
      "received_at": "2024-03-15T10:00:00Z",
      "email_type": "initial_inquiry"
    }
  ],
  "tasks": [
    {
      "id": "task-uuid-1",
      "title": "腳本提交",
      "due_date": "2024-03-20",
      "status": "completed",
      "completed_at": "2024-03-19T16:00:00Z"
    },
    {
      "id": "task-uuid-2",
      "title": "影片初剪",
      "due_date": "2024-03-25",
      "status": "pending"
    }
  ],
  "updates": [
    {
      "id": "update-uuid-1",
      "update_type": "status_change",
      "old_value": "to_confirm",
      "new_value": "in_progress",
      "created_at": "2024-03-16T10:00:00Z"
    }
  ],
  "created_at": "2024-03-15T10:00:00Z",
  "updated_at": "2024-03-16T14:20:00Z"
}
```

---

### 5. 任務管理

#### `GET /cases/:caseId/tasks`
**描述**：取得案件的所有任務

**認證**：Required

**回應**：
```json
{
  "data": [
    {
      "id": "task-uuid-1",
      "title": "腳本提交",
      "description": "提交影片腳本給品牌審核",
      "due_date": "2024-03-20",
      "due_time": "17:00:00",
      "status": "completed",
      "completed_at": "2024-03-19T16:00:00Z",
      "source": "auto_generated",
      "created_at": "2024-03-15T10:00:00Z"
    }
  ]
}
```

---

#### `POST /cases/:caseId/tasks`
**描述**：為案件建立新任務

**認證**：Required

**Request Body**：
```json
{
  "title": "影片初剪",
  "description": "完成影片初步剪輯",
  "due_date": "2024-03-25",
  "due_time": "17:00",
  "reminder_days": 1
}
```

**回應**：
```json
{
  "id": "task-uuid-2",
  "title": "影片初剪",
  "status": "pending",
  "created_at": "2024-03-15T10:00:00Z"
}
```

---

#### `POST /tasks/:id/complete`
**描述**：標記任務為完成

**認證**：Required

**回應**：
```json
{
  "id": "task-uuid-2",
  "status": "completed",
  "completed_at": "2024-03-25T15:30:00Z"
}
```

---

### 6. 回覆生成

#### `POST /replies/generate`
**描述**：使用 AI 生成回覆草稿

**認證**：Required

**Request Body**：
```json
{
  "email_id": "email-uuid-1",
  "case_id": "case-uuid-1",
  "reply_tone": "friendly",
  "additional_context": "提到我最近檔期較滿，但可以接 4 月的案子",
  "include_pricing_package_id": "package-uuid-1"
}
```

**欄位說明**：
- `email_id`: 要回覆的郵件 ID
- `case_id`: 相關案件 ID（選填）
- `reply_tone`: professional, friendly, concise
- `additional_context`: 使用者補充資訊（選填）
- `include_pricing_package_id`: 要插入的報價方案 ID（選填）

**回應**：
```json
{
  "id": "reply-uuid-1",
  "ai_generated_content": "王小明您好，\n\n感謝您的合作邀約...",
  "model_used": "gpt-4o-mini",
  "tokens_used": 450,
  "estimated_cost": 0.0003,
  "created_at": "2024-03-15T10:30:00Z"
}
```

**錯誤回應**：
- `400 Bad Request`: email_id 無效或 API Key 未設定
- `500 Internal Server Error`: OpenAI API 錯誤

---

#### `POST /replies/send`
**描述**：寄送回覆（透過 Gmail API）

**認證**：Required

**Request Body**：
```json
{
  "reply_id": "reply-uuid-1",
  "final_content": "使用者編輯後的最終內容..."
}
```

**回應**：
```json
{
  "id": "reply-uuid-1",
  "status": "sent",
  "gmail_message_id": "18d1234567890abc",
  "sent_at": "2024-03-15T10:35:00Z"
}
```

**錯誤回應**：
- `400 Bad Request`: Gmail 未連接或 token 過期
- `500 Internal Server Error`: 寄送失敗

---

### 7. 報價方案

#### `GET /pricing-packages`
**描述**：取得使用者的所有報價方案

**認證**：Required

**Query Parameters**：
- `is_active` (boolean, optional): 只顯示啟用的方案

**回應**：
```json
{
  "data": [
    {
      "id": "package-uuid-1",
      "name": "YouTube 業配影片",
      "description": "標準 YouTube 業配方案",
      "price": 30000,
      "currency": "TWD",
      "items": [
        {
          "name": "60-90 秒業配段落",
          "description": "在影片中置入產品介紹"
        }
      ],
      "terms": "需提前 2 週預約檔期",
      "is_active": true,
      "created_at": "2024-01-10T10:00:00Z"
    }
  ]
}
```

---

#### `POST /pricing-packages`
**描述**：建立新的報價方案

**認證**：Required

**Request Body**：
```json
{
  "name": "Instagram Reels 業配",
  "description": "Instagram Reels 短影片業配",
  "price": 15000,
  "items": [
    {
      "name": "15-30 秒 Reels",
      "description": "短影片業配內容"
    }
  ],
  "terms": "保留影片 14 天",
  "default_duration_days": 7
}
```

**回應**：
```json
{
  "id": "package-uuid-2",
  "name": "Instagram Reels 業配",
  "created_at": "2024-03-15T10:00:00Z"
}
```

---

### 8. 通知

#### `GET /notifications`
**描述**：取得通知列表

**認證**：Required

**Query Parameters**：
- `page` (integer, default: 1)
- `per_page` (integer, default: 20)
- `is_read` (boolean, optional): 過濾已讀/未讀

**回應**：
```json
{
  "data": [
    {
      "id": "notif-uuid-1",
      "type": "new_case",
      "title": "新案件：XX 品牌合作邀約",
      "message": "系統偵測到新的合作邀約郵件並自動建立案件",
      "related_case_id": "case-uuid-1",
      "is_read": false,
      "action_url": "/cases/case-uuid-1",
      "created_at": "2024-03-15T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 50,
    "total_pages": 3
  },
  "unread_count": 12
}
```

---

### 9. 統計資料

#### `GET /stats/dashboard`
**描述**：儀表板統計資料

**認證**：Required

**Query Parameters**：
- `period` (string, default: "month"): month, quarter, year

**回應**：
```json
{
  "period": "month",
  "date_range": {
    "start": "2024-03-01",
    "end": "2024-03-31"
  },
  "cases": {
    "total": 8,
    "to_confirm": 2,
    "in_progress": 4,
    "completed": 2,
    "cancelled": 0
  },
  "revenue": {
    "total": 180000,
    "paid": 120000,
    "pending": 60000,
    "currency": "TWD"
  },
  "emails": {
    "total": 45,
    "unread": 8,
    "ai_classified": 40
  },
  "tasks": {
    "total": 15,
    "pending": 8,
    "completed": 7,
    "overdue": 1
  },
  "top_brands": [
    {
      "brand_name": "XX 品牌",
      "case_count": 3,
      "total_revenue": 90000
    }
  ]
}
```

---

## 📊 資料格式規範

### 日期時間
- 使用 **ISO 8601** 格式
- 時區：**UTC**
- 範例：`2024-03-15T10:30:00Z`

### 金額
- 使用**數字型別**（不使用字串）
- 精度：小數點後 2 位
- 範例：`30000.00`

### UUID
- 使用標準 UUID v4 格式
- 範例：`550e8400-e29b-41d4-a716-446655440000`

---

## ⚠️ 錯誤處理

### 錯誤回應格式
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "案件標題不能為空",
    "details": {
      "field": "title",
      "reason": "required"
    }
  }
}
```

### HTTP 狀態碼
- `200 OK`: 成功
- `201 Created`: 資源建立成功
- `400 Bad Request`: 請求參數錯誤
- `401 Unauthorized`: 未認證或 token 無效
- `403 Forbidden`: 無權限存取
- `404 Not Found`: 資源不存在
- `429 Too Many Requests`: 超過速率限制
- `500 Internal Server Error`: 伺服器錯誤

### 常見錯誤碼
- `INVALID_REQUEST`: 請求參數錯誤
- `UNAUTHORIZED`: 未認證
- `FORBIDDEN`: 無權限
- `NOT_FOUND`: 資源不存在
- `GMAIL_NOT_CONNECTED`: Gmail 未連接
- `GMAIL_TOKEN_EXPIRED`: Gmail token 過期
- `OPENAI_API_ERROR`: OpenAI API 錯誤
- `RATE_LIMIT_EXCEEDED`: 超過速率限制

---

## 🚀 速率限制

### 全域限制
- **每分鐘**: 60 requests
- **每小時**: 1000 requests

### 特定端點限制
- `POST /gmail/sync`: 1 request / 分鐘
- `POST /replies/generate`: 10 requests / 分鐘

### 速率限制 Header
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1710504600
```

---

## ✅ 分頁規範

### 分頁參數
- `page`: 頁碼（從 1 開始）
- `per_page`: 每頁筆數（預設 20，最大 100）

### 分頁回應
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

**API 規格文件完成！** 🎉

