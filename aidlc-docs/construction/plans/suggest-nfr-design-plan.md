# Unit D (`suggest`) — NFR Design Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / NFR Design
**Unit**: D — `suggest`（学習・先回り）
**Depth**: Standard
**Prerequisite**: NFR Requirements 承認済み（PR #101 マージ済み）、Unit C NFR Design（PR #79 マージ済み）= パターン継承元
**Related**: [suggest/nfr-requirements/](../suggest/nfr-requirements/)（NFRD-Dxx）、継承元 [order/nfr-design/](../order/nfr-design/)（P-* / LC-ORDER-*）、凍結契約 [unit-interfaces.md](../interfaces/unit-interfaces.md) §5・§7

---

## 1. Plan の目的

Unit D `suggest` の NFR Design（設計パターン + 論理コンポーネント）を確定する。Unit D は Bedrock・観測性・DI・テストの基盤を **Unit C NFR Design（P-* パターン）から再利用**し、suggest 固有の orchestration（履歴十分判定 → 抑制 → Bedrock 推論 → フォールバック → 保存）と論理コンポーネント（`LC-SUGGEST-xx`）を新規定義する。NFRD-D01〜D20 を実装方針に落とす。

## 2. Unit C からの再利用候補（NFR Design パターン）

| Unit C パターン | 内容 | Unit D での再利用 |
|---|---|---|
| P-RETRY-01 | Bedrock Retry Classification（リトライ対象/非対象エラー分類） | `InferSuggestion` でそのまま流用（共有 RetryClassifier LC-ORDER-07） |
| P-PLAN-01 | Plan Construction Strategy（Bedrock→fallback orchestration） | suggest 版 `SuggestionBuilder` の雛形 |
| P-OBS-01 | Latency Measurement（Measure helper） | bedrockLatencyMs 計測に流用 |
| P-OBS-02 | LogSummary（構造化ログ集約） | suggest 版 `SuggestLogSummary` |
| P-OBS-03 | Layered Logging | そのまま流用 |
| P-INIT-01 | Lambda Cold Start（package-level client init） | Bedrock/DynamoDB client 共有 |
| P-DI-01 | Manual Dependency Injection | そのまま流用（main.go 配線） |
| P-MOCK-01 | Function-Field Mock | SuggestService テストに流用 |
| P-PBT-01 | Property-Based Testing | suggest 軽量版（NFRD-D12） |
| P-FE-LOAD-01 | Loading State | useSuggestion の isLoading に流用 |

> P-FE-ERR-01 / P-FE-TOAST-01〜02 / P-FE-LOCK-01 は注文固有（エラートースト・連打ロック）であり、suggest は表示のみのため**再利用しない**（カード非表示に丸めるだけ、NFRD-D17）。

## 3. 生成する Artifacts（Plan 承認 + 回答後）

1. `aidlc-docs/construction/suggest/nfr-design/nfr-design-patterns.md`（P-SG-xx + Unit C パターン再利用マップ）
2. `aidlc-docs/construction/suggest/nfr-design/logical-components.md`（LC-SUGGEST-xx + 共有 LC-ORDER-* 参照）

## 4. 作業手順（Checkboxes）

- [x] §5 の確認質問（Q-DD1〜Q-DD6）にユーザが回答（全 A・推奨）
- [x] 曖昧さ・矛盾を点検（Unit C パターン整合、矛盾なし）
- [x] ユーザによる Plan 承認（2026-05-25「すべて A・推奨で」）
- [x] `nfr-design-patterns.md` 生成
- [x] `logical-components.md` 生成
- [ ] 完了メッセージ（2-option ゲート）
- [ ] aidlc-state.md / audit.md 更新、push + PR

---

## 5. 確認質問（Q-DD1 〜 Q-DD6）

> 1 問ずつ対話形式で提示。**(推奨)** は調査に基づく既定案。「全部推奨で」で一括採用可。

#### Q-DD1 — Bedrock/観測性/DI/テストの Unit C パターン再利用
P-RETRY-01 / P-OBS-01〜03 / P-INIT-01 / P-DI-01 / P-MOCK-01 を suggest にそのまま再利用するか。

- A) **すべて再利用**（共有 RetryClassifier / LogSummary パターン / package-level client / Manual DI / Function-Field Mock）。suggest 固有は LogSummary の項目とフォールバック分岐のみ差分 **(推奨)**
- B) 一部のみ再利用（[Answer] に指定）
- C) Other

[Answer]: **A**（P-RETRY-01 / P-OBS-01〜03 / P-INIT-01 / P-DI-01 / P-MOCK-01 をすべて再利用。suggest 固有差分は LogSummary 項目とフォールバック分岐のみ）

#### Q-DD2 — `SuggestionBuilder` orchestration パターン（P-PLAN-01 相当）
GetSuggestion の中核（履歴十分判定 → 抑制判定 → Bedrock 推論 + リトライ → フォールバック → suggestionId 採番 + 保存）を 1 つの Builder パターンに集約するか。

- A) **`SuggestionBuilder`（P-PLAN-01 相当）に集約**。依存: BedrockAdapter / RetryClassifier（共有）/ FallbackSuggestProvider（共有）/ OrderHistoryReader / SuggestionStore。`SuggestService` は Builder と抑制判定を orchestrate **(推奨)**
- B) Builder に集約せず `SuggestService` に直接記述（薄い Unit なので）
- C) Other

[Answer]: **A**（`SuggestionBuilder`（P-PLAN-01 相当）に集約。依存: BedrockAdapter/RetryClassifier/FallbackSuggestProvider（共有）+ OrderHistoryReader/SuggestionStore。SuggestService が Builder + 抑制判定を orchestrate）

#### Q-DD3 — 論理コンポーネントの分割と命名
suggest 所有コンポーネントを `LC-SUGGEST-xx` で定義し、共有（Bedrock/Fallback/Retry/observability）は Unit C の `LC-ORDER-*` を参照するか。

- A) **`LC-SUGGEST-01〜`（SuggestService / SuggestHandler / SuggestionStore / SuggestionBuilder / SuggestLogSummary / useSuggestion / SuggestBubble + DTO + test mock）を定義、共有 Adapter 系は LC-ORDER-05/07/09/14 を参照**（重複定義しない） **(推奨)**
- B) 別の粒度・命名（[Answer] に指定）
- C) Other

[Answer]: **A**（LC-SUGGEST-01〜 を定義（SuggestService/SuggestHandler/SuggestionStore/SuggestionBuilder/SuggestLogSummary/useSuggestion/SuggestBubble + DTO + test mock）、共有 Adapter 系は LC-ORDER-05/07/09/14 を参照）

#### Q-DD4 — 観測性（SuggestLogSummary）
NFRD-D11 のログ項目（suggestionId/hasSuggestion/fallbackUsed/historyCount/suppressedRecentOrder/bedrockLatencyMs/bedrockAttempt）を LogSummary パターンで集約するか。

- A) **`SuggestLogSummary`（P-OBS-02 相当）に集約**。共通 8 項目（Unit A LC-AUTH-05）+ suggest 固有 7 項目。Bedrock 本文は非記録（NFRD-D16）。カスタムメトリクスなし（NFRD-D10） **(推奨)**
- B) 別方式（[Answer] に指定）
- C) Other

[Answer]: **A**（`SuggestLogSummary`（P-OBS-02 相当）に集約。共通 8 項目 + suggest 固有 7 項目。Bedrock 本文非記録、カスタムメトリクスなし）

#### Q-DD5 — フロント（useSuggestion + SuggestBubble）の設計
NFRD-D17（マウント時1回・retry0・失敗はカード非表示）と design spec §3.3 を実装パターンに落とす。

- A) **`useSuggestion`（TanStack Query、P-FE-LOAD-01 の isLoading 流用、staleTime 実質無限）+ `SuggestBubble`（GoroButton 内蔵、表示のみ）。エラートースト/連打ロックは不要（表示のみ、カード非表示に丸め）** **(推奨)**
- B) Unit C の FE パターン（toast/lock）も一部適用
- C) Other

[Answer]: **A**（`useSuggestion`（TanStack, P-FE-LOAD-01 の isLoading 流用, staleTime 実質無限）+ `SuggestBubble`（GoroButton 内蔵, 表示のみ）。toast/lock は不要、失敗はカード非表示に丸め）

#### Q-DD6 — PBT 設計パターン
NFRD-D12（軽量 PBT）を P-PBT-01 相当でどう設計するか。

- A) **`P-SG-PBT-01`（P-PBT-01 軽量版）**: backend gopter で `BuildFromHistory` 最頻判定の決定性（同入力→同出力、同点は最新優先）、frontend fast-check でレスポンス→props ラウンドトリップ。Function-Field Mock（P-MOCK-01）で Bedrock を決定論化 **(推奨)**
- B) PBT パターンは設けず通常テストのみ
- C) Other

[Answer]: **A**（`P-SG-PBT-01`（P-PBT-01 軽量版）: gopter で BuildFromHistory 最頻判定の決定性、fast-check でレスポンス→props ラウンドトリップ。P-MOCK-01 で Bedrock 決定論化）

---

## 6. 矛盾チェック観点

1. Q-DD1〜Q-DD6 が Unit C P-* パターン・NFRD-D01〜D20 と整合するか
2. 共有コンポーネント（BedrockAdapter/RetryClassifier/FallbackSuggestProvider）の再利用が凍結契約 §7 と矛盾しないか
3. frontend パターンが FD frontend-components.md / design spec §3.3 と整合するか
4. 「depends / たぶん / 標準で」等の曖昧回答がないか

## 7. 完了条件

- Q-DD1〜Q-DD6 回答 + 矛盾なし確認 + Plan 承認
- `nfr-design-patterns.md` / `logical-components.md` 生成
- 2-option 完了ゲート確認

## 8. 次ステージ（Infrastructure Design）への引き継ぎ

LC-SUGGEST-xx と再利用 LC-ORDER-* の配線、`GoroPay_Suggestion` テーブル、`GET /api/suggest` ルート、Bedrock IAM 共有参照を Infrastructure Design に引き継ぐ。
