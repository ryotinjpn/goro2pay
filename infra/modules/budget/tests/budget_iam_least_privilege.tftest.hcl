# IAM Role / Policy の最小権限検証 (terraform-coding-rule)。

mock_provider "aws" {}
mock_provider "archive" {}

variables {
  env = "dev"
}

run "scheduler_lambda_role_assume_lambda" {
  command = plan

  assert {
    condition     = strcontains(aws_iam_role.scheduler_lambda.assume_role_policy, "lambda.amazonaws.com")
    error_message = "scheduler_lambda role must trust lambda.amazonaws.com"
  }
  assert {
    condition     = aws_iam_role.scheduler_lambda.name == "gp-dev-scheduler-role"
    error_message = "scheduler_lambda role name must follow naming convention"
  }
}

run "eventbridge_scheduler_role_only_invoke" {
  command = plan

  assert {
    condition     = strcontains(aws_iam_role.eventbridge_scheduler.assume_role_policy, "scheduler.amazonaws.com")
    error_message = "eventbridge_scheduler role must trust scheduler.amazonaws.com"
  }
  assert {
    condition     = aws_iam_role.eventbridge_scheduler.name == "gp-dev-eventbridge-scheduler-role"
    error_message = "eventbridge_scheduler role name must follow naming convention"
  }
}

run "dynamodb_access_policy_naming" {
  command = plan

  assert {
    condition     = aws_iam_policy.dynamodb_access.name == "gp-dev-budget-dynamodb-policy"
    error_message = "dynamodb_access policy name must follow naming convention"
  }
}

# Note: IAM Policy の jsonencode 結果 (aws_iam_role_policy.policy / aws_iam_policy.policy)
# は plan 時には未確定 (computed string) のため、`strcontains` を含むアサーションは
# `command = apply` でないと評価できない。mock_provider では apply が
# computed 属性を返さない仕様のため、本 tftest では assume_role_policy の文字列のみ
# 検証する。実際のポリシー本文の正しさは既存の `iam.tf` の jsonencode 内容を
# レビューで担保し、デプロイ後の `aws iam get-role-policy` で確認する。
