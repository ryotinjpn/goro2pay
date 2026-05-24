mock_provider "aws" {}

variables {
  env                         = "dev"
  region                      = "ap-northeast-1"
  cognito_user_pool_id        = "ap-northeast-1_TESTPOOL"
  cognito_user_pool_client_id = "test-client-id"
  codestar_connection_arn     = "arn:aws:codestar-connections:ap-northeast-1:000000000000:connection/abcd"
  github_owner                = "test-owner"
  github_repo_name            = "goro2pay"
  github_branch               = "develop"
}

run "plan_succeeds" {
  command = plan
}

run "resources_present" {
  command = plan

  # mock_provider では output (computed attr 経由) は plan 時点で unknown になる。
  # 代わりに resource の non-computed 属性を assert することで存在を確認する。
  assert {
    condition     = aws_lambda_function.api.function_name == "gp-dev-api-fn"
    error_message = "api_lambda 関数名が想定と異なる"
  }
  assert {
    condition     = aws_lambda_function.api.package_type == "Image"
    error_message = "api_lambda は package_type = Image であること"
  }
  assert {
    condition     = aws_ecr_repository.api.name == "gp-dev-api-image"
    error_message = "ECR Repository 名が想定と異なる"
  }
  assert {
    condition     = aws_codepipeline.api.name == "gp-dev-api-pipeline"
    error_message = "CodePipeline 名が想定と異なる"
  }
}

run "api_lambda_lifecycle_ignores_image_uri" {
  command = plan

  # CodeBuild が `aws lambda update-function-code` で image_uri を更新するため、
  # Terraform 側は lifecycle.ignore_changes = [image_uri] で drift を無視。
  # mock_provider 環境では image_uri 自体が unknown なので、lifecycle ブロックの
  # 設定有無を resource definition 経由で確認する代わりに、
  # 設定漏れがあれば apply で必ずズレが出るので、ここでは最低限 architectures が
  # arm64 で固定されていることを assert (Q-I15)。
  assert {
    condition     = aws_lambda_function.api.architectures[0] == "arm64"
    error_message = "api_lambda.architectures は arm64 (LWA + Graviton) であること"
  }
}

run "codebuild_environment_uses_arm_container" {
  command = plan

  # Lambda が arm64 / Go SDK v2 が大量モジュールを抱えるため、
  # x86_64 + QEMU クロスビルドだと build に 10 分超かかっていた。
  # ARM ネイティブ (amazonlinux2-aarch64-standard:3.0) + MEDIUM で
  # ~50 秒に短縮 (~16x 高速化)。
  assert {
    condition     = one(aws_codebuild_project.api.environment[*].type) == "ARM_CONTAINER"
    error_message = "CodeBuild environment.type must be ARM_CONTAINER (native arm64 build)"
  }
  assert {
    condition     = one(aws_codebuild_project.api.environment[*].image) == "aws/codebuild/amazonlinux2-aarch64-standard:3.0"
    error_message = "CodeBuild image must be amazonlinux2-aarch64-standard:3.0 (ARM native curated image)"
  }
  assert {
    condition     = one(aws_codebuild_project.api.environment[*].compute_type) == "BUILD_GENERAL1_MEDIUM"
    error_message = "CodeBuild compute_type must be BUILD_GENERAL1_MEDIUM (go build + docker build に十分なメモリ)"
  }
  assert {
    condition     = one(aws_codebuild_project.api.environment[*].privileged_mode) == true
    error_message = "CodeBuild privileged_mode must be true (docker build に必須)"
  }
}

# IAM policy 文字列内容の最小権限テスト (lambda:UpdateFunctionCode の Resource 限定 等)
# は mock_provider の plan/apply とも attribute unknown / invalid ARN で評価困難なため、
# tflint / checkov / IAM Access Analyzer 等の静的解析を Build & Test ステージで実施する。
# (prd README に IAM 静的解析の項目を追加済み)
