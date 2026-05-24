# Unit C — Infrastructure Summary

**Generated**: 2026-05-24
**Stage**: Construction / Code Generation
**Unit**: C (`order` + 横串 observability)

## 生成ファイル

### 新規 module 3 種

| Module | ファイル数 | 主リソース |
|---|---|---|
| `infra/modules/order_history/` | 6 (main / locals / variables / outputs / README + tftest) | DynamoDB OrderHistory + IAM Policy |
| `infra/modules/bedrock/` | 7 (+ data.tf) | Bedrock IAM Policy (Claude 3.5 Haiku 限定) |
| `infra/modules/observability/` | 7 (+ data.tf) | SNS + Metric Filter ×3 + Alarm ×3 + Budgets |

### 既存 module への追記

| ファイル | 変更内容 |
|---|---|
| `infra/modules/api_gateway/routes.tf` | `POST /api/orders` / `GET /api/orders` route 追加 |
| `infra/modules/lambda_api/variables.tf` | `additional_policy_arns` / `order_history_table_name` 変数追加 |
| `infra/modules/lambda_api/iam.tf` | `aws_iam_role_policy_attachment.additional` (`for_each`) 追加 |
| `infra/modules/lambda_api/api_lambda.tf` | environment.variables に `BEDROCK_INFERENCE_PROFILE_ID` + `ORDER_HISTORY_TABLE_NAME` 追加 |
| `infra/modules/lambda_api/outputs.tf` | `api_log_group_name` output 追加 |

### envs/dev/

| ファイル | 変更内容 |
|---|---|
| `infra/envs/dev/locals.tf` | `alarm_email = "alerts@example.com"` 追加 |
| `infra/envs/dev/main.tf` | `module "order_history"` / `module "bedrock"` / `module "observability"` 呼出追加、`module "lambda_api"` に Unit C パラメータ注入 |
| `infra/envs/dev/outputs.tf` | `order_history_table_name` / `sns_topic_arn` output 追加 |

## terraform-test 結果

```
tests/dynamodb_schema.tftest.hcl
  4 passed, 0 failed

tests/bedrock_iam_least_privilege.tftest.hcl
  3 passed, 0 failed

tests/cloudwatch_alarms.tftest.hcl
  5 passed, 0 failed
```

合計 **12 tftest 全パス**、`mock_provider` で AWS API 呼出なしのオフラインテスト。

## terraform validate

`infra/envs/dev/` で `terraform validate` 通過 (configuration is valid)。

## Infrastructure Design ↔ コード対応

| Infrastructure Design 項目 | 実装 |
|---|---|
| Q-I1 見直し版 (機能別 3 module) | `modules/order_history/` + `bedrock/` + `observability/` |
| Q-I2 = A (DynamoDB プロビジョンド 1/1) | `modules/order_history/main.tf` |
| Q-I3 = A (Bedrock 最小権限) | `modules/bedrock/main.tf` (InvokeModel + Converse のみ) |
| Q-I4 = C (アラーム閾値外部化) | `modules/observability/variables.tf` (4 つの threshold) |
| Q-I5 = C (横串 SNS Topic) | `modules/observability/main.tf` (`gp-{env}-alarms`) |
| Q-I6 = A (var.alarm_email + email 購読) | `modules/observability/main.tf` |
| Q-I7 = A (Budgets 80% + 100%) | `modules/observability/main.tf` |
| Q-I8 = A (PITR 無効) | `modules/order_history/main.tf` |
| Q-I9 = A (AES256) | `modules/order_history/main.tf` |
| Q-I10 = A (Inference Profile) | `modules/bedrock/locals.tf` |
| Q-I11 = B (既存 routes.tf 拡張) | `modules/api_gateway/routes.tf` |
| Q-I12 改定版 (additional_policy_arns) | `modules/lambda_api/iam.tf` (`for_each`) |
| Q-I13 = C (terraform-test 3 種) | 12 tftest 全パス |

## NFR 達成根拠

| NFR | 実装 |
|---|---|
| NFRC-C13 (Alarms 3 種) | `modules/observability/main.tf` の 3 alarm + 3 metric filter |
| NFRC-C18 (Lambda 256MB / arm64) | Unit A 既存設定 (lambda_api 既存)、Unit C 追加変更なし |
| NFRC-C20 (Bedrock $10/月、Budgets $5 で警告) | `modules/observability` で Budgets 設定 |

## 後続ステージへの引き継ぎ

- 初回 apply 時に SNS 購読確認メール (5 分以内、Unit A の deployment-runbook と同じ手順)
- AWS Console で Bedrock Claude 3.5 Haiku モデルアクセス申請が必要 (deployment-runbook.md §2.1)
- Unit B / D / E は同 `additional_policy_arns` パターンで Policy 追加可能
- 本番環境構築時は `infra/envs/prd/` を `dev` の構造を踏襲して作成、`alarm_email` / `bedrock_budget_limit_usd` を本番値で上書き
