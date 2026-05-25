# DynamoDB Suggestion スキーマ検証 (Q-DI6)
# mock_provider で AWS API 呼出なしのオフラインテスト。

mock_provider "aws" {}

variables {
  env = "dev"
}

run "dynamodb_table_basic" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.suggestion.name == "gp-dev-suggestion"
    error_message = "table name must be gp-dev-suggestion (Q-I4 命名規則)"
  }

  assert {
    condition     = aws_dynamodb_table.suggestion.hash_key == "suggestionId"
    error_message = "hash_key must be suggestionId (凍結契約 §5.3)"
  }

  assert {
    condition     = aws_dynamodb_table.suggestion.billing_mode == "PROVISIONED"
    error_message = "billing_mode must be PROVISIONED (Q-DI1 / NFRD-D08)"
  }

  assert {
    condition     = aws_dynamodb_table.suggestion.read_capacity == 1
    error_message = "read_capacity must be 1 (NFRD-D08)"
  }

  assert {
    condition     = aws_dynamodb_table.suggestion.write_capacity == 1
    error_message = "write_capacity must be 1 (NFRD-D08)"
  }
}

run "dynamodb_table_ttl_attribute" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.suggestion.ttl[0].attribute_name == "expiresAt"
    error_message = "TTL attribute must be expiresAt (凍結契約 §5.3 / BR-D08)"
  }

  assert {
    condition     = aws_dynamodb_table.suggestion.ttl[0].enabled == true
    error_message = "TTL must be enabled (30 分自動削除)"
  }
}

run "dynamodb_encryption_and_pitr" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.suggestion.server_side_encryption[0].enabled == true
    error_message = "server-side encryption (AES256) must be enabled"
  }

  assert {
    condition     = aws_dynamodb_table.suggestion.point_in_time_recovery[0].enabled == false
    error_message = "PITR must be disabled (一時キャッシュ、本番化時に再検討)"
  }
}

run "iam_policy_least_privilege" {
  command = plan

  assert {
    condition     = aws_iam_policy.dynamodb_suggestion.name == "gp-dev-suggestion-policy"
    error_message = "policy name must follow naming convention"
  }
}
