# API Gateway module: HTTP API + JWT Authorizer + Stage + Logout/Health route
# unit-of-work.md §4.1 の `infra/modules/api_gateway/` (Unit 横串) 定義に準拠。
# Authorizer 設定値は cognito module の output を受け取る。
# Lambda integration は lambda_api module の output (invoke_arn) を受け取る。
#
# resource 定義は機能別ファイル (api_gateway.tf / routes.tf) に配置している。
