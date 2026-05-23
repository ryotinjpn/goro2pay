# `infra/modules/lambda_api/` — API Lambda + CD Module (Unit 横串)

API Lambda 本体 (Go + Gin + LWA, container image) + ECR + CodePipeline + CodeBuild + S3 artifacts + 関連 IAM をまとめた module。

## リソース

- `aws_lambda_function.api` — API Lambda (image_uri = ECR `:bootstrap`、`lifecycle.ignore_changes = [image_uri]`)
- `aws_lambda_permission.apigw_invoke_api` — API Gateway → Lambda invoke 許可
- `aws_ecr_repository.api` + `aws_ecr_lifecycle_policy.api`
- `aws_codepipeline.api` (Source: GitHub via CodeStar Connection / Build: CodeBuild)
- `aws_codebuild_project.api` (ARM_CONTAINER, privileged_mode、buildspec = apps/api/buildspec.yml)
- `aws_s3_bucket.codepipeline_artifacts` (lifecycle 30 日)
- `aws_iam_role.api_lambda` / `codepipeline_api` / `codebuild_api`
- `aws_cloudwatch_log_group.api` / `codebuild_api`

## 利用例

```hcl
module "lambda_api" {
  source                      = "../../modules/lambda_api"
  env                         = "dev"
  api_gateway_execution_arn   = module.api_gateway.api_execution_arn
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
  codestar_connection_arn     = aws_codestarconnections_connection.github.arn
  github_owner                = local.github_owner
  github_repo_name            = local.github_repo
  github_branch               = local.github_branch
}
```

## 他 Unit からの利用

Unit B/C/D/E は本 module の `output.api_lambda_role_arn` / `api_lambda_role_name` に対して権限ポリシーを attach する形で機能を追加する。

```hcl
# Unit B (envs/dev/main.tf 等で)
resource "aws_iam_role_policy_attachment" "wallet_dynamodb" {
  role       = module.lambda_api.api_lambda_role_name
  policy_arn = aws_iam_policy.wallet_dynamodb.arn
}
```

## Outputs

- `api_lambda_function_name`
- `api_lambda_invoke_arn` (api_gateway integration target)
- `api_lambda_execution_arn`
- `api_lambda_role_arn` / `api_lambda_role_name`
- `ecr_repository_url` (bootstrap-ecr-initial.sh で利用)
- `codepipeline_name`
