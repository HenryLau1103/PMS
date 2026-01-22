import axios from 'axios';
import type {
  OHLCVData,
  MAData,
  RSIData,
  MACDData,
  BBData,
  KDJData,
} from '@/types/chart';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Fetch OHLCV candlestick data
export async function getOHLCV(
  symbol: string,
  from?: string,
  to?: string,
  limit: number = 500
): Promise<OHLCVData[]> {
  const params = new URLSearchParams();
  if (from) params.append('from', from);
  if (to) params.append('to', to);
  params.append('limit', limit.toString());

  const response = await api.get(`/api/v1/stocks/${symbol}/ohlcv?${params.toString()}`);
  return response.data.data || [];
}

// Fetch Moving Average
export async function getMA(
  symbol: string,
  period: number = 20,
  type: 'SMA' | 'EMA' = 'SMA',
  limit: number = 500
): Promise<MAData[]> {
  const response = await api.get(`/api/v1/indicators/${symbol}/ma`, {
    params: { period, type, limit },
  });
  return response.data.data || [];
}

// Fetch RSI
export async function getRSI(
  symbol: string,
  period: number = 14,
  limit: number = 500
): Promise<RSIData[]> {
  const response = await api.get(`/api/v1/indicators/${symbol}/rsi`, {
    params: { period, limit },
  });
  return response.data.data || [];
}

// Fetch MACD
export async function getMACD(
  symbol: string,
  fast: number = 12,
  slow: number = 26,
  signal: number = 9,
  limit: number = 500
): Promise<MACDData[]> {
  const response = await api.get(`/api/v1/indicators/${symbol}/macd`, {
    params: { fast, slow, signal, limit },
  });
  return response.data.data || [];
}

// Fetch Bollinger Bands
export async function getBollingerBands(
  symbol: string,
  period: number = 20,
  stddev: number = 2,
  limit: number = 500
): Promise<BBData[]> {
  const response = await api.get(`/api/v1/indicators/${symbol}/bb`, {
    params: { period, stddev, limit },
  });
  return response.data.data || [];
}

// Fetch KDJ
export async function getKDJ(
  symbol: string,
  period: number = 9,
  limit: number = 500
): Promise<KDJData[]> {
  const response = await api.get(`/api/v1/indicators/${symbol}/kdj`, {
    params: { period, limit },
  });
  return response.data.data || [];
}

// Helper: Convert timestamp to seconds
export function toUnixTimestamp(timestamp: string): number {
  return Math.floor(new Date(timestamp).getTime() / 1000);
}

// Helper: Parse OHLCV data to chart format
export function parseOHLCVData(data: OHLCVData[]) {
  return data.map((d) => ({
    time: toUnixTimestamp(d.timestamp),
    open: parseFloat(d.open),
    high: parseFloat(d.high),
    low: parseFloat(d.low),
    close: parseFloat(d.close),
  }));
}

// Helper: Parse volume data
export function parseVolumeData(data: OHLCVData[]) {
  return data.map((d, i, arr) => {
    const close = parseFloat(d.close);
    const prevClose = i > 0 ? parseFloat(arr[i - 1].close) : close;
    const color = close >= prevClose ? '#26a69a' : '#ef5350'; // Green if up, red if down

    return {
      time: toUnixTimestamp(d.timestamp),
      value: d.volume,
      color,
    };
  });
}

// Helper: Parse indicator data
export function parseIndicatorData(data: MAData[] | RSIData[]): { time: number; value: number }[] {
  return data.map((d) => ({
    time: toUnixTimestamp(d.timestamp),
    value: parseFloat(d.value),
  }));
}
