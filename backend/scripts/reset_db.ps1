# 重置資料庫腳本（僅限開發環境使用）
# 警告：此腳本會刪除所有資料！

Write-Host "⚠️  警告：此操作會刪除所有資料！" -ForegroundColor Red
Write-Host "請確認您要重置開發資料庫 (y/N): " -NoNewline
$confirmation = Read-Host

if ($confirmation -ne 'y') {
    Write-Host "操作已取消" -ForegroundColor Yellow
    exit
}

Write-Host "`n📦 正在重置資料庫..." -ForegroundColor Cyan

# 方式 1: 使用 Docker Compose 重新建立資料庫容器
Write-Host "停止並移除資料庫容器..." -ForegroundColor Yellow
docker-compose down -v postgres

Write-Host "重新啟動資料庫..." -ForegroundColor Yellow
docker-compose up -d postgres

# 等待資料庫啟動
Write-Host "等待資料庫啟動..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# 執行 migrations
Write-Host "`n🔄 執行 migrations..." -ForegroundColor Cyan
go run cmd/migrate/main.go up

Write-Host "`n✅ 資料庫重置完成！" -ForegroundColor Green

