'use client';

import { useState, useEffect } from 'react';
import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface SyncStatus {
  is_running: boolean;
  total_symbols: number;
  processed_count: number;
  success_count: number;
  failed_count: number;
  current_symbol: string;
  started_at?: string;
  completed_at?: string;
  error_message?: string;
  failed_symbols?: string[];
}

interface DataSyncPanelProps {
  portfolioId: string;
}

export default function DataSyncPanel({ portfolioId }: DataSyncPanelProps) {
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showFailedSymbols, setShowFailedSymbols] = useState(false);

  // Poll sync status every 2 seconds when sync is running
  useEffect(() => {
    let interval: NodeJS.Timeout | null = null;

    const fetchStatus = async () => {
      try {
        const response = await axios.get(`${API_BASE_URL}/api/v1/market/bulk-sync/status`);
        if (response.data.success) {
          setSyncStatus(response.data.status);
        }
      } catch (err) {
        console.error('Failed to fetch sync status:', err);
      }
    };

    fetchStatus(); // Initial fetch

    if (syncStatus?.is_running) {
      interval = setInterval(fetchStatus, 2000);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [syncStatus?.is_running]);

  const handleStartSync = async (priorityHoldings: boolean) => {
    setLoading(true);
    setError(null);

    try {
      const response = await axios.post(`${API_BASE_URL}/api/v1/market/bulk-sync/start`, {
        portfolio_id: portfolioId,
        start_date: '2024-01-01',
        end_date: '2024-12-31',
        priority_holdings: priorityHoldings,
      });

      if (response.data.success) {
        // Status will be updated by polling
      }
    } catch (err: any) {
      setError(err.response?.data?.error || '啟動同步失敗');
    } finally {
      setLoading(false);
    }
  };

  const handleStopSync = async () => {
    try {
      await axios.post(`${API_BASE_URL}/api/v1/market/bulk-sync/stop`);
    } catch (err: any) {
      setError(err.response?.data?.error || '停止同步失敗');
    }
  };

  const progress = syncStatus && syncStatus.total_symbols > 0
    ? (syncStatus.processed_count / syncStatus.total_symbols) * 100
    : 0;

  const formatDuration = (start?: string, end?: string) => {
    if (!start) return '';
    const startTime = new Date(start).getTime();
    const endTime = end ? new Date(end).getTime() : Date.now();
    const duration = Math.floor((endTime - startTime) / 1000);
    const minutes = Math.floor(duration / 60);
    const seconds = duration % 60;
    return `${minutes}分${seconds}秒`;
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold text-gray-900">市場數據同步</h2>
        {syncStatus?.is_running && (
          <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800 animate-pulse">
            <span className="w-2 h-2 bg-blue-600 rounded-full mr-2"></span>
            同步中...
          </span>
        )}
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-red-800 text-sm">{error}</p>
        </div>
      )}

      {/* Control Buttons */}
      <div className="flex space-x-3 mb-6">
        <button
          onClick={() => handleStartSync(true)}
          disabled={loading || syncStatus?.is_running}
          className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors text-sm font-medium"
        >
          優先同步持倉股票
        </button>
        <button
          onClick={() => handleStartSync(false)}
          disabled={loading || syncStatus?.is_running}
          className="flex-1 px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors text-sm font-medium"
        >
          同步所有股票
        </button>
        {syncStatus?.is_running && (
          <button
            onClick={handleStopSync}
            className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors text-sm font-medium"
          >
            停止
          </button>
        )}
      </div>

      {/* Progress Display */}
      {syncStatus && (
        <div className="space-y-4">
          {/* Progress Bar */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-700">
                進度: {syncStatus.processed_count} / {syncStatus.total_symbols}
              </span>
              <span className="text-sm font-medium text-gray-700">
                {progress.toFixed(1)}%
              </span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
              <div
                className="bg-blue-600 h-3 rounded-full transition-all duration-300 ease-out"
                style={{ width: `${progress}%` }}
              ></div>
            </div>
          </div>

          {/* Status Grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-gray-50 rounded-lg p-3">
              <p className="text-xs text-gray-500 mb-1">總數</p>
              <p className="text-xl font-bold text-gray-900">{syncStatus.total_symbols}</p>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <p className="text-xs text-gray-500 mb-1">成功</p>
              <p className="text-xl font-bold text-green-600">{syncStatus.success_count}</p>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <p className="text-xs text-gray-500 mb-1">失敗</p>
              <p className="text-xl font-bold text-red-600">{syncStatus.failed_count}</p>
            </div>
            <div className="bg-blue-50 rounded-lg p-3">
              <p className="text-xs text-gray-500 mb-1">已處理</p>
              <p className="text-xl font-bold text-blue-600">{syncStatus.processed_count}</p>
            </div>
          </div>

          {/* Current Symbol */}
          {syncStatus.is_running && syncStatus.current_symbol && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
              <p className="text-sm text-blue-800">
                <span className="font-medium">當前處理:</span> {syncStatus.current_symbol}
              </p>
            </div>
          )}

          {/* Duration */}
          {syncStatus.started_at && (
            <div className="text-sm text-gray-600">
              <span className="font-medium">用時:</span> {formatDuration(syncStatus.started_at, syncStatus.completed_at)}
            </div>
          )}

          {/* Failed Symbols */}
          {syncStatus.failed_count > 0 && syncStatus.failed_symbols && syncStatus.failed_symbols.length > 0 && (
            <div>
              <button
                onClick={() => setShowFailedSymbols(!showFailedSymbols)}
                className="text-sm text-red-600 hover:text-red-700 font-medium"
              >
                {showFailedSymbols ? '隱藏' : '顯示'}失敗的股票 ({syncStatus.failed_count})
              </button>
              {showFailedSymbols && (
                <div className="mt-2 p-3 bg-red-50 border border-red-200 rounded-md">
                  <div className="flex flex-wrap gap-2">
                    {syncStatus.failed_symbols.map((symbol) => (
                      <span
                        key={symbol}
                        className="px-2 py-1 bg-red-100 text-red-800 text-xs rounded"
                      >
                        {symbol}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Completion Message */}
          {!syncStatus.is_running && syncStatus.completed_at && (
            <div className="bg-green-50 border border-green-200 rounded-lg p-3">
              <p className="text-sm text-green-800">
                ✓ 同步完成 - 成功 {syncStatus.success_count} 支股票
              </p>
            </div>
          )}

          {/* Error Message */}
          {syncStatus.error_message && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
              <p className="text-sm text-yellow-800">
                ⚠ {syncStatus.error_message}
              </p>
            </div>
          )}
        </div>
      )}

      {/* Info */}
      {!syncStatus && (
        <div className="text-sm text-gray-500">
          <p>點擊按鈕開始同步2024年市場數據</p>
          <p className="mt-2">• 優先同步: 先同步您的持倉股票，再同步其他股票</p>
          <p>• 全部同步: 同步所有1,943支台股 (約需2-3小時)</p>
        </div>
      )}
    </div>
  );
}
