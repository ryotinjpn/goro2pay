# Unit E `metrics` — Functional Design Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard
**Stage**: Functional Design (Construction Phase, per-unit)

---

## 0. このドキュメントの目的

Unit E `metrics` の Functional Design ステージで生成する成果物（business-logic-model.md / business-rules.md / domain-entities.md / frontend-components.md）の前段として、ビジネスロジックを確定するために必要な意思決定を質問形式で収集する。

ヒアリング方針: チャットで 1 問ずつ提示する。回答受領のたびに本ファイルの `[Answer]:` を更新する。

---

## 1. Plan 実施チェックリスト

- [x] Step 1: Unit E の前段成果物を読み込み済み（unit-of-work.md / unit-of-work-story-map.md / unit-interfaces.md / requirements.md / stories.md）
- [x] Step 2: 業務ロジック上の確認事項を質問形式で抽出（Q-F1〜Q-F8）
- [x] Step 3: 質問を本 Plan ファイルに記載
- [x] Step 4: チャットで 1 問ずつ提示してヒアリング（Q-F1〜Q-F8 全問完了）
- [x] Step 5: 回答を `[Answer]:` タグへ反映、矛盾検出（全 6 観点で整合確認）
- [x] Step 6: 矛盾解消後、Functional Design 成果物の生成準備完了

---

## 2. Unit E が担当するストーリー（再掲）

| ID | ストーリー要約 | 主要関心 |
|---|---|---|
| US-3-01 | 今月のダメ化回数を見て優越感と空虚感を同時に感じる | MetricsService.GetMetrics / DamageCount 集計 |
| US-3-02 | ダメ予算消化率を見る | ConsumptionRate 計算 / 80% 超警告色 |
| US-3-03 | 残高 0 円時に「今月ダメになれません」画面を見る | BudgetEmptyScreen 遷移トリガー |
| US-3-04 | 翌月予算の増額誘導モーダルで意思薄弱になる | BudgetRaiseService / RaiseModal UX |
| US-X-03 | 自力で何もできない自分に気づき、退化ループを完成する | 増額受諾フロー / NFR-DEG-04 |

---

## 3. 凍結済み前提（unit-interfaces.md より、質問不要）

| 項目 | 確定値 |
|---|---|
| `MetricsService.GetMetrics` シグネチャ | `(ctx, userID) -> (*Metrics, error)` |
| `Metrics` 構造体フィールド | DamageCount / ConsumptionRate(0..1) / MonthlyBudget / RemainingBalance / ThresholdExceeded(>0.8) / SummaryText |
| `BudgetRaiseService` シグネチャ | `ComputeRecommendedBudget` / `Accept(ctx, userID, newMonthlyBudget) -> BudgetRaiseResult` |
| `BudgetRaiseResult` | NewMonthlyBudget / AppliedFrom(翌月1日 JST) |
| REST エンドポイント | GET /api/metrics / GET /api/budget/raise/recommendation / POST /api/budget/raise |
| 新規 DynamoDB テーブル | なし（既存テーブルの読取参照のみ） |
| 内部依存 | WalletReader / BudgetSettingsReader / OrderHistoryReader / BudgetSettingsWriter |
| 予算上限 | 100,000 円（unit-of-work.md 確定済み） |
| 推奨増額式 | `min(currentBudget * 1.5, 100_000)` 円（unit-of-work.md 確定済み） |

---

## 4. 質問

> ヒアリングは下記の Q 順にチャットで 1 問ずつ実施する。回答ごとに `[Answer]:` を更新する。

### Q-F1: DamageCount（ダメ化回数）の集計範囲

`MetricsService.GetMetrics` で返す `DamageCount` は何を集計するか。Unit B の月初リセットは EventBridge Scheduler (月初 00:00 JST) で動作する。

A) 当月 1 日 00:00 JST 〜 現在 を `OrderHistoryReader` の `CreatedAt` で絞り込んで件数を返す（カレンダー月ベース、JST タイムゾーン固定）
B) 前回リセットイベント時刻 〜 現在 を集計する（`BudgetResetLog` の最終エントリを参照してカットオフを決定する）
C) OrderHistory にある全件を返す（月次リセット考慮なし、シンプル実装）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — カレンダー月ベース (JST)、OrderHistoryReader.CreatedAt で当月 1 日以降を絞り込み

---

### Q-F2: ConsumptionRate のエッジケース処理

`ConsumptionRate = (monthlyBudget - remainingBalance) / monthlyBudget` の算出において、以下のエッジケースをどう扱うか。

A) `monthlyBudget = 0` → `ConsumptionRate = 0.0` で固定（初回未設定ユーザーは常に安全扱い）
B) `monthlyBudget = 0` → `ErrNoBudgetSet` エラーを返す（未設定ユーザーは必ず BudgetSetupScreen に誘導）
C) `remainingBalance < 0`（仮に発生した場合）→ `ConsumptionRate = 1.0` にクランプ
D) A + C（0 除算は 0.0 固定 + 負残高は 1.0 にクランプ、両方適用）
X) Other (please describe after [Answer]: tag below)

[Answer]: B — monthlyBudget = 0 の場合は ErrNoBudgetSet を返す。なお remainingBalance < 0 は Unit B の不変条件（NFR-REL-01）で発生しないため追加クランプは不要

---

### Q-F3: SummaryText の生成場所とフォーマット

`Metrics.SummaryText` は unit-interfaces.md で「今月のダメ化回数: 12 回、消化額 ¥24,600」例示がある。どこで生成するか。

A) Backend (`MetricsService`) で生成して文字列として返す（フォーマットはサーバ固定）
B) Frontend で `DamageCount` / `MonthlyBudget` / `RemainingBalance` から組み立てる（`SummaryText` フィールドは空文字か廃止）
C) Backend でテンプレート文字列を返すが、フロントがローカライズできるように数値フィールドも並列で返す（現状 A と同じ、将来の多言語対応への布石）
X) Other (please describe after [Answer]: tag below)

[Answer]: C — Backend が SummaryText を生成しつつ数値フィールドも並列で返す。Frontend は SummaryText を表示するが、必要なら数値フィールドから独自フォーマットも組み立て可能

---

### Q-F4: BudgetEmptyScreen 遷移トリガー

US-3-03「残高 0 円時に今月ダメになれません画面を見る」のトリガー。残高 0 の検知をどこで行うか。

A) Frontend が `useWallet` / `useMetrics` の `remainingBalance === 0` を監視し、`router.push('/budget-empty')` でリダイレクト（クライアントサイドルーティング）
B) GET /api/metrics レスポンスに `shouldRedirectToEmpty: boolean` を追加し、Frontend はそのフラグに従ってリダイレクト
C) 注文時 (POST /api/orders) に 402 が返った場合（Unit C がすでに実装済み）のみ BudgetEmpty へ遷移。メトリクス画面からは誘導しない
X) Other (please describe after [Answer]: tag below)

[Answer]: A — useMetrics の remainingBalance === 0 を MainScreen で監視し router.push('/budget-empty') でリダイレクト

---

### Q-F5: RaiseModal の表示トリガー

US-3-04「翌月予算の増額誘導モーダル」はいつ表示するか（NFR-DEG-04: 退化ループ）。

A) BudgetEmptyScreen に到達したときに自動表示（残高 0 になったユーザーへ自動的に誘導）
B) BudgetEmptyScreen に「翌月の予算を増やす」ボタンを設置し、ユーザー操作で表示
C) MetricsPanel の `ThresholdExceeded = true`（80% 超）のタイミングでも RaiseModal を表示（残高 0 になる前から誘導）
D) A + C（BudgetEmpty 到達で自動 + 80% 超でも誘導）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — BudgetEmptyScreen マウント時に自動で RaiseModal を表示。ユーザーは「増やす」以外の選択肢を考える間もなく退化ループへ誘導

---

### Q-F6: 増額 Accept フローの EffectiveFrom 計算

`BudgetRaiseService.Accept` が返す `BudgetRaiseResult.AppliedFrom` は「翌月 1 日 00:00 JST」。この計算はどこで行うか。

A) Backend (`BudgetRaiseService`) がサーバ時刻を JST で翌月 1 日に丸めて返す。Client はこの値をそのまま表示する
B) Client が翌月 1 日を計算して POST リクエストに `effectiveFrom` を含める。Backend は受け取った値をそのまま BudgetSettings に書く
C) A と同じだが、即時適用のオプションも用意する（将来拡張、今は翌月固定で実装）
X) Other (please describe after [Answer]: tag below)

[Answer]: A — Backend が JST で翌月 1 日 00:00 を計算して AppliedFrom に詰める。Unit B SetBudget のサーバサイド計算と一貫

---

### Q-F7: MetricsPanel の更新頻度と MainScreen 上の配置

`MetricsPanel.tsx` は `app/page.tsx`（MainScreen）に配置される。どのタイミングで最新データを取得するか。

A) アプリ起動（MainScreen マウント）時に 1 度取得、以降はユーザー操作で手動リフレッシュ（シンプル実装）
B) MainScreen マウント時に取得 + 注文完了後（`useOrder` の mutation 成功）に TanStack Query の invalidateQueries で自動更新
C) 一定間隔 polling（例: 30 秒）で定期取得（残高をリアルタイム反映）
X) Other (please describe after [Answer]: tag below)

[Answer]: B — マウント時取得 + useOrder mutation 成功時に invalidateQueries(['metrics']) で自動更新

---

### Q-F8: 新規ユーザー（BudgetSettings 未設定）の GET /api/metrics 動作

ダメ予算を一度も設定していないユーザーが GET /api/metrics を呼んだ場合の動作。

A) HTTP 404 または専用エラー（`ErrNoBudgetSet`）を返す。Frontend はこのエラーを受け取り BudgetSetupScreen へ誘導する
B) `Metrics` をゼロ値（DamageCount=0, ConsumptionRate=0.0, MonthlyBudget=0, RemainingBalance=0）で返す。Frontend はゼロ値を表示するだけで遷移しない
C) HTTP 200 で `Metrics` ゼロ値を返しつつ、`needsBudgetSetup: true` フィールドを追加してフロントが BudgetSetupScreen に誘導
X) Other (please describe after [Answer]: tag below)

[Answer]: A — HTTP 400 + エラーコード "ERR_NO_BUDGET_SET" を返す。useMetrics の isError で BudgetSetupScreen (/budget) へリダイレクト

---

## 5. 矛盾チェックメモ（回答収集後に記入）

| 観点 | チェック内容 | 結果 |
|---|---|---|
| Q-F1 × Q-F2 | 集計期間と ConsumptionRate の分母が同一月ベースか | ✅ 整合。DamageCount は当月 OrderHistory 件数、ConsumptionRate の分母は BudgetSettings.MonthlyBudget（月次リセット済み値）で独立して問題なし |
| Q-F3 × Q-F4 | SummaryText 生成場所と BudgetEmpty トリガーの独立性 | ✅ 整合。SummaryText は Backend 生成、BudgetEmpty トリガーは Frontend の remainingBalance===0 判定で独立 |
| Q-F4 × Q-F5 | BudgetEmpty 遷移と RaiseModal 表示の順序整合 | ✅ 整合。useMetrics → remainingBalance===0 → router.push('/budget-empty') → BudgetEmptyScreen マウント → RaiseModal 自動表示 の順序で矛盾なし |
| Q-F6 × Unit B Q-B2 | EffectiveFrom 計算が Unit B BudgetSettings 書き込みと整合するか | ✅ 整合。Unit B Q-B2=B では変更時は BudgetSettings.EffectiveFrom=翌月1日を即時更新。Unit E Accept も同様にサーバサイドで JST 翌月1日を計算して BudgetSettingsWriter.Set を呼ぶ |
| Q-F7 × Q-F4 | MetricsPanel 更新後に BudgetEmpty リダイレクトが正しく発火するか | ✅ 整合。useOrder mutation 成功 → invalidateQueries(['metrics']) → useMetrics 再フェッチ → remainingBalance===0 なら router.push('/budget-empty') の順序で発火 |
| Q-F8 × Q-F2 | 未設定ユーザーの返却値と ConsumptionRate 計算の整合 | ✅ 整合。Q-F2=B で monthlyBudget=0 → ErrNoBudgetSet、Q-F8=A で HTTP 400 を返す。同一エラーパスで統一されており重複なし |
