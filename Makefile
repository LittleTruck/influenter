# Influenter - Makefile
.PHONY: help dev up down logs clean backend-init frontend-init migrate-up migrate-down test

# 預設目標
.DEFAULT_GOAL := help

# 顏色定義
COLOR_RESET = \033[0m
COLOR_BOLD = \033[1m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m
COLOR_BLUE = \033[34m

## help: 顯示此幫助訊息
help:
	@echo "$(COLOR_BOLD)Influenter - 可用命令:$(COLOR_RESET)"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  $(COLOR_GREEN)/' | sed 's/:/ $(COLOR_RESET)-/'
	@echo ""

## dev: 啟動完整開發環境 (所有服務在 Docker 中)
dev:
	@echo "$(COLOR_BLUE)🚀 啟動開發環境...$(COLOR_RESET)"
	docker-compose up -d
	@echo "$(COLOR_GREEN)✅ Docker 服務已啟動$(COLOR_RESET)"
	@echo ""
	@echo "服務列表:"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "  - Backend API: http://localhost:8080"
	@echo "  - Frontend: http://localhost:3000"
	@echo "  - Asynq Monitor: http://localhost:8180"

## up: 啟動所有 Docker 服務
up:
	@echo "$(COLOR_BLUE)🚀 啟動 Docker 服務...$(COLOR_RESET)"
	docker-compose up -d
	@echo "$(COLOR_GREEN)✅ 完成$(COLOR_RESET)"

## down: 停止所有 Docker 服務
down:
	@echo "$(COLOR_YELLOW)🛑 停止 Docker 服務...$(COLOR_RESET)"
	docker-compose down
	@echo "$(COLOR_GREEN)✅ 完成$(COLOR_RESET)"

## logs: 查看服務日誌
logs:
	docker-compose logs -f

## logs-api: 查看 API 日誌
logs-api:
	docker-compose logs -f backend-api

## logs-worker: 查看 Worker 日誌
logs-worker:
	docker-compose logs -f backend-worker

## logs-frontend: 查看 Frontend 日誌
logs-frontend:
	docker-compose logs -f frontend

## clean: 清理所有容器和資料卷
clean:
	@echo "$(COLOR_YELLOW)⚠️  警告: 這將刪除所有資料！$(COLOR_RESET)"
	@read -p "確定要繼續嗎? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		docker-compose down -v; \
		echo "$(COLOR_GREEN)✅ 清理完成$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_BLUE)已取消$(COLOR_RESET)"; \
	fi

## backend-init: 初始化後端專案
backend-init:
	@echo "$(COLOR_BLUE)📦 初始化後端專案...$(COLOR_RESET)"
	cd backend && go mod init github.com/yourusername/influenter-backend
	cd backend && go mod tidy
	@echo "$(COLOR_GREEN)✅ 後端專案已初始化$(COLOR_RESET)"

## frontend-init: 初始化前端專案
frontend-init:
	@echo "$(COLOR_BLUE)📦 初始化前端專案...$(COLOR_RESET)"
	cd frontend && npm install
	@echo "$(COLOR_GREEN)✅ 前端專案已初始化$(COLOR_RESET)"

## frontend-dev: 啟動前端開發伺服器 (本機)
frontend-dev:
	@echo "$(COLOR_BLUE)🎨 啟動前端開發伺服器...$(COLOR_RESET)"
	cd frontend && npm run dev

## frontend-build: 構建前端專案
frontend-build:
	@echo "$(COLOR_BLUE)🔨 構建前端專案...$(COLOR_RESET)"
	docker-compose exec frontend npm run build
	@echo "$(COLOR_GREEN)✅ 構建完成$(COLOR_RESET)"

## migrate-up: 執行資料庫遷移 (升級)
migrate-up:
	@echo "$(COLOR_BLUE)📊 執行資料庫遷移...$(COLOR_RESET)"
	docker-compose exec backend-api go run ./cmd/migrate/main.go up
	@echo "$(COLOR_GREEN)✅ 遷移完成$(COLOR_RESET)"

## migrate-down: 回滾資料庫遷移
migrate-down:
	@echo "$(COLOR_YELLOW)⚠️  回滾資料庫遷移...$(COLOR_RESET)"
	docker-compose exec backend-api go run ./cmd/migrate/main.go down
	@echo "$(COLOR_GREEN)✅ 回滾完成$(COLOR_RESET)"

## migrate-status: 查看遷移狀態
migrate-status:
	@echo "$(COLOR_BLUE)📋 查看遷移狀態...$(COLOR_RESET)"
	docker-compose exec backend-api go run ./cmd/migrate/main.go status

## migrate-create: 創建新遷移 (使用方式: make migrate-create NAME=your_migration_name)
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "$(COLOR_YELLOW)❌ 請指定遷移名稱: make migrate-create NAME=your_migration_name$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BLUE)📝 創建遷移: $(NAME)...$(COLOR_RESET)"
	docker-compose exec backend-api go run ./cmd/migrate/main.go create $(NAME)
	@echo "$(COLOR_GREEN)✅ 遷移檔案已建立$(COLOR_RESET)"

## db-reset: 重置資料庫（刪除所有資料並重新執行 migrations）
db-reset:
	@echo "$(COLOR_YELLOW)⚠️  警告: 這將刪除所有資料！$(COLOR_RESET)"
	@read -p "確定要繼續嗎? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		echo "$(COLOR_BLUE)🔄 重置資料庫...$(COLOR_RESET)"; \
		docker-compose down postgres; \
		docker volume rm influenter_postgres_data 2>/dev/null || true; \
		docker-compose up -d postgres; \
		echo "⏳ 等待資料庫啟動..."; \
		sleep 8; \
		docker-compose exec backend-api go run ./cmd/migrate/main.go up; \
		echo "$(COLOR_GREEN)✅ 資料庫重置完成$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_BLUE)已取消$(COLOR_RESET)"; \
	fi

## test: 執行測試
test:
	@echo "$(COLOR_BLUE)🧪 執行測試...$(COLOR_RESET)"
	cd backend && go test ./... -v
	@echo "$(COLOR_GREEN)✅ 測試完成$(COLOR_RESET)"

## ps: 查看運行中的服務
ps:
	docker-compose ps

## restart: 重啟所有服務
restart:
	@echo "$(COLOR_YELLOW)🔄 重啟服務...$(COLOR_RESET)"
	docker-compose restart
	@echo "$(COLOR_GREEN)✅ 完成$(COLOR_RESET)"

## shell-api: 進入 API 容器的 shell
shell-api:
	docker-compose exec backend-api sh

## shell-db: 進入 PostgreSQL 容器
shell-db:
	docker-compose exec postgres psql -U influenter_user -d influenter

## shell-frontend: 進入 Frontend 容器的 shell
shell-frontend:
	docker-compose exec frontend sh

## prod-up: 啟動生產環境
prod-up:
	@echo "$(COLOR_BLUE)🚀 啟動生產環境...$(COLOR_RESET)"
	docker-compose -f docker-compose.prod.yml up -d
	@echo "$(COLOR_GREEN)✅ 生產環境已啟動$(COLOR_RESET)"

## prod-down: 停止生產環境
prod-down:
	docker-compose -f docker-compose.prod.yml down

