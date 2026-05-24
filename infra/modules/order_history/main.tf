# Unit C: OrderHistory DynamoDB テーブル + アクセス用 IAM Policy
# 凍結契約 §3.2 / NFRC-C09 / NFRC-C12 / Q-I2 = A / Q-I8 = A / Q-I9 = A 整合

resource "aws_dynamodb_table" "order_history" {
  name         = local.table_name
  billing_mode = "PROVISIONED"
  # Q-I2 = A: プロビジョンド 1 RCU / 1 WCU (Unit B 統一、無料枠内)
  read_capacity  = 1
  write_capacity = 1

  # 凍結契約 §3.2: PK = USER#{userID}, SK = ORDER#{orderedAt}#{orderID}
  hash_key  = "PK"
  range_key = "SK"

  attribute {
    name = "PK"
    type = "S"
  }
  attribute {
    name = "SK"
    type = "S"
  }

  # NFRC-C12 / 凍結契約整合: TTL 属性は expiresAt (90 日)
  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  # Q-I9 = A: AWS マネージド AES256 (Unit B 統一)
  server_side_encryption {
    enabled = true
  }

  # Q-I8 = A: PITR 無効 (dev / ハッカソン、本番化時に再検討)
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "order" }
}

# Unit C: API Lambda Role に attach する DynamoDB CRUD Policy
# terraform-coding-rule の最小権限原則に従い、対象テーブル ARN に限定。
resource "aws_iam_policy" "dynamodb_order_history" {
  name        = local.policy_name
  description = "Unit C: OrderHistory DynamoDB CRUD permissions (least privilege)"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "OrderHistoryCRUD"
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:UpdateItem",
        ]
        Resource = aws_dynamodb_table.order_history.arn
      },
    ]
  })

  tags = { Unit = "order" }
}
