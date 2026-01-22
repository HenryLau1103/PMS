'use client';

import { useState } from 'react';
import TransactionForm from '@/components/TransactionForm';
import PortfolioDashboard from '@/components/PortfolioDashboard';

// Demo portfolio ID (matches database seed data)
const DEMO_PORTFOLIO_ID = '00000000-0000-0000-0000-000000000011';

export default function Home() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleTransactionSuccess = () => {
    // Trigger portfolio refresh when new transaction is added
    setRefreshTrigger((prev) => prev + 1);
  };

  return (
    <main className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                台股智能投資組合管理系統
              </h1>
              <p className="mt-1 text-sm text-gray-500">
                Portfolio Stock Management - Phase 1 MVP
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                Phase 1: 核心功能
              </span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Welcome Section */}
        <div className="bg-gradient-to-r from-primary-600 to-blue-600 rounded-lg shadow-lg p-6 mb-8 text-white">
          <h2 className="text-2xl font-bold mb-2">歡迎使用 PSM 系統</h2>
          <p className="text-primary-100">
            開始記錄您的台股交易，系統將自動計算持倉、成本與損益
          </p>
          <div className="mt-4 flex space-x-4 text-sm">
            <div className="flex items-center">
              <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clipRule="evenodd"
                />
              </svg>
              自動計算手續費與稅金
            </div>
            <div className="flex items-center">
              <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clipRule="evenodd"
                />
              </svg>
              FIFO 成本會計
            </div>
            <div className="flex items-center">
              <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clipRule="evenodd"
                />
              </svg>
              即時持倉追蹤
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Left Column: Transaction Form */}
          <div>
            <TransactionForm
              portfolioId={DEMO_PORTFOLIO_ID}
              onSuccess={handleTransactionSuccess}
            />
          </div>

          {/* Right Column: Portfolio Dashboard */}
          <div>
            <PortfolioDashboard
              portfolioId={DEMO_PORTFOLIO_ID}
              refreshTrigger={refreshTrigger}
            />
          </div>
        </div>

        {/* Footer Info */}
        <div className="mt-12 bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">系統功能狀態</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
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
              <span className="w-3 h-3 bg-yellow-500 rounded-full mr-3"></span>
              <div>
                <p className="font-medium">技術分析</p>
                <p className="text-sm text-gray-500">Phase 2 開發中</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
