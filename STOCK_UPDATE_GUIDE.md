# 台股股票清單更新指南

## 📊 目前資料來源

股票清單是手動整理的常見台股，目前包含：
- 52 檔常見上市櫃股票
- 涵蓋：半導體、金融、傳產、電信、航運、電子等產業

## 🔄 如何新增股票

### 方法 1: 直接使用 SQL（推薦）

```bash
# 單筆新增
docker-compose exec -T timescaledb psql -U psm_user -d portfolio_db << EOF
INSERT INTO taiwan_stocks (symbol, name, name_en, market, industry) VALUES
('股票代號', '中文名稱', 'English Name', '市場', '產業');
EOF

# 範例：新增長興化工
docker-compose exec -T timescaledb psql -U psm_user -d portfolio_db << EOF
INSERT INTO taiwan_stocks (symbol, name, name_en, market, industry) VALUES
('1717', '長興', 'Chang Chun Plastics', 'TSE', '化學');
EOF
```

### 方法 2: 批次新增多檔股票

建立檔案 `add_stocks.sql`:
```sql
INSERT INTO taiwan_stocks (symbol, name, name_en, market, industry) VALUES
('1234', '股票A', 'Stock A', 'TSE', '產業類別'),
('5678', '股票B', 'Stock B', 'OTC', '產業類別'),
('9012', '股票C', 'Stock C', 'TSE', '產業類別');
```

執行：
```bash
cat add_stocks.sql | docker-compose exec -T timescaledb psql -U psm_user -d portfolio_db
```

## 📝 欄位說明

| 欄位 | 說明 | 範例 | 必填 |
|------|------|------|------|
| symbol | 股票代號（4位數字） | `2330` | ✅ |
| name | 中文名稱 | `台積電` | ✅ |
| name_en | 英文名稱 | `TSMC` | ❌ |
| market | 市場別 | `TSE`(上市) / `OTC`(上櫃) | ✅ |
| industry | 產業類別 | `半導體` | ❌ |

## 🏷️ 常見產業分類

- `半導體` - 晶圓代工、IC設計
- `金融保險` - 銀行、證券、保險
- `電子零組件` - 被動元件、連接器
- `電腦週邊` - 主機板、筆電
- `通信網路` - 電信、網通設備
- `光電` - 面板、LED
- `化學` - 塑化、特化
- `鋼鐵` - 鋼材、鋼鐵製品
- `零售` - 百貨、超商
- `食品` - 食品加工、飲料
- `航運` - 貨櫃、散裝
- `水泥` - 水泥製品
- `油電燃氣` - 能源、瓦斯
- `生技醫療` - 製藥、醫材

## 🔍 查詢與驗證

### 查看所有股票
```bash
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "SELECT symbol, name, market, industry FROM taiwan_stocks ORDER BY symbol;"
```

### 搜尋特定股票
```bash
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "SELECT * FROM taiwan_stocks WHERE symbol = '2330';"
```

### 查詢特定產業
```bash
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "SELECT symbol, name FROM taiwan_stocks WHERE industry = '半導體';"
```

### 統計股票數量
```bash
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "SELECT COUNT(*) as total, market FROM taiwan_stocks GROUP BY market;"
```

## ✏️ 修改股票資料

### 更新股票名稱
```bash
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "UPDATE taiwan_stocks SET name = '新名稱' WHERE symbol = '2330';"
```

### 更新產業分類
```bash
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "UPDATE taiwan_stocks SET industry = '新產業' WHERE symbol = '2330';"
```

## 🗑️ 刪除股票

### 停用股票（推薦）
```bash
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "UPDATE taiwan_stocks SET is_active = false WHERE symbol = '2330';"
```

### 永久刪除
```bash
docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "DELETE FROM taiwan_stocks WHERE symbol = '2330';"
```

## 🚀 未來擴充方案

### 方案 1: 爬蟲自動更新（推薦）
從公開資料源定期更新：
- 台灣證券交易所 API
- 證券櫃檯買賣中心 API
- Yahoo Finance Taiwan

### 方案 2: CSV 批次匯入
```bash
# 準備 stocks.csv 檔案格式：
# symbol,name,name_en,market,industry
# 2330,台積電,TSMC,TSE,半導體
# 2317,鴻海,Hon Hai,TSE,電子

docker-compose exec timescaledb psql -U psm_user -d portfolio_db -c "\COPY taiwan_stocks(symbol, name, name_en, market, industry) FROM '/path/to/stocks.csv' WITH CSV HEADER;"
```

### 方案 3: Admin API 端點
建立管理後台允許手動新增/編輯股票（需實作認證）

## 📋 完整台股清單資源

如需匯入完整台股清單（1700+ 檔），可從以下來源取得：
1. [台灣證券交易所 - 上市公司資料](https://www.twse.com.tw/)
2. [證券櫃檯買賣中心 - 上櫃公司資料](https://www.tpex.org.tw/)
3. [Goodinfo 台灣股市資訊網](https://goodinfo.tw/)

## ⚠️ 注意事項

1. **資料庫持久化**: 資料儲存在 Docker volume `timescale_data`，重啟容器不會遺失
2. **重建容器**: 若執行 `docker-compose down -v` 會清空所有資料，需重新匯入
3. **字元編碼**: 確保中文名稱使用 UTF-8 編碼
4. **唯一性**: `symbol` 為主鍵，不可重複

## 🛠️ 備份與還原

### 備份股票清單
```bash
docker-compose exec timescaledb pg_dump -U psm_user -d portfolio_db -t taiwan_stocks > stocks_backup.sql
```

### 還原股票清單
```bash
cat stocks_backup.sql | docker-compose exec -T timescaledb psql -U psm_user -d portfolio_db
```
