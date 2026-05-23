# Order Unit (Unit C) — Deployment Architecture

**Document Version**: 1.0
**Created**: 2026-05-24
**Stage**: Construction / Infrastructure Design
**Unit**: C — `order` (代行手配コア / MVP の心臓)
**Depth**: Comprehensive
**Related**:
- [infrastructure-design.md](./infrastructure-design.md)（Terraform リソース詳細）
- [logical-components.md](../nfr-design/logical-components.md)（論理コンポーネント）
- [nfr-design-patterns.md](../nfr-design/nfr-design-patterns.md)（設計パターン）
- Unit A `auth/infrastructure-design/deployment-architecture.md`

本ドキュメントは Unit C の **デプロイ構成と実行時データフロー** を可視化する。Unit A 既存インフラ（API Gateway / Lambda / Cognito / Amplify Hosting）を再利用しつつ、Unit C 固有の DynamoDB / Bedrock / 観測性レイヤーの統合を示す。

---

## 1. 全体構成図

```
[ユーザブラウザ / iOS Safari]
        │
        ▼
[Amplify Hosting (Unit A)]                    ┌─────────────────────────────┐
  Next.js App Router (BFF パターン)           │  Frontend Components (Unit C)│
  ├─ /  : MainScreen                          │  ├─ GoroButton (LC-ORDER-21)│
  ├─ /order/[id]/complete : OrderComplete    │  ├─ OrderHistoryList (24)   │
  ├─ /budget-empty : BudgetEmpty (Unit B)    │  ├─ ToastHost (30) ★全画面  │
  └─ /api/[...path] : BFF Route Handler      │  └─ OrderCompletionScreen(32)│
        │ Authorization: Bearer (透過)        └─────────────────────────────┘
        ▼
[API Gateway HTTP API (Unit A)]
  ├─ JWT Authorizer (Cognito)
  ├─ Stage Throttling: 100 req/s (Unit A 確定)
  ├─ POST /api/auth/logout (Unit A)
  ├─ POST /api/wallet/* (Unit B)
  ├─ POST /api/orders   ★ Unit C 追加 (LC-ORDER-22 useOrder)
  └─ GET  /api/orders   ★ Unit C 追加 (LC-ORDER-23 useOrderHistory)
        │
        ▼
[API Lambda (Unit A、横串)]
  Function: gp-{env}-api-fn
  Memory: 256MB / Arch: arm64 / Timeout: 10s (NFRC-C18)
  Runtime: provided.al2023 (Go + Gin + LWA)
  ├─ AttachUserID middleware (Unit A LC-AUTH-XX)
  ├─ LatencyMiddleware (LC-ORDER-12) ★ E2E レイテンシ計測
  ├─ ContextAwareSlogHandler (LC-AUTH-05) ★ traceId/requestId/userId 自動付与
  └─ OrderHandler (LC-ORDER-02)
        │
        │   ┌──────────────────────────────────────┐
        │   │  init() で SDK 初期化 (P-INIT-01)     │
        │   │   - bedrockruntime.Client (LC-14)    │
        │   │   - dynamodb.Client      (LC-15)     │
        │   └──────────────────────────────────────┘
        ▼
[OrderService (LC-01)] ──── 呼出 ──── [WalletService (Unit B 既存)]
        │
        ├─→ [PlanBuilder (LC-08)] ─── [BedrockAdapter (LC-05)] ─── Amazon Bedrock
        │       │                                                  └ Inference Profile
        │       │                                                    apac.anthropic.claude-3-5-haiku-*
        │       │
        │       └─→ [FallbackSuggestProvider (LC-09)]
        │
        ├─→ [DeliveryAdapter (LC-06)] (MockDeliveryAdapter)
        │
        └─→ [OrderHistoryRepository (LC-04)] ──── DynamoDB
                                                  gp-{env}-order-history
                                                  ├ PK = USER#{userID}
                                                  ├ SK = ORDER#{orderedAt}#{orderID}
                                                  └ TTL = expiresAt (90 日)

[全 PlaceOrder で並走]
  LogSummary (LC-10) ──defer LogComplete──┐
  EventLogger (LC-11) ──slog.Warn──────┐  │
  Measure (LC-13) ──slog.Info─────────┐│  │
                                      ▼▼  ▼
                            [CloudWatch Logs (Unit A 既存 Log Group)]
                                      │
                                      ▼
                  ┌───────────────────────────────────┐
                  │ Metric Filter ×3 (NFRC-C13)       │
                  │  - place_order_latency             │
                  │  - bedrock_retry                   │
                  │  - fallback_triggered              │
                  └───────────────────────────────────┘
                                      │
                                      ▼
                  ┌───────────────────────────────────┐
                  │ CloudWatch Alarm ×3               │
                  │  - p95 > 3000ms (3 connectives)   │
                  │  - retry ≥ 5 / 5 min               │
                  │  - fallback ≥ 3 / 5 min            │
                  └───────────────────────────────────┘
                                      │
                                      ▼
                  ┌───────────────────────────────────┐
                  │ AWS Budgets                       │
                  │  Bedrock $5 / month                │
                  │  - 80% threshold (warning)         │
                  │  - 100% threshold (alert)          │
                  └───────────────────────────────────┘
                                      │
                                      ▼
                          [SNS Topic gp-{env}-alarms]
                                      │
                                      ▼
                                  [email]
                                  alerts@example.com
```

### 凡例

- 実線 ─→: リクエスト / 同期呼出
- 二重線 ══: 非同期通知（SNS / Email）
- 点線 ┄┄: 既存リソース（Unit A / B 構築済み）
- 太字: Unit C で新規構築するコンポーネント
- ★: 本書のスコープ

---

## 2. リクエストフロー詳細

### 2.1 通常パス: `POST /api/orders` 成功フロー（NFRC-C01 通常パス p95 2.5s）

```
1. [Browser] GoroButton 押下
       │
       ▼
2. [Browser] useOrder.mutate(req) (LC-22)
       │ apiClient.placeOrder(req) (LC-33)
       │ POST /api/orders
       │ Authorization: Bearer <accessToken>
       │ Body: { category: "food", idempotencyKey: <ULID>, suggestionId? }
       ▼
3. [Next.js Server (BFF)] catch-all Route Handler (LC-AUTH-18)
       │ → API Gateway へ透過プロキシ
       ▼
4. [API Gateway] JWT Authorizer (Cognito) で accessToken 検証
       │ claims.sub → context.userID
       ▼
5. [API Lambda] LatencyMiddleware (LC-12) start = time.Now()
       │ AttachUserID → ctx に userID 注入
       ▼
6. [OrderHandler.PlaceOrder (LC-02)] req パース
       │ ↓
7. [OrderService.PlaceOrder (LC-01)]
       │ summary := NewLogSummary(ctx)
       │ defer summary.LogComplete()
       │
       ├─ a. [PlanBuilder.Build (LC-08)] Bedrock 呼出
       │     │ measure(ctx, "bedrock", ...)
       │     │ ├─ history := historyRepo.Query(ctx, userID, 30)  → DynamoDB
       │     │ ├─ bedrockAdapter.InferOrderPlan(history) → Bedrock
       │     │ │   timeout: 1.5s (NFRC-C07)
       │     │ │   retry: max 1 (NFRC-C06、RetryClassifier 経由)
       │     │ └─ summary.SetBedrockLatencyMs(ms) / SetBedrockAttempt(n) / SetFallbackTriggered(false)
       │     ▼
       │     plan := { Store, Menu, Amount, Category, Source: "bedrock" }
       │
       ├─ b. [WalletService.Deduct (Unit B)] 残高引き落とし
       │     │ measure(ctx, "wallet", ...)
       │     │ TryAcquire idempotencyKey → DeductConditional
       │     ▼
       │     deductResult := { OrderID, RemainingBalance, Idempotent: false }
       │
       ├─ c. [DeliveryAdapter.Place (LC-06)] モック手配
       │     │ measure(ctx, "delivery", ...)
       │     │ MockDeliveryAdapter は noop で nil を返す（log.Info 出力のみ）
       │     ▼
       │     (success)
       │
       ├─ d. [OrderHistoryRepository.Insert (LC-04)] 履歴保存
       │     │ measure(ctx, "history", ...)
       │     │ PutItem PK=USER#{userID} SK=ORDER#{orderedAt}#{orderID} expiresAt=now+90d
       │     │ 失敗時も 200 応答（NFRC-C09、historyInsertFailed=true をログ付与）
       │     ▼
       │     (success)
       │
       └─ summary.SetOrderId / SetCategory / SetAmount / SetStoreName / SetMenuName / SetHistoryCount / SetSource
          ▼
       result := PlaceOrderResult { OrderID, StoreName, MenuName, Amount, RemainingBalance, Idempotent: false }
       ▼
8. [API Lambda] LatencyMiddleware end → slog.Info "request_complete" latencyMs=2150ms
       │ defer summary.LogComplete() → slog.Info "place_order_complete" + 11 項目
       ▼
9. [API Gateway] 200 + JSON
       ▼
10. [Next.js Server] BFF レスポンス透過
       ▼
11. [Browser] useOrder.onSuccess
       │ triggerLock(1000ms) (LC-25)
       │ queryClient.invalidateQueries({ queryKey: ['orderHistory'] })
       │ queryClient.invalidateQueries({ queryKey: ['balance'] }) (Unit B)
       │ router.push(`/order/${result.OrderID}/complete`)
       ▼
12. [Browser] OrderCompletionScreen (LC-32) 表示
       │ useEffect で setTimeout(5000) → router.push('/')
       ▼
13. [Browser] MainScreen に戻る
       useOrderHistory 再 fetch（placeholderData で前回値表示維持、LC-23 / P-FE-LOAD-01）
```

**所要時間目標** (NFRC-C01 通常パス p95 ≤ 2.5s):
- Bedrock: 1500ms（NFRC-C07 タイムアウト）
- Wallet: 500ms（Unit B p95、NFRC-C01 内訳）
- Delivery: 50ms（Mock）
- History: 50ms（DynamoDB Put）
- LambdaOverhead + Network/UI: 400ms
- 合計: 2500ms（p95 達成）

### 2.2 リトライ後成功フロー（NFRC-C01 リトライ発動パス p95 ≤ 4.0s）

7-a で Bedrock が ThrottlingException を返した場合:

```
7-a-1. RetryClassifier.ShouldRetry(err) → true
7-a-2. EventLogger.LogBedrockRetry(ctx, attempt=2, errorClass="ThrottlingException", elapsedMs=1500)
       slog.Warn "bedrock_retry" → CloudWatch Logs
7-a-3. bedrockAdapter.InferOrderPlan(history) を再実行（attempt 2、もう 1500ms）
7-a-4. 成功 → plan := { ...Source: "bedrock" }
       summary.SetBedrockAttempt(2) / SetBedrockLatencyMs(2500ms 累計)
       以降フローは 2.1 と同じ
```

**Metric Filter 検知**: `bedrock_retry` イベントが metric filter で抽出 → `BedrockRetryCount` メトリクス +1
**Alarm 発動条件**: 5 分間に ≥ 5 回 → SNS 通知

### 2.3 フォールバック発動フロー（NFRC-C01 フォールバック発動パス p95 ≤ 2.5s）

7-a で Bedrock が 2 連続失敗した場合:

```
7-a-1. RetryClassifier.ShouldRetry(err) → true
7-a-2. EventLogger.LogBedrockRetry → "bedrock_retry"
7-a-3. bedrockAdapter.InferOrderPlan → 再失敗
7-a-4. PlanBuilder: 履歴件数による分岐
       │ if historyCount >= 5 → FallbackSuggestProvider.BuildFromHistory(history)
       │ else                 → FallbackSuggestProvider.Default()
7-a-5. EventLogger.LogFallbackTriggered(ctx, reason="bedrock_double_failure", historyCount, fallbackType)
       slog.Warn "fallback_triggered" → CloudWatch Logs
7-a-6. plan := { ...Source: "fallback_history" or "fallback_default" }
       summary.SetFallbackTriggered(true) / SetSource("fallback_*")
       以降フローは 2.1 と同じ
```

**Metric Filter 検知**: `fallback_triggered` イベントが metric filter で抽出 → `FallbackTriggeredCount` メトリクス +1
**Alarm 発動条件**: 5 分間に ≥ 3 回 → SNS 通知

### 2.4 残高不足エラーフロー（NFRC-C22）

```
7-b で WalletService.Deduct が ErrInsufficientFunds を返した場合:

8'. OrderHandler エラーマッピング (BR-C39 / LC-02)
    wallet.ErrInsufficientFunds → 402 INSUFFICIENT_FUNDS
9'. API Gateway 402
10'. Browser useOrder.onError
     mapOrderError(err) (LC-26 P-FE-ERR-01)
     → { type: 'navigate', path: '/budget-empty', transitionMs: 0 }
11'. router.push('/budget-empty') (Unit B BudgetEmptyScreen)
     triggerLock(1000ms) (LC-25)
```

### 2.5 冪等命中フロー（NFRC-C05）

```
7-b で WalletService.Deduct が冪等命中した場合:
deductResult := { OrderID: <既存>, RemainingBalance: <変化なし>, Idempotent: true }

7-c-d は実行スキップ（Plan / Adapter / History は再実行しない）

OrderHistoryRepository.GetItem(orderID) で 1 回目の OrderRecord を取得
result := PlaceOrderResult { ...同じ payload, Idempotent: true }
summary.SetIdempotent(true) / SetOrderId(既存)
```

NFRC-C15 PBT P-3 が「2 回目以降のレスポンス payload 完全一致」を保証する。

### 2.6 サーバエラー → 自虐トーストフロー（NFRC-C22）

```
8''. OrderHandler エラーマッピング (LC-02)
     その他エラー → 500 INTERNAL_ERROR
9''. API Gateway 500
10''. Browser useOrder.onError
      mapOrderError(err) (LC-26)
      → { type: 'toast', text: getRandomToast() (LC-27), durationMs: 5000 }
11''. useToast.showToast(text, 5000) (LC-29)
      → toastsAtom 追加 → ToastHost 表示 (LC-30)
      triggerLock(1000ms) (LC-25)
```

---

## 3. 観測性データフロー

### 3.1 ログ → メトリクス → アラーム → 通知

```
[API Lambda]
  ├─ LatencyMiddleware → slog.Info "request_complete" {latencyMs, statusCode, path}
  ├─ Measure → slog.Info "<name>_complete" {latencyMs, success}
  ├─ EventLogger → slog.Warn "bedrock_retry" / "fallback_triggered"
  └─ LogSummary → slog.Info "place_order_complete" {19 項目}
        │
        ▼ JSON 構造化ログ
[CloudWatch Logs]
  Log Group: /aws/lambda/gp-{env}-api-fn (Unit A 既存)
  Stream: <Lambda invocation ID>
        │
        ▼ Metric Filter（pattern マッチ）
[CloudWatch Metrics]
  Namespace: GoroPay/Order
  Metrics:
    - PlaceOrderLatencyMs (Milliseconds)
    - BedrockRetryCount (Count)
    - FallbackTriggeredCount (Count)
        │
        ▼ Alarm 評価（5 min × 1〜3 datapoints）
[CloudWatch Alarms]
  - place_order_p95_breach (p95 > 3000ms × 3)
  - bedrock_retry_burst (≥ 5 / 5 min)
  - fallback_triggered_burst (≥ 3 / 5 min)
        │
        ▼ Alarm Action（ALARM 状態遷移）
[SNS Topic gp-{env}-alarms]
        │
        ▼ Subscription
[email: alerts@example.com]
```

### 3.2 コスト監視データフロー

```
[Amazon Bedrock]
  InvokeModel / Converse 呼出
  → Service Charge: Bedrock
        │
        ▼
[AWS Cost Explorer]
  Service フィルタ: Amazon Bedrock
        │
        ▼ 日次集計
[AWS Budgets]
  gp-{env}-bedrock-budget
  Limit: $5 / month
  Cost Filter: Service = Amazon Bedrock
        │
        ▼ Threshold 検知
  - 80% > $4 → notification (ACTUAL)
  - 100% > $5 → notification (ACTUAL)
        │
        ▼ SNS Topic 通知
[SNS Topic gp-{env}-alarms]
        │
        ▼
[email: alerts@example.com]
```

---

## 4. リソース配置とライフサイクル

### 4.1 Lambda INIT フェーズ（Cold Start）

```
Container 起動
  ├─ Go バイナリロード (約 200ms)
  └─ package init() 実行
      ├─ bedrockruntime.NewFromConfig() (約 250ms)
      └─ dynamodb.NewFromConfig() (約 150ms)
  → Total INIT: 約 600ms（NFRC-C18 目標 P95 600ms 以内）
```

### 4.2 Lambda INVOKE フェーズ（Warm）

```
Handler 実行
  ├─ 既に init 済み → SDK Client 再利用
  ├─ Bedrock 呼出 (1500ms タイムアウト)
  ├─ DynamoDB 呼出 (Wallet / Insert / Query 各 50〜500ms)
  └─ レスポンス組み立て
  → Total INVOKE: 通常パス 約 2100ms（NFRC-C01 内訳通り）
```

---

## 5. デプロイ手順詳細

### 5.1 初回 dev 環境構築（Unit C 追加）

| 順序 | 操作 | 場所 |
|---|---|---|
| 1 | `infra/envs/dev/locals.tf` の `alarm_email` を実メールアドレスに変更 | ローカル |
| 2 | `cd infra/envs/dev && terraform init` | ローカル |
| 3 | `terraform plan` で変更内容確認（既存リソース変更なし、3 module 新規 + 既存 2 module 拡張） | ローカル |
| 4 | `terraform apply` 実行 | ローカル |
| 5 | SNS 購読確認メール（5 分以内）の `Confirm subscription` リンクをクリック | メール |
| 6 | AWS Console → Bedrock → Model access で `Anthropic Claude 3.5 Haiku` を有効化（未承認なら申請） | AWS Console |
| 7 | `terraform output order_history_table_name` で `gp-dev-order-history` を確認 | ローカル |
| 8 | Code Generation 完了後、git push → CodePipeline 起動で API Lambda の最新 image をデプロイ | GitHub |
| 9 | Frontend は Amplify Hosting で自動デプロイ（develop ブランチ push） | GitHub |

### 5.2 ハッカソン実演時のチェック項目

| チェック項目 | 確認方法 |
|---|---|
| API Lambda が Bedrock を呼べる | CloudWatch Logs で `bedrock_complete latencyMs=...` ログ確認 |
| OrderHistory に Insert 成功 | DynamoDB Console で `gp-dev-order-history` テーブル確認 |
| アラームが正常に設定されている | CloudWatch Alarms Console で 3 つの ALARM 状態（INSUFFICIENT_DATA or OK）を確認 |
| SNS 購読が confirmed 状態 | SNS Console で Subscription の `Status = Confirmed` 確認 |
| Bedrock コストアラートが動作可能 | Budgets Console で `gp-dev-bedrock-budget` の `Status = Active` 確認 |
| Frontend からの注文成功 | Browser → MainScreen → 「ご飯めんどくさい」押下 → OrderCompletionScreen 表示 → MainScreen へ自動遷移確認 |

### 5.3 prd 環境への横展開（将来）

`infra/envs/prd/main.tf` を `infra/envs/dev/main.tf` の構造を踏襲して作成:

- `local.env = "prd"` に変更
- `local.alarm_email = "alerts-prd@example.com"` 等の環境別アドレス
- DynamoDB PITR 有効化検討（infrastructure-design.md §10.2）
- Budgets limit を本番想定値に引き上げ（例: $50 / month）
- アラーム閾値を緩める（`p95_threshold_ms = 5000` 等の variables 上書き検討、Q-I4 = C で外部化済みのため tfvars で対応可能）

---

## 6. リスクと緩和策

| リスク | 影響 | 緩和策 |
|---|---|---|
| Bedrock スロットリング多発 | フォールバック乱発 → ビジネス価値毀損 | NFRC-C13-2 アラーム + AWS Budgets 早期検知、本番化時に Provisioned Throughput 検討 |
| DynamoDB バースト容量超過 | OrderHistory Insert 失敗 | NFRC-C09 で 200 応答（フェイルオープン）+ historyInsertFailed=true ログで運用検知 |
| Lambda コールドスタート長期化 | NFRC-C18 違反 | arm64 + 256MB + init で SDK 初期化（P-INIT-01）、本番化時に Provisioned Concurrency 検討 |
| アラームメールスパム | オペレータの認知疲労 | しきい値を tfvars で外部化（Q-I4 = C）、誤報多発時は dev で緩める |
| Budgets メール遅延 | コスト超過認知遅延 | 80% で警告（Q-I7 = A）、SNS は通常 1〜2 分以内に到達 |

---

## 7. Code Generation への引き継ぎ

Code Generation ステージで実装すべき項目（NFR Design 由来）:

| 引き継ぎ項目 | 関連 LC | 由来 |
|---|---|---|
| Bedrock SDK init() / package-level client | LC-14 | P-INIT-01 |
| DynamoDB SDK init() / package-level client | LC-15 | P-INIT-01 |
| OrderService 実装（オーケストレーション） | LC-01 | NFR Design |
| OrderHandler 実装（Gin handler、エラーマッピング） | LC-02 | NFR Design |
| OrderHistoryRepository 実装 | LC-04 | NFR Design |
| BedrockAdapter / RetryClassifier 実装 | LC-05 / LC-07 | P-RETRY-01 |
| PlanBuilder 実装（リトライ + フォールバック分岐） | LC-08 | P-PLAN-01 |
| FallbackSuggestProvider 実装（5 店舗ランダム選択） | LC-09 | BR-C07 |
| LatencyMiddleware / Measure 実装 | LC-12 / LC-13 | P-OBS-01 |
| LogSummary / EventLogger 実装 | LC-10 / LC-11 | P-OBS-02 / P-OBS-03 |
| MockBedrockAdapter / WalletStub / InmemoryHistory 実装 | LC-16 / LC-19 / LC-20 | P-MOCK-01 / P-PBT-01 |
| PBT テストコード（gopter、P-1 + P-3） | — | P-PBT-01 |
| Frontend GoroButton / OrderCompletionScreen / OrderHistoryList / Toast / ToastHost | LC-21 / LC-32 / LC-24 / LC-31 / LC-30 | NFR Design |
| useOrder / useOrderHistory / useDisableLock / useToast | LC-22 / LC-23 / LC-25 / LC-29 | NFR Design |
| errorMappers / getRandomToast | LC-26 / LC-27 | P-FE-ERR-01 / P-FE-TOAST-01 |
| toastsAtom / lib/api/orders.ts / lib/ulid.ts | LC-28 / LC-33 / LC-34 | NFR Design |
| Bedrock プロンプトテンプレート | — | FD `business-logic-model.md` |
| 自虐トースト最終文言（3 種） | — | NFRC-C22 / Q-D10 |

**環境変数注入**:
- `ORDER_HISTORY_TABLE_NAME` → `gp-dev-order-history` （Terraform output 経由）
- `BEDROCK_INFERENCE_PROFILE_ID` → `apac.anthropic.claude-3-5-haiku-20241022-v1:0`（Terraform 直書き）

**Test 戦略**:
- local / CI: mock 必須（NFRC-C16）
- dev: Bedrock 実呼出し許容（NFRC-C16）
- terraform-test: 3 ファイル（Q-I13 = C）

---

## 8. 文書管理

- **承認**: ユーザ承認待ち（Construction フェーズの per-unit ループ承認ゲート）
- **凍結契約への影響**: なし
- **次ステージ**: Code Generation（per-unit、Comprehensive 深度）
