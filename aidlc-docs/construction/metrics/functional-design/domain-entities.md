# Domain Entities — Unit E `metrics`

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard

---

## 1. Unit E が所有するエンティティ

Unit E は新規 DynamoDB テーブルを持たない。既存テーブル（Unit B / C 所有）を読取参照のみ行う。

---

## 2. 出力 DTO（unit-interfaces.md 凍結済み）

### 2.1 Metrics

```go
package metrics

type Metrics struct {
    DamageCount       int
    ConsumptionRate   float64 // 0.0 .. 1.0、clamp 済み
    MonthlyBudget     int     // 円
    RemainingBalance  int     // 円
    ThresholdExceeded bool    // ConsumptionRate > 0.8
    SummaryText       string  // "今月のダメ化回数: N 回、消化額 ¥X,XXX"
}
```

**JSON レスポンス例**:
```json
{
  "damageCount": 12,
  "consumptionRate": 0.82,
  "monthlyBudget": 30000,
  "remainingBalance": 5400,
  "thresholdExceeded": true,
  "summaryText": "今月のダメ化回数: 12 回、消化額 ¥24,600"
}
```

---

### 2.2 BudgetRaiseResult

```go
package budget_raise

type BudgetRaiseResult struct {
    NewMonthlyBudget int
    AppliedFrom      time.Time // 翌月 1 日 00:00:00 JST
}
```

**JSON レスポンス例**:
```json
{
  "newMonthlyBudget": 45000,
  "appliedFrom": "2026-06-01T00:00:00+09:00"
}
```

---

### 2.3 RecommendationResponse

```go
// GET /api/budget/raise/recommendation のレスポンス（DTO）
type RecommendationResponse struct {
    CurrentMonthlyBudget     int
    RecommendedMonthlyBudget int
}
```

**JSON レスポンス例**:
```json
{
  "currentMonthlyBudget": 30000,
  "recommendedMonthlyBudget": 45000
}
```

---

## 3. 読取参照する他 Unit のエンティティ

### 3.1 WalletReader（Unit B 所有）

```go
package wallet_repo

type WalletReader interface {
    Get(ctx context.Context, userID string) (*Wallet, error)
}

type Wallet struct {
    UserID           string
    Balance          int     // 残高（円）
    // 他フィールドは Unit B 内部
}
```

Unit E が参照するフィールド: `Balance`（→ `RemainingBalance` として利用）

---

### 3.2 BudgetSettingsReader（Unit B 所有）

```go
package budget_settings

type BudgetSettingsReader interface {
    Get(ctx context.Context, userID string) (*BudgetSettings, error)
}

type BudgetSettings struct {
    UserID        string
    MonthlyBudget int
    // 他フィールドは Unit B 内部
}
```

Unit E が参照するフィールド: `MonthlyBudget`

---

### 3.3 BudgetSettingsWriter（Unit B が Unit E 向けに公開）

```go
package budget_settings

type BudgetSettingsWriter interface {
    Set(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error
}
```

`BudgetRaiseService.Accept` のみが使用する書き込みインターフェース。

---

### 3.4 OrderHistoryReader（Unit C 所有）

unit-interfaces.md で Unit D / E 向けに公開済みのインターフェース（変更不要）:

```go
package order_history

type OrderHistoryReader interface {
    ListRecent(ctx context.Context, userID string, limit int) ([]OrderRecord, error)
    CountThisMonth(ctx context.Context, userID string) (int, error)
    SumThisMonth(ctx context.Context, userID string) (int, error)
}
```

Unit E が使用するメソッド: `CountThisMonth`（DamageCount 取得）。`SumThisMonth` は参照しない（消化額は `monthlyBudget - remainingBalance` で Wallet から計算）。

---

## 4. 内部依存関係まとめ

```
MetricsService
  +-- WalletReader          (Unit B 所有、読取のみ)
  +-- BudgetSettingsReader  (Unit B 所有、読取のみ)
  +-- OrderHistoryReader    (Unit C 所有、読取のみ / CountSince 追加必要)

BudgetRaiseService
  +-- BudgetSettingsReader  (Unit B 所有、読取)
  +-- BudgetSettingsWriter  (Unit B が Unit E 向けに公開)
```

---

## 5. エラー型

```go
package metrics

var (
    ErrNoBudgetSet  = errors.New("no budget set")
    ErrInvalidBudget = errors.New("invalid budget amount")
)
```
