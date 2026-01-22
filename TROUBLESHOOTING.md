# PSM 系統啟動故障排除指南

## 🔧 問題：start.bat 沒有反應或 Docker 連接失敗

### ✅ 解決方案

#### **步驟 1: 確認 Docker Desktop 正常運行**

1. **打開 Docker Desktop 應用程式**
   - 在 Windows 搜尋列輸入 "Docker Desktop"
   - 啟動 Docker Desktop
   - 等待底部狀態顯示 "Engine running" (綠色圖標)

2. **重啟 Docker Desktop (如果需要)**
   - 右鍵點擊系統托盤的 Docker 圖標
   - 選擇 "Restart Docker Desktop"
   - 等待約 30-60 秒

#### **步驟 2: 驗證 Docker 是否可用**

打開 PowerShell 或命令提示字元，執行：

```powershell
docker --version
docker ps
```

如果看到錯誤訊息：
```
error during connect: ... pipe/dockerDesktopLinuxEngine ...
```

這表示 Docker Desktop 沒有完全啟動。

#### **步驟 3: 手動啟動系統 (推薦方法)**

在專案根目錄 `C:\Users\Henry\OneDrive\桌面\PSM`，使用以下任一方法：

##### **方法 A: PowerShell (推薦)**

```powershell
# 1. 開啟 PowerShell (以管理員身份)
# 2. 進入專案目錄
cd "C:\Users\Henry\OneDrive\桌面\PSM"

# 3. 啟動服務
docker-compose up -d

# 4. 等待 30 秒
Start-Sleep -Seconds 30

# 5. 檢查狀態
docker-compose ps

# 6. 查看日誌
docker-compose logs -f
```

##### **方法 B: 命令提示字元 (CMD)**

```cmd
cd C:\Users\Henry\OneDrive\桌面\PSM
docker-compose up -d
timeout /t 30
docker-compose ps
```

##### **方法 C: Git Bash**

```bash
cd /c/Users/Henry/OneDrive/桌面/PSM
docker-compose up -d
sleep 30
docker-compose ps
```

#### **步驟 4: 驗證服務啟動**

執行以下命令檢查服務狀態：

```powershell
# 查看運行中的容器
docker-compose ps

# 查看 Backend 日誌
docker-compose logs backend

# 查看 Database 日誌
docker-compose logs timescaledb

# 查看 Frontend 日誌
docker-compose logs frontend
```

**成功的輸出應該顯示:**
```
NAME                  STATUS
psm-backend           Up
psm-frontend          Up
psm-timescaledb       Up (healthy)
psm-redis             Up (healthy)
```

#### **步驟 5: 訪問應用**

一旦所有服務狀態為 "Up"，打開瀏覽器訪問：

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/health

---

## 🐛 常見問題與解決方案

### **問題 1: Docker Desktop 無法啟動**

**症狀:** Docker Desktop 一直顯示 "Starting..." 或錯誤

**解決方案:**
1. 完全關閉 Docker Desktop
2. 打開工作管理員 (Ctrl+Shift+Esc)
3. 結束所有 Docker 相關進程
4. 重新啟動 Docker Desktop
5. 如果還是失敗，重啟電腦

### **問題 2: Port 已被占用**

**症狀:** 錯誤訊息顯示 "port is already allocated"

**解決方案:**
```powershell
# 查看占用 3000 端口的程序
netstat -ano | findstr :3000

# 查看占用 8080 端口的程序  
netstat -ano | findstr :8080

# 結束進程 (替換 PID)
taskkill /PID <進程ID> /F
```

或修改 docker-compose.yml 中的端口映射。

### **問題 3: 容器無法啟動**

**症狀:** `docker-compose ps` 顯示 "Exit" 狀態

**解決方案:**
```powershell
# 查看詳細錯誤日誌
docker-compose logs <service-name>

# 重新構建並啟動
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### **問題 4: 前端無法連接後端**

**症狀:** 前端顯示 API 連接錯誤

**解決方案:**
1. 確認後端服務運行: `docker-compose logs backend`
2. 測試後端 API: `curl http://localhost:8080/health`
3. 檢查前端環境變數: `frontend/.env.local`

### **問題 5: 資料庫初始化失敗**

**症狀:** Backend 日誌顯示 "database connection failed"

**解決方案:**
```powershell
# 完全清理並重新啟動
docker-compose down -v
docker-compose up -d

# 等待資料庫完全啟動 (約 30 秒)
Start-Sleep -Seconds 30

# 檢查資料庫日誌
docker-compose logs timescaledb
```

---

## 🚀 一鍵啟動腳本 (改良版)

創建新文件 `start-improved.ps1`:

```powershell
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
    Write-Host "❌ 無法連接到 Docker" -ForegroundColor Red
    Write-Host "請確認 Docker Desktop 正在運行" -ForegroundColor Red
    Read-Host "按 Enter 鍵退出"
    exit 1
}

Write-Host ""
Write-Host "📦 啟動服務..." -ForegroundColor Yellow
docker-compose up -d

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
    Write-Host "⚠️  Backend API: 尚未就緒" -ForegroundColor Yellow
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
Write-Host "==========================================" -ForegroundColor Cyan
Read-Host "按 Enter 鍵退出"
```

**使用方法:**
```powershell
# 右鍵點擊 start-improved.ps1
# 選擇 "使用 PowerShell 運行"
```

---

## 📞 還是無法啟動？

### **最小化測試方案**

嘗試僅啟動資料庫進行測試：

```powershell
# 只啟動資料庫和 Redis
docker-compose up -d timescaledb redis

# 等待啟動
Start-Sleep -Seconds 20

# 檢查狀態
docker-compose ps

# 查看日誌
docker-compose logs timescaledb
```

如果資料庫可以正常啟動，再逐個添加其他服務。

### **聯絡支援**

如果以上方法都無法解決，請提供以下資訊：

1. Docker Desktop 版本
2. Windows 版本
3. 錯誤訊息截圖
4. `docker-compose logs` 的完整輸出

---

**快速診斷命令:**
```powershell
docker --version
docker-compose --version
docker ps
docker-compose ps
docker-compose logs --tail=50
```
