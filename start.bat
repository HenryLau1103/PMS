@echo off
echo 🚀 啟動 PSM 台股智能投資組合管理系統
echo ==========================================
echo.

REM Check if Docker is running
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Error: Docker 未運行，請先啟動 Docker Desktop
    pause
    exit /b 1
)

echo ✅ Docker 已就緒
echo.

REM Check if docker-compose.yml exists
if not exist "docker-compose.yml" (
    echo ❌ Error: docker-compose.yml 不存在
    pause
    exit /b 1
)

echo 📦 啟動服務...
docker-compose up -d

echo.
echo ⏳ 等待服務啟動 (約 30 秒)...
timeout /t 30 /nobreak >nul

echo.
echo 🔍 檢查服務狀態...
docker-compose ps

echo.
echo 🏥 健康檢查...
curl -s http://localhost:8080/health 2>nul | findstr /C:"healthy" >nul
if %errorlevel% equ 0 (
    echo ✅ Backend API 健康狀態: 正常
) else (
    echo ⚠️  Backend API 可能尚未完全啟動，請稍候再試
)

echo.
echo ==========================================
echo ✅ PSM 系統啟動完成！
echo.
echo 訪問以下服務:
echo   🌐 Frontend:  http://localhost:3000
echo   🔌 Backend API: http://localhost:8080
echo   🗄️  Database:   localhost:5432
echo.
echo 查看日誌:
echo   docker-compose logs -f
echo.
echo 停止服務:
echo   docker-compose down
echo.
echo ==========================================
pause
