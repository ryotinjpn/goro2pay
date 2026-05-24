# DynamoDB OrderHistory スキーマ検証 (Q-I13 = C)
# mock_provider で AWS API 呼出なしのオフラインテスト。

mock_provider "aws" {}

variables {
  env = "dev"
}

run "dynamodb_table_basic" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.order_history.name == "gp-dev-order-history"
    error_message = "table name must be gp-dev-order-history (Q-I4 命名規則)"
  }

  assert {
    condition     = aws_dynamodb_table.order_history.hash_key == "PK"
    error_message = "hash_key must be PK (凍結契約 §3.2)"
  }

  assert {
    condition     = aws_dynamodb_table.order_history.range_key == "SK"
    error_message = "range_key must be SK (凍結契約 §3.2)"
  }

  assert {
    condition     = aws_dynamodb_table.order_history.billing_mode == "PROVISIONED"
    error_message = "billing_mode must be PROVISIONED (Q-I2 = A)"
  }

  assert {
    condition     = aws_dynamodb_table.order_history.read_capacity == 1
    error_message = "read_capacity must be 1 (Q-I2 = A)"
  }

  assert {
    condition     = aws_dynamodb_table.order_history.write_capacity == 1
    error_message = "write_capacity must be 1 (Q-I2 = A)"
  }
}

run "dynamodb_table_ttl_attribute" {
  command = plan

  assert {
    condition     = length(aws_dynamodb_table.order_history.ttl) > 0
    error_message = "ttl block must be present"
  }

  assert {
    condition     = aws_dynamodb_table.order_history.ttl[0].attribute_name == "expiresAt"
    error_message = "TTL attribute must be expiresAt (NFRC-C12 / 凍結契約整合)"
  }

  assert {
    condition     = aws_dynamodb_table.order_history.ttl[0].enabled == true
    error_message = "TTL must be enabled"
  }
}

run "dynamodb_encryption_and_pitr" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.order_history.server_side_encryption[0].enabled == true
    error_message = "server-side encryption (AES256) must be enabled (Q-I9 = A)"
  }

  assert {
    condition     = aws_dynamodb_table.order_history.point_in_time_recovery[0].enabled == false
    error_message = "PITR must be disabled (Q-I8 = A、本番化時に再検討)"
  }
}

run "iam_policy_least_privilege" {
  command = plan

  assert {
    condition     = aws_iam_policy.dynamodb_order_history.name == "gp-dev-order-history-policy"
    error_message = "policy name must follow naming convention"
  }
}
