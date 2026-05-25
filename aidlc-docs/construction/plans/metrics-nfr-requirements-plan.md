# Unit E `metrics` — NFR Requirements Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard
**Stage**: NFR Requirements (Construction Phase, per-unit)

---

## 0. このドキュメントの目的

Unit E `metrics` の NFR Requirements ステージで生成する成果物（nfr-requirements.md / tech-stack-decisions.md）の前段として、非機能要件とテックスタック選定に必要な意思決定を収集する。

**前提**: Unit A/B/C で確定済みのテックスタック（Go / Gin / DynamoDB SDK v2 / Next.js / TanStack Query）を継承する。Unit E は新規インフラなし・読取専用の Standard 深度 Unit のため、確認事項は最小限。

---

## 1. Plan 実施チェックリスト

- [x] Step 1: Functional Design 成果物を読み込み済み
- [x] Step 2: NFR 確認事項を抽出（Q-N1〜Q-N5）
- [x] Step 3: 質問を本 Plan ファイルに記載
- [x] Step 4: チャットで 1 問ずつ提示してヒアリング（Q-N1〜Q-N5 全問完了）
- [x] Step 5: 回答を `[Answer]:` タグへ反映、矛盾検出（全 4 観点で整合確認）
- [x] Step 6: NFR Requirements 成果物の生成準備完了

---

## 2. 継承済みテックスタック（確認不要）

| 項目 | 確定値 | 出典 |
|---|---|---|
| 言語 | Go 1.24 | Unit A/B/C |
| フレームワーク | Gin | Unit A |
| DynamoDB SDK | aws-sdk-go-v2 | Unit A/B/C |
| Frontend | Next.js 14 / TanStack Query v5 / Jotai | Unit A/B/C |
| テスト | Vitest (Frontend) / Go testing (Backend) | Unit A/B/C |
| IaC | Terraform | Unit A |
| 認証・認可 | Unit A の AttachUserID middleware | Unit A |
| 新規 DynamoDB テーブル | なし | unit-of-work.md |

---

## 3. 質問

### Q-N1: GET /api/metrics の DynamoDB 読み取り並列化

`MetricsService.GetMetrics` は内部で 3 回の DynamoDB 読み取りを行う（WalletReader / BudgetSettingsReader / OrderHistoryReader.CountThisMonth）。これを直列 vs 並列で実行するか。

A) `sync.WaitGroup` / `errgroup` で 3 読み取りを並列実行（レイテンシ最小化、Unit C P-INIT-01 の方針と整合）
B) 直列実行（シンプル実装、ハッカソン規模では差異なし）
C) BudgetSettings と Wallet のみ並列、CountThisMonth は後から直列（BudgetSettings が NotFound なら CountThisMonth は不要なため）
X) Other (please describe after [Answer]: tag below)

[Answer]: C — BudgetSettings + Wallet を errgroup で並列取得 → NotFound チェック後に CountThisMonth を直列呼び出し

---

### Q-N2: DynamoDB 部分障害時の挙動

WalletReader / BudgetSettingsReader / OrderHistoryReader のいずれか 1 つが DynamoDB エラーを返した場合。

A) いずれかが失敗したら即座に HTTP 500 を返す（全 or 何もなし）
B) BudgetSettings / Wallet の失敗は 500、OrderHistoryReader の失敗は DamageCount=0 で部分返却（メトリクス表示は続ける）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — いずれかの DynamoDB 読み取りが失敗したら HTTP 500。部分返却なし

---

### Q-N3: PBT（Property-Based Testing）適用対象

Unit E には以下の純関数がある（PBT Partial が有効）。どこに適用するか。

A) `ConsumptionRate` 計算（`0.0 <= rate <= 1.0` の不変条件）と `ComputeRecommendedBudget`（`result >= current` かつ `result <= 100_000` の不変条件）の 2 関数に適用
B) `ComputeRecommendedBudget` のみ（入出力が単純で PBT の恩恵が大きい）
C) PBT は適用しない（Unit E の純関数はテーブルテストで十分）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — ConsumptionRate と ComputeRecommendedBudget の 2 関数に gopter で PBT 適用

---

### Q-N4: NFR-DEG-03 警告色の具体仕様

`ThresholdExceeded = true`（消化率 80% 超）のときの `MetricsPanel` 演出。

A) プログレスバーと消化率テキストを赤（`#EF4444` / Tailwind `text-red-500`）に変更するのみ（シンプル）
B) A + パルスアニメーション（Tailwind `animate-pulse`）でさらに不安を煽る（NFR-DEG-03: 不安の演出を強化）
X) Other (please describe after [Answer]: tag below)

[Answer]: B — 赤色 (text-red-500 / bg-red-500) + animate-pulse で不安を最大化

---

### Q-N5: RaiseModal の再表示制御

BudgetEmptyScreen で RaiseModal を「今月はがんばる」で閉じた後、再度 BudgetEmptyScreen に来た場合の挙動。

A) 再度マウントしたら再表示（セッション内で何度でも表示、退化ループを最大化）
B) `sessionStorage` にフラグを持ち、同一セッション内で 1 度閉じたら再表示しない
X) Other (please describe after [Answer]: tag below)

[Answer]: A — BudgetEmptyScreen マウントのたびに RaiseModal を再表示。逃げられない退化ループ

---

## 4. 矛盾チェックメモ（回答収集後に記入）

| 観点 | チェック内容 | 結果 |
|---|---|---|
| Q-N1 × Q-N2 | 並列実行時のエラーハンドリング方針の整合 | ✅ 整合。errgroup で並列取得し、いずれかのエラーを errgroup が捕捉して即 500 返却。矛盾なし |
| Q-N3 × FD business-rules.md | PBT 対象関数が純粋関数であることの確認 | ✅ 整合。ConsumptionRate 計算（BR-M03）と ComputeRecommendedBudget（BR-R01）はいずれも外部 I/O なし・副作用なしの純関数 |
| Q-N4 × NFR-DEG-03 | 警告色仕様が要件定義の NFR-DEG-03 と整合するか | ✅ 整合。NFR-DEG-03「不安の演出」に animate-pulse が加わることで強化される |
| Q-N5 × FD Q-F5 | RaiseModal 再表示と退化ループ（NFR-DEG-04）の整合 | ✅ 整合。マウントのたびに再表示は Q-F5=A（BudgetEmpty 到達で自動表示）の強化版。sessionStorage なしでシンプル |
