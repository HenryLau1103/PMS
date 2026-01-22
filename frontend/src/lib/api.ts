import axios from 'axios';
import type { CreateEventRequest, LedgerEvent, Position, Portfolio, UnrealizedPnL, TaiwanStock } from '@/types/api';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const apiClient = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Event/Transaction APIs
export const createEvent = async (data: CreateEventRequest): Promise<LedgerEvent> => {
  const response = await apiClient.post<LedgerEvent>('/events', data);
  return response.data;
};

export const getEvents = async (portfolioId: string, limit: number = 100): Promise<LedgerEvent[]> => {
  const response = await apiClient.get<LedgerEvent[]>(`/portfolios/${portfolioId}/events`, {
    params: { limit },
  });
  return response.data;
};

export const getEventsBySymbol = async (portfolioId: string, symbol: string): Promise<LedgerEvent[]> => {
  const response = await apiClient.get<LedgerEvent[]>(`/portfolios/${portfolioId}/events/${symbol}`);
  return response.data;
};

// Position APIs
export const getPositions = async (portfolioId: string): Promise<Position[]> => {
  const response = await apiClient.get<Position[]>(`/portfolios/${portfolioId}/positions`);
  return response.data;
};

export const getPosition = async (portfolioId: string, symbol: string): Promise<Position> => {
  const response = await apiClient.get<Position>(`/portfolios/${portfolioId}/positions/${symbol}`);
  return response.data;
};

export const getUnrealizedPnL = async (
  portfolioId: string,
  symbol: string,
  currentPrice: string
): Promise<UnrealizedPnL> => {
  const response = await apiClient.get<UnrealizedPnL>(
    `/portfolios/${portfolioId}/positions/${symbol}/pnl`,
    {
      params: { current_price: currentPrice },
    }
  );
  return response.data;
};

// Portfolio APIs
export const getPortfolio = async (portfolioId: string): Promise<Portfolio> => {
  const response = await apiClient.get<Portfolio>(`/portfolios/${portfolioId}`);
  return response.data;
};

export const getPortfolios = async (): Promise<Portfolio[]> => {
  const response = await apiClient.get<Portfolio[]>('/portfolios');
  return response.data;
};

// Health check
export const healthCheck = async (): Promise<{ status: string }> => {
  const response = await axios.get(`${API_BASE_URL}/health`);
  return response.data;
};

// Stock APIs
export const searchStocks = async (query: string, limit: number = 20): Promise<TaiwanStock[]> => {
  const response = await apiClient.get<TaiwanStock[]>('/stocks/search', {
    params: { q: query, limit },
  });
  return response.data;
};

export const getStock = async (symbol: string): Promise<TaiwanStock> => {
  const response = await apiClient.get<TaiwanStock>(`/stocks/${symbol}`);
  return response.data;
};
