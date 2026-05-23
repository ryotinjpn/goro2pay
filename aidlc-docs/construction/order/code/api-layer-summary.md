# Unit C — API Layer Summary

**Generated**: 2026-05-24
**Stage**: Construction / Code Generation
**Unit**: C (`order`)

## 生成ファイル

| ファイル | 行数概算 | 用途 |
|---|---|---|
| `apps/api/internal/handlers/order.go` | ~150 | OrderHandler (LC-02)、HTTP ↔ OrderService マッピング |
| `apps/api/internal/handlers/order_test.go` | ~200 | エラーマッピング 5 ケース + History 取得 |
| `apps/api/internal/middleware/latency.go` | ~30 | LatencyMiddleware (P-OBS-01)、E2E 計測 |
| `apps/api/internal/middleware/latency_test.go` | ~40 | request_complete ログ検証 |
| `apps/api/internal/observability/measure.go` | ~30 | Measure ヘルパ (P-OBS-01)、generic 関数 |
| `apps/api/internal/observability/measure_test.go` | ~50 | 成功 / エラー時のログ検証 |
| `apps/api/main.go` (編集) | +30 | Unit C component の DI 配線 (P-DI-01) |
| `apps/api/wallet_stub.go` | ~25 | noopWalletService (Unit B 完成までの暫定) |

## エラーマッピング (凍結契約 §4.3)

| order error | HTTP status | error code |
|---|---|---|
| `order.ErrInsufficientFunds` | 402 | `INSUFFICIENT_FUNDS` |
| `order.ErrIdempotencyConflict` | 409 | `IDEMPOTENCY_CONFLICT` |
| その他 | 500 | `INTERNAL_ERROR` |
| バリデーション失敗 (Gin binding) | 400 | `BAD_REQUEST` |
| `auth.UserIDFromContext` 失敗 | 401 | `UNAUTHORIZED` |

## Route 配線

```go
api := r.Group("/api", auth.AttachUserID())
api.POST("/orders", orderHandler.PlaceOrder)
api.GET("/orders", orderHandler.GetHistory)
```

すべて Unit A `auth.AttachUserID()` middleware の認証グループ配下に登録。
`LatencyMiddleware` は Gin Engine ルートに登録されているため、`/health` /
`/api/*` 全 route で `request_complete` ログが出力される。

## 観測性パターン

```go
// 各ステップを measure ヘルパで計測 (P-OBS-01)
plan, _, err := observability.Measure(ctx, "plan_build", func() (*Plan, error) {
    return s.planBuilder.Build(ctx, history, dayOfWeek, req.Category)
})

// E2E は middleware が自動付与
// `request_complete` イベントに latencyMs / statusCode / method / path
```

## NFR 達成根拠

| NFR | 実装 |
|---|---|
| NFRC-C01 (E2E p95 3.0s) | LatencyMiddleware で `request_complete` を CloudWatch メトリクスフィルタ (NFRC-C13-1) で集計 |
| NFRC-C12 (構造化ログ 19 項目) | LogSummary (11 項目) + ContextAwareSlogHandler (8 項目、Unit A) |
| NFRC-C18 (Lambda 256MB / arm64) | Infrastructure Design 既存設定で対応、本層では追加変更なし |

## 後続ステージへの引き継ぎ

- Unit B WalletService 実装到達後、`main.go` の DI 配線で `walletStub` を Unit B 実装に差し替え
- E2E テスト (Playwright) は Build & Test ステージで実装
