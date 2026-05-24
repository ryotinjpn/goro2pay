# Order Unit (Unit C) — Code Generation Plan

**Document Version**: 0.1 (Draft, awaiting user approval)
**Created**: 2026-05-24
**Unit**: C (`order` / 代行手配コア / MVP の心臓)
**Construction Depth**: Comprehensive
**Stage**: Code Generation (Construction Phase)
**Prerequisite**: Functional Design / NFR Requirements / NFR Design / Infrastructure Design 全て承認済み (PR #65, #78, #79, #80 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit C（`order` / 代行手配コア）の **Code Generation ステージ** を遂行するための作業計画を定義する。設計フェーズで凍結した仕様（FD Q-1〜Q-12 / NFR Req Q-N1〜Q-N13 / NFR Design Q-D1〜Q-D14 / Infra Design Q-I1〜Q-I13）に従い、実コード・設定・テスト・スクリプトを生成する。

### 1.1 ストーリー範囲（unit-of-work-story-map.md より）

Unit C が実装するストーリー:

| Story ID | 内容 | 受入基準ポイント |
|---|---|---|
| **US-1-01** | コアユースケース（「ご飯めんどくさい」押下） | 1 タップで注文完了、3 秒以内、フォールバック動作 |
| **US-1-03** | 注文履歴記録 | DynamoDB に Insert、TTL 90 日、PK/SK 整合 |
| **US-1-07** | 完了画面の自動遷移 | 5 秒後に MainScreen へ自動遷移 |
| **US-2-03** | サジェストからの 1 タップ注文（Unit D と共同） | suggestionId 経由、Bedrock 再呼出しなし |
| **US-X-01** | 決定疲れ解放の体験 | NFR-DEG-01 / NFR-DEG-05 ダメ化UX 体現 |

加えて以下の機能要件:
- FR-ORDER-01〜06（注文オーケストレーション全般）
- 凍結契約 §4.1 公開 Interface（OrderService / PlaceOrderRequest / PlaceOrderResult / OrderRecord）

### 1.2 Unit 依存と契約

unit-interfaces.md §4:
- **依存先**:
  - Unit A `AttachUserID` middleware（`gin.Context` から userID 取得）
  - Unit B `WalletService.Deduct`（残高引き落とし、冪等性保証）
  - Unit D `SuggestService.ResolveSuggestion`（サジェスト経由注文時、本 PR では interface 連携のみ、Unit D 実装は別 PR）
  - 横串 `BedrockAdapter` / `DeliveryAdapter` / `FallbackSuggestProvider`
- **依存元**: なし（最上位ユースケース）
- **公開 Interface**: `OrderService` (`PlaceOrder` / `GetHistory`)
- **公開 DTO**: `PlaceOrderRequest` / `PlaceOrderResult` / `OrderRecord`
- **公開 Sentinel Error**: `order.ErrInsufficientFunds` / `order.ErrIdempotencyConflict`（凍結契約 §4.3）
- **公開 Output (Terraform module)**:
  - `module.order_history`: `dynamodb_table_name` / `dynamodb_table_arn` / `dynamodb_policy_arn`
  - `module.bedrock`: `bedrock_policy_arn`
  - `module.observability`: `sns_topic_arn`

### 1.3 NFR Design / Infrastructure Design 由来の主要設計判断

| 判断 | 由来 | 影響 |
|---|---|---|
| 手動 DI（main.go で組み立て） | P-DI-01 / Q-D5 | `apps/api/main.go` への配線追加 |
| package-level `init()` で SDK 初期化 | P-INIT-01 / Q-D4 | `internal/adapters/bedrock/` / `internal/repo/order_history/` |
| `RetryClassifier` interface 抽出 | P-RETRY-01 / Q-D1 | 横串 Adapter として `internal/adapters/bedrock/retry.go` |
| `PlanBuilder` 新設 | P-PLAN-01 / Q-D2 | `internal/order/plan_builder.go` |
| Latency 計測（middleware + measure ヘルパ） | P-OBS-01 / Q-D3 | `internal/middleware/latency.go` + `internal/observability/measure.go` |
| `LogSummary` ラッパ | P-OBS-02 / Q-D6 | `internal/order/logger.go`、PII 構造的防御 |
| 3 層ログ戦略 | P-OBS-03 / Q-D13 | `internal/order/event_logger.go` で WARNING 出力 |
| Function-Field Mock | P-MOCK-01 / Q-D8 | 各 Adapter mock を closure 注入で実装 |
| PBT (gopter, P-1 + P-3) | P-PBT-01 / Q-D7 | mock + inmemory ハイブリッド |
| `useDisableLock(durationMs)` hook | P-FE-LOCK-01 / Q-D12 | `web/hooks/useDisableLock.ts` |
| `mapOrderError` 純関数 | P-FE-ERR-01 / Q-D9 | `web/lib/errorMappers.ts` |
| `getRandomToast` 純関数 | P-FE-TOAST-01 / Q-D10 | `web/lib/toasts.ts` |
| `placeholderData: keepPreviousData` | P-FE-LOAD-01 / Q-D11 | `web/hooks/useOrderHistory.ts` |
| ToastHost + Jotai atom | P-FE-TOAST-02 / Q-D14 | `web/components/ToastHost.tsx` + `web/state/toastAtoms.ts` |
| 機能別 3 module 新規 | Q-I1 見直し版 | `infra/modules/order_history/` / `bedrock/` / `observability/` |
| 既存 `api_gateway/routes.tf` 追記 | Q-I11 = B | Unit C ルート 2 本追加 |
| 既存 `lambda_api/iam.tf` 追記 | Q-I12 改定版 | `additional_policy_arns` 受け取り attach |

---

## 2. プロジェクト構造（生成対象）

Unit A 既存構造を踏襲、Unit C で追加・拡張するファイルを `★` で示す。

```
goro2pay/
├── apps/
│   └── api/                              # API Lambda (Unit A 既存)
│       ├── go.mod                        # ★ 依存追加: aws-sdk-go-v2 bedrockruntime / dynamodb / oklog/ulid v2 / gopter
│       ├── go.sum                        # ★ go mod tidy で更新
│       ├── main.go                       # ★ Unit C component の DI 配線追加（P-DI-01）
│       └── internal/
│           ├── auth/                     # Unit A 既存（変更なし）
│           ├── logging/                  # Unit A 既存（変更なし）
│           ├── handlers/                 # Unit A 既存
│           │   ├── order.go              # ★ 新規 OrderHandler (LC-ORDER-02)
│           │   └── order_test.go         # ★ 単体テスト
│           ├── apperrors/                # Unit A 既存
│           ├── middleware/               # ★ 新規（横串）
│           │   ├── latency.go            # ★ LatencyMiddleware (LC-12 / P-OBS-01)
│           │   └── latency_test.go
│           ├── observability/            # ★ 新規（横串）
│           │   ├── measure.go            # ★ Measure ヘルパ (LC-13 / P-OBS-01)
│           │   └── measure_test.go
│           ├── order/                    # ★ 新規（Unit C 中核）
│           │   ├── service.go            # ★ OrderService (LC-01)
│           │   ├── service_test.go       # ★ 単体テスト + 統合 9 シナリオ
│           │   ├── service_pbt_test.go   # ★ PBT (gopter, P-1 + P-3)
│           │   ├── handler.go            # ★ → handlers/order.go へ移動（責務分離）
│           │   ├── plan_builder.go       # ★ PlanBuilder (LC-08 / P-PLAN-01)
│           │   ├── plan_builder_test.go
│           │   ├── logger.go             # ★ LogSummary (LC-10 / P-OBS-02)
│           │   ├── logger_test.go
│           │   ├── event_logger.go       # ★ EventLogger (LC-11 / P-OBS-03)
│           │   ├── event_logger_test.go
│           │   ├── types.go              # ★ DTO (LC-03)
│           │   ├── wallet_stub_test.go   # ★ inmemory WalletStub (LC-19、PBT 用)
│           │   └── wallet.go             # ★ WalletService interface (Unit B 連携、unit-interfaces §3.1)
│           ├── repo/
│           │   └── order_history/        # ★ 新規（Unit C 専用）
│           │       ├── repository.go     # ★ DynamoDB CRUD + init() で SDK 初期化 (LC-04 / LC-15 / P-INIT-01)
│           │       ├── repository_test.go
│           │       └── inmemory.go       # ★ inmemory 実装（テスト専用 LC-20）
│           └── adapters/                 # ★ 新規（横串、Unit C / D 共有候補）
│               ├── bedrock/
│               │   ├── adapter.go        # ★ ClaudeBedrockAdapter + init() (LC-05 / LC-14 / P-INIT-01)
│               │   ├── adapter_test.go
│               │   ├── retry.go          # ★ RetryClassifier (LC-07 / P-RETRY-01)
│               │   ├── retry_test.go
│               │   ├── prompt.go         # ★ プロンプトテンプレート（FD §2 由来）
│               │   ├── prompt_test.go
│               │   └── mock.go           # ★ MockBedrockAdapter (LC-16 / P-MOCK-01)
│               ├── delivery/
│               │   ├── adapter.go        # ★ DeliveryAdapter interface + MockDeliveryAdapter (LC-06)
│               │   ├── adapter_test.go
│               │   └── mock.go           # ★ closure 注入対応 mock 拡張 (LC-17)
│               └── fallback/
│                   ├── provider.go       # ★ FallbackSuggestProvider (LC-09)
│                   ├── provider_test.go
│                   ├── stores.go         # ★ 5 店舗ラインナップ定数（BR-C07）
│                   └── mock.go           # ★ MockFallbackProvider (LC-18)
│
├── web/                                  # Next.js (Unit A 既存)
│   ├── package.json                      # ★ 依存追加: ulid (Frontend ULID)
│   ├── app/
│   │   ├── layout.tsx                    # ★ <ToastHost /> 追加
│   │   ├── page.tsx                      # ★ MainScreen に GoroButton + OrderHistoryList 追加（Unit A の placeholder を置換）
│   │   └── order/                        # ★ 新規
│   │       └── [id]/
│   │           └── complete/
│   │               └── page.tsx          # ★ OrderCompletionScreen (LC-32)
│   ├── components/
│   │   ├── order/                        # ★ 新規
│   │   │   ├── GoroButton.tsx            # ★ LC-21
│   │   │   ├── OrderHistoryList.tsx      # ★ LC-24
│   │   │   ├── OrderHistorySkeleton.tsx  # ★ LC-24
│   │   │   ├── ToastHost.tsx             # ★ LC-30
│   │   │   └── Toast.tsx                 # ★ LC-31
│   │   └── auth/                         # Unit A 既存（変更なし）
│   ├── hooks/
│   │   ├── useOrder.ts                   # ★ LC-22
│   │   ├── useOrderHistory.ts            # ★ LC-23
│   │   ├── useDisableLock.ts             # ★ LC-25 / P-FE-LOCK-01
│   │   └── useToast.ts                   # ★ LC-29
│   ├── lib/
│   │   ├── errorMappers.ts               # ★ LC-26 / P-FE-ERR-01
│   │   ├── toasts.ts                     # ★ LC-27 / P-FE-TOAST-01
│   │   ├── ulid.ts                       # ★ LC-34
│   │   └── api/
│   │       └── orders.ts                 # ★ LC-33（apiClient.placeOrder / fetchOrderHistory）
│   ├── state/
│   │   └── toastAtoms.ts                 # ★ LC-28
│   └── tests/
│       ├── useOrder.test.tsx             # ★
│       ├── useOrderHistory.test.tsx      # ★
│       ├── useDisableLock.test.tsx       # ★ fakeTimers で時間境界検証
│       ├── useToast.test.tsx             # ★
│       ├── errorMappers.test.ts          # ★ 純関数テスト
│       ├── toasts.test.ts                # ★ Math.random mock
│       ├── orders-api.test.ts            # ★ msw で API 呼出テスト
│       └── ToastHost.test.tsx            # ★ 4 件追加で 3 件表示
│
└── infra/
    ├── modules/
    │   ├── (Unit A 既存 5 module: codestar_connection / cognito / api_gateway / lambda_api / amplify)
    │   ├── api_gateway/                  # ★ routes.tf に Unit C ルート 2 本追加
    │   │   └── routes.tf                 # ★ POST /api/orders + GET /api/orders integration + route
    │   ├── lambda_api/                   # ★ additional_policy_arns 受け取り
    │   │   ├── variables.tf              # ★ additional_policy_arns + order_history_table_name 変数追加
    │   │   ├── iam.tf                    # ★ for_each で attach
    │   │   ├── api_lambda.tf             # ★ environment 変数追加
    │   │   └── outputs.tf                # ★ api_log_group_name output 追加
    │   ├── budget/                       # Unit B 既存（変更なし）
    │   ├── order_history/                # ★ 新規（Unit C 専用）
    │   │   ├── README.md
    │   │   ├── main.tf                   # DynamoDB OrderHistory + IAM Policy
    │   │   ├── locals.tf                 # table_name / policy_name
    │   │   ├── variables.tf              # env / tags
    │   │   ├── outputs.tf                # dynamodb_table_name / arn / policy_arn
    │   │   └── tests/
    │   │       └── dynamodb_schema.tftest.hcl
    │   ├── bedrock/                      # ★ 新規（Unit C / D 共有候補）
    │   │   ├── README.md
    │   │   ├── main.tf                   # Bedrock IAM Policy
    │   │   ├── data.tf                   # aws_caller_identity / aws_region
    │   │   ├── locals.tf                 # model_id / inference_profile_id / model_arns
    │   │   ├── variables.tf              # env / region / tags
    │   │   ├── outputs.tf                # bedrock_policy_arn
    │   │   └── tests/
    │   │       └── bedrock_iam_least_privilege.tftest.hcl
    │   └── observability/                # ★ 新規（横串、Unit D/E 再利用可）
    │       ├── README.md
    │       ├── main.tf                   # SNS Topic + metric filter ×3 + alarm ×3 + Budgets
    │       ├── data.tf                   # aws_caller_identity (Budgets account_id)
    │       ├── locals.tf                 # alarm 名・しきい値・metric namespace
    │       ├── variables.tf              # alarm_email / api_log_group_name / 4 種閾値 / env / tags
    │       ├── outputs.tf                # sns_topic_arn
    │       └── tests/
    │           └── cloudwatch_alarms.tftest.hcl
    └── envs/
        └── dev/
            ├── locals.tf                 # ★ alarm_email = "alerts@example.com" 追加
            ├── main.tf                   # ★ Unit C 3 module 呼出 + module.lambda_api への additional_policy_arns 追加
            └── outputs.tf                # ★ order_history_table_name / sns_topic_arn 追加
```

---

## 3. 作業手順（番号付きステップ、Plan 承認後に実行）

各ステップは「実装 → 単体テスト → サマリ」の小ループを取る。サマリは `aidlc-docs/construction/order/code/` 配下に Markdown で記録（コード本体ではなく要約）。

### Step 1: プロジェクト構造セットアップ + Go モジュール依存追加

- [x] `apps/api/go.mod` に依存追加:
  - `github.com/aws/aws-sdk-go-v2/service/bedrockruntime`（Bedrock SDK）
  - `github.com/aws/aws-sdk-go-v2/service/dynamodb`（DynamoDB SDK）
  - `github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue`
  - `github.com/oklog/ulid/v2`（OrderID 採番）
  - （`gopter`, `aws-sdk-go-v2/config`, `aws/smithy-go` は Unit A で既に追加済みなら再利用）
- [x] `apps/api/go.sum` を `go mod tidy` で更新
- [x] `web/package.json` に `ulid` npm パッケージ追加
- [x] `package-lock.json` を `npm install` で更新
- [x] 新規ディレクトリ作成:
  - `apps/api/internal/middleware/`
  - `apps/api/internal/observability/`
  - `apps/api/internal/order/`
  - `apps/api/internal/repo/order_history/`
  - `apps/api/internal/adapters/{bedrock,delivery,fallback}/`
  - `web/components/order/`
  - `web/app/order/[id]/complete/`
  - `infra/modules/{order_history,bedrock,observability}/`
  - 各 `tests/` サブディレクトリ

### Step 2: Backend Adapter 層実装（横串）

#### Step 2.1: BedrockAdapter + RetryClassifier
- [x] `internal/adapters/bedrock/retry.go` — `RetryClassifier` interface + `BedrockRetryClassifier` 実装（`errors.As` で SDK 型判定、P-RETRY-01）
- [x] `internal/adapters/bedrock/retry_test.go` — 各エラー型のリトライ判定検証
- [x] `internal/adapters/bedrock/prompt.go` — プロンプトテンプレート（履歴 + 曜日 + カテゴリ → JSON 出力指示、FD §2.2）
- [x] `internal/adapters/bedrock/prompt_test.go` — テンプレート組み立て検証
- [x] `internal/adapters/bedrock/adapter.go` — `BedrockAdapter` interface + `ClaudeBedrockAdapter` 実装、package-level `init()` で SDK Client 初期化（P-INIT-01）、`InferOrderPlan` 内でリトライループ + Context タイムアウト 1.5s
- [x] `internal/adapters/bedrock/adapter_test.go` — Mock Bedrock SDK でリトライ・タイムアウト・成功・永続エラーの 4 シナリオ検証
- [x] `internal/adapters/bedrock/mock.go` — `MockBedrockAdapter`（function-field、`InferOrderPlanFunc` + `Calls`、P-MOCK-01）

#### Step 2.2: DeliveryAdapter
- [x] `internal/adapters/delivery/adapter.go` — `DeliveryAdapter` interface + `MockDeliveryAdapter`（noop で nil、`slog.InfoContext` ログ出力のみ）
- [x] `internal/adapters/delivery/adapter_test.go`
- [x] `internal/adapters/delivery/mock.go` — closure 注入対応 mock（テスト専用、`PlaceFunc` + `Calls`）

#### Step 2.3: FallbackSuggestProvider
- [x] `internal/adapters/fallback/stores.go` — 5 店舗ラインナップ定数配列（BR-C07）
- [x] `internal/adapters/fallback/provider.go` — `FallbackSuggestProvider` interface + 実装（`BuildFromHistory` 最頻パターン抽出 + `Default` ランダム選択）
- [x] `internal/adapters/fallback/provider_test.go` — 各メソッドの境界値テスト
- [x] `internal/adapters/fallback/mock.go` — closure 注入対応 mock

**ストーリー対応**: BR-C01〜C07 / FD §2.2 / NFRC-C06 / NFRC-C07 / NFRC-C20

### Step 3: Backend Repository 層実装

- [x] `internal/repo/order_history/repository.go` — `OrderHistoryRepository` interface + 実装、package-level `init()` で SDK Client 初期化（P-INIT-01）、`Insert` / `GetItem` / `Query` 実装。属性名は camelCase（`userId / orderId / orderedAt / expiresAt` 等）
- [x] `internal/repo/order_history/repository_test.go` — DynamoDB Local / DynamoDB SDK Mock で CRUD 検証
- [x] `internal/repo/order_history/inmemory.go` — `InmemoryHistory` 実装（テスト専用、PBT 用、LC-20）

**ストーリー対応**: US-1-03 / domain-entities.md §2.1 / 凍結契約 §3.2

### Step 4: Backend Order 中核実装

#### Step 4.1: 型定義 + Wallet interface
- [x] `internal/order/types.go` — `PlaceOrderRequest` / `PlaceOrderResult` / `OrderRecord` / `Plan` / `Source` enum（凍結契約 §4.1 + Unit C 内部追加属性）
- [x] `internal/order/wallet.go` — `WalletService` interface 定義（unit-interfaces §3.1 整合、Unit B 実装を import 不要にするため Unit C 側に再宣言）

#### Step 4.2: PlanBuilder（P-PLAN-01）
- [x] `internal/order/plan_builder.go` — `PlanBuilder` interface + `BedrockPlanBuilder` 実装。Bedrock 呼出 → リトライ判定 → 失敗時フォールバック分岐（履歴 5 件閾値、BR-C06）
- [x] `internal/order/plan_builder_test.go` — 4 シナリオ（成功 / リトライ後成功 / フォールバック履歴 / フォールバック Default）

#### Step 4.3: LogSummary + EventLogger（P-OBS-02 / P-OBS-03）
- [x] `internal/order/logger.go` — `LogSummary` struct（11 項目）+ setter + `LogComplete()`
- [x] `internal/order/logger_test.go` — 11 項目の出力検証、PII 構造的防御確認（Bedrock 本文フィールド非存在）
- [x] `internal/order/event_logger.go` — `LogBedrockRetry` / `LogFallbackTriggered` 関数
- [x] `internal/order/event_logger_test.go` — slog.Warn 出力フォーマット検証

#### Step 4.4: OrderService（LC-01）
- [x] `internal/order/service.go` — `OrderService` interface + struct 実装。`PlaceOrder` でオーケストレーション（冪等性 → PlanBuilder → Wallet.Deduct → DeliveryAdapter.Place → OrderHistory.Insert → 結果組み立て）、`GetHistory` で OrderHistoryRepository.Query
- [x] `internal/order/service_test.go` — NFRC-C17 統合テスト 9 シナリオ（通常 / リトライ成功 / フォールバック / 永続エラー / 冪等命中 / 連打 / 残高不足 / Context Timeout / Context Cancel）

#### Step 4.5: PBT（P-PBT-01）
- [x] `internal/order/wallet_stub_test.go` — `WalletStub`（inmemory、LC-19、`Deduct` 実装で同一 idempotencyKey に対して同一結果を返す）
- [x] `internal/order/service_pbt_test.go` — gopter で P-1（残高不変）+ P-3（冪等レスポンス一貫性）プロパティ実装、各 100 サンプル

**ストーリー対応**: US-1-01 / US-1-03 / US-2-03 / FD 全般 / NFRC-C01 / NFRC-C05 / NFRC-C09 / NFRC-C12 / NFRC-C15 / NFRC-C17

### Step 5: Backend 観測性（横串）

- [x] `internal/middleware/latency.go` — `LatencyMiddleware`（Gin handler chain 最初、E2E レイテンシ計測、P-OBS-01）
- [x] `internal/middleware/latency_test.go` — `httptest` + `bytes.Buffer` で `request_complete` ログ出力検証
- [x] `internal/observability/measure.go` — generic 関数 `Measure[T any](ctx, name, fn) (T, time.Duration, error)`
- [x] `internal/observability/measure_test.go` — 各シナリオ（成功 / 失敗 / panic 伝播）の検証

**ストーリー対応**: NFRC-C01 / NFRC-C12 / NFRC-C13

### Step 6: Backend Handler 層 + main.go 配線

- [x] `internal/handlers/order.go` — `OrderHandler` 実装（Gin handler、`POST /api/orders` / `GET /api/orders`、エラーマッピング 402/409/500/504、LC-02）
- [x] `internal/handlers/order_test.go` — エラーマッピング、HTTP ステータス検証
- [x] `apps/api/main.go` — Unit C component の DI 配線追加:
  1. `bedrock.NewClaudeBedrockAdapter()` 生成（init() で SDK 初期化済み）
  2. `delivery.NewMockDeliveryAdapter()` 生成
  3. `fallback.NewFallbackSuggestProvider()` 生成
  4. `order.NewBedrockPlanBuilder(bedrockAdapter, fallbackProvider, retryClassifier)` 生成
  5. `orderhistory.NewRepository()` 生成（init() で SDK 初期化済み）
  6. Unit B WalletService 取得（既存 main.go の組み立てに依存、要 Unit B 実装連携）
  7. `order.NewOrderService(historyRepo, planBuilder, deliveryAdapter, walletSvc)` 生成
  8. `handlers.NewOrderHandler(orderSvc)` 生成
  9. `LatencyMiddleware()` を Gin の middleware chain 最初に登録
  10. `r.POST("/api/orders", orderHandler.PlaceOrder)` / `r.GET("/api/orders", orderHandler.GetHistory)` 登録
- [ ] `main_test.go`（既存に追記、ヘルスチェック等で Unit C component 配線が壊れていないことを確認） — **Build & Test ステージへ繰越**（D-I1: 配線テストは E2E (Playwright) と統合してまとめて実装）

**ストーリー対応**: 凍結契約 §4.1 / FD § 全般 / P-DI-01

### Step 7: Backend サマリ
- [x] `aidlc-docs/construction/order/code/business-logic-summary.md` — Order 中核（OrderService / PlanBuilder / FallbackSuggestProvider）の生成ファイル + 関数シグネチャ + テスト結果概要
- [x] `aidlc-docs/construction/order/code/api-layer-summary.md` — handler + middleware + observability の概要
- [x] `aidlc-docs/construction/order/code/repository-layer-summary.md` — OrderHistoryRepository + DynamoDB スキーマ
- [x] `aidlc-docs/construction/order/code/adapters-summary.md` — Bedrock / Delivery / Fallback Adapter の概要

### Step 8: Frontend lib + state 実装

- [x] `web/lib/ulid.ts` — `generateUlid()` 純関数（`ulid` npm パッケージラップ、LC-34）
- [x] `web/lib/toasts.ts` — `TOAST_VARIANTS` readonly 配列 + `getRandomToast()` 純関数（LC-27 / P-FE-TOAST-01）
- [x] `web/lib/errorMappers.ts` — `OrderErrorAction` 型 + `mapOrderError(err)` 純関数（LC-26 / P-FE-ERR-01）
- [x] `web/lib/api/orders.ts` — `apiClient.placeOrder` / `fetchOrderHistory`（既存 Unit A `apiClient` ラッパ利用、Authorization: Bearer 透過、LC-33）
- [x] `web/state/toastAtoms.ts` — `toastsAtom: atom<ToastItem[]>` + `ToastItem` 型（LC-28）

**ストーリー対応**: NFRC-C19 / NFRC-C22 / FD frontend-components.md

### Step 9: Frontend hooks 実装

- [x] `web/hooks/useDisableLock.ts` — `useDisableLock(durationMs) → { isLocked, triggerLock }`、`useEffect` cleanup でタイマー破棄（LC-25 / P-FE-LOCK-01）
- [x] `web/hooks/useToast.ts` — `useToast() → { toasts, showToast }`、Jotai atom + setTimeout 自動消去（LC-29）
- [x] `web/hooks/useOrderHistory.ts` — `useQuery` で `placeholderData: keepPreviousData` + `staleTime: 60_000` + `gcTime: 300_000`（LC-23 / P-FE-LOAD-01）
- [x] `web/hooks/useOrder.ts` — `useMutation`、`retry: 0`、`onSuccess` で `triggerLock` + `invalidateQueries(['orderHistory'])` + `invalidateQueries(['balance'])`、`onError` で `triggerLock` + `mapOrderError(err)` の戻り値で分岐実行（LC-22 / P-FE-ERR-01 / P-FE-LOCK-01）

### Step 10: Frontend components 実装

- [x] `web/components/order/Toast.tsx` — 個別トースト、`role="status"`、アニメーション（LC-31）
- [x] `web/components/order/ToastHost.tsx` — トースト一覧表示、`slice(0, 3)` で最大 3 件、`role="region"` + `aria-live="polite"`（LC-30 / P-FE-TOAST-02）
- [x] `web/components/order/GoroButton.tsx` — `useOrder` hook 使用、`disabled` 属性で連打抑制（LC-21）
- [x] `web/components/order/OrderHistoryList.tsx` — `useOrderHistory` hook 使用、空状態（0 件）はダメ化文言（LC-24）
- [x] `web/components/order/OrderHistorySkeleton.tsx` — スケルトン表示（LC-24）

### Step 11: Frontend pages 配置 + ToastHost 統合

- [x] `web/app/layout.tsx` — `<ToastHost />` を `<body>` 直下に追加（既存 `<AppProviders>` 配下に配置）
- [x] `web/app/page.tsx` — MainScreen に `<GoroButton />` + `<OrderHistoryList />` を統合（Unit A の MainScreen placeholder を Unit C 用に更新）
- [x] `web/app/order/[id]/complete/page.tsx` — `OrderCompletionScreen`、`useEffect` + `setTimeout(5000)` で `router.push('/')`（LC-32）

### Step 12: Frontend テスト

- [x] `web/tests/toasts.test.ts` — `vi.spyOn(Math, 'random')` で決定論的検証
- [x] `web/tests/errorMappers.test.ts` — 全エラーケース（402 / 409 / 500 / NetworkError / その他）の純関数検証
- [x] `web/tests/useDisableLock.test.tsx` — `vi.useFakeTimers()` + `renderHook` + `act` で 999ms / 1000ms / 1001ms 境界検証
- [x] `web/tests/useToast.test.tsx` — `showToast` で atom 更新 + setTimeout 消去検証
- [ ] `web/tests/useOrder.test.tsx` — msw + React Testing Library で `onError` callback 起動検証 — **Build & Test ステージへ繰越**（D-I1: msw 導入は本 PR スコープ外、`web/tests/OrderHistoryList.test.tsx` で render-level 検証は実装済み）
- [ ] `web/tests/useOrderHistory.test.tsx` — msw で 2 回目 fetch 遅延 → `placeholderData` 動作検証 — **Build & Test ステージへ繰越**（D-I1: 同上、msw 環境整備とセットで実装）
- [x] `web/tests/orders-api.test.ts` — `apiClient.placeOrder` + `fetchOrderHistory` の HTTP 呼出検証
- [x] `web/tests/ToastHost.test.tsx` — 4 件追加で 3 件のみ表示、4 件目はキューイング検証

### Step 13: Frontend サマリ
- [x] `aidlc-docs/construction/order/code/frontend-summary.md` — 生成ファイル + コンポーネント階層 + テスト結果概要

### Step 14: Infrastructure - 新規 module 3 種

#### Step 14.1: `infra/modules/order_history/`
- [x] `main.tf` — `aws_dynamodb_table.order_history` + `aws_iam_policy.dynamodb_order_history`（infrastructure-design.md §3.1）
- [x] `locals.tf` — `table_name` / `policy_name` / `tags`
- [x] `variables.tf` — `env` / `tags`
- [x] `outputs.tf` — `dynamodb_table_name` / `dynamodb_table_arn` / `dynamodb_policy_arn`
- [x] `README.md`
- [x] `tests/dynamodb_schema.tftest.hcl` — PK/SK/TTL/暗号化/PITR/キャパシティ検証（mock_provider）

#### Step 14.2: `infra/modules/bedrock/`
- [x] `main.tf` — `aws_iam_policy.bedrock_inference`（Foundation Model + Inference Profile ARN 限定）
- [x] `data.tf` — `aws_caller_identity.current` / `aws_region.current`
- [x] `locals.tf` — `model_id` / `inference_profile_id` / `model_arns` / `policy_name`
- [x] `variables.tf` — `env` / `region` / `tags`
- [x] `outputs.tf` — `bedrock_policy_arn`
- [x] `README.md`
- [x] `tests/bedrock_iam_least_privilege.tftest.hcl` — Action / Resource ARN 制限検証

#### Step 14.3: `infra/modules/observability/`
- [x] `main.tf` — SNS Topic + Subscription + metric filter ×3 + alarm ×3 + Budgets（infrastructure-design.md §3.3）
- [x] `data.tf` — `aws_caller_identity.current`（Budgets account_id）
- [x] `locals.tf` — `env_prefix` / `sns_topic_name` / `metric_namespace` / `budget_start` / `tags`
- [x] `variables.tf` — `env` / `alarm_email` / `api_log_group_name` / `p95_threshold_ms` (default 3000) / `retry_threshold_count` (default 5) / `fallback_threshold_count` (default 3) / `bedrock_budget_limit_usd` (default 5) / `tags`
- [x] `outputs.tf` — `sns_topic_arn`
- [x] `README.md`
- [x] `tests/cloudwatch_alarms.tftest.hcl` — 3 alarm の閾値・SNS 紐付け / Budgets 通知検証

### Step 15: Infrastructure - 既存 module への追記

#### Step 15.1: `infra/modules/api_gateway/routes.tf` 追記
- [x] Unit C ルート 2 本（`POST /api/orders` / `GET /api/orders`）の `aws_apigatewayv2_integration` + `aws_apigatewayv2_route` 追加
- [ ] tests/api_gateway_basic.tftest.hcl に Unit C routes 検証ケース追加 — **Build & Test ステージへ繰越**（D-I1: 既存 Unit A tftest を不変としたため Unit C 側の `aws_apigatewayv2_route` リソース追加検証は次ステージで対応）

#### Step 15.2: `infra/modules/lambda_api/`
- [x] `variables.tf` — `additional_policy_arns: list(string)`（default `[]`）+ `order_history_table_name: string` 追加
- [x] `iam.tf` — `aws_iam_role_policy_attachment.additional` を `for_each = toset(var.additional_policy_arns)` で追加
- [x] `api_lambda.tf` — `environment.variables` に `ORDER_HISTORY_TABLE_NAME` + `BEDROCK_INFERENCE_PROFILE_ID` 追加
- [x] `outputs.tf` — `api_log_group_name` output 追加
- [ ] tests/lambda_api_basic.tftest.hcl に additional_policy_arns 検証ケース追加 — **Build & Test ステージへ繰越**（D-I1: 既存 Unit A tftest を不変としたため、追加 attach の検証は次ステージで対応）

### Step 16: Infrastructure - envs/dev/ への追記

- [x] `infra/envs/dev/locals.tf` — `alarm_email = "alerts@example.com"` 追加
- [x] `infra/envs/dev/main.tf` — `module "order_history"` / `module "bedrock"` / `module "observability"` 呼出追加 + 既存 `module "lambda_api"` に `additional_policy_arns` + `order_history_table_name` パラメータ追加
- [x] `infra/envs/dev/outputs.tf` — `order_history_table_name` / `sns_topic_arn` output 追加

### Step 17: Infrastructure サマリ
- [x] `aidlc-docs/construction/order/code/infrastructure-summary.md` — 新規 3 module + 既存 2 module 追記の概要 + terraform-test 結果

### Step 18: Documentation 更新

- [x] `aidlc-docs/construction/order/code/deployment-runbook.md` — Unit C 初回 apply 手順、SNS 購読確認、Bedrock モデルアクセス申請、CodePipeline トリガー、Frontend Amplify 自動デプロイ手順、ハッカソン実演チェックリスト
- [ ] ルート `README.md` 更新（Unit C 完了後の利用手順を追記） — **Build & Test ステージへ繰越**（D-I1: ルート README は Unit B/D/E 完了後にまとめて全 Unit 視点で更新する方が情報の一貫性が保たれる）

### Step 19: 完了確認とサマリ

- [x] `aidlc-docs/construction/order/code/code-generation-summary.md` — 全成果物一覧、ステップ別チェック結果、生成ファイル数、テスト結果サマリ
- [x] `aidlc-state.md` の Construction → Code Generation (Unit C) を [x] に更新
- [x] `audit.md` に完了記録を追記
- [x] Plan のチェックボックス全て [x] に更新

---

## 4. ストーリー トレーサビリティ

| Step | 関連ストーリー / FR | 関連 NFR | 関連 LC | 関連パターン |
|---|---|---|---|---|
| Step 2 | FR-ORDER-01〜06 (Bedrock 呼出 / フォールバック) | NFRC-C06 / C07 / C20 | LC-05 / 07 / 09 | P-RETRY-01 / P-INIT-01 / P-MOCK-01 |
| Step 3 | US-1-03 (履歴記録) | NFRC-C12 / C19 | LC-04 / 15 / 20 | P-INIT-01 |
| Step 4.1〜4.4 | US-1-01 / US-1-03 / US-2-03 (中核ユースケース) | NFRC-C01 / C05 / C08 / C09 | LC-01 / 03 / 08 / 10 / 11 | P-PLAN-01 / P-OBS-02 / P-OBS-03 |
| Step 4.5 | (品質投資) | NFRC-C15 / C17 | LC-19 / 20 | P-PBT-01 |
| Step 5 | (横串観測性) | NFRC-C01 / C12 / C13 | LC-12 / 13 | P-OBS-01 |
| Step 6 | 凍結契約 §4.1 | NFRC-C18 | LC-02 | P-DI-01 |
| Step 8 | NFR-DEG-05 (自虐文言) | NFRC-C22 / C24 | LC-26 / 27 / 28 / 33 / 34 | P-FE-ERR-01 / P-FE-TOAST-01 |
| Step 9 | NFR-DEG-01 (低摩擦) | NFRC-C19 / C22 | LC-22 / 23 / 25 / 29 | P-FE-LOAD-01 / P-FE-LOCK-01 |
| Step 10 | US-1-01 / US-1-07 | NFRC-C21 / C22 | LC-21 / 24 / 30 / 31 | P-FE-TOAST-02 |
| Step 11 | US-1-07 (5 秒自動遷移) | NFRC-C21 | LC-32 | — |
| Step 14 | (Unit C インフラ) | NFRC-C13 / C18 / C20 | LC-04 / 05 | Q-I1〜Q-I13 |
| Step 15 | (横串インフラ追記) | NFRC-C13 | — | Q-I11 / Q-I12 |
| Step 18 | (デプロイ手順) | — | — | — |

---

## 5. テスト戦略

### 5.1 Backend テストカバレッジ

| 種別 | 対象 | フレームワーク | カバレッジ目標 |
|---|---|---|---|
| 単体テスト | 全 Adapter / Repository / 純関数 | `testing` 標準 | 各ファイル ≥ 80% |
| PBT | OrderService の P-1 / P-3 | `gopter` | 各 100 サンプル |
| 統合テスト | NFRC-C17 9 シナリオ | `testing` + mock | 全シナリオ通過 |

### 5.2 Frontend テストカバレッジ

| 種別 | 対象 | フレームワーク | カバレッジ目標 |
|---|---|---|---|
| 単体テスト（純関数） | errorMappers / toasts / ulid | Vitest | 100% branch coverage |
| Hook テスト | useOrder / useOrderHistory / useDisableLock / useToast | Vitest + `renderHook` + `vi.useFakeTimers()` | 主要ケース網羅 |
| Component テスト | GoroButton / OrderHistoryList / Toast / ToastHost | Vitest + React Testing Library | レンダリング検証 |
| API モック | orders-api | msw | 401/402/409/500 各ケース |

### 5.3 Infrastructure テスト

| 種別 | 対象 | フレームワーク |
|---|---|---|
| terraform-test | DynamoDB スキーマ / Bedrock IAM 最小権限 / Alarms 閾値 | `mock_provider` でオフライン |

---

## 6. テストスタブ方針（NFRC-C16 整合）

| 環境 | Bedrock 呼出し | スタブ手法 |
|---|---|---|
| `go test`（local） | mock 必須 | `MockBedrockAdapter`（function-field closure） |
| CI（GitHub Actions） | mock 必須 | 同上、AWS 認証情報なし |
| dev 環境（AWS） | 実呼出し許容 | `IS_TEST` 環境変数なし時に実 SDK |
| stg / prd 環境 | 実呼出し | 同上 |

---

## 7. Code Generation 段階での確定事項

### 7.1 自虐トースト最終文言（NFRC-C22 / Q-D10）

NFR Design で例示した 3 文言を最終確定:

```typescript
export const TOAST_VARIANTS: readonly string[] = [
  "サーバーがやる気を失いました…もう一度お試しください",
  "システムがふぬけてます。少し待ってあげてください",
  "今日はちょっとダメ化に失敗しました。再挑戦しますか？",
] as const;
```

### 7.2 Bedrock プロンプトテンプレート（FD §2.2）

```
あなたは「人をダメにする」食事代行アプリのアシスタントです。
以下のユーザの直近の注文履歴と現在の曜日を踏まえ、最も「ダメ化を促す」食事プランを 1 つ JSON で提案してください。

## 履歴
{history_json}

## 現在の曜日
{day_of_week}

## カテゴリ
food

## 出力フォーマット（JSON のみ、それ以外の文字列は一切含めないこと）
{
  "storeName": "店舗名（短く）",
  "menuName": "メニュー名",
  "amount": 整数（円、500〜3000 の範囲）,
  "category": "food"
}
```

### 7.3 環境変数（API Lambda、Unit C 関連）

| 環境変数 | デフォルト値 | 由来 |
|---|---|---|
| `ORDER_HISTORY_TABLE_NAME` | `gp-dev-order-history`（Terraform output 経由） | Step 15.2 |
| `BEDROCK_INFERENCE_PROFILE_ID` | `apac.anthropic.claude-3-5-haiku-20241022-v1:0` | NFRC-C20 |
| `AWS_REGION` | `ap-northeast-1`（Lambda 標準注入） | P-INIT-01 |

### 7.4 Bedrock Mock の 4 シナリオ

| シナリオ | closure 設定例 |
|---|---|
| 成功 | `mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) { return successPlan, nil }` |
| Throttle | `mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) { return nil, &types.ThrottlingException{Message: aws.String("Rate exceeded")} }` |
| Timeout | `mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) { return nil, context.DeadlineExceeded }` |
| 永続エラー | `mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) { return nil, &types.ValidationException{} }` |

---

## 8. 想定スコープと工数

- **生成ファイル数: 約 80 ファイル**
  - Backend Go: 約 35 ファイル（コード ~20 + テスト ~15）
  - Frontend TypeScript: 約 25 ファイル（コード ~15 + テスト ~10）
  - Infrastructure Terraform: 約 20 ファイル（module ~16 + tests ~3 + envs/dev 編集 ~3）
- **ドキュメントサマリ**: 6 種（business-logic / api-layer / repository-layer / adapters / frontend / infrastructure / code-generation 全体）
- **テストファイル**: 約 25 種（Go ユニット + PBT + 統合、TS ユニット + Hook + Component、Terraform tftest）

Unit C は MVP の心臓部（Comprehensive 深度）で、Unit A（70 ファイル）と同等以上の規模。Bedrock 統合 + PBT + 統合テスト 9 シナリオが Unit A との主な差分。

---

## 9. 想定外の論点（後続ステージへの引き継ぎ）

- **Build & Test ステージ**: `go test ./...` / `npm test` / `terraform test` の CI 統合、E2E (Playwright) シナリオ実行、Bedrock 実呼出し検証、初回デプロイ動作確認
- **Unit B 連携**: 本 PR 時点で Unit B `WalletService` の Go 実装が完了している前提（Unit B Code Generation が後続 Unit C より先行している場合）。Unit B が未実装の場合は Unit C 側で `WalletService` interface のみ定義し、本実装は Unit B 完了時に main.go の DI 配線を更新
- **Unit D 連携**: `SuggestService.ResolveSuggestion` は Unit D 未実装のため、Unit C は `suggestionId != nil` 時の処理を `if suggestSvc != nil { ... }` ガードでスキップ可能にしておく（Unit D 完成時に main.go で注入）
- **本番化**: Bedrock コスト最適化（Provisioned Throughput 検討）、DynamoDB PITR 有効化、Customer Managed KMS、CloudWatch Dashboard 構築

---

## 10. 承認ゲート

本 Plan の構成（生成範囲・ステップ順序・テスト方針）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — ステップ追加削除、生成範囲調整、優先順位変更等
- ✅ **Approve & Start Generation** — Plan 承認、Part 2 Generation を開始（Step 1 から順次実行、約 80 ファイル生成）
