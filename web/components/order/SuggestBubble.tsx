// LC-SUGGEST-11 SuggestBubble
//
// 先回りサジェスト時に GoroButton 上部へ表示する吹き出し (design spec §3.3)。
// 文言は固定「そろそろだろ。」(BR-D13: API title より design spec の固定文言を優先)。
// 表示専用。タップ等のインタラクションは GoroButton 側が担う。
"use client";

import type { SuggestionPlan } from "@/lib/api/suggest";

// design spec §3.3 / BR-D13: 吹き出し固定文言。
const BUBBLE_TEXT = "そろそろだろ。";

export function SuggestBubble({ plan }: { plan: SuggestionPlan }) {
  return (
    <div
      data-testid="suggest-bubble"
      aria-hidden="true"
      style={{
        marginBottom: 12,
        padding: "10px 16px",
        background: "rgba(26,18,8,0.92)",
        color: "#f4ecd8",
        border: "1px solid rgba(201,169,107,0.3)",
        borderRadius: 12,
        textAlign: "center",
        maxWidth: 280,
      }}
    >
      <p style={{ margin: 0, fontSize: 16, fontWeight: 700 }}>{BUBBLE_TEXT}</p>
      <p data-testid="suggest-plan-label" style={{ margin: "4px 0 0", fontSize: 13, color: "#c9a96b" }}>
        — {plan.storeName} ¥{plan.amount.toLocaleString()} だ。
      </p>
    </div>
  );
}
