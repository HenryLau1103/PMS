'use client';

import { useState, useEffect, useRef } from 'react';
import { searchStocks } from '@/lib/api';
import type { TaiwanStock } from '@/types/api';

interface StockAutocompleteProps {
  value: string;
  onSelect: (stock: TaiwanStock) => void;
  placeholder?: string;
  className?: string;
}

export default function StockAutocomplete({ value, onSelect, placeholder, className }: StockAutocompleteProps) {
  const [inputValue, setInputValue] = useState(value);
  const [suggestions, setSuggestions] = useState<TaiwanStock[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const debounceTimerRef = useRef<NodeJS.Timeout>();

  // Update input when parent value changes
  useEffect(() => {
    setInputValue(value);
  }, [value]);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleInputChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    setInputValue(newValue);
    setHighlightedIndex(-1);

    // Clear existing timer
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    // Don't search if input is too short
    if (newValue.trim().length < 1) {
      setSuggestions([]);
      setIsOpen(false);
      return;
    }

    // Debounce search
    debounceTimerRef.current = setTimeout(async () => {
      setLoading(true);
      try {
        const results = await searchStocks(newValue.trim(), 10);
        setSuggestions(results);
        setIsOpen(results.length > 0);
      } catch (error) {
        console.error('Stock search failed:', error);
        setSuggestions([]);
      } finally {
        setLoading(false);
      }
    }, 300); // 300ms debounce
  };

  const handleSelectStock = (stock: TaiwanStock) => {
    setInputValue(`${stock.symbol} ${stock.name}`);
    setSuggestions([]);
    setIsOpen(false);
    onSelect(stock);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!isOpen || suggestions.length === 0) return;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setHighlightedIndex((prev) => (prev < suggestions.length - 1 ? prev + 1 : 0));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setHighlightedIndex((prev) => (prev > 0 ? prev - 1 : suggestions.length - 1));
        break;
      case 'Enter':
        e.preventDefault();
        if (highlightedIndex >= 0 && highlightedIndex < suggestions.length) {
          handleSelectStock(suggestions[highlightedIndex]);
        }
        break;
      case 'Escape':
        setIsOpen(false);
        break;
    }
  };

  return (
    <div ref={wrapperRef} className="relative">
      <input
        type="text"
        value={inputValue}
        onChange={handleInputChange}
        onKeyDown={handleKeyDown}
        onFocus={() => {
          if (suggestions.length > 0) {
            setIsOpen(true);
          }
        }}
        placeholder={placeholder || '輸入股票代號（例: 2330, 1707）'}
        className={className}
        autoComplete="off"
      />

      {/* Suggestions Dropdown */}
      {isOpen && (
        <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto">
          {loading ? (
            <div className="px-4 py-3 text-sm text-gray-500">搜尋中...</div>
          ) : suggestions.length === 0 ? (
            <div className="px-4 py-3 text-sm text-gray-500">無符合結果</div>
          ) : (
            suggestions.map((stock, index) => (
              <div
                key={stock.symbol}
                className={`px-4 py-3 cursor-pointer transition-colors ${
                  index === highlightedIndex
                    ? 'bg-primary-100'
                    : 'hover:bg-gray-100'
                }`}
                onClick={() => handleSelectStock(stock)}
                onMouseEnter={() => setHighlightedIndex(index)}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <span className="font-semibold text-gray-900">{stock.symbol}</span>
                    <span className="text-gray-700">{stock.name}</span>
                  </div>
                  <div className="flex items-center space-x-2 text-xs text-gray-500">
                    {stock.industry && (
                      <span className="px-2 py-1 bg-gray-100 rounded">{stock.industry}</span>
                    )}
                    <span className={`px-2 py-1 rounded ${stock.market === 'TSE' ? 'bg-blue-100 text-blue-700' : 'bg-green-100 text-green-700'}`}>
                      {stock.market === 'TSE' ? '上市' : '上櫃'}
                    </span>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
