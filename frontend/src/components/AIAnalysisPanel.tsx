'use client';

import { useState, useEffect } from 'react';
import axios from 'axios';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface AIAnalysis {
  symbol: string;
  analysis_type: string;
  content: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  created_at: string;
  cached: boolean;
}

interface SentimentSummary {
  symbol: string;
  days: number;
  total_articles: number;
  positive_count: number;
  negative_count: number;
  neutral_count: number;
  average_score: number;
  overall_sentiment: string;
}

interface AIStatus {
  configured: boolean;
  message: string;
}

interface Props {
  symbol: string;
}

export default function AIAnalysisPanel({ symbol }: Props) {
  const [aiStatus, setAiStatus] = useState<AIStatus | null>(null);
  const [analysis, setAnalysis] = useState<AIAnalysis | null>(null);
  const [sentiment, setSentiment] = useState<SentimentSummary | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [analysisType, setAnalysisType] = useState<string>('daily_summary');

  // Check AI status on mount
  useEffect(() => {
    const checkStatus = async () => {
      try {
        const res = await axios.get(`${API_BASE}/api/v1/ai/status`);
        setAiStatus(res.data);
      } catch {
        setAiStatus({ configured: false, message: 'AI 服務無法連線' });
      }
    };
    checkStatus();
  }, []);

  // Fetch sentiment when symbol changes
  useEffect(() => {
    const fetchSentiment = async () => {
      try {
        const res = await axios.get(`${API_BASE}/api/v1/sentiment/${symbol}?days=7`);
        if (res.data.success) {
          setSentiment(res.data.data);
        }
      } catch {
        setSentiment(null);
      }
    };
    if (symbol) {
      fetchSentiment();
    }
  }, [symbol]);

  const fetchAnalysis = async () => {
    if (!aiStatus?.configured) return;
    
    setLoading(true);
    setError(null);
    try {
      const res = await axios.get(`${API_BASE}/api/v1/ai/${symbol}/analysis?type=${analysisType}`);
      if (res.data.success) {
        setAnalysis(res.data.data);
      } else {
        setError(res.data.error || '分析失敗');
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : '請求失敗';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const getSentimentColor = (sentiment: string) => {
    switch (sentiment) {
      case 'positive': return 'text-red-600 bg-red-50 border-red-200';
      case 'negative': return 'text-green-600 bg-green-50 border-green-200';
      default: return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  };

  const getSentimentText = (sentiment: string) => {
    switch (sentiment) {
      case 'positive': return '正面';
      case 'negative': return '負面';
      default: return '中性';
    }
  };

  const analysisTypes = [
    { value: 'daily_summary', label: '每日摘要' },
    { value: 'investment_advice', label: '投資建議' },
    { value: 'risk_assessment', label: '風險評估' },
    { value: 'news_digest', label: '新聞分析' },
  ];

  return (
    <div className="bg-white rounded-lg shadow-lg p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h3 className="text-lg font-bold text-gray-900 flex items-center">
          <span className="mr-2">🤖</span>
          AI 智能分析
          <span className="ml-3 px-3 py-1 text-sm font-medium bg-blue-100 text-blue-800 rounded-full">
            {symbol}
          </span>
        </h3>
        {aiStatus?.configured && (
          <span className="text-xs text-green-600 flex items-center">
            <span className="w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse"></span>
            Gemini AI 已連線
          </span>
        )}
      </div>

      {/* Main Grid Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left: Controls (3/12) */}
        <div className="lg:col-span-3 space-y-4">
          {/* Sentiment Summary */}
          {sentiment && sentiment.total_articles > 0 && (
            <div className="p-4 bg-gray-50 rounded-lg border border-gray-200">
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm font-semibold text-gray-700">近7日新聞情緒</span>
                <span className={`px-2 py-1 rounded text-xs font-bold border ${getSentimentColor(sentiment.overall_sentiment)}`}>
                  {getSentimentText(sentiment.overall_sentiment)}
                </span>
              </div>
              <div className="grid grid-cols-3 gap-2 text-center">
                <div className="p-3 bg-red-50 rounded-lg border border-red-100">
                  <div className="text-2xl font-bold text-red-600">{sentiment.positive_count}</div>
                  <div className="text-xs text-gray-500">正面</div>
                </div>
                <div className="p-3 bg-gray-100 rounded-lg border border-gray-200">
                  <div className="text-2xl font-bold text-gray-600">{sentiment.neutral_count}</div>
                  <div className="text-xs text-gray-500">中性</div>
                </div>
                <div className="p-3 bg-green-50 rounded-lg border border-green-100">
                  <div className="text-2xl font-bold text-green-600">{sentiment.negative_count}</div>
                  <div className="text-xs text-gray-500">負面</div>
                </div>
              </div>
              <div className="mt-3 flex items-center justify-between text-xs text-gray-500">
                <span>共 {sentiment.total_articles} 篇新聞</span>
                <span>情緒分數: <strong className={sentiment.average_score > 0 ? 'text-red-600' : sentiment.average_score < 0 ? 'text-green-600' : 'text-gray-600'}>{sentiment.average_score.toFixed(2)}</strong></span>
              </div>
            </div>
          )}

          {/* AI Status Warning */}
          {!aiStatus?.configured && (
            <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
              <div className="flex items-start text-yellow-800">
                <svg className="w-5 h-5 mr-2 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                <div>
                  <p className="text-sm font-medium">AI 分析功能未啟用</p>
                  <p className="text-xs mt-1">請在 docker-compose.yml 設定 GEMINI_API_KEY</p>
                </div>
              </div>
            </div>
          )}

          {/* Analysis Type Selector */}
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">分析類型</label>
            <div className="grid grid-cols-2 gap-2">
              {analysisTypes.map((type) => (
                <button
                  key={type.value}
                  onClick={() => setAnalysisType(type.value)}
                  disabled={!aiStatus?.configured || loading}
                  className={`px-3 py-2 text-sm rounded-lg border transition-colors ${
                    analysisType === type.value
                      ? 'bg-blue-600 text-white border-blue-600'
                      : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                  } ${!aiStatus?.configured || loading ? 'opacity-50 cursor-not-allowed' : ''}`}
                >
                  {type.label}
                </button>
              ))}
            </div>
          </div>

          {/* Generate Button */}
          <button
            onClick={fetchAnalysis}
            disabled={!aiStatus?.configured || loading}
            className={`w-full py-3 px-4 rounded-lg font-semibold text-white transition-all ${
              aiStatus?.configured && !loading
                ? 'bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 shadow-md hover:shadow-lg'
                : 'bg-gray-400 cursor-not-allowed'
            }`}
          >
            {loading ? (
              <span className="flex items-center justify-center">
                <svg className="animate-spin -ml-1 mr-2 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                AI 分析中...
              </span>
            ) : (
              <span className="flex items-center justify-center">
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                生成 AI 分析
              </span>
            )}
          </button>
        </div>

        {/* Right: Analysis Result (9/12) */}
        <div className="lg:col-span-9">
          {/* Error Display */}
          {error && (
            <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 mb-4">
              <div className="flex items-center">
                <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
                {error}
              </div>
            </div>
          )}

          {/* Analysis Result */}
          {analysis ? (
            <div className="h-full">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center space-x-3">
                  <span className="text-sm font-semibold text-gray-700">分析結果</span>
                  <span className="text-xs px-2 py-1 bg-purple-100 text-purple-700 rounded">
                    {analysisTypes.find(t => t.value === analysis.analysis_type)?.label || analysis.analysis_type}
                  </span>
                </div>
                <div className="flex items-center space-x-3 text-xs text-gray-500">
                  {analysis.cached && (
                    <span className="px-2 py-1 bg-blue-100 text-blue-700 rounded font-medium">快取</span>
                  )}
                  <span className="px-2 py-1 bg-gray-100 rounded">模型: {analysis.model}</span>
                  <span>Tokens: {analysis.input_tokens} → {analysis.output_tokens}</span>
                </div>
              </div>
              <div className="p-5 bg-gradient-to-br from-gray-50 to-gray-100 rounded-lg border border-gray-200 min-h-[300px] max-h-[400px] overflow-y-auto">
                <div className="prose prose-sm max-w-none text-gray-700 whitespace-pre-wrap leading-relaxed">
                  {analysis.content}
                </div>
              </div>
            </div>
          ) : (
            <div className="h-full flex items-center justify-center min-h-[300px] bg-gray-50 rounded-lg border-2 border-dashed border-gray-300">
              <div className="text-center text-gray-400">
                <svg className="w-16 h-16 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
                <p className="text-lg font-medium">選擇分析類型並點擊生成</p>
                <p className="text-sm mt-1">AI 將根據股票數據和新聞情緒提供分析</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
