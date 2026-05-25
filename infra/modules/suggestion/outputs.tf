output "dynamodb_table_name" {
  value       = aws_dynamodb_table.suggestion.name
  description = "Suggestion table name (環境変数 DDB_TABLE_SUGGESTION 経由で API Lambda に注入)"
}

output "dynamodb_table_arn" {
  value       = aws_dynamodb_table.suggestion.arn
  description = "Suggestion table ARN"
}

output "dynamodb_policy_arn" {
  value       = aws_iam_policy.dynamodb_suggestion.arn
  description = "Unit D 用 IAM policy ARN (lambda_api の additional_policy_arns で attach する)"
}
