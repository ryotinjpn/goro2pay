# Unit D — API Layer Summary

## Handler（LC-SUGGEST-02）
`apps/api/internal/handlers/suggest_handler.go`
- `GET /api/suggest`（凍結契約 §5.2）。`auth.UserIDFromContext` で userID 取得 → `SuggestService.GetSuggestion`
- レスポンス DTO は `hasSuggestion` / `suggestionId` / `title` / `plan` のみ。**`FallbackUsed` は出さない**（BR-D12）
- Service エラーは 200 + `hasSuggestion:false` に丸め（NFRD-D07、メイン機能を阻害しない二重防御）

## 横串 BedrockAdapter（変更）
`apps/api/internal/adapters/bedrock/`
- `adapter.go`: interface に `InferSuggestion(ctx, history, dayOfWeek)` 追加。`ClaudeBedrockAdapter` は `BuildSuggestPrompt` でプロンプトのみ差し替え、リトライ/タイムアウト/パースは `inferWithPrompt` を `InferOrderPlan` と**共通化**（既存挙動保持）
- `prompt.go`: `BuildSuggestPrompt`（先回り提案用、`ParsePlanResponse` 共有、PII 非送出）
- `mock.go`: `MockBedrockAdapter.InferSuggestion`（テスト用）

## ルーティング（main.go）
- `api.GET("/suggest", suggestHandler.GetSuggestion)` を認証必須グループに追加
- DI: `suggestionStore` / `suggestionBuilder` / `suggestSvc` を配線、`orderSvc.SetSuggestResolver(suggest.NewOrderResolverAdapter(suggestSvc))`（Q-DG1=B）
