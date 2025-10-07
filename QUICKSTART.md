# 🚀 Influenter - 快速啟動指南

## 📋 第一步：環境準備

### 1. 安裝必要軟體

確保您已安裝：
- ✅ **Docker Desktop** (包含 Docker Compose)
- ✅ **Git**
- ✅ **Node.js 18+** (用於前端開發)
- ✅ **Go 1.21+** (選用，若要在本機開發後端)

### 2. 驗證安裝

```bash
# 檢查 Docker
docker --version
docker-compose --version

# 檢查 Node.js
node --version
npm --version

# 檢查 Git
git --version
```

---

## 🔧 第二步：專案設置

### 1. 建立環境變數檔案

複製以下內容並儲存為 `.env` 檔案在專案根目錄：

```env
# Google OAuth（暫時使用假資料，稍後設定）
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# JWT Secret（暫時使用這個，生產環境要改）
JWT_SECRET=dev-jwt-secret-please-change-in-production-12345678

# Encryption Key（32 bytes，暫時使用這個）
ENCRYPTION_KEY=12345678901234567890123456789012
```

> ⚠️ **注意**：這些是開發用的假資料。實際使用前需要設定真實的 Google OAuth 憑證。詳見 `ENV_SETUP.md`。

### 2. 初始化 Git（如果還沒有）

```bash
git init
git add .
git commit -m "Initial commit: Spec-Kit setup with Docker configuration"
```

---

## 🐳 第三步：啟動 Docker 服務

### 使用 Make（推薦）

```bash
# 查看所有可用命令
make help

# 啟動開發環境
make dev
```

### 使用 Docker Compose

```bash
# 啟動所有服務
docker-compose up -d

# 查看服務狀態
docker-compose ps

# 查看日誌
docker-compose logs -f
```

### 服務啟動後

您應該會看到以下服務運行：
- ✅ PostgreSQL (port 5432)
- ✅ Redis (port 6379)
- ✅ Backend API (port 8080)
- ✅ Backend Worker
- ✅ Asynq Monitor (port 8081)

---

## 📦 第四步：初始化專案

### 1. 初始化後端專案

```bash
# 進入後端目錄
cd backend

# 初始化 Go module
go mod init github.com/yourusername/influenter-backend

# 安裝依賴
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/oauth2
go get google.golang.org/api/gmail/v1
go get github.com/sashabaranov/go-openai
go get github.com/hibiken/asynq

# 整理依賴
go mod tidy

# 回到根目錄
cd ..
```

### 2. 初始化前端專案

```bash
# 進入前端目錄
cd frontend

# 使用 Nuxt 3 初始化（如果還沒有）
npx nuxi@latest init . --force

# 或者手動安裝依賴
npm install

# 安裝 Vuetify 和其他套件
npm install vuetify @mdi/font pinia @vite-pwa/nuxt

# 回到根目錄
cd ..
```

---

## 🎨 第五步：啟動前端開發伺服器

開啟**新的終端機視窗**：

```bash
cd frontend
npm run dev
```

前端將在 http://localhost:3000 啟動

---

## ✅ 驗證安裝

### 1. 檢查 Docker 服務

```bash
# 查看運行中的容器
docker ps

# 應該看到 5 個容器在運行：
# - influenter-postgres
# - influenter-redis
# - influenter-backend-api
# - influenter-backend-worker
# - influenter-asynqmon
```

### 2. 測試後端 API

開啟瀏覽器或使用 curl：

```bash
# Health check
curl http://localhost:8080/health

# 應該返回 JSON 回應（目前會失敗，因為還沒建立 API）
```

### 3. 訪問服務

- 🎨 **前端**: http://localhost:3000
- 🔌 **後端 API**: http://localhost:8080
- 📊 **Asynq Monitor**: http://localhost:8081

---

## 📝 當前狀態

您現在已經完成：
- ✅ Spec-Kit 完整規劃（所有文件）
- ✅ Docker 配置（開發 + 生產）
- ✅ 專案目錄結構
- ✅ Makefile 快捷命令
- ✅ Docker 服務運行中

### 下一步：開始開發！

按照 **Phase 0** 的任務開始實作：

1. **建立後端基礎程式碼**
   - `backend/cmd/server/main.go` (API server 進入點)
   - `backend/internal/config/config.go` (設定管理)
   - `backend/internal/database/db.go` (資料庫連線)

2. **建立前端基礎程式碼**
   - 設定 Vuetify
   - 建立 Layout
   - 建立基礎頁面

詳見：`specs/001-influenter-mvp/tasks.md` 的 **Phase 0** 任務清單。

---

## 🛠️ 常用命令速查

```bash
# 啟動開發環境
make dev

# 停止所有服務
make down

# 查看日誌
make logs

# 重啟服務
make restart

# 進入 API 容器
make shell-api

# 進入資料庫
make shell-db

# 清理所有資料
make clean
```

---

## 🆘 遇到問題？

### Docker 相關

**問題：容器無法啟動**
```bash
# 查看詳細錯誤
docker-compose logs backend-api

# 重新建置容器
docker-compose up -d --build
```

**問題：Port 已被佔用**
```bash
# 查看佔用 port 的程式
# Windows:
netstat -ano | findstr :8080

# macOS/Linux:
lsof -i :8080

# 修改 docker-compose.yml 中的 port mapping
```

### 前端相關

**問題：npm install 失敗**
```bash
# 清除 cache 重試
npm cache clean --force
npm install
```

**問題：Nuxt 啟動失敗**
```bash
# 刪除 node_modules 重新安裝
rm -rf node_modules package-lock.json
npm install
```

### 後端相關

**問題：go mod 錯誤**
```bash
# 重新整理依賴
go mod tidy
go mod download
```

---

## 📚 參考文件

- [完整 README](README.md)
- [環境變數設定指南](ENV_SETUP.md)
- [功能規格](specs/001-influenter-mvp/spec.md)
- [實作計劃](specs/001-influenter-mvp/plan.md)
- [任務分解](specs/001-influenter-mvp/tasks.md)

---

**恭喜！您已經完成環境設置，可以開始開發了！** 🎉

建議從 `specs/001-influenter-mvp/tasks.md` 的 **P0-BACKEND-001** 開始執行。

