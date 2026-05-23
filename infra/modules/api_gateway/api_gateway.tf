# API Gateway HTTP API (Q-I2=B) + JWT Authorizer (LC-17) + Stage Throttling

resource "aws_apigatewayv2_api" "main" {
  name          = "${local.prefix}-api"
  protocol_type = "HTTP"
  # BFF パターン採用のため cors_configuration 未設定 (Q-I2=B)
}

resource "aws_apigatewayv2_authorizer" "cognito" {
  name             = "${local.prefix}-auth-authorizer"
  api_id           = aws_apigatewayv2_api.main.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]

  jwt_configuration {
    audience = [var.cognito_user_pool_client_id]
    issuer   = "https://cognito-idp.${var.region}.amazonaws.com/${var.cognito_user_pool_id}"
  }

  authorizer_result_ttl_in_seconds = 60 # Q-I8=C
}

resource "aws_cloudwatch_log_group" "apigw_access" {
  name              = "/aws/apigateway/${local.prefix}-api"
  retention_in_days = 7
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  default_route_settings {
    throttling_burst_limit   = 200
    throttling_rate_limit    = 100
    detailed_metrics_enabled = false
  }

  # Authorizer による 401/403 等、Lambda に到達しないリクエストの調査ログ。
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.apigw_access.arn
    format = jsonencode({
      requestId         = "$context.requestId"
      ip                = "$context.identity.sourceIp"
      requestTime       = "$context.requestTime"
      httpMethod        = "$context.httpMethod"
      routeKey          = "$context.routeKey"
      status            = "$context.status"
      protocol          = "$context.protocol"
      responseLength    = "$context.responseLength"
      authorizerError   = "$context.authorizer.error"
      integrationStatus = "$context.integrationStatus"
    })
  }
}
