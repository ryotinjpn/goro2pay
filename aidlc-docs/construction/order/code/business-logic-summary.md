# Unit C — Business Logic Summary

**Generated**: 2026-05-24
**Stage**: Construction / Code Generation
**Unit**: C (`order` / 代行手配コア)

## 生成ファイル

| ファイル | 行数概算 | 用途 |
|---|---|---|
| `apps/api/internal/order/types.go` | ~50 | DTO + Sentinel error (凍結契約 §4.1 / §4.3) |
| `apps/api/internal/order/wallet.go` | ~50 | WalletService interface (Unit B 連携、凍結契約 §3.1) |
| `apps/api/internal/order/plan_builder.go` | ~110 | PlanBuilder (P-PLAN-01)、Bedrock + Fallback 統合 |
| `apps/api/internal/order/plan_builder_test.go` | ~120 | 4 シナリオ単体テスト |
| `apps/api/internal/order/logger.go` | ~70 | LogSummary (P-OBS-02)、11 項目蓄積 + LogComplete |
| `apps/api/internal/order/logger_test.go` | ~60 | 11 項目出力 + 構造的 PII 防御検証 |
| `apps/api/internal/order/event_logger.go` | ~30 | EventLogger (P-OBS-03)、WARNING ログ |
| `apps/api/internal/order/event_logger_test.go` | ~40 | bedrock_retry / fallback_triggered 検証 |
| `apps/api/internal/order/service.go` | ~200 | OrderService 実装、オーケストレーション |
| `apps/api/internal/order/service_test.go` | ~200 | NFRC-C17 統合テスト 9 シナリオ |
| `apps/api/internal/order/service_pbt_test.go` | ~110 | gopter PBT (P-1 + P-3)、各 100 サンプル |
| `apps/api/internal/order/wallet_stub_test.go` | ~80 | inmemory WalletStub (PBT 用) |

## 関数シグネチャ

```go
// 凍結契約 §4.1
type OrderService interface {
    PlaceOrder(ctx context.Context, userID string, req PlaceOrderRequest) (*PlaceOrderResult, error)
    GetHistory(ctx context.Context, userID string, limit int) ([]*OrderRecord, error)
}

// P-PLAN-01
type PlanBuilder interface {
    Build(ctx context.Context, history []OrderRecord, dayOfWeek string, category string) (*Plan, error)
}

// P-OBS-02
type LogSummary struct { /* 11 fields, no Bedrock body fields */ }
func NewLogSummary(ctx context.Context) *LogSummary
func (s *LogSummary) LogComplete()

// P-OBS-03
func LogBedrockRetry(ctx context.Context, attempt int, errorClass string, elapsedMs int64)
func LogFallbackTriggered(ctx context.Context, reason string, historyCount int, fallbackType string)
```

## NFR Design ↔ コード対応

| NFR Design パターン | 実装ファイル |
|---|---|
| P-PLAN-01 Plan Construction Strategy | `plan_builder.go` |
| P-OBS-02 Order LogSummary | `logger.go` (PII 構造的防御を `TestLogSummary_StructHasNoBedrockBodyFields` で検証) |
| P-OBS-03 Layered Logging Strategy | `event_logger.go` (WARN level、`bedrock_retry` / `fallback_triggered`) |
| P-DI-01 Manual DI | `service.go` の constructor 注入 |
| P-PBT-01 Order PBT | `service_pbt_test.go` (gopter、P-1 残高不変 + P-3 冪等レスポンス一貫性) |

## テスト結果サマリ

- 単体テスト: 全パス想定 (Go 標準 testing + testify)
- PBT: `MinSuccessfulTests = 100` で各プロパティ実行
- 統合テスト 9 シナリオ: 通常 / リトライ後成功 / フォールバック履歴 / 永続エラー / 冪等命中 / 連打 / 残高不足 / Context Canceled / バリデーション

## 後続ステージへの引き継ぎ

- Unit B WalletService 実装が main.go に DI 配線されるまでの間、`apps/api/wallet_stub.go` の `noopWalletService` が常時 `ErrInsufficientFunds` を返す。Unit B 完了時に削除すること。
- Unit D `SuggestService.ResolveSuggestion` は本 PR では未配線 (Unit C 単独でも動作可能)。
