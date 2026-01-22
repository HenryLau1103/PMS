# PSM Startup Script (PowerShell)
# Encoding: UTF-8

Write-Host "Starting PSM Taiwan Stock Portfolio Management System" -ForegroundColor Cyan
Write-Host "======================================================" -ForegroundColor Cyan
Write-Host ""

# Check Docker
Write-Host "Checking Docker status..." -ForegroundColor Yellow
try {
    $dockerVersion = docker --version
    Write-Host "Docker installed: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Docker not installed or not running" -ForegroundColor Red
    Write-Host "Please install and start Docker Desktop first" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Test Docker connection
Write-Host "Testing Docker connection..." -ForegroundColor Yellow
try {
    docker ps | Out-Null
    Write-Host "Docker connection OK" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Cannot connect to Docker Engine" -ForegroundColor Red
    Write-Host "" -ForegroundColor Red
    Write-Host "Please follow these steps:" -ForegroundColor Yellow
    Write-Host "1. Open Docker Desktop application" -ForegroundColor White
    Write-Host "2. Wait for status 'Engine running' (green icon)" -ForegroundColor White
    Write-Host "3. If Docker Desktop fails to start, restart your computer" -ForegroundColor White
    Write-Host "" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""
Write-Host "Starting services..." -ForegroundColor Yellow
docker-compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Service startup failed" -ForegroundColor Red
    Write-Host "View error logs: docker-compose logs" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""
Write-Host "Waiting for services to be ready (30 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

Write-Host ""
Write-Host "Checking service status..." -ForegroundColor Yellow
docker-compose ps

Write-Host ""
Write-Host "Health check..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Host "Backend API: Running normally" -ForegroundColor Green
    }
} catch {
    Write-Host "Backend API: Not fully ready yet, please try again later" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "======================================================" -ForegroundColor Cyan
Write-Host "PSM System startup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Access services:" -ForegroundColor White
Write-Host "  Frontend:    http://localhost:3000" -ForegroundColor Cyan
Write-Host "  Backend API: http://localhost:8080" -ForegroundColor Cyan
Write-Host "  Database:    localhost:5432" -ForegroundColor Cyan
Write-Host ""
Write-Host "Useful commands:" -ForegroundColor White
Write-Host "  View logs:    docker-compose logs -f" -ForegroundColor Gray
Write-Host "  Stop services: docker-compose down" -ForegroundColor Gray
Write-Host "  Restart:      docker-compose restart" -ForegroundColor Gray
Write-Host ""
Write-Host "For troubleshooting, see: TROUBLESHOOTING.md" -ForegroundColor Yellow
Write-Host "======================================================" -ForegroundColor Cyan
Write-Host ""
Read-Host "Press Enter to exit"
