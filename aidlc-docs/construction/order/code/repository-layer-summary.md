# Unit C — Repository Layer Summary

**Generated**: 2026-05-24
**Stage**: Construction / Code Generation
**Unit**: C (`order`)

## 生成ファイル

| ファイル | 行数概算 | 用途 |
|---|---|---|
| `apps/api/internal/repo/order_history/repository.go` | ~200 | DynamoDB OrderHistory CRUD + init() で SDK 初期化 (LC-04 / LC-15 / P-INIT-01) |
| `apps/api/internal/repo/order_history/repository_test.go` | ~180 | DynamoDB SDK Mock で CRUD 検証 |
| `apps/api/internal/repo/order_history/inmemory.go` | ~80 | inmemory 実装 (テスト専用 LC-20、PBT 用) |

## DynamoDB スキーマ (凍結契約 §3.2 整合)

| 属性 | 型 | 用途 |
|---|---|---|
| `PK` | S | `USER#{userID}` |
| `SK` | S | `ORDER#{orderedAt}#{orderID}` |
| `orderId` | S | ULID |
| `userId` | S | Cognito sub |
| `category` | S | "food" 固定 (MVP) |
| `storeName` | S | 例: "ゴロゴロ食堂" |
| `menuName` | S | 例: "おまかせ定食" |
| `amount` | N | 円 |
| `orderedAt` | S | RFC3339 |
| `idempotencyKey` | S | ULID (Frontend 発行) |
| `dayOfWeek` | S | 例: "Mon" |
| `source` | S | "bedrock" / "fallback_history" / "fallback_default" |
| `expiresAt` | N | Unix epoch、TTL 属性 (90 日) |

GSI なし (凍結契約整合修正で `GSI_IdempotencyKey` を撤回済み)。

## 公開メソッド

```go
type OrderHistoryRepository interface {
    Insert(ctx context.Context, record *OrderRecord) error
    GetItem(ctx context.Context, userID, orderID, orderedAt string) (*OrderRecord, error)
    Query(ctx context.Context, userID string, limit int) ([]*OrderRecord, error)
}
```

`Insert` は ConditionExpression `attribute_not_exists(PK) AND attribute_not_exists(SK)` で重複検知。
`Query` は ScanIndexForward = false で降順、limit はデフォルト 20 / 最大 100 にクランプ (FD Q-11=A)。

## P-INIT-01 (Lambda Cold Start Optimization)

```go
var defaultClient *dynamodb.Client

func init() {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        panic(...)
    }
    defaultClient = dynamodb.NewFromConfig(cfg)
}
```

Lambda INIT フェーズの burst CPU で SDK 初期化を済ませることで、INVOKE 600ms (NFRC-C18) を確保する。失敗時は panic で fail-fast し CloudWatch で即検知。

## NFR 達成根拠

| NFR | 実装 |
|---|---|
| NFRC-C02 (GetHistory P50 100ms / P95 500ms) | DynamoDB Query (PK/SK インデックス、ScanIndexForward false で降順)、TTL で古い項目自動削除 |
| NFRC-C09 (Insert 失敗時 200 応答) | OrderService.PlaceOrder で `_ = ierr` で握りつぶし、observability.Measure 経由でログ記録 |
| NFRC-C18 (Lambda init 600ms 以内) | P-INIT-01 で達成 |

## テスト戦略

- 本番経路: `repository_test.go` で DynamoDB SDK の `DynamoDBAPI` interface mock 経由で検証
- PBT 経路: `inmemory.go` の InmemoryRepository を `service_pbt_test.go` で使用
- 統合テスト: `service_test.go` 9 シナリオで InmemoryRepository を使った E2E 検証

## 後続ステージへの引き継ぎ

- 本番デプロイ後、CloudWatch Logs Insights で `history_query_complete` / `history_insert_complete` の latencyMs を観測し、NFRC-C02 達成度を評価
- DynamoDB Local を使った integration test は Build & Test ステージで追加検討
