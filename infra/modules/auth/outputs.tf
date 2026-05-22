# Module outputs (Infrastructure Design §4.2、13 種)
# 他 Unit (B/C/D/E) や Frontend が参照する値を公開する

# --- Cognito ---
output "user_pool_id" {
  description = "Cognito User Pool ID。Amplify env や他 Unit Lambda env が利用"
  value       = aws_cognito_user_pool.main.id
}

output "user_pool_arn" {
  description = "Cognito User Pool ARN"
  value       = aws_cognito_user_pool.main.arn
}

output "user_pool_endpoint" {
  description = "Cognito JWT issuer URL (cognito-idp endpoint)"
  value       = aws_cognito_user_pool.main.endpoint
}

output "user_pool_client_id" {
  description = "Cognito App Client ID。Frontend Amplify が利用"
  value       = aws_cognito_user_pool_client.web.id
}

# --- API Gateway ---
output "api_id" {
  description = "API Gateway HTTP API ID。他 Unit が route 追加時に参照"
  value       = aws_apigatewayv2_api.main.id
}

output "api_endpoint" {
  description = "API Gateway 公開 URL。Frontend の API_ENDPOINT (server-only env)"
  value       = aws_apigatewayv2_api.main.api_endpoint
}

output "cognito_authorizer_id" {
  description = "JWT Authorizer ID。他 Unit の route の authorizer_id として参照"
  value       = aws_apigatewayv2_authorizer.cognito.id
}

# --- API Lambda ---
output "api_lambda_function_name" {
  description = "API Lambda 関数名。他 Unit が CodeBuild 等から参照、または route 追加で利用"
  value       = aws_lambda_function.api.function_name
}

output "api_lambda_invoke_arn" {
  description = "API Lambda invoke ARN。他 Unit の integration target"
  value       = aws_lambda_function.api.invoke_arn
}

output "api_lambda_role_arn" {
  description = "API Lambda IAM Role ARN。他 Unit が DynamoDB / Bedrock 権限を attach"
  value       = aws_iam_role.api_lambda.arn
}

# --- Amplify ---
output "amplify_app_id" {
  description = "Amplify App ID。デプロイ確認 / Console URL 用"
  value       = aws_amplify_app.web.id
}

output "amplify_default_domain" {
  description = "Amplify default domain (xxx.amplifyapp.com)。Frontend 公開 URL"
  value       = aws_amplify_app.web.default_domain
}

# --- CD パイプライン関連 ---
output "ecr_repository_url" {
  description = "ECR Repository URL。bootstrap-ecr-initial.sh で利用"
  value       = aws_ecr_repository.api.repository_url
}

output "codepipeline_name" {
  description = "CodePipeline 名。デプロイ状態確認用"
  value       = aws_codepipeline.api.name
}
