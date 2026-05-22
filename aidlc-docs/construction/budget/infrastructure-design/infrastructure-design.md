# Budget Unit — Infrastructure Design

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Construction Depth**: Standard
**Stage**: Infrastructure Design / Construction
**Predecessors**: NFR Design 完了 (PR #75)

本ドキュメントは Unit B の **Terraform リソース定義** を凍結する。NFR Design で論理化した LC-BUDGET-01〜12 を実 AWS リソースに展開する。

参照: [budget-infrastructure-design-plan.md](../../plans/budget-infrastructure-design-plan.md), [logical-components.md](../nfr-design/logical-components.md), [unit-interfaces.md](../../interfaces/unit-interfaces.md)

---

## 1. スコープと前提

### 1.1 本書のスコープ（Q-I1 確定）

新規 `infra/modules/budget/` モジュールとして構築するリソース:

1. **DynamoDB テーブル** 4 本（Wallet / BudgetSettings / IdempotencyKeys / BudgetResetLog）
2. **Scheduler Lambda**（Go, zip 方式）← Q-I2=A
3. **EventBridge Scheduler**（月末リセット cron）
4. **API Gateway ルート** 2 本（既存 API Gateway への追加）
5. **IAM 権限追加**（API Lambda Role + 新規 Scheduler Lambda Role + EventBridge Scheduler Role）
6. **CloudWatch Log Group**（Scheduler Lambda 用）

### 1.2 スコープ外（他 Unit / 後続 Unit が担当）

- API Lambda 本体（Unit A が構築済み `gp-{env}-api-fn`）
- API Gateway 本体・Cognito Authorizer（Unit A が構築済み）
- Unit C/D/E の API Gateway ルート追加 → 各 Unit の Infrastructure Design

### 1.3 不変前提

| 項目 | 値 | 出典 |
|---|---|---|
| Region | `ap-northeast-1` | NFR-COMP / Q-15 |
| IaC | Terraform | 要件 Q-12=C |
| 命名規則 | `gp-{env}-{resource}` | Unit A Q-I4 |
| タグ | `Project=goro2pay / Env={env} / Unit=budget / ManagedBy=terraform` | Unit A Q-I11 |
| DynamoDB Provisioned | 1 RCU / 1 WCU | Q-N4=B |
| DynamoDB 暗号化 | AWS マネージド（AES256 デフォルト） | Q-I4=A |
| PITR | 無効 | Q-I6=A |
| Scheduler Lambda タイムアウト | 30 秒 | Q-N3=A |
| Scheduler Lambda メモリ | 128MB | Q-N9=A |

---

## 2. ディレクトリ構造

```
infra/
├── envs/
│   └── dev/
│       └── main.tf               # module "budget" 呼出を追加
└── modules/
    └── budget/                   # 新規 (Q-I1=A)
        ├── README.md
        ├── dynamodb.tf           # DynamoDB テーブル 4 本
        ├── scheduler_lambda.tf   # Scheduler Lambda + EventBridge Scheduler
        ├── api_gateway_routes.tf # GET /api/wallet + POST /api/wallet/budget
        ├── iam.tf                # IAM Role/Policy (Scheduler + API Lambda 権限追加)
        ├── log_groups.tf         # CloudWatch Log Group (Scheduler Lambda 用)
        ├── variables.tf
        └── outputs.tf

apps/
└── scheduler/
    ├── main.go                   # Lambda handler (aws-lambda-go, bootstrap バイナリ)
    └── Makefile                  # build: GOOS=linux GOARCH=arm64 go build -o bootstrap .
                                   # (Code Generation で確定)
```

`infra/envs/dev/main.tf` に以下を追加:

```hcl
module "budget" {
  source = "../../modules/budget"

  env                      = var.env
  region                   = var.region
  api_id                   = module.auth.api_id
  api_lambda_invoke_arn    = module.auth.api_lambda_invoke_arn
  api_lambda_role_arn      = module.auth.api_lambda_role_arn
  cognito_authorizer_id    = module.auth.cognito_authorizer_id
}
```

---

## 3. リソース詳細

### 3.1 DynamoDB テーブル

#### 3.1.1 `aws_dynamodb_table.wallet`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-wallet` | Q-I4 |
| `billing_mode` | `PROVISIONED` | Q-N4=B |
| `read_capacity` | 1 | NFR-CAP-01 |
| `write_capacity` | 1 | NFR-CAP-01 |
| `hash_key` | `userId` | unit-interfaces.md §3.4 |
| `attribute(userId)` | S（文字列） | — |
| `server_side_encryption.enabled` | true（AES256 デフォルト） | Q-I4=A |
| `point_in_time_recovery.enabled` | false | Q-I6=A |
| `ttl` | 未設定 | — |

> **ConsistentRead** は Lambda コード側で指定（Terraform 設定項目ではない）。

#### 3.1.2 `aws_dynamodb_table.budget_settings`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-budget-settings` | Q-I4 |
| `billing_mode` | `PROVISIONED` | Q-N4=B |
| `read_capacity` | 1 | — |
| `write_capacity` | 1 | — |
| `hash_key` | `userId` | unit-interfaces.md §3.4 |
| `attribute(userId)` | S | — |
| `server_side_encryption.enabled` | true | Q-I4=A |
| `point_in_time_recovery.enabled` | false | Q-I6=A |

#### 3.1.3 `aws_dynamodb_table.idempotency_keys`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-idempotency-keys` | Q-I4 |
| `billing_mode` | `PROVISIONED` | Q-N4=B |
| `read_capacity` | 1 | — |
| `write_capacity` | 1 | — |
| `hash_key` | `key` | unit-interfaces.md §3.4 |
| `attribute(key)` | S | — |
| `ttl.attribute_name` | `expiresAt` | NFR-REL-02（TTL 24h） |
| `ttl.enabled` | true | — |
| `server_side_encryption.enabled` | true | Q-I4=A |
| `point_in_time_recovery.enabled` | false | Q-I6=A |

#### 3.1.4 `aws_dynamodb_table.budget_reset_log`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-budget-reset-log` | Q-I4 |
| `billing_mode` | `PROVISIONED` | Q-N4=B |
| `read_capacity` | 1 | — |
| `write_capacity` | 1 | — |
| `hash_key` | `resetDate` | unit-interfaces.md §3.4 |
| `range_key` | `userId` | unit-interfaces.md §3.4（SK あり） |
| `attribute(resetDate)` | S（`YYYY-MM` 形式） | — |
| `attribute(userId)` | S | — |
| `server_side_encryption.enabled` | true | Q-I4=A |
| `point_in_time_recovery.enabled` | false | Q-I6=A |

---

### 3.2 Scheduler Lambda（Q-I2=A: zip 方式）

#### 3.2.1 `data.archive_file.scheduler`

| 設定 | 値 |
|---|---|
| `type` | `zip` |
| `source_file` | `${path.root}/../../apps/scheduler/bootstrap`（pre-built Go バイナリ） |
| `output_path` | `${path.module}/.terraform/tmp/scheduler.zip` |

**前提**: `terraform apply` 前に `make -C apps/scheduler build` を実行し `bootstrap` バイナリを生成しておく（Q-I3=A、CI/CD なし）。

#### 3.2.2 `aws_lambda_function.scheduler`

| 設定 | 値 | 根拠 |
|---|---|---|
| `function_name` | `gp-${var.env}-scheduler-fn` | Q-I4 |
| `runtime` | `provided.al2023` | Go カスタムランタイム |
| `handler` | `bootstrap` | Go Lambda 標準ハンドラ名 |
| `role` | `aws_iam_role.scheduler_lambda.arn` | — |
| `filename` | `data.archive_file.scheduler.output_path` | Q-I2=A |
| `source_code_hash` | `data.archive_file.scheduler.output_base64sha256` | デプロイトリガ |
| `memory_size` | 128 | Q-N9=A |
| `timeout` | 30 | Q-N3=A |
| `architectures` | `["arm64"]` | Unit A と統一、コスト最適化 |
| `environment.DDB_TABLE_WALLET` | `aws_dynamodb_table.wallet.name` | unit-interfaces §10 |
| `environment.DDB_TABLE_BUDGET_SETTINGS` | `aws_dynamodb_table.budget_settings.name` | 同上 |
| `environment.DDB_TABLE_BUDGET_RESET_LOG` | `aws_dynamodb_table.budget_reset_log.name` | 同上 |
| `environment.AWS_REGION` | `ap-northeast-1` | — |
| `environment.LOG_LEVEL` | `info` | NFR-OBS-01 |

#### 3.2.3 `aws_lambda_permission.eventbridge_invoke_scheduler`

EventBridge Scheduler が Scheduler Lambda を呼び出せるよう許可。

```
principal  = "scheduler.amazonaws.com"
source_arn = aws_scheduler_schedule.monthly_reset.arn
action     = "lambda:InvokeFunction"
```

#### 3.2.4 `aws_scheduler_schedule.monthly_reset`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-monthly-reset` | Q-I4 |
| `schedule_expression` | `cron(0 15 L * ? *)` | Q-B8=A（月末最終日 0:00 JST） |
| `flexible_time_window.mode` | `OFF` | 厳密な時刻実行 |
| `target.arn` | `aws_lambda_function.scheduler.arn` | — |
| `target.role_arn` | `aws_iam_role.eventbridge_scheduler.arn` | — |
| `target.input` | `"{}"` | 空の JSON イベント |

> **失敗ハンドリング**: Q-I5=A。再試行なし、CloudWatch Logs への ERROR ログのみ。EventBridge Scheduler の `retry_policy.maximum_retry_attempts = 0` を設定。

---

### 3.3 API Gateway ルート（既存 API Gateway への追加）

#### 3.3.1 `aws_apigatewayv2_integration.api_lambda`

| 設定 | 値 | 根拠 |
|---|---|---|
| `api_id` | `var.api_id` | Unit A module output |
| `integration_type` | `AWS_PROXY` | — |
| `integration_uri` | `var.api_lambda_invoke_arn` | Unit A module output |
| `payload_format_version` | `2.0` | Unit A と統一 |
| `integration_method` | `POST` | — |

#### 3.3.2 `aws_apigatewayv2_route.get_wallet`

| 設定 | 値 | 根拠 |
|---|---|---|
| `api_id` | `var.api_id` | — |
| `route_key` | `GET /api/wallet` | unit-interfaces §3.3 |
| `target` | `integrations/${aws_apigatewayv2_integration.api_lambda.id}` | — |
| `authorization_type` | `JWT` | 認証必須 |
| `authorizer_id` | `var.cognito_authorizer_id` | Unit A module output |

#### 3.3.3 `aws_apigatewayv2_route.post_wallet_budget`

| 設定 | 値 | 根拠 |
|---|---|---|
| `api_id` | `var.api_id` | — |
| `route_key` | `POST /api/wallet/budget` | unit-interfaces §3.3 |
| `target` | `integrations/${aws_apigatewayv2_integration.api_lambda.id}` | — |
| `authorization_type` | `JWT` | 認証必須 |
| `authorizer_id` | `var.cognito_authorizer_id` | Unit A module output |

> **`POST /api/wallet/deduct` はルートなし**: `Deduct` は Unit C の `OrderService` が Go 関数呼び出し（同一 Lambda 内の DI）で実行するため、HTTP ルートは不要。

---

### 3.4 IAM Roles & Policies

#### 3.4.1 API Lambda Role への DynamoDB 権限追加

Unit A が作成した `aws_iam_role.api_lambda`（`var.api_lambda_role_arn`）にインラインポリシーを追加。

```json
{
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan"
    ],
    "Resource": [
      "${aws_dynamodb_table.wallet.arn}",
      "${aws_dynamodb_table.budget_settings.arn}",
      "${aws_dynamodb_table.idempotency_keys.arn}",
      "${aws_dynamodb_table.budget_reset_log.arn}"
    ]
  }]
}
```

Terraform リソース: `aws_iam_role_policy.api_lambda_dynamodb`

#### 3.4.2 `aws_iam_role.scheduler_lambda`（新規）

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-scheduler-role` |
| `assume_role_policy` | `lambda.amazonaws.com` service principal |

インラインポリシー `aws_iam_role_policy.scheduler_lambda_dynamodb`:

```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"],
      "Resource": ["${aws_cloudwatch_log_group.scheduler.arn}", "${aws_cloudwatch_log_group.scheduler.arn}:*"]
    },
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:UpdateItem",
        "dynamodb:Scan"
      ],
      "Resource": ["${aws_dynamodb_table.wallet.arn}"]
    },
    {
      "Effect": "Allow",
      "Action": ["dynamodb:GetItem"],
      "Resource": ["${aws_dynamodb_table.budget_settings.arn}"]
    },
    {
      "Effect": "Allow",
      "Action": ["dynamodb:PutItem", "dynamodb:GetItem"],
      "Resource": ["${aws_dynamodb_table.budget_reset_log.arn}"]
    }
  ]
}
```

#### 3.4.3 `aws_iam_role.eventbridge_scheduler`（新規）

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-eventbridge-scheduler-role` |
| `assume_role_policy` | `scheduler.amazonaws.com` service principal |

インラインポリシー: `lambda:InvokeFunction` on `aws_lambda_function.scheduler.arn` のみ

---

### 3.5 CloudWatch Log Group

| Resource | name | retention_in_days | 根拠 |
|---|---|---|---|
| `aws_cloudwatch_log_group.scheduler` | `/aws/lambda/gp-${var.env}-scheduler-fn` | 7 | Unit A と統一 |

---

## 4. Variables / Outputs

### 4.1 Module Variables（`infra/modules/budget/variables.tf`）

| 名前 | 型 | 説明 |
|---|---|---|
| `env` | string | 環境識別子（例: `dev`） |
| `region` | string | AWS Region（デフォルト: `ap-northeast-1`） |
| `api_id` | string | Unit A が出力した API Gateway ID |
| `api_lambda_invoke_arn` | string | Unit A が出力した API Lambda invoke ARN |
| `api_lambda_role_arn` | string | Unit A が出力した API Lambda IAM Role ARN |
| `cognito_authorizer_id` | string | Unit A が出力した Cognito Authorizer ID |

### 4.2 Module Outputs（`infra/modules/budget/outputs.tf`）

| 名前 | 値 | 用途 |
|---|---|---|
| `wallet_table_name` | `aws_dynamodb_table.wallet.name` | Lambda env var / Unit E 参照 |
| `wallet_table_arn` | `aws_dynamodb_table.wallet.arn` | Unit E の IAM 権限追加時 |
| `budget_settings_table_name` | `aws_dynamodb_table.budget_settings.name` | Lambda env var / Unit E 参照 |
| `budget_settings_table_arn` | `aws_dynamodb_table.budget_settings.arn` | Unit E の IAM 権限追加時 |
| `idempotency_keys_table_name` | `aws_dynamodb_table.idempotency_keys.name` | Lambda env var |
| `budget_reset_log_table_name` | `aws_dynamodb_table.budget_reset_log.name` | Lambda env var |
| `scheduler_lambda_function_name` | `aws_lambda_function.scheduler.function_name` | デプロイ確認 |
| `scheduler_lambda_arn` | `aws_lambda_function.scheduler.arn` | — |

---

## 5. API Lambda 環境変数の追加

`envs/dev/main.tf` で Unit A の API Lambda に Unit B の DynamoDB テーブル名を環境変数として追加する。Unit A の API Lambda は環境変数を Terraform 外から更新できないため、**Unit A module の variables に Unit B の table name を渡す形** で対応する。

実装方針: `envs/dev/main.tf` で以下のように `aws_lambda_function_event_invoke_config` または `aws_lambda_function` の `environment` を更新するか、**別リソース `aws_lambda_function.api` の `environment` を `lifecycle` の後から manage する形**。

> **引き継ぎ**: 具体的な `envs/dev/main.tf` の統合方法は Code Generation で確定。推奨は `envs/dev/main.tf` にて `module.budget` の outputs を `module.auth` の API Lambda 環境変数として渡す構成。

---

## 6. terraform-test 統合

| テストファイル | 目的 |
|---|---|
| `budget_basic.tftest.hcl` | mock_provider で `terraform plan` がエラーなく成立することを確認 |
| `budget_dynamodb_tables.tftest.hcl` | 4 テーブルの billing_mode / ttl 設定が NFR と一致することを確認 |
| `budget_scheduler_lambda.tftest.hcl` | Scheduler Lambda の memory / timeout 設定を確認 |
| `budget_iam_least_privilege.tftest.hcl` | Scheduler Lambda IAM が最小権限（Wallet Scan / BudgetSettings Get / ResetLog PutItem）であることを確認 |

---

## 7. リソース → 論理コンポーネント / NFR トレーサビリティ

| Terraform リソース | 対応 LC | 対応 NFR / Q |
|---|---|---|
| `aws_dynamodb_table.wallet` | LC-BUDGET-03 | NFR-REL-01 / Q-N4 |
| `aws_dynamodb_table.budget_settings` | LC-BUDGET-04 | Q-N4 |
| `aws_dynamodb_table.idempotency_keys` | LC-BUDGET-05 | NFR-REL-02（TTL 24h） |
| `aws_dynamodb_table.budget_reset_log` | LC-BUDGET-06 | NFR-REL-03 |
| `aws_lambda_function.scheduler` | LC-BUDGET-07 | Q-N3=A / Q-N9=A |
| `aws_scheduler_schedule.monthly_reset` | LC-BUDGET-07 | Q-B8=A |
| `aws_apigatewayv2_route.get_wallet` | LC-BUDGET-01 | unit-interfaces §3.3 |
| `aws_apigatewayv2_route.post_wallet_budget` | LC-BUDGET-01 | unit-interfaces §3.3 |
| `aws_iam_role_policy.api_lambda_dynamodb` | LC-BUDGET-03〜06 | 最小権限原則 |
| `aws_iam_role.scheduler_lambda` | LC-BUDGET-07 | 最小権限原則 |
| `aws_cloudwatch_log_group.scheduler` | LC-AUTH-05（再利用）| NFR-OBS-01 |

---

## 8. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Code Generation** | `infra/modules/budget/` 配下の全 `.tf` ファイル実装、`envs/dev/main.tf` への `module "budget"` 追加、`apps/scheduler/main.go`（Go Lambda handler）、`apps/scheduler/Makefile`（バイナリビルド手順） |
| **Unit E Infrastructure Design** | `aws_dynamodb_table.wallet.arn` / `aws_dynamodb_table.budget_settings.arn` を `module.budget` outputs 経由で参照し、Unit E の IAM ポリシーに追加 |
