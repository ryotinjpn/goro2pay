# dev env: 4 つの機能別 module を組み合わせて Auth Unit + 横串インフラを構築する。
# unit-of-work.md §4.1 の機能別 module 構成 (cognito / api_gateway / lambda_api / amplify) に準拠。

# CodeStar Connection (CodePipeline と Amplify で共有、envs 側で作成)
resource "aws_codestarconnections_connection" "github" {
  name          = "gp-${local.env}-github-conn"
  provider_type = "GitHub"
  # 注意: 作成直後は Pending 状態。AWS Console で手動承認が必要 (1 度だけ)。
}

# Cognito (Auth Unit 所有)
module "cognito" {
  source = "../../modules/cognito"
  env    = local.env
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
  api_gateway_execution_arn   = module.api_gateway.api_execution_arn
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
  codestar_connection_arn     = aws_codestarconnections_connection.github.arn
  github_owner                = local.github_owner
  github_repo_name            = local.github_repo
  github_branch               = local.github_branch
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
}
