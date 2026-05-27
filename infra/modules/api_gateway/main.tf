# API Gateway module: HTTP API + JWT Authorizer + Stage + Routes
# unit-of-work.md §4.1 の `infra/modules/api_gateway/` (Unit 横串) 定義に準拠。
# Authorizer 設定値は cognito module の output を受け取る。
# Lambda integration は lambda_api module の output (invoke_arn) を受け取る。

# ----------------------------------------------------------------------------
# HTTP API (Q-I2=B) + JWT Authorizer (LC-17) + Stage Throttling
# ----------------------------------------------------------------------------

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

  # HTTP API (v2) の JWT Authorizer はキャッシュ非対応 (= TTL は 0 固定)。
  # 0 以外を指定すると CreateAuthorizer が "Cache is not available for JWT
  # authorizer. TTL must be set to 0 to disable cache." で失敗する。
  authorizer_result_ttl_in_seconds = 0
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

# ----------------------------------------------------------------------------
# Routes + Lambda integration
# 全 Unit (A/B/C/D/E) の route が同一 `aws_apigatewayv2_integration.api_lambda`
# を経由して同じ API Lambda function に転送される構造 (Code Review Minor 10)。
# ----------------------------------------------------------------------------

resource "aws_apigatewayv2_integration" "api_lambda" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_uri        = var.api_lambda_invoke_arn
  payload_format_version = "2.0"
  # NOTE: AWS_PROXY (Lambda) integration では integration_method は AWS API 側で
  # 無視される (Lambda invoke は常に POST)。冗長指定を排し DRY を保つ。
}

# Unit A: Logout (JWT 必須) + Health (認証不要)
resource "aws_apigatewayv2_route" "logout" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "POST /api/auth/logout"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

resource "aws_apigatewayv2_route" "health" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /health"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "NONE"
}

# Unit C: POST /api/orders / GET /api/orders (NFRC-C01 / 凍結契約 §4.1)
resource "aws_apigatewayv2_route" "place_order" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "POST /api/orders"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

resource "aws_apigatewayv2_route" "get_order_history" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /api/orders"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

# Unit B: GET /api/wallet / POST /api/wallet/budget (凍結契約 §3.3)
resource "aws_apigatewayv2_route" "get_wallet" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /api/wallet"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

resource "aws_apigatewayv2_route" "post_wallet_budget" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "POST /api/wallet/budget"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

# Unit D: GET /api/suggest (起動時の先回りサジェスト、凍結契約 §5.2)
resource "aws_apigatewayv2_route" "get_suggest" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /api/suggest"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

# Unit E: GET /api/metrics / GET /api/budget/raise/recommendation /
#         POST /api/budget/raise (LC-ME-02 / LC-ME-04)
resource "aws_apigatewayv2_route" "get_metrics" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /api/metrics"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

resource "aws_apigatewayv2_route" "get_budget_raise_recommendation" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "GET /api/budget/raise/recommendation"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}

resource "aws_apigatewayv2_route" "post_budget_raise" {
  api_id             = aws_apigatewayv2_api.main.id
  route_key          = "POST /api/budget/raise"
  target             = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito.id
}
