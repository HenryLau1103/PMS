'use client';

import React, { useState, useEffect, useRef } from 'react';
import { RealtimeQuote, MarketStatus } from '@/types/api';
import { getRealtimeQuote, getMarketStatus } from '@/lib/realtimeApi';

interface TradeTick {
  id: number;
  time: string;
  price: string;
  volume: number;
  trend: 'up' | 'down' | 'equal';
}

interface TradeTickPanelProps {
  symbol: string;
  maxTicks?: number;
}

export default function TradeTickPanel({ symbol, maxTicks = 10 }: TradeTickPanelProps) {
  const [ticks, setTicks] = useState<TradeTick[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentQuote, setCurrentQuote] = useState<RealtimeQuote | null>(null);
  const [marketStatus, setMarketStatus] = useState<MarketStatus | null>(null);
  
  // Refs to track history without triggering re-renders for logic
  const lastQuoteRef = useRef<RealtimeQuote | null>(null);
  const tickIdRef = useRef(0);
  const isFirstLoad = useRef(true);

  useEffect(() => {
    // Reset state when symbol changes
    setTicks([]);
    setLoading(true);
    setError(null);
    setCurrentQuote(null);
    lastQuoteRef.current = null;
    isFirstLoad.current = true;

    const fetchData = async () => {
      try {
        // Always fetch market status first (this should always work)
        const status = await getMarketStatus();
        setMarketStatus(status);
        
        // Try to fetch quote (may fail if market is closed)
        try {
          const quote = await getRealtimeQuote(symbol);
          
          // Handle first load
          if (isFirstLoad.current) {
            lastQuoteRef.current = quote;
            setCurrentQuote(quote);
            setLoading(false);
            isFirstLoad.current = false;
            setError(null);
            return;
          }

          const lastQuote = lastQuoteRef.current;
          
          // Only process ticks if market is open
          if (status.is_open && lastQuote) {
            const prevVol = lastQuote.volume;
            const currVol = quote.volume;
            const volDelta = currVol - prevVol;

            const prevPrice = parseFloat(lastQuote.price);
            const currPrice = parseFloat(quote.price);
            
            if (volDelta > 0 || prevPrice !== currPrice) {
              const trend = currPrice > prevPrice ? 'up' : currPrice < prevPrice ? 'down' : 'equal';
              
              const now = new Date();
              const timeStr = now.toLocaleTimeString('en-US', { 
                hour12: false, 
                hour: '2-digit', 
                minute: '2-digit', 
                second: '2-digit' 
              });

              const newTick: TradeTick = {
                id: tickIdRef.current++,
                time: timeStr,
                price: quote.price,
                volume: volDelta > 0 ? volDelta : 0,
                trend
              };

              setTicks(prev => {
                const newTicks = [newTick, ...prev];
                return newTicks.slice(0, maxTicks);
              });
            }
          }
          
          setCurrentQuote(quote);
          lastQuoteRef.current = quote;
          setError(null);
        } catch (quoteErr) {
          // Quote fetch failed - this is expected when market is closed
          console.log('Quote fetch failed (expected if market closed):', quoteErr);
          if (isFirstLoad.current) {
            setLoading(false);
            isFirstLoad.current = false;
          }
          // Don't set error if market is closed - this is expected behavior
          if (status.is_open) {
            setError('無法載入即時行情');
          }
        }
      } catch (err) {
        // Market status fetch failed - this is a real error
        console.error('Failed to fetch market status:', err);
        if (isFirstLoad.current) {
          setError('無法載入即時行情');
          setLoading(false);
          isFirstLoad.current = false;
        }
      }
    };

    // Initial fetch
    fetchData();

    // Poll interval
    const intervalId = setInterval(fetchData, 3000);

    return () => clearInterval(intervalId);
  }, [symbol, maxTicks]);

  // Determine display state based on market status
  const isMarketOpen = marketStatus?.is_open ?? false;
  const marketMessage = marketStatus?.message || '';

  if (error) {
    return (
      <div className="w-full max-w-[280px] p-4 bg-white rounded-xl shadow-sm border border-gray-100 text-center text-red-500 text-sm">
        {error}
      </div>
    );
  }

  return (
    <div className="w-full max-w-[280px] bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden font-sans">
      {/* Header */}
      <div className="px-4 py-3 border-b border-gray-100 bg-gray-50/50 flex justify-between items-center">
        <div>
          <h3 className="text-sm font-semibold text-gray-900">即時成交</h3>
          <p className="text-xs text-gray-500 font-mono mt-0.5">{currentQuote?.name || symbol}</p>
        </div>
        {isMarketOpen ? (
          <div className="flex items-center gap-1.5">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
            <span className="text-[10px] font-medium text-emerald-600 uppercase tracking-wider">LIVE</span>
          </div>
        ) : (
          <div className="flex items-center gap-1.5">
            <span className="relative inline-flex rounded-full h-2 w-2 bg-gray-400"></span>
            <span className="text-[10px] font-medium text-gray-500 uppercase tracking-wider">CLOSED</span>
          </div>
        )}
      </div>

      {/* Ticker Content */}
      <style>{`
        @keyframes slideIn {
          from { opacity: 0; transform: translateY(-10px); }
          to { opacity: 1; transform: translateY(0); }
        }
        .animate-slide-in {
          animation: slideIn 0.3s ease-out forwards;
        }
      `}</style>
      <div className="relative">
        {/* Column Headers */}
        <div className="grid grid-cols-3 px-4 py-2 bg-white text-[10px] font-medium text-gray-400 uppercase tracking-wider border-b border-gray-50">
          <div className="text-left">Time</div>
          <div className="text-center">Price</div>
          <div className="text-right">Vol</div>
        </div>

        {/* Scrollable Area */}
        <div className="min-h-[200px] max-h-[300px] overflow-hidden bg-white relative">
          {loading ? (
            <div className="absolute inset-0 flex flex-col items-center justify-center text-gray-400 gap-2">
              <div className="w-5 h-5 border-2 border-gray-200 border-t-indigo-500 rounded-full animate-spin"></div>
              <span className="text-xs">載入中...</span>
            </div>
          ) : !isMarketOpen ? (
            <div className="absolute inset-0 flex flex-col items-center justify-center text-gray-400 gap-2">
              <svg className="w-8 h-8 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-sm font-medium text-gray-500">已收盤</span>
              <span className="text-xs text-gray-400">{marketMessage}</span>
            </div>
          ) : ticks.length === 0 ? (
            <div className="absolute inset-0 flex flex-col items-center justify-center text-gray-300">
              <span className="text-xs">等待成交...</span>
            </div>
          ) : (
            <div className="flex flex-col w-full">
              {ticks.map((tick) => (
                <div 
                  key={tick.id}
                  className="grid grid-cols-3 px-4 py-2.5 border-b border-gray-50 last:border-0 hover:bg-gray-50 transition-colors duration-200 animate-slide-in items-center"
                >
                  {/* Time */}
                  <div className="text-xs font-mono text-gray-400 tabular-nums text-left">
                    {tick.time}
                  </div>
                  
                  {/* Price */}
                  <div className={`text-sm font-bold font-mono text-center flex items-center justify-center gap-1 ${
                    tick.trend === 'up' ? 'text-rose-500' : 
                    tick.trend === 'down' ? 'text-emerald-500' : 'text-gray-600'
                  }`}>
                    {tick.price}
                    {tick.trend === 'up' && <span className="text-[10px]">▲</span>}
                    {tick.trend === 'down' && <span className="text-[10px]">▼</span>}
                  </div>
                  
                  {/* Volume */}
                  <div className="text-xs font-mono text-gray-600 tabular-nums text-right">
                    {tick.volume.toLocaleString()}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
      
      {/* Current Quote Footer (Optional Summary) */}
      {currentQuote && (
        <div className="px-4 py-2 bg-gray-50 border-t border-gray-100 flex justify-between items-center text-xs text-gray-500">
            <span>總量</span>
            <span className="font-mono font-medium text-gray-700">{currentQuote.volume.toLocaleString()}</span>
        </div>
      )}
    </div>
  );
}
