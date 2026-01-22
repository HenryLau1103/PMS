'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { RealtimeQuote, OrderBookLevel, MarketStatus } from '@/types/api';
import { getRealtimeQuote, getMarketStatus } from '@/lib/realtimeApi';

interface OrderBookPanelProps {
  symbol: string;
}

export default function OrderBookPanel({ symbol }: OrderBookPanelProps) {
  const [quote, setQuote] = useState<RealtimeQuote | null>(null);
  const [marketStatus, setMarketStatus] = useState<MarketStatus | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    try {
      // Always fetch market status first
      const status = await getMarketStatus();
      setMarketStatus(status);

      // Try to fetch quote data
      try {
        const data = await getRealtimeQuote(symbol);
        setQuote(data);
        setError(null);
      } catch (quoteErr) {
        // If market is closed, this is expected
        if (!status.is_open) {
          setError(null); // Clear error since this is expected
        } else if (!quote) {
          setError('無法載入報價');
        }
      }
    } catch (err) {
      console.error('Failed to fetch data:', err);
      if (!quote && !marketStatus) {
        setError('無法載入資料');
      }
    } finally {
      setLoading(false);
    }
  }, [symbol, quote, marketStatus]);

  useEffect(() => {
    setLoading(true);
    setQuote(null);
    fetchData();

    const intervalId = setInterval(fetchData, 5000);
    return () => clearInterval(intervalId);
  }, [symbol]); // Only depend on symbol, not fetchData

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat('en-US').format(num);
  };

  // Calculate max volume for relative bars
  const getMaxVolume = () => {
    if (!quote?.order_book) return 0;
    const bidMax = Math.max(...quote.order_book.bids.map(b => b.volume), 0);
    const askMax = Math.max(...quote.order_book.asks.map(a => a.volume), 0);
    return Math.max(bidMax, askMax, 1); // Avoid division by zero
  };

  const maxVolume = getMaxVolume();

  const isLimitUp = (price: string) => quote?.limit_up === price;
  const isLimitDown = (price: string) => quote?.limit_down === price;

  // Render a row for the order book
  const renderRow = (bid?: OrderBookLevel, ask?: OrderBookLevel, index?: number) => {
    const bidWidth = bid ? (bid.volume / maxVolume) * 100 : 0;
    const askWidth = ask ? (ask.volume / maxVolume) * 100 : 0;

    return (
      <div key={index} className="grid grid-cols-2 h-8 text-sm relative border-b border-gray-100 last:border-0">
        {/* Left Side: Bids (Volume | Price) */}
        <div className="grid grid-cols-2 border-r border-gray-100 relative">
          {/* Bid Volume */}
          <div className="relative flex items-center justify-end pr-2 overflow-hidden">
            {bid && (
              <>
                <div 
                  className="absolute right-0 top-0 bottom-0 bg-red-100 transition-all duration-500 ease-out"
                  style={{ width: `${bidWidth}%` }}
                />
                <span className="relative z-10 font-mono text-gray-700 text-xs">{formatNumber(bid.volume)}</span>
              </>
            )}
          </div>
          
          {/* Bid Price */}
          <div className={`relative flex items-center justify-center font-bold ${
            bid ? (isLimitUp(bid.price) ? 'bg-red-500 text-white' : isLimitDown(bid.price) ? 'bg-green-500 text-white' : 'text-red-600') : ''
          }`}>
            {bid?.price}
          </div>
        </div>

        {/* Right Side: Asks (Price | Volume) */}
        <div className="grid grid-cols-2 relative">
          {/* Ask Price */}
          <div className={`relative flex items-center justify-center font-bold border-r border-gray-100 ${
             ask ? (isLimitUp(ask.price) ? 'bg-red-500 text-white' : isLimitDown(ask.price) ? 'bg-green-500 text-white' : 'text-green-600') : ''
          }`}>
            {ask?.price}
          </div>

          {/* Ask Volume */}
          <div className="relative flex items-center justify-start pl-2 overflow-hidden">
            {ask && (
              <>
                <div 
                  className="absolute left-0 top-0 bottom-0 bg-green-100 transition-all duration-500 ease-out"
                  style={{ width: `${askWidth}%` }}
                />
                <span className="relative z-10 font-mono text-gray-700 text-xs">{formatNumber(ask.volume)}</span>
              </>
            )}
          </div>
        </div>
      </div>
    );
  };

  // Prepare data rows (ensure 5 rows)
  const rows = Array(5).fill(null).map((_, i) => {
    const bid = quote?.order_book?.bids[i];
    const ask = quote?.order_book?.asks[i];
    return { bid, ask };
  });

  const isMarketOpen = marketStatus?.is_open ?? false;
  const marketMessage = marketStatus?.message || '';

  return (
    <div className="w-full max-w-[350px] bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden font-sans">
      {/* Header */}
      <div className="bg-gray-50 px-3 py-2 border-b border-gray-200 flex justify-between items-center">
        <div className="font-bold text-gray-800 text-sm">
          五檔報價 - {symbol} {quote?.name}
        </div>
        {isMarketOpen ? (
          loading && !quote && (
            <div className="animate-spin h-3 w-3 border-2 border-gray-400 border-t-transparent rounded-full" />
          )
        ) : (
          <span className="text-[10px] font-medium text-gray-500 uppercase tracking-wider">CLOSED</span>
        )}
      </div>

      {/* Column Headers */}
      <div className="grid grid-cols-2 bg-gray-50 text-xs text-gray-500 font-medium border-b border-gray-200">
        <div className="grid grid-cols-2 border-r border-gray-200">
          <div className="text-right py-1 pr-2">買量</div>
          <div className="text-center py-1">買價</div>
        </div>
        <div className="grid grid-cols-2">
          <div className="text-center py-1 border-r border-gray-200">賣價</div>
          <div className="text-left py-1 pl-2">賣量</div>
        </div>
      </div>

      {/* Content */}
      <div className="relative min-h-[160px]">
        {loading && !quote && !marketStatus ? (
          // Skeleton Loading
          <div className="animate-pulse p-4 space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-4 bg-gray-100 rounded w-full" />
            ))}
          </div>
        ) : !isMarketOpen ? (
          // Market Closed
          <div className="absolute inset-0 flex flex-col items-center justify-center text-gray-400 gap-2">
            <svg className="w-8 h-8 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm font-medium text-gray-500">已收盤</span>
            <span className="text-xs text-gray-400">{marketMessage}</span>
          </div>
        ) : error ? (
          <div className="absolute inset-0 flex items-center justify-center text-gray-400 text-sm">
            {error}
          </div>
        ) : !quote?.order_book || (quote.order_book.bids.length === 0 && quote.order_book.asks.length === 0) ? (
          <div className="absolute inset-0 flex items-center justify-center text-gray-400 text-sm">
            尚無掛單資料
          </div>
        ) : (
          // Order Book Rows
          <div className="bg-white">
            {rows.map((row, i) => renderRow(row.bid, row.ask, i))}
          </div>
        )}
      </div>
    </div>
  );
}
