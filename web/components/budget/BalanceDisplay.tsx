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

import styles from "./BalanceDisplay.module.css";

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
        className={styles.skeleton}
      >
        <div className={styles.skeletonBar} />
        <div className={styles.skeletonBarLg} />
      </div>
    );
  }

  if (query.isError || !query.data) {
    return (
      <div
        data-testid="budget-balance-display"
        role="status"
        className={styles.errorCard}
      >
        <span>残高取得に失敗しました</span>
        <button
          data-testid="budget-balance-retry"
          onClick={() => void query.refetch()}
          className={styles.retry}
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
      className={`${styles.card} ${query.isFetching ? styles.cardFetching : ""}`}
    >
      <div className={styles.label}>残りダメ予算</div>
      <div
        data-testid="budget-balance-amount"
        className={`${styles.amount} ${isWarning ? styles.amountWarning : ""}`}
      >
        {isWarning ? "⚠ " : ""}
        {formatYen(balance)}
      </div>
      <div className={styles.meta}>月間予算 {formatYen(monthlyBudget)}</div>
      <div data-testid="budget-reset-countdown" className={styles.metaCountdown}>
        月末リセットまで あと {remainingDays} 日
      </div>
    </div>
  );
}
