output "wallet_table_name" {
  value       = aws_dynamodb_table.wallet.name
  description = "Wallet table name (環境変数 DDB_TABLE_WALLET 経由で API Lambda に注入)"
}

output "wallet_table_arn" {
  value       = aws_dynamodb_table.wallet.arn
  description = "Wallet table ARN (Unit E の IAM 権限追加時に参照)"
}

output "budget_settings_table_name" {
  value       = aws_dynamodb_table.budget_settings.name
  description = "BudgetSettings table name"
}

output "budget_settings_table_arn" {
  value       = aws_dynamodb_table.budget_settings.arn
  description = "BudgetSettings table ARN"
}

output "idempotency_keys_table_name" {
  value       = aws_dynamodb_table.idempotency_keys.name
  description = "IdempotencyKeys table name (TTL 24h)"
}

output "idempotency_keys_table_arn" {
  value       = aws_dynamodb_table.idempotency_keys.arn
  description = "IdempotencyKeys table ARN"
}

output "budget_reset_log_table_name" {
  value       = aws_dynamodb_table.budget_reset_log.name
  description = "BudgetResetLog table name"
}

output "budget_reset_log_table_arn" {
  value       = aws_dynamodb_table.budget_reset_log.arn
  description = "BudgetResetLog table ARN"
}

output "scheduler_lambda_function_name" {
  value       = aws_lambda_function.scheduler.function_name
  description = "Scheduler Lambda function name (デプロイ確認用)"
}

output "scheduler_lambda_arn" {
  value       = aws_lambda_function.scheduler.arn
  description = "Scheduler Lambda ARN"
}

output "dynamodb_policy_arn" {
  value       = aws_iam_policy.dynamodb_access.arn
  description = "API Lambda attach 用 DynamoDB CRUD Policy ARN (envs/dev で additional_policy_arns に追加)"
}
