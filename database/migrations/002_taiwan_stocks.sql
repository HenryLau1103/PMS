-- ================================================
-- Taiwan Stock Symbols Master Table
-- ================================================

CREATE TABLE taiwan_stocks (
    symbol VARCHAR(10) PRIMARY KEY,     -- Stock code without .TW suffix (e.g., '2330')
    name VARCHAR(100) NOT NULL,         -- Chinese name (e.g., '台積電')
    name_en VARCHAR(100),                -- English name (e.g., 'TSMC')
    market VARCHAR(10) NOT NULL,        -- 'TSE' or 'OTC' (上市/上櫃)
    industry VARCHAR(50),                -- Industry category
    is_active BOOLEAN DEFAULT TRUE,     -- Whether still trading
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for fast searching
CREATE INDEX idx_stocks_name ON taiwan_stocks USING gin(to_tsvector('simple', name));
CREATE INDEX idx_stocks_symbol_prefix ON taiwan_stocks(symbol text_pattern_ops);
CREATE INDEX idx_stocks_market ON taiwan_stocks(market);
CREATE INDEX idx_stocks_active ON taiwan_stocks(is_active);

-- ================================================
-- Insert Popular Taiwan Stocks (Top 50 by market cap)
-- ================================================

INSERT INTO taiwan_stocks (symbol, name, name_en, market, industry) VALUES
-- 半導體
('2330', '台積電', 'TSMC', 'TSE', '半導體'),
('2454', '聯發科', 'MediaTek', 'TSE', '半導體'),
('2308', '台達電', 'Delta Electronics', 'TSE', '電子零組件'),
('3034', '聯詠', 'Novatek', 'TSE', '半導體'),
('2303', '聯電', 'UMC', 'TSE', '半導體'),
('3231', '緯創', 'Wistron', 'TSE', '電腦週邊'),
('2379', '瑞昱', 'Realtek', 'TSE', '半導體'),
('3711', '日月光投控', 'ASE Technology Holding', 'TSE', '半導體'),
('2408', '南亞科', 'Nanya Technology', 'TSE', '半導體'),
('2301', '光寶科', 'Lite-On Technology', 'TSE', '電子零組件'),

-- 金融
('2881', '富邦金', 'Fubon Financial', 'TSE', '金融保險'),
('2882', '國泰金', 'Cathay Financial', 'TSE', '金融保險'),
('2884', '玉山金', 'E.Sun Financial', 'TSE', '金融保險'),
('2886', '兆豐金', 'Mega Financial', 'TSE', '金融保險'),
('2891', '中信金', 'CTBC Financial', 'TSE', '金融保險'),
('2885', '元大金', 'Yuanta Financial', 'TSE', '金融保險'),
('2883', '開發金', 'China Development Financial', 'TSE', '金融保險'),
('2892', '第一金', 'First Financial', 'TSE', '金融保險'),
('5880', '合庫金', 'Taiwan Business Bank', 'TSE', '金融保險'),

-- 傳產
('1301', '台塑', 'Formosa Plastics', 'TSE', '塑膠'),
('1303', '南亞', 'Nan Ya Plastics', 'TSE', '塑膠'),
('1326', '台化', 'Formosa Chemicals', 'TSE', '化學'),
('2002', '中鋼', 'China Steel', 'TSE', '鋼鐵'),
('2912', '統一超', '7-Eleven', 'TSE', '零售'),
('1216', '統一', 'Uni-President', 'TSE', '食品'),

-- 電信
('2412', '中華電', 'Chunghwa Telecom', 'TSE', '通信網路'),
('4904', '遠傳', 'Far EasTone', 'TSE', '通信網路'),
('3045', '台灣大', 'Taiwan Mobile', 'TSE', '通信網路'),

-- 航運
('2603', '長榮', 'Evergreen Marine', 'TSE', '航運'),
('2609', '陽明', 'Yang Ming Marine', 'TSE', '航運'),
('2615', '萬海', 'Wan Hai Lines', 'TSE', '航運'),

-- 電子
('2317', '鴻海', 'Hon Hai/Foxconn', 'TSE', '電子'),
('2382', '廣達', 'Quanta Computer', 'TSE', '電腦週邊'),
('2357', '華碩', 'ASUS', 'TSE', '電腦週邊'),
('2353', '宏碁', 'Acer', 'TSE', '電腦週邊'),

-- 生技醫療
('4938', '和碩', 'Pegatron', 'TSE', '電腦週邊'),
('1101', '台泥', 'Taiwan Cement', 'TSE', '水泥'),
('1102', '亞泥', 'Asia Cement', 'TSE', '水泥'),

-- 櫃買中心熱門股
('6505', '台塑化', 'Formosa Petrochemical', 'TSE', '油電燃氣'),
('6415', '矽力-KY', 'Silergy', 'TSE', '半導體'),
('6669', '緯穎', 'Wiwynn', 'TSE', '電腦週邊'),
('3008', '大立光', 'Largan Precision', 'TSE', '光學'),

-- 更多常見股票
('1702', '南僑', 'Nam Chow', 'TSE', '化學'),
('1707', '葡萄王', 'Grape King Bio', 'TSE', '生技醫療'),
('2324', '仁寶', 'Compal Electronics', 'TSE', '電腦週邊'),
('2327', '國巨', 'Yageo', 'TSE', '電子零組件'),
('2345', '智邦', 'Accton Technology', 'TSE', '通信網路'),
('2355', '敬鵬', 'King Board', 'TSE', '電子零組件'),
('2356', '英業達', 'Inventec', 'TSE', '電腦週邊'),
('2395', '研華', 'Advantech', 'TSE', '電腦週邊'),
('2409', '友達', 'AUO', 'TSE', '光電'),
('2474', '可成', 'Catcher Technology', 'TSE', '電子零組件');

COMMENT ON TABLE taiwan_stocks IS '台灣股票代號主檔（支援自動完成功能）';
COMMENT ON COLUMN taiwan_stocks.symbol IS '股票代號（不含.TW後綴）';
COMMENT ON COLUMN taiwan_stocks.name IS '股票中文名稱';
COMMENT ON COLUMN taiwan_stocks.market IS '市場別：TSE=上市, OTC=上櫃';
