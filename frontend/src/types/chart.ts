// Chart-related type definitions for TradingView Lightweight Charts

export interface OHLCVData {
  symbol: string;
  timestamp: string;
  open: string;
  high: string;
  low: string;
  close: string;
  volume: number;
  turnover: string;
}

export interface CandlestickData {
  time: number; // Unix timestamp in seconds
  open: number;
  high: number;
  low: number;
  close: number;
}

export interface VolumeData {
  time: number;
  value: number;
  color?: string;
}

export interface IndicatorValue {
  time: number;
  value: number;
}

export interface MAData {
  timestamp: string;
  value: string;
}

export interface RSIData {
  timestamp: string;
  value: string;
}

export interface MACDData {
  timestamp: string;
  macd: string;
  signal: string;
  histogram: string;
}

export interface BBData {
  timestamp: string;
  upper: string;
  middle: string;
  lower: string;
}

export interface KDJData {
  timestamp: string;
  k: string;
  d: string;
  j: string;
}

export interface IndicatorConfig {
  enabled: boolean;
  period?: number;
  type?: string;
  color?: string;
}

export interface ChartConfig {
  ma5: IndicatorConfig;
  ma10: IndicatorConfig;
  ma20: IndicatorConfig;
  rsi: IndicatorConfig;
  macd: IndicatorConfig;
  bb: IndicatorConfig;
  kdj: IndicatorConfig;
}

export type Timeframe = '1d' | '1w' | '1m';
