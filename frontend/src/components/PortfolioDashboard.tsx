'use client';

import { useEffect, useState } from 'react';
import { getPositions, getStock } from '@/lib/api';
import type { Position, TaiwanStock, RealtimeQuote } from '@/types/api';
import { formatCurrency, formatPercentage, getPnLColorClass } from '@/lib/utils';
import RealtimePriceCell from './RealtimePriceCell';

interface PortfolioDashboardProps {
  portfolioId: string;
  refreshTrigger?: number;
}

interface PositionWithName extends Position {
  stock_name?: string;
  current_price?: number;
  unrealized_pnl?: number;
  unrealized_pnl_pct?: number;
}

export default function PortfolioDashboard({ portfolioId, refreshTrigger = 0 }: PortfolioDashboardProps) {
  const [positions, setPositions] = useState<PositionWithName[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPositions();
  }, [portfolioId, refreshTrigger]);

  const fetchPositions = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getPositions(portfolioId);
      
      // Fetch stock names for each position
      const positionsWithNames = await Promise.all(
        (data || []).map(async (position) => {
          try {
            const stock = await getStock(position.symbol);
            return { ...position, stock_name: stock.name };
          } catch {
            // If stock not found, just use symbol
            return { ...position, stock_name: undefined };
          }
        })
      );
      
      setPositions(positionsWithNames);
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || '載入持倉失敗');
      setPositions([]);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-white shadow-md rounded-lg p-6">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/4 mb-6"></div>
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-16 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white shadow-md rounded-lg p-6">
        <div className="text-center text-red-600">
          <p>{error}</p>
          <button
            onClick={fetchPositions}
            className="mt-4 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
          >
            重試
          </button>
        </div>
      </div>
    );
  }

  // Calculate portfolio summary
  const totalCost = positions?.reduce((sum, pos) => sum + parseFloat(pos.total_cost), 0) || 0;
  const totalPositions = positions?.length || 0;

  return (
    <div className="bg-white shadow-md rounded-lg p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">投資組合持倉</h2>
        <button
          onClick={fetchPositions}
          className="px-4 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
        >
          刷新
        </button>
      </div>

      {/* Summary Cards */}
      <div className="space-y-4 mb-6">
        {/* Row 1: 持倉數 & 持股類型 */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-primary-50 border border-primary-200 rounded-lg p-4">
            <p className="text-sm text-primary-600 font-medium mb-2">總持倉數</p>
            <p className="text-3xl font-bold text-primary-800">{totalPositions}</p>
          </div>
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
            <p className="text-sm text-gray-600 font-medium mb-2">持股類型</p>
            <p className="text-3xl font-bold text-gray-800">台股</p>
          </div>
        </div>
        
        {/* Row 2: 總成本 (單獨一列) */}
        <div className="bg-gradient-to-r from-blue-50 to-blue-100 border border-blue-200 rounded-lg p-6">
          <p className="text-sm text-blue-600 font-medium mb-2">總投資成本</p>
          <p className="text-4xl font-bold text-blue-800">{formatCurrency(totalCost)}</p>
        </div>
      </div>

      {/* Positions Table */}
      {!positions || positions.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          <p className="text-lg">目前沒有持倉</p>
          <p className="text-sm mt-2">請先新增交易記錄</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  股票代號
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  持有股數
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  平均成本
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  現價
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  總成本
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  更新時間
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {positions.map((position) => (
                <tr key={position.symbol} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex flex-col">
                      <div className="text-sm font-medium text-gray-900">{position.symbol}</div>
                      {position.stock_name && (
                        <div className="text-xs text-gray-500">{position.stock_name}</div>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="text-sm text-gray-900">
                      {parseFloat(position.total_quantity).toLocaleString('zh-TW')}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="text-sm text-gray-900">
                      {formatCurrency(position.avg_cost_per_share)}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <RealtimePriceCell 
                      symbol={position.symbol} 
                      showChange={true}
                      showLimitAlert={true}
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="text-sm font-medium text-gray-900">
                      {formatCurrency(position.total_cost)}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm text-gray-500">
                    {new Date(position.last_updated).toLocaleDateString('zh-TW')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
