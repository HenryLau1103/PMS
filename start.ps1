# PSM 啟動腳本 (PowerShell)

Write-Host "🚀 啟動 PSM 台股智能投資組合管理系統" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# 檢查 Docker
Write-Host "🔍 檢查 Docker 狀態..." -ForegroundColor Yellow
try {
    $dockerVersion = docker --version
    Write-Host "✅ Docker 已安裝: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker 未安裝或未啟動" -ForegroundColor Red
    Write-Host "請先安裝並啟動 Docker Desktop" -ForegroundColor Red
    Read-Host "按 Enter 鍵退出"
    exit 1
}

# 測試 Docker 連接
Write-Host "🔍 測試 Docker 連接..." -ForegroundColor Yellow
try {
    docker ps | Out-Null
    Write-Host "✅ Docker 連接正常" -ForegroundColor Green
} catch {
    Write-Host "❌ 無法連接到 Docker Engine" -ForegroundColor Red
    Write-Host "" -ForegroundColor Red
    Write-Host "請執行以下步驟:" -ForegroundColor Yellow
    Write-Host "1. 打開 Docker Desktop 應用程式" -ForegroundColor White
    Write-Host "2. 等待底部狀態顯示 'Engine running' (綠色)" -ForegroundColor White
    Write-Host "3. 如果 Docker Desktop 無法啟動，請重啟電腦" -ForegroundColor White
    Write-Host "" -ForegroundColor Red
    Read-Host "按 Enter 鍵退出"
    exit 1
}

Write-Host ""
Write-Host "📦 啟動服務..." -ForegroundColor Yellow
docker-compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 服務啟動失敗" -ForegroundColor Red
    Write-Host "查看錯誤日誌: docker-compose logs" -ForegroundColor Yellow
    Read-Host "按 Enter 鍵退出"
    exit 1
}

Write-Host ""
Write-Host "⏳ 等待服務就緒 (30秒)..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

Write-Host ""
Write-Host "🔍 檢查服務狀態..." -ForegroundColor Yellow
docker-compose ps

Write-Host ""
Write-Host "🏥 健康檢查..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Backend API: 正常運行" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️  Backend API: 尚未完全就緒，請稍後再試" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "✅ PSM 系統啟動完成！" -ForegroundColor Green
Write-Host ""
Write-Host "訪問以下服務:" -ForegroundColor White
Write-Host "  🌐 Frontend:    http://localhost:3000" -ForegroundColor Cyan
Write-Host "  🔌 Backend API: http://localhost:8080" -ForegroundColor Cyan
Write-Host "  🗄️  Database:    localhost:5432" -ForegroundColor Cyan
Write-Host ""
Write-Host "實用命令:" -ForegroundColor White
Write-Host "  查看日誌:   docker-compose logs -f" -ForegroundColor Gray
Write-Host "  停止服務:   docker-compose down" -ForegroundColor Gray
Write-Host "  重啟服務:   docker-compose restart" -ForegroundColor Gray
Write-Host ""
Write-Host "如遇問題，請查看: TROUBLESHOOTING.md" -ForegroundColor Yellow
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Read-Host "按 Enter 鍵退出"
