// LC-24 OrderHistoryList (P-FE-LOAD-01)
//
// 注文履歴のリスト表示。
//   - 初回ロード中    : OrderHistorySkeleton
//   - 再 fetch 中     : 前回データを opacity 0.5 で表示 (placeholderData 経由)
//   - 0 件            : ダメ化文言「履歴がまだありません」
"use client";

import { useOrderHistory } from "@/hooks/useOrderHistory";
import { OrderHistorySkeleton } from "./OrderHistorySkeleton";

export function OrderHistoryList() {
  const { data, isLoading, isFetching, isError } = useOrderHistory(20);

  if (isLoading) {
    return <OrderHistorySkeleton />;
  }

  const items = data ?? [];

  if (items.length === 0) {
    return (
      <p
        data-testid="order-history-empty"
        style={{ color: "#888", fontSize: 14, textAlign: "center", padding: 16 }}
      >
        履歴がまだありません
      </p>
    );
  }

  return (
    <div
      data-testid="order-history-list"
      style={{
        display: "flex",
        flexDirection: "column",
        gap: 8,
        opacity: isFetching ? 0.5 : 1,
        transition: "opacity 200ms ease",
      }}
    >
      {isError ? (
        <div
          role="alert"
          style={{ background: "#fdecea", color: "#a33", padding: 8, borderRadius: 4 }}
        >
          履歴の取得に失敗しました
        </div>
      ) : null}
      {items.map((r) => (
        <div
          key={r.orderId}
          style={{
            background: "#fff",
            border: "1px solid #eee",
            borderRadius: 8,
            padding: 12,
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div>
            <div style={{ fontWeight: 600 }}>{r.storeName}</div>
            <div style={{ fontSize: 12, color: "#888" }}>{r.menuName}</div>
          </div>
          <div style={{ fontSize: 14, color: "#444" }}>
            ¥{r.amount.toLocaleString("ja-JP")}
          </div>
        </div>
      ))}
    </div>
  );
}
