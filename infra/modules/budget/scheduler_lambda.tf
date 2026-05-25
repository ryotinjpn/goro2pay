# Unit B Scheduler Lambda + EventBridge Scheduler (infrastructure-design.md §3.2)
# - runtime: provided.al2023 (Go custom)
# - architecture: arm64 (Unit A 統一、コスト最適化)
# - memory: 128MB / timeout: 30s (Q-N9=A / Q-N3=A)
# - schedule: cron(0 15 L * ? *) UTC
#     = UTC 月末最終日 15:00 UTC
#     = JST (UTC+9) 翌日 00:00 JST (= JST における月初 0:00)
#   ※ JST 月初 0:00 = UTC 月末日 15:00 UTC のため `(L)` 表記で正しい (Q-B8=A)。

# Scheduler Lambda の bootstrap バイナリを zip 化 (Q-I2=A)。
# 前提: terraform apply 前に `make -C apps/api/cmd/scheduler build` を実行し
# `apps/api/cmd/scheduler/bootstrap` を生成しておく必要がある (Q-I3=A、CI/CD なし)。
#
# var.scheduler_bootstrap_path が空文字列のときはリポジトリ構造を仮定したデフォルト
# パスを使う (path.module から 3 階層上)。CI/CD で workspace 構造が異なる場合は
# 呼び出し側で絶対パスを渡せる (Code Review Minor 12)。
locals {
  bootstrap_path = var.scheduler_bootstrap_path != "" ? var.scheduler_bootstrap_path : "${path.module}/../../../apps/api/cmd/scheduler/bootstrap"
}

data "archive_file" "scheduler" {
  type        = "zip"
  source_file = local.bootstrap_path
  output_path = "${path.module}/.terraform/tmp/scheduler.zip"
}

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
