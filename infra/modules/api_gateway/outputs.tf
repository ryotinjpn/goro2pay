output "api_id" {
  description = "API Gateway HTTP API ID。他 Unit が後続で route を追加する際に参照"
  value       = aws_apigatewayv2_api.main.id
}

output "api_endpoint" {
  description = "API Gateway 公開 URL。Frontend Amplify env (server-only API_ENDPOINT) で利用"
  value       = aws_apigatewayv2_api.main.api_endpoint
}

output "api_execution_arn" {
  description = "API Gateway execution ARN (lambda_permission の source_arn 用)"
  value       = aws_apigatewayv2_api.main.execution_arn
}

output "cognito_authorizer_id" {
  description = "JWT Authorizer ID。他 Unit の route の authorizer_id"
  value       = aws_apigatewayv2_authorizer.cognito.id
}

output "api_lambda_integration_id" {
  description = "API Lambda 共通 integration ID。他 Unit の route の target で再利用"
  value       = aws_apigatewayv2_integration.api_lambda.id
}
