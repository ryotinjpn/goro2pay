# Cognito password_policy が A-NFR-SEC-02 の値と一致することを確認

mock_provider "aws" {}

variables {
  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = "test-owner"
  github_repo_name = "goro2pay"
  github_branch    = "develop"
}

run "password_policy_matches_a_nfr_sec_02" {
  command = plan

  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].minimum_length == 8
    error_message = "password_policy.minimum_length must be 8 (A-NFR-SEC-02)"
  }
  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].require_uppercase == true
    error_message = "password_policy.require_uppercase must be true (A-NFR-SEC-02)"
  }
  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].require_lowercase == true
    error_message = "password_policy.require_lowercase must be true (A-NFR-SEC-02)"
  }
  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].require_numbers == true
    error_message = "password_policy.require_numbers must be true (A-NFR-SEC-02)"
  }
  assert {
    condition     = aws_cognito_user_pool.main.password_policy[0].require_symbols == false
    error_message = "password_policy.require_symbols must be false (A-NFR-SEC-02 / Q-A2=B)"
  }
}
