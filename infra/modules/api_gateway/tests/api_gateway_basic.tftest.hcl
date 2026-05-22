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

run "outputs_present" {
  command = plan

  assert {
    condition     = output.api_id != null
    error_message = "api_id output should not be null"
  }
  assert {
    condition     = output.cognito_authorizer_id != null
    error_message = "cognito_authorizer_id output should not be null"
  }
  assert {
    condition     = output.api_lambda_integration_id != null
    error_message = "api_lambda_integration_id output should not be null"
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

run "authorizer_ttl_matches_q_i8" {
  command = plan

  assert {
    condition     = aws_apigatewayv2_authorizer.cognito.authorizer_result_ttl_in_seconds == 60
    error_message = "authorizer_result_ttl_in_seconds must be 60 (Q-I8=C)"
  }
}
