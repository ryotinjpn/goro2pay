# API Lambda の image_uri が ECR :bootstrap タグを参照していること、
# token validity が A-NFR-SEC-03 の値であることを確認

mock_provider "aws" {}

variables {
  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = "test-owner"
  github_repo_name = "goro2pay"
  github_branch    = "develop"
}

run "api_lambda_uses_bootstrap_image_initially" {
  command = plan

  assert {
    condition     = endswith(aws_lambda_function.api.image_uri, ":bootstrap")
    error_message = "api_lambda.image_uri should initially point to ECR :bootstrap tag (Q-I15)"
  }
}

run "token_validity_matches_a_nfr_sec_03" {
  command = plan

  assert {
    condition = (
      aws_cognito_user_pool_client.web.id_token_validity == 8 &&
      aws_cognito_user_pool_client.web.access_token_validity == 8 &&
      aws_cognito_user_pool_client.web.refresh_token_validity == 30
    )
    error_message = "Token validity must be IdToken/AccessToken=8h, RefreshToken=30d (A-NFR-SEC-03 / Q-N1=B)"
  }
}
