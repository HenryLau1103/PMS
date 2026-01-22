# 台股股票清單同步工具

## 🚀 快速使用（推薦）

### 方法 1: 使用 API 端點同步（需 Docker 運行）

```bash
# 觸發同步
curl -X POST http://localhost:8080/api/v1/stocks/sync

# 範例回應：
# {
#   "success": true,
#   "message": "Stock synchronization completed",
#   "result": {
#     "tse": 1075,
#     "otc": 782,
#     "total": 1857
#   }
# }
```

### 方法 2: 使用 CLI 同步工具（Docker 停止時）

```bash
# 進入 backend 目錄
cd backend

# 執行同步工具
go run cmd/sync-stocks/main.go

# 輸出範例：
# 🔄 Starting stock synchronization from TWSE/TPEx Open API...
# 
# ✅ Synchronization completed successfully!
# 
# 📊 Results:
#    - TSE (上市): 1075 stocks
#    - OTC (上櫃): 782 stocks
#    - Total:      1857 stocks
# 
# ✨ Stock database is now up to date!
```

### 方法 3: 使用 Docker 執行同步

```bash
# Docker 容器內執行
docker-compose exec backend go run cmd/sync-stocks/main.go
```

## 📊 資料來源

### 台灣證券交易所 (TWSE) - 上市股票
- API: `https://openapi.twse.com.tw/v1/opendata/t187ap03_L`
- 約 1,075 檔股票
- 包含完整的公司資訊、產業分類

### 櫃買中心 (TPEx) - 上櫃股票
- API: `https://www.tpex.org.tw/openapi/v1/tpex_mainboard_peratio_analysis`
- 約 782 檔股票
- 自動過濾 ETF 與債券

## 🏷️ 自動產業分類

系統會自動將 TWSE 提供的產業代碼轉換為中文名稱：

| 代碼 | 產業名稱 |
|------|----------|
| 01   | 水泥 |
| 02   | 食品 |
| 03   | 塑膠 |
| 17   | 金融保險 |
| 21   | 化學生技醫療 |
| 22   | 半導體 |
| 23   | 電腦週邊 |
| 24   | 光電 |
| 25   | 通信網路 |
| 26   | 電子零組件 |
| ...  | (完整對應表見 stock_sync_service.go) |

## 🔄 同步策略

### Upsert 機制
- 若股票代號不存在 → 新增
- 若股票代號已存在 → 更新名稱、產業、更新時間
- 保留 `is_active` 狀態不被覆蓋

### 同步內容
- ✅ 股票代號 (symbol)
- ✅ 公司簡稱 (name)
- ✅ 市場別 (market): TSE / OTC
- ✅ 產業分類 (industry)
- ✅ 更新時間 (updated_at)

### 不覆蓋項目
- `is_active` - 手動停用的股票不會被重新啟用
- `name_en` - 保留手動輸入的英文名稱

## 📝 同步後驗證

```bash
# 查詢總數
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "SELECT market, COUNT(*) FROM taiwan_stocks GROUP BY market;"

# 輸出範例：
#  market | count 
# --------+-------
#  TSE    |  1075
#  OTC    |   782

# 查詢特定股票
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "SELECT * FROM taiwan_stocks WHERE symbol = '2330';"
```

## ⏰ 自動排程（選用）

### 方法 1: Linux Cron Job
```bash
# 編輯 crontab
crontab -e

# 每天凌晨 2 點同步
0 2 * * * curl -X POST http://localhost:8080/api/v1/stocks/sync
```

### 方法 2: Windows 工作排程器
1. 開啟「工作排程器」
2. 建立基本工作
3. 觸發程序：每日
4. 動作：啟動程式
5. 程式：`curl`
6. 引數：`-X POST http://localhost:8080/api/v1/stocks/sync`

### 方法 3: Docker 容器內 (未來實作)
建立背景 worker 定期執行同步

## 🚨 注意事項

### API 限制
- TWSE/TPEx API 無明確限流說明
- 建議不要過於頻繁呼叫（每日一次即可）
- 若遇到 HTTP 429，請延後再試

### 錯誤處理
- 個別股票同步失敗不會中斷整體流程
- 錯誤會記錄在 console 但繼續處理
- 最終回傳成功同步的數量

### 資料一致性
- 使用 UPSERT 確保冪等性
- 可安全重複執行，不會產生重複資料
- 已刪除的上市公司會保留在資料庫（`is_active=false`）

## 🛠️ 故障排除

### 問題：API 回傳 404/500
**解決：**
- 檢查 TWSE/TPEx 網站是否正常
- API 端點可能變更，需更新 URL

### 問題：同步時間過長
**解決：**
- 正常情況約需 10-30 秒
- 若超過 1 分鐘，檢查網路連線

### 問題：部分股票未同步
**解決：**
- 查看 console 日誌找出失敗原因
- 可能是資料格式問題或欄位缺失

## 📈 進階使用

### 僅同步上市股票
修改 `stock_sync_service.go`，註解掉 OTC 部分

### 自訂產業分類
修改 `industryMap` 對應表

### 擴充欄位
API 提供更多欄位（地址、資本額等），可依需求擴充

---

**最後更新：** 2026-01-21  
**資料來源：** TWSE Open API, TPEx Open API
