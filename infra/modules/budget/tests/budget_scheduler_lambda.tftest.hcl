# Scheduler Lambda + EventBridge Scheduler の設定検証 (infrastructure-design.md §3.2)。

mock_provider "aws" {}
mock_provider "archive" {}

variables {
  env = "dev"
}

run "lambda_runtime_and_size" {
  command = plan

  assert {
    condition     = aws_lambda_function.scheduler.runtime == "provided.al2023"
    error_message = "runtime must be provided.al2023 (Go custom)"
  }
  assert {
    condition     = aws_lambda_function.scheduler.handler == "bootstrap"
    error_message = "handler must be bootstrap (Go Lambda 標準)"
  }
  assert {
    condition     = contains(aws_lambda_function.scheduler.architectures, "arm64")
    error_message = "architecture must be arm64 (Q-Cost / Unit A 統一)"
  }
  assert {
    condition     = aws_lambda_function.scheduler.memory_size == 128
    error_message = "memory_size must be 128MB (Q-N9=A)"
  }
  assert {
    condition     = aws_lambda_function.scheduler.timeout == 30
    error_message = "timeout must be 30s (Q-N3=A)"
  }
}

run "schedule_cron" {
  command = plan

  assert {
    condition     = aws_scheduler_schedule.monthly_reset.schedule_expression == "cron(0 15 L * ? *)"
    error_message = "schedule must be cron(0 15 L * ? *) UTC = 月末最終日 0:00 JST (Q-B8=A)"
  }
  assert {
    condition     = aws_scheduler_schedule.monthly_reset.flexible_time_window[0].mode == "OFF"
    error_message = "flexible_time_window must be OFF for strict timing"
  }
  assert {
    condition     = aws_scheduler_schedule.monthly_reset.target[0].retry_policy[0].maximum_retry_attempts == 0
    error_message = "retry must be 0 (Q-I5=A: ERROR ログのみ)"
  }
}

run "log_group_retention" {
  command = plan

  assert {
    condition     = aws_cloudwatch_log_group.scheduler.retention_in_days == 7
    error_message = "log retention must be 7 days (Unit A 統一)"
  }
}
