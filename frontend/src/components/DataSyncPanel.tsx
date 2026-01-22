'use client';

import { useState, useEffect } from 'react';
import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface SyncStatus {
  is_running: boolean;
  mode: string;
  total_days: number;
  processed_days: number;
  total_symbols: number;
  processed_count: number;
  success_count: number;
  failed_count: number;
  skipped_count: number;
  current_date: string;
  current_symbol: string;
  started_at?: string;
  completed_at?: string;
  error_message?: string;
  failed_dates?: string[];
  failed_symbols?: string[];
  estimated_time?: string;
}

interface SyncInfo {
  first_synced_date: string;
  last_synced_date: string;
  synced_days_count: number;
  gaps_count: number;
}

interface DataSyncPanelProps {
  portfolioId: string;
}

export default function DataSyncPanel({ portfolioId }: DataSyncPanelProps) {
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [syncInfo, setSyncInfo] = useState<SyncInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showFailedItems, setShowFailedItems] = useState(false);

  // Fetch sync info on mount
  useEffect(() => {
    const fetchSyncInfo = async () => {
      try {
        const response = await axios.get(`${API_BASE_URL}/api/v1/market/bulk-sync/info`);
        if (response.data.success) {
          setSyncInfo(response.data.info);
        }
      } catch (err) {
        console.error('Failed to fetch sync info:', err);
      }
    };
    fetchSyncInfo();
  }, []);

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

  // Refresh sync info after sync completes
  useEffect(() => {
    if (syncStatus && !syncStatus.is_running && syncStatus.completed_at && !syncStatus.completed_at.startsWith('0001')) {
      const fetchSyncInfo = async () => {
        try {
          const response = await axios.get(`${API_BASE_URL}/api/v1/market/bulk-sync/info`);
          if (response.data.success) {
            setSyncInfo(response.data.info);
          }
        } catch (err) {
          console.error('Failed to fetch sync info:', err);
        }
      };
      fetchSyncInfo();
    }
  }, [syncStatus?.is_running, syncStatus?.completed_at]);

  const handleStartSync = async (mode: 'incremental' | 'full') => {
    setLoading(true);
    setError(null);

    try {
      const today = new Date();
      let startDate: string;
      
      if (mode === 'incremental' && syncInfo?.last_synced_date) {
        // Incremental: Start from last synced date (to fill any gaps and get new data)
        const lastDate = new Date(syncInfo.last_synced_date);
        lastDate.setDate(lastDate.getDate() + 1); // Start from next day
        startDate = lastDate.toISOString().split('T')[0];
        
        // If last synced date is today or future, no need to sync
        if (lastDate >= today) {
          setError('資料已是最新');
          setLoading(false);
          return;
        }
      } else {
        // Full: Last 2 years
        const twoYearsAgo = new Date();
        twoYearsAgo.setFullYear(today.getFullYear() - 2);
        startDate = twoYearsAgo.toISOString().split('T')[0];
      }
      
      const endDate = today.toISOString().split('T')[0];

      const response = await axios.post(`${API_BASE_URL}/api/v1/market/bulk-sync/start`, {
        portfolio_id: portfolioId,
        start_date: startDate,
        end_date: endDate,
        priority_holdings: false,
        skip_synced: true,
      });

      if (response.data.success) {
        setSyncStatus(prev => ({
          ...prev,
          is_running: true,
          mode: 'date',
          total_days: 0,
          processed_days: 0,
          success_count: 0,
          failed_count: 0,
          skipped_count: 0,
          processed_count: 0,
          current_date: '計算中...',
          started_at: new Date().toISOString(),
          failed_dates: [],
          estimated_time: '計算中...',
        } as SyncStatus));
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

  // Progress based on mode
  const isDateMode = syncStatus?.mode === 'date';
  const progress = isDateMode
    ? (syncStatus && syncStatus.total_days > 0 ? (syncStatus.processed_days / syncStatus.total_days) * 100 : 0)
    : (syncStatus && syncStatus.total_symbols > 0 ? (syncStatus.processed_count / syncStatus.total_symbols) * 100 : 0);

  const formatDuration = (start?: string, end?: string) => {
    if (!start) return '';
    const startTime = new Date(start).getTime();
    
    const isValidEndTime = end && !end.startsWith('0001-01-01');
    const endTime = isValidEndTime ? new Date(end).getTime() : Date.now();
    
    const duration = Math.floor((endTime - startTime) / 1000);
    if (duration < 0) return '';
    
    const hours = Math.floor(duration / 3600);
    const minutes = Math.floor((duration % 3600) / 60);
    const seconds = duration % 60;
    
    if (hours > 0) {
      return `${hours}時${minutes}分${seconds}秒`;
    }
    return `${minutes}分${seconds}秒`;
  };

  const failedItems = isDateMode ? syncStatus?.failed_dates : syncStatus?.failed_symbols;
  const failedLabel = isDateMode ? '日期' : '股票';

  // Calculate if incremental sync is available
  const canIncrementalSync = syncInfo?.last_synced_date && 
    new Date(syncInfo.last_synced_date) < new Date();

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold text-gray-900">市場數據同步</h2>
        <div className="flex items-center gap-2">
          {syncStatus?.mode === 'date' && (
            <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-purple-100 text-purple-800">
              快速模式
            </span>
          )}
          {syncStatus?.is_running && (
            <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800 animate-pulse">
              <span className="w-2 h-2 bg-blue-600 rounded-full mr-2"></span>
              同步中...
            </span>
          )}
        </div>
      </div>

      {/* Sync Info */}
      {syncInfo && !syncStatus?.is_running && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div>
              <span className="text-gray-500">資料範圍:</span>
              <span className="ml-2 font-medium">
                {syncInfo.first_synced_date || '無'} ~ {syncInfo.last_synced_date || '無'}
              </span>
            </div>
            <div>
              <span className="text-gray-500">已同步天數:</span>
              <span className="ml-2 font-medium text-green-600">{syncInfo.synced_days_count} 天</span>
              {syncInfo.gaps_count > 0 && (
                <span className="ml-2 text-orange-500">({syncInfo.gaps_count} 個缺口)</span>
              )}
            </div>
          </div>
        </div>
      )}

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-red-800 text-sm">{error}</p>
        </div>
      )}

      {/* Control Buttons */}
      <div className="flex space-x-3 mb-6">
        {canIncrementalSync && (
          <button
            onClick={() => handleStartSync('incremental')}
            disabled={loading || syncStatus?.is_running}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors text-sm font-medium flex items-center justify-center gap-2"
          >
            {loading ? (
              <>
                <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                啟動中...
              </>
            ) : (
              <>
                🔄 更新至今天
                <span className="text-xs opacity-75">
                  (從 {syncInfo?.last_synced_date})
                </span>
              </>
            )}
          </button>
        )}
        <button
          onClick={() => handleStartSync('full')}
          disabled={loading || syncStatus?.is_running}
          className={`${canIncrementalSync ? 'flex-1' : 'w-full'} px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors text-sm font-medium flex items-center justify-center gap-2`}
        >
          {loading ? (
            <>
              <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
              啟動中...
            </>
          ) : (
            canIncrementalSync ? '📦 完整同步 (2年)' : '📦 同步所有股票 (2年)'
          )}
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
      {syncStatus && syncStatus.is_running && (
        <div className="space-y-4">
          {/* Progress Bar */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-700">
                {isDateMode ? (
                  <>進度: {syncStatus.processed_days} / {syncStatus.total_days} 天</>
                ) : (
                  <>進度: {syncStatus.processed_count} / {syncStatus.total_symbols}</>
                )}
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
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div className="bg-green-50 rounded-lg p-3">
              <p className="text-xs text-gray-500 mb-1">成功</p>
              <p className="text-xl font-bold text-green-600">{syncStatus.success_count}</p>
            </div>
            <div className="bg-purple-50 rounded-lg p-3">
              <p className="text-xs text-gray-500 mb-1">跳過</p>
              <p className="text-xl font-bold text-purple-600">{syncStatus.skipped_count || 0}</p>
            </div>
            <div className="bg-red-50 rounded-lg p-3">
              <p className="text-xs text-gray-500 mb-1">失敗</p>
              <p className="text-xl font-bold text-red-600">{syncStatus.failed_count}</p>
            </div>
            <div className="bg-blue-50 rounded-lg p-3">
              <p className="text-xs text-gray-500 mb-1">資料筆數</p>
              <p className="text-xl font-bold text-blue-600">{syncStatus.processed_count.toLocaleString()}</p>
            </div>
          </div>

          {/* Current Progress Detail */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 space-y-1">
            {syncStatus.current_date && (
              <p className="text-sm text-blue-800">
                <span className="font-medium">當前日期:</span> {syncStatus.current_date}
              </p>
            )}
            {syncStatus.estimated_time && syncStatus.estimated_time !== 'completed' && (
              <p className="text-sm text-blue-800">
                <span className="font-medium">預估剩餘:</span> {syncStatus.estimated_time}
              </p>
            )}
          </div>

          {/* Duration */}
          {syncStatus.started_at && (
            <div className="text-sm text-gray-600">
              <span className="font-medium">已用時:</span> {formatDuration(syncStatus.started_at, syncStatus.completed_at)}
            </div>
          )}
        </div>
      )}

      {/* Completion/Error States */}
      {syncStatus && !syncStatus.is_running && (
        <div className="space-y-3">
          {/* Failed Items */}
          {syncStatus.failed_count > 0 && failedItems && failedItems.length > 0 && (
            <div>
              <button
                onClick={() => setShowFailedItems(!showFailedItems)}
                className="text-sm text-red-600 hover:text-red-700 font-medium"
              >
                {showFailedItems ? '隱藏' : '顯示'}失敗的{failedLabel} ({syncStatus.failed_count})
              </button>
              {showFailedItems && (
                <div className="mt-2 p-3 bg-red-50 border border-red-200 rounded-md">
                  <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto">
                    {failedItems.slice(0, 50).map((item) => (
                      <span key={item} className="px-2 py-1 bg-red-100 text-red-800 text-xs rounded">
                        {item}
                      </span>
                    ))}
                    {failedItems.length > 50 && (
                      <span className="px-2 py-1 text-red-600 text-xs">...還有 {failedItems.length - 50} 個</span>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Completion Message */}
          {syncStatus.completed_at && !syncStatus.completed_at.startsWith('0001') && !syncStatus.error_message && (
            <div className="bg-green-50 border border-green-200 rounded-lg p-3">
              <p className="text-sm text-green-800">
                ✓ 同步完成 - 成功 {syncStatus.success_count} 天，共 {syncStatus.processed_count.toLocaleString()} 筆資料
              </p>
            </div>
          )}

          {/* Error Message */}
          {syncStatus.error_message && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
              <p className="text-sm text-yellow-800">⚠ {syncStatus.error_message}</p>
            </div>
          )}
        </div>
      )}

      {/* Initial Info */}
      {!syncStatus?.is_running && !syncInfo?.last_synced_date && (
        <div className="text-sm text-gray-500">
          <p className="font-medium text-gray-700 mb-2">🚀 快速同步模式</p>
          <p>• 一次抓取所有股票當日資料</p>
          <p>• 預估時間: ~42 分鐘 (2年資料)</p>
          <p>• 5秒/次請求，避免 API 限制</p>
        </div>
      )}
    </div>
  );
}
