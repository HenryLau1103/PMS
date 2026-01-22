import axios from 'axios';
import type { RealtimeQuote, MarketStatus, WSMessage, WSResponse } from '@/types/api';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const WS_BASE_URL = API_BASE_URL.replace('http', 'ws');

const apiClient = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// REST API functions
export const getMarketStatus = async (): Promise<MarketStatus> => {
  const response = await apiClient.get<{ success: boolean; data: MarketStatus }>('/market/status');
  return response.data.data;
};

export const getRealtimeQuote = async (symbol: string): Promise<RealtimeQuote> => {
  const response = await apiClient.get<{ success: boolean; data: RealtimeQuote }>(`/realtime/${symbol}`);
  return response.data.data;
};

export const getBatchQuotes = async (symbols: string[]): Promise<RealtimeQuote[]> => {
  const response = await apiClient.get<{ success: boolean; data: RealtimeQuote[]; count: number }>(
    '/realtime',
    { params: { symbols: symbols.join(',') } }
  );
  return response.data.data;
};

// WebSocket connection manager
export class RealtimeWebSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private subscribedSymbols: Set<string> = new Set();
  
  private onQuoteCallback: ((quote: RealtimeQuote) => void) | null = null;
  private onStatusCallback: ((status: MarketStatus) => void) | null = null;
  private onConnectedCallback: (() => void) | null = null;
  private onDisconnectedCallback: (() => void) | null = null;
  private onErrorCallback: ((error: string) => void) | null = null;

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    const wsUrl = `${WS_BASE_URL}/ws/realtime`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.onConnectedCallback?.();

      // Re-subscribe to symbols if any
      if (this.subscribedSymbols.size > 0) {
        this.subscribe(Array.from(this.subscribedSymbols));
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.onDisconnectedCallback?.();
      this.attemptReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.onErrorCallback?.('WebSocket connection error');
    };

    this.ws.onmessage = (event) => {
      try {
        const response: WSResponse = JSON.parse(event.data);
        this.handleMessage(response);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };
  }

  private handleMessage(response: WSResponse): void {
    switch (response.type) {
      case 'quote':
        if (response.data && this.onQuoteCallback) {
          this.onQuoteCallback(response.data as RealtimeQuote);
        }
        break;
      case 'status':
        if (response.data && this.onStatusCallback) {
          this.onStatusCallback(response.data as MarketStatus);
        }
        break;
      case 'subscribed':
        console.log('Subscribed to symbols:', response.data);
        break;
      case 'unsubscribed':
        console.log('Unsubscribed from symbols:', response.data);
        break;
      case 'error':
        console.error('WebSocket error:', response.message);
        this.onErrorCallback?.(response.message || 'Unknown error');
        break;
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnect attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
    
    setTimeout(() => {
      this.connect();
    }, delay);
  }

  subscribe(symbols: string[]): void {
    symbols.forEach(s => this.subscribedSymbols.add(s.toUpperCase()));
    
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message: WSMessage = {
        action: 'subscribe',
        symbols: symbols.map(s => s.toUpperCase()),
      };
      this.ws.send(JSON.stringify(message));
    }
  }

  unsubscribe(symbols: string[]): void {
    symbols.forEach(s => this.subscribedSymbols.delete(s.toUpperCase()));
    
    if (this.ws?.readyState === WebSocket.OPEN) {
      const message: WSMessage = {
        action: 'unsubscribe',
        symbols: symbols.map(s => s.toUpperCase()),
      };
      this.ws.send(JSON.stringify(message));
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.subscribedSymbols.clear();
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  // Event handlers
  onQuote(callback: (quote: RealtimeQuote) => void): void {
    this.onQuoteCallback = callback;
  }

  onMarketStatus(callback: (status: MarketStatus) => void): void {
    this.onStatusCallback = callback;
  }

  onConnected(callback: () => void): void {
    this.onConnectedCallback = callback;
  }

  onDisconnected(callback: () => void): void {
    this.onDisconnectedCallback = callback;
  }

  onError(callback: (error: string) => void): void {
    this.onErrorCallback = callback;
  }
}

// Singleton instance for global use
let wsInstance: RealtimeWebSocket | null = null;

export const getRealtimeWS = (): RealtimeWebSocket => {
  if (!wsInstance) {
    wsInstance = new RealtimeWebSocket();
  }
  return wsInstance;
};
