# Plan が mock_provider で成立することを確認 (terraform-test プラグイン規約)

mock_provider "aws" {}

variables {
  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = "test-owner"
  github_repo_name = "goro2pay"
  github_branch    = "develop"
}

run "plan_succeeds" {
  command = plan
}
