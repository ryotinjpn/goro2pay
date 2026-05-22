# CodeBuild IAM Role が Lambda UpdateFunctionCode を最小スコープ (本 Lambda のみ) で持つことを確認 (Q-I15)

mock_provider "aws" {}

variables {
  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = "test-owner"
  github_repo_name = "goro2pay"
  github_branch    = "develop"
}

run "codebuild_iam_role_exists" {
  command = plan

  assert {
    condition     = aws_iam_role.codebuild_api.name == "gp-dev-codebuild-role"
    error_message = "CodeBuild IAM Role name must follow naming convention"
  }
}

run "codebuild_iam_policy_exists" {
  command = plan

  assert {
    condition     = aws_iam_role_policy.codebuild_api.name == "policy"
    error_message = "CodeBuild role inline policy must be present"
  }
  assert {
    condition     = aws_iam_role_policy.codebuild_api.role == aws_iam_role.codebuild_api.id
    error_message = "policy must be attached to codebuild role"
  }
}
