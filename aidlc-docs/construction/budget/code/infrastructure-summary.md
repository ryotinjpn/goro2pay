# Infrastructure Summary — Unit B (`budget`)

## 新規 Module: `infra/modules/budget/`

| ファイル | 役割 |
|---|---|
| `versions.tf` | terraform >= 1.10.0 / aws ~> 6.46 / archive ~> 2.4 |
| `variables.tf` | env, region |
| `locals.tf` | name_prefix `gp-${var.env}` + 全リソース名 |
| `dynamodb.tf` | 4 テーブル (wallet / budget_settings / idempotency_keys + TTL / budget_reset_log + composite key) |
| `iam.tf` | scheduler_lambda role (Wallet GUS + BudgetSettings G + BudgetResetLog PG + Logs) / eventbridge_scheduler role (lambda:InvokeFunction only) / API Lambda 用 dynamodb_access policy (4 テーブル CRUD) |
| `log_groups.tf` | /aws/lambda/gp-{env}-scheduler-fn (retention 7 日) |
| `scheduler_lambda.tf` | archive_file (bootstrap) + Lambda (provided.al2023, arm64, 128MB, 30s) + permission + Schedule (cron(0 15 L * ? *) UTC) |
| `outputs.tf` | 4 テーブルの name/arn + scheduler arn + dynamodb_policy_arn |
| `README.md` | モジュール概要 + デプロイ前提 |
| `tests/budget_basic.tftest.hcl` | plan 成立 |
| `tests/budget_dynamodb_tables.tftest.hcl` | 4 テーブルの billing/hash/TTL/encryption/PITR 検証 |
| `tests/budget_scheduler_lambda.tftest.hcl` | runtime/handler/arch/memory/timeout/cron/retry 検証 |
| `tests/budget_iam_least_privilege.tftest.hcl` | role naming / assume_role_policy 検証 |

## DynamoDB テーブル一覧

| テーブル | PK | SK | TTL | 用途 |
|---|---|---|---|---|
| `gp-dev-wallet` | `userId` | - | - | 残高 (PROVISIONED 1/1, AES256, PITR off) |
| `gp-dev-budget-settings` | `userId` | - | - | 月間予算 (PROVISIONED 1/1) |
| `gp-dev-idempotency-keys` | `key` | - | `expiresAt` (24h) | Deduct 冪等性 |
| `gp-dev-budget-reset-log` | `resetDate` | `userId` | - | 月初リセット履歴 (composite key で冪等性確保) |

## Scheduler Lambda

| 項目 | 値 |
|---|---|
| function_name | `gp-dev-scheduler-fn` |
| runtime | `provided.al2023` (Go custom) |
| handler | `bootstrap` |
| architecture | arm64 |
| memory | 128 MB (Q-N9=A) |
| timeout | 30s (Q-N3=A) |
| package | data.archive_file (zip) — `apps/api/cmd/scheduler/bootstrap` を要事前ビルド |
| schedule | `cron(0 15 L * ? *)` UTC = 月末最終日 0:00 JST (Q-B8=A) |
| retry | 0 (Q-I5=A: ERROR ログのみ) |

## IAM (最小権限原則)

### Scheduler Lambda Role (`gp-dev-scheduler-role`)

| Statement | Resource | Actions |
|---|---|---|
| WalletReadWrite | wallet table | GetItem, UpdateItem, Scan |
| BudgetSettingsRead | budget_settings table | GetItem |
| BudgetResetLogWrite | budget_reset_log table | PutItem, GetItem |
| Logs | scheduler log group | CreateLogGroup, CreateLogStream, PutLogEvents |

### EventBridge Scheduler Role (`gp-dev-eventbridge-scheduler-role`)

| Statement | Resource | Actions |
|---|---|---|
| InvokeSchedulerLambda | scheduler lambda arn | lambda:InvokeFunction |

### API Lambda 用 Policy (`gp-dev-budget-dynamodb-policy`)

`outputs.dynamodb_policy_arn` で公開し、`envs/dev/main.tf` で `module.lambda_api.additional_policy_arns` に渡して attach。Unit C `order_history` と同 pattern。

| Statement | Resource | Actions |
|---|---|---|
| BudgetTablesCRUD | 4 テーブル ARN 全部 | GetItem, PutItem, UpdateItem, DeleteItem, Query, Scan |

## 既存 Module への変更

### `infra/modules/api_gateway/routes.tf` (追記)

Unit B 用ルート 2 本を追加 (既存 `aws_apigatewayv2_integration.api_lambda` を再利用):
- `aws_apigatewayv2_route.get_wallet` — `GET /api/wallet` (JWT)
- `aws_apigatewayv2_route.post_wallet_budget` — `POST /api/wallet/budget` (JWT)

### `infra/modules/lambda_api/variables.tf` (追記)

Unit B 用 4 つの変数を追加 (空文字列なら env 注入 skip):
- `wallet_table_name` → `DDB_TABLE_WALLET`
- `budget_settings_table_name` → `DDB_TABLE_BUDGET_SETTINGS`
- `idempotency_keys_table_name` → `DDB_TABLE_IDEMPOTENCY`
- `budget_reset_log_table_name` → `DDB_TABLE_BUDGET_RESET_LOG`

### `infra/modules/lambda_api/api_lambda.tf` (追記)

`environment.variables` の merge ブロックに 4 つの conditional map を追加 (Unit C `order_history_table_name` と同 pattern)。

### `infra/envs/dev/main.tf` (追記)

```hcl
module "budget" {
  source = "../../modules/budget"
  env    = local.env
  region = local.region
}

# lambda_api への注入
module "lambda_api" {
  ...
  additional_policy_arns = [
    module.order_history.dynamodb_policy_arn,
    module.bedrock.bedrock_policy_arn,
    module.budget.dynamodb_policy_arn,           # ← 追加
  ]
  wallet_table_name           = module.budget.wallet_table_name           # ← 追加
  budget_settings_table_name  = module.budget.budget_settings_table_name  # ← 追加
  idempotency_keys_table_name = module.budget.idempotency_keys_table_name # ← 追加
  budget_reset_log_table_name = module.budget.budget_reset_log_table_name # ← 追加
}
```

## 検証結果

| 項目 | 結果 |
|---|---|
| `cd infra/modules/budget && terraform init -backend=false && terraform validate` | PASS |
| `cd infra/modules/budget && terraform fmt -recursive` | PASS |
| `cd infra/modules/budget && terraform test` | 11 PASS |
| `cd infra/modules/lambda_api && terraform test` | 既存 4 PASS (regression なし) |
| `cd infra/modules/api_gateway && terraform test` | 既存 5 PASS (regression なし) |
| `cd infra/envs/dev && terraform init -backend=false && terraform validate` | PASS |
