# Budget module (Unit B): Wallet/BudgetSettings/IdempotencyKeys/BudgetResetLog
# DynamoDB + Scheduler Lambda + EventBridge Scheduler + IAM (最小権限)
# infrastructure-design.md §3.x

# ----------------------------------------------------------------------------
# DynamoDB テーブル 4 本 (infrastructure-design.md §3.1)
# 全テーブル PROVISIONED 1 RCU / 1 WCU (Q-N4=B)、AES256 (Q-I4=A)、PITR 無効 (Q-I6=A)
# ----------------------------------------------------------------------------

# 3.1.1 Wallet — 残高保持
resource "aws_dynamodb_table" "wallet" {
  name           = local.wallet_table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1

  hash_key = "userId"
  attribute {
    name = "userId"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "budget" }
}

# 3.1.2 BudgetSettings — 月間予算 + 増額履歴
resource "aws_dynamodb_table" "budget_settings" {
  name           = local.budget_settings_table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1

  hash_key = "userId"
  attribute {
    name = "userId"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "budget" }
}

# 3.1.3 IdempotencyKeys — Deduct 冪等性 (TTL 24h)
resource "aws_dynamodb_table" "idempotency_keys" {
  name           = local.idempotency_table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1

  hash_key = "key"
  attribute {
    name = "key"
    type = "S"
  }

  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  server_side_encryption {
    enabled = true
  }
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "budget" }
}

# 3.1.4 BudgetResetLog — 月初リセット履歴 (PK: resetDate, SK: userId)
resource "aws_dynamodb_table" "budget_reset_log" {
  name           = local.budget_reset_log_table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1

  hash_key  = "resetDate"
  range_key = "userId"

  attribute {
    name = "resetDate"
    type = "S"
  }
  attribute {
    name = "userId"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "budget" }
}

# ----------------------------------------------------------------------------
# Scheduler Lambda + EventBridge Scheduler (infrastructure-design.md §3.2)
# - runtime: provided.al2023 (Go custom)
# - architecture: arm64 (Unit A 統一、コスト最適化)
# - memory: 128MB / timeout: 30s (Q-N9=A / Q-N3=A)
# - schedule: cron(0 15 L * ? *) UTC
#     = UTC 月末最終日 15:00 UTC
#     = JST (UTC+9) 翌日 00:00 JST (= JST における月初 0:00)
#   ※ JST 月初 0:00 = UTC 月末日 15:00 UTC のため `(L)` 表記で正しい (Q-B8=A)。
# ----------------------------------------------------------------------------

resource "aws_lambda_function" "scheduler" {
  function_name = local.scheduler_function_name
  role          = aws_iam_role.scheduler_lambda.arn

  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]
  memory_size   = 128
  timeout       = 30

  filename         = data.archive_file.scheduler.output_path
  source_code_hash = data.archive_file.scheduler.output_base64sha256

  environment {
    variables = {
      # Scheduler は ResetAll しか呼ばないため Wallet / BudgetSettings /
      # BudgetResetLog のみ env で渡す。Idempotency は API Lambda 専用 (Code
      # Review Important 5)。
      DDB_TABLE_WALLET           = aws_dynamodb_table.wallet.name
      DDB_TABLE_BUDGET_SETTINGS  = aws_dynamodb_table.budget_settings.name
      DDB_TABLE_BUDGET_RESET_LOG = aws_dynamodb_table.budget_reset_log.name
      AWS_LAMBDA_LOG_LEVEL       = "INFO"
      LOG_LEVEL                  = "info"
    }
  }

  depends_on = [aws_cloudwatch_log_group.scheduler]

  tags = { Unit = "budget" }
}

# EventBridge Scheduler が Scheduler Lambda を invoke する許可
resource "aws_lambda_permission" "eventbridge_invoke_scheduler" {
  statement_id  = "AllowEventBridgeSchedulerInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.scheduler.function_name
  principal     = "scheduler.amazonaws.com"
  source_arn    = aws_scheduler_schedule.monthly_reset.arn
}

# 月初リセット Schedule
resource "aws_scheduler_schedule" "monthly_reset" {
  name = local.monthly_reset_schedule_name

  flexible_time_window {
    mode = "OFF"
  }

  # cron(0 15 L * ? *) UTC
  #   = UTC の月末最終日 15:00 UTC
  #   = JST 翌日 00:00 (= JST における月初 0:00) (Q-B8=A)
  # `L` (last day of month) は UTC タイムゾーンで評価されるため、JST と UTC の
  # 月末日が一致する月 (大多数) では JST 月初 0:00 ぴったりに発火する。
  schedule_expression = "cron(0 15 L * ? *)"

  target {
    arn      = aws_lambda_function.scheduler.arn
    role_arn = aws_iam_role.eventbridge_scheduler.arn
    input    = "{}"

    retry_policy {
      maximum_retry_attempts = 0 # Q-I5=A: 再試行なし、ERROR ログのみ
    }
  }
}

# ----------------------------------------------------------------------------
# CloudWatch Log Group (Unit A と同方針 retention 7 日)
# ----------------------------------------------------------------------------

resource "aws_cloudwatch_log_group" "scheduler" {
  name              = local.scheduler_log_group_name
  retention_in_days = 7

  tags = { Unit = "budget" }
}

# ----------------------------------------------------------------------------
# IAM Role / Policy
# - Scheduler Lambda Role + 最小権限 Policy
# - EventBridge Scheduler Role + Lambda Invoke Policy
# - API Lambda 用 DynamoDB CRUD Policy (output で公開、envs/dev で attach)
# ----------------------------------------------------------------------------

# Scheduler Lambda 実行ロール
resource "aws_iam_role" "scheduler_lambda" {
  name = local.scheduler_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = { Unit = "budget" }
}

# Scheduler Lambda 用 inline policy
# 最小権限原則 (terraform-coding-rule):
# - Wallet: GetItem / UpdateItem / Scan
# - BudgetSettings: GetItem
# - BudgetResetLog: PutItem / GetItem
# - Logs: scheduler log group のみ
resource "aws_iam_role_policy" "scheduler_lambda" {
  name = "${local.scheduler_role_name}-policy"
  role = aws_iam_role.scheduler_lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "WalletReadWrite"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:Scan",
        ]
        Resource = aws_dynamodb_table.wallet.arn
      },
      {
        Sid      = "BudgetSettingsRead"
        Effect   = "Allow"
        Action   = ["dynamodb:GetItem"]
        Resource = aws_dynamodb_table.budget_settings.arn
      },
      {
        Sid    = "BudgetResetLogWrite"
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
        ]
        Resource = aws_dynamodb_table.budget_reset_log.arn
      },
      {
        Sid    = "Logs"
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ]
        Resource = [
          aws_cloudwatch_log_group.scheduler.arn,
          "${aws_cloudwatch_log_group.scheduler.arn}:*",
        ]
      },
    ]
  })
}

# EventBridge Scheduler 実行ロール (Lambda invoke 専用)
resource "aws_iam_role" "eventbridge_scheduler" {
  name = local.eventbridge_scheduler_role

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "scheduler.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = { Unit = "budget" }
}

resource "aws_iam_role_policy" "eventbridge_scheduler_invoke" {
  name = "${local.eventbridge_scheduler_role}-invoke"
  role = aws_iam_role.eventbridge_scheduler.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid      = "InvokeSchedulerLambda"
      Effect   = "Allow"
      Action   = "lambda:InvokeFunction"
      Resource = aws_lambda_function.scheduler.arn
    }]
  })
}

# API Lambda 用 DynamoDB CRUD Policy
# Unit C `order_history` と同 pattern: output で arn を公開し、envs/dev で
# `module.lambda_api.additional_policy_arns` に渡す。
#
# Code Review Minor 8 修正: 最小権限化。
# テーブルごとに必要な action のみ付与:
#   - wallet:           Get / Put (Create) / Update (Deduct/UpdateBalance/SetBalance)
#   - budget_settings:  Get / Put / Update (Set with UpdateItem)
#   - idempotency_keys: Get / Put (TryAcquire) / Update (SaveResponse)
#   - budget_reset_log: API Lambda は触らない (Scheduler 専用) → 含めない
# DeleteItem は MVP で使わないため除外。
# Scan は Wallet ListAllUserIDs が Scheduler 専用のため API Lambda には付与しない。
resource "aws_iam_policy" "dynamodb_access" {
  name        = local.dynamodb_access_policy_name
  description = "Unit B: Wallet/BudgetSettings/IdempotencyKeys DynamoDB permissions for API Lambda (least privilege)"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "WalletReadWrite"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
        ]
        Resource = aws_dynamodb_table.wallet.arn
      },
      {
        Sid    = "BudgetSettingsReadWrite"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
        ]
        Resource = aws_dynamodb_table.budget_settings.arn
      },
      {
        Sid    = "IdempotencyReadWrite"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
        ]
        Resource = aws_dynamodb_table.idempotency_keys.arn
      },
    ]
  })

  tags = { Unit = "budget" }
}
