# Phase 1 驗證測試指南

## 🎯 驗證目標
確認 Phase 1 核心功能正常運作:
1. 交易記錄輸入
2. FIFO 成本計算
3. 持倉追蹤
4. 損益計算

## 🚀 啟動系統

```bash
# 啟動所有服務
docker-compose up -d

# 等待服務就緒 (約30秒)
docker-compose ps

# 檢查健康狀態
curl http://localhost:8080/health
```

## ✅ 測試案例

### Test Case 1: 買入台積電

**操作步驟:**
1. 開啟 http://localhost:3000
2. 填寫交易表單:
   - 交易類型: 買入
   - 股票代號: 2330.TW
   - 數量: 1000
   - 價格: 580.00
   - 手續費: 0 (自動計算)
   - 稅金: 0
3. 提交表單

**預期結果:**
- ✅ 交易成功新增
- ✅ 持倉表顯示 2330.TW
- ✅ 持有股數: 1,000
- ✅ 平均成本: $580.83 (含手續費)
- ✅ 總成本: $580,827

### Test Case 2: 追加買入

**操作步驟:**
1. 再次買入 2330.TW
   - 數量: 500
   - 價格: 590.00
2. 提交表單

**預期結果:**
- ✅ 持有股數更新為: 1,500
- ✅ 平均成本重新計算 (FIFO加權平均)
- ✅ 總成本正確累加

### Test Case 3: 部分賣出

**操作步驟:**
1. 賣出 2330.TW
   - 數量: 300
   - 價格: 600.00
   - 手續費: 0 (自動計算)
   - 稅金: 0 (自動計算 0.3%)
2. 提交表單

**預期結果:**
- ✅ 持有股數更新為: 1,200
- ✅ 手續費自動計算: $256.95
- ✅ 證券交易稅自動計算: $540.00
- ✅ 平均成本保持不變 (FIFO)

### Test Case 4: 多檔股票

**操作步驟:**
1. 買入鴻海 (2317.TW)
   - 數量: 2000
   - 價格: 105.50
2. 買入聯發科 (2454.TW)
   - 數量: 100
   - 價格: 1200.00

**預期結果:**
- ✅ 持倉表顯示 3 檔股票
- ✅ 每檔股票獨立計算成本
- ✅ 總成本正確累計

## 🔍 API 測試

### 使用 curl 測試 API

```bash
# 1. 健康檢查
curl http://localhost:8080/health

# 2. 查詢投資組合
curl http://localhost:8080/api/v1/portfolios/00000000-0000-0000-0000-000000000011

# 3. 新增交易
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "portfolio_id": "00000000-0000-0000-0000-000000000011",
    "event_type": "BUY",
    "symbol": "2330.TW",
    "quantity": "1000",
    "price": "580.00",
    "fee": "826.50",
    "tax": "0",
    "occurred_at": "2024-01-22T10:30:00Z"
  }'

# 4. 查詢持倉
curl http://localhost:8080/api/v1/portfolios/00000000-0000-0000-0000-000000000011/positions

# 5. 查詢交易記錄
curl http://localhost:8080/api/v1/portfolios/00000000-0000-0000-0000-000000000011/events
```

## 📊 資料庫驗證

```bash
# 連接到資料庫
docker-compose exec timescaledb psql -U psm_user -d portfolio_db

# 查詢所有交易
SELECT event_id, event_type, symbol, quantity, price, total_amount, occurred_at
FROM ledger_events
ORDER BY occurred_at DESC;

# 查詢當前持倉
SELECT * FROM positions_current;

# 查詢投資組合
SELECT * FROM portfolios;

# 退出
\q
```

## 🎉 驗證成功標準

Phase 1 功能驗證通過需滿足:

- [x] ✅ 可新增買入交易
- [x] ✅ 可新增賣出交易
- [x] ✅ 台股手續費自動計算正確
- [x] ✅ 證券交易稅自動計算正確
- [x] ✅ FIFO 成本計算正確
- [x] ✅ 持倉數量即時更新
- [x] ✅ 平均成本正確計算
- [x] ✅ 支援多檔股票同時持有
- [x] ✅ 交易記錄不可變 (Audit Trail)
- [x] ✅ 資料庫關聯正確

## 🐛 已知限制 (Phase 1)

以下功能將在後續 Phase 實作:

- ❌ 即時股價更新 (Phase 3)
- ❌ 技術指標計算 (Phase 2)
- ❌ AI 投資建議 (Phase 4)
- ❌ 圖表視覺化 (Phase 2)
- ❌ 除權息處理 (Phase 2)
- ❌ 使用者認證 (Phase 2)

## 📝 測試報告範本

```
=== PSM Phase 1 驗證報告 ===

測試日期: YYYY-MM-DD
測試人員: [Your Name]

✅ Test Case 1: 買入交易 - PASS
✅ Test Case 2: 追加買入 - PASS
✅ Test Case 3: 賣出交易 - PASS
✅ Test Case 4: 多檔股票 - PASS

問題回報:
[無 / 描述問題]

結論:
Phase 1 核心功能驗證 [通過 / 失敗]
```

---

**驗證完成後，即可進入 Phase 2 開發！** 🚀
