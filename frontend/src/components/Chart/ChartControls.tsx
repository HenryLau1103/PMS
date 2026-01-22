'use client';

import React from 'react';
import { ChartConfig, IndicatorConfig } from '@/types/chart';
import { Switch } from '@headlessui/react'; // Ensure this is installed or use standard
import { clsx } from 'clsx';

// Since I cannot verify if @headlessui/react is installed (it is in package.json), I will use it.
// "dependencies": { "@headlessui/react": "^1.7.18" }

interface ChartControlsProps {
  config: ChartConfig;
  onConfigChange: (newConfig: ChartConfig) => void;
  className?: string;
}

export default function ChartControls({ config, onConfigChange, className }: ChartControlsProps) {
  
  const toggleIndicator = (key: keyof ChartConfig) => {
    const newConfig = { ...config };
    newConfig[key] = {
      ...newConfig[key],
      enabled: !newConfig[key].enabled,
    };
    onConfigChange(newConfig);
  };

  const updatePeriod = (key: keyof ChartConfig, period: number) => {
     const newConfig = { ...config };
     newConfig[key] = {
       ...newConfig[key],
       period: period,
     };
     onConfigChange(newConfig);
  };

  return (
    <div className={`bg-gray-800 p-4 rounded-lg border border-gray-700 text-white ${className}`}>
      <h3 className="text-lg font-semibold mb-4 text-gray-200">指標設定</h3>
      
      <div className="space-y-4">
        {/* Moving Averages */}
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-400 uppercase tracking-wider">均線 (Moving Averages)</div>
          <ControlRow 
             label="MA5" 
             color={config.ma5.color} 
             enabled={config.ma5.enabled} 
             period={config.ma5.period}
             onToggle={() => toggleIndicator('ma5')}
             onPeriodChange={(v) => updatePeriod('ma5', v)}
          />
          <ControlRow 
             label="MA10" 
             color={config.ma10.color} 
             enabled={config.ma10.enabled} 
             period={config.ma10.period}
             onToggle={() => toggleIndicator('ma10')}
             onPeriodChange={(v) => updatePeriod('ma10', v)}
          />
          <ControlRow 
             label="MA20" 
             color={config.ma20.color} 
             enabled={config.ma20.enabled} 
             period={config.ma20.period}
             onToggle={() => toggleIndicator('ma20')}
             onPeriodChange={(v) => updatePeriod('ma20', v)}
          />
        </div>

        <div className="border-t border-gray-700 my-2"></div>

        {/* Technical Indicators */}
        <div className="space-y-2">
          <div className="text-xs font-medium text-gray-400 uppercase tracking-wider">技術指標 (Indicators)</div>
          <ControlRow 
             label="Bollinger Bands" 
             color={config.bb.color} 
             enabled={config.bb.enabled} 
             period={config.bb.period}
             onToggle={() => toggleIndicator('bb')}
             onPeriodChange={(v) => updatePeriod('bb', v)}
          />
          <ControlRow 
             label="RSI" 
             color={config.rsi.color} 
             enabled={config.rsi.enabled} 
             period={config.rsi.period}
             onToggle={() => toggleIndicator('rsi')}
             onPeriodChange={(v) => updatePeriod('rsi', v)}
          />
          {/* 
          <ControlRow 
             label="MACD" 
             color={config.macd.color} 
             enabled={config.macd.enabled} 
             period={config.macd.period} // MACD has multiple params, simplistic here
             onToggle={() => toggleIndicator('macd')}
             // onPeriodChange handled differently for MACD
          />
          */}
        </div>
      </div>
    </div>
  );
}

interface ControlRowProps {
  label: string;
  color?: string;
  enabled: boolean;
  period?: number;
  onToggle: () => void;
  onPeriodChange?: (val: number) => void;
}

function ControlRow({ label, color, enabled, period, onToggle, onPeriodChange }: ControlRowProps) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center space-x-3">
        <Switch
          checked={enabled}
          onChange={onToggle}
          className={clsx(
            enabled ? 'bg-primary-600' : 'bg-gray-600',
            'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2'
          )}
        >
          <span
            aria-hidden="true"
            className={clsx(
              enabled ? 'translate-x-4' : 'translate-x-0',
              'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out'
            )}
          />
        </Switch>
        <div className="flex items-center space-x-2">
          <span className="text-sm font-medium text-gray-300">{label}</span>
          {color && <span className="w-2 h-2 rounded-full" style={{ backgroundColor: color }}></span>}
        </div>
      </div>
      
      {period !== undefined && enabled && (
        <div className="flex items-center space-x-1">
            <span className="text-xs text-gray-500">N=</span>
            <input 
                type="number" 
                value={period} 
                onChange={(e) => onPeriodChange?.(parseInt(e.target.value) || 0)}
                className="w-12 px-1 py-0.5 bg-gray-700 border border-gray-600 rounded text-xs text-center text-white focus:border-primary-500 focus:outline-none"
            />
        </div>
      )}
    </div>
  );
}
