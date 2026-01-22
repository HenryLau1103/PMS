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
