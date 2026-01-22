# 台股智能投資組合管理系統 (PSM)

[![Phase](https://img.shields.io/badge/Phase-2%20Complete-success?style=flat-square)](https://github.com/HenryLau1103/PMS)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-14-000000?style=flat-square&logo=next.js)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org/)
[![TimescaleDB](https://img.shields.io/badge/TimescaleDB-PostgreSQL%2015-FDB515?style=flat-square&logo=timescale)](https://www.timescale.com/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Taiwan Stock](https://img.shields.io/badge/Taiwan%20Stock-1943%20stocks-red?style=flat-square)](https://www.twse.com.tw/)

Portfolio Stock Management System - 專業的台股投資組合管理平台

## 🎯 Phase 1: 核心基礎功能 (MVP) ✅

### 已完成功能
- ✅ PostgreSQL + TimescaleDB 資料庫架構
- ✅ Go Backend API (Fiber框架)
- ✅ Next.js Frontend (React + TypeScript)
- ✅ 交易記錄輸入 (買入/賣出/股利)
- ✅ FIFO 成本會計
- ✅ 持倉追蹤與損益計算
- ✅ Docker容器化部署
- ✅ 台股手續費與稅金自動計算

## 📈 Phase 2: 技術分析與市場數據 ✅

### Phase 2.1: 市場數據基礎設施 ✅
- ✅ TimescaleDB Hypertable 時間序列架構
- ✅ TWSE/TPEx API 整合 (1,943支台股)
- ✅ 連續聚合 (Daily/Weekly/Monthly)
- ✅ OHLCV 數據API
- ✅ 自動同步機制

### Phase 2.2: 技術分析引擎 ✅
- ✅ TA-Lib 整合 (markcheno/go-talib)
- ✅ 5大核心指標:
  - MA (移動平均線)
  - RSI (相對強弱指標)
  - MACD (指數平滑移動平均線)
  - Bollinger Bands (布林通道)
  - KDJ (隨機指標)
- ✅ Redis 快取層 (24小時TTL)
- ✅ 批次指標查詢API

### Phase 2.3: TradingView 圖表前端 ✅
- ✅ TradingView Lightweight Charts v4.1.3
- ✅ 蠟燭圖 + 成交量顯示
- ✅ 深色主題 (專業配色)
- ✅ 多指標疊加顯示
- ✅ 指標參數動態調整
- ✅ 響應式圖表設計

### Phase 2.4: 主頁整合 ✅
- ✅ 技術圖表整合到儀表板
- ✅ 自動填入最新收盤價
- ✅ 優化版面配置
- ✅ 股票代號快速切換

### Phase 2.5: 批量數據同步 ✅
- ✅ 批量同步1,943支台股
- ✅ 即時進度追蹤
- ✅ 優先同步持倉股票
- ✅ 失敗重試與錯誤追蹤
- ✅ 速率限制 (符合TWSE規範)
- ✅ 背景異步處理
- ✅ 最近2年歷史數據 (2024-2026)

## 🏗️ 系統架構

```
┌─────────────────────┐    ┌──────────────────────┐    ┌─────────────────────┐
│   Frontend (React)  │◄───│  API Gateway (Go)    │◄───│  TimescaleDB        │
│   Next.js + Chart   │    │  Fiber + TA-Lib      │    │  Hypertables        │
│   TradingView       │    │  Redis Cache         │    │  Aggregates         │
└─────────────────────┘    └──────────────────────┘    └─────────────────────┘
```

### 技術堆疊

**Frontend:**
- Next.js 14 (App Router)
- React 18
- TypeScript
- Tailwind CSS
- TradingView Lightweight Charts v4.1.3
- Axios

**Backend:**
- Go 1.21
- Fiber Web Framework
- TA-Lib (Technical Analysis Library)
- PostgreSQL Driver
- Redis Client
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
git clone https://github.com/HenryLau1103/PMS.git
cd PSM

# 2. 啟動所有服務
docker-compose up -d

# 3. 查看服務狀態
docker-compose ps

# 4. 查看日誌
docker-compose logs -f
```

服務端口:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Database**: localhost:5432
- **Redis**: localhost:6379

### 首次使用 - 同步市場數據

1. 打開 http://localhost:3000
2. 在右側「市場數據同步」面板
3. 點擊「同步所有股票」
4. 等待同步完成 (~1.6小時)

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

**stock_ohlcv** - OHLCV時間序列 (Hypertable)
- 每日開高低收成交量數據
- TimescaleDB 壓縮與分區
- 連續聚合支援

**technical_indicators** - 技術指標快取
- 計算結果快取
- 定期更新機制

## 🔌 API 端點

### 交易管理
- `POST /api/v1/events` - 新增交易
- `GET /api/v1/portfolios/:id/events` - 查詢交易記錄
- `GET /api/v1/portfolios/:id/events/:symbol` - 查詢特定股票交易

### 持倉管理
- `GET /api/v1/portfolios/:id/positions` - 查詢所有持倉
- `GET /api/v1/portfolios/:id/positions/:symbol` - 查詢特定持倉
- `GET /api/v1/portfolios/:id/positions/:symbol/pnl` - 計算未實現損益

### 市場數據
- `GET /api/v1/stocks/:symbol/ohlcv` - 查詢OHLCV數據
- `POST /api/v1/market/sync` - 單一股票同步
- `POST /api/v1/market/bulk-sync/start` - 批量同步
- `GET /api/v1/market/bulk-sync/status` - 同步進度
- `POST /api/v1/market/bulk-sync/stop` - 停止同步

### 技術指標
- `GET /api/v1/indicators/:symbol/ma` - 移動平均線
- `GET /api/v1/indicators/:symbol/rsi` - RSI指標
- `GET /api/v1/indicators/:symbol/macd` - MACD指標
- `GET /api/v1/indicators/:symbol/bb` - 布林通道
- `GET /api/v1/indicators/:symbol/kdj` - KDJ指標
- `POST /api/v1/indicators/:symbol/batch` - 批次查詢

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
- TSE (台灣證券交易所): `2330`, `2454`
- TPEx (櫃買中心): `6488`, `5347`

## 🎨 功能展示

### 主儀表板
- 交易表單 (左側)
- 投資組合概覽 (右上)
- 市場數據同步面板 (右下)
- 技術分析圖表 (底部)

### 技術分析圖表
- K線圖 + 成交量
- 動態指標切換
- 參數即時調整
- 多時間框架支援

### 批量同步功能
- 即時進度條
- 成功/失敗統計
- 失敗股票列表
- 用時追蹤

## 📁 專案結構

```
PSM/
├── backend/
│   ├── cmd/api/              # 主程式進入點
│   ├── internal/
│   │   ├── database/         # 資料庫連接
│   │   ├── handlers/         # HTTP handlers
│   │   │   ├── bulk_sync_handler.go
│   │   │   ├── indicator_handler.go
│   │   │   └── market_data_handler.go
│   │   ├── models/           # 資料模型
│   │   └── services/         # 業務邏輯
│   │       ├── market_data_service.go
│   │       └── technical_analysis_service.go
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── app/              # Next.js App Router
│   │   │   ├── page.tsx      # 主儀表板
│   │   │   └── analysis/     # 技術分析頁
│   │   ├── components/       # React 組件
│   │   │   ├── Chart/        # 圖表組件
│   │   │   │   ├── StockChart.tsx
│   │   │   │   └── ChartControls.tsx
│   │   │   ├── DataSyncPanel.tsx
│   │   │   └── PortfolioDashboard.tsx
│   │   ├── lib/              # 工具函數
│   │   │   └── chartApi.ts
│   │   └── types/            # TypeScript 類型
│   ├── Dockerfile
│   └── package.json
├── database/
│   └── migrations/           # SQL 遷移腳本
│       ├── 001_init.sql
│       ├── 002_taiwan_stocks.sql
│       └── 003_market_data.sql
├── docker-compose.yml
└── README.md
```

## 🔒 安全性考量

- SQL 參數化查詢 (防止 SQL Injection)
- CORS 設定
- Input 驗證 (台股代號格式、數值範圍)
- 不可變交易記錄 (Audit Trail)
- Redis 快取過期機制
- API 速率限制 (TWSE規範)

## 📈 開發路線圖

### Phase 3: 即時數據 (計劃中)
- [ ] WebSocket 即時價格推送
- [ ] 台股交易時間限制
- [ ] 漲跌停視覺化
- [ ] 盤中五檔報價
- [ ] 個股成交明細

### Phase 4: AI 分析 (計劃中)
- [ ] 鉅亨網新聞爬取
- [ ] 中文情感分析
- [ ] GPT-4 投資建議
- [ ] 異常交易偵測
- [ ] 智能選股推薦

### Phase 5: 優化與擴展 (計劃中)
- [ ] 效能優化
- [ ] 行動響應式設計
- [ ] 資料匯出功能 (CSV/Excel)
- [ ] 多帳戶管理
- [ ] 權限控制系統
- [ ] 回測系統

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

### 圖表無數據顯示
1. 檢查是否已同步市場數據
2. 查看同步進度: 訪問主頁右下角「市場數據同步」面板
3. 手動觸發同步: 點擊「同步所有股票」按鈕

### 清空所有資料重新開始
```bash
docker-compose down -v
docker-compose up -d
```

## 🧪 測試

### 測試OHLCV API
```bash
curl "http://localhost:8080/api/v1/stocks/2330/ohlcv?limit=10"
```

### 測試技術指標API
```bash
curl "http://localhost:8080/api/v1/indicators/2330/ma?period=20"
```

### 測試同步狀態
```bash
curl "http://localhost:8080/api/v1/market/bulk-sync/status"
```

## 📝 License

MIT License

## 👥 貢獻者

Developed with ❤️ for Taiwan Stock Market Investors

---

**Phase 1 Status:** ✅ 完成 (2024-01-22)  
**Phase 2 Status:** ✅ 完成 (2026-01-22)  
**Last Updated:** 2026-01-22
