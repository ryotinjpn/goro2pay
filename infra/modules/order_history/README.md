# `modules/order_history`

Unit C `order` の DynamoDB OrderHistory テーブルとアクセス用 IAM Policy を提供する module。

## リソース

- `aws_dynamodb_table.order_history`: PK/SK + TTL 90 日 + AES256 暗号化 + PITR 無効
- `aws_iam_policy.dynamodb_order_history`: API Lambda Role に attach する CRUD policy (PutItem / GetItem / Query / UpdateItem のみ)

## Outputs

| Output | 用途 |
|---|---|
| `dynamodb_table_name` | API Lambda 環境変数 `ORDER_HISTORY_TABLE_NAME` |
| `dynamodb_table_arn` | 監視・他モジュールからの参照 |
| `dynamodb_policy_arn` | `modules/lambda_api` に `additional_policy_arns` で渡して attach |

## 仕様根拠

- 凍結契約 §3.2 (PK/SK、TTL 属性 `expiresAt`)
- Q-I2 = A (プロビジョンド 1 RCU / 1 WCU、Unit B 統一)
- Q-I8 = A (PITR 無効、本番化時に再検討)
- Q-I9 = A (AES256、Unit B 統一)
- NFRC-C12 (TTL 90 日)
