// LC-24 OrderHistorySkeleton (P-FE-LOAD-01)
//
// 初回ロード中のスケルトン表示。再 fetch 時は OrderHistoryList が
// placeholderData の前回値を opacity 0.5 で表示するため本コンポーネントは未使用。
"use client";

export function OrderHistorySkeleton() {
  return (
    <div
      data-testid="order-history-skeleton"
      aria-busy="true"
      aria-label="履歴を読み込み中"
      style={{ display: "flex", flexDirection: "column", gap: 12 }}
    >
      {[0, 1, 2].map((i) => (
        <div
          key={i}
          style={{
            height: 56,
            background: "linear-gradient(90deg, #eee 0%, #f5f5f5 50%, #eee 100%)",
            borderRadius: 8,
          }}
        />
      ))}
    </div>
  );
}
