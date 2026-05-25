# Unit D — Frontend Summary

## コンポーネント

| ファイル | 内容 | LC |
|---|---|---|
| `web/lib/api/suggest.ts` | `fetchSuggestion`（`apiClient.request` 経由、取得失敗は `hasSuggestion:false` に丸め NFRD-D17） | — |
| `web/hooks/useSuggestion.ts` | TanStack Query（`['suggestion']` / staleTime ∞ / refetchOnWindowFocus:false / retry:0、マウント時1回 BR-D14） | LC-SUGGEST-10 |
| `web/components/order/SuggestBubble.tsx` | 固定文言「そろそろだろ。」+ 提案ラベル（design spec §3.3 / BR-D13）、表示専用 | LC-SUGGEST-11 |
| `web/components/order/GoroButton.tsx` | useSuggestion 連携。提案あり時 SuggestBubble 表示 + 押下時 `suggestionId` 送信（US-2-03） | LC-21 改修 |

## 凍結契約 ⇄ design spec 突合（Q-DF8/D9）
- **IF=凍結契約**: `useSuggestion()` / `GET /api/suggest` / DTO
- **視覚=design spec §3.3**: `SuggestBubble`（吹き出し固定文言）。API の `title` より固定文言を優先表示（BR-D13）

## テスト
`suggest-api.test.ts`（200履歴十分/不足/HTTPエラー丸め/path）、`suggestBubble.test.tsx`（固定文言・提案ラベル描画）。`npm ci` 632pkg → vitest 92 PASS / tsc PASS

## 注記
- `['balance']` invalidate は Unit C `useOrder` が既に実施（凍結契約 §9.1）。Unit D は追加不要
- 視覚（色/アニメ）は横串デザインシステム（PR #94）が後続で上書き予定
