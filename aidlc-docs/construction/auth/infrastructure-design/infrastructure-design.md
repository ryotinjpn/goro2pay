# Auth Unit — Infrastructure Design

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: A (`auth`)
**Construction Depth**: Standard
**Stage**: Infrastructure Design / Construction
**Predecessors**: NFR Design 完了 (PR #68)

本ドキュメントは Unit A の **Terraform リソース定義** を凍結する。NFR Design で論理化した LC-15/16/17 / LC-07 を実 AWS リソースに展開する。

参照: [auth-infrastructure-design-plan.md](../../plans/auth-infrastructure-design-plan.md), [logical-components.md](../nfr-design/logical-components.md), [nfr-design-patterns.md](../nfr-design/nfr-design-patterns.md), [unit-interfaces.md](../../interfaces/unit-interfaces.md)

---

## 1. スコープと前提

### 1.1 本書のスコープ（Q-I10=A4 確定）

Unit A PR で構築するリソース:

1. **Cognito** (User Pool / App Client) ← LC-15/16
2. **Pre Sign-up Lambda** (Node.js, auto-confirm) ← LC-07
3. **API Gateway HTTP API** (本体) ← Q-I2=B
4. **JWT Authorizer** (Cognito 連携) ← LC-17
5. **`POST /api/auth/logout` route + integration**
6. **API Lambda 本体** (Go + Gin + LWA, 当面は Hello World + Logout のみ)
7. CloudWatch Log Groups（Lambda 用）
8. IAM Roles / Policies（Pre Sign-up Lambda + API Lambda）

### 1.2 スコープ外（他 PR / 後続 Unit が担当）

- Frontend Amplify Hosting → 別 PR
- Unit B/C/D/E のハンドラ実装 + route 追加 → 各 Unit の Construction
- DynamoDB テーブル定義 → 各 Unit の Infrastructure Design
- Bedrock IAM → Unit C の Infrastructure Design

### 1.3 不変前提

| 項目 | 値 | 出典 |
|---|---|---|
| Region | `ap-northeast-1` | NFR-COMP / Q-15 |
| IaC | Terraform | 要件 Q12=C |
| 命名 | `gp-{env}-{resource}` | Q-I4 |
| Tags | `Project=goro2pay` / `Env=dev` / `Unit=auth` / `ManagedBy=terraform` | Q-I11 |
| Module 規約 | terraform-module-design / terraform-coding-rule / terraform-test 準拠 | A-NFR-MAINT-01 |

---

## 2. ディレクトリ構造

```
infra/
├── envs/
│   ├── dev/
│   │   ├── backend.tf            # S3 + use_lockfile (Q-I6)
│   │   ├── providers.tf          # default_tags
│   │   ├── main.tf               # auth module 呼出
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── prd/
│       └── README.md             # placeholder (Q-I5)
├── modules/
│   └── auth/                     # Q-I1=A 単一モジュール
│       ├── README.md
│       ├── main.tf               # メイン定義
│       ├── cognito.tf            # User Pool + App Client
│       ├── pre_signup_lambda.tf  # Pre Sign-up Lambda (Node.js)
│       ├── api_gateway.tf        # HTTP API + Authorizer + Stage
│       ├── api_lambda.tf         # API Lambda (Go + Gin + LWA, Hello World)
│       ├── logout_route.tf       # POST /api/auth/logout route + integration
│       ├── iam.tf                # 各 Lambda 用 IAM Role / Policy
│       ├── log_groups.tf         # CloudWatch Log Groups
│       ├── variables.tf
│       └── outputs.tf
└── lambdas/
    ├── pre-signup/
    │   └── index.js              # 5 行 auto-confirm 実装 (Code Generation)
    └── api/
        └── (Code Generation で配置、Go バイナリ + Dockerfile)
```

`modules/auth/` の内訳は責務別にファイル分割し、1 ファイル ~100 行以内を目安。

---

## 3. リソース詳細

以下、`infra/modules/auth/` 配下に置く Terraform リソースを論理仕様で記述する（HCL の細かいコード本体は Code Generation で書く）。

### 3.1 Cognito (LC-15 / LC-16)

#### 3.1.1 `aws_cognito_user_pool.main`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-userpool` | Q-I4 |
| `username_attributes` | `["email"]` | LC-15 |
| `auto_verified_attributes` | `["email"]` | LC-15 |
| `password_policy.minimum_length` | 8 | A-NFR-SEC-02 |
| `password_policy.require_uppercase` | true | 同上 |
| `password_policy.require_lowercase` | true | 同上 |
| `password_policy.require_numbers` | true | 同上 |
| `password_policy.require_symbols` | false | 同上 |
| `mfa_configuration` | `OFF` | LC-15 |
| `account_recovery_setting.recovery_mechanism` | `verified_email` 1 つだけ宣言（最小設定） | LC-15 |
| `admin_create_user_config.allow_admin_create_user_only` | false | LC-15 |
| `lambda_config.pre_sign_up` | `aws_lambda_function.pre_signup.arn` | LC-15 |
| `email_configuration.email_sending_account` | `COGNITO_DEFAULT` | Q-I7 |
| `deletion_protection` | `INACTIVE` | Q-I13 |

#### 3.1.2 `aws_cognito_user_pool_client.web`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-appclient-web` | Q-I4 |
| `user_pool_id` | `aws_cognito_user_pool.main.id` | — |
| `generate_secret` | false | LC-16 (PWA Public Client) |
| `explicit_auth_flows` | `["ALLOW_USER_PASSWORD_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"]` | LC-16 |
| `id_token_validity` | 8 | A-NFR-SEC-03 |
| `access_token_validity` | 8 | 同上 |
| `refresh_token_validity` | 30 | 同上 |
| `token_validity_units.id_token` | `hours` | LC-16 |
| `token_validity_units.access_token` | `hours` | 同上 |
| `token_validity_units.refresh_token` | `days` | 同上 |
| `prevent_user_existence_errors` | `ENABLED` | A-NFR-SEC-08 (情報漏洩抑止 / R-Err-1) |
| `enable_token_revocation` | true | A-NFR-SEC-03 (GlobalSignOut 用) |

#### 3.1.3 Cognito Lambda Trigger 連携

`aws_lambda_permission.cognito_invoke_pre_signup` で Cognito User Pool が `aws_lambda_function.pre_signup` を呼び出せるよう許可。

```
principal              = "cognito-idp.amazonaws.com"
source_arn             = aws_cognito_user_pool.main.arn
action                 = "lambda:InvokeFunction"
```

### 3.2 Pre Sign-up Lambda (LC-07)

#### 3.2.1 `data.archive_file.pre_signup`

| 設定 | 値 |
|---|---|
| `type` | `zip` |
| `source_dir` | `${path.module}/../../lambdas/pre-signup` |
| `output_path` | `${path.module}/.terraform/tmp/pre-signup.zip` |

Q-I3 = A、Code Generation で `infra/lambdas/pre-signup/index.js` を配置。

#### 3.2.2 `aws_lambda_function.pre_signup`

| 設定 | 値 | 根拠 |
|---|---|---|
| `function_name` | `gp-${var.env}-presignup-fn` | Q-I4 |
| `runtime` | `nodejs20.x` | Q-D7 |
| `handler` | `index.handler` | LC-07 |
| `role` | `aws_iam_role.pre_signup_lambda.arn` | — |
| `filename` | `data.archive_file.pre_signup.output_path` | Q-I3 |
| `source_code_hash` | `data.archive_file.pre_signup.output_base64sha256` | デプロイトリガ |
| `memory_size` | 128 | 最小、5 行コード |
| `timeout` | 5 (seconds) | Cognito Trigger 推奨 |
| `architectures` | `["arm64"]` | コスト・性能で arm64 を選好 |
| `environment.LOG_LEVEL` | `info` | A-NFR-OBS-01 |

### 3.3 API Lambda (Hello World + Logout)

#### 3.3.1 `aws_lambda_function.api`

| 設定 | 値 | 根拠 |
|---|---|---|
| `function_name` | `gp-${var.env}-api-fn` | Q-I4 |
| `package_type` | `Image` | LWA + Go コンテナ |
| `image_uri` | `${var.api_image_uri}` (ECR タグ、Code Generation で push) | Q-I10=A4 |
| `role` | `aws_iam_role.api_lambda.arn` | — |
| `memory_size` | 512 | LWA + Gin 起動余裕 |
| `timeout` | 30 (seconds) | NFR-PERF-01 (3 秒) は handler 内、Lambda timeout は安全マージン |
| `architectures` | `["arm64"]` | 同上 |
| `environment.AWS_LWA_PORT` | `8080` | LWA 標準 |
| `environment.LOG_LEVEL` | `info` | A-NFR-OBS-01 |
| `environment.COGNITO_USER_POOL_ID` | `aws_cognito_user_pool.main.id` | unit-interfaces §10 |
| `environment.COGNITO_APP_CLIENT_ID` | `aws_cognito_user_pool_client.web.id` | 同上 |
| `environment.AWS_REGION` | `ap-northeast-1` | 同上 |

#### 3.3.2 `aws_lambda_permission.apigw_invoke_api`

API Gateway が API Lambda を呼び出せるよう許可。

### 3.4 API Gateway HTTP API (Q-I2=B)

#### 3.4.1 `aws_apigatewayv2_api.main`

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-api` |
| `protocol_type` | `HTTP` |
| `cors_configuration.allow_origins` | `["*"]` (本 MVP は緩く、Frontend ドメインに絞るのは本番化時) |
| `cors_configuration.allow_methods` | `["GET", "POST", "OPTIONS"]` |
| `cors_configuration.allow_headers` | `["Authorization", "Content-Type"]` |

#### 3.4.2 `aws_apigatewayv2_authorizer.cognito`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-auth-authorizer` | Q-I4 |
| `api_id` | `aws_apigatewayv2_api.main.id` | — |
| `authorizer_type` | `JWT` | Cognito = マネージド JWT |
| `identity_sources` | `["$request.header.Authorization"]` | 標準 |
| `jwt_configuration.audience` | `[aws_cognito_user_pool_client.web.id]` | LC-17 |
| `jwt_configuration.issuer` | `https://cognito-idp.${var.region}.amazonaws.com/${aws_cognito_user_pool.main.id}` | 同上 |
| `authorizer_result_ttl_in_seconds` | 60 | Q-I8 |

#### 3.4.3 `aws_apigatewayv2_stage.default`

| 設定 | 値 | 根拠 |
|---|---|---|
| `api_id` | `aws_apigatewayv2_api.main.id` | — |
| `name` | `$default` | デフォルトステージ |
| `auto_deploy` | true | HTTP API のシンプル運用 |
| `default_route_settings.throttling_burst_limit` | 200 | A-NFR-SEC-04 |
| `default_route_settings.throttling_rate_limit` | 100 | 同上 |
| `default_route_settings.detailed_metrics_enabled` | false | A-NFR-OBS-02 |
| `access_log_settings` | （未設定。本 MVP では Lambda 内ログのみで十分。本番化時 access log 追加） | — |

### 3.5 Logout Route (Q-I10=A4 範囲)

#### 3.5.1 `aws_apigatewayv2_integration.api_lambda`

| 設定 | 値 |
|---|---|
| `api_id` | `aws_apigatewayv2_api.main.id` |
| `integration_type` | `AWS_PROXY` |
| `integration_uri` | `aws_lambda_function.api.invoke_arn` |
| `payload_format_version` | `2.0` |
| `integration_method` | `POST` |

このリソースは Logout 専用ではなく **API Lambda への汎用 integration** として再利用される（Unit B 以降の route も同 integration を target にする）。

#### 3.5.2 `aws_apigatewayv2_route.logout`

| 設定 | 値 |
|---|---|
| `api_id` | `aws_apigatewayv2_api.main.id` |
| `route_key` | `POST /api/auth/logout` |
| `target` | `integrations/${aws_apigatewayv2_integration.api_lambda.id}` |
| `authorization_type` | `JWT` |
| `authorizer_id` | `aws_apigatewayv2_authorizer.cognito.id` |

### 3.6 IAM Roles & Policies (Q-I9)

#### 3.6.1 `aws_iam_role.pre_signup_lambda`

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-presignup-role` |
| `assume_role_policy` | Lambda service principal (`lambda.amazonaws.com`) |

インラインポリシー: CloudWatch Logs 書込のみ（Q-I9=A）

```
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"],
    "Resource": [aws_cloudwatch_log_group.pre_signup.arn, "${aws_cloudwatch_log_group.pre_signup.arn}:*"]
  }]
}
```

#### 3.6.2 `aws_iam_role.api_lambda`

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-api-role` |
| `assume_role_policy` | Lambda service principal |

インラインポリシー: 当面 CloudWatch Logs 書込のみ（後続 Unit でテーブル読書権限が追加される）

```
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"],
    "Resource": [aws_cloudwatch_log_group.api.arn, "${aws_cloudwatch_log_group.api.arn}:*"]
  }]
}
```

### 3.7 CloudWatch Log Groups (Q-I12)

| Resource | name | retention_in_days |
|---|---|---|
| `aws_cloudwatch_log_group.pre_signup` | `/aws/lambda/gp-${var.env}-presignup-fn` | 7 |
| `aws_cloudwatch_log_group.api` | `/aws/lambda/gp-${var.env}-api-fn` | 7 |

---

## 4. Variables / Outputs

### 4.1 Module Variables (`infra/modules/auth/variables.tf`)

| 名前 | 型 | デフォルト | 説明 |
|---|---|---|---|
| `env` | string | — | 環境識別子 (例: `dev`) |
| `region` | string | `ap-northeast-1` | AWS Region |
| `api_image_uri` | string | — | API Lambda 用コンテナイメージ URI (ECR タグ) |
| `tags` | map(string) | (default_tags で代替可) | 追加タグ |

### 4.2 Module Outputs (`infra/modules/auth/outputs.tf`)

他 module / 後続 Unit が参照する値:

| 名前 | 値 | 用途 |
|---|---|---|
| `user_pool_id` | `aws_cognito_user_pool.main.id` | Frontend Amplify 設定 / 他 Unit の Lambda env |
| `user_pool_arn` | `aws_cognito_user_pool.main.arn` | 他 IAM 連携 |
| `user_pool_endpoint` | `aws_cognito_user_pool.main.endpoint` | JWT issuer URL |
| `user_pool_client_id` | `aws_cognito_user_pool_client.web.id` | Frontend Amplify 設定 |
| `api_id` | `aws_apigatewayv2_api.main.id` | 他 Unit の route 追加時 |
| `api_endpoint` | `aws_apigatewayv2_api.main.api_endpoint` | Frontend が呼ぶ URL |
| `api_lambda_function_name` | `aws_lambda_function.api.function_name` | 他 Unit の route で使用 |
| `api_lambda_invoke_arn` | `aws_lambda_function.api.invoke_arn` | 他 Unit の integration |
| `api_lambda_role_arn` | `aws_iam_role.api_lambda.arn` | 他 Unit が DynamoDB 等の権限を attach |
| `cognito_authorizer_id` | `aws_apigatewayv2_authorizer.cognito.id` | 他 Unit の route の `authorizer_id` |

---

## 5. backend.tf (Q-I6 確定版)

`infra/envs/dev/backend.tf`:

```hcl
terraform {
  backend "s3" {
    bucket       = "gp-tfstate-dev"
    key          = "auth/terraform.tfstate"   # Unit ごとに分離する場合は "auth/" プレフィックス
    region       = "ap-northeast-1"
    encrypt      = true
    use_lockfile = true                        # S3 ネイティブ Lock (Q-I6)
  }
}
```

**注意**: bucket `gp-tfstate-dev` は手動 or bootstrap スクリプトで先行作成すること。Unit A PR の README に明記。

---

## 6. providers.tf (Q-I11)

`infra/envs/dev/providers.tf`:

```hcl
provider "aws" {
  region = "ap-northeast-1"

  default_tags {
    tags = {
      Project   = "goro2pay"
      Env       = "dev"
      Unit      = "auth"
      ManagedBy = "terraform"
    }
  }
}
```

`Unit` タグは Unit A の責任範囲を示すために `auth` を default に置くが、他 Unit が後続で別の dev env から module 呼出を追加する場合は `Unit` を override する設計とする（または各 module 内で `tags` 引数を受けて override）。

---

## 7. terraform-test 統合

A-NFR-MAINT-01 / terraform-test プラグイン規約に従い、以下のテストを `infra/modules/auth/tests/` に配置（Code Generation で実装）:

| テストファイル | 目的 |
|---|---|
| `auth_basic.tftest.hcl` | mock_provider で `terraform plan` がエラーなく成立することを確認 |
| `auth_outputs.tftest.hcl` | 主要 output が空文字でないことを確認 |
| `auth_cognito_password_policy.tftest.hcl` | password_policy が A-NFR-SEC-02 と一致することを確認 |

---

## 8. リソース → 論理コンポーネント / NFR トレーサビリティ

| Terraform リソース | 対応 LC | 対応 A-NFR / Q-I |
|---|---|---|
| `aws_cognito_user_pool.main` | LC-15 | A-NFR-SEC-01/02 / Q-I7/I13 |
| `aws_cognito_user_pool_client.web` | LC-16 | A-NFR-SEC-03 |
| `aws_lambda_function.pre_signup` | LC-07 | A-NFR-REL-01 / Q-D7/I3 |
| `aws_apigatewayv2_api.main` | (横串) | Q-I2/I10 |
| `aws_apigatewayv2_authorizer.cognito` | LC-17 | A-NFR-SEC-01 / Q-I8 |
| `aws_apigatewayv2_stage.default` | (横串) | A-NFR-SEC-04 |
| `aws_apigatewayv2_route.logout` + integration | (Logout 用) | LC-AUTH-06 |
| `aws_lambda_function.api` | (横串、本 PR で先行) | Q-I10=A4 |
| `aws_iam_role.*` | (各 Lambda) | Q-I9 |
| `aws_cloudwatch_log_group.*` | (各 Lambda) | A-NFR-OBS-01 / Q-I12 |

---

## 9. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Code Generation** | `infra/lambdas/pre-signup/index.js` (5 行) / `infra/lambdas/api/` の Go コード + Dockerfile + ECR push スクリプト / Terraform `*.tf` の HCL 本体 / mock_provider テスト |
| **横串改善 PR (将来)** | `lambda_api/` を独立 module として切り出し、Auth module から API Lambda 関連を移動。本 MVP では Auth module 内の `api_lambda.tf` で先行構築 |
| **Frontend Amplify Hosting** | 別 PR で `infra/modules/amplify/` を構築、Unit A から `user_pool_id` / `user_pool_client_id` / `api_endpoint` を output 経由で取得 |

---

## 10. 既存ドキュメント整合性メモ

### 10.1 Application Design / NFR Design との差分

- **API Gateway 種類**: 既存ドキュメント「REST」表記 → 本書で **HTTP API** に変更（Q-I2=B）。`unit-interfaces.md` §3.3 の「path」表記は HTTP API でも同じ動作。`A-NFR-SEC-04 Stage Throttling` は HTTP API では default_route_settings として表現される
- **API Lambda 構築タイミング**: unit-of-work.md §4.1 では `lambda_api/` を「Unit 横串」と記載 → 本書で **Unit A PR で先行構築する**（横串 PR の所在不明確のため、Q-I10=A4）。後続 PR で必要なら独立 module への切り出しを検討

### 10.2 整合修正メモ

本 PR では既存ドキュメントは変更しないが、Code Generation 完了後にレビューで以下の調整を検討:

- `unit-interfaces.md` §3.3 の API path prefix 確認（`/api/...` で統一済み、PR #66）
- `unit-of-work.md` §4.1 の `lambda_api/` 横串記述に「Unit A PR で先行構築」の脚注追加（任意）
