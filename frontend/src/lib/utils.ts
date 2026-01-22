// Utility functions for the app

/**
 * Format number as Taiwan dollar currency
 */
export function formatCurrency(value: string | number): string {
  const num = typeof value === 'string' ? parseFloat(value) : value;
  return new Intl.NumberFormat('zh-TW', {
    style: 'currency',
    currency: 'TWD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(num);
}

/**
 * Format number as percentage
 */
export function formatPercentage(value: string | number, decimals: number = 2): string {
  const num = typeof value === 'string' ? parseFloat(value) : value;
  return `${num.toFixed(decimals)}%`;
}

/**
 * Format large numbers with K, M, B suffixes
 */
export function formatCompactNumber(value: string | number): string {
  const num = typeof value === 'string' ? parseFloat(value) : value;
  return new Intl.NumberFormat('zh-TW', {
    notation: 'compact',
    compactDisplay: 'short',
  }).format(num);
}

/**
 * Validate Taiwan stock symbol (4 digits + .TW or .TWO)
 */
export function validateTaiwanSymbol(symbol: string): boolean {
  const pattern = /^[0-9]{4}\.(TW|TWO)$/;
  return pattern.test(symbol);
}

/**
 * Get color class for P&L value
 */
export function getPnLColorClass(value: string | number): string {
  const num = typeof value === 'string' ? parseFloat(value) : value;
  if (num > 0) return 'text-success';
  if (num < 0) return 'text-danger';
  return 'text-gray-600';
}

/**
 * Calculate percentage change
 */
export function calculatePercentageChange(current: number, original: number): number {
  if (original === 0) return 0;
  return ((current - original) / original) * 100;
}

/**
 * Format date to Taiwan locale
 */
export function formatDate(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return new Intl.DateTimeFormat('zh-TW', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(d);
}

/**
 * Format datetime to Taiwan locale
 */
export function formatDateTime(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return new Intl.DateTimeFormat('zh-TW', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(d);
}

/**
 * Combine class names conditionally
 */
export function cn(...classes: (string | boolean | undefined)[]): string {
  return classes.filter(Boolean).join(' ');
}
