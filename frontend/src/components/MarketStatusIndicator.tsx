'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { MarketStatus } from '@/types/api';
import { getMarketStatus, getRealtimeWS } from '@/lib/realtimeApi';

const STATUS_CONFIG = {
  open: {
    color: 'bg-green-500',
    shadow: 'shadow-[0_0_10px_rgba(34,197,94,0.5)]',
    pulse: 'animate-pulse',
    label: '市場開盤',
  },
  pre_market: {
    color: 'bg-yellow-500',
    shadow: 'shadow-[0_0_10px_rgba(234,179,8,0.5)]',
    pulse: 'animate-pulse',
    label: '盤前試撮',
  },
  after_hours: {
    color: 'bg-yellow-500',
    shadow: 'shadow-[0_0_10px_rgba(234,179,8,0.5)]',
    pulse: 'animate-pulse',
    label: '盤後交易',
  },
  closed: {
    color: 'bg-gray-500',
    shadow: '',
    pulse: '',
    label: '已收盤',
  },
  holiday: {
    color: 'bg-gray-500',
    shadow: '',
    pulse: '',
    label: '休市日',
  },
};

export default function MarketStatusIndicator() {
  const [status, setStatus] = useState<MarketStatus | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [displayTime, setDisplayTime] = useState<string>('--:--:--');
  const [timeOffset, setTimeOffset] = useState<number>(0);
  const wsRef = useRef<ReturnType<typeof getRealtimeWS> | null>(null);

  // Calculate server time offset when status updates
  useEffect(() => {
    if (status?.server_time) {
      const serverTime = new Date(status.server_time).getTime();
      const localTime = Date.now();
      setTimeOffset(serverTime - localTime);
    }
  }, [status?.server_time]);

  // Tick local clock using server time offset
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      const currentServerTime = new Date(now + timeOffset);
      setDisplayTime(
        new Intl.DateTimeFormat('en-GB', {
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          hour12: false,
          timeZone: 'Asia/Taipei',
        }).format(currentServerTime)
      );
    }, 1000);
    return () => clearInterval(interval);
  }, [timeOffset]);

  const fetchStatus = useCallback(async () => {
    try {
      const data = await getMarketStatus();
      setStatus(data);
    } catch (error) {
      console.error('Failed to fetch market status:', error);
    }
  }, []);

  // Initialize WebSocket and polling
  useEffect(() => {
    fetchStatus();

    const pollInterval = setInterval(fetchStatus, 30000);

    const ws = getRealtimeWS();
    wsRef.current = ws;

    ws.connect();
    setIsConnected(ws.isConnected());

    ws.onConnected(() => {
      setIsConnected(true);
    });

    ws.onDisconnected(() => {
      setIsConnected(false);
    });

    ws.onMarketStatus((newStatus) => {
      setStatus(newStatus);
    });

    return () => {
      clearInterval(pollInterval);
    };
  }, [fetchStatus]);

  const currentConfig = status 
    ? STATUS_CONFIG[status.status] || STATUS_CONFIG.closed
    : STATUS_CONFIG.closed;

  return (
    <div className="flex items-center gap-3 px-3 py-1.5 bg-slate-800 rounded-full border border-slate-600 shadow-lg select-none transition-all hover:border-slate-500">
      
      {/* Status Dot & Label */}
      <div className="flex items-center gap-2 pr-3 border-r border-slate-600">
        <div className="relative flex items-center justify-center w-2.5 h-2.5">
          <span className={`absolute inline-flex h-full w-full rounded-full opacity-75 ${currentConfig.color} ${currentConfig.pulse}`}></span>
          <span className={`relative inline-flex rounded-full w-2.5 h-2.5 ${currentConfig.color} ${currentConfig.shadow}`}></span>
        </div>
        <span className="text-xs font-semibold tracking-wide text-slate-100">
          {status ? currentConfig.label : '載入中...'}
        </span>
      </div>

      {/* Time & Connection */}
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-1.5 text-slate-300">
          <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span className="font-mono text-xs tracking-wider font-medium text-slate-200">
            {displayTime}
          </span>
        </div>

        {/* Connection Indicator */}
        <div 
          className="group relative flex items-center justify-center"
          title={isConnected ? "WebSocket Connected" : "Disconnected"}
        >
          <div className={`w-2 h-2 rounded-full transition-colors duration-300 ${isConnected ? 'bg-emerald-400' : 'bg-rose-400'}`}></div>
          
          {/* Tooltip on hover */}
          <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 text-[10px] text-white bg-slate-900 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none border border-slate-700 z-50">
            {isConnected ? '已連線' : '斷線'}
          </div>
        </div>
      </div>
    </div>
  );
}
