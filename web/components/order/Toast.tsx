// LC-31 Toast コンポーネント
//
// 個別トーストの表示。role="status" でスクリーンリーダーにステータス情報として通知。
"use client";

type ToastProps = {
  id: string;
  text: string;
  durationMs: number;
};

export function Toast({ text }: ToastProps) {
  return (
    <div
      role="status"
      style={{
        background: "rgba(40, 40, 40, 0.92)",
        color: "#fff",
        padding: "10px 16px",
        borderRadius: 8,
        marginTop: 8,
        boxShadow: "0 4px 12px rgba(0,0,0,0.18)",
        maxWidth: 360,
        fontSize: 14,
        lineHeight: 1.4,
      }}
    >
      {text}
    </div>
  );
}
