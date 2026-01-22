# Quick Start Guide - PSM System

## Method 1: Simple Batch File (Recommended)

1. Double-click: **`start-simple.bat`**
2. Wait for startup to complete
3. Open browser: http://localhost:3000

---

## Method 2: Manual Commands (PowerShell)

Open PowerShell and run:

```powershell
# Navigate to project directory
cd "C:\Users\Henry\OneDrive\桌面\PSM"

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

---

## Method 3: Manual Commands (Command Prompt)

Open CMD and run:

```cmd
cd C:\Users\Henry\OneDrive\桌面\PSM
docker-compose up -d
docker-compose ps
```

---

## Before Starting - Check Docker

1. **Open Docker Desktop** application
2. Wait for bottom status to show **"Engine running"** (green icon)
3. This may take 30-60 seconds

---

## Verify Services

After starting, check these URLs:

- **Frontend**: http://localhost:3000
- **Backend Health Check**: http://localhost:8080/health
- Should see: `{"status":"healthy"}`

---

## Stop Services

```powershell
docker-compose down
```

---

## Reset Everything (Clean Start)

```powershell
docker-compose down -v
docker-compose up -d
```

---

## Troubleshooting

### Problem: Docker connection error

**Solution:**
1. Close Docker Desktop completely
2. Restart Docker Desktop
3. Wait for "Engine running" status
4. Try again

### Problem: Port already in use

**Solution:**
```powershell
# Check what's using port 3000
netstat -ano | findstr :3000

# Check what's using port 8080
netstat -ano | findstr :8080

# Kill process (replace PID)
taskkill /PID <process_id> /F
```

### Problem: Services not starting

**Solution:**
```powershell
# View detailed logs
docker-compose logs

# Restart specific service
docker-compose restart backend
docker-compose restart frontend
```

---

## Quick Test

Test if system is working:

1. Open http://localhost:3000
2. Fill in transaction form:
   - Type: BUY
   - Symbol: 2330.TW
   - Quantity: 1000
   - Price: 580
3. Submit
4. Right side should show position with 1000 shares

---

## Files You Can Use

- **start-simple.bat** - Simple startup (no Chinese characters)
- **start-en.ps1** - English PowerShell script
- **docker-compose.yml** - Direct docker-compose usage

Choose whichever works best for you!

---

## Need Help?

See **TROUBLESHOOTING.md** for detailed solutions.
