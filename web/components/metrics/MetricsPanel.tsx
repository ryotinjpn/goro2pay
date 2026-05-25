"use client";

import { useMetrics } from "@/hooks/useMetrics";

// MetricsPanel は MainScreen に埋め込まれる今月のダメ化メトリクス表示コンポーネント (LC-ME-09)。
// P-ME-FE-DEG-01: ThresholdExceeded=true のとき animate-pulse + text-red-500 で演出 (NFRE-E08)。
export function MetricsPanel() {
  const { data: metrics, isLoading } = useMetrics();

  if (isLoading) {
    return (
      <div data-testid="metrics-panel-loading" className="animate-pulse space-y-2">
        <div className="h-4 bg-gray-200 rounded w-3/4" />
        <div className="h-4 bg-gray-200 rounded w-1/2" />
      </div>
    );
  }

  if (!metrics) return null;

  const barClass = metrics.thresholdExceeded
    ? "bg-red-500 animate-pulse"
    : "bg-blue-500";

  const rateClass = metrics.thresholdExceeded
    ? "text-red-500 animate-pulse font-bold"
    : "text-gray-700";

  const pct = Math.round(metrics.consumptionRate * 100);

  return (
    <div data-testid="metrics-panel" className="space-y-3">
      <p className="text-sm text-gray-500">今月のダメ化</p>
      <p className="text-2xl font-bold">{metrics.damageCount} 回</p>

      <div className="space-y-1">
        <div className="flex justify-between text-sm">
          <span className="text-gray-600">消化率</span>
          <span data-testid="consumption-rate" className={rateClass}>{pct}%</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div
            data-testid="consumption-bar"
            className={`h-2 rounded-full ${barClass}`}
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>

      <p className="text-xs text-gray-400">{metrics.summaryText}</p>
    </div>
  );
}
