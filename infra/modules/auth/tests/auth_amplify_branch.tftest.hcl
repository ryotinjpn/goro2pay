# Amplify branch の env vars が必要キーを含むことを確認 (Q-I14 / BFF パターン)

mock_provider "aws" {}

variables {
  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = "test-owner"
  github_repo_name = "goro2pay"
  github_branch    = "develop"
}

run "amplify_branch_has_required_env_vars" {
  command = plan

  assert {
    condition = contains(
      keys(aws_amplify_branch.develop.environment_variables),
      "AMPLIFY_MONOREPO_APP_ROOT"
    )
    error_message = "AMPLIFY_MONOREPO_APP_ROOT is required for monorepo build"
  }
  assert {
    condition = (
      aws_amplify_branch.develop.environment_variables["AMPLIFY_MONOREPO_APP_ROOT"] == "web"
    )
    error_message = "AMPLIFY_MONOREPO_APP_ROOT must be 'web' (Inception §4.1)"
  }
  assert {
    condition = contains(
      keys(aws_amplify_branch.develop.environment_variables),
      "API_ENDPOINT"
    )
    error_message = "API_ENDPOINT (server-only) must be present for BFF Route Handler"
  }
  assert {
    condition = !contains(
      keys(aws_amplify_branch.develop.environment_variables),
      "NEXT_PUBLIC_API_ENDPOINT"
    )
    error_message = "NEXT_PUBLIC_API_ENDPOINT must NOT exist (BFF パターン: API URL 秘匿化)"
  }
  assert {
    condition = contains(
      keys(aws_amplify_branch.develop.environment_variables),
      "NEXT_PUBLIC_USER_POOL_ID"
    )
    error_message = "NEXT_PUBLIC_USER_POOL_ID is required for Browser-side Amplify Auth"
  }
}
