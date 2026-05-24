# Order Unit (Unit C) — Infrastructure Design

**Document Version**: 1.0
**Created**: 2026-05-24
**Stage**: Construction / Infrastructure Design
**Unit**: C — `order` (代行手配コア / MVP の心臓)
**Depth**: Comprehensive
**Related**:
- Plan: [order-infrastructure-design-plan.md](../../plans/order-infrastructure-design-plan.md)
- NFR Design: [nfr-design-patterns.md](../nfr-design/nfr-design-patterns.md), [logical-components.md](../nfr-design/logical-components.md)
- NFR Requirements: [nfr-requirements.md](../nfr-requirements/nfr-requirements.md), [tech-stack-decisions.md](../nfr-requirements/tech-stack-decisions.md)
- Functional Design: [business-logic-model.md](../functional-design/business-logic-model.md), [domain-entities.md](../functional-design/domain-entities.md)
- 凍結契約: [unit-interfaces.md](../../interfaces/unit-interfaces.md)
- Unit A Infrastructure Design: `aidlc-docs/construction/auth/infrastructure-design/infrastructure-design.md`

本ドキュメントは Unit C `order` の **Terraform リソース定義** を凍結する。NFR Design で論理化した LC-ORDER-04（OrderHistoryRepository）/ LC-ORDER-05（BedrockAdapter）/ LC-ORDER-12〜13（観測性）/ LC-ORDER-14〜15（init）の物理実装と、CloudWatch Alarms / SNS Topic / AWS Budgets / API Gateway routes / IAM Policy 拡張を実 AWS リソースに展開する。

---

## 1. スコープと前提

### 1.1 本書のスコープ

Unit C Construction の Infrastructure Design として、以下を構築する:

1. **DynamoDB `OrderHistory` テーブル**（PK / SK / TTL 90 日 / GSI なし）← LC-ORDER-04
2. **Bedrock IAM Policy**（Claude Haiku 4.5 Foundation Model + Inference Profile ARN 限定）← LC-ORDER-05 / NFRC-C20
3. **DynamoDB IAM Policy**（OrderHistory CRUD 権限）
4. **API Gateway routes 2 本追加**（`POST /api/orders` / `GET /api/orders`）← 既存 `modules/api_gateway/routes.tf` 拡張
5. **CloudWatch Logs metric filter 3 種**（`bedrock_retry` / `fallback_triggered` / `place_order_complete` p95）← NFRC-C13
6. **CloudWatch Alarms 3 種**（NFRC-C13-1/C13-2/C13-3）
7. **SNS Topic**（横串 `gp-{env}-alarms`）+ email subscription
8. **AWS Budgets**（Bedrock 月 $5 / 80% 警告 + 100% 警報）← NFRC-C20
9. **Terraform module 構成**（機能別 3 module 新規 + 既存 2 module 拡張）
10. **terraform-test**（DynamoDB スキーマ / Bedrock IAM 最小権限 / Alarms 閾値の 3 ファイル）

### 1.2 スコープ外（他 Unit / 後続ステージが担当）

- API Lambda 本体実装 → Unit A 既存（`gp-{env}-api-fn`、Go + Gin + LWA、ECR + CodePipeline でデプロイ済み）
- API Gateway 本体 / JWT Authorizer / Stage / Throttling → Unit A 既存
- Cognito User Pool / App Client → Unit A 既存
- `Wallet` / `IdempotencyKeys` / `BudgetSettings` DynamoDB テーブル → Unit B 既存（Unit C は WalletService 経由で参照のみ）
- BedrockAdapter / DeliveryAdapter / OrderHistoryRepository の Go 実装 → Code Generation
- `useOrder` / `GoroButton` / `ToastHost` 等の Frontend 実装 → Code Generation
- CloudWatch Dashboard → NFRC-C14 で MVP 不実装
- X-Ray / カスタムメトリクス（PutMetricData）→ NFRC-C14 で MVP 不実装

### 1.3 不変前提

| 項目 | 値 | 出典 |
|---|---|---|
| Region | `ap-northeast-1` | NFR-COMP / Q-15 |
| IaC | Terraform | 要件 Q-12 = C |
| 命名規則 | `gp-{env}-{resource}` | Unit A Q-I4 |
| Tags | `Project=goro2pay` / `Env={env}` / `Unit=order` or `Unit=observability` / `ManagedBy=terraform` | Unit A Q-I11 |
| Module 規約 | terraform-module-design / terraform-coding-rule / terraform-test 準拠 | チームベースライン |
| API Lambda Memory | 256MB / arm64 / Timeout 10s | NFRC-C18 / Q-N9 = B |
| Bedrock モデル | `jp.anthropic.claude-haiku-4-5-20251001-v1:0`（Inference Profile） | NFRC-C20 / Q-N8 = B / Q-I10 = A |
| OrderHistory PK / SK | `PK=USER#{userID}` / `SK=ORDER#{orderedAt}#{orderID}` | unit-interfaces.md §3.2 |
| TTL 属性名 | `expiresAt`（90 日） | NFRC-C12 / 凍結契約整合修正済み |
| OrderHistory キャパシティ | プロビジョンド 1 RCU / 1 WCU | Q-I2 = A / Unit B 統一 |
| 暗号化 | AWS マネージド AES256 | Q-I9 = A / Unit B 統一 |
| PITR | 無効 | Q-I8 = A / Unit B 統一 |

---

## 2. ディレクトリ構造

Unit A の機能別 5 module + Unit B 既存 + Unit C で新規 3 module を追加。Unit A の構造（`unit-of-work.md §4.1` 整合）を踏襲し、各 module 配下のファイル分割は terraform-module-design 準拠（`main.tf` / `data.tf` / `locals.tf` / `variables.tf` / `outputs.tf`）。

```
infra/
├── envs/
│   ├── dev/
│   │   ├── backend.tf            # 既存（変更なし）
│   │   ├── providers.tf          # 既存（default_tags は Project / Env / ManagedBy のみ、Unit タグは module 個別）
│   │   ├── versions.tf           # 既存（変更なし）
│   │   ├── locals.tf             # 既存 + alarm_email = "alerts@example.com" 追加
│   │   ├── main.tf               # 既存 + Unit C 3 module 追記、module.lambda_api に additional_policy_arns 追加
│   │   └── outputs.tf            # 既存 + dynamodb_table_name / sns_topic_arn 追加
│   └── prd/
│       └── README.md             # placeholder
├── modules/
│   ├── codestar_connection/      # Unit A 既存（変更なし）
│   ├── cognito/                  # Unit A 既存（変更なし）
│   ├── api_gateway/              # Unit A 既存
│   │   ├── main.tf
│   │   ├── api_gateway.tf
│   │   ├── routes.tf             # ★ Unit C routes 2 本を追加（POST /api/orders / GET /api/orders）
│   │   ├── locals.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── tests/
│   ├── lambda_api/               # Unit A 既存
│   │   ├── main.tf
│   │   ├── api_lambda.tf
│   │   ├── ecr.tf
│   │   ├── codepipeline.tf
│   │   ├── iam.tf                # ★ for_each で additional_policy_arns を attach
│   │   ├── locals.tf
│   │   ├── variables.tf          # ★ additional_policy_arns 変数追加
│   │   ├── outputs.tf            # ★ api_log_group_name 追加（observability から参照）
│   │   └── tests/
│   ├── amplify/                  # Unit A 既存（変更なし）
│   ├── budget/                   # Unit B 既存（別 PR で見直し対象）
│   ├── order_history/            # ★ 新規（Unit C 専用）
│   │   ├── main.tf               # aws_dynamodb_table + aws_iam_policy (DynamoDB CRUD)
│   │   ├── locals.tf             # table_name / policy_name 等の locals
│   │   ├── variables.tf          # env / tags
│   │   ├── outputs.tf            # dynamodb_table_name / dynamodb_table_arn / dynamodb_policy_arn
│   │   └── tests/
│   │       └── dynamodb_schema.tftest.hcl
│   ├── bedrock/                  # ★ 新規（Unit C / D 共有候補）
│   │   ├── main.tf               # aws_iam_policy (InvokeModel + Converse, Foundation Model + Inference Profile ARN 限定)
│   │   ├── data.tf               # aws_caller_identity / aws_region
│   │   ├── locals.tf             # model_id / inference_profile_id / model_arns
│   │   ├── variables.tf          # env / region
│   │   ├── outputs.tf            # bedrock_policy_arn
│   │   └── tests/
│   │       └── bedrock_iam_least_privilege.tftest.hcl
│   └── observability/            # ★ 新規（横串、Unit D/E 再利用可）
│       ├── main.tf               # SNS Topic + Subscription / metric filter ×3 / alarm ×3 / Budgets
│       ├── data.tf               # aws_caller_identity (Budgets account_id)
│       ├── locals.tf             # alarm 名・しきい値・metric namespace
│       ├── variables.tf          # alarm_email / api_log_group_name / p95_threshold_ms / retry_threshold_count / fallback_threshold_count / bedrock_budget_limit_usd / env
│       ├── outputs.tf            # sns_topic_arn
│       └── tests/
│           └── cloudwatch_alarms.tftest.hcl
└── lambdas/
    └── (既存)
```

---

## 3. リソース詳細

### 3.1 `modules/order_history/`（Unit C 専用）

#### 3.1.1 DynamoDB OrderHistory（main.tf）

```hcl
resource "aws_dynamodb_table" "order_history" {
  name         = local.table_name
  billing_mode = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1
  hash_key     = "PK"
  range_key    = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  server_side_encryption {
    enabled = true # AWS マネージド AES256（Q-I9 = A）
  }

  point_in_time_recovery {
    enabled = false # Q-I8 = A
  }

  tags = merge(local.tags, { Unit = "order" })
}
```

| プロパティ | 値 | 由来 |
|---|---|---|
| `name` | `gp-{env}-order-history` | Q-I4 命名規則 |
| `billing_mode` | `PROVISIONED` | Q-I2 = A |
| `read_capacity` / `write_capacity` | 1 / 1 | Q-I2 = A、無料枠内 |
| `hash_key` | `PK` (String、`USER#{userID}`) | unit-interfaces.md §3.2 |
| `range_key` | `SK` (String、`ORDER#{orderedAt}#{orderID}`) | unit-interfaces.md §3.2 |
| `ttl.attribute_name` | `expiresAt`（90 日 epoch 秒） | NFRC-C12 / 凍結契約整合 |
| `server_side_encryption.enabled` | true（AWS マネージド AES256） | Q-I9 = A |
| `point_in_time_recovery.enabled` | false | Q-I8 = A |
| GSI | なし（凍結契約整合修正で `GSI_IdempotencyKey` を撤回済み） | unit-interfaces.md §3.2 |

#### 3.1.2 DynamoDB アクセス用 IAM Policy（main.tf）

```hcl
resource "aws_iam_policy" "dynamodb_order_history" {
  name        = local.policy_name
  description = "Unit C: OrderHistory DynamoDB CRUD permissions"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "OrderHistoryCRUD"
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:UpdateItem",
        ]
        Resource = aws_dynamodb_table.order_history.arn
      },
    ]
  })

  tags = merge(local.tags, { Unit = "order" })
}
```

許可するアクション:
- `PutItem`: `OrderHistoryRepository.Insert`（LC-ORDER-04）
- `GetItem`: 単一注文取得（OrderCompletionScreen 用、Server Component から呼出予定）
- `Query`: `OrderHistoryRepository.Query`（履歴一覧、フォールバック分岐閾値判定 BR-C06）
- `UpdateItem`: 将来の状態遷移用（Q-D14 でステート保持を選択する場合の余地）

許可しないアクション: `DeleteItem`（TTL 自動削除のみ）/ `Scan` / `BatchWrite`（最小権限）

#### 3.1.3 locals.tf

```hcl
locals {
  table_name  = "gp-${var.env}-order-history"
  policy_name = "gp-${var.env}-order-history-policy"
  tags        = var.tags
}
```

#### 3.1.4 variables.tf

```hcl
variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}
```

#### 3.1.5 outputs.tf

```hcl
output "dynamodb_table_name" {
  value       = aws_dynamodb_table.order_history.name
  description = "OrderHistory table name (for env var injection to Lambda)"
}

output "dynamodb_table_arn" {
  value       = aws_dynamodb_table.order_history.arn
  description = "OrderHistory table ARN"
}

output "dynamodb_policy_arn" {
  value       = aws_iam_policy.dynamodb_order_history.arn
  description = "IAM policy ARN for DynamoDB OrderHistory CRUD (for attach to API Lambda Role)"
}
```

---

### 3.2 `modules/bedrock/`（Unit C / D 共有候補）

#### 3.2.1 Bedrock IAM Policy（main.tf）

```hcl
resource "aws_iam_policy" "bedrock_inference" {
  name        = local.policy_name
  description = "Bedrock Claude Haiku 4.5 inference (InvokeModel + Converse, Foundation Model + Inference Profile ARN limited)"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "BedrockInferenceClaudeHaiku"
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:Converse",
        ]
        Resource = local.model_arns
      },
    ]
  })

  tags = merge(local.tags, { Unit = "bedrock" })
}
```

#### 3.2.2 data.tf

```hcl
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
```

#### 3.2.3 locals.tf

```hcl
locals {
  model_id            = "anthropic.claude-haiku-4-5-20251001-v1:0"
  inference_profile_id = "jp.anthropic.claude-haiku-4-5-20251001-v1:0"
  policy_name          = "gp-${var.env}-bedrock-inference-policy"
  account_id           = data.aws_caller_identity.current.account_id

  # Foundation Model ARN (region 横断、Inference Profile が内部で複数リージョンを使用するため)
  # + Inference Profile ARN (アカウント別、リージョン固定)
  model_arns = [
    "arn:aws:bedrock:*::foundation-model/${local.model_id}",
    "arn:aws:bedrock:${var.region}:${local.account_id}:inference-profile/${local.inference_profile_id}",
  ]

  tags = var.tags
}
```

#### 3.2.4 variables.tf

```hcl
variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

variable "region" {
  description = "AWS region (ap-northeast-1 想定)"
  type        = string
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}
```

#### 3.2.5 outputs.tf

```hcl
output "bedrock_policy_arn" {
  value       = aws_iam_policy.bedrock_inference.arn
  description = "IAM policy ARN for Bedrock inference (for attach to API Lambda Role)"
}
```

許可するアクション:
- `bedrock:InvokeModel`: 互換性確保（NFRC-C20 で Converse API 確定、ただし将来移行用）
- `bedrock:Converse`: NFRC-C20 / Q-N8 = B 確定（FD で確定済み）

許可しないアクション:
- `bedrock:InvokeModelWithResponseStream`（MVP では未使用、Q-I3 = A）
- `bedrock:*`（最小権限）
- 他モデル ARN（Q-I3 = A、Foundation Model + Inference Profile に限定）

---

### 3.3 `modules/observability/`（横串、Unit D/E 再利用可）

#### 3.3.1 SNS Topic（main.tf）

```hcl
resource "aws_sns_topic" "alarms" {
  name = local.sns_topic_name
  tags = merge(local.tags, { Unit = "observability" })
}

resource "aws_sns_topic_subscription" "alarm_email" {
  topic_arn              = aws_sns_topic.alarms.arn
  protocol               = "email"
  endpoint               = var.alarm_email
  confirmation_timeout_in_minutes = 5
}
```

#### 3.3.2 CloudWatch Logs metric filter ×3（main.tf）

NFRC-C13 のアラーム検知に必要なログイベントを抽出（NFRC-C13 / P-OBS-03 整合）。

```hcl
# NFRC-C13-1: PlaceOrder p95 > 3000ms
resource "aws_cloudwatch_log_metric_filter" "place_order_latency" {
  name           = "${local.env_prefix}-place-order-latency"
  log_group_name = var.api_log_group_name
  pattern        = "{ $.event = \"place_order_complete\" && $.latencyMs = * }"

  metric_transformation {
    name          = "PlaceOrderLatencyMs"
    namespace     = local.metric_namespace
    value         = "$.latencyMs"
    default_value = null
    unit          = "Milliseconds"
  }
}

# NFRC-C13-2: Bedrock リトライ発動
resource "aws_cloudwatch_log_metric_filter" "bedrock_retry" {
  name           = "${local.env_prefix}-bedrock-retry"
  log_group_name = var.api_log_group_name
  pattern        = "{ $.event = \"bedrock_retry\" }"

  metric_transformation {
    name          = "BedrockRetryCount"
    namespace     = local.metric_namespace
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}

# NFRC-C13-3: フォールバック発動
resource "aws_cloudwatch_log_metric_filter" "fallback_triggered" {
  name           = "${local.env_prefix}-fallback-triggered"
  log_group_name = var.api_log_group_name
  pattern        = "{ $.event = \"fallback_triggered\" }"

  metric_transformation {
    name          = "FallbackTriggeredCount"
    namespace     = local.metric_namespace
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}
```

#### 3.3.3 CloudWatch Alarms ×3（main.tf）

```hcl
# NFRC-C13-1: PlaceOrder p95 違反
resource "aws_cloudwatch_metric_alarm" "place_order_p95_breach" {
  alarm_name          = "${local.env_prefix}-place-order-p95-breach"
  alarm_description   = "PlaceOrder p95 > ${var.p95_threshold_ms}ms (NFRC-C13-1)"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  threshold           = var.p95_threshold_ms
  treat_missing_data  = "notBreaching"

  metric_query {
    id          = "p95"
    return_data = true
    metric {
      metric_name = "PlaceOrderLatencyMs"
      namespace   = local.metric_namespace
      period      = 300 # 5 min
      stat        = "p95"
    }
  }

  alarm_actions = [aws_sns_topic.alarms.arn]
  ok_actions    = [aws_sns_topic.alarms.arn]
  tags          = merge(local.tags, { Unit = "observability", Severity = "High" })
}

# NFRC-C13-2: Bedrock リトライ発動
resource "aws_cloudwatch_metric_alarm" "bedrock_retry_burst" {
  alarm_name          = "${local.env_prefix}-bedrock-retry-burst"
  alarm_description   = "Bedrock retry events ≥ ${var.retry_threshold_count} per 5 min (NFRC-C13-2)"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "BedrockRetryCount"
  namespace           = local.metric_namespace
  period              = 300
  statistic           = "Sum"
  threshold           = var.retry_threshold_count
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.alarms.arn]
  ok_actions    = [aws_sns_topic.alarms.arn]
  tags          = merge(local.tags, { Unit = "observability", Severity = "Medium" })
}

# NFRC-C13-3: フォールバック発動
resource "aws_cloudwatch_metric_alarm" "fallback_triggered_burst" {
  alarm_name          = "${local.env_prefix}-fallback-triggered-burst"
  alarm_description   = "Fallback triggered events ≥ ${var.fallback_threshold_count} per 5 min (NFRC-C13-3)"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "FallbackTriggeredCount"
  namespace           = local.metric_namespace
  period              = 300
  statistic           = "Sum"
  threshold           = var.fallback_threshold_count
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.alarms.arn]
  ok_actions    = [aws_sns_topic.alarms.arn]
  tags          = merge(local.tags, { Unit = "observability", Severity = "High" })
}
```

| Alarm | 閾値（デフォルト） | 評価期間 | Severity |
|---|---|---|---|
| `place_order_p95_breach` (NFRC-C13-1) | p95 > 3000ms を 3 回連続検知 | 5 min × 3 datapoints | High |
| `bedrock_retry_burst` (NFRC-C13-2) | リトライ ≥ 5 / 5 min | 5 min × 1 datapoint | Medium |
| `fallback_triggered_burst` (NFRC-C13-3) | フォールバック ≥ 3 / 5 min | 5 min × 1 datapoint | High |

#### 3.3.4 AWS Budgets（Bedrock コスト監視、main.tf）

```hcl
resource "aws_budgets_budget" "bedrock" {
  name              = "${local.env_prefix}-bedrock-budget"
  budget_type       = "COST"
  limit_amount      = tostring(var.bedrock_budget_limit_usd)
  limit_unit        = "USD"
  time_unit         = "MONTHLY"
  time_period_start = local.budget_start

  cost_filter {
    name   = "Service"
    values = ["Amazon Bedrock"]
  }

  # Q-I7 = A: 80% 警告 + 100% 警報
  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type           = "ACTUAL"
    subscriber_sns_topic_arns  = [aws_sns_topic.alarms.arn]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type           = "ACTUAL"
    subscriber_sns_topic_arns  = [aws_sns_topic.alarms.arn]
  }
}
```

#### 3.3.5 locals.tf

```hcl
locals {
  env_prefix       = "gp-${var.env}"
  sns_topic_name   = "${local.env_prefix}-alarms"
  metric_namespace = "GoroPay/Order"
  budget_start     = "2026-05-01_00:00"
  tags             = var.tags
}
```

#### 3.3.6 variables.tf

```hcl
variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

variable "alarm_email" {
  description = "Email address to receive CloudWatch alarm notifications via SNS"
  type        = string
}

variable "api_log_group_name" {
  description = "CloudWatch Log Group name of API Lambda (for metric filters)"
  type        = string
}

variable "p95_threshold_ms" {
  description = "PlaceOrder p95 latency alarm threshold (NFRC-C13-1)"
  type        = number
  default     = 3000
}

variable "retry_threshold_count" {
  description = "Bedrock retry burst alarm threshold per 5 min (NFRC-C13-2)"
  type        = number
  default     = 5
}

variable "fallback_threshold_count" {
  description = "Fallback triggered burst alarm threshold per 5 min (NFRC-C13-3)"
  type        = number
  default     = 3
}

variable "bedrock_budget_limit_usd" {
  description = "Monthly Bedrock budget limit in USD (NFRC-C20)"
  type        = number
  default     = 5
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}
```

#### 3.3.7 outputs.tf

```hcl
output "sns_topic_arn" {
  value       = aws_sns_topic.alarms.arn
  description = "SNS Topic ARN for alarm notifications (for reuse by Unit D/E)"
}
```

---

### 3.4 既存 `modules/api_gateway/` への追記（Q-I11 = B）

`modules/api_gateway/routes.tf` に Unit C ルート 2 本を追記。

```hcl
# Unit C: POST /api/orders (PlaceOrder)
resource "aws_apigatewayv2_integration" "post_orders" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_uri        = var.api_lambda_invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "post_orders" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "POST /api/orders"
  target             = "integrations/${aws_apigatewayv2_integration.post_orders.id}"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
  authorization_type = "JWT"
}

# Unit C: GET /api/orders (GetHistory)
resource "aws_apigatewayv2_integration" "get_orders" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_uri        = var.api_lambda_invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "get_orders" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /api/orders"
  target             = "integrations/${aws_apigatewayv2_integration.get_orders.id}"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
  authorization_type = "JWT"
}
```

`aws_lambda_permission.apigw_invoke_api` は `envs/dev/main.tf` 既存のもの（`source_arn = "${execution_arn}/*/*"` ワイルドカード）でカバー済み、追加不要。

---

### 3.5 既存 `modules/lambda_api/` への追記（Q-I12 改定版）

#### 3.5.1 variables.tf に追加

```hcl
variable "additional_policy_arns" {
  description = "Additional IAM policy ARNs to attach to the API Lambda execution role"
  type        = list(string)
  default     = []
}
```

#### 3.5.2 iam.tf に追加（既存末尾）

```hcl
resource "aws_iam_role_policy_attachment" "additional" {
  for_each   = toset(var.additional_policy_arns)
  role       = aws_iam_role.api_lambda.name
  policy_arn = each.value
}
```

#### 3.5.3 outputs.tf に追加

```hcl
output "api_log_group_name" {
  value       = aws_cloudwatch_log_group.api_lambda.name
  description = "CloudWatch Log Group name of API Lambda (for observability metric filters)"
}
```

`api_lambda_function_name` を環境変数として渡す部分（既存）に Unit C 用環境変数を追加:

```hcl
# api_lambda.tf 既存リソースの environment 設定に追記
resource "aws_lambda_function" "api" {
  # ...既存...
  environment {
    variables = {
      # 既存
      COGNITO_USER_POOL_ID = var.cognito_user_pool_id
      COGNITO_CLIENT_ID    = var.cognito_user_pool_client_id
      # Unit C 追加
      ORDER_HISTORY_TABLE_NAME       = var.order_history_table_name
      BEDROCK_INFERENCE_PROFILE_ID   = "jp.anthropic.claude-haiku-4-5-20251001-v1:0"
    }
  }
}
```

`variables.tf` に `order_history_table_name` 変数を追加（envs から注入）。

---

### 3.6 envs/dev/ への追記

#### 3.6.1 locals.tf 追記

```hcl
locals {
  # 既存
  env           = "dev"
  region        = "ap-northeast-1"
  github_owner  = "ryotinjpn"
  github_repo   = "goro2pay"
  github_branch = "develop"

  # Unit C / 横串 observability 追加
  alarm_email = "alerts@example.com" # 注: 本番化時は terraform.tfvars or env var で外部注入
}
```

#### 3.6.2 main.tf 追記（末尾、Unit C 3 module + lambda_api 拡張）

```hcl
# Unit C: OrderHistory DynamoDB + アクセス用 IAM Policy
module "order_history" {
  source = "../../modules/order_history"
  env    = local.env
}

# Unit C / D 共有: Bedrock IAM Policy
module "bedrock" {
  source = "../../modules/bedrock"
  env    = local.env
  region = local.region
}

# 横串 Observability: CloudWatch Alarms / SNS Topic / Budgets
module "observability" {
  source                   = "../../modules/observability"
  env                      = local.env
  alarm_email              = local.alarm_email
  api_log_group_name       = module.lambda_api.api_log_group_name
  # しきい値は NFRC-C13 デフォルト値を踏襲（必要時 tfvars で上書き）
}
```

`module "lambda_api"` 既存呼出に `additional_policy_arns` と `order_history_table_name` を追加:

```hcl
module "lambda_api" {
  source                      = "../../modules/lambda_api"
  env                         = local.env
  region                      = local.region
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
  codestar_connection_arn     = module.codestar_connection.connection_arn
  github_owner                = local.github_owner
  github_repo_name            = local.github_repo
  github_branch               = local.github_branch
  # Unit C 追加
  additional_policy_arns      = [
    module.order_history.dynamodb_policy_arn,
    module.bedrock.bedrock_policy_arn,
  ]
  order_history_table_name = module.order_history.dynamodb_table_name
}
```

#### 3.6.3 outputs.tf 追記

```hcl
# 既存
output "user_pool_id" { value = module.cognito.user_pool_id }
output "user_pool_client_id" { value = module.cognito.user_pool_client_id }
output "api_endpoint" { value = module.api_gateway.api_endpoint }
output "amplify_default_domain" { value = module.amplify.amplify_default_domain }
output "ecr_repository_url" { value = module.lambda_api.ecr_repository_url }
output "codepipeline_name" { value = module.lambda_api.codepipeline_name }

# Unit C 追加
output "order_history_table_name" { value = module.order_history.dynamodb_table_name }
output "sns_topic_arn" { value = module.observability.sns_topic_arn }
```

---

## 4. 環境変数（Lambda 経由）

API Lambda に注入される環境変数の Unit C 関連（既存に追加）:

| 環境変数 | 値 | 用途 |
|---|---|---|
| `ORDER_HISTORY_TABLE_NAME` | `gp-{env}-order-history` | LC-ORDER-04 OrderHistoryRepository |
| `BEDROCK_INFERENCE_PROFILE_ID` | `jp.anthropic.claude-haiku-4-5-20251001-v1:0` | LC-ORDER-05 BedrockAdapter |
| `AWS_REGION` | `ap-northeast-1`（Lambda 標準注入） | P-INIT-01 init() で利用 |

`AWS_REGION` は AWS Lambda が自動注入するため Terraform 側では明示的に設定不要。

---

## 5. terraform-test 戦略（Q-I13 = C）

mock_provider を使用し、AWS API 呼出なしのオフラインテストを 3 ファイルで実装する（terraform-test 規約準拠）。

### 5.1 `modules/order_history/tests/dynamodb_schema.tftest.hcl`

検証項目:
- `aws_dynamodb_table.order_history.name == "gp-dev-order-history"`
- `hash_key == "PK"` / `range_key == "SK"`
- `ttl[0].attribute_name == "expiresAt"` / `ttl[0].enabled == true`
- `server_side_encryption[0].enabled == true`（AES256）
- `point_in_time_recovery[0].enabled == false`
- `billing_mode == "PROVISIONED"` / `read_capacity == 1` / `write_capacity == 1`

### 5.2 `modules/bedrock/tests/bedrock_iam_least_privilege.tftest.hcl`

検証項目:
- `aws_iam_policy.bedrock_inference` の Statement に `bedrock:InvokeModel` と `bedrock:Converse` のみ含まれる
- `bedrock:*` や `bedrock:InvokeModelWithResponseStream` が含まれない
- `Resource` が `arn:aws:bedrock:*::foundation-model/anthropic.claude-haiku-4-5-20251001-v1:0` と Inference Profile ARN の 2 つに限定されている
- 他モデル ARN を含まない

### 5.3 `modules/observability/tests/cloudwatch_alarms.tftest.hcl`

検証項目:
- `aws_cloudwatch_metric_alarm.place_order_p95_breach.threshold == 3000`
- `aws_cloudwatch_metric_alarm.bedrock_retry_burst.threshold == 5`
- `aws_cloudwatch_metric_alarm.fallback_triggered_burst.threshold == 3`
- 各 alarm の `alarm_actions` に SNS Topic ARN が含まれる
- `aws_sns_topic.alarms.name == "gp-dev-alarms"`
- AWS Budgets の `limit_amount == "5"` / 通知 80% + 100% の 2 つを設定

---

## 6. デプロイ手順

### 6.1 初回 apply 手順

1. **`infra/envs/dev/locals.tf` の `alarm_email` を実メールアドレスに変更**（or `terraform.tfvars` で override 検討）
2. `cd infra/envs/dev && terraform init`（Unit C 3 module の初期化）
3. `terraform plan`（既存リソース変更なし、新規リソース追加のみであることを確認）
4. `terraform apply`
5. **SNS 購読確認メール承認**（5 分以内）
   - 受信メール本文の `Confirm subscription` リンクをクリック
6. AWS Console で Bedrock モデルアクセス申請（リージョン `ap-northeast-1`）が完了済みであることを確認
   - Bedrock Console → Model access → Anthropic Claude Haiku 4.5 を有効化
7. `terraform output order_history_table_name` で DynamoDB テーブル名を確認
8. CodePipeline で API Lambda の最新 image を deploy（Code Generation 完了後）

### 6.2 デプロイ手順（Code 変更時）

| 変更箇所 | デプロイ方法 |
|---|---|
| Terraform リソース変更 | `terraform plan` → `terraform apply` |
| API Lambda Go コード変更 | git push → GitHub Actions or 手動で CodePipeline 起動 |
| Frontend Next.js 変更 | git push → Amplify Hosting 自動デプロイ |

---

## 7. コスト見積（月額、デモ規模）

| サービス | 内訳 | 月額（USD） |
|---|---|---|
| DynamoDB OrderHistory | プロビジョンド 1 RCU + 1 WCU + AES256 + PITR 無効 | ~$0.50 |
| DynamoDB ストレージ | TTL 90 日 / 想定 27,000 records × 2KB = 54MB | ~$0.01 |
| Bedrock Claude Haiku 4.5 | NFRC-C20 で月 $1.08 想定（負荷 10 倍で $10.80） | $1.08 ~ $10.80 |
| CloudWatch Logs | API Lambda ログ + metric filter 抽出 | 無料枠内（5GB/月） |
| CloudWatch Alarms | 3 alarms × $0.10 / month | $0.30 |
| SNS Topic | 標準トピック / email 購読 | 無料枠内（1,000 通知 / 月） |
| AWS Budgets | 2 budget actions / month | 無料枠内（最初の 2 つの予算は無料） |
| **合計** | | **$1.89 〜 $11.61 / 月** |

Unit A + Unit B 既存コストとの合算で、本 MVP 全体は月 $5〜$30 程度の見込み（Bedrock コスト次第）。

---

## 8. トレーサビリティ

| Logical Component | AWS リソース | Terraform 配置 |
|---|---|---|
| LC-ORDER-04 OrderHistoryRepository | DynamoDB `gp-{env}-order-history` | `modules/order_history/main.tf` |
| LC-ORDER-05 BedrockAdapter | IAM Policy `gp-{env}-bedrock-inference-policy` | `modules/bedrock/main.tf` |
| LC-ORDER-12 LatencyMiddleware | CloudWatch Logs（Unit A 既存 Log Group） + metric filter | `modules/observability/main.tf` |
| LC-ORDER-13 MeasureHelper | （ログ出力先のみ、リソースなし） | — |
| LC-ORDER-14 BedrockClientInit | IAM Policy で region 指定 | `modules/bedrock/main.tf` |
| LC-ORDER-15 DynamoClientInit | IAM Policy で table ARN 指定 | `modules/order_history/main.tf` |
| NFRC-C13-1 p95 alarm | CloudWatch metric filter + Alarm | `modules/observability/main.tf` §3.3.2/§3.3.3 |
| NFRC-C13-2 retry alarm | 同上 | 同上 |
| NFRC-C13-3 fallback alarm | 同上 | 同上 |
| NFRC-C20 月次予算 $5 | AWS Budgets | `modules/observability/main.tf` §3.3.4 |

---

## 9. NFR 達成根拠

| NFR | 達成根拠 |
|---|---|
| NFRC-C01 (E2E p95 3.0s / p99 5.0s) | DynamoDB プロビジョンド 1 RCU/1 WCU + Bedrock Inference Profile + Lambda 256MB / arm64（Unit A 既存） |
| NFRC-C02 (GetHistory P95 500ms) | DynamoDB Query（PK / SK インデックス）、TTL で古いデータ自動排除 |
| NFRC-C05 (冪等性 TTL 24h) | Unit B IdempotencyKeys テーブル（既存）+ Unit C で Wallet payload から OrderID 復元 |
| NFRC-C12 (構造化ログ 19 項目) | `modules/lambda_api/` 既存（Unit A）の `slog` 出力先 = CloudWatch Logs |
| NFRC-C13 (CloudWatch アラーム 3 種) | `modules/observability/` で metric filter 3 + alarm 3 + SNS 通知 |
| NFRC-C18 (Lambda 256MB / arm64 / 600ms cold) | Unit A 既存 Lambda 設定（Unit C で変更不要、CodePipeline で同一 Lambda にデプロイ） |
| NFRC-C20 (Bedrock 月次予算 $5) | AWS Budgets で 80% 警告 + 100% 警報 |
| NFRC-C24 (Bedrock 本文ログ非記録) | Code Generation 段階で `LogSummary` struct に Bedrock 本文フィールドを含めない（構造的防御） |

---

## 10. 整合性メモ

### 10.1 凍結契約整合

- DynamoDB OrderHistory の PK/SK/TTL 属性名（`expiresAt`）は unit-interfaces.md §3.2 で凍結済みの定義に整合
- IAM Policy のリソース ARN 制限により、unit-interfaces.md の Bedrock モデル指定（NFRC-C20）と整合
- `OrderHistory` GSI なし（凍結契約整合修正で `GSI_IdempotencyKey` を撤回済み）

### 10.2 後続 Unit / 本番化時の対応

- **Unit D `suggest`**: `modules/bedrock/` の Policy ARN を Unit C と同様に attach 可能（横串再利用）。`modules/observability/` の SNS Topic ARN も Unit D アラームから再利用可能
- **Unit E `metrics`**: Unit B / C の DynamoDB テーブルを読取権限のみで参照する Policy を追加する想定（Unit E Infrastructure Design で確定）
- **本番化時**:
  - DynamoDB PITR 有効化（Q-I8 再検討）
  - Customer Managed KMS key（Q-I9 再検討、NFR-COMP-03）
  - AWS Budgets の予測ベース予算追加（Q-I7 = C への昇格検討）
  - Bedrock Provisioned Throughput 検討（NFRC-C11 / Q-N3 = B で MVP スコープ外）

### 10.3 セキュリティ観点

- **最小権限原則**: DynamoDB は CRUD 権限のみ（Scan / DeleteItem / BatchWrite なし）、Bedrock は InvokeModel + Converse のみ（リトライ系・モデル管理系なし）
- **PII 防御**: `LogSummary` struct（NFR Design P-OBS-02）に Bedrock 本文を含めない構造的防御は Code Generation 段階で実装、Infrastructure 段階では metric filter pattern が `event` フィールドのみ参照する設計で漏洩を補完
- **暗号化**: DynamoDB AWS マネージド AES256（Unit B 統一）、SNS は AWS マネージド暗号化（デフォルト）

---

## 11. 文書管理

- **承認**: ユーザ承認待ち（Construction フェーズの per-unit ループ承認ゲート）
- **凍結契約への影響**: なし（DynamoDB スキーマは unit-interfaces.md §3.2 整合、IAM Policy / CloudWatch Alarms / SNS / Budgets は内部仕様）
- **次ステージ**: Code Generation（per-unit、Comprehensive 深度）
