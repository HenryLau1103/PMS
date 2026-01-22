-- ============================================================================
-- Phase 2: Technical Analysis - Market Data Schema
-- Migration 003: OHLCV Time-Series Data & Continuous Aggregates
-- ============================================================================

-- ============================================================================
-- 1. Create Hypertable for OHLCV Data (Candlestick Data)
-- ============================================================================

CREATE TABLE IF NOT EXISTS stock_ohlcv (
    symbol VARCHAR(10) NOT NULL,           -- Stock symbol (e.g., '2330')
    timestamp TIMESTAMPTZ NOT NULL,        -- Time of the candle
    open NUMERIC(12, 2) NOT NULL,          -- Opening price
    high NUMERIC(12, 2) NOT NULL,          -- Highest price
    low NUMERIC(12, 2) NOT NULL,           -- Lowest price
    close NUMERIC(12, 2) NOT NULL,         -- Closing price
    volume BIGINT NOT NULL DEFAULT 0,      -- Trading volume (shares)
    turnover NUMERIC(20, 2) DEFAULT 0,     -- Turnover (TWD)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT pk_stock_ohlcv PRIMARY KEY (symbol, timestamp),
    CONSTRAINT chk_ohlcv_prices CHECK (
        high >= low AND 
        high >= open AND 
        high >= close AND 
        low <= open AND 
        low <= close
    ),
    CONSTRAINT chk_volume_positive CHECK (volume >= 0)
);

-- Convert to TimescaleDB hypertable (partitioned by time)
SELECT create_hypertable(
    'stock_ohlcv', 
    'timestamp',
    chunk_time_interval => INTERVAL '1 month',
    if_not_exists => TRUE
);

-- Add indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_ohlcv_symbol_time 
    ON stock_ohlcv (symbol, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_ohlcv_timestamp 
    ON stock_ohlcv (timestamp DESC);

-- Comment
COMMENT ON TABLE stock_ohlcv IS 'Time-series OHLCV (candlestick) data for all Taiwan stocks';

-- ============================================================================
-- 2. Continuous Aggregates - Multi-Timeframe Auto-Aggregation
-- ============================================================================

-- Daily aggregates (from minute/hourly data if available)
CREATE MATERIALIZED VIEW IF NOT EXISTS ohlcv_daily
WITH (timescaledb.continuous) AS
SELECT 
    symbol,
    time_bucket('1 day', timestamp) AS day,
    FIRST(open, timestamp) AS open,
    MAX(high) AS high,
    MIN(low) AS low,
    LAST(close, timestamp) AS close,
    SUM(volume) AS volume,
    SUM(turnover) AS turnover
FROM stock_ohlcv
GROUP BY symbol, day
WITH NO DATA;

-- Weekly aggregates
CREATE MATERIALIZED VIEW IF NOT EXISTS ohlcv_weekly
WITH (timescaledb.continuous) AS
SELECT 
    symbol,
    time_bucket('1 week', timestamp) AS week,
    FIRST(open, timestamp) AS open,
    MAX(high) AS high,
    MIN(low) AS low,
    LAST(close, timestamp) AS close,
    SUM(volume) AS volume,
    SUM(turnover) AS turnover
FROM stock_ohlcv
GROUP BY symbol, week
WITH NO DATA;

-- Monthly aggregates
CREATE MATERIALIZED VIEW IF NOT EXISTS ohlcv_monthly
WITH (timescaledb.continuous) AS
SELECT 
    symbol,
    time_bucket('1 month', timestamp) AS month,
    FIRST(open, timestamp) AS open,
    MAX(high) AS high,
    MIN(low) AS low,
    LAST(close, timestamp) AS close,
    SUM(volume) AS volume,
    SUM(turnover) AS turnover
FROM stock_ohlcv
GROUP BY symbol, month
WITH NO DATA;

-- Add refresh policies (auto-refresh every hour)
SELECT add_continuous_aggregate_policy('ohlcv_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

SELECT add_continuous_aggregate_policy('ohlcv_weekly',
    start_offset => INTERVAL '3 weeks',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

SELECT add_continuous_aggregate_policy('ohlcv_monthly',
    start_offset => INTERVAL '3 months',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- ============================================================================
-- 3. Technical Indicators Cache Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS technical_indicators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(10) NOT NULL,
    indicator_type VARCHAR(50) NOT NULL,    -- 'MA', 'RSI', 'MACD', 'BB', 'KDJ'
    timeframe VARCHAR(20) NOT NULL,         -- '1d', '1w', '1m'
    params JSONB NOT NULL,                  -- Indicator parameters (e.g., {"period": 20})
    data JSONB NOT NULL,                    -- Calculated indicator data
    calculated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,                 -- Cache expiration
    
    CONSTRAINT uq_indicator_cache UNIQUE (symbol, indicator_type, timeframe, params)
);

-- Index for fast lookup
CREATE INDEX IF NOT EXISTS idx_indicators_lookup 
    ON technical_indicators (symbol, indicator_type, timeframe);

CREATE INDEX IF NOT EXISTS idx_indicators_expiry 
    ON technical_indicators (expires_at) 
    WHERE expires_at IS NOT NULL;

COMMENT ON TABLE technical_indicators IS 'Cache table for pre-calculated technical indicators';

-- ============================================================================
-- 4. Helper Functions
-- ============================================================================

-- Function to get latest OHLCV for a symbol
CREATE OR REPLACE FUNCTION get_latest_ohlcv(
    p_symbol VARCHAR(10),
    p_limit INT DEFAULT 100
)
RETURNS TABLE (
    time_stamp TIMESTAMPTZ,
    open_price NUMERIC,
    high_price NUMERIC,
    low_price NUMERIC,
    close_price NUMERIC,
    trade_volume BIGINT,
    trade_turnover NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        o.timestamp,
        o.open,
        o.high,
        o.low,
        o.close,
        o.volume,
        o.turnover
    FROM stock_ohlcv o
    WHERE o.symbol = p_symbol
    ORDER BY o.timestamp DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Function to get OHLCV in time range
CREATE OR REPLACE FUNCTION get_ohlcv_range(
    p_symbol VARCHAR(10),
    p_start TIMESTAMPTZ,
    p_end TIMESTAMPTZ,
    p_timeframe VARCHAR(20) DEFAULT '1d'
)
RETURNS TABLE (
    time_stamp TIMESTAMPTZ,
    open_price NUMERIC,
    high_price NUMERIC,
    low_price NUMERIC,
    close_price NUMERIC,
    trade_volume BIGINT,
    trade_turnover NUMERIC
) AS $$
BEGIN
    -- Select appropriate aggregation based on timeframe
    CASE p_timeframe
        WHEN '1w' THEN
            RETURN QUERY
            SELECT week, open, high, low, close, volume, turnover
            FROM ohlcv_weekly
            WHERE symbol = p_symbol 
                AND week BETWEEN p_start AND p_end
            ORDER BY week;
        WHEN '1m' THEN
            RETURN QUERY
            SELECT month, open, high, low, close, volume, turnover
            FROM ohlcv_monthly
            WHERE symbol = p_symbol 
                AND month BETWEEN p_start AND p_end
            ORDER BY month;
        ELSE
            -- Default to daily or raw data
            RETURN QUERY
            SELECT o.timestamp, o.open, o.high, o.low, o.close, o.volume, o.turnover
            FROM stock_ohlcv o
            WHERE o.symbol = p_symbol 
                AND o.timestamp BETWEEN p_start AND p_end
            ORDER BY o.timestamp;
    END CASE;
END;
$$ LANGUAGE plpgsql;

-- Function to clean expired indicator cache
CREATE OR REPLACE FUNCTION clean_expired_indicators()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM technical_indicators
    WHERE expires_at IS NOT NULL AND expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 5. Data Retention Policy (Optional - Keep 5 years of data)
-- ============================================================================

-- Automatically drop chunks older than 5 years
SELECT add_retention_policy('stock_ohlcv', INTERVAL '5 years', if_not_exists => TRUE);

-- ============================================================================
-- 6. Grant Permissions
-- ============================================================================

GRANT SELECT, INSERT, UPDATE, DELETE ON stock_ohlcv TO psm_user;
GRANT SELECT ON ohlcv_daily, ohlcv_weekly, ohlcv_monthly TO psm_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON technical_indicators TO psm_user;

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Insert migration record (if you have a migrations tracking table)
-- INSERT INTO schema_migrations (version, description, applied_at) 
-- VALUES ('003', 'Market data OHLCV schema and continuous aggregates', NOW());

COMMENT ON SCHEMA public IS 'Phase 2 market data schema applied successfully';
