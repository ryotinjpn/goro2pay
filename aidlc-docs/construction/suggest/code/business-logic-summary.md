# Unit D — Business Logic Summary

**Package**: `apps/api/internal/suggest/`

## コンポーネント

| ファイル | 内容 | LC |
|---|---|---|
| `types.go` | `Suggestion` / `SuggestionPlan`（凍結契約 §5.1） | LC-SUGGEST-03 |
| `service.go` | `SuggestService`（GetSuggestion / ResolveSuggestion）+ 履歴十分・抑制判定、now 注入 | LC-SUGGEST-01 |
| `builder.go` | `SuggestionBuilder`（Bedrock 推論→失敗時 BuildFromHistory フォールバック） | LC-SUGGEST-05 |
| `logsummary.go` | `SuggestLogSummary`（構造化ログ 1 行集約、PII 非保持） | LC-SUGGEST-06 |
| `order_adapter.go` | `OrderResolverAdapter`（suggest→order.SuggestResolver ブリッジ、Q-DG1=B） | — |

## GetSuggestion ロジック（BR-D 反映）
1. `OrderHistoryReader.ListRecent(userID, 30)` → `withinWindow` で直近30日に絞り `historyCount`
2. `< 5` → `hasSuggestion:false`（BR-D01/D02）
3. `hasRecentSameCategoryOrder`（直近3h food）→ 抑制 `false`（BR-D03）
4. `SuggestionBuilder.Build` → Bedrock 1.5s×1→fallback（BR-D04/D05）、空Plan→false（BR-D06）
5. `ulid.Make()` 採番 → `SuggestionStore.Save`（TTL30分、保存失敗は非致命 BR-C10）
6. `Suggestion{hasSuggestion:true, suggestionId, title(固定), plan, fallbackUsed(内部)}`

## ResolveSuggestion ロジック
- `Store.Get` → 命中で `SuggestionPlan` / 失効・不在で `nil`（BR-D10、Unit C 透過フォールバック）。Bedrock 再検証なし（BR-D11）

## テスト
`service_test.go`（履歴不足/十分/抑制/空Plan/Resolve命中・失効）、`builder_test.go`（Bedrock成功/失敗フォールバック）、`pbt_test.go`（gopter: withinWindow / 抑制判定の性質）
