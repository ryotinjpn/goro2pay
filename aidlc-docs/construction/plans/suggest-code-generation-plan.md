# Unit D (`suggest`) — Code Generation Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / Code Generation（Part 1: Planning）
**Unit**: D — `suggest`（学習・先回り）
**Prerequisite**: FD / NFR-R / NFR Design / Infrastructure Design 全承認・マージ済み（PR #100/#101/#102/#103）
**Related**: [suggest/functional-design/](../suggest/functional-design/)、[suggest/nfr-design/logical-components.md](../suggest/nfr-design/logical-components.md)（LC-SUGGEST-*）、[suggest/infrastructure-design/](../suggest/infrastructure-design/)、凍結契約 [unit-interfaces.md](../interfaces/unit-interfaces.md) §5・§7
**方針**: 既存実装（Unit A/B/C）に合わせて生成。order パッケージを雛形に suggest を実装。共有 adapters / observability / logging を再利用。

---

## 1. Unit Context

### 1.1 担当ストーリー
US-2-01（起動時サジェスト）/ US-2-02（履歴不足非表示）/ US-2-03（1タップ注文 Resolve、Unit C 共同）/ US-2-04（学習）/ US-X-02（自己委譲）

### 1.2 依存・インターフェース
- **公開**: `SuggestService`（`GetSuggestion` / `ResolveSuggestion`）、`GET /api/suggest`、`useSuggestion`（契約 §5・§9）
- **再利用（既存・develop）**: `adapters/bedrock`（InferOrderPlan 済、**InferSuggestion 未実装→追加**）、`adapters/fallback`（`BuildFromHistory` 済→流用）、`adapters/bedrock/retry`（RetryClassifier）、observability/logging/middleware、`apiClient`/BFF（Unit A）
- **所有テーブル**: `GoroPay_Suggestion`（Infra Design）

### 1.3 ⚠️ 既存コードとのギャップ（要判断、§3 の Q-DG1）
develop の `order.NewService(orderHistoryRepo, planBuilder, deliveryAdapter, walletAdapter)` は **SuggestResolver を受け取っていない**（Unit C 実装時に Unit D 未存在）。凍結契約 §4.5 は OrderService が `SuggestResolver` を持つ設計。1 タップ注文（US-2-03）を成立させるには、ResolveSuggestion を order フローに接続する必要がある。

---

## 2. 生成ファイル一覧（manifest）

### 2.1 Backend（Go）
| ファイル | 内容 | 区分 |
|---|---|---|
| `apps/api/internal/suggest/types.go` | Suggestion / SuggestionPlan / 内部型（契約 §5.1） | 新規 |
| `apps/api/internal/suggest/service.go` | `SuggestService`（GetSuggestion/ResolveSuggestion）+ 抑制判定（BR-D03） | 新規 |
| `apps/api/internal/suggest/builder.go` | `SuggestionBuilder`（P-SG-BUILD-01: 履歴十分→Bedrock+リトライ→fallback） | 新規 |
| `apps/api/internal/suggest/logsummary.go` | `SuggestLogSummary`（P-SG-OBS-01） | 新規 |
| `apps/api/internal/suggest/*_test.go` | service / builder / logsummary の単体テスト | 新規 |
| `apps/api/internal/suggest/builder_pbt_test.go` | P-SG-PBT-01（gopter、BuildFromHistory 決定性） | 新規 |
| `apps/api/internal/repo/suggestion/repository.go` | SuggestionStore（DynamoDB Save/Get、TTL30分） | 新規 |
| `apps/api/internal/repo/suggestion/inmemory.go` + test | test-only in-memory + repo test | 新規 |
| `apps/api/internal/handlers/suggest_handler.go` | `GET /api/suggest` handler（FallbackUsed 除外） | 新規 |
| `apps/api/internal/adapters/bedrock/adapter.go` | **`InferSuggestion` メソッド追加** | 変更 |
| `apps/api/internal/adapters/bedrock/prompt.go` | suggest プロンプト追加 | 変更 |
| `apps/api/internal/adapters/bedrock/mock.go` | mock に InferSuggestion 追加 | 変更 |
| `apps/api/main.go` | SuggestService 配線 + route 登録 + **OrderService への SuggestResolver 接続（Q-DG1）** | 変更 |

### 2.2 Frontend（TypeScript）
| ファイル | 内容 | 区分 |
|---|---|---|
| `web/hooks/useSuggestion.ts` | TanStack（マウント時1回、staleTime∞、retry0、queryKey ['suggestion']） | 新規 |
| `web/components/order/SuggestBubble.tsx` | GoroButton 内蔵の吹き出し「そろそろだろ。」（design spec §3.3） | 新規 |
| `web/components/order/GoroButton.tsx` | suggested 状態で SuggestBubble 表示 + 副ラベル | 変更 |
| `web/lib/api/suggest.ts`（または既存 api ラッパ拡張） | GET /api/suggest クライアント | 新規/変更 |
| `web/tests/*` | useSuggestion / SuggestBubble テスト（Vitest + fast-check） | 新規 |

### 2.3 Infrastructure（Terraform）
| ファイル | 内容 | 区分 |
|---|---|---|
| `infra/modules/suggestion/{main,locals,variables,outputs}.tf` | GoroPay_Suggestion + IAM policy | 新規 |
| `infra/modules/suggestion/tests/dynamodb_schema.tftest.hcl` | スキーマ assertion | 新規 |
| `infra/modules/api_gateway/routes.tf` | `GET /api/suggest` 追記 | 変更 |
| `infra/envs/dev/main.tf` | module "suggestion" + lambda_api 呼出更新（policy/env） | 変更 |

### 2.4 ドキュメント（aidlc-docs）
| ファイル | 内容 |
|---|---|
| `aidlc-docs/construction/suggest/code/*.md` | business-logic / api-layer / repository / frontend / infrastructure / code-generation サマリ |

---

## 3. 確認質問（Q-DG1 〜 Q-DG4）

> 1 問ずつ対話形式。**(推奨)** は既定案。

#### Q-DG1 — ResolveSuggestion を order フローに接続する方法（§1.3 ギャップ）
1 タップ注文（US-2-03）で `suggestionId` から保存済み plan を解決する経路。

- A) **`OrderHandler` 層で解決**: `POST /api/orders` で `suggestionId` があれば handler が `SuggestService.ResolveSuggestion` を呼び、解決済み plan を OrderService に渡す。OrderService の既存シグネチャ（NewService）は変更最小 **(推奨)**
- B) 契約どおり `OrderService` に `SuggestResolver` を注入: `order.NewService(...)` に引数追加、main.go で SuggestService を配線（Unit C コードに踏み込む）
- C) 今回は GetSuggestion/ResolveSuggestion を standalone 実装に留め、order との 1 タップ接続は別 PR の follow-up
- D) Other

[Answer]: **B**（契約 §4.5 / FD business-logic-model §2.3 どおり OrderService に SuggestResolver を注入。service.go の plan 生成前に「suggestionId→ResolveSuggestion、nil なら planBuilder（BR-C10 フォールバック）」分岐を追加、NewService にオプション引数追加、main.go で SuggestService を配線。Unit C の merged コードに変更が入る）

#### Q-DG2 — ブランチ/PR 分割
Unit C は feat/unit-c-backend / frontend / infra + docs の 4 分割だった。

- A) **Unit C と同じ 4 分割**（feat/unit-d-backend / feat/unit-d-frontend / feat/unit-d-infra + 本 docs ブランチに plan+サマリ）。本ブランチ `docs/construction-suggest-code-generation` には plan + code サマリのみ **(推奨)**
- B) Unit D は小さいので 1 PR にまとめる（本ブランチに全部）
- C) Other

[Answer]: **B**（約 38 ファイルとまとまった規模のため 1 PR にまとめる。本ブランチ `docs/construction-suggest-code-generation` に Backend/Frontend/Infra/Docs を全部。Backend に含む Unit C order 変更〔Q-DG1=B〕は PR 説明で明示）

#### Q-DG3 — InferSuggestion の Bedrock プロンプト
共有 BedrockAdapter に InferSuggestion を追加する際のプロンプト方針。

- A) **InferOrderPlan のプロンプト様式を踏襲**し suggest 用に調整（履歴要約 → {store,menu,amount,category} を tool use/JSON で。Title は固定文言運用のため任意）。NFRD-D16（PII 非送出）遵守 **(推奨)**
- B) 別方式
- C) Other

[Answer]: **A**（InferOrderPlan のプロンプト様式を踏襲し suggest 用に調整。履歴要約→{store,menu,amount,category} を JSON/tool use、PII 非送出、Title は固定文言運用）

#### Q-DG4 — 生成スコープ（Part 2 の進め方）
Part 2（実コード生成）は大規模。どう進めるか。

- A) **本 Plan 承認後、Backend → Frontend → Infra → サマリ の順に生成**し、各まとまりでレビュー可能にする。検証（go test / vitest / terraform test）も実施 **(推奨)**
- B) 一括生成
- C) Other

[Answer]: **A**（Backend → Frontend → Infra → サマリ の順に生成、各まとまりで検証）

---

## ✅ Part 1（Planning）完了サマリ
- Q-DG1=**B**（OrderService に SuggestResolver 注入、Unit C order コード変更あり）/ Q-DG2=**B**（1 PR にまとめる）/ Q-DG3=**A**（InferOrderPlan プロンプト様式踏襲）/ Q-DG4=**A**（レイヤ順生成 + 検証）
- ユーザ Plan 承認: 2026-05-25。Part 2 へ移行。

---

## 4. 作業手順（Part 2、承認後に実行）

- [x] Step 1: 共有 BedrockAdapter に InferSuggestion + prompt + mock 追加（inferWithPrompt 共通化、InferOrderPlan 挙動保持）
- [x] Step 2: suggest パッケージ生成（types/builder/service/logsummary/order_adapter + service_test/builder_test/pbt_test）
- [x] Step 3: suggestion repository 生成（repository/inmemory + repository_test）
- [x] Step 4: suggest_handler 生成 + main.go 配線（GET /api/suggest route + DI + Q-DG1=B の order SuggestResolver 接続）
- [x] Step 5: Backend 検証（`go build` PASS / `go test ./...` 16 pkg PASS / `go vet` PASS / `gofmt` 適用）
- [ ] Step 6: Frontend（useSuggestion / SuggestBubble / GoroButton 連携 + tests）
- [ ] Step 7: Frontend 検証（`vitest` / `tsc --noEmit`）
- [ ] Step 8: Infra（modules/suggestion + routes/envs 追記 + tftest）
- [ ] Step 9: Infra 検証（`terraform validate` / `fmt` / `test`）
- [ ] Step 10: code サマリ生成（aidlc-docs/construction/suggest/code/）
- [ ] Step 11: aidlc-state.md / audit.md 更新、ブランチ分割 push + PR

---

## 5. 完了条件

- Q-DG1〜Q-DG4 回答 + Plan 承認（Part 1 完了）
- Step 1〜11 完了、各検証パス
- 2-option 完了ゲート確認

## 6. Story トレーサビリティ
US-2-01 → service.GetSuggestion / SuggestBubble、US-2-02 → 履歴十分判定、US-2-03 → ResolveSuggestion + order 接続（Q-DG1）、US-2-04 → OrderHistoryReader 流用、US-X-02 → 先回り体験全体
