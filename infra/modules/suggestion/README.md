# modules/suggestion (Unit D)

Unit D `suggest` の `GoroPay_Suggestion` 一時保存テーブルと、API Lambda 用の
最小権限 IAM Policy を提供する。

## リソース

| リソース | 内容 |
|---|---|
| `aws_dynamodb_table.suggestion` | PK `suggestionId`(ULID) / TTL `expiresAt`(30分) / PROVISIONED 1RCU・1WCU / AES256 / PITR 無効 / GSI なし |
| `aws_iam_policy.dynamodb_suggestion` | `PutItem` + `GetItem` のみ (SuggestionStore.Save/Get、最小権限) |

## 入出力

| 種別 | 名前 | 説明 |
|---|---|---|
| input | `env` | 環境識別子 (dev/stg/prd) |
| output | `dynamodb_table_name` | env `DDB_TABLE_SUGGESTION` 経由で API Lambda に注入 |
| output | `dynamodb_table_arn` | テーブル ARN |
| output | `dynamodb_policy_arn` | `lambda_api` の `additional_policy_arns` で attach |

## 整合

- 凍結契約 [unit-interfaces.md](../../../aidlc-docs/construction/interfaces/unit-interfaces.md) §5.3 / §10
- NFR Design [logical-components.md](../../../aidlc-docs/construction/suggest/nfr-design/logical-components.md) LC-SUGGEST-04
- Infrastructure Design [infrastructure-design.md](../../../aidlc-docs/construction/suggest/infrastructure-design/infrastructure-design.md) §2

## テスト

`tests/dynamodb_schema.tftest.hcl` (mock_provider、`terraform test` でオフライン検証)。
