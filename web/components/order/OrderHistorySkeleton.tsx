// LC-24 OrderHistorySkeleton (P-FE-LOAD-01)
//
// 初回ロード中のスケルトン表示。再 fetch 時は OrderHistoryList が
// placeholderData の前回値を opacity 0.5 で表示するため本コンポーネントは未使用。
"use client";

import styles from "./OrderHistorySkeleton.module.css";

export function OrderHistorySkeleton() {
  return (
    <div
      data-testid="order-history-skeleton"
      aria-busy="true"
      aria-label="履歴を読み込み中"
      className={styles.skeleton}
    >
      {[0, 1, 2].map((i) => (
        <div key={i} className={styles.bar} />
      ))}
    </div>
  );
}
