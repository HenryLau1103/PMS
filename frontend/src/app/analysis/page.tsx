'use client';

import React, { useState, useEffect } from 'react';
import StockChart from '@/components/Chart/StockChart';
import ChartControls from '@/components/Chart/ChartControls';
import StockAutocomplete from '@/components/StockAutocomplete';
import { ChartConfig } from '@/types/chart';
import { TaiwanStock } from '@/types/api';

const DEFAULT_CONFIG: ChartConfig = {
  ma5: { enabled: true, period: 5, color: '#2962FF' },
  ma10: { enabled: true, period: 10, color: '#B71C1C' },
  ma20: { enabled: true, period: 20, color: '#FF6D00' },
  rsi: { enabled: false, period: 14, color: '#9C27B0' },
  macd: { enabled: false, period: 0, color: '#00E676' }, // Disabled for MVP
  bb: { enabled: false, period: 20, color: '#2196F3' },
  kdj: { enabled: false, period: 9, color: '#E91E63' }, // Disabled for MVP
};

export default function AnalysisPage() {
  const [selectedSymbol, setSelectedSymbol] = useState<string>('2330');
  const [config, setConfig] = useState<ChartConfig>(DEFAULT_CONFIG);

  const handleSymbolSelect = (stock: TaiwanStock) => {
    setSelectedSymbol(stock.symbol);
  };

  return (
    <div className="min-h-screen bg-[#121212] text-gray-100 flex flex-col">
      {/* Top Bar */}
      <header className="bg-[#1e1e1e] border-b border-gray-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold bg-gradient-to-r from-blue-400 to-teal-400 bg-clip-text text-transparent">
            PSM Technical Analysis
          </h1>
          <div className="h-6 w-px bg-gray-700 mx-2"></div>
          <div className="w-64 relative">
             {/* Wrapper for StockAutocomplete to handle dark mode context somewhat gracefully */}
             <div className="text-gray-900">
                <StockAutocomplete 
                  value={selectedSymbol} 
                  onSelect={handleSymbolSelect}
                  placeholder="搜尋股票 (2330)"
                  className="w-full px-3 py-2 bg-gray-100 border-none rounded focus:ring-2 focus:ring-primary-500"
                />
             </div>
          </div>
        </div>
        
        <div className="flex items-center gap-4">
           <div className="text-sm text-gray-400">
              <span className="inline-block w-2 h-2 rounded-full bg-green-500 mr-2 animate-pulse"></span>
              Market Data Ready
           </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 flex overflow-hidden">
        {/* Chart Area */}
        <div className="flex-1 relative bg-[#1e1e1e]">
          <StockChart 
            symbol={selectedSymbol} 
            config={config} 
            className="w-full h-full"
          />
        </div>

        {/* Right Sidebar Controls */}
        <aside className="w-72 bg-[#181818] border-l border-gray-800 p-4 overflow-y-auto">
          <ChartControls 
            config={config} 
            onConfigChange={setConfig} 
            className="bg-transparent border-0 p-0"
          />
          
          <div className="mt-8 p-4 bg-gray-800/50 rounded-lg text-xs text-gray-500">
            <h4 className="font-semibold text-gray-400 mb-2">Instructions</h4>
            <ul className="list-disc pl-4 space-y-1">
              <li>Use the search bar to find stocks.</li>
              <li>Toggle indicators to overlay on chart.</li>
              <li>Scroll to zoom, drag to pan.</li>
            </ul>
          </div>
        </aside>
      </main>
    </div>
  );
}
