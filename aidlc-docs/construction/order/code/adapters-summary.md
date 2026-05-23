# Unit C — Adapters Summary

**Generated**: 2026-05-24
**Stage**: Construction / Code Generation
**Unit**: C (`order`、横串 Adapter は Unit C/D 共有候補)

## 生成ファイル

| ファイル | 行数概算 | 用途 |
|---|---|---|
| `apps/api/internal/adapters/bedrock/retry.go` | ~90 | RetryClassifier interface + BedrockRetryClassifier (P-RETRY-01) |
| `apps/api/internal/adapters/bedrock/retry_test.go` | ~70 | エラー型ごとのリトライ判定 10 ケース |
| `apps/api/internal/adapters/bedrock/prompt.go` | ~80 | BuildPrompt + ParsePlanResponse (FD §2.2) |
| `apps/api/internal/adapters/bedrock/prompt_test.go` | ~75 | プロンプト構築 + JSON パース検証 |
| `apps/api/internal/adapters/bedrock/adapter.go` | ~170 | ClaudeBedrockAdapter + init() で SDK 初期化 (LC-05 / LC-14 / P-INIT-01) |
| `apps/api/internal/adapters/bedrock/adapter_test.go` | ~150 | 4 シナリオ + Context Canceled 検証 |
| `apps/api/internal/adapters/bedrock/mock.go` | ~30 | MockBedrockAdapter (function-field、P-MOCK-01) |
| `apps/api/internal/adapters/delivery/adapter.go` | ~60 | DeliveryAdapter interface + MockDeliveryAdapter (LC-06) |
| `apps/api/internal/adapters/delivery/adapter_test.go` | ~40 | Mock + Fake adapter 検証 |
| `apps/api/internal/adapters/delivery/mock.go` | ~25 | FakeDeliveryAdapter (テスト専用、closure 注入) |
| `apps/api/internal/adapters/fallback/stores.go` | ~25 | DefaultStores 5 店舗ラインナップ (BR-C07) |
| `apps/api/internal/adapters/fallback/provider.go` | ~80 | SimpleFallbackProvider (LC-09) |
| `apps/api/internal/adapters/fallback/provider_test.go` | ~70 | 履歴最頻 + Default ランダム + 決定論性検証 |
| `apps/api/internal/adapters/fallback/mock.go` | ~30 | FakeFallbackProvider (closure 注入) |

## 主要 interface

```go
// P-RETRY-01
type RetryClassifier interface {
    ShouldRetry(err error) bool
}

// LC-05
type BedrockAdapter interface {
    InferOrderPlan(ctx context.Context, history []HistoryItem, dayOfWeek string, category string) (*Plan, error)
}

// LC-06
type DeliveryAdapter interface {
    Place(ctx context.Context, req PlaceOrderRequest) error
}

// LC-09
type FallbackSuggestProvider interface {
    BuildFromHistory(history []HistoryItem) *Plan
    Default() *Plan
}
```

## P-RETRY-01 リトライ分類

| エラー | 判定 |
|---|---|
| `*types.ThrottlingException` | リトライ |
| `*types.ServiceUnavailableException` | リトライ |
| `*types.InternalServerException` | リトライ |
| `context.DeadlineExceeded` | リトライ |
| `smithy.APIError` (FaultServer) | リトライ |
| `*types.ValidationException` | 即フォールバック |
| `*types.AccessDeniedException` | 即フォールバック |
| `*types.ResourceNotFoundException` | 即フォールバック |
| `context.Canceled` | リトライしない (上位の意思決定優先) |
| 不明な error | リトライ (ネットワーク系想定) |

## P-INIT-01 (Lambda Cold Start)

```go
// internal/adapters/bedrock/adapter.go
var defaultClient *bedrockruntime.Client

func init() {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil { panic(...) }
    defaultClient = bedrockruntime.NewFromConfig(cfg)
}
```

`NewClaudeBedrockAdapter()` は `defaultClient` を使い、`NewClaudeBedrockAdapterWithClient(c)` でテスト時 mock SDK を注入可能。

## P-MOCK-01 (Function-Field Mock)

```go
type MockBedrockAdapter struct {
    InferOrderPlanFunc func(ctx, history, dayOfWeek, category) (*Plan, error)
    Calls              int
}
```

テスト関数内で `mock.InferOrderPlanFunc = func(...) { ... }` を再代入してシナリオ切替。
NFRC-C16 4 シナリオ (成功 / Throttle / Timeout / 永続エラー) を closure で網羅。

## 5 店舗ラインナップ (BR-C07)

| Index | Store | Menu | Amount |
|---|---|---|---|
| 0 | ゴロゴロ食堂 | おまかせ定食 | ¥1,000 |
| 1 | ぐうたら亭 | 手抜き丼 | ¥800 |
| 2 | ダメ屋 | やる気なしカレー | ¥1,200 |
| 3 | 怠惰キッチン | 何でもよし弁当 | ¥1,500 |
| 4 | ふぬけ食堂 | しょうがない定食 | ¥900 |

## NFR 達成根拠

| NFR | 実装 |
|---|---|
| NFRC-C06 (リトライ 1 回) | adapter.go の maxAttempts = 2 + RetryClassifier |
| NFRC-C07 (タイムアウト 1.5s) | adapter.go の `bedrockTimeoutPerCall = 1500 * time.Millisecond` |
| NFRC-C08 (フォールバック閾値 5 件) | plan_builder.go の `fallbackThreshold = 5` |
| NFRC-C16 (mock 必須) | MockBedrockAdapter / FakeDeliveryAdapter / FakeFallbackProvider |
| NFRC-C20 (Bedrock 3.5 Haiku Inference Profile) | adapter.go の `defaultInferenceProfileID` + `BEDROCK_INFERENCE_PROFILE_ID` 環境変数 |
| NFRC-C24 (Bedrock 本文ログ非記録) | prompt.go の HistoryItem に PII フィールドなし、 ParsePlanResponse のログ出力なし |

## 後続ステージへの引き継ぎ

- 本番デプロイ後、Bedrock 実呼出しで JSON パース成功率を CloudWatch で観測 (`bedrock_response_invalid` ログ件数)
- Unit D `SuggestService` で同じ BedrockAdapter / FallbackSuggestProvider を再利用予定
