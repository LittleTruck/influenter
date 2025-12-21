# 釋放 Windows 保留的端口範圍
# 注意：此腳本需要以管理員權限執行

Write-Host "檢查當前保留的端口範圍..." -ForegroundColor Cyan
netsh interface ipv4 show excludedportrange protocol=tcp

Write-Host "`n警告：釋放端口範圍可能會影響 Hyper-V 或其他 Windows 功能" -ForegroundColor Yellow
Write-Host "如果您的系統使用 Hyper-V，建議不要執行此操作" -ForegroundColor Yellow

$confirm = Read-Host "`n確定要釋放保留的端口範圍嗎？(y/N)"
if ($confirm -ne "y" -and $confirm -ne "Y") {
    Write-Host "已取消操作" -ForegroundColor Yellow
    exit
}

# 常見的保留範圍（根據之前看到的）
$ranges = @(
    @{start=7973; end=8072},
    @{start=8073; end=8172},
    @{start=8266; end=8365}
)

foreach ($range in $ranges) {
    Write-Host "嘗試移除端口範圍 $($range.start)-$($range.end)..." -ForegroundColor Cyan
    try {
        # 注意：這個命令的語法可能因 Windows 版本而異
        # 如果這個命令失敗，您可能需要通過其他方式釋放端口
        netsh interface ipv4 delete excludedportrange protocol=tcp startport=$($range.start) numberofports=$($range.end - $range.start + 1)
        Write-Host "✓ 成功移除端口範圍 $($range.start)-$($range.end)" -ForegroundColor Green
    } catch {
        Write-Host "✗ 無法移除端口範圍 $($range.start)-$($range.end): $_" -ForegroundColor Red
        Write-Host "提示：這些範圍可能由 Hyper-V 或其他服務保留，需要先停用相關服務" -ForegroundColor Yellow
    }
}

Write-Host "`n檢查更新後的保留端口範圍..." -ForegroundColor Cyan
netsh interface ipv4 show excludedportrange protocol=tcp

Write-Host "`n如果端口範圍仍然存在，您可能需要：" -ForegroundColor Yellow
Write-Host "1. 停用 Hyper-V (如果未使用): Disable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All"
Write-Host "2. 或使用其他端口（如 5000, 9000 等）" -ForegroundColor Yellow


