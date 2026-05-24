output "api_lambda_function_name" {
  description = "API Lambda 関数名"
  value       = aws_lambda_function.api.function_name
}

output "api_lambda_invoke_arn" {
  description = "API Lambda invoke ARN (api_gateway integration target)"
  value       = aws_lambda_function.api.invoke_arn
}

output "api_lambda_execution_arn" {
  description = "API Lambda execution ARN (lambda_permission source_arn 用)"
  value       = aws_lambda_function.api.arn
}

output "api_lambda_role_arn" {
  description = "API Lambda IAM Role ARN。他 Unit が DynamoDB / Bedrock 権限を attach する"
  value       = aws_iam_role.api_lambda.arn
}

output "api_lambda_role_name" {
  description = "API Lambda IAM Role name (aws_iam_role_policy_attachment.role 用)"
  value       = aws_iam_role.api_lambda.name
}

output "ecr_repository_url" {
  description = "ECR Repository URL (bootstrap-ecr-initial.sh で利用)"
  value       = aws_ecr_repository.api.repository_url
}

output "codepipeline_name" {
  description = "CodePipeline 名 (デプロイ状態確認用)"
  value       = aws_codepipeline.api.name
}

# Unit C 追加: 横串 observability module の metric filter source として参照させる。
output "api_log_group_name" {
  description = "API Lambda CloudWatch Log Group 名 (modules/observability の metric filter source)"
  value       = aws_cloudwatch_log_group.api.name
}
