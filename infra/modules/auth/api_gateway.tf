# API Gateway HTTP API (Q-I2=B) + JWT Authorizer (LC-17) + Stage Throttling

resource "aws_apigatewayv2_api" "main" {
  name          = "${local.prefix}-api"
  protocol_type = "HTTP"
  # BFF パターン採用のため、ブラウザ直叩きが発生せず CORS 不要 (cors_configuration 未設定)。
  # 本番化時に直叩きユースケースが出てきたら追加する。
}

resource "aws_apigatewayv2_authorizer" "cognito" {
  name             = "${local.prefix}-auth-authorizer"
  api_id           = aws_apigatewayv2_api.main.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]

  jwt_configuration {
    audience = [aws_cognito_user_pool_client.web.id]
    issuer   = "https://cognito-idp.${var.region}.amazonaws.com/${aws_cognito_user_pool.main.id}"
  }

  authorizer_result_ttl_in_seconds = 60 # Q-I8=C
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  default_route_settings {
    # NFR Design A-NFR-SEC-04 / Q-I10 範囲
    throttling_burst_limit   = 200
    throttling_rate_limit    = 100
    detailed_metrics_enabled = false # A-NFR-OBS-02
  }
}
