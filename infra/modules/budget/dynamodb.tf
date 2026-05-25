# Unit B: DynamoDB テーブル 4 本 (infrastructure-design.md §3.1)
# 全テーブル PROVISIONED 1 RCU / 1 WCU (Q-N4=B)、AES256 (Q-I4=A)、PITR 無効 (Q-I6=A)。

# 3.1.1 Wallet — 残高保持
resource "aws_dynamodb_table" "wallet" {
  name           = local.wallet_table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1

  hash_key = "userId"
  attribute {
    name = "userId"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "budget" }
}

# 3.1.2 BudgetSettings — 月間予算 + 増額履歴
resource "aws_dynamodb_table" "budget_settings" {
  name           = local.budget_settings_table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1

  hash_key = "userId"
  attribute {
    name = "userId"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "budget" }
}

# 3.1.3 IdempotencyKeys — Deduct 冪等性 (TTL 24h)
resource "aws_dynamodb_table" "idempotency_keys" {
  name           = local.idempotency_table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1

  hash_key = "key"
  attribute {
    name = "key"
    type = "S"
  }

  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  server_side_encryption {
    enabled = true
  }
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "budget" }
}

# 3.1.4 BudgetResetLog — 月初リセット履歴 (PK: resetDate, SK: userId)
resource "aws_dynamodb_table" "budget_reset_log" {
  name           = local.budget_reset_log_table_name
  billing_mode   = "PROVISIONED"
  read_capacity  = 1
  write_capacity = 1

  hash_key  = "resetDate"
  range_key = "userId"

  attribute {
    name = "resetDate"
    type = "S"
  }
  attribute {
    name = "userId"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "budget" }
}
