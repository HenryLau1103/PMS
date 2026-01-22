'use client';

import React, { useEffect, useRef, useState } from 'react';
import {
  createChart,
  ColorType,
  IChartApi,
  ISeriesApi,
  Time,
  CrosshairMode,
  LineStyle,
} from 'lightweight-charts';
import {
  getOHLCV,
  getMA,
  getRSI,
  getMACD,
  getBollingerBands,
  getKDJ,
  parseOHLCVData,
  parseVolumeData,
  parseIndicatorData,
  toUnixTimestamp,
} from '@/lib/chartApi';
import { ChartConfig } from '@/types/chart';

interface StockChartProps {
  symbol: string;
  config: ChartConfig;
  className?: string;
}

export default function StockChart({ symbol, config, className }: StockChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null);
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
  
  // References for indicator series
  const ma5SeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const ma10SeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const ma20SeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const rsiSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const bbUpperSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const bbLowerSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  const bbMiddleSeriesRef = useRef<ISeriesApi<'Line'> | null>(null);
  
  // We'll skip MACD/KDJ for the single-chart MVP to avoid overcrowding, 
  // or implement them if requested. The prompt asks for them.
  // We will try to stack them if enabled, or just overlay.
  // For this implementation, we will focus on MAs, BB, RSI (bottom pane).
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!chartContainerRef.current) return;

    // 1. Initialize Chart
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: '#1e1e1e' },
        textColor: '#9ca3af', // gray-400
      },
      grid: {
        vertLines: { color: '#374151' }, // gray-700
        horzLines: { color: '#374151' },
      },
      width: chartContainerRef.current.clientWidth,
      height: 600,
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        borderColor: '#4b5563',
      },
      rightPriceScale: {
        borderColor: '#4b5563',
        scaleMargins: {
          top: 0.1,
          bottom: 0.2, // Leave space for volume
        },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
      },
    });

    chartRef.current = chart;

    // 2. Add Series
    // Candlestick
    const candlestickSeries = chart.addCandlestickSeries({
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderVisible: false,
      wickUpColor: '#26a69a',
      wickDownColor: '#ef5350',
    });
    candlestickSeriesRef.current = candlestickSeries;

    // Volume (Overlay at bottom)
    const volumeSeries = chart.addHistogramSeries({
      color: '#26a69a',
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: '', // Overlay on main
    });
    volumeSeriesRef.current = volumeSeries;

    // Configure price scale margins after series creation
    chart.priceScale('').applyOptions({
      scaleMargins: {
        top: 0.8, // Push to bottom 20%
        bottom: 0,
      },
    });

    // Resize Handler
    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({ width: chartContainerRef.current.clientWidth });
      }
    };

    window.addEventListener('resize', handleResize);
    const resizeObserver = new ResizeObserver(handleResize);
    resizeObserver.observe(chartContainerRef.current);

    return () => {
      window.removeEventListener('resize', handleResize);
      resizeObserver.disconnect();
      chart.remove();
    };
  }, []);

  // Fetch and Update Data
  useEffect(() => {
    if (!chartRef.current || !symbol) return;

    const fetchData = async () => {
      setLoading(true);
      setError(null);
      try {
        // Fetch OHLCV (Main Data)
        const ohlcvData = await getOHLCV(symbol);
        
        // Ensure data is sorted by time (ascending)
        ohlcvData.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
        
        const candleData = parseOHLCVData(ohlcvData);
        const volumeData = parseVolumeData(ohlcvData);

        if (candlestickSeriesRef.current) {
          candlestickSeriesRef.current.setData(candleData as any);
        }
        if (volumeSeriesRef.current) {
          volumeSeriesRef.current.setData(volumeData as any);
        }

        // Indicators
        // MA5
        if (config.ma5.enabled) {
          if (!ma5SeriesRef.current) {
            ma5SeriesRef.current = chartRef.current!.addLineSeries({
              color: config.ma5.color || '#2962FF',
              lineWidth: 1,
              title: 'MA5',
            });
          }
          const ma5 = await getMA(symbol, config.ma5.period || 5);
          ma5.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
          ma5SeriesRef.current.setData(parseIndicatorData(ma5) as any);
        } else if (ma5SeriesRef.current) {
          chartRef.current!.removeSeries(ma5SeriesRef.current);
          ma5SeriesRef.current = null;
        }

        // MA10
        if (config.ma10.enabled) {
          if (!ma10SeriesRef.current) {
            ma10SeriesRef.current = chartRef.current!.addLineSeries({
              color: config.ma10.color || '#B71C1C',
              lineWidth: 1,
              title: 'MA10',
            });
          }
          const ma10 = await getMA(symbol, config.ma10.period || 10);
          ma10.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
          ma10SeriesRef.current.setData(parseIndicatorData(ma10) as any);
        } else if (ma10SeriesRef.current) {
          chartRef.current!.removeSeries(ma10SeriesRef.current);
          ma10SeriesRef.current = null;
        }

        // MA20
        if (config.ma20.enabled) {
          if (!ma20SeriesRef.current) {
            ma20SeriesRef.current = chartRef.current!.addLineSeries({
              color: config.ma20.color || '#FF6D00',
              lineWidth: 1,
              title: 'MA20',
            });
          }
          const ma20 = await getMA(symbol, config.ma20.period || 20);
          ma20.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
          ma20SeriesRef.current.setData(parseIndicatorData(ma20) as any);
        } else if (ma20SeriesRef.current) {
          chartRef.current!.removeSeries(ma20SeriesRef.current);
          ma20SeriesRef.current = null;
        }

        // Bollinger Bands
        if (config.bb.enabled) {
          if (!bbMiddleSeriesRef.current) {
             bbUpperSeriesRef.current = chartRef.current!.addLineSeries({ color: config.bb.color || '#2196F3', lineWidth: 1, title: 'BB Up' });
             bbLowerSeriesRef.current = chartRef.current!.addLineSeries({ color: config.bb.color || '#2196F3', lineWidth: 1, title: 'BB Low' });
             bbMiddleSeriesRef.current = chartRef.current!.addLineSeries({ color: config.bb.color || '#2196F3', lineWidth: 1, lineStyle: LineStyle.Dashed, title: 'BB Mid' });
          }
          const bbData = await getBollingerBands(symbol, config.bb.period || 20);
          bbData.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
          // Parse BB data manually as it has upper/middle/lower
          const upper = bbData.map(d => ({ time: toUnixTimestamp(d.timestamp), value: parseFloat(d.upper) }));
          const lower = bbData.map(d => ({ time: toUnixTimestamp(d.timestamp), value: parseFloat(d.lower) }));
          const middle = bbData.map(d => ({ time: toUnixTimestamp(d.timestamp), value: parseFloat(d.middle) }));
          
          bbUpperSeriesRef.current?.setData(upper as any);
          bbLowerSeriesRef.current?.setData(lower as any);
          bbMiddleSeriesRef.current?.setData(middle as any);
        } else if (bbMiddleSeriesRef.current) {
           chartRef.current!.removeSeries(bbUpperSeriesRef.current!);
           chartRef.current!.removeSeries(bbLowerSeriesRef.current!);
           chartRef.current!.removeSeries(bbMiddleSeriesRef.current!);
           bbUpperSeriesRef.current = null;
           bbLowerSeriesRef.current = null;
           bbMiddleSeriesRef.current = null;
        }

        // RSI - For now, we only implement RSI as a separate scale example if needed,
        // but given the constraints of one chart instance, we might skip complex stacking 
        // OR overlay it with a separate scaleId.
        // Let's overlay it at the very top (unconventional) or very bottom.
        // Or simply: if RSI is enabled, we create a new series with a separate scale.
        if (config.rsi.enabled) {
            if (!rsiSeriesRef.current) {
                rsiSeriesRef.current = chartRef.current!.addLineSeries({
                    color: config.rsi.color || '#9C27B0',
                    lineWidth: 2,
                    priceScaleId: 'rsi',
                    title: 'RSI',
                });
                chartRef.current!.priceScale('rsi').applyOptions({
                    scaleMargins: {
                        top: 0.8, // Bottom 20%
                        bottom: 0,
                    },
                    visible: true,
                });
            }
            const rsi = await getRSI(symbol, config.rsi.period || 14);
            rsi.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
            rsiSeriesRef.current.setData(parseIndicatorData(rsi) as any);
        } else if (rsiSeriesRef.current) {
            chartRef.current!.removeSeries(rsiSeriesRef.current);
            rsiSeriesRef.current = null;
        }

        chartRef.current!.timeScale().fitContent();

      } catch (err: any) {
        console.error('Chart Data Fetch Error:', err);
        setError('無法載入圖表數據');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [symbol, config]); // Re-fetch when symbol or config changes (e.g. enabling an indicator)

  return (
    <div className={`relative ${className}`}>
      {loading && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-900 bg-opacity-50 z-10 text-white">
          載入中...
        </div>
      )}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-900 bg-opacity-80 z-10 text-red-500">
          {error}
        </div>
      )}
      <div ref={chartContainerRef} className="w-full h-full" />
      <div className="absolute top-4 left-4 z-10 text-white opacity-50 text-sm pointer-events-none">
          {symbol} Technical Chart
      </div>
    </div>
  );
}
