# `infra/modules/api_gateway/` — API Gateway Module (Unit 横串)

API Gateway HTTP API + Cognito JWT Authorizer + Stage Throttling。
Auth Unit 範囲では Logout / Health の route と integration もここに含む。
他 Unit (B/C/D/E) は本 module の `output.api_lambda_integration_id`
と `output.cognito_authorizer_id` を参照して route を envs/ で追加する。

## リソース

- `aws_apigatewayv2_api.main` — HTTP API
  - cors_configuration 未設定 (BFF パターン)
- `aws_apigatewayv2_authorizer.cognito` — JWT Authorizer (TTL 0, HTTP API は cache 非対応)
- `aws_apigatewayv2_stage.default` — `$default` (Throttling 100/200)
- `aws_apigatewayv2_integration.api_lambda` — 共通 integration
- `aws_apigatewayv2_route.logout` — POST /api/auth/logout (JWT)
- `aws_apigatewayv2_route.health` — GET /health (認証不要)

## 利用例

```hcl
module "api_gateway" {
  source                      = "../../modules/api_gateway"
  env                         = "dev"
  region                      = "ap-northeast-1"
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
  api_lambda_invoke_arn       = module.lambda_api.api_lambda_invoke_arn
}
```

## 他 Unit が route を追加する例

Unit B (budget) が envs/dev で route を追加する場合:

```hcl
resource "aws_apigatewayv2_route" "wallet_get" {
  api_id             = module.api_gateway.api_id
  route_key          = "GET /api/wallet"
  target             = "integrations/${module.api_gateway.api_lambda_integration_id}"
  authorization_type = "JWT"
  authorizer_id      = module.api_gateway.cognito_authorizer_id
}
```
