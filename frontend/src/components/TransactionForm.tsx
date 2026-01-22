'use client';

import { useState } from 'react';
import { createEvent } from '@/lib/api';
import type { EventType, TaiwanStock } from '@/types/api';
import StockAutocomplete from './StockAutocomplete';

interface TransactionFormProps {
  portfolioId: string;
  onSuccess?: () => void;
}

export default function TransactionForm({ portfolioId, onSuccess }: TransactionFormProps) {
  const [formData, setFormData] = useState({
    event_type: 'BUY' as EventType,
    symbol: '',
    stock_name: '',
    quantity: '',
    price: '',
    fee: '0',
    tax: '0',
    occurred_at: new Date().toISOString().slice(0, 16),
    notes: '',
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    // Validate symbol is provided
    if (!formData.symbol) {
      setError('請選擇股票');
      setLoading(false);
      return;
    }

    try {
      // Auto-calculate Taiwan stock trading fees and taxes
      const quantity = parseFloat(formData.quantity);
      const price = parseFloat(formData.price);
      const totalValue = quantity * price;

      let fee = parseFloat(formData.fee);
      let tax = parseFloat(formData.tax);

      // Auto-calculate if not manually set
      if (fee === 0 && formData.event_type === 'BUY') {
        // 手續費 0.1425% (最低20元)
        fee = Math.max(20, totalValue * 0.001425);
      } else if (fee === 0 && formData.event_type === 'SELL') {
        // 手續費 0.1425% (最低20元)
        fee = Math.max(20, totalValue * 0.001425);
      }

      if (tax === 0 && formData.event_type === 'SELL') {
        // 證券交易稅 0.3%
        tax = totalValue * 0.003;
      }

      await createEvent({
        portfolio_id: portfolioId,
        event_type: formData.event_type,
        symbol: formData.symbol,
        quantity: formData.quantity,
        price: formData.price,
        fee: fee.toFixed(2),
        tax: tax.toFixed(2),
        occurred_at: new Date(formData.occurred_at).toISOString(),
        notes: formData.notes || undefined,
      });

      // Reset form
      setFormData({
        event_type: 'BUY',
        symbol: '',
        stock_name: '',
        quantity: '',
        price: '',
        fee: '0',
        tax: '0',
        occurred_at: new Date().toISOString().slice(0, 16),
        notes: '',
      });

      onSuccess?.();
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || '新增交易失敗');
    } finally {
      setLoading(false);
    }
  };

  const handleStockSelect = (stock: TaiwanStock) => {
    setFormData({
      ...formData,
      symbol: stock.symbol,
      stock_name: stock.name,
    });
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="bg-white shadow-md rounded-lg p-6">
      <h2 className="text-2xl font-bold mb-6">新增交易記錄</h2>

      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-md">
          <p className="text-red-800">{error}</p>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Transaction Type */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            交易類型
          </label>
          <select
            name="event_type"
            value={formData.event_type}
            onChange={handleChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            required
          >
            <option value="BUY">買入</option>
            <option value="SELL">賣出</option>
            <option value="DIVIDEND">股利</option>
          </select>
        </div>

        {/* Symbol - Stock Autocomplete */}
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            股票代號 / 名稱
          </label>
          <StockAutocomplete
            value={formData.symbol ? `${formData.symbol} ${formData.stock_name}` : ''}
            onSelect={handleStockSelect}
            placeholder="輸入代號（例: 2330, 1707）"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>

        {/* Quantity */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            數量 (股)
          </label>
          <input
            type="number"
            name="quantity"
            value={formData.quantity}
            onChange={handleChange}
            placeholder="1000"
            step="1"
            min="1"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            required
          />
        </div>

        {/* Price */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            價格 (元)
          </label>
          <input
            type="number"
            name="price"
            value={formData.price}
            onChange={handleChange}
            placeholder="100.00"
            step="0.01"
            min="0"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            required
          />
        </div>

        {/* Fee */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            手續費 (元) - 留空自動計算
          </label>
          <input
            type="number"
            name="fee"
            value={formData.fee}
            onChange={handleChange}
            placeholder="0"
            step="0.01"
            min="0"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>

        {/* Tax */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            證券交易稅 (元) - 留空自動計算
          </label>
          <input
            type="number"
            name="tax"
            value={formData.tax}
            onChange={handleChange}
            placeholder="0"
            step="0.01"
            min="0"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
          />
        </div>

        {/* Occurred At */}
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            交易日期時間
          </label>
          <input
            type="datetime-local"
            name="occurred_at"
            value={formData.occurred_at}
            onChange={handleChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            required
          />
        </div>

        {/* Notes */}
        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            備註 (選填)
          </label>
          <textarea
            name="notes"
            value={formData.notes}
            onChange={handleChange}
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            placeholder="記錄交易原因或其他備註..."
          />
        </div>
      </div>

      <div className="mt-6">
        <button
          type="submit"
          disabled={loading}
          className="w-full bg-primary-600 text-white py-3 px-4 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {loading ? '處理中...' : '新增交易'}
        </button>
      </div>
    </form>
  );
}
