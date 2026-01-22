# PSM 系統狀態

## 當前狀態

✅ **Backend**: 正常運行
- Health: http://localhost:8080/health → {"status":"healthy"}
- API: http://localhost:8080/api/v1/portfolios → 返回數據正常

✅ **Frontend**: 服務運行中
- Port: 3000
- Build: 成功
- Start: 正常

⚠️ **問題**: 瀏覽器顯示 "Application error: a client-side exception has occurred"

## 可能原因

這個錯誤通常由以下原因引起：

1. **環境變數問題** - NEXT_PUBLIC_API_URL 需要在構建時設置
2. **API連接問題** - 瀏覽器無法訪問 http://localhost:8080
3. **React組件錯誤** - 客戶端渲染時出錯

## 解決方案

### 方案 1: 使用開發模式 (推薦用於測試)

不使用Docker，本地運行：

```bash
# 1. 保持 Docker 的 Database 和 Backend 運行
docker-compose up -d timescaledb redis backend

# 2. 本地運行 Frontend (開發模式)
cd frontend
npm install
npm run dev
```

然後訪問: http://localhost:3000

### 方案 2: 檢查瀏覽器控制台

1. 打開 http://localhost:3000
2. 按 F12 打開開發者工具
3. 查看 Console 標籤中的錯誤訊息
4. 將錯誤訊息告訴我

### 方案 3: 查看 Docker 日誌

```bash
# 查看 frontend 詳細日誌
docker-compose logs frontend --tail=100

# 查看 backend 日誌
docker-compose logs backend --tail=50
```

## 快速測試 - 本地開發模式

```bash
# 停止 frontend 容器
docker-compose stop frontend

# 本地運行
cd frontend
npm run dev
```

這樣可以看到更詳細的錯誤信息。

## 當前建議

**立即嘗試本地開發模式：**

```bash
# 在 PSM 目錄執行
docker-compose stop frontend
cd frontend
npm run dev
```

然後訪問 http://localhost:3000 並告訴我是否有錯誤訊息。
