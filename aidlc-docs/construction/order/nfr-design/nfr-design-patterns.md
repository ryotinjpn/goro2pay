# Unit C (`order`) — NFR Design Patterns

**Document Version**: 1.0
**Created**: 2026-05-23
**Stage**: Construction / NFR Design
**Unit**: C — `order` (代行手配コア / MVP の心臓)
**Depth**: Comprehensive
**Related**:
- Plan: [order-nfr-design-plan.md](../../plans/order-nfr-design-plan.md)
- NFR Requirements: [nfr-requirements.md](../nfr-requirements/nfr-requirements.md), [tech-stack-decisions.md](../nfr-requirements/tech-stack-decisions.md)
- Functional Design: [business-logic-model.md](../functional-design/business-logic-model.md), [business-rules.md](../functional-design/business-rules.md), [domain-entities.md](../functional-design/domain-entities.md), [frontend-components.md](../functional-design/frontend-components.md)
- 凍結契約: [unit-interfaces.md](../../interfaces/unit-interfaces.md)

本ドキュメントは Unit C `order` の **設計パターン集** を、NFR Requirements (NFRC-C01〜C25) を実現するための具体的な実装方針として網羅する。Comprehensive 深度として 14 パターン（`P-RETRY-01` / `P-PLAN-01` / `P-OBS-01〜03` / `P-INIT-01` / `P-DI-01` / `P-MOCK-01` / `P-PBT-01` / `P-FE-ERR-01` / `P-FE-TOAST-01〜02` / `P-FE-LOAD-01` / `P-FE-LOCK-01`）を ID 付きで管理する。

---

## 1. パターン識別子規則

- 形式: `P-{カテゴリ}-{番号}`
- カテゴリ:
  - `RETRY`: リトライ・エラー分類
  - `PLAN`: Bedrock + フォールバックの統合
  - `OBS`: 観測性（レイテンシ・ログ）
  - `INIT`: コールドスタート最適化
  - `DI`: 依存性注入
  - `MOCK`: テストスタブ
  - `PBT`: Property-Based Testing
  - `FE-ERR`: Frontend エラーハンドリング
  - `FE-TOAST`: Frontend トースト表示
  - `FE-LOAD`: Frontend ローディング状態
  - `FE-LOCK`: Frontend 連打抑制

---

## 2. Backend パターン

### P-RETRY-01: Bedrock Retry Classification

**由来**: Plan Q-D1 = D（A + C ハイブリッド）、NFRC-C06

**目的**: Bedrock 呼出失敗時の「リトライすべきか / 即フォールバックか」の判定を、SDK 型を活用しつつ interface 抽象化で実現。

**設計**:
- `RetryClassifier` interface を切り出し、`BedrockRetryClassifier` implementation に SDK 型判定を集約
- `OrderService` / `PlanBuilder` は interface 経由で利用 → テスト容易性 + Unit D 共有性

**実装イメージ**:
```go
// internal/adapters/bedrock/retry.go
type RetryClassifier interface {
    ShouldRetry(err error) bool
}

type BedrockRetryClassifier struct{}

func (c *BedrockRetryClassifier) ShouldRetry(err error) bool {
    // リトライ対象
    var throttle *types.ThrottlingException
    var unavailable *types.ServiceUnavailableException
    if errors.As(err, &throttle) || errors.As(err, &unavailable) {
        return true
    }
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }
    // Smithy ネットワークエラー
    var apiErr *smithy.GenericAPIError
    if errors.As(err, &apiErr) && apiErr.ErrorFault() == smithy.FaultServer {
        return true
    }
    // 永続エラー（即フォールバック）
    return false
}
```

**リトライ対象 / 非対象**:

| エラー | 判定 |
|---|---|
| `*types.ThrottlingException` | リトライ |
| `*types.ServiceUnavailableException` | リトライ |
| `context.DeadlineExceeded` | リトライ |
| `*smithy.GenericAPIError` (FaultServer) | リトライ |
| `*types.ValidationException` | 即フォールバック |
| `*types.AccessDeniedException` | 即フォールバック |
| `*types.ResourceNotFoundException` | 即フォールバック |

**Unit D との共有可能性**: 横串 `internal/adapters/bedrock/retry.go` に配置、Unit D `SuggestService` でも再利用可能。

**テスト戦略**: `fakeClassifier` を使った単体テスト + Q-D7=C の PBT で間接的に検証。

---

### P-PLAN-01: Plan Construction Strategy

**由来**: Plan Q-D2 = B、NFRC-C06 / NFRC-C08

**目的**: Bedrock 呼出 + リトライ + フォールバック分岐の一連のロジックを `PlanBuilder` に集約し、`OrderService` の認知負荷を下げる。

**設計**:
- `PlanBuilder` interface を新設、`Build(ctx, history)` で `Plan` を返す
- `BedrockPlanBuilder` implementation は `BedrockAdapter` + `RetryClassifier` (P-RETRY-01) + `FallbackSuggestProvider` を依存に持つ
- `OrderService` は冪等性 / Wallet / Adapter / History のオーケストレーションのみ担当

**実装イメージ**:
```go
// internal/order/plan_builder.go
type PlanBuilder interface {
    Build(ctx context.Context, history []OrderRecord) (*Plan, error)
}

type Plan struct {
    StoreName        string
    MenuName         string
    Amount           int
    Category         string
    Source           string // "bedrock" | "fallback_history" | "fallback_default"
    BedrockLatencyMs int64
    BedrockAttempt   int
    FallbackTriggered bool
}

type BedrockPlanBuilder struct {
    bedrock    BedrockAdapter
    fallback   FallbackSuggestProvider
    classifier RetryClassifier
}

func (b *BedrockPlanBuilder) Build(ctx context.Context, history []OrderRecord) (*Plan, error) {
    // 1. Bedrock 呼出 + リトライ判定（P-RETRY-01）
    // 2. 失敗時はフォールバック（履歴件数で BuildFromHistory or Default 分岐）
    // 3. Plan を返す（Source / メトリクス情報付き）
}
```

**分岐ロジック**:
- Bedrock 呼出（NFRC-C06/C07: リトライ 1 回 / タイムアウト 1.5s）
- Bedrock 失敗 + 履歴 ≥ 5 件 → `FallbackSuggestProvider.BuildFromHistory(history)`
- Bedrock 失敗 + 履歴 < 5 件 → `FallbackSuggestProvider.Default()`（5 店舗ランダム選択、BR-C07）

**配置**: `internal/order/plan_builder.go`（Unit C 内、Unit D 共有時は横串移動可能な interface 設計）

**テスト戦略**:
- `PlanBuilder` 単体テスト: 4 シナリオ（成功 / リトライ後成功 / フォールバック履歴 / フォールバック Default）
- `OrderService` 統合テスト: `PlanBuilder` を mock 化、オーケストレーションのみ検証

---

### P-OBS-01: Latency Measurement

**由来**: Plan Q-D3 = C、NFRC-C01 / NFRC-C12

**目的**: E2E と各ステップのレイテンシを計測し、CloudWatch Logs メトリクスフィルタ（NFRC-C13）で集計可能にする。

**設計（2 層）**:

**1. E2E レイテンシ: `LatencyMiddleware`（Gin handler chain の最初）**

```go
// internal/middleware/latency.go
func LatencyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        slog.InfoContext(c.Request.Context(), "request_complete",
            "latencyMs", time.Since(start).Milliseconds(),
            "statusCode", c.Writer.Status(),
            "path", c.FullPath(),
        )
    }
}
```

→ 全 API リクエストに自動付与、書き忘れ不可能。

**2. 各ステップレイテンシ: `measure(ctx, name, fn)` ヘルパ**

```go
// internal/observability/measure.go
func Measure[T any](ctx context.Context, name string, fn func() (T, error)) (T, time.Duration, error) {
    start := time.Now()
    result, err := fn()
    elapsed := time.Since(start)
    slog.InfoContext(ctx, name+"_complete",
        "latencyMs", elapsed.Milliseconds(),
        "success", err == nil,
    )
    return result, elapsed, err
}
```

**計測対象**: Bedrock / Wallet.Deduct / DeliveryAdapter.Place / OrderHistory.Insert / PlanBuilder.Build

**NFRC-C12 マッピング**:
- `bedrockLatencyMs` → `Measure(ctx, "bedrock", ...)` の戻り値を `LogSummary.SetBedrockLatencyMs(ms)` で蓄積（P-OBS-02 連携）
- `bedrockAttempt` → リトライカウンタを `LogSummary.SetBedrockAttempt(n)` で蓄積
- `fallbackTriggered` → PlanBuilder で判定、`LogSummary.SetFallbackTriggered(b)` で蓄積

**Unit A `ContextAwareSlogHandler` (LC-AUTH-05) との連携**: middleware と `Measure` の出力は `slog.InfoContext(ctx, ...)` 経由のため、`traceId` / `requestId` / `userId` が自動付与される。

**Unit B / C / D / E 全 Unit 横串化**: `internal/middleware/latency.go` / `internal/observability/measure.go` は全 Unit から利用可能。

---

### P-OBS-02: Order LogSummary

**由来**: Plan Q-D6 = B、NFRC-C12 / NFRC-C24

**目的**: NFRC-C12 の Unit C 固有 11 項目を `PlaceOrder` リクエスト中に蓄積し、完了時に 1 件のサマリログ（`place_order_complete` event）として出力。Bedrock 本文を構造的に排除（NFRC-C24）。

**設計**:
- `LogSummary` struct が 11 項目を保持
- 各項目の setter（`SetIdempotencyKey` / `SetBedrockLatencyMs` 等）
- `defer summary.LogComplete()` パターンで完了時一括出力（成功・エラー両方で確実に出力）

**実装イメージ**:
```go
// internal/order/logger.go
type LogSummary struct {
    ctx               context.Context
    startedAt         time.Time
    idempotencyKey    string
    idempotent        bool
    orderId           string
    bedrockLatencyMs  int64
    bedrockAttempt    int
    fallbackTriggered bool
    category          string
    amount            int
    storeName         string
    menuName          string
    historyCount      int
    source            string
}

func NewLogSummary(ctx context.Context) *LogSummary {
    return &LogSummary{ctx: ctx, startedAt: time.Now()}
}

func (s *LogSummary) SetIdempotencyKey(key string) { s.idempotencyKey = key }
// ...各項目の setter
func (s *LogSummary) LogComplete() {
    slog.InfoContext(s.ctx, "place_order_complete",
        "idempotencyKey", s.idempotencyKey,
        "idempotent", s.idempotent,
        "orderId", s.orderId,
        "bedrockLatencyMs", s.bedrockLatencyMs,
        "bedrockAttempt", s.bedrockAttempt,
        "fallbackTriggered", s.fallbackTriggered,
        "category", s.category,
        "amount", s.amount,
        "storeName", s.storeName,
        "menuName", s.menuName,
        "historyCount", s.historyCount,
        "source", s.source,
    )
}
```

**使用パターン**:
```go
func (s *OrderService) PlaceOrder(ctx, req) (*PlaceOrderResult, error) {
    summary := order.NewLogSummary(ctx)
    defer summary.LogComplete()
    summary.SetIdempotencyKey(req.IdempotencyKey)
    summary.SetSource(req.Source())
    // ... 各ステップで蓄積
    return result, nil
}
```

**Unit A `ContextAwareSlogHandler` (LC-AUTH-05) との責務分離**:
- 共通 8 項目: `ContextAwareSlogHandler` が context から自動付与（`level / timestamp / userId / action / traceId / requestId / email_hash / userAgent`）
- Unit C 固有 11 項目: `LogSummary` がメモリで蓄積、`LogComplete` で出力

**PII 防御**: `LogSummary` struct に Bedrock プロンプト本文・レスポンス本文のフィールドを意図的に含めないことで、構造的に漏洩リスクを排除（NFRC-C24 整合、タイプセーフな防御）。

---

### P-OBS-03: Layered Logging Strategy

**由来**: Plan Q-D13 = C、NFRC-C13

**目的**: NFRC-C13 アラーム 3 種を最小ログ件数で検知可能にする 3 層ログ戦略。

**3 層構造**:

| 層 | レベル | 出力タイミング | 用途 |
|---|---|---|---|
| サマリ | INFO | `defer summary.LogComplete()` で 1 リクエスト終了時 | NFRC-C12 19 項目、最終集計、p95 検知（NFRC-C13-1） |
| イベント WARNING | WARN | リトライ発動時 / フォールバック発動時 | NFRC-C13-2 / NFRC-C13-3 のシグナル |
| エラー | ERROR | エラー戻り直前 | 異常終了の検知（CloudWatch Errors メトリクス） |

**WARNING ログのスキーマ**:
```go
// Bedrock リトライ発動時
slog.WarnContext(ctx, "bedrock_retry",
    "event", "bedrock_retry",
    "attempt", 2,
    "errorClass", "ThrottlingException", // RetryClassifier 経由で分類
    "elapsedMs", 1500,
)

// フォールバック発動時
slog.WarnContext(ctx, "fallback_triggered",
    "event", "fallback_triggered",
    "reason", "bedrock_double_failure", // or "permanent_error"
    "historyCount", 5,
    "fallbackType", "build_from_history", // or "default"
)
```

**CloudWatch Logs Insights クエリ例**（Infrastructure Design 引き継ぎ）:

```
# NFRC-C13-2 リトライ発動率
fields @timestamp
| filter level = "WARN" and event = "bedrock_retry"
| stats count() by bin(5m)

# NFRC-C13-3 フォールバック発動率
fields @timestamp, fallbackType
| filter level = "WARN" and event = "fallback_triggered"
| stats count() by bin(5m)

# NFRC-C13-1 p95 違反検知
fields @timestamp, latencyMs
| filter event = "place_order_complete"
| stats percentile(latencyMs, 95) by bin(5m)
```

**P-OBS-02 (LogSummary) との責務分離**:
- `LogSummary`: 1 リクエストの状態を蓄積、完了時 INFO 出力
- `EventLogger`: イベント駆動の WARNING 出力（リトライ・フォールバック）
- 標準 `slog`: ERROR 出力

**Unit B との差分根拠**: Unit B は Bedrock 呼出なし、リトライ発動シグナル不要 → A 相当で十分。Unit C は Bedrock 呼出あり、Comprehensive 深度 → C で運用シグナル強化。

---

### P-INIT-01: Lambda Cold Start Optimization

**由来**: Plan Q-D4 = A、NFRC-C18

**目的**: コールドスタート目標 P95 600ms を達成するため、SDK 初期化を Lambda INIT フェーズで実施し burst CPU を活用。

**設計**:
- **Bedrock SDK 初期化**: `internal/adapters/bedrock/adapter.go` の `init()` で `bedrockruntime.Client` を package-level 変数に保持
- **DynamoDB SDK 初期化**: `internal/repo/order_history/repository.go` の `init()` で `dynamodb.Client` を package-level 変数に保持
- **AWS Config**: `init()` 内で `config.LoadDefaultConfig(context.Background())` を呼出。`AWS_REGION` 環境変数（Lambda 標準）を使用
- **Fail-fast**: `init()` で SDK 初期化失敗時は `panic`（Lambda 起動失敗 → CloudWatch 即検知）

**実装イメージ**:
```go
// internal/adapters/bedrock/adapter.go
package bedrock

var defaultClient *bedrockruntime.Client

func init() {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        panic(fmt.Sprintf("failed to load AWS config: %v", err))
    }
    defaultClient = bedrockruntime.NewFromConfig(cfg)
}

// Production 用
func NewClaudeBedrockAdapter() *ClaudeBedrockAdapter {
    return &ClaudeBedrockAdapter{
        client:     defaultClient,
        classifier: &BedrockRetryClassifier{},
    }
}

// テスト用 (P-MOCK-01 連携)
func NewClaudeBedrockAdapterWithClient(c BedrockRuntimeAPI) *ClaudeBedrockAdapter {
    return &ClaudeBedrockAdapter{client: c, classifier: &BedrockRetryClassifier{}}
}
```

**コールドスタート目標達成の根拠**:
- INIT フェーズで AWS 提供の burst CPU を活用 → SDK 初期化を 250〜400ms で完了
- INVOKE フェーズは SDK 再利用 → 600ms 以内（NFRC-C18 達成）

**Unit B / Unit D との整合**: 横串 `internal/adapters/bedrock/` は Unit D との共有のため、init パターンを Unit D 側でも統一。

---

### P-DI-01: Manual Dependency Injection

**由来**: Plan Q-D5 = A、Unit A / B 統一

**目的**: コンポーネント数 < 10 のデモ規模に最適、Unit A / B / D / E と完全統一の DI パターン。

**設計**:
- 配線箇所: `apps/api/main.go` で全 component を組み立て
- DI 順序:
  1. P-INIT-01 の `init()` で SDK client を package-level 変数に設定（Bedrock / DynamoDB）
  2. `main.go` で Adapter / Provider を生成（`NewClaudeBedrockAdapter()` 等、内部で `defaultClient` 使用）
  3. PlanBuilder / Service を組み立て（Adapter 注入）
  4. Handler を組み立て（Service 注入）
  5. Gin Router に Handler 登録

**実装イメージ**（`main.go` Unit C 部分）:
```go
// apps/api/main.go
bedrockAdapter := bedrock.NewClaudeBedrockAdapter()
deliveryAdapter := delivery.NewMockDeliveryAdapter()
fallbackProvider := fallback.NewFallbackSuggestProvider()
planBuilder := order.NewBedrockPlanBuilder(bedrockAdapter, fallbackProvider)
historyRepo := orderhistory.NewRepository()
orderSvc := order.NewOrderService(historyRepo, planBuilder, deliveryAdapter, walletService)
orderHandler := order.NewOrderHandler(orderSvc)
r.POST("/api/orders", orderHandler.PlaceOrder)
r.GET("/api/orders", orderHandler.GetHistory)
```

**interface 定義配置**:
- `internal/adapters/bedrock/adapter.go`: `BedrockAdapter` interface + `ClaudeBedrockAdapter` implementation
- `internal/adapters/delivery/adapter.go`: `DeliveryAdapter` interface + `MockDeliveryAdapter` implementation
- `internal/adapters/fallback/provider.go`: `FallbackSuggestProvider` interface + implementation
- `internal/order/plan_builder.go`: `PlanBuilder` interface + `BedrockPlanBuilder` implementation (P-PLAN-01)
- `internal/order/service.go`: `OrderService` interface + struct implementation
- `internal/order/handler.go`: `OrderHandler` struct
- `internal/repo/order_history/repository.go`: `OrderHistoryRepository` interface + implementation

**`wire` 不採用の根拠**: コンポーネント数 < 10 でメリット薄、AI-DLC ワークフローで `wire_gen.go` の生成・管理が複雑化。

---

### P-MOCK-01: Function-Field Mock

**由来**: Plan Q-D8 = C、NFRC-C16

**目的**: NFRC-C16 の Bedrock スタブ（4 シナリオ）を Go 慣用句的な closure 注入パターンで実装。

**設計**:
- 単一 mock struct + 関数フィールド（closure 注入）
- テスト関数内で `mock.XxxFunc = func(...) {...}` を再代入してシナリオ切替

**実装イメージ**:
```go
// internal/adapters/bedrock/mock.go
type MockBedrockAdapter struct {
    InferOrderPlanFunc func(ctx context.Context, history []OrderRecord) (*Plan, error)
    Calls              int  // 呼び出し回数記録（テスト assertion 用）
}

func (m *MockBedrockAdapter) InferOrderPlan(ctx context.Context, history []OrderRecord) (*Plan, error) {
    m.Calls++
    if m.InferOrderPlanFunc != nil {
        return m.InferOrderPlanFunc(ctx, history)
    }
    return defaultSuccessPlan, nil  // Func 未設定時のフォールバック
}
```

**NFRC-C16 4 シナリオ実装例**:
```go
// 成功
mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) {
    return &Plan{StoreName: "ゴロゴロ食堂", MenuName: "おまかせ定食", Amount: 1000}, nil
}

// Throttle
mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) {
    return nil, &types.ThrottlingException{Message: aws.String("Rate exceeded")}
}

// Timeout
mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) {
    return nil, context.DeadlineExceeded
}

// 永続エラー
mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) {
    return nil, &types.ValidationException{Message: aws.String("invalid params")}
}
```

**P-PBT-01 との連携**: PBT 実行時は決定論的な closure をセット、Sample ごとに再代入することで「異なる Bedrock 挙動」をシミュレート可能。

**NFRC-C17 統合テスト 9 シナリオ**: テスト関数内で closure を切り替えて全シナリオ（成功 / リトライ成功 / フォールバック / 永続エラー / 冪等命中 / 連打 / 残高不足 / Context Timeout / Context Cancel）を検証。

**Unit D との整合**: Bedrock を共有するため `MockBedrockAdapter` を Unit D テストでも再利用（横串、Code Generation で確認）。

---

### P-PBT-01: Order Property-Based Testing

**由来**: Plan Q-D7 = C、NFRC-C15 / NFRC-C16

**目的**: Comprehensive 深度の品質投資として、`gopter` PBT で P-1（残高不変）/ P-3（冪等レスポンス一貫性）を検証。状態遷移の整合性まで含めて検証するため、Repository / Wallet を inmemory 実装。

**設計**: ハイブリッド mock + inmemory
- mock 化対象: BedrockAdapter / DeliveryAdapter / FallbackSuggestProvider / PlanBuilder
- inmemory 実装対象: WalletService（テスト用 stub） / OrderHistoryRepository

**配置**:
- `internal/order/service_pbt_test.go`: PBT テストコード
- `internal/order/wallet_stub_test.go`: inmemory `WalletStub`（Unit B WalletService interface 実装）
- `internal/repo/order_history/inmemory.go`: inmemory `OrderHistoryRepository` 実装

**Generator 設計**:
```go
ulidGen := gen.RegexMatch("^[0-9A-HJKMNP-TV-Z]{26}$")  // ULID 形式
userIdGen := gen.AlphaString  // Cognito sub 形式
categoryGen := gen.OneConstOf("food")  // MVP 範囲、BR-C27
historyCountGen := gen.IntRange(0, 30)  // フォールバック分岐閾値検証範囲
```

**Property 設計**:

```go
// P-1: 残高不変
properties.Property("P-1: same idempotencyKey deducts at most once",
    prop.ForAll(
        func(key string, n int, initialBalance int) bool {
            wallet := NewWalletStub(initialBalance)
            history := NewInmemoryHistory()
            svc := setupOrderService(wallet, history)
            for i := 0; i < n; i++ {
                svc.PlaceOrder(ctx, userID, PlaceOrderRequest{IdempotencyKey: key, ...})
            }
            balanceDiff := initialBalance - wallet.Balance(userID)
            return balanceDiff <= maxAmount  // 1 回分以下
        },
        ulidGen, gen.IntRange(1, 10), gen.IntRange(1000, 100000),
    ))

// P-3: 冪等性レスポンス一貫性
properties.Property("P-3: idempotent responses are identical",
    prop.ForAll(
        func(key string, n int) bool {
            svc := setupOrderService(...)
            firstResult, _ := svc.PlaceOrder(ctx, userID, req)
            for i := 1; i < n; i++ {
                result, _ := svc.PlaceOrder(ctx, userID, req)
                if result.OrderID != firstResult.OrderID { return false }
                if result.Amount != firstResult.Amount { return false }
                if result.StoreName != firstResult.StoreName { return false }
                if result.MenuName != firstResult.MenuName { return false }
            }
            return true
        },
        ulidGen, gen.IntRange(2, 10),
    ))
```

**実装/PBT 互換性の保証**:
- inmemory 実装は production と同じ interface を実装
- production の DynamoDB-specific な race condition は別途 NFRC-C17 統合テストで検証

**Unit B との連携**:
- Unit B PBT で `WalletService.Deduct` の冪等性が保証
- Unit C PBT は WalletStub で「Unit C 視点の N 回送信挙動」を検証
- → Unit 境界の契約（unit-interfaces.md §3.1）が双方向に保証

**サンプル数 / shrinking 設定**: デフォルト 100 サンプル、CI 実行時間 100ms 以内（合計 PBT スイートで 1 秒以内）。

**P-1 / P-3 適用外の根拠**:
- P-2（履歴整合）: Unit 境界をまたぐ統合検証のため E2E テスト or 観測（NFRC-C13）で対応
- フォールバック分岐閾値（履歴 4/5 件）: 境界値テーブルドリブンテストで対応（PBT 不採用、shrinking のメリットが薄い）

---

## 3. Frontend パターン

### P-FE-ERR-01: Order Error Mapping

**由来**: Plan Q-D9 = B、NFRC-C22

**目的**: NFRC-C22 のエラー UX（402 遷移 / 500 自虐トースト / 連打抑制）を純関数化、テスト容易性・再利用性を担保。

**設計**:
- `mapOrderError(err) → OrderErrorAction` 純関数
- `useOrder` の `onError` で `mapOrderError` を呼び、戻り値の `action.type` で `router.push` / `showToast` / 何もしない を分岐実行

**配置**:
- `web/lib/errorMappers.ts`: 純関数（テスト容易）
- `web/lib/toasts.ts`: `getRandomToast()` / `TOAST_VARIANTS`（P-FE-TOAST-01）
- `web/hooks/useOrder.ts`: mapper 結果をアクション実行

**`OrderErrorAction` 型**:
```typescript
type OrderErrorAction =
  | { type: 'navigate'; path: string; transitionMs?: number }
  | { type: 'toast'; text: string; durationMs?: number }
  | { type: 'silent' };
```

**エラーマッピング**:

| エラー | Action |
|---|---|
| `ApiError(402)` | `{ type: 'navigate', path: '/budget-empty', transitionMs: 0 }` |
| `ApiError(409)` | `{ type: 'silent' }`（冪等性衝突は実質成功扱い、BR-C39） |
| `ApiError(500)` | `{ type: 'toast', text: getRandomToast(), durationMs: 5000 }` |
| `NetworkError` | `{ type: 'toast', text: getRandomToast(), durationMs: 5000 }` |
| その他 | `{ type: 'toast', text: getRandomToast(), durationMs: 5000 }` |

**`useOrder` 実装パターン**:
```typescript
export function useOrder() {
  const router = useRouter();
  const { showToast } = useToast();
  const { isLocked, triggerLock } = useDisableLock(1000);  // P-FE-LOCK-01

  const mutation = useMutation({
    mutationFn: placeOrder,
    retry: 0,  // NFRC-C19, FD Q-12=A
    onError: (err) => {
      triggerLock();
      const action = mapOrderError(err);
      switch (action.type) {
        case 'navigate':
          startTransition(() => router.push(action.path));
          break;
        case 'toast':
          showToast(action.text, action.durationMs);
          break;
        case 'silent':
          break;
      }
    },
    onSuccess: () => {
      triggerLock();
      queryClient.invalidateQueries({ queryKey: ['orderHistory'] });
      queryClient.invalidateQueries({ queryKey: ['balance'] });
    },
  });

  return { ...mutation, disabled: mutation.isPending || isLocked };
}
```

**テスト戦略**:
- `errorMappers.test.ts`: 純関数として全エラーケース検証（Vitest）
- `useOrder.test.tsx`: React Testing Library + msw で mock、`onError` callback の起動を検証
- `BudgetEmptyScreen` への遷移は E2E テスト（dev 環境、Build and Test ステージで詳細化）

**NFR-OBS-02 整合**: Sentry / Datadog 等の外部観測ツール連携は含めない、placeholder も残さない。

**Unit A `apiClient` (LC-AUTH-09) との連携**: `apiClient` の `ApiError` 型を `mapOrderError` で受ける、HTTP ステータスコードベースの分岐（`err.status` フィールド）。

---

### P-FE-TOAST-01: Random Toast Variant

**由来**: Plan Q-D10 = A、NFRC-C22 / NFR-DEG-05

**目的**: 自虐トースト 3 種ローテーションを純関数で実装、NFR-DEG-05 の「予測不能性によるダメ化エンタメ」を体現。

**設計**:
- `lib/toasts.ts` に readonly 配列定義
- `getRandomToast()` 純関数（`Math.floor(Math.random() * TOAST_VARIANTS.length)`）

**実装イメージ**:
```typescript
// web/lib/toasts.ts
export const TOAST_VARIANTS: readonly string[] = [
  'サーバーがやる気を失いました…もう一度お試しください',
  'システムがふぬけてます。少し待ってあげてください',
  '今日はちょっとダメ化に失敗しました。再挑戦しますか？',
] as const;

export type ToastVariant = typeof TOAST_VARIANTS[number];

export function getRandomToast(): ToastVariant {
  const index = Math.floor(Math.random() * TOAST_VARIANTS.length);
  return TOAST_VARIANTS[index];
}
```

**テスト戦略**:
- `vi.spyOn(Math, 'random').mockReturnValue(value)` で決定論的にテスト
- 全 3 文言が選択可能であることを検証
- 配列が空でないこと / 文言が空文字でないことのバリデーション

**P-FE-ERR-01 との連携**: `mapOrderError` が `getRandomToast()` を呼んで `OrderErrorAction.text` に格納、純関数同士の組み合わせで責務分離。

**最終文言の確定**: 例示 3 文言は NFR Design 段階の参考、最終文言は Code Generation で確定（NFR-DEG-05 文言ガイドラインに従う）。

**連続避けロジック不採用の根拠**: NFRC-C22 の連打抑制 1 秒（P-FE-LOCK-01）で連続表示はほぼ起きない、ランダム性そのものが NFR-DEG-05 のダメ化エンタメ体現。

---

### P-FE-TOAST-02: Toast Host & Queue

**由来**: Plan Q-D14 = A、NFRC-C22

**目的**: トースト表示を Jotai + React で自前実装、外部依存なし、最大 3 件キューで多重表示制御。

**設計**:
- `<ToastHost />` を `app/layout.tsx` の `<body>` 直下に配置
- Jotai `toastsAtom` でグローバル状態管理
- `useToast.showToast(text, durationMs)` API
- 最大 3 件表示、4 件目以降はキューイング

**配置**:
- `web/components/ToastHost.tsx`: トースト表示コンポーネント
- `web/components/Toast.tsx`: 個別トースト（アニメーション含む）
- `web/hooks/useToast.ts`: `showToast` API（Jotai atom ベース）
- `web/state/toastAtoms.ts`: `toastsAtom`（Jotai atom 定義）

**実装イメージ**:
```typescript
// web/state/toastAtoms.ts
import { atom } from 'jotai';

export type ToastItem = {
  id: string;
  text: string;
  durationMs: number;
};

export const toastsAtom = atom<ToastItem[]>([]);

// web/hooks/useToast.ts
import { useCallback } from 'react';
import { useAtom } from 'jotai';
import { toastsAtom } from '@/state/toastAtoms';

export function useToast() {
  const [toasts, setToasts] = useAtom(toastsAtom);
  const showToast = useCallback((text: string, durationMs = 5000) => {
    const id = crypto.randomUUID();
    setToasts((prev) => [...prev, { id, text, durationMs }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, durationMs);
  }, [setToasts]);
  return { toasts, showToast };
}

// web/components/ToastHost.tsx
'use client';

import { useToast } from '@/hooks/useToast';
import { Toast } from './Toast';

const MAX_VISIBLE = 3;

export function ToastHost() {
  const { toasts } = useToast();
  const visible = toasts.slice(0, MAX_VISIBLE);
  return (
    <div role="region" aria-live="polite" className="toast-container">
      {visible.map((t) => (
        <Toast key={t.id} {...t} />
      ))}
    </div>
  );
}

// web/app/layout.tsx
import { ToastHost } from '@/components/ToastHost';

export default function RootLayout({ children }) {
  return (
    <html lang="ja">
      <body>
        <Providers>{children}</Providers>
        <ToastHost />
      </body>
    </html>
  );
}
```

**キュー制御**:
- `toastsAtom` に追加された順に `ToastHost` が `slice(0, 3)` で先頭 3 件のみ表示
- 4 件目以降は配列に保持、先頭が dismiss されると自動的に次が表示
- 各トーストは個別 `setTimeout` で独立に消去

**a11y**:
- `role="region"` + `aria-live="polite"`（コンテナ）
- `role="status"`（個別トースト）
- 自動消去後もスクリーンリーダーが読み上げ済み

**最大 3 件キューの根拠**: 表示時間 5 秒 × 1 秒連打抑制 = 5 秒間に最大 3 回エラー発生想定。それ以上はキューイング。

**Unit A 既存 ToastHost との整合**: Unit A Code Generation 時に `ToastHost` 等が既に実装されている可能性あり、その場合は Unit C で再利用（横串化）。なければ Unit C で新規実装。

---

### P-FE-LOAD-01: Order History Loading State

**由来**: Plan Q-D11 = B、NFRC-C19 / NFR-DEG-03

**目的**: NFR-DEG-03（履歴常時可視化）と NFRC-C19（60 秒 + invalidate）を `placeholderData` で両立。

**設計**: TanStack Query v5 の `keepPreviousData` を使い、再 fetch 中も前回データを薄く表示。

**実装イメージ**:
```typescript
// web/hooks/useOrderHistory.ts
import { keepPreviousData, useQuery } from '@tanstack/react-query';

export function useOrderHistory() {
  return useQuery({
    queryKey: ['orderHistory'],
    queryFn: fetchOrderHistory,
    placeholderData: keepPreviousData,  // ★ 核
    staleTime: 60_000,                   // NFRC-C19
    gcTime: 300_000,                     // NFRC-C19
    refetchOnWindowFocus: true,          // NFRC-C19
  });
}
```

**状態遷移**:

| 状態 | `isLoading` | `isFetching` | `data` | UI |
|---|---|---|---|---|
| 初回ロード中 | true | true | undefined | スケルトン |
| データ取得済 | false | false | OrderRecord[] | 通常表示 |
| 再 fetch 中 | false | true | OrderRecord[] (前回値) | 薄く表示（opacity: 0.5） |
| エラー時 | false | false | OrderRecord[] (前回値) | 通常表示 + エラーバナー重畳 |
| 0 件 | false | false | [] | 「履歴がまだありません」ダメ化文言 |

**空状態（0 件）の扱い**:
- 「履歴がまだありません」ダメ化文言（NFR-DEG-05 体現、最終文言は Code Generation で確定）
- サジェストカードは出さない（FD `business-logic-model.md` §6 整合、フォールバック分岐閾値 5 件未満）
- 「ご飯めんどくさい」ボタンは常に表示

**Unit B Q-D3=A との差分根拠**:
- Unit B `BalanceDisplay`: 数値 1 つ → スケルトン切替が自然
- Unit C `OrderHistoryList`: リスト → `placeholderData` で常時可視化が UX 良
- データ特性に応じた適材適所（NFRC-C19 stale time の判断と同じ思想）

**テスト戦略**: Vitest + React Testing Library で各状態（isLoading / isFetching / isError / 空配列）を検証、`placeholderData: keepPreviousData` の挙動は msw で 2 回目以降の fetch を遅延させて検証。

---

### P-FE-LOCK-01: Disable Lock Hook

**由来**: Plan Q-D12 = B、NFRC-C22

**目的**: ボタン連打抑制 1 秒を独立 hook で実装、テスト容易性・再利用性・a11y 対応を担保。

**設計**:
- `useDisableLock(durationMs)` hook
- `lockedUntil: number`（lock 解除時刻 epoch ms）を `useState` で保持
- `triggerLock()` で `lockedUntil = Date.now() + durationMs` を設定
- `useEffect` cleanup で `clearTimeout`、unmount 時にタイマー破棄

**実装イメージ**:
```typescript
// web/hooks/useDisableLock.ts
import { useCallback, useEffect, useState } from 'react';

export function useDisableLock(durationMs: number): {
  isLocked: boolean;
  triggerLock: () => void;
} {
  const [lockedUntil, setLockedUntil] = useState<number>(0);
  const isLocked = Date.now() < lockedUntil;

  const triggerLock = useCallback(() => {
    setLockedUntil(Date.now() + durationMs);
  }, [durationMs]);

  useEffect(() => {
    if (!isLocked) return;
    const remaining = lockedUntil - Date.now();
    const timer = setTimeout(() => setLockedUntil(0), remaining);
    return () => clearTimeout(timer);
  }, [lockedUntil, isLocked]);

  return { isLocked, triggerLock };
}
```

**`useOrder` での合成**: P-FE-ERR-01 の `useOrder` 実装内で `triggerLock()` を `onSuccess` / `onError` で呼出、`disabled = mutation.isPending || isLocked` を返す。

**`GoroButton` での使用**:
```tsx
function GoroButton() {
  const { mutate, disabled } = useOrder();
  return (
    <button disabled={disabled} onClick={() => mutate(...)}>
      ご飯めんどくさい
    </button>
  );
}
```

**テスト戦略**:
- `useDisableLock.test.tsx`: `vi.useFakeTimers()` + `renderHook` + `act` で時間進行を制御、各境界（999ms / 1000ms / 1001ms）を検証
- `useOrder.test.tsx`: `useDisableLock` を mock 化、`triggerLock` 呼出を検証
- `GoroButton.test.tsx`: `disabled` 属性の変化を検証

**a11y 観点**:
- HTML `disabled` 属性使用 → スクリーンリーダー「無効」読み上げ、キーボード submit 不能
- `pointerEvents: none` 等の CSS ハック不採用

**設計一貫性**: Q-D9=B / Q-D10=A / Q-D11=B と同様「責務単位のファイル分割 + 振る舞い hook 化」を継承。

**将来の拡張余地**: 他箇所（増額ボタン、履歴削除等）でも `useDisableLock` を再利用可能、localStorage 永続化は MVP 範囲外（placeholder も残さない）。

---

## 4. パターン適用マトリクス（NFR Requirements との対応）

| NFR ID | 適用パターン |
|---|---|
| NFRC-C01 (E2E p95 3.0s / p99 5.0s) | P-OBS-01 (Latency Measurement) + P-OBS-02 (LogSummary) |
| NFRC-C03 (親 Context タイムアウト 5s) | P-PLAN-01 内で `context.WithTimeout` 適用 |
| NFRC-C05 (冪等性 TTL 24h) | P-PBT-01 P-3 で検証 |
| NFRC-C06 (Bedrock リトライ 1 回) | P-RETRY-01 + P-PLAN-01 |
| NFRC-C07 (Bedrock 1.5s タイムアウト) | P-PLAN-01 内で `context.WithTimeout(parent, 1500ms)` |
| NFRC-C08 (フォールバック閾値 5 件) | P-PLAN-01 |
| NFRC-C09 (History Insert 失敗時 200) | P-OBS-03 WARNING で記録、サマリには `historyInsertFailed=true` 付与 |
| NFRC-C10 (親 Context Cancel 即時伝播) | P-PLAN-01 / P-OBS-01 内で context 伝播 |
| NFRC-C12 (構造化ログ 19 項目) | P-OBS-02 (LogSummary 11 項目) + Unit A LC-AUTH-05 (共通 8 項目) |
| NFRC-C13 (CloudWatch アラーム 3 種) | P-OBS-01 (E2E latency) + P-OBS-03 (WARNING シグナル) |
| NFRC-C15 (PBT P-1 + P-3) | P-PBT-01 |
| NFRC-C16 (Bedrock スタブマトリクス) | P-MOCK-01 |
| NFRC-C17 (統合テスト 9 シナリオ) | P-MOCK-01 + P-DI-01 |
| NFRC-C18 (Lambda 256MB / arm64 / 600ms cold) | P-INIT-01 |
| NFRC-C19 (TanStack Query 60s + invalidate) | P-FE-LOAD-01 |
| NFRC-C20 (Bedrock Haiku 4.5 Converse) | P-INIT-01 + P-PLAN-01 |
| NFRC-C22 (エラー UX) | P-FE-ERR-01 + P-FE-TOAST-01 + P-FE-TOAST-02 + P-FE-LOCK-01 |
| NFRC-C24 (Bedrock 本文ログ非記録) | P-OBS-02 (LogSummary 構造的に PII 排除) |

---

## 5. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Infrastructure Design** | DynamoDB `OrderHistory` テーブル定義、Bedrock IAM Role / モデル ARN 制限、API Gateway ルート / Throttle / CORS、Lambda 関数 Terraform、CloudWatch Alarms / メトリクスフィルタ Terraform、AWS Budgets 設定 |
| **Code Generation** | 各パターンの Go / TypeScript 実装、Bedrock プロンプトテンプレート最終、`gopter` テストコード、`MockBedrockAdapter` 実装、自虐トースト最終文言、`useOrder` / `useOrderHistory` / `useDisableLock` / `useToast` 実装、`GoroButton` / `OrderCompletionScreen` / `ToastHost` / `Toast` / `OrderHistoryList` 実装、`errorMappers` / `getRandomToast` 純関数実装 |
| **Build and Test** | dev 環境での手動 E2E 検証手順、CI ワークフロー定義、Bedrock スタブのテスト戦略実装 |

---

## 6. 文書管理

- **承認**: ユーザ承認待ち（Construction フェーズの per-unit ループ承認ゲート）
- **凍結契約への影響**: なし（P-RETRY-01 / P-PLAN-01 の interface は Unit C 内部、P-DI-01 / P-MOCK-01 はテスト範囲）
- **次ステージ**: Infrastructure Design（per-unit、Comprehensive 深度）
