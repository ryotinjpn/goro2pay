mock_provider "aws" {}

variables {
  env                         = "dev"
  region                      = "ap-northeast-1"
  cognito_user_pool_id        = "ap-northeast-1_TESTPOOL"
  cognito_user_pool_client_id = "test-client-id"
  api_lambda_invoke_arn       = "arn:aws:apigateway:ap-northeast-1:lambda:path/2015-03-31/functions/arn:aws:lambda:ap-northeast-1:000000000000:function:test/invocations"
}

run "plan_succeeds" {
  command = plan
}

run "resources_present" {
  command = plan

  # mock_provider では output (computed attr 経由) は plan 時点で unknown のため、
  # resource の non-computed 属性で存在を確認する。
  assert {
    condition     = aws_apigatewayv2_api.main.name == "gp-dev-api"
    error_message = "HTTP API 名が想定と異なる"
  }
  assert {
    condition     = aws_apigatewayv2_api.main.protocol_type == "HTTP"
    error_message = "API は HTTP API であること"
  }
  assert {
    condition     = aws_apigatewayv2_authorizer.cognito.authorizer_type == "JWT"
    error_message = "Authorizer は JWT であること"
  }
}

run "access_log_enabled" {
  command = plan

  # Authorizer による 401/403 ログを CloudWatch に残す。
  assert {
    condition     = length(aws_apigatewayv2_stage.default.access_log_settings) == 1
    error_message = "access_log_settings が有効化されていること"
  }
}

run "throttling_matches_a_nfr_sec_04" {
  command = plan

  assert {
    condition     = aws_apigatewayv2_stage.default.default_route_settings[0].throttling_rate_limit == 100
    error_message = "throttling_rate_limit must be 100 (A-NFR-SEC-04)"
  }
  assert {
    condition     = aws_apigatewayv2_stage.default.default_route_settings[0].throttling_burst_limit == 200
    error_message = "throttling_burst_limit must be 200 (A-NFR-SEC-04)"
  }
}

run "authorizer_ttl_is_zero" {
  command = plan

  # HTTP API (v2) JWT Authorizer は cache 非対応で TTL=0 固定。
  # Q-I8=C (60秒) は HTTP API 仕様適用不可と実装時に判明。
  assert {
    condition     = aws_apigatewayv2_authorizer.cognito.authorizer_result_ttl_in_seconds == 0
    error_message = "authorizer_result_ttl_in_seconds must be 0 (HTTP API JWT Authorizer は cache 非対応)"
  }
}
