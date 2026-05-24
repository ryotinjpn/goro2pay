output "dynamodb_table_name" {
  value       = aws_dynamodb_table.order_history.name
  description = "OrderHistory table name (環境変数 ORDER_HISTORY_TABLE_NAME 経由で API Lambda に注入)"
}

output "dynamodb_table_arn" {
  value       = aws_dynamodb_table.order_history.arn
  description = "OrderHistory table ARN"
}

output "dynamodb_policy_arn" {
  value       = aws_iam_policy.dynamodb_order_history.arn
  description = "Unit C 用 IAM policy ARN (lambda_api/iam.tf で attach する)"
}
