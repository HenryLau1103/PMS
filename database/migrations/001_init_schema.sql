-- ================================================
-- Taiwan Stock Portfolio Management System
-- Phase 1: Core Database Schema
-- ================================================

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ================================================
-- USERS & PORTFOLIOS
-- ================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE portfolios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    currency VARCHAR(3) DEFAULT 'TWD',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_portfolios_user_id ON portfolios(user_id);

-- ================================================
-- LEDGER EVENTS (Immutable Transaction Log)
-- Event Sourcing Pattern
-- ================================================

CREATE TABLE ledger_events (
    event_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    
    -- Event metadata
    event_type VARCHAR(50) NOT NULL, -- 'BUY', 'SELL', 'DIVIDEND', 'SPLIT', 'RIGHTS', 'CORRECTION'
    
    -- Transaction details
    symbol VARCHAR(20) NOT NULL,     -- '2330.TW' format
    quantity DECIMAL(15, 6),         -- Number of shares
    price DECIMAL(12, 2),            -- Price per share
    
    -- Fees and taxes (Taiwan specific)
    fee DECIMAL(10, 2) DEFAULT 0,              -- 手續費
    tax DECIMAL(10, 2) DEFAULT 0,              -- 證券交易稅
    total_amount DECIMAL(15, 2),               -- Total transaction amount
    
    -- Timestamps
    occurred_at TIMESTAMPTZ NOT NULL,          -- When the transaction occurred
    recorded_at TIMESTAMPTZ DEFAULT NOW(),     -- When it was recorded in system
    
    -- Audit trail
    source VARCHAR(50) DEFAULT 'manual',       -- 'manual', 'import', 'api'
    notes TEXT,
    
    -- Additional data (JSONB for flexibility)
    payload JSONB,
    
    -- Constraints
    CONSTRAINT valid_event_type CHECK (event_type IN ('BUY', 'SELL', 'DIVIDEND', 'SPLIT', 'RIGHTS', 'CORRECTION')),
    CONSTRAINT valid_symbol CHECK (symbol ~ '^[0-9]{4}\.(TW|TWO)$'),
    CONSTRAINT positive_quantity CHECK (quantity IS NULL OR quantity > 0),
    CONSTRAINT positive_price CHECK (price IS NULL OR price >= 0)
);

-- Indexes for performance
CREATE INDEX idx_ledger_events_user_id ON ledger_events(user_id);
CREATE INDEX idx_ledger_events_portfolio_id ON ledger_events(portfolio_id);
CREATE INDEX idx_ledger_events_symbol ON ledger_events(symbol);
CREATE INDEX idx_ledger_events_occurred_at ON ledger_events(occurred_at DESC);
CREATE INDEX idx_ledger_events_type ON ledger_events(event_type);
CREATE INDEX idx_ledger_events_composite ON ledger_events(portfolio_id, symbol, occurred_at);

-- ================================================
-- TAX LOTS (FIFO Cost Basis Tracking)
-- ================================================

CREATE TABLE tax_lots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    
    -- Lot details
    purchase_event_id UUID NOT NULL REFERENCES ledger_events(event_id),
    purchase_date TIMESTAMPTZ NOT NULL,
    purchase_price DECIMAL(12, 2) NOT NULL,
    
    -- Quantity tracking
    original_quantity DECIMAL(15, 6) NOT NULL,
    remaining_quantity DECIMAL(15, 6) NOT NULL,
    
    -- Status
    is_closed BOOLEAN DEFAULT FALSE,
    closed_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT positive_quantities CHECK (
        original_quantity > 0 AND 
        remaining_quantity >= 0 AND 
        remaining_quantity <= original_quantity
    )
);

CREATE INDEX idx_tax_lots_portfolio_symbol ON tax_lots(portfolio_id, symbol);
CREATE INDEX idx_tax_lots_purchase_date ON tax_lots(purchase_date);
CREATE INDEX idx_tax_lots_open ON tax_lots(portfolio_id, symbol) WHERE is_closed = FALSE;

-- ================================================
-- POSITIONS (Current Holdings - Materialized View)
-- ================================================

CREATE MATERIALIZED VIEW positions_current AS
WITH aggregated_positions AS (
    SELECT 
        portfolio_id,
        symbol,
        SUM(
            CASE 
                WHEN event_type = 'BUY' THEN quantity
                WHEN event_type = 'SELL' THEN -quantity
                WHEN event_type = 'SPLIT' THEN quantity * (payload->>'ratio')::DECIMAL
                ELSE 0
            END
        ) as total_quantity,
        SUM(
            CASE 
                WHEN event_type = 'BUY' THEN total_amount
                WHEN event_type = 'SELL' THEN -total_amount
                ELSE 0
            END
        ) as total_cost
    FROM ledger_events
    WHERE event_type IN ('BUY', 'SELL', 'SPLIT')
    GROUP BY portfolio_id, symbol
)
SELECT 
    portfolio_id,
    symbol,
    total_quantity,
    total_cost,
    CASE 
        WHEN total_quantity > 0 THEN total_cost / total_quantity
        ELSE 0
    END as avg_cost_per_share,
    NOW() as last_updated
FROM aggregated_positions
WHERE total_quantity > 0;

CREATE UNIQUE INDEX idx_positions_current_unique ON positions_current(portfolio_id, symbol);

-- ================================================
-- REALIZED P&L (Closed Positions)
-- ================================================

CREATE TABLE realized_pnl (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    
    -- Transaction references
    buy_event_id UUID NOT NULL REFERENCES ledger_events(event_id),
    sell_event_id UUID NOT NULL REFERENCES ledger_events(event_id),
    
    -- P&L calculation
    quantity DECIMAL(15, 6) NOT NULL,
    buy_price DECIMAL(12, 2) NOT NULL,
    sell_price DECIMAL(12, 2) NOT NULL,
    realized_pnl DECIMAL(15, 2) NOT NULL,
    
    -- Fees
    total_fees DECIMAL(10, 2) DEFAULT 0,
    total_taxes DECIMAL(10, 2) DEFAULT 0,
    
    -- Holding period
    buy_date TIMESTAMPTZ NOT NULL,
    sell_date TIMESTAMPTZ NOT NULL,
    holding_days INTEGER,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_realized_pnl_portfolio ON realized_pnl(portfolio_id);
CREATE INDEX idx_realized_pnl_symbol ON realized_pnl(symbol);
CREATE INDEX idx_realized_pnl_dates ON realized_pnl(sell_date DESC);

-- ================================================
-- CORPORATE ACTIONS (Taiwan Stock Specific)
-- ================================================

CREATE TABLE corporate_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(20) NOT NULL,
    action_type VARCHAR(50) NOT NULL, -- 'DIVIDEND', 'STOCK_DIVIDEND', 'SPLIT', 'RIGHTS', 'MERGER'
    
    -- Important dates
    announcement_date DATE,
    ex_date DATE NOT NULL,           -- 除權息日
    record_date DATE,                -- 停止過戶日
    payment_date DATE,               -- 發放日
    
    -- Action details
    cash_dividend DECIMAL(10, 4),    -- 現金股利 (per share)
    stock_dividend DECIMAL(10, 4),   -- 股票股利 (per share)
    split_ratio DECIMAL(10, 4),      -- 分割/合併比例
    rights_ratio DECIMAL(10, 4),     -- 認購比例
    subscription_price DECIMAL(12, 2), -- 認購價格
    
    -- Price adjustment factor
    adjustment_factor DECIMAL(12, 8) DEFAULT 1.0,
    
    -- Metadata
    source VARCHAR(100),
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT valid_action_type CHECK (
        action_type IN ('DIVIDEND', 'STOCK_DIVIDEND', 'SPLIT', 'RIGHTS', 'MERGER')
    ),
    CONSTRAINT valid_symbol_ca CHECK (symbol ~ '^[0-9]{4}\.(TW|TWO)$')
);

CREATE INDEX idx_corporate_actions_symbol ON corporate_actions(symbol);
CREATE INDEX idx_corporate_actions_ex_date ON corporate_actions(ex_date DESC);

-- ================================================
-- HELPER FUNCTIONS
-- ================================================

-- Function to refresh positions materialized view
CREATE OR REPLACE FUNCTION refresh_positions()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY positions_current;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate unrealized P&L
CREATE OR REPLACE FUNCTION calculate_unrealized_pnl(
    p_portfolio_id UUID,
    p_symbol VARCHAR(20),
    p_current_price DECIMAL(12, 2)
)
RETURNS TABLE(
    symbol VARCHAR(20),
    quantity DECIMAL(15, 6),
    avg_cost DECIMAL(12, 2),
    current_price DECIMAL(12, 2),
    market_value DECIMAL(15, 2),
    cost_basis DECIMAL(15, 2),
    unrealized_pnl DECIMAL(15, 2),
    unrealized_pnl_pct DECIMAL(10, 4)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pc.symbol,
        pc.total_quantity,
        pc.avg_cost_per_share,
        p_current_price,
        pc.total_quantity * p_current_price as market_value,
        pc.total_cost as cost_basis,
        (pc.total_quantity * p_current_price) - pc.total_cost as unrealized_pnl,
        CASE 
            WHEN pc.total_cost > 0 THEN 
                ((pc.total_quantity * p_current_price) - pc.total_cost) / pc.total_cost * 100
            ELSE 0
        END as unrealized_pnl_pct
    FROM positions_current pc
    WHERE pc.portfolio_id = p_portfolio_id
      AND pc.symbol = p_symbol;
END;
$$ LANGUAGE plpgsql;

-- ================================================
-- TRIGGER FUNCTIONS
-- ================================================

-- Auto-update timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_portfolios_updated_at BEFORE UPDATE ON portfolios
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ================================================
-- INITIAL DATA (Demo User)
-- ================================================

INSERT INTO users (id, email, username, password_hash) VALUES
    ('00000000-0000-0000-0000-000000000001', 'demo@psm.tw', 'demo_user', '$2a$10$demo_hash_placeholder');

INSERT INTO portfolios (id, user_id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001', '台股投資組合', '個人台股投資帳戶');

-- ================================================
-- SCHEMA VERSION
-- ================================================

CREATE TABLE schema_version (
    version INTEGER PRIMARY KEY,
    description TEXT,
    applied_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO schema_version (version, description) VALUES
    (1, 'Phase 1: Core ledger, positions, and Taiwan stock support');
