# Unit B: IAM Role / Policy
# - Scheduler Lambda Role + 最小権限 Policy
# - EventBridge Scheduler Role + Lambda Invoke Policy
# - API Lambda 用 DynamoDB CRUD Policy (output で公開、envs/dev で attach)

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
resource "aws_iam_policy" "dynamodb_access" {
  name        = local.dynamodb_access_policy_name
  description = "Unit B: Wallet/BudgetSettings/IdempotencyKeys DynamoDB permissions for API Lambda (least privilege)"

  # Code Review Minor 8 修正: 最小権限化。
  # テーブルごとに必要な action のみ付与:
  #   - wallet:           Get / Put (Create) / Update (Deduct/UpdateBalance/SetBalance)
  #   - budget_settings:  Get / Put / Update (Set with UpdateItem)
  #   - idempotency_keys: Get / Put (TryAcquire) / Update (SaveResponse)
  #   - budget_reset_log: API Lambda は触らない (Scheduler 専用) → 含めない
  # DeleteItem は MVP で使わないため除外。
  # Scan は Wallet ListAllUserIDs が Scheduler 専用のため API Lambda には付与しない。
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
