// LC-BUDGET-09 BalanceDisplay (P-DEG-01 / NFR-DEG-03 残高常時可視化)。
//
// 内部で useQuery を直接 subscribe する (FD §3.1 注: 凍結 IF §9 の useWallet では
// isError が公開されないため、エラー UI 制御に直接アクセス)。
//
// 表示分岐:
//   - isLoading=true (初回ロード)        → スケルトン
//   - isFetching=true && data !== null   → 薄い表示 (opacity 0.5)
//   - 通常                                → 残高 + ラベル + リセット日カウントダウン
"use client";

import { useQuery } from "@tanstack/react-query";

import { type WalletResponse } from "@/lib/api/wallet";
import { WALLET_QUERY_OPTIONS } from "@/hooks/useWallet";

function computeRemainingDays(now: Date): number {
  const next = new Date(now.getFullYear(), now.getMonth() + 1, 1);
  const diffMs = next.getTime() - now.getTime();
  return Math.max(1, Math.ceil(diffMs / (24 * 60 * 60 * 1000)));
}

function formatYen(n: number): string {
  return `¥${new Intl.NumberFormat("ja-JP").format(n)}`;
}

export function BalanceDisplay() {
  // useWallet と同じ TanStack Query 設定を共有 (Code Review Important 7)。
  // 内部で useQuery を直接 subscribe するのは FD §3.1 注: 凍結 IF §9 の
  // useWallet では isError / isFetching が公開されないため。
  const query = useQuery<WalletResponse>(WALLET_QUERY_OPTIONS);

  if (query.isLoading) {
    return (
      <div
        data-testid="budget-balance-skeleton"
        aria-busy="true"
        style={{
          padding: 16,
          borderRadius: 12,
          background: "#f3f4f6",
          minHeight: 120,
          animation: "pulse 1.4s ease-in-out infinite",
        }}
      >
        <div
          style={{
            height: 14,
            width: 96,
            background: "#e5e7eb",
            borderRadius: 4,
            marginBottom: 12,
          }}
        />
        <div
          style={{
            height: 36,
            width: 160,
            background: "#e5e7eb",
            borderRadius: 6,
          }}
        />
      </div>
    );
  }

  if (query.isError || !query.data) {
    return (
      <div
        data-testid="budget-balance-display"
        role="status"
        style={{
          padding: 16,
          borderRadius: 12,
          background: "#fff7ed",
          color: "#9a3412",
        }}
      >
        残高取得に失敗しました
        <button
          data-testid="budget-balance-retry"
          onClick={() => void query.refetch()}
          style={{
            marginLeft: 12,
            background: "#9a3412",
            color: "#fff",
            border: "none",
            borderRadius: 6,
            padding: "4px 10px",
            cursor: "pointer",
          }}
        >
          再試行
        </button>
      </div>
    );
  }

  const { balance, monthlyBudget } = query.data;
  const consumptionRate = monthlyBudget > 0 ? 1 - balance / monthlyBudget : 0;
  const isWarning = consumptionRate > 0.8;
  const remainingDays = computeRemainingDays(new Date());

  return (
    <div
      data-testid="budget-balance-display"
      role="status"
      aria-live="polite"
      style={{
        padding: 20,
        borderRadius: 12,
        background: "#ffffff",
        boxShadow: "0 2px 8px rgba(0,0,0,0.06)",
        opacity: query.isFetching ? 0.5 : 1,
        transition: "opacity 0.2s",
      }}
    >
      <div style={{ fontSize: 13, color: "#6b7280", marginBottom: 6 }}>
        残りダメ予算
      </div>
      <div
        data-testid="budget-balance-amount"
        style={{
          fontSize: 40,
          fontWeight: 700,
          color: isWarning ? "#dc2626" : "#111827",
        }}
      >
        {isWarning ? "⚠ " : ""}
        {formatYen(balance)}
      </div>
      <div style={{ fontSize: 12, color: "#9ca3af", marginTop: 8 }}>
        月間予算 {formatYen(monthlyBudget)}
      </div>
      <div
        data-testid="budget-reset-countdown"
        style={{ fontSize: 12, color: "#9ca3af", marginTop: 4 }}
      >
        月末リセットまで あと {remainingDays} 日
      </div>
    </div>
  );
}
