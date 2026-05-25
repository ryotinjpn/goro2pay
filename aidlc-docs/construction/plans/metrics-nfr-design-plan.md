# Unit E `metrics` — NFR Design Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard
**Stage**: NFR Design (Construction Phase, per-unit)

---

## 0. このドキュメントの目的

Unit E `metrics` の NFR Design ステージで生成する成果物（nfr-design-patterns.md / logical-components.md）の前段として、パターン定義と論理コンポーネント設計を確定するために必要な意思決定を収集する。

方針: Unit D と同様に Standard 深度でヒアリングは 5 問（Q-DD1〜Q-DD5）。回答受領のたびに `[Answer]:` を更新する。

---

## 1. Plan 実施チェックリスト

- [x] Step 1: NFR Requirements / Functional Design 成果物を読み込み済み（NFRE-E01〜E10 / business-logic-model / unit-interfaces.md §6）
- [x] Step 2: Unit C / D の nfr-design-patterns.md / logical-components.md を参照して再利用候補を特定
- [x] Step 3: ヒアリング質問を抽出（Q-DD1〜Q-DD5）
- [x] Step 4: チャットで 1 問ずつ提示してヒアリング（全 5 問完了）
- [x] Step 5: 回答を `[Answer]:` タグへ反映、矛盾検出
- [x] Step 6: nfr-design-patterns.md / logical-components.md 生成

---

## 2. Unit E が担当する NFR 要件（再掲）

| ID | カテゴリ | 要件概要 |
|---|---|---|
| NFRE-E01 | Performance | GET /api/metrics P95 500ms。errgroup 並列読取（BudgetSettings + Wallet）|
| NFRE-E02 | Performance | GET /api/budget/raise/recommendation P95 200ms |
| NFRE-E03 | Performance | POST /api/budget/raise P95 300ms |
| NFRE-E04 | Reliability | DynamoDB 障害 → 即 HTTP 500（部分返却なし） |
| NFRE-E05 | Reliability | ErrNoBudgetSet → HTTP 400 |
| NFRE-E06 | Scalability | 新規テーブルなし、既存読取のみ |
| NFRE-E07 | Maintainability | gopter PBT: ConsumptionRate / ComputeRecommendedBudget |
| NFRE-E08 | Usability | ThresholdExceeded → animate-pulse + text-red-500 |
| NFRE-E09 | Usability | BudgetEmptyScreen マウント → RaiseModal 強制表示 |
| NFRE-E10 | Security | Unit A AttachUserID middleware 適用 |

---

## 3. Unit C / D パターン再利用候補

| Unit C パターン | 内容 | E で再利用候補か |
|---|---|---|
| P-INIT-01 | Lambda INIT フェーズ client 初期化 | ✅ DynamoDB client 共有 |
| P-DI-01 | Manual Dependency Injection (main.go) | ✅ MetricsService / BudgetRaiseService 配線 |
| P-MOCK-01 | Function-Field Mock | ✅ BudgetSettingsReader / WalletReader / OrderHistoryReader モック |
| P-OBS-01 | Latency Measurement | △ MetricsService の latency 計測（Optional）|
| P-OBS-02 | LogSummary（構造化ログ 1 行集約） | △ Q-DD4 で判断 |
| P-OBS-03 | Layered Logging | ✅ INFO/WARN 層別 |
| P-PBT-01 | Property-Based Testing (gopter) | 類似。Unit E 独自 P-E-PBT-01/02 として定義 |
| P-FE-LOAD-01 | Loading State | ✅ useMetrics / useBudgetRaise の isLoading |
| P-RETRY-01 | Bedrock Retry | ❌ 不要（DynamoDB 読取、Bedrock なし） |
| P-PLAN-01 | Plan Construction | ❌ 不要（Bedrock なし） |

---

## 4. 質問

### Q-DD1: Unit C パターン再利用範囲

Unit E で Unit C（または D）から**再利用するパターン**として適切なのはどれか。

A) P-INIT-01 / P-DI-01 / P-MOCK-01 / P-OBS-03 / P-FE-LOAD-01 を再利用し、P-OBS-01（latency 計測）は metrics の観測性として追加採用する（全 5 + OBS-01）
B) P-INIT-01 / P-DI-01 / P-MOCK-01 / P-FE-LOAD-01 のみ再利用し、OBS 系は省略（Unit E は観測性要件なし）
C) P-INIT-01 / P-DI-01 / P-MOCK-01 / P-OBS-01 / P-OBS-02 / P-OBS-03 / P-FE-LOAD-01 を全採用（Unit C と同水準の観測性）
X) Other

[Answer]: A — P-INIT-01 / P-DI-01 / P-MOCK-01 / P-OBS-01 / P-OBS-03 / P-FE-LOAD-01 を再利用（P-OBS-01 の latency 計測を追加採用）

---

### Q-DD2: errgroup 並列読取パターンの定義方法

NFRE-E01 の `errgroup` 並列読取（BudgetSettings + Wallet）を NFR Design でどう扱うか。

A) 新規パターン `P-ME-PARALLEL-01` として定義し、errgroup の使い方・エラー集約・CountThisMonth 呼出タイミングを明文化する（後続 Code Generation で参照しやすい）
B) nfr-requirements.md §2.2 の擬似コードを参照するだけで十分。新規パターン ID は不要
X) Other

[Answer]: A — 新規パターン `P-ME-PARALLEL-01` として定義（errgroup 構造 / ErrNoBudgetSet 判定タイミング / CountThisMonth 直列呼出 を明文化）

---

### Q-DD3: DEG UX パターンの定義方法

NFRE-E08（animate-pulse + red color）と NFRE-E09（RaiseModal 強制表示）を NFR Design でどう扱うか。

A) 新規パターン `P-ME-FE-DEG-01` にまとめ、ThresholdExceeded 条件 / CSS クラス / RaiseModal 自動オープン / ダメコピーを 1 パターンとして明文化する
B) NFRE-E08 / E09 は nfr-requirements.md §5 に詳細があるので、logical-components の Frontend 記述に条件だけ書けば十分。新規パターン不要
X) Other

[Answer]: A — 新規パターン `P-ME-FE-DEG-01` として定義（ThresholdExceeded → CSS クラス切替 / BudgetEmptyScreen useState(true) / ダメコピー一覧 を 1 パターンに集約）

---

### Q-DD4: MetricsService / BudgetRaiseService の LogSummary

各エンドポイントのリクエスト単位で構造化ログを集約する LogSummary（P-OBS-02 流用）を定義するか。

A) P-ME-OBS-01 として定義し、共通 8 項目 + Unit E 固有項目（action / damageCount / consumptionRate / thresholdExceeded / raisedAmount 等）を明示する
B) Unit E は軽量実装のため LogSummary 定義不要。標準 slog.Info で対応し P-OBS-02 の再利用のみ記述する
X) Other

[Answer]: B — LogSummary 定義なし。Unit E は観測性 NFR なし（NFRE-E01 は latency 目標のみ）。標準 slog.Info + P-OBS-01 latency 計測で対応

---

### Q-DD5: Logical Components の粒度

Unit E の Logical Components をどの粒度で定義するか（LC-ME-xx）。

A) Backend: MetricsService / MetricsHandler / BudgetRaiseService / BudgetRaiseHandler / Metrics DTO + BudgetRaiseResult DTO（6 コンポーネント）。Frontend: useMetrics / useBudgetRaise / MetricsPanel / BudgetEmptyScreen / RaiseModal（5 コンポーネント）。計 11 LC
B) Backend: MetricsService + Handler を統合 / BudgetRaiseService + Handler を統合（4 コンポーネント）。Frontend は同様（5 コンポーネント）。計 9 LC（粗め）
C) Handler は独立 LC として切り出さず、Service LC の責務説明に handler の役割も含める（Backend 3 + Frontend 5 = 8 LC、最も粗め）
X) Other

[Answer]: A — Backend: MetricsService / MetricsHandler / BudgetRaiseService / BudgetRaiseHandler / Metrics DTO / BudgetRaiseResult DTO（6 LC）+ Frontend: useMetrics / useBudgetRaise / MetricsPanel / BudgetEmptyScreen / RaiseModal（5 LC）= 計 11 LC

---

## 5. 矛盾チェックメモ（回答収集後に記入）

| 観点 | チェック内容 | 結果 |
|---|---|---|
| Q-DD1 × NFRE-E01 | 採用 OBS パターンと P95 目標の整合 | ✅ P-OBS-01 latency 計測を採用し NFRE-E01（P95 500ms）達成の観測手段を確保。矛盾なし |
| Q-DD2 × Q-DD5 | errgroup パターンと MetricsService LC の責務整合 | ✅ P-ME-PARALLEL-01 は MetricsService（LC-ME-01）の実装パターンとして帰属。LC 粒度と整合 |
| Q-DD3 × Unit C P-FE-* | DEG UX パターンと既存 FE パターンの重複排除 | ✅ P-ME-FE-DEG-01 は loading/lock パターンとは別軸（状態スタイル）。P-FE-LOAD-01 は useMetrics の isLoading に再利用し重複なし |
| Q-DD4 × Q-DD1 | LogSummary 採否と OBS 系採用方針の整合 | ✅ Q-DD4=B（LogSummary なし）と Q-DD1=A（P-OBS-01 latency のみ）は一貫。latency は slog.Info 1 行に含めるのみ |
| Q-DD5 × unit-interfaces.md §6 | LC 粒度と凍結契約の公開 interface 整合 | ✅ MetricsService（凍結契約 §6.1）/ BudgetRaiseService（§6.2）を独立 LC として定義。Handler は内部実装として分離。矛盾なし |
