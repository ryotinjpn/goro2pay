# 4 DynamoDB テーブルの設定検証 (infrastructure-design.md §3.1)。

mock_provider "aws" {}
mock_provider "archive" {}

variables {
  env = "dev"
}

run "wallet_table" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.wallet.billing_mode == "PROVISIONED"
    error_message = "wallet must be PROVISIONED (Q-N4=B)"
  }
  assert {
    condition     = aws_dynamodb_table.wallet.read_capacity == 1 && aws_dynamodb_table.wallet.write_capacity == 1
    error_message = "wallet must be 1 RCU / 1 WCU"
  }
  assert {
    condition     = aws_dynamodb_table.wallet.hash_key == "userId"
    error_message = "wallet hash_key must be userId (凍結 IF §3.4)"
  }
  assert {
    condition     = aws_dynamodb_table.wallet.server_side_encryption[0].enabled == true
    error_message = "wallet must have AES256 encryption (Q-I4=A)"
  }
  assert {
    condition     = aws_dynamodb_table.wallet.point_in_time_recovery[0].enabled == false
    error_message = "wallet PITR must be disabled (Q-I6=A)"
  }
}

run "budget_settings_table" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.budget_settings.hash_key == "userId"
    error_message = "budget_settings hash_key must be userId"
  }
  assert {
    condition     = aws_dynamodb_table.budget_settings.billing_mode == "PROVISIONED"
    error_message = "budget_settings must be PROVISIONED"
  }
}

run "idempotency_table_ttl" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.idempotency_keys.hash_key == "key"
    error_message = "idempotency_keys hash_key must be key"
  }
  assert {
    condition     = length(aws_dynamodb_table.idempotency_keys.ttl) > 0
    error_message = "idempotency_keys must have TTL block"
  }
  assert {
    condition     = aws_dynamodb_table.idempotency_keys.ttl[0].attribute_name == "expiresAt"
    error_message = "TTL attribute_name must be expiresAt (NFR-REL-02 24h)"
  }
  assert {
    condition     = aws_dynamodb_table.idempotency_keys.ttl[0].enabled == true
    error_message = "TTL must be enabled"
  }
}

run "budget_reset_log_composite_key" {
  command = plan

  assert {
    condition     = aws_dynamodb_table.budget_reset_log.hash_key == "resetDate"
    error_message = "budget_reset_log hash_key must be resetDate"
  }
  assert {
    condition     = aws_dynamodb_table.budget_reset_log.range_key == "userId"
    error_message = "budget_reset_log range_key must be userId (CR-B-05 複合キー)"
  }
}
