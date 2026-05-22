# 主要 output が空文字でないことを確認

mock_provider "aws" {}

variables {
  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = "test-owner"
  github_repo_name = "goro2pay"
  github_branch    = "develop"
}

run "outputs_are_present" {
  command = plan

  assert {
    condition     = output.user_pool_id != null
    error_message = "user_pool_id output should not be null"
  }
  assert {
    condition     = output.api_id != null
    error_message = "api_id output should not be null"
  }
  assert {
    condition     = output.cognito_authorizer_id != null
    error_message = "cognito_authorizer_id output should not be null"
  }
  assert {
    condition     = output.api_lambda_role_arn != null
    error_message = "api_lambda_role_arn output should not be null"
  }
  assert {
    condition     = output.amplify_app_id != null
    error_message = "amplify_app_id output should not be null"
  }
  assert {
    condition     = output.ecr_repository_url != null
    error_message = "ecr_repository_url output should not be null"
  }
}
