// API Types
export type EventType = 'BUY' | 'SELL' | 'DIVIDEND' | 'SPLIT' | 'RIGHTS' | 'CORRECTION';

export interface LedgerEvent {
  event_id: string;
  user_id: string;
  portfolio_id: string;
  event_type: EventType;
  symbol: string;
  quantity: string;
  price: string;
  fee: string;
  tax: string;
  total_amount: string;
  occurred_at: string;
  recorded_at: string;
  source: string;
  notes?: string;
  payload?: string;
}

export interface CreateEventRequest {
  portfolio_id: string;
  event_type: EventType;
  symbol: string;
  quantity: string;
  price: string;
  fee: string;
  tax: string;
  occurred_at: string;
  notes?: string;
}

export interface Position {
  portfolio_id: string;
  symbol: string;
  total_quantity: string;
  total_cost: string;
  avg_cost_per_share: string;
  last_updated: string;
}

export interface UnrealizedPnL {
  symbol: string;
  quantity: string;
  avg_cost: string;
  current_price: string;
  market_value: string;
  cost_basis: string;
  unrealized_pnl: string;
  unrealized_pnl_pct: string;
}

export interface Portfolio {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  currency: string;
  created_at: string;
  updated_at: string;
}

export interface TaiwanStock {
  symbol: string;
  name: string;
  name_en?: string;
  market: string;
  industry?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Real-time data types (Phase 3)
export interface OrderBookLevel {
  price: string;
  volume: number;
}

export interface OrderBook {
  bids: OrderBookLevel[];
  asks: OrderBookLevel[];
}

export interface RealtimeQuote {
  symbol: string;
  name: string;
  price: string;
  change: string;
  change_percent: string;
  open: string;
  high: string;
  low: string;
  prev_close: string;
  volume: number;
  turnover: string;
  bid_price: string;
  ask_price: string;
  bid_volume: number;
  ask_volume: number;
  trade_time: string;
  is_open: boolean;
  limit_up: string;
  limit_down: string;
  updated_at: string;
  order_book?: OrderBook;
}

export interface MarketStatus {
  is_open: boolean;
  status: 'pre_market' | 'open' | 'after_hours' | 'closed' | 'holiday';
  message: string;
  next_open_time?: string;
  server_time: string;
}

export interface WSMessage {
  action: 'subscribe' | 'unsubscribe';
  symbols: string[];
}

export interface WSResponse {
  type: 'quote' | 'status' | 'error' | 'subscribed' | 'unsubscribed';
  data?: RealtimeQuote | MarketStatus | string[];
  message?: string;
}
