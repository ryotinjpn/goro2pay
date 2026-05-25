# Business Rules — Unit E `metrics`

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard

---

## 1. MetricsService ビジネスルール

### BR-M01: 予算未設定ユーザーへのメトリクス提供禁止

**ルール**: `BudgetSettings` が存在しない、または `monthlyBudget = 0` のユーザーに対して `GetMetrics` は `ErrNoBudgetSet` を返さなければならない。

**根拠**: 予算未設定状態でメトリクスを返すと ConsumptionRate の分母がゼロになり、ゼロ除算が発生する。また未設定ユーザーはまず BudgetSetupScreen で予算を設定する必要がある（US-0-03）。

**HTTP**: 400 + `{"code": "ERR_NO_BUDGET_SET"}`

---

### BR-M02: DamageCount は当月 JST カレンダーベースで集計

**ルール**: DamageCount は `OrderHistory.CreatedAt >= 当月 1 日 00:00:00 JST` の件数とする。月の境界は Asia/Tokyo タイムゾーンで決定する。

**根拠**: 月初リセット（Unit B EventBridge Scheduler）と同じ月の境界定義を使うことで、ユーザー体験上の一貫性を保つ（Q-F1=A）。

---

### BR-M03: ConsumptionRate のクランプ

**ルール**:
- `ConsumptionRate = clamp((monthlyBudget - remainingBalance) / monthlyBudget, 0.0, 1.0)`
- `monthlyBudget > 0` は BR-M01 で保証済み
- `remainingBalance < 0` は Unit B NFR-REL-01 で発生しないが、防御的に `rate = min(max(rate, 0.0), 1.0)` を適用する

---

### BR-M04: ThresholdExceeded の判定閾値

**ルール**: `ThresholdExceeded = ConsumptionRate > 0.8`

**根拠**: 消化率 80% 超で警告色演出（NFR-DEG-03: 不安の演出）。厳密な等号は含まない（`> 0.8`、`= 0.8` では通常色）。

---

### BR-M05: SummaryText のフォーマット

**ルール**: `"今月のダメ化回数: {DamageCount} 回、消化額 ¥{spent:,}"` の固定フォーマットで Backend が生成する。`spent = monthlyBudget - remainingBalance`。金額は 3 桁カンマ区切り。

**例**: `"今月のダメ化回数: 12 回、消化額 ¥24,600"`

---

## 2. BudgetRaiseService ビジネスルール

### BR-R01: 増額推奨値の上限

**ルール**: `recommendedBudget = min(floor(currentBudget * 1.5), 100_000)`

**根拠**: unit-of-work.md で確定済み（NFR-DEG-04: 退化ループ）。上限 100,000 円は BR-R02 のバリデーション上限と一致させる。

---

### BR-R02: Accept の入力バリデーション

**ルール**: `1 <= newMonthlyBudget <= 100,000` を満たさない場合は `ErrInvalidBudget` を返す。

**理由**: Unit B `SetBudget` の許容範囲（Q-B1=A: `1 ≤ monthlyBudget ≤ 100,000`）と統一する。

---

### BR-R03: EffectiveFrom はサーバサイドで決定

**ルール**: `BudgetRaiseService.Accept` は `effectiveFrom = 翌月 1 日 00:00:00 JST` をサーバ時刻から計算する。クライアントからの入力は受け付けない。

**根拠**: タイムゾーンずれ・クライアント時刻改ざんを防止（Q-F6=A）。Unit B Q-B2=B の設計と一貫する。

---

### BR-R04: 増額は翌月から適用、当月残高は変更しない

**ルール**: `Accept` の呼び出しは `BudgetSettingsWriter.Set(userID, newMonthlyBudget, effectiveFrom)` を呼ぶのみ。当月の `Wallet.Balance` は変更しない。

**根拠**: Unit B Q-B2=B では「BudgetRaise による変更は翌月リセット時から適用」と定義（BudgetRaise は増額のみの特殊フロー。当月即時加算は `SetBudget` のルールであり、BudgetRaise の性質とは異なる）。

---

## 3. フロントエンド ビジネスルール

### BR-FE01: BudgetEmpty 遷移条件

**ルール**: `useMetrics` が `remainingBalance === 0` を返した場合、Frontend は直ちに `/budget-empty` へ `router.push` する。

**補足**: `isError` かつ `code === "ERR_NO_BUDGET_SET"` の場合は `/budget` へリダイレクトする（BR-M01 と対応）。

---

### BR-FE02: RaiseModal の自動表示

**ルール**: `BudgetEmptyScreen` のマウント時に `RaiseModal` を自動的にオープン状態で表示する。

**根拠**: ユーザーが意思決定する前に増額の選択肢を提示することで退化ループへの誘導を最大化する（NFR-DEG-04 / NFR-DEG-01）。

---

### BR-FE03: MetricsPanel の自動更新

**ルール**: `useOrder` の mutation 成功時に `queryClient.invalidateQueries({ queryKey: ['metrics'] })` を呼び出す。MetricsPanel は次のレンダリングサイクルで最新データを表示する。

**根拠**: 注文完了後に DamageCount と残高が即時反映されないと数値の不整合が生じる（Q-F7=B）。

---

## 4. エラーコード一覧

| コード | HTTP | 発生箇所 | 意味 |
|---|---|---|---|
| `ERR_NO_BUDGET_SET` | 400 | MetricsHandler / BudgetRaiseHandler | 予算未設定 |
| `ERR_INVALID_BUDGET` | 400 | BudgetRaiseHandler | newMonthlyBudget が範囲外 |
| `ERR_UNAUTHORIZED` | 401 | Unit A middleware | 認証失敗 |
