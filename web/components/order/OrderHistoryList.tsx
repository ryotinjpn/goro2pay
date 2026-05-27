// LC-24 OrderHistoryList (P-FE-LOAD-01)
//
// 注文履歴のリスト表示。
//   - 初回ロード中           : OrderHistorySkeleton
//   - 初回ロード失敗 (data なし): エラーバナー (F-I1 修正)
//   - 再 fetch 中            : 前回データを opacity 0.5 + aria-busy=true (F-I2)
//   - 0 件                   : ダメ化文言「履歴がまだありません」
"use client";

import { useOrderHistory } from "@/hooks/useOrderHistory";
import { OrderHistorySkeleton } from "./OrderHistorySkeleton";

import styles from "./OrderHistoryList.module.css";

export function OrderHistoryList() {
  const { data, isLoading, isFetching, isError } = useOrderHistory(20);

  if (isLoading) {
    return <OrderHistorySkeleton />;
  }

  // F-I1 修正: 初回ロード失敗時 (data === undefined && isError) は
  // 「履歴がまだありません」ではなくエラーバナーを表示。
  // エラーを空状態として誤表示するのはユーザ認知を歪めるため明示。
  if (isError && !data) {
    return (
      <div role="alert" data-testid="order-history-error" className={styles.error}>
        履歴の取得に失敗しました…再読込でやり直してください
      </div>
    );
  }

  const items = data ?? [];

  if (items.length === 0) {
    return (
      <p data-testid="order-history-empty" className={styles.empty}>
        履歴がまだありません
      </p>
    );
  }

  return (
    <div
      data-testid="order-history-list"
      // F-I2 修正: aria-busy で refetch 状態を screen reader に伝える
      aria-busy={isFetching}
      className={`${styles.list} ${isFetching ? styles.listFetching : ""}`}
    >
      {isError ? (
        <div role="alert" className={styles.errorInline}>
          履歴の取得に失敗しました
        </div>
      ) : null}
      {items.map((r) => (
        <div key={r.orderId} className={styles.item}>
          <div className={styles.itemMain}>
            <div className={styles.storeName}>{r.storeName}</div>
            <div className={styles.menuName}>{r.menuName}</div>
          </div>
          <div className={styles.amount}>¥{r.amount.toLocaleString("ja-JP")}</div>
        </div>
      ))}
    </div>
  );
}
