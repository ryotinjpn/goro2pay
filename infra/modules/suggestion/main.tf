# Unit D: Suggestion DynamoDB テーブル + アクセス用 IAM Policy
# 凍結契約 §5.3 / NFRD-D05 / NFRD-D08 / Q-DI1 整合

resource "aws_dynamodb_table" "suggestion" {
  name         = local.table_name
  billing_mode = "PROVISIONED"
  # Q-DI1 / NFRD-D08: プロビジョンド 1 RCU / 1 WCU (Unit B/C 統一、無料枠内)
  read_capacity  = 1
  write_capacity = 1

  # 凍結契約 §5.3: PK = suggestionId (ULID 単独、SK / GSI なし)
  hash_key = "suggestionId"

  attribute {
    name = "suggestionId"
    type = "S"
  }

  # 凍結契約 §5.3 / BR-D08: TTL 属性は expiresAt (30 分)
  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  # AWS マネージド AES256 (Unit B/C 統一、NFR-SEC-02)
  server_side_encryption {
    enabled = true
  }

  # PITR 無効 (dev / ハッカソン、本番化時に再検討。一時キャッシュのため不要)
  point_in_time_recovery {
    enabled = false
  }

  tags = { Unit = "suggest" }
}

# Unit D: API Lambda Role に attach する DynamoDB Policy
# terraform-coding-rule の最小権限原則。SuggestionStore は Save (PutItem) /
# Get (GetItem) のみ使用するため Query / Update / Delete は付与しない。
resource "aws_iam_policy" "dynamodb_suggestion" {
  name        = local.policy_name
  description = "Unit D: Suggestion DynamoDB Put/Get permissions (least privilege)"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "SuggestionPutGet"
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
        ]
        Resource = aws_dynamodb_table.suggestion.arn
      },
    ]
  })

  tags = { Unit = "suggest" }
}
