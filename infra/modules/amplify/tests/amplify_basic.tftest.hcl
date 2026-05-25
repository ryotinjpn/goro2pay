mock_provider "aws" {}

variables {
  env                         = "dev"
  region                      = "ap-northeast-1"
  github_owner                = "test-owner"
  github_repo_name            = "goro2pay"
  github_branch               = "develop"
  cognito_user_pool_id        = "ap-northeast-1_TESTPOOL"
  cognito_user_pool_client_id = "test-client-id"
  api_endpoint                = "https://example.execute-api.ap-northeast-1.amazonaws.com"
  # tftest 用ダミー値 (provider の min length 1 validation を回避)。
  # 実環境では envs 側で SSM から実値を注入する。
  github_oauth_token = "dummy-tftest-token"
  # 本 module 単体テスト用の固定 fixture (Frontend ブランチに依存しない)。
  # 配置: infra/modules/amplify/tests/fixtures/amplify.yml
  # `file()` は configuration root 起点なので modules/amplify/ からの相対で渡す。
  amplify_yml_path = "tests/fixtures/amplify.yml"
}

run "plan_succeeds" {
  command = plan
}

run "resources_present" {
  command = plan

  # mock_provider では output (computed attr 経由) は plan 時点で unknown のため、
  # resource の non-computed 属性で存在を確認する。
  assert {
    condition     = aws_amplify_app.web.name == "gp-dev-web"
    error_message = "Amplify App 名が想定と異なる"
  }
  assert {
    condition     = aws_amplify_app.web.platform == "WEB_COMPUTE"
    error_message = "Amplify platform は WEB_COMPUTE (Next.js SSR) であること"
  }
  assert {
    condition     = aws_amplify_branch.develop.branch_name == "develop"
    error_message = "Amplify branch_name は develop であること"
  }
}

run "app_has_monorepo_app_root_env_var" {
  command = plan

  # AMPLIFY_MONOREPO_APP_ROOT は App-level に置く必要がある。
  # Branch-level に置くと framework auto-detection 段階で読まれず、
  # build phase 到達前に "Cannot read 'next' version in package.json" で失敗する。
  assert {
    condition = contains(
      keys(aws_amplify_app.web.environment_variables),
      "AMPLIFY_MONOREPO_APP_ROOT"
    )
    error_message = "AMPLIFY_MONOREPO_APP_ROOT must be set at App-level for monorepo build"
  }
  assert {
    condition = (
      aws_amplify_app.web.environment_variables["AMPLIFY_MONOREPO_APP_ROOT"] == "web"
    )
    error_message = "AMPLIFY_MONOREPO_APP_ROOT must be 'web' (Inception §4.1)"
  }
}

run "branch_has_required_env_vars_bff_pattern" {
  command = plan

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
