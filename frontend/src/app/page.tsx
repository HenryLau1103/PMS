'use client';

import { useState } from 'react';
import TransactionForm from '@/components/TransactionForm';
import PortfolioDashboard from '@/components/PortfolioDashboard';
import DataSyncPanel from '@/components/DataSyncPanel';
import StockChart from '@/components/Chart/StockChart';
import ChartControls from '@/components/Chart/ChartControls';
import type { ChartConfig } from '@/types/chart';

// Demo portfolio ID (matches database seed data)
const DEMO_PORTFOLIO_ID = '00000000-0000-0000-0000-000000000011';

export default function Home() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [selectedSymbol, setSelectedSymbol] = useState('2330'); // Default to TSMC
  const [chartConfig, setChartConfig] = useState<ChartConfig>({
    ma5: { enabled: false, period: 5, color: '#2196F3' },
    ma10: { enabled: false, period: 10, color: '#FF9800' },
    ma20: { enabled: false, period: 20, color: '#9C27B0' },
    rsi: { enabled: false, period: 14, color: '#00BCD4' },
    macd: { enabled: false, color: '#4CAF50' },
    bb: { enabled: false, period: 20, color: '#E91E63' },
    kdj: { enabled: false, period: 9, color: '#FFC107' },
  });

  const handleTransactionSuccess = () => {
    // Trigger portfolio refresh when new transaction is added
    setRefreshTrigger((prev) => prev + 1);
  };

  return (
    <main className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-[1920px] mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                台股智能投資組合管理系統
              </h1>
              <p className="mt-1 text-sm text-gray-500">
                Portfolio Stock Management - 技術分析整合版
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <a
                href="/analysis"
                className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
              >
                完整技術分析頁面
              </a>
              <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                Phase 2: 技術分析完成
              </span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-[1920px] mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Top Section: Transaction Form + Portfolio Dashboard */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Left Column: Transaction Form */}
          <div>
            <TransactionForm
              portfolioId={DEMO_PORTFOLIO_ID}
              onSuccess={handleTransactionSuccess}
            />
          </div>

          {/* Right Column: Portfolio Dashboard + Data Sync Panel */}
          <div className="space-y-6">
            <PortfolioDashboard
              portfolioId={DEMO_PORTFOLIO_ID}
              refreshTrigger={refreshTrigger}
            />
            
            {/* Data Sync Panel - Below Portfolio Dashboard */}
            <DataSyncPanel portfolioId={DEMO_PORTFOLIO_ID} />
          </div>
        </div>

        {/* Bottom Section: Technical Analysis Chart */}
        <div className="grid grid-cols-1 xl:grid-cols-12 gap-6">
          {/* Chart Area (10/12) */}
          <div className="xl:col-span-10">
            <div className="bg-white rounded-lg shadow-lg p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-bold text-gray-900">
                  技術分析圖表
                </h2>
                <div className="flex items-center space-x-2">
                  <label className="text-sm text-gray-600">股票代號:</label>
                  <input
                    type="text"
                    value={selectedSymbol}
                    onChange={(e) => setSelectedSymbol(e.target.value.toUpperCase())}
                    className="px-3 py-1 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    placeholder="2330"
                  />
                </div>
              </div>
              
              {/* Chart Component */}
              <div className="h-[600px]">
                <StockChart
                  symbol={selectedSymbol}
                  config={chartConfig}
                />
              </div>
            </div>
          </div>

          {/* Chart Controls (2/12) */}
          <div className="xl:col-span-2">
            <ChartControls
              config={chartConfig}
              onConfigChange={setChartConfig}
            />
          </div>
        </div>

        {/* Footer Info */}
        <div className="mt-8 bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">系統功能狀態</h3>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="flex items-center">
              <span className="w-3 h-3 bg-green-500 rounded-full mr-3"></span>
              <div>
                <p className="font-medium">交易記錄</p>
                <p className="text-sm text-gray-500">已完成</p>
              </div>
            </div>
            <div className="flex items-center">
              <span className="w-3 h-3 bg-green-500 rounded-full mr-3"></span>
              <div>
                <p className="font-medium">持倉管理</p>
                <p className="text-sm text-gray-500">已完成</p>
              </div>
            </div>
            <div className="flex items-center">
              <span className="w-3 h-3 bg-green-500 rounded-full mr-3"></span>
              <div>
                <p className="font-medium">技術分析</p>
                <p className="text-sm text-gray-500">Phase 2 已完成</p>
              </div>
            </div>
            <div className="flex items-center">
              <span className="w-3 h-3 bg-green-500 rounded-full mr-3"></span>
              <div>
                <p className="font-medium">K線圖表</p>
                <p className="text-sm text-gray-500">即時顯示</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
