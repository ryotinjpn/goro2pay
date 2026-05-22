# Routes + Lambda integration
# Unit A 範囲: Logout (JWT 必須) + Health (認証不要)
# 他 Unit (B/C/D/E) の route は各 Unit Construction で route 単位の
# resource を追加する (本 module 内で integration を再利用する)。

resource "aws_apigatewayv2_integration" "api_lambda" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_uri        = var.api_lambda_invoke_arn
  payload_format_version = "2.0"
  integration_method     = "POST"
}

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
