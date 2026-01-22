'use client';

import React, { useEffect, useState, useCallback } from 'react';
import type { RealtimeQuote, MarketStatus } from '@/types/api';
import { getRealtimeQuote, getMarketStatus } from '@/lib/realtimeApi';

interface RealtimePriceCellProps {
  symbol: string;
  onPriceUpdate?: (quote: RealtimeQuote) => void;
  showChange?: boolean;
  showLimitAlert?: boolean;
  className?: string;
}

// Price status based on limit up/down
type PriceStatus = 'normal' | 'limit_up' | 'limit_down' | 'near_limit_up' | 'near_limit_down';

function getPriceStatus(quote: RealtimeQuote): PriceStatus {
  const price = parseFloat(quote.price);
  const limitUp = parseFloat(quote.limit_up);
  const limitDown = parseFloat(quote.limit_down);
  const prevClose = parseFloat(quote.prev_close);
  
  if (price <= 0 || limitUp <= 0 || limitDown <= 0) return 'normal';
  
  // Exactly at limit
  if (Math.abs(price - limitUp) < 0.01) return 'limit_up';
  if (Math.abs(price - limitDown) < 0.01) return 'limit_down';
  
  // Near limit (within 2% of limit range)
  const limitRange = limitUp - limitDown;
  const nearThreshold = limitRange * 0.05; // 5% of range
  
  if (price >= limitUp - nearThreshold) return 'near_limit_up';
  if (price <= limitDown + nearThreshold) return 'near_limit_down';
  
  return 'normal';
}

function getChangeColor(changePercent: number): string {
  if (changePercent > 0) return 'text-red-600'; // Taiwan: red = up
  if (changePercent < 0) return 'text-green-600'; // Taiwan: green = down
  return 'text-gray-600';
}

function getPriceColorClass(status: PriceStatus): string {
  switch (status) {
    case 'limit_up':
      return 'text-red-600 font-bold bg-red-50 border border-red-200';
    case 'limit_down':
      return 'text-green-600 font-bold bg-green-50 border border-green-200';
    case 'near_limit_up':
      return 'text-red-500';
    case 'near_limit_down':
      return 'text-green-500';
    default:
      return 'text-gray-900';
  }
}

function getLimitBadge(status: PriceStatus): React.ReactNode {
  if (status === 'limit_up') {
    return (
      <span className="ml-1.5 inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold bg-red-500 text-white animate-pulse">
        漲停
      </span>
    );
  }
  if (status === 'limit_down') {
    return (
      <span className="ml-1.5 inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold bg-green-500 text-white animate-pulse">
        跌停
      </span>
    );
  }
  return null;
}

export default function RealtimePriceCell({
  symbol,
  onPriceUpdate,
  showChange = true,
  showLimitAlert = true,
  className = '',
}: RealtimePriceCellProps) {
  const [quote, setQuote] = useState<RealtimeQuote | null>(null);
  const [marketStatus, setMarketStatus] = useState<MarketStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastPrice, setLastPrice] = useState<number | null>(null);
  const [priceFlash, setPriceFlash] = useState<'up' | 'down' | null>(null);

  const fetchQuote = useCallback(async () => {
    try {
      // Always get market status
      const status = await getMarketStatus();
      setMarketStatus(status);

      // Try to get quote
      try {
        const data = await getRealtimeQuote(symbol);
        
        const newPrice = parseFloat(data.price);
        if (lastPrice !== null && newPrice !== lastPrice && newPrice > 0) {
          setPriceFlash(newPrice > lastPrice ? 'up' : 'down');
          setTimeout(() => setPriceFlash(null), 500);
        }
        if (newPrice > 0) {
          setLastPrice(newPrice);
        }
        
        setQuote(data);
        setError(null);
        onPriceUpdate?.(data);
      } catch (quoteErr) {
        // If market is closed, don't show error
        if (!status.is_open) {
          setError(null);
        } else {
          setError('--');
        }
      }
    } catch (err: any) {
      setError('--');
    } finally {
      setLoading(false);
    }
  }, [symbol, lastPrice, onPriceUpdate]);

  useEffect(() => {
    fetchQuote();
    
    const interval = setInterval(fetchQuote, 10000);
    
    return () => clearInterval(interval);
  }, [symbol]);

  if (loading) {
    return (
      <div className={`animate-pulse ${className}`}>
        <div className="h-5 bg-gray-200 rounded w-16"></div>
      </div>
    );
  }

  // If market is closed and no quote, show closed status
  if (!marketStatus?.is_open && !quote) {
    return (
      <div className={`text-gray-400 text-sm ${className}`}>
        已收盤
      </div>
    );
  }

  if (error || !quote) {
    return (
      <div className={`text-gray-400 text-sm ${className}`}>
        --
      </div>
    );
  }

  const price = parseFloat(quote.price);
  const changePercent = parseFloat(quote.change_percent);
  const change = parseFloat(quote.change);
  const status = getPriceStatus(quote);
  
  // Handle zero price (no trade yet)
  if (price <= 0) {
    return (
      <div className={`text-gray-400 text-sm ${className}`}>
        尚無成交
      </div>
    );
  }

  const flashClass = priceFlash === 'up' 
    ? 'bg-red-100 transition-colors duration-300' 
    : priceFlash === 'down' 
      ? 'bg-green-100 transition-colors duration-300'
      : '';

  return (
    <div className={`${className}`}>
      <div className="flex items-center">
        <span className={`text-sm font-medium px-1 rounded ${getPriceColorClass(status)} ${flashClass}`}>
          {price.toFixed(2)}
        </span>
        {showLimitAlert && getLimitBadge(status)}
      </div>
      
      {showChange && (
        <div className={`text-xs mt-0.5 ${getChangeColor(changePercent)}`}>
          {change >= 0 ? '+' : ''}{change.toFixed(2)} ({changePercent >= 0 ? '+' : ''}{changePercent.toFixed(2)}%)
        </div>
      )}
    </div>
  );
}
