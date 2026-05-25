# `infra/modules/budget` — Unit B (`budget` / ダメ予算) Infrastructure

Unit B 専用の Terraform モジュール。DynamoDB 4 テーブル + Scheduler Lambda + EventBridge Scheduler + IAM (最小権限) を提供する。

## 提供リソース

| 種別 | 名前 | 用途 |
|---|---|---|
| DynamoDB | `gp-{env}-wallet` | 残高保持 (PK: userId) |
| DynamoDB | `gp-{env}-budget-settings` | 月間予算 + 増額履歴 (PK: userId) |
| DynamoDB | `gp-{env}-idempotency-keys` | Deduct 冪等性 (PK: key, TTL: expiresAt 24h) |
| DynamoDB | `gp-{env}-budget-reset-log` | 月初リセット履歴 (PK: resetDate, SK: userId) |
| Lambda | `gp-{env}-scheduler-fn` | 月初リセット (provided.al2023, arm64, 128MB, 30s) |
| EventBridge | `gp-{env}-monthly-reset` | cron(0 15 L * ? *) UTC = 月末最終日 0:00 JST |
| IAM Role | `gp-{env}-scheduler-role` | Scheduler Lambda 実行 (最小権限) |
| IAM Role | `gp-{env}-eventbridge-scheduler-role` | EventBridge Scheduler の Lambda invoke |
| IAM Policy | `gp-{env}-budget-dynamodb-policy` | API Lambda attach 用 DynamoDB CRUD (output) |
| CloudWatch Log Group | `/aws/lambda/gp-{env}-scheduler-fn` | Scheduler Lambda ログ (retention 7 日) |

## 入出力

### Variables

| 変数 | 型 | デフォルト | 説明 |
|---|---|---|---|
| `env` | string | (必須) | 環境識別子 (dev/stg/prd) |
| `region` | string | `ap-northeast-1` | AWS region |

### Outputs

- 4 テーブルの `_table_name` / `_table_arn`
- `scheduler_lambda_function_name` / `scheduler_lambda_arn`
- `dynamodb_policy_arn` ← envs/dev で `module.lambda_api.additional_policy_arns` に渡す

## デプロイ前提

`terraform apply` の **前** に Scheduler Lambda の bootstrap バイナリを生成する必要がある:

```bash
make -C apps/api/cmd/scheduler build
```

`apps/api/cmd/scheduler/bootstrap` (Linux arm64) が生成される。これが `archive_file.scheduler` で zip 化されて Lambda にデプロイされる。

## terraform-test (Q-I13=C)

`tests/` 配下の 4 件の tftest で `mock_provider` を使ったオフライン検証:

- `budget_basic.tftest.hcl` — plan が成立すること
- `budget_dynamodb_tables.tftest.hcl` — 4 テーブルの設定値検証
- `budget_scheduler_lambda.tftest.hcl` — Lambda + Schedule の設定値検証
- `budget_iam_least_privilege.tftest.hcl` — IAM 最小権限の検証

```bash
cd infra/modules/budget && terraform init -backend=false && terraform test
```

## 設計成果物

- [Plan](../../../aidlc-docs/construction/plans/budget-code-generation-plan.md)
- [Infrastructure Design](../../../aidlc-docs/construction/budget/infrastructure-design/infrastructure-design.md)
- [NFR Design](../../../aidlc-docs/construction/budget/nfr-design/)
- [Functional Design](../../../aidlc-docs/construction/budget/functional-design/)
- [凍結 IF (unit-interfaces.md)](../../../aidlc-docs/construction/interfaces/unit-interfaces.md)
