mock_provider "aws" {}

variables {
  env                         = "dev"
  api_gateway_execution_arn   = "arn:aws:execute-api:ap-northeast-1:000000000000:abcdef"
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

run "outputs_present" {
  command = plan

  assert {
    condition     = output.api_lambda_function_name != null
    error_message = "api_lambda_function_name should not be null"
  }
  assert {
    condition     = output.api_lambda_role_arn != null
    error_message = "api_lambda_role_arn should not be null"
  }
  assert {
    condition     = output.ecr_repository_url != null
    error_message = "ecr_repository_url should not be null"
  }
  assert {
    condition     = output.codepipeline_name != null
    error_message = "codepipeline_name should not be null"
  }
}

run "api_lambda_uses_bootstrap_image_initially" {
  command = plan

  assert {
    condition     = endswith(aws_lambda_function.api.image_uri, ":bootstrap")
    error_message = "api_lambda.image_uri should initially point to ECR :bootstrap tag (Q-I15)"
  }
}
