# Component Methods — ゴロゴロPay

**Document Version**: 1.0
**Created**: 2026-05-07

本ドキュメントは各コンポーネントの **メソッドシグネチャ（Go interface ベース）** と高レベルな入出力仕様を記述する。詳細なビジネスロジック（残高計算の具体的なアルゴリズム、Bedrock プロンプト文面、エラーコード体系等）は Construction フェーズの Functional Design（per-unit）で展開する。

**記述方針**:
- Go の関数シグネチャをそのまま記載（型は仮、Functional Design で最終化）
- `context.Context` は全ての I/O 伴うメソッドに必須
- エラーは Go 慣例の `(result, error)` 形式
- 構造体名は `PascalCase`、メソッド名も `PascalCase`、定義場所のパッケージ名は `snake_case`

---

## 1. Gin Handlers

### 1.1 HealthHandler

```go
package handlers

type HealthHandler struct{}

// GET /health
func (h *HealthHandler) Check(c *gin.Context)
```
- **Input**: なし
- **Output**: `200 {status: "ok", version: "x.y.z"}`
- **認証**: 不要

### 1.2 WalletHandler

```go
package handlers

type WalletHandler struct {
    walletService WalletService
}

// GET /wallet
func (h *WalletHandler) GetBalance(c *gin.Context)

// POST /wallet/budget
func (h *WalletHandler) SetBudget(c *gin.Context)
```
- `GetBalance`: userId (JWT claims) → `200 {balance: int, monthlyBudget: int}`
- `SetBudget`: `{monthlyBudget: int}` → `200 {monthlyBudget: int}`

### 1.3 OrderHandler

```go
package handlers

type OrderHandler struct {
    orderService OrderService
}

// POST /orders
func (h *OrderHandler) PlaceOrder(c *gin.Context)

// GET /orders
func (h *OrderHandler) GetHistory(c *gin.Context)
```
- `PlaceOrder`: `{category: "food", idempotencyKey: string, suggestionId?: string}` → `201 {orderId, storeName, menuName, amount, remainingBalance}`
- `GetHistory`: クエリ `?limit=20` → `200 {items: [...]}`

### 1.4 MetricsHandler

```go
package handlers

type MetricsHandler struct {
    metricsService MetricsService
}

// GET /metrics
func (h *MetricsHandler) GetMetrics(c *gin.Context)
```
- **Output**: `200 {damageCount: int, consumptionRate: float, monthlyBudget: int, remainingBalance: int, thresholdExceeded: bool}`

### 1.5 SuggestHandler

```go
package handlers

type SuggestHandler struct {
    suggestService SuggestService
}

// GET /suggest
func (h *SuggestHandler) GetSuggestion(c *gin.Context)
```
- **Output**: `200 {hasSuggestion: bool, suggestion?: {title, storeName, menuName, amount, suggestionId}}`
- `hasSuggestion=false` の場合、履歴不足のためデフォルト UI 表示

### 1.6 BudgetRaiseHandler

```go
package handlers

type BudgetRaiseHandler struct {
    budgetRaiseService BudgetRaiseService
}

// POST /budget/raise
func (h *BudgetRaiseHandler) Accept(c *gin.Context)
```
- **Input**: `{newMonthlyBudget: int}` (推奨額をそのまま採用、または独自額)
- **Output**: `200 {newMonthlyBudget: int, appliedFrom: "YYYY-MM-01"}`

---

## 2. Services

### 2.1 AuthContextService (Gin middleware)

```go
package auth

// AttachUserID は JWT claims から userId を抽出し gin.Context に載せる
func AttachUserID() gin.HandlerFunc

// UserIDFromContext は middleware で載せた userId を取り出す
func UserIDFromContext(c *gin.Context) (string, error)
```

### 2.2 WalletService

```go
package wallet

type WalletService interface {
    // GetBalance は指定ユーザの残高と月間予算を取得する
    GetBalance(ctx context.Context, userID string) (*WalletSnapshot, error)

    // SetBudget は月間予算を更新する（当月残高は据え置き、翌月リセットから適用）
    SetBudget(ctx context.Context, userID string, monthlyBudget int) error

    // Deduct は条件付き書き込みで残高を減算する
    // idempotencyKey により二重引き落としを防止
    // 残高不足時は ErrInsufficientBalance を返す
    Deduct(ctx context.Context, userID string, amount int, idempotencyKey string) (*DeductResult, error)

    // ResetAll は全ユーザの残高を monthlyBudget にリセットする（スケジューラ用）
    ResetAll(ctx context.Context) (*ResetResult, error)
}

type WalletSnapshot struct {
    UserID         string
    Balance        int
    MonthlyBudget  int
    UpdatedAt      time.Time
}

type DeductResult struct {
    NewBalance     int
    Idempotent     bool  // true の場合、既存の結果を返した
}

type ResetResult struct {
    ProcessedUsers int
    Errors         []error
}
```

### 2.3 OrderService

```go
package order

type OrderService interface {
    // PlaceOrder は代行手配ユースケースの全体をオーケストレーションする
    // 1. Bedrock で最適プラン推論 (or suggestionId を解決)
    // 2. Wallet 減算（idempotencyKey 付き）
    // 3. DeliveryAdapter で外部手配
    // 4. 履歴記録
    PlaceOrder(ctx context.Context, userID string, req PlaceOrderRequest) (*PlaceOrderResult, error)

    // GetHistory は指定ユーザの注文履歴を取得する
    GetHistory(ctx context.Context, userID string, limit int) ([]*OrderRecord, error)
}

type PlaceOrderRequest struct {
    Category        string   // "food" 等
    IdempotencyKey  string
    SuggestionID    *string  // Suggest からの 1 タップ注文時
}

type PlaceOrderResult struct {
    OrderID            string
    StoreName          string
    MenuName           string
    Amount             int
    RemainingBalance   int
    Idempotent         bool
}

type OrderRecord struct {
    OrderID    string
    UserID     string
    Category   string
    StoreName  string
    MenuName   string
    Amount     int
    OrderedAt  time.Time
}
```

### 2.4 SuggestService

```go
package suggest

type SuggestService interface {
    // GetSuggestion は起動時の先回り提案を生成する
    // 履歴不足なら hasSuggestion=false を返す
    // Bedrock 失敗時はフォールバック（履歴最頻）を返す
    GetSuggestion(ctx context.Context, userID string) (*Suggestion, error)

    // ResolveSuggestion は Suggest 経由の注文時、saved suggestion を復元する
    ResolveSuggestion(ctx context.Context, suggestionID string) (*SuggestionPlan, error)
}

type Suggestion struct {
    HasSuggestion  bool
    SuggestionID   string
    Title          string  // 例: "そろそろご飯めんどくさいですよね？"
    Plan           *SuggestionPlan
    FallbackUsed   bool    // 内部ログ用
}

type SuggestionPlan struct {
    StoreName  string
    MenuName   string
    Amount     int
    Category   string
}
```

### 2.5 MetricsService

```go
package metrics

type MetricsService interface {
    // GetMetrics は今月のダメ化回数・消化率を返す
    GetMetrics(ctx context.Context, userID string) (*Metrics, error)
}

type Metrics struct {
    DamageCount         int
    ConsumptionRate     float64  // 0.0 - 1.0
    MonthlyBudget       int
    RemainingBalance    int
    ThresholdExceeded   bool     // ConsumptionRate > 0.8
    SummaryText         string   // "今月のダメ化回数: 47 回、消化額 ¥30,000"
}
```

### 2.6 BudgetRaiseService

```go
package budget_raise

type BudgetRaiseService interface {
    // ComputeRecommendedBudget は現予算と消化実績から推奨額を算出する
    // 本 MVP: min(current * 1.5, cap 100_000)
    ComputeRecommendedBudget(ctx context.Context, userID string) (int, error)

    // Accept は新しい月間予算を設定する（翌月リセットから適用）
    Accept(ctx context.Context, userID string, newMonthlyBudget int) (*BudgetRaiseResult, error)
}

type BudgetRaiseResult struct {
    NewMonthlyBudget  int
    AppliedFrom       time.Time  // 翌月1日 00:00 JST
}
```

---

## 3. Adapters

### 3.1 DeliveryAdapter

```go
package delivery

type DeliveryAdapter interface {
    // PlaceOrder は外部デリバリーサービスに注文を送る
    // 本 MVP 実装 (MockDeliveryAdapter) は固定応答を返す
    PlaceOrder(ctx context.Context, req PlaceOrderInput) (*PlaceOrderOutput, error)
}

type PlaceOrderInput struct {
    UserID     string
    Category   string
    StoreName  string
    MenuName   string
    Amount     int
}

type PlaceOrderOutput struct {
    ExternalOrderID  string
    Status           string  // "accepted" 等
    ETAMinutes       int     // 配達時間見込み (モックでは固定 30)
}

// 実装例
type MockDeliveryAdapter struct{}

func (a *MockDeliveryAdapter) PlaceOrder(ctx context.Context, req PlaceOrderInput) (*PlaceOrderOutput, error) {
    return &PlaceOrderOutput{
        ExternalOrderID: "mock-" + ulid.Make().String(),
        Status:          "accepted",
        ETAMinutes:      30,
    }, nil
}
```

### 3.2 BedrockAdapter

```go
package bedrock_adapter

type BedrockAdapter interface {
    // InferOrderPlan は注文内容を推論する（OrderService から使用）
    InferOrderPlan(ctx context.Context, req InferOrderPlanInput) (*InferOrderPlanOutput, error)

    // InferSuggestion は先回り提案を推論する（SuggestService から使用）
    InferSuggestion(ctx context.Context, req InferSuggestionInput) (*InferSuggestionOutput, error)
}

type InferOrderPlanInput struct {
    UserID     string
    History    []OrderHistoryBrief  // 直近履歴
    NowJST     time.Time
    Category   string
}

type InferOrderPlanOutput struct {
    StoreName  string
    MenuName   string
    Amount     int
    Rationale  string  // デバッグ・ログ用
}

type InferSuggestionInput struct {
    UserID   string
    History  []OrderHistoryBrief
    NowJST   time.Time
}

type InferSuggestionOutput struct {
    HasSuggestion  bool
    Title          string
    Plan           *SuggestionPlan
}

// 実装: ClaudeBedrockAdapter（Converse API 利用）
// リトライと fallback は Service 層で実装し、Adapter は純粋な Bedrock 呼び出しのみ
```

### 3.3 FallbackSuggestProvider

```go
package fallback

type FallbackSuggestProvider interface {
    // BuildFromHistory は履歴から最頻パターンの固定プランを返す
    BuildFromHistory(ctx context.Context, history []OrderHistoryBrief) (*SuggestionPlan, error)

    // Default は履歴すら取れないときのハードコード値を返す（CoCo壱カレー想定）
    Default() *SuggestionPlan
}
```

---

## 4. Repositories

### 4.1 WalletRepository

```go
package wallet_repo

type WalletRepository interface {
    // Get は残高を取得する
    Get(ctx context.Context, userID string) (*WalletRecord, error)

    // DeductConditional は残高が amount 以上の条件付きで減算する
    // 残高不足時は ErrInsufficientBalance を返す
    DeductConditional(ctx context.Context, userID string, amount int) (*WalletRecord, error)

    // ResetTo は残高を指定値にリセットする（月初リセット用）
    ResetTo(ctx context.Context, userID string, to int) error

    // ListAllUserIDs はリセット対象のユーザIDを列挙する
    ListAllUserIDs(ctx context.Context) ([]string, error)
}

type WalletRecord struct {
    UserID     string
    Balance    int
    UpdatedAt  time.Time
}
```

### 4.2 BudgetSettingsRepository

```go
package budget_settings

type BudgetSettingsRepository interface {
    Get(ctx context.Context, userID string) (*BudgetSettings, error)
    Set(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error
}

type BudgetSettings struct {
    UserID         string
    MonthlyBudget  int
    EffectiveFrom  time.Time   // 翌月リセットから適用
    RaiseHistory   []RaiseLog  // 増額履歴（NFR-DEG-04 のトレース用）
}

type RaiseLog struct {
    At        time.Time
    PrevBudget int
    NewBudget  int
}
```

### 4.3 OrderHistoryRepository

```go
package order_history

type OrderHistoryRepository interface {
    Insert(ctx context.Context, rec OrderRecord) error
    ListRecent(ctx context.Context, userID string, limit int) ([]OrderRecord, error)
    CountThisMonth(ctx context.Context, userID string) (int, error)
    SumThisMonth(ctx context.Context, userID string) (int, error)  // 消化額合計
}
```

### 4.4 IdempotencyRepository

```go
package idempotency

type IdempotencyRepository interface {
    // TryAcquire は idempotencyKey の初取得なら true を返す。既存なら payload も返す。
    TryAcquire(ctx context.Context, key string, payload []byte, ttl time.Duration) (acquired bool, prev []byte, err error)
}
```

### 4.5 BudgetResetLogRepository

```go
package budget_reset_log

type BudgetResetLogRepository interface {
    Insert(ctx context.Context, log BudgetResetLog) error
}

type BudgetResetLog struct {
    ResetDate    string  // "YYYY-MM"
    UserID       string
    PrevBalance  int
    NewBalance   int
    At           time.Time
}
```

---

## 5. SchedulerLambda (monthlyResetHandler)

```go
package main // scheduler

// Lambda handler（Gin/LWA 不要、単純 Lambda handler）
func HandleMonthlyReset(ctx context.Context, event events.EventBridgeEvent) error
```
- **処理**:
  1. `BudgetSettingsRepository.ListAll()` で全ユーザ列挙（実装上は `WalletRepository.ListAllUserIDs` と組み合わせ）
  2. 各ユーザに対して `WalletService.ResetAll` 相当の処理
  3. 各リセットを `BudgetResetLogRepository.Insert` に記録
- **失敗時**: 構造化ログに ERROR 出力し、運用者が手動リカバリ（設計書記載）

---

## 6. Front-end Methods / Hooks（Next.js Client Components）

Next.js / React 側のカスタムフックを定義する。

### 6.1 useAuth

```typescript
function useAuth(): {
  user: User | null;
  login(email: string, password: string): Promise<void>;
  signup(email: string, password: string): Promise<void>;
  logout(): Promise<void>;
  isAuthenticated: boolean;
}
```

### 6.2 useWallet

```typescript
function useWallet(): {
  balance: number;
  monthlyBudget: number;
  isLoading: boolean;
  refetch(): void;
}
```
- TanStack Query の `useQuery('wallet', fetchWallet)` のラッパ

### 6.3 useOrder

```typescript
function useOrder(): {
  placeOrder(input: { suggestionId?: string }): Promise<OrderResult>;
  isPlacing: boolean;
}
```
- TanStack Query の `useMutation` + 冪等性キー自動生成

### 6.4 useSuggestion

```typescript
function useSuggestion(): {
  suggestion: Suggestion | null;
  isLoading: boolean;
}
```

### 6.5 useMetrics

```typescript
function useMetrics(): {
  damageCount: number;
  consumptionRate: number;
  thresholdExceeded: boolean;
  remainingBalance: number;
}
```

### 6.6 useBudgetRaise

```typescript
function useBudgetRaise(): {
  recommendedBudget: number;
  accept(newBudget: number): Promise<void>;
  dismiss(): void;
}
```

---

## 7. エラー定義

Go 側の主要 sentinel エラー（Functional Design で追加拡張）:

```go
package errors

var (
    ErrInsufficientBalance   = errors.New("insufficient balance")
    ErrIdempotencyConflict   = errors.New("idempotency key conflict with different payload")
    ErrUnauthorized          = errors.New("unauthorized")
    ErrBedrockUnavailable    = errors.New("bedrock unavailable after retry")  // Service 層でハンドリング
    ErrValidation            = errors.New("validation failed")
)
```

API 応答のエラーコード体系（仮）:

| HTTP | Code | 意味 |
|---|---|---|
| 400 | `VALIDATION_FAILED` | リクエストパラメータ不備 |
| 401 | `UNAUTHORIZED` | JWT 不正・期限切れ |
| 402 | `INSUFFICIENT_BALANCE` | 残高不足（「ダメになれません」応答） |
| 409 | `IDEMPOTENCY_CONFLICT` | 同一キーで異なるリクエスト |
| 500 | `INTERNAL_ERROR` | 想定外エラー |
| 503 | `SUGGESTION_FALLBACK` | Bedrock フォールバック発動（内部的には 200 で処理継続、ログにのみ記録） |

---

## 8. 審査観点へのトレーサビリティ

| 審査観点 | 本ドキュメントでの対応 |
|---|---|
| ビジネス意図の明確さ | 各 Service / Handler のコメントで対応ストーリー・Intent を明記 |
| 創造性とテーマ適合性 | `BudgetRaiseService.ComputeRecommendedBudget`, `MetricsService.GetMetrics.SummaryText` など、ダメ化UX のロジックを型付きで示す |
| Unit 分解の適切さ | パッケージ名（`wallet` / `order` / `suggest` / `metrics` / `budget_raise` 等）が Unit 候補と一致 |
| ドキュメント品質 | インターフェース・入出力型・エラー・フックを横断的に整理 |
