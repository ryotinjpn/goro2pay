# Unit D (`suggest`) — Code Generation Summary

**Created**: 2026-05-25
**Stage**: Construction / Code Generation（Part 2）
**Unit**: D — `suggest`（学習・先回り）
**Plan**: [suggest-code-generation-plan.md](../../plans/suggest-code-generation-plan.md)（Q-DG1=B / Q-DG2=B / Q-DG3=A / Q-DG4=A）

Unit D の実コードを Backend → Frontend → Infra → Docs の順で生成し、各層で検証した。本ドキュメントは生成物の索引。

---

## 1. 生成・変更ファイル一覧

### Backend (Go) — [business-logic-summary](./business-logic-summary.md) / [api-layer-summary](./api-layer-summary.md) / [repository-layer-summary](./repository-layer-summary.md)
| ファイル | 区分 |
|---|---|
| `apps/api/internal/suggest/{types,service,builder,logsummary,order_adapter}.go` | 新規 |
| `apps/api/internal/suggest/{service_test,builder_test,pbt_test}.go` | 新規（テスト） |
| `apps/api/internal/repo/suggestion/{repository,inmemory,repository_test}.go` | 新規 |
| `apps/api/internal/handlers/suggest_handler.go` | 新規 |
| `apps/api/internal/adapters/bedrock/{adapter,prompt,mock}.go` | 変更（InferSuggestion 追加） |
| `apps/api/internal/order/{service,suggest_resolver}.go` | 変更（SuggestResolver 注入、Q-DG1=B） |
| `apps/api/main.go` | 変更（DI + route 配線） |

### Frontend (TS) — [frontend-summary](./frontend-summary.md)
| ファイル | 区分 |
|---|---|
| `web/lib/api/suggest.ts` / `web/hooks/useSuggestion.ts` / `web/components/order/SuggestBubble.tsx` | 新規 |
| `web/components/order/GoroButton.tsx` | 変更 |
| `web/tests/{suggest-api,suggestBubble}.test.*` | 新規（テスト） |

### Infra (Terraform) — [infrastructure-summary](./infrastructure-summary.md)
| ファイル | 区分 |
|---|---|
| `infra/modules/suggestion/*` (main/locals/variables/outputs/versions/README + tests) | 新規 |
| `infra/modules/api_gateway/routes.tf` / `infra/modules/lambda_api/{variables,api_lambda}.tf` / `infra/envs/dev/main.tf` | 変更 |

---

## 2. 検証結果

| 層 | 検証 | 結果 |
|---|---|---|
| Backend | `go build ./...` / `go test ./...` / `go vet` / `gofmt` | **16 pkg PASS**（Unit C order/bedrock リグレッションなし） |
| Frontend | `npm ci` 632pkg / `vitest run` / `tsc --noEmit` | **92 tests PASS** / 型 PASS |
| Infra | `terraform fmt` / suggestion `terraform test` / envs/dev `validate`（terraform v1.15.4） | **test 4 PASS** / validate Success |

---

## 3. Story トレーサビリティ

| Story | 実装 |
|---|---|
| US-2-01 起動時サジェスト | `SuggestService.GetSuggestion` / `useSuggestion` / `SuggestBubble` |
| US-2-02 履歴不足非表示 | GetSuggestion の履歴十分判定（5件、BR-D01/D02） |
| US-2-03 1タップ注文（Unit C 共同） | `ResolveSuggestion` + order `SuggestResolver` 注入（Q-DG1=B） |
| US-2-04 学習 | OrderHistoryReader 読取（Unit C 所有） |
| US-X-02 自己委譲 | 先回り提案 + 透過フォールバック |

---

## 4. 主要な設計判断（Plan Q-DG）

- **Q-DG1=B**: OrderService に `SuggestResolver` を注入。`PlaceOrder` で suggestionId→保存値（BR-C09）、失効時は Bedrock 推論へ透過フォールバック（BR-C10）。Unit C の `order/service.go` / `NewService` / `main.go` を変更（既存テストはリグレッションなし）
- **Q-DG2=B**: Backend/Frontend/Infra/Docs を 1 PR にまとめる
- **Q-DG3=A**: `InferSuggestion` は `InferOrderPlan` のプロンプト様式踏襲（`inferWithPrompt` 共通化）
- **Q-DG4=A**: レイヤ順生成 + 各層検証

## 5. 後続（Build and Test）への申し送り
- dev 環境での実 Bedrock E2E（履歴十分→提案→1タップ注文）は Build and Test で実施
- 横串デザインシステム（PR #94）マージ後、GoroButton/SuggestBubble の視覚が上書きされる前提
