// LC-21 GoroButton
//
// メイン画面の「ご飯めんどくさい」ボタン。useOrder hook 経由で
// PlaceOrder mutation を発火する。disabled 属性で連打抑制 1 秒 + ローディング表示。
"use client";

import { useRouter } from "next/navigation";

import { useOrder } from "@/hooks/useOrder";
import { generateUlid } from "@/lib/ulid";

export function GoroButton() {
  const router = useRouter();
  const { mutate, disabled, isPending } = useOrder();

  const handleClick = () => {
    mutate(
      {
        category: "food",
        idempotencyKey: generateUlid(),
      },
      {
        onSuccess: (res) => {
          // 完了画面に遷移 (LC-32)
          router.push(`/order/${res.orderId}/complete`);
        },
      },
    );
  };

  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={disabled}
      data-testid="goro-button"
      aria-label="ご飯めんどくさい"
      style={{
        padding: "20px 32px",
        fontSize: 24,
        fontWeight: 700,
        background: disabled ? "#bbb" : "#ff7043",
        color: "#fff",
        border: "none",
        borderRadius: 12,
        cursor: disabled ? "not-allowed" : "pointer",
        boxShadow: disabled ? "none" : "0 4px 12px rgba(255,112,67,0.4)",
        minWidth: 240,
      }}
    >
      {isPending ? "ダメ化中…" : "ご飯めんどくさい"}
    </button>
  );
}
