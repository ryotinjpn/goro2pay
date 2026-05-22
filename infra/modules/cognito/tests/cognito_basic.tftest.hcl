mock_provider "aws" {}

variables {
  env = "dev"
}

run "plan_succeeds" {
  command = plan
}

run "resources_present" {
  command = plan

  # mock_provider では output (computed attr 経由) は plan 時点で unknown のため、
  # resource の non-computed 属性で存在を確認する。
  assert {
    condition     = aws_cognito_user_pool.main.name == "gp-dev-userpool"
    error_message = "User Pool 名が想定と異なる"
  }
  assert {
    condition     = aws_cognito_user_pool_client.web.name == "gp-dev-appclient-web"
    error_message = "User Pool Client 名が想定と異なる"
  }
}

run "password_policy_matches_a_nfr_sec_02" {
  command = plan

  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].minimum_length == 8
    error_message = "password_policy.minimum_length must be 8 (A-NFR-SEC-02)"
  }
  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].require_uppercase == true
    error_message = "require_uppercase must be true"
  }
  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].require_lowercase == true
    error_message = "require_lowercase must be true"
  }
  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].require_numbers == true
    error_message = "require_numbers must be true"
  }
  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].require_symbols == false
    error_message = "require_symbols must be false (Q-A2=B)"
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
    error_message = "Token validity must be 8h/8h/30d (A-NFR-SEC-03)"
  }
}
