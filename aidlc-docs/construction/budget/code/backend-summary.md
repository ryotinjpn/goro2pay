# Backend Summary — Unit B (`budget`)

## ディレクトリ構造

```
apps/api/
├── main.go                                  # ★ DI 配線追加
├── internal/wallet/                         # ★ 新規パッケージ
│   ├── doc.go                                # パッケージドキュメント
│   ├── errors.go                             # ErrInsufficientBalance / ErrIdempotencyConflict / ErrBudgetOutOfRange
│   ├── types.go                              # WalletService / WalletSnapshot / DeductResult / ResetResult
│   ├── validation.go                         # VR-B-01〜06 純関数
│   ├── validation_test.go                    # 境界値 table-driven test
│   ├── service.go                            # GetBalance / SetBudget / Deduct / ResetAll 本実装
│   ├── service_test.go                      # function-field mock で 17 シナリオ
│   ├── service_pbt_test.go                   # gopter PBT 3 プロパティ × 100 sample
│   ├── handler.go                            # GET /api/wallet, POST /api/wallet/budget
│   ├── handler_test.go                       # httptest で 6 シナリオ
│   ├── order_adapter.go                      # order.WalletService interface 実装
│   └── order_adapter_test.go                # 4 シナリオ
├── internal/repo/wallet_repo/                # ★ 新規 (Wallet + WalletReader)
├── internal/repo/budget_settings/             # ★ 新規 (Reader + Writer for Unit E)
├── internal/repo/idempotency/                # ★ 新規 (TryAcquire / SaveResponse)
├── internal/repo/budget_reset_log/            # ★ 新規 (Get / Insert with attribute_not_exists)
└── cmd/scheduler/                           # ★ 新規 月初リセット Lambda
    ├── main.go                               # lambda.Start, runReset(ctx, svc) 抽出
    ├── main_test.go                          # mock svc で 2 シナリオ
    ├── Makefile                              # bootstrap arm64 ビルド
    └── README.md
```

`apps/api/wallet_stub.go` は **削除済み** (Unit B 実装到達)。

## ドメイン層 (`internal/wallet/`)

### errors.go (Sentinel Errors / 凍結 IF §3.1)

| Sentinel | HTTP マップ |
|---|---|
| `ErrInsufficientBalance` | 402 INSUFFICIENT_BALANCE |
| `ErrIdempotencyConflict` | 409 IDEMPOTENCY_CONFLICT |
| `ErrBudgetOutOfRange` | 400 VALIDATION_FAILED |
| `ErrInvalidInput` (内部) | 400 VALIDATION_FAILED |

### validation.go

純関数 5 種で VR-B-01〜06 を実装:
- `ValidateMonthlyBudget(int)` — 範囲 [1000, 100000] + 1000 円刻み
- `ValidateUserID(string)` — 空文字列 / 空白のみ拒否
- `ValidateDeductInput(userID, amount, key)` — 正値 + ULID 形式 + プレフィックス userID 一致

### service.go (`Service`)

依存 (interface 経由でテスト容易):
- `WalletRepository` (Get/Create/DeductConditional/UpdateBalance/SetBalance/ListAllUserIDs)
- `BudgetSettingsRepository` (Get/Set)
- `IdempotencyRepository` (TryAcquire/SaveResponse)
- `BudgetResetLogRepository` (Get/Insert)
- `now func() time.Time` (テスト用に `SetClock`)

主要ロジック:

| メソッド | UC | 主処理 |
|---|---|---|
| `GetBalance(ctx, userID)` | UC-B-03 | Wallet.Get + BudgetSettings.Get を WalletSnapshot に合成 |
| `SetBudget(ctx, userID, monthlyBudget)` | UC-B-01/02 | DR-B-01 初回判定 → 初回パス (Wallet.Create) or 変更パス (DR-B-02 差分調整: 増額 UpdateBalance / 減額 SetBalance with min 打ち切り) |
| `Deduct(ctx, userID, amount, key)` | UC-B-04 | TryAcquire → 既存 hit (DR-B-03 hash 一致 → response 再生 / 不一致 → ErrIdempotencyConflict) → DeductConditional → SaveResponse (PR-B-04 失敗結果も保存) |
| `ResetAll(ctx)` | UC-B-05 | ListAllUserIDs → 各ユーザで BudgetResetLog.Get で skip 判定 (DR-B-05) → BudgetSettings.Get → Wallet.SetBalance(monthlyBudget) → BudgetResetLog.Insert、失敗は `[]error` 集約 (P-OBS-02) |

## Repository 層 (`internal/repo/`)

各 Repository は `package init()` で `dynamodb.NewFromConfig` を初期化 (P-INIT-01)。テーブル名は env から取得:

| Package | env | 主要メソッド |
|---|---|---|
| `wallet_repo` | `DDB_TABLE_WALLET` | Get (ConsistentRead), Create (attribute_not_exists), DeductConditional (balance >= :amount, ConditionalCheckFailedException → ErrInsufficientBalance), UpdateBalance, SetBalance, ListAllUserIDs (Scan) |
| `budget_settings` | `DDB_TABLE_BUDGET_SETTINGS` | Get, Set (Upsert) |
| `idempotency` | `DDB_TABLE_IDEMPOTENCY` | TryAcquire (PutItem with `attribute_not_exists(key)`、TTL `expiresAt`), SaveResponse (UpdateItem) |
| `budget_reset_log` | `DDB_TABLE_BUDGET_RESET_LOG` | Get, Insert (`attribute_not_exists(resetDate) AND attribute_not_exists(userId)`、ErrAlreadyLogged) |

## Handler 層 (`internal/wallet/handler.go`)

エラーマッピング (凍結 IF §8):
```
ErrBudgetOutOfRange     → 400 VALIDATION_FAILED
ErrInsufficientBalance  → 402 INSUFFICIENT_BALANCE
ErrIdempotencyConflict  → 409 IDEMPOTENCY_CONFLICT
ErrInvalidInput         → 400 VALIDATION_FAILED
apperrors.ErrUnauthorized → 401 UNAUTHORIZED
default                 → 500 INTERNAL_ERROR
```

slog (P-OBS-01): `action="get_balance" | "set_budget"`、`amount` / `newBalance` を明示 attrs。

## Adapter 層 (`internal/wallet/order_adapter.go`)

`order.WalletService` interface (引数順 `(userID, idempotencyKey, amount)`) を `wallet.Service.Deduct(userID, amount, idempotencyKey)` に委譲。`OrderID = idempotencyKey` で流用。sentinel 変換:
- `wallet.ErrInsufficientBalance` → `order.ErrInsufficientFunds`
- `wallet.ErrIdempotencyConflict` → `order.ErrIdempotencyConflict`

## Scheduler Lambda (`cmd/scheduler/`)

- `lambda.Start(handler)` パターン
- `init()` で `slog.SetDefault(logging.NewContextAwareSlogHandler(...))` (LC-AUTH-05 再利用)
- `runReset(ctx, svc)` を関数として抽出してテスト可能
- bootstrap ビルド: `make build` で Linux arm64 (`provided.al2023` 対応) 14MB バイナリ

## main.go DI 配線

```go
walletRepo := wallet_repo.NewRepository()
settingsRepo := budget_settings.NewRepository()
idemRepo := idempotency.NewRepository()
resetLogRepo := budget_reset_log.NewRepository()
walletSvc := wallet.NewService(walletRepo, settingsRepo, idemRepo, resetLogRepo)

walletAdapter := wallet.NewOrderAdapter(walletSvc)
orderSvc := order.NewService(orderHistoryRepo, planBuilder, deliveryAdapter, walletAdapter)

walletHandler := wallet.NewHandler(walletSvc)
api.GET("/wallet", walletHandler.GetBalance)
api.POST("/wallet/budget", walletHandler.SetBudget)
```

`unconfiguredWalletService` ベースの `wallet_stub.go` は完全削除。`POST /api/orders` は本物の Wallet Service で 402 / 200 を返すようになった。

## テスト集計

| パッケージ | テスト数 | 種別 |
|---|---|---|
| `internal/wallet` | 31 | validation 8 + service 17 + handler 6 + adapter 4 + sentinel 1 + PBT 3 |
| `internal/repo/wallet_repo` | 3 | mock dynamodb client |
| `internal/repo/idempotency` | 3 | mock dynamodb client |
| `cmd/scheduler` | 2 | mock svc |

## 凍結 IF 整合確認

- `var _ wallet.WalletService = (*wallet.Service)(nil)` でコンパイル時保証
- `var _ wallet_repo.WalletReader = (*wallet_repo.Repository)(nil)` (Unit E 公開 IF)
- `var _ budget_settings.BudgetSettingsReader = (*budget_settings.Repository)(nil)` (Unit E)
- `var _ budget_settings.BudgetSettingsWriter = (*budget_settings.Repository)(nil)` (Unit E)
- `var _ order.WalletService = (*wallet.OrderAdapter)(nil)` (Unit C 連携)
