# Routes + Lambda integration
# Unit A 範囲: Logout (JWT 必須) + Health (認証不要)
# 他 Unit (B/C/D/E) の route は各 Unit Construction で route 単位の
# resource を追加する (本 module 内で integration を再利用する)。

resource "aws_apigatewayv2_integration" "api_lambda" {
  api_id                 = aws_apigatewayv2_api.main.id
  integration_type       = "AWS_PROXY"
  integration_uri        = var.api_lambda_invoke_arn
  payload_format_version = "2.0"
  # NOTE: AWS_PROXY (Lambda) integration では integration_method は AWS API 側で
  # 無視される (Lambda invoke は常に POST)。冗長指定を排し DRY を保つ。
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

# ----------------------------------------------------------------------------
# Unit C: POST /api/orders / GET /api/orders (NFRC-C01 / 凍結契約 §4.1)
# integration は既存 api_lambda を再利用 (single Lambda、route 切替で済むため)。
# ----------------------------------------------------------------------------

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

# ----------------------------------------------------------------------------
# Unit B: GET /api/wallet / POST /api/wallet/budget (凍結契約 §3.3)
# 既存 api_lambda integration を再利用 (single Lambda、route 切替で済む)。
# 全 Unit (A/B/C/D/E) の route が同一 `aws_apigatewayv2_integration.api_lambda`
# を経由して同じ API Lambda function に転送される構造 (Code Review Minor 10)。
# ----------------------------------------------------------------------------

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
