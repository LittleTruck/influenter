# Influenter MVP - 快速開始指南

> **給開發者的快速參考**

---

## 📖 文件導覽

### 核心文件（必讀）
1. **[Constitution](../../memory/constitution.md)** - 專案憲章與價值觀
2. **[Spec](spec.md)** - 完整功能規格
3. **[Clarifications](clarifications.md)** - 需求釐清結果
4. **[Research](research.md)** - 技術研究與選型
5. **[Data Model](data-model.md)** - 資料庫設計
6. **[API Spec](contracts/api-spec.md)** - RESTful API 文件
7. **[Plan](plan.md)** - 實作計劃（6 個 Phase）
8. **[Tasks](tasks.md)** - 詳細任務分解

---

## 🎯 專案概覽

**Influenter** 是為台灣創作者設計的 AI 驅動合作案件管理系統。

### 核心功能
- 📧 Gmail 自動同步
- 🤖 AI 智慧分類與資訊抽取
- 📊 案件專案化管理
- ✍️ AI 協助生成回覆
- 📅 時程管理與提醒
- 💰 報價方案管理

---

## 🛠️ 技術棧

| 領域 | 技術 |
|------|------|
| **前端** | Vue 3 + Nuxt 3 + Vuetify 3 + TypeScript |
| **後端** | Go + Gin + GORM |
| **資料庫** | PostgreSQL 15+ |
| **快取/任務** | Redis + Asynq |
| **AI** | OpenAI GPT-4o-mini |
| **認證** | Google OAuth 2.0 + JWT |
| **整合** | Gmail API |

---

## 🚀 開始開發

### Step 1: 環境準備
```bash
# 安裝依賴
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis
```

### Step 2: Clone 並設定
```bash
# Clone repository
git clone <repo-url>
cd influenter

# 設定環境變數
cp .env.example .env
# 編輯 .env 填入必要資訊
```

### Step 3: 初始化資料庫
```bash
# 建立資料庫
createdb influenter

# 執行遷移
cd backend
go run cmd/migrate/main.go up
```

### Step 4: 啟動服務
```bash
# Terminal 1: 後端
cd backend
go run cmd/server/main.go

# Terminal 2: Worker
cd backend
go run cmd/worker/main.go

# Terminal 3: 前端
cd frontend
npm install
npm run dev
```

### Step 5: 開啟瀏覽器
訪問 `http://localhost:3000`

---

## 📋 開發流程（Spec-Driven）

我們已經完成了 Spec-Driven Development 的前置階段：

### ✅ 已完成
1. ✅ **憲章建立** - 定義專案價值觀與原則
2. ✅ **規格撰寫** - 詳細功能規格（spec.md）
3. ✅ **需求釐清** - 解決模糊需求（clarifications.md）
4. ✅ **技術研究** - 框架選型與最佳實踐（research.md）
5. ✅ **資料模型** - 完整資料庫設計（data-model.md）
6. ✅ **API 設計** - RESTful API 規格（contracts/api-spec.md）
7. ✅ **實作計劃** - 6 個 Phase 的詳細計劃（plan.md）
8. ✅ **任務分解** - 480 小時的可執行任務（tasks.md）

### 🎯 下一步：開始實作！
按照 **[Plan.md](plan.md)** 的順序，逐步完成每個 Phase：

- **Phase 0** (Week 1): 專案初始化
- **Phase 1** (Week 2-3): 認證系統
- **Phase 2** (Week 4-5): Gmail 整合
- **Phase 3** (Week 6-7): AI 分析
- **Phase 4** (Week 8-9): 案件管理
- **Phase 5** (Week 10-11): 回覆生成與進階功能
- **Phase 6** (Week 12): 測試與部署

---

## 💡 重要原則

### 從 Constitution 中記住的核心價值
1. **使用者至上** - 以台灣創作者實際需求為核心
2. **AI 賦能** - AI 是輔助，決策權在使用者
3. **漸進式交付** - 優先核心功能，快速迭代
4. **技術卓越** - 乾淨程式碼、完善測試

### 不要做的事情
- ❌ 不做社群媒體排程
- ❌ 不做金流處理
- ❌ 不做多語言（專注台灣）
- ❌ 不過度設計

---

## 📚 參考資源

### 官方文件
- [Gmail API](https://developers.google.com/gmail/api)
- [OpenAI API](https://platform.openai.com/docs)
- [Nuxt 3](https://nuxt.com/docs)
- [Vuetify 3](https://vuetifyjs.com/)
- [Gin](https://gin-gonic.com/docs/)
- [GORM](https://gorm.io/docs/)

### Spec-Kit
- [GitHub Spec-Kit](https://github.com/github/spec-kit)
- [Spec-Driven Development](https://github.com/github/spec-kit/blob/main/spec-driven.md)

---

## ✅ 檢查清單

開始開發前，確認：
- [ ] 已閱讀 Constitution
- [ ] 已閱讀 Spec
- [ ] 已閱讀 Clarifications
- [ ] 了解技術選型原因（Research）
- [ ] 理解資料模型（Data Model）
- [ ] 熟悉 API 規格
- [ ] 清楚開發計劃（Plan）
- [ ] 準備好開始執行任務（Tasks）

---

## 🎉 準備就緒！

現在您已經擁有完整的開發藍圖，可以開始建構 Influenter 了！

按照 **tasks.md** 中的任務，逐一完成，每個任務都設計成 4 小時內可完成的大小。

祝開發順利！🚀

---

**有任何問題請參考對應文件，或回顧 Constitution 中的原則。**

