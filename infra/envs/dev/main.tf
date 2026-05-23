# dev env: 機能別 module を組み合わせて Auth Unit + 横串インフラを構築する。
# unit-of-work.md §4.1 の機能別 module 構成
# (codestar_connection / cognito / api_gateway / lambda_api / amplify) に準拠。

# CodeStar Connection (CodePipeline と Amplify で共有、env ごとに 1 つ)
# 初回 apply 後は AWS Console で手動承認が必要 (deployment-runbook.md §4)。
module "codestar_connection" {
  source = "../../modules/codestar_connection"
  env    = local.env
}

# Cognito (Auth Unit 所有)
module "cognito" {
  source = "../../modules/cognito"
  env    = local.env
  region = local.region
}

# API Gateway (Unit 横串)
module "api_gateway" {
  source                      = "../../modules/api_gateway"
  env                         = local.env
  region                      = local.region
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
  api_lambda_invoke_arn       = module.lambda_api.api_lambda_invoke_arn
}

# Lambda API (Unit 横串、API Lambda + ECR + CodePipeline + CodeBuild)
module "lambda_api" {
  source                      = "../../modules/lambda_api"
  env                         = local.env
  region                      = local.region
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
  codestar_connection_arn     = module.codestar_connection.connection_arn
  github_owner                = local.github_owner
  github_repo_name            = local.github_repo
  github_branch               = local.github_branch
}

# API Gateway → API Lambda invoke 許可
# 両 module の output を必要とするため、循環依存を避けるため envs 側に置く。
#
# source_arn は `${execution_arn}/*/*` (= 全 stage / 全 route から invoke 可) と
# 広めに設定する。Auth Unit の Logout / Health route だけなら
# `${execution_arn}/*/POST/api/auth/logout` 等に絞れるが、後続 Unit B/C/D/E が
# 同じ API Lambda を再利用して新 route を追加するため、route ごとに
# permission を増やす方式は煩雑。Lambda 自体は IAM Role で個別権限を絞っており、
# API Gateway → Lambda 経路は同一 AWS アカウント内で閉じているため、本 MVP では
# stage / route ワイルドカードで許容する。
resource "aws_lambda_permission" "apigw_invoke_api" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.lambda_api.api_lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${module.api_gateway.api_execution_arn}/*/*"
}

# Amplify Hosting (Unit 横串、Frontend 配信)
module "amplify" {
  source                      = "../../modules/amplify"
  env                         = local.env
  region                      = local.region
  github_owner                = local.github_owner
  github_repo_name            = local.github_repo
  github_branch               = local.github_branch
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
  api_endpoint                = module.api_gateway.api_endpoint
  amplify_yml_path            = "${path.root}/../../../web/amplify.yml"
  codestar_connection_arn     = module.codestar_connection.connection_arn
  # 現状 Amplify 側では Console 手動接続運用のため connection_arn 自体は
  # tf resource では使わず、変数受け口だけ揃える (将来 Console 操作レス化時用)。
}
