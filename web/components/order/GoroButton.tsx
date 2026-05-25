// LC-21 GoroButton (+ LC-SUGGEST-11 連携)
//
// メイン画面の「ご飯めんどくさい」ボタン。useOrder hook 経由で
// PlaceOrder mutation を発火する。disabled 属性で連打抑制 1 秒 + ローディング表示。
//
// Unit D: useSuggestion で先回り提案がある場合は SuggestBubble を上部に表示し、
// 押下時に suggestionId を送る (US-2-03 1 タップ注文)。提案がなければ従来どおり。
// 注: 視覚 (色・アニメ) は横串デザインシステム (PR #94) が後続で上書きする。
"use client";

import { useRouter } from "next/navigation";

import { useOrder } from "@/hooks/useOrder";
import { useSuggestion } from "@/hooks/useSuggestion";
import { generateUlid } from "@/lib/ulid";

import { SuggestBubble } from "./SuggestBubble";

export function GoroButton() {
  const router = useRouter();
  const { mutate, disabled, isPending } = useOrder();
  const { suggestion } = useSuggestion();

  const plan = suggestion?.hasSuggestion ? suggestion.plan : undefined;
  const hasSuggestion = Boolean(plan);

  const handleClick = () => {
    mutate(
      {
        category: "food",
        idempotencyKey: generateUlid(),
        // サジェスト表示中はその suggestionId を送信し、Unit C が保存値を解決する
        // (US-2-03、失効時は Unit C が透過的に Bedrock 推論へフォールバック BR-C10)。
        suggestionId: hasSuggestion ? suggestion?.suggestionId : undefined,
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
    <div style={{ display: "flex", flexDirection: "column", alignItems: "center" }}>
      {plan ? <SuggestBubble plan={plan} /> : null}
      <button
        type="button"
        onClick={handleClick}
        disabled={disabled}
        data-testid="goro-button"
        aria-label={
          plan ? `${plan.storeName} ¥${plan.amount} を注文` : "ご飯めんどくさい"
        }
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
        {isPending ? "ダメ化中…" : hasSuggestion ? "押す。" : "ご飯めんどくさい"}
      </button>
    </div>
  );
}
