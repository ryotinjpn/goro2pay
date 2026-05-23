# Budget Unit — Deployment Architecture

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Stage**: Infrastructure Design / Construction
**Predecessors**: [infrastructure-design.md](./infrastructure-design.md)

本ドキュメントは Unit B の **デプロイメント時の AWS リソース構成と相互関係、リクエストフロー** を記述する。

---

## 1. AWS リソース全体図

```
                       ┌───────────────────────────── ap-northeast-1 (Tokyo) ──────────────────┐
                       │                                                                         │
                       │  ┌──────────────── ランタイム ─────────────────────────────────────┐   │
                       │  │                                                                  │   │
   ┌────────────┐      │  │ ┌──────────────────────────────────────────────────────────────┐│   │
   │ ブラウザ    │──1──►│  │ │ Amplify Hosting (Unit A 構築済み)                            ││   │
   │ (太郎)     │      │  │ │  Next.js App Router (SSR)                                    ││   │
   └────────────┘      │  │ │  catch-all /api/[...path] → API Gateway (BFF)                ││   │
                       │  │ └───────────────────────────────┬──────────────────────────────┘│   │
                       │  │                                  │ 2: /api/wallet/*              │   │
                       │  │                                  ▼ Bearer <accessToken>          │   │
                       │  │ ┌──────────────────────────────────────────────────────────────┐│   │
                       │  │ │ API Gateway HTTP API (gp-dev-api) — Unit A 構築済み          ││   │
                       │  │ │  JWT Authorizer (Cognito)                                     ││   │
                       │  │ │  ┌─ GET  /api/wallet          ─┐                             ││   │
                       │  │ │  └─ POST /api/wallet/budget    ─┘ ← Unit B が追加            ││   │
                       │  │ └───────────────────────────────┬──────────────────────────────┘│   │
                       │  │                                  │ AWS_PROXY                     │   │
                       │  │                                  ▼                               │   │
                       │  │ ┌──────────────────────────────────────────────────────────────┐│   │
                       │  │ │ API Lambda (gp-dev-api-fn) — Unit A 構築済み                 ││   │
                       │  │ │  Go + Gin + LWA, arm64, 512MB                                ││   │
                       │  │ │  ENV: DDB_TABLE_WALLET / DDB_TABLE_BUDGET_SETTINGS           ││   │
                       │  │ │       DDB_TABLE_IDEMPOTENCY / DDB_TABLE_BUDGET_RESET_LOG     ││   │
                       │  │ │                                                              ││   │
                       │  │ │  WalletHandler (LC-BUDGET-01)                                ││   │
                       │  │ │    └► WalletService (LC-BUDGET-02)                           ││   │
                       │  │ │         ├► WalletRepository         ─────────────────────►  ││   │
                       │  │ │         ├► BudgetSettingsRepository  ──────────────────────► ││   │
                       │  │ │         ├► IdempotencyRepository     ─────────────────────►  ││   │
                       │  │ │         └► BudgetResetLogRepository  ──────────────────────► ││   │
                       │  │ └──────────────────────────────────────┬───────────────────────┘│   │
                       │  │                                         │                        │   │
                       │  │                  ┌──────────────────────┴───────────────────┐   │   │
                       │  │                  │          DynamoDB (Unit B 所有)           │   │   │
                       │  │                  │  ┌─────────────────────────────────────┐ │   │   │
                       │  │                  │  │ gp-dev-wallet (PK: userId)           │ │   │   │
                       │  │                  │  │  Provisioned 1R/1W, ConsistentRead  │ │   │   │
                       │  │                  │  ├─────────────────────────────────────┤ │   │   │
                       │  │                  │  │ gp-dev-budget-settings (PK: userId)  │ │   │   │
                       │  │                  │  │  Provisioned 1R/1W                  │ │   │   │
                       │  │                  │  ├─────────────────────────────────────┤ │   │   │
                       │  │                  │  │ gp-dev-idempotency-keys (PK: key)    │ │   │   │
                       │  │                  │  │  Provisioned 1R/1W, TTL: expiresAt  │ │   │   │
                       │  │                  │  ├─────────────────────────────────────┤ │   │   │
                       │  │                  │  │ gp-dev-budget-reset-log              │ │   │   │
                       │  │                  │  │  PK: resetDate / SK: userId          │ │   │   │
                       │  │                  │  │  Provisioned 1R/1W                  │ │   │   │
                       │  │                  │  └─────────────────────────────────────┘ │   │   │
                       │  │                  └──────────────────────────────────────────┘   │   │
                       │  │                                                                  │   │
                       │  │ ┌──────────── 月次リセットフロー ─────────────────────────────┐ │   │
                       │  │ │                                                               │ │   │
                       │  │ │  EventBridge Scheduler (gp-dev-monthly-reset)                │ │   │
                       │  │ │  cron(0 15 L * ? *) UTC = 月末最終日 0:00 JST               │ │   │
                       │  │ │                    │                                          │ │   │
                       │  │ │                    ▼                                          │ │   │
                       │  │ │  Scheduler Lambda (gp-dev-scheduler-fn)                      │ │   │
                       │  │ │  Go (zip/bootstrap), arm64, 128MB, 30s timeout               │ │   │
                       │  │ │    └► WalletService.ResetAll()                                │ │   │
                       │  │ │         ├► gp-dev-wallet (Scan → UpdateItem × N)             │ │   │
                       │  │ │         ├► gp-dev-budget-settings (GetItem × N)              │ │   │
                       │  │ │         └► gp-dev-budget-reset-log (PutItem × N)             │ │   │
                       │  │ │                    │                                          │ │   │
                       │  │ │                    ▼                                          │ │   │
                       │  │ │  CloudWatch Logs: /aws/lambda/gp-dev-scheduler-fn            │ │   │
                       │  │ │  ERROR per user + INFO summary (P-OBS-02)                    │ │   │
                       │  │ │  失敗時: 再試行なし、手動リカバリ (Q-I5=A, NFR-REL-04)      │ │   │
                       │  │ └───────────────────────────────────────────────────────────────┘ │   │
                       │  │                                                                  │   │
                       │  │ ┌──────────────────────────────────────────────────────────────┐│   │
                       │  │ │ CloudWatch Log Groups (retention 7日)                        ││   │
                       │  │ │  /aws/lambda/gp-dev-api-fn     (Unit A 構築済み)             ││   │
                       │  │ │  /aws/lambda/gp-dev-scheduler-fn (Unit B 追加)               ││   │
                       │  │ └──────────────────────────────────────────────────────────────┘│   │
                       │  └──────────────────────────────────────────────────────────────────┘   │
                       └─────────────────────────────────────────────────────────────────────────┘
```

---

## 2. リクエストフロー

### 2.1 `GET /api/wallet`（残高取得）

```
Browser
  → apiClient.request({ path: '/api/wallet' })      # LC-AUTH-09, Bearer <accessToken>
  → Next.js BffProxyRouteHandler (LC-AUTH-18)       # server-only: API_ENDPOINT
  → API Gateway: GET /api/wallet (JWT Authorizer)
  → API Lambda: WalletHandler.GetBalance()
      → WalletService.GetBalance(ctx, userID)
          → WalletRepository.Get(ConsistentRead: true)  # gp-dev-wallet
  ← 200 { balance, monthlyBudget, updatedAt }
```

### 2.2 `POST /api/wallet/budget`（予算設定）

```
Browser
  → apiClient.request({ path: '/api/wallet/budget', method: 'POST', body })
  → Next.js BffProxyRouteHandler
  → API Gateway: POST /api/wallet/budget (JWT Authorizer)
  → API Lambda: WalletHandler.SetBudget()
      → WalletService.SetBudget(ctx, userID, monthlyBudget)
          1. BudgetSettingsRepository.Upsert()     # gp-dev-budget-settings
          2. WalletRepository.SetBalance()         # gp-dev-wallet
  ← 200 { monthlyBudget, appliedFrom }
```

### 2.3 `Deduct`（残高減算 — Unit C 内部呼び出し）

```
API Lambda: OrderHandler (Unit C)
  → OrderService.PlaceOrder()
      → WalletService.Deduct(ctx, userID, amount, idempotencyKey)
          1. IdempotencyRepository.TryAcquire()   # gp-dev-idempotency-keys (TTL 24h)
          2. WalletRepository.Deduct()            # gp-dev-wallet (ConditionExpression: balance >= :amount)
               ↳ ConditionalCheckFailed → ErrInsufficientBalance → HTTP 402
          3. IdempotencyRepository.SaveResponse()
  ← { newBalance, idempotent }
```

> `Deduct` は HTTP ルートを持たない。Unit C と Unit B は同一 Lambda 内で Go 関数として連携する。

### 2.4 月次リセット（EventBridge Scheduler → Scheduler Lambda）

```
EventBridge Scheduler
  cron(0 15 L * ? *) UTC = 月末最終日 0:00 JST
  → Scheduler Lambda: ResetSchedulerHandler (LC-BUDGET-07)
      → WalletService.ResetAll(ctx)
          → gp-dev-wallet: Scan 全ユーザ
          → 各ユーザ:
              1. BudgetSettingsRepository.Get()      # gp-dev-budget-settings
              2. WalletRepository.SetBalance()       # gp-dev-wallet
              3. BudgetResetLogRepository.Save()     # gp-dev-budget-reset-log (PutItem + Condition)
          → 失敗ユーザ: ERROR ログ (P-OBS-02)
      → 集計サマリ: INFO ログ { processedUsers, failedUsers }
```

---

## 3. Terraform モジュール構成とデプロイ手順

### 3.1 初回デプロイ手順

```bash
# 1. Scheduler Lambda バイナリをビルド
make -C apps/scheduler build
# → apps/scheduler/bootstrap (linux/arm64 バイナリ) が生成される

# 2. Terraform 実行（infra/envs/dev/ から）
cd infra/envs/dev
terraform init
terraform plan
terraform apply
```

### 3.2 Scheduler Lambda 更新時（Q-I3=A: 手動デプロイ）

```bash
# コード変更後
make -C apps/scheduler build
cd infra/envs/dev
terraform apply -target=module.budget.aws_lambda_function.scheduler
```

### 3.3 モジュール依存関係

```
envs/dev/main.tf
  └── module "auth"     (infra/modules/auth/)    # 先に apply 済みが前提
  └── module "budget"   (infra/modules/budget/)  # auth outputs を参照
        ├── var.api_id                 ← module.auth.api_id
        ├── var.api_lambda_invoke_arn  ← module.auth.api_lambda_invoke_arn
        ├── var.api_lambda_role_arn    ← module.auth.api_lambda_role_arn
        └── var.cognito_authorizer_id  ← module.auth.cognito_authorizer_id
```

---

## 4. 環境変数まとめ（Unit B 追加分）

| 変数名 | 設定先 Lambda | 値 | 出典 |
|---|---|---|---|
| `DDB_TABLE_WALLET` | API Lambda / Scheduler Lambda | `aws_dynamodb_table.wallet.name` | unit-interfaces §10 |
| `DDB_TABLE_BUDGET_SETTINGS` | API Lambda / Scheduler Lambda | `aws_dynamodb_table.budget_settings.name` | 同上 |
| `DDB_TABLE_IDEMPOTENCY` | API Lambda | `aws_dynamodb_table.idempotency_keys.name` | 同上 |
| `DDB_TABLE_BUDGET_RESET_LOG` | Scheduler Lambda | `aws_dynamodb_table.budget_reset_log.name` | 同上 |

---

## 5. 後続 Unit への影響

| 後続 Unit | 内容 |
|---|---|
| **Unit C** | `PlaceOrder` 内で `WalletService.Deduct()` を呼ぶ。Unit B の DI が必要（Code Generation で統合） |
| **Unit E** | `module.budget.wallet_table_arn` / `module.budget.budget_settings_table_arn` を参照し、Unit E の IAM ポリシーに Read 権限を追加 |
