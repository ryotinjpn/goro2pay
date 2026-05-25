# budget module の基本 plan 検証 (Q-I13=C / Unit C 同方針)。
# mock_provider で AWS / archive API 呼出なしのオフラインテスト。

mock_provider "aws" {}
mock_provider "archive" {}

variables {
  env = "dev"
}

run "plan_succeeds" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.wallet.name == "gp-dev-wallet"
    error_message = "wallet table name must be gp-dev-wallet"
  }

  assert {
    condition     = aws_lambda_function.scheduler.function_name == "gp-dev-scheduler-fn"
    error_message = "scheduler function name must be gp-dev-scheduler-fn"
  }
}
