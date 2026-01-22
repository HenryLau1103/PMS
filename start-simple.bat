@echo off
chcp 65001 >nul
cls

echo ========================================
echo PSM System Startup
echo ========================================
echo.

echo [1/4] Checking Docker...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker not found
    echo Please install Docker Desktop first
    pause
    exit /b 1
)
echo OK: Docker installed

echo.
echo [2/4] Testing Docker connection...
docker ps >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Cannot connect to Docker
    echo.
    echo Please:
    echo 1. Open Docker Desktop
    echo 2. Wait for "Engine running" status
    echo 3. Run this script again
    pause
    exit /b 1
)
echo OK: Docker is running

echo.
echo [3/4] Starting services...
docker-compose up -d
if %errorlevel% neq 0 (
    echo ERROR: Failed to start services
    pause
    exit /b 1
)

echo.
echo [4/4] Waiting for services (30 seconds)...
timeout /t 30 /nobreak >nul

echo.
echo ========================================
echo Service Status:
echo ========================================
docker-compose ps

echo.
echo ========================================
echo PSM System Started!
echo ========================================
echo.
echo Open in browser:
echo   Frontend:  http://localhost:3000
echo   API:       http://localhost:8080
echo.
echo Commands:
echo   View logs: docker-compose logs -f
echo   Stop:      docker-compose down
echo.
pause
