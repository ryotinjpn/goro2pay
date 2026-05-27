# Unit D — Infrastructure Summary

## 新規 module `infra/modules/suggestion/`
- `aws_dynamodb_table.suggestion`: PK `suggestionId` / TTL `expiresAt`(30分) / PROVISIONED 1RCU・1WCU / AES256 / PITR 無効 / GSI なし（凍結契約 §5.3 / NFRD-D08）
- `aws_iam_policy.dynamodb_suggestion`: `PutItem` + `GetItem` のみ（最小権限）
- outputs: `dynamodb_table_name` / `dynamodb_table_arn` / `dynamodb_policy_arn`
- `tests/dynamodb_schema.tftest.hcl`: mock_provider で PK/TTL/capacity/SSE/PITR/命名を assert

## 既存 module への追記
| ファイル | 追記 |
|---|---|
| `api_gateway/routes.tf` | `GET /api/suggest`（JWT、既存 lambda integration 再利用） |
| `lambda_api/variables.tf` | `suggestion_table_name` 変数 |
| `lambda_api/api_lambda.tf` | env `DDB_TABLE_SUGGESTION` 注入（空文字ならスキップ） |
| `envs/dev/main.tf` | `module "suggestion"` + `additional_policy_arns` に suggestion policy + `suggestion_table_name` 配線 |

## 再利用（新規なし）
- `modules/bedrock/`（Bedrock IAM、Unit C/D 共有、既に lambda_api に付与）
- `modules/observability/`（横串、suggest 専用アラームは今回追加せず Q-DI5）
- Lambda 設定（512MB/arm64/30s）は変更なし

## 検証
terraform v1.15.4（公式 zip 導入）: `fmt -recursive` 整形済 / suggestion `terraform test` **4 PASS** / envs/dev `terraform validate` **Success**
