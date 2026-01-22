# 台股智能投資組合管理系統 (PSM)

Portfolio Stock Management System - 專業的台股投資組合管理平台

## 🎯 Phase 1: 核心基礎功能 (MVP)

### ✅ 已完成功能
- ✅ PostgreSQL + TimescaleDB 資料庫架構
- ✅ Go Backend API (Fiber框架)
- ✅ Next.js Frontend (React + TypeScript)
- ✅ 交易記錄輸入 (買入/賣出/股利)
- ✅ FIFO 成本會計
- ✅ 持倉追蹤與損益計算
- ✅ Docker容器化部署
- ✅ 台股手續費與稅金自動計算

## 🏗️ 系統架構

```
┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐
│   Frontend (React)  │◄───│  API Gateway (Go)    │◄───│  Database           │
│   Next.js + Tailwind│    │  Fiber Framework     │    │  TimescaleDB        │
└─────────────────────┘    └──────────────────────┘    └─────────────────────┘
```

### 技術堆疊

**Frontend:**
- Next.js 14 (App Router)
- React 18
- TypeScript
- Tailwind CSS
- Axios

**Backend:**
- Go 1.21
- Fiber Web Framework
- PostgreSQL Driver
- UUID & Decimal 處理

**Database:**
- PostgreSQL 15
- TimescaleDB Extension
- Redis 7 (快取層)

## 🚀 快速開始

### 前置需求
- Docker & Docker Compose
- Node.js 20+ (本地開發)
- Go 1.21+ (本地開發)

### 使用 Docker Compose 啟動 (推薦)

```bash
# 1. Clone 專案
git clone <repository-url>
cd PSM

# 2. 啟動所有服務
docker-compose up -d

# 3. 查看服務狀態
docker-compose ps

# 4. 查看日誌
docker-compose logs -f
```

服務端口:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Database: localhost:5432
- Redis: localhost:6379

### 本地開發模式

#### 啟動 Database

```bash
docker-compose up -d timescaledb redis
```

#### 啟動 Backend

```bash
cd backend
cp .env.example .env
go mod download
go run cmd/api/main.go
```

#### 啟動 Frontend

```bash
cd frontend
npm install
npm run dev
```

## 📊 資料庫 Schema

### 核心表格

**ledger_events** - 不可變交易總帳
- 所有交易的完整審計追蹤
- Event Sourcing 模式
- 支援: BUY, SELL, DIVIDEND, SPLIT, RIGHTS

**positions_current** - 當前持倉 (Materialized View)
- 從 ledger_events 聚合計算
- FIFO 成本會計
- 自動刷新機制

**tax_lots** - 稅務批次追蹤
- FIFO 成本基礎追蹤
- 已實現/未實現損益計算

**corporate_actions** - 公司行動事件
- 除權息處理
- 股票分割/合併
- 價格調整因子

## 🔌 API 端點

### 交易管理
- `POST /api/v1/events` - 新增交易
- `GET /api/v1/portfolios/:id/events` - 查詢交易記錄
- `GET /api/v1/portfolios/:id/events/:symbol` - 查詢特定股票交易

### 持倉管理
- `GET /api/v1/portfolios/:id/positions` - 查詢所有持倉
- `GET /api/v1/portfolios/:id/positions/:symbol` - 查詢特定持倉
- `GET /api/v1/portfolios/:id/positions/:symbol/pnl` - 計算未實現損益

### 投資組合
- `GET /api/v1/portfolios/:id` - 查詢投資組合
- `GET /api/v1/portfolios` - 查詢用戶所有投資組合

### 健康檢查
- `GET /health` - 系統健康狀態

## 💡 台股特殊功能

### 自動計算台灣證券交易費用

**買入交易:**
- 手續費: 0.1425% (最低 $20 TWD)

**賣出交易:**
- 手續費: 0.1425% (最低 $20 TWD)
- 證券交易稅: 0.3%

### 股票代號格式
- TSE (台灣證券交易所): `2330.TW`
- TPEx (櫃買中心): `6488.TWO`

## 🧪 測試與驗證

### 測試案例: 台積電 (2330.TW) 交易

```json
{
  "portfolio_id": "00000000-0000-0000-0000-000000000011",
  "event_type": "BUY",
  "symbol": "2330.TW",
  "quantity": "1000",
  "price": "580.00",
  "fee": "0",
  "tax": "0",
  "occurred_at": "2024-01-22T10:30:00Z",
  "notes": "測試買入台積電"
}
```

系統自動計算:
- 交易金額: $580,000
- 手續費: $826.50 (0.1425%)
- 總成本: $580,826.50

## 📁 專案結構

```
PSM/
├── backend/
│   ├── cmd/api/          # 主程式進入點
│   ├── internal/
│   │   ├── database/     # 資料庫連接
│   │   ├── handlers/     # HTTP handlers
│   │   ├── models/       # 資料模型
│   │   └── services/     # 業務邏輯
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── app/          # Next.js App Router
│   │   ├── components/   # React 組件
│   │   ├── lib/          # 工具函數
│   │   └── types/        # TypeScript 類型
│   ├── Dockerfile
│   └── package.json
├── database/
│   └── migrations/       # SQL 遷移腳本
├── docker-compose.yml
└── README.md
```

## 🔒 安全性考量

- SQL 參數化查詢 (防止 SQL Injection)
- CORS 設定
- Input 驗證 (台股代號格式、數值範圍)
- 不可變交易記錄 (Audit Trail)

## 📈 下一階段功能 (Phase 2-5)

### Phase 2: 技術分析 (規劃中)
- [ ] TA-Lib 200+ 技術指標
- [ ] TradingView Lightweight Charts
- [ ] 多時間框架分析

### Phase 3: 即時數據 (規劃中)
- [ ] WebSocket 即時價格
- [ ] 台股交易時間限制
- [ ] 漲跌停視覺化

### Phase 4: AI 分析 (規劃中)
- [ ] 鉅亨網新聞爬取
- [ ] 中文情感分析
- [ ] GPT-4 投資建議

### Phase 5: 優化 (規劃中)
- [ ] 效能優化
- [ ] 行動響應式設計
- [ ] 資料匯出功能

## 🐛 疑難排解

### Database 連接失敗
```bash
# 檢查 TimescaleDB 是否運行
docker-compose ps timescaledb

# 查看資料庫日誌
docker-compose logs timescaledb
```

### Frontend 無法連接 Backend
檢查 `.env.local` 設定:
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 清空所有資料重新開始
```bash
docker-compose down -v
docker-compose up -d
```

## 📝 License

MIT License

## 👥 貢獻者

Developed with ❤️ for Taiwan Stock Market Investors

---

**Phase 1 Status:** ✅ 完成  
**Last Updated:** 2024-01-22
