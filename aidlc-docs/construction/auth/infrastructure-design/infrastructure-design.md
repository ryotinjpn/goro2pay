# Auth Unit — Infrastructure Design

**Document Version**: 1.2
**Created**: 2026-05-22
**Updated**: 2026-05-22 (Q-I14/Q-I15 追加: Amplify Hosting + CodePipeline/CodeBuild を Unit A スコープに追加、横串インフラ = Unit A 方針)
**Updated**: 2026-05-22 (BFF パターン採用 + back/ ディレクトリリネーム + API 認証を AccessToken に統一、cors_configuration を未設定化)
**Unit**: A (`auth`)
**Construction Depth**: Standard
**Stage**: Infrastructure Design / Construction
**Predecessors**: NFR Design 完了 (PR #68)

本ドキュメントは Unit A の **Terraform リソース定義** を凍結する。NFR Design で論理化した LC-15/16/17 / LC-07 を実 AWS リソースに展開する。

参照: [auth-infrastructure-design-plan.md](../../plans/auth-infrastructure-design-plan.md), [logical-components.md](../nfr-design/logical-components.md), [nfr-design-patterns.md](../nfr-design/nfr-design-patterns.md), [unit-interfaces.md](../../interfaces/unit-interfaces.md)

---

## 1. スコープと前提

### 1.1 本書のスコープ（Q-I10=A4 / Q-I14 / Q-I15 確定）

横串インフラ = Unit A 方針に従い、Unit A PR で構築するリソース:

1. **Cognito** (User Pool / App Client) ← LC-15/16
2. **Pre Sign-up Lambda** (Node.js, auto-confirm) ← LC-07
3. **API Gateway HTTP API** (本体) ← Q-I2=B
4. **JWT Authorizer** (Cognito 連携) ← LC-17
5. **`POST /api/auth/logout` route + integration**
6. **API Lambda 本体** (Go + Gin + LWA, 当面は Hello World + Logout のみ)
7. **AWS Amplify Hosting** (Next.js App Router、GitHub auto deploy) ← Q-I14=A
8. **CodePipeline + CodeBuild** (API Lambda の CD) ← Q-I15=D
9. **ECR Repository** (API Lambda image)
10. CloudWatch Log Groups（Lambda + CodeBuild 用）
11. IAM Roles / Policies（各 Lambda + Amplify SSR + CodePipeline + CodeBuild）

### 1.2 スコープ外（他 PR / 後続 Unit が担当）

- Unit B/C/D/E のハンドラ実装 + route 追加 → 各 Unit の Construction
- DynamoDB テーブル定義 → 各 Unit の Infrastructure Design
- Bedrock IAM → Unit C の Infrastructure Design
- Pre Sign-up Lambda の CD（archive_file + terraform apply のまま手動運用）

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
│   └── auth/                     # Q-I1=A 単一モジュール（横串インフラ含む）
│       ├── README.md
│       ├── main.tf               # メイン定義
│       ├── cognito.tf            # User Pool + App Client
│       ├── pre_signup_lambda.tf  # Pre Sign-up Lambda (Node.js, archive_file)
│       ├── api_gateway.tf        # HTTP API + Authorizer + Stage
│       ├── api_lambda.tf         # API Lambda (Go + Gin + LWA, ECR image)
│       ├── ecr.tf                # ECR Repository (API Lambda image, Q-I15)
│       ├── logout_route.tf       # POST /api/auth/logout route + integration
│       ├── amplify.tf            # Amplify Hosting (Next.js App Router) (Q-I14)
│       ├── codepipeline.tf       # CodePipeline + CodeBuild (API Lambda CD) (Q-I15)
│       ├── iam.tf                # 全 IAM Role / Policy
│       ├── log_groups.tf         # CloudWatch Log Groups
│       ├── variables.tf
│       └── outputs.tf
├── lambdas/
│   └── pre-signup/
│       └── index.js              # 5 行 auto-confirm 実装 (Code Generation)
│                                  # ※ archive_file の source_dir 都合で infra/ 配下
└── scripts/
    ├── bootstrap-backend.sh      # S3 tfstate bucket 作成 (Code Generation)
    └── bootstrap-ecr-initial.sh  # ECR 初回 image push (Code Generation)
```

加えて、リポジトリルート直下の `back/` ディレクトリに API Lambda 関連の Go コードと Docker build 資材を配置:

```
back/
└── api/                          # API Lambda (Go + Gin + LWA)
    ├── Dockerfile                # LWA + arm64 Go バイナリ build (Code Generation)
    ├── buildspec.yml             # CodeBuild build 仕様 (Code Generation)
    ├── go.mod
    ├── main.go                   # Gin 起動 + 依存注入
    └── internal/                 # Unit A〜E の実装パッケージ
        ├── auth/
        ├── logging/
        ├── handlers/
        └── (将来 Unit B/C/D/E が追加)
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
| `image_uri` | `${aws_ecr_repository.api.repository_url}:bootstrap` (初期値、CD 後は CodeBuild が更新) | Q-I10=A4 / Q-I15 |
| `role` | `aws_iam_role.api_lambda.arn` | — |
| `memory_size` | 512 | LWA + Gin 起動余裕 |
| `timeout` | 30 (seconds) | NFR-PERF-01 (3 秒) は handler 内、Lambda timeout は安全マージン |
| `architectures` | `["arm64"]` | 同上 |
| `environment.AWS_LWA_PORT` | `8080` | LWA 標準 |
| `environment.LOG_LEVEL` | `info` | A-NFR-OBS-01 |
| `environment.COGNITO_USER_POOL_ID` | `aws_cognito_user_pool.main.id` | unit-interfaces §10 |
| `environment.COGNITO_APP_CLIENT_ID` | `aws_cognito_user_pool_client.web.id` | 同上 |
| `environment.AWS_REGION` | `ap-northeast-1` | 同上 |
| `lifecycle.ignore_changes` | `[image_uri]` | Q-I15: CodeBuild が image_uri を更新するため Terraform は無視 |

#### 3.3.2 `aws_lambda_permission.apigw_invoke_api`

API Gateway が API Lambda を呼び出せるよう許可。

#### 3.3.3 `aws_ecr_repository.api`

API Lambda の Docker image を保存。

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-api-image` | Q-I4 |
| `image_tag_mutability` | `MUTABLE` | CodeBuild が `latest` タグを上書きするため (Q-I15) |
| `image_scanning_configuration.scan_on_push` | true | セキュリティ標準 |
| `encryption_configuration.encryption_type` | `AES256` | A-NFR-SEC-05 マネージドデフォルト |

`aws_ecr_lifecycle_policy.api`: 古いイメージを削除（直近 5 個保持、ストレージコスト削減）。

### 3.4 API Gateway HTTP API (Q-I2=B)

#### 3.4.1 `aws_apigatewayv2_api.main`

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-api` |
| `protocol_type` | `HTTP` |
| `cors_configuration` | **未設定** (BFF パターン採用、API Gateway を直接ブラウザから叩かないため CORS 不要) |

**Note**: BFF パターンにより、ブラウザは `/api/*` を Next.js Server (Amplify Hosting) 経由で呼び出す。API Gateway へのリクエストは常に server-to-server となり、CORS preflight は発生しない。本番化時に直叩きユースケースが出てきたら CORS 設定を追加する。

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
| `aws_cloudwatch_log_group.codebuild_api` | `/aws/codebuild/gp-${var.env}-api-build` | 7 |

### 3.8 Amplify Hosting (Q-I14=A)

#### 3.8.1 `aws_codestarconnections_connection.github`

GitHub と AWS の接続を提供。Amplify と CodePipeline の両方が参照する。

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-github-conn` |
| `provider_type` | `GitHub` |

**注意**: Connection 作成後、AWS Console で **Pending → Available** への手動承認が 1 回必要（GitHub App インストール）。Terraform apply 直後は Pending 状態。Unit A PR README に手動承認手順を明記。

#### 3.8.2 `aws_amplify_app.web`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-web` | Q-I4 |
| `repository` | `https://github.com/{owner}/goro2pay` | Q-I14 |
| `platform` | `WEB_COMPUTE` | Next.js App Router SSR/SSG（PR #10 確定） |
| `iam_service_role_arn` | `aws_iam_role.amplify_ssr.arn` | SSR 実行用 |
| `enable_branch_auto_build` | true | Q-I14 |
| `build_spec` | （後述 §3.8.5） | Next.js build |
| `environment_variables` | `_LIVE_UPDATES`、`AMPLIFY_DIFF_DEPLOY=false` 等の Amplify 標準 | — |
| `custom_rule[*]` | `</^[^.]+$|\.(?!(css|gif|ico|jpg|js|png|txt|svg|woff|woff2|ttf|map|json)$)([^.]+$)/`<br>`{ source: "</^[^.]+$|\.(?!(css|...)$)/", target: "/index.html", status: "200" }` | App Router の SPA fallback |
| `connection_arn` | `aws_codestarconnections_connection.github.arn` | OAuth 不要 |

#### 3.8.3 `aws_amplify_branch.develop`

| 設定 | 値 |
|---|---|
| `app_id` | `aws_amplify_app.web.id` |
| `branch_name` | `develop` |
| `enable_auto_build` | true |
| `framework` | `Next.js - SSR` |
| `stage` | `DEVELOPMENT` |
| `environment_variables.NEXT_PUBLIC_USER_POOL_ID` | `aws_cognito_user_pool.main.id` | Browser に露出 OK (Amplify Auth が利用) |
| `environment_variables.NEXT_PUBLIC_USER_POOL_CLIENT_ID` | `aws_cognito_user_pool_client.web.id` | 同上 |
| `environment_variables.NEXT_PUBLIC_AWS_REGION` | `ap-northeast-1` | 同上 |
| `environment_variables.API_ENDPOINT` | `aws_apigatewayv2_api.main.api_endpoint` | **server-only**（NEXT_PUBLIC_ なし、catch-all Route Handler が proxy 先として使用） |

このブランチへの GitHub push が auto deploy トリガとなる。

**env var 命名ポリシー** (BFF パターン整合):
- `NEXT_PUBLIC_*` → ブラウザバンドルに含まれて OK な値（Cognito User Pool ID / Client ID は public な情報、Amplify Auth がブラウザで利用するため必須）
- prefix なし → server-only（Next.js Server からのみアクセス可、ブラウザバンドル除外）。`API_ENDPOINT` を秘匿することで API Gateway URL がブラウザに露出しない

#### 3.8.4 `aws_iam_role.amplify_ssr`

Amplify Hosting の SSR Compute role。

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-amplify-ssr-role` |
| `assume_role_policy` | `amplify.amazonaws.com` service principal |

インラインポリシー: CloudWatch Logs 書込のみ。Amplify SSR が他 AWS サービスを呼ぶ必要は本 MVP では無し（Cognito 呼出は Frontend ブラウザ側、API は API Gateway 経由）。

#### 3.8.5 buildSpec（YAML 文字列を Terraform 内で）

```yaml
version: 1
applications:
  - frontend:
      phases:
        preBuild:
          commands:
            - cd web
            - npm ci
        build:
          commands:
            - npm run build
      artifacts:
        baseDirectory: web/.next
        files:
          - '**/*'
      cache:
        paths:
          - web/node_modules/**/*
          - web/.next/cache/**/*
```

具体 YAML は Code Generation で確定。

### 3.9 CodePipeline + CodeBuild (Q-I15=D)

API Lambda の CD パイプライン。

#### 3.9.1 `aws_codebuild_project.api`

| 設定 | 値 | 根拠 |
|---|---|---|
| `name` | `gp-${var.env}-api-build` | Q-I4 |
| `service_role` | `aws_iam_role.codebuild_api.arn` | — |
| `artifacts.type` | `CODEPIPELINE` | Pipeline 内で実行 |
| `environment.compute_type` | `BUILD_GENERAL1_SMALL` | 最小コスト |
| `environment.image` | `aws/codebuild/standard:7.0` | Docker / Go ビルド可能 |
| `environment.type` | `LINUX_CONTAINER` | — |
| `environment.privileged_mode` | true | Docker build に必要 |
| `environment.environment_variable.AWS_DEFAULT_REGION` | `ap-northeast-1` | — |
| `environment.environment_variable.ECR_REPOSITORY_URI` | `aws_ecr_repository.api.repository_url` | — |
| `environment.environment_variable.LAMBDA_FUNCTION_NAME` | `aws_lambda_function.api.function_name` | image 更新用 |
| `source.type` | `CODEPIPELINE` | — |
| `source.buildspec` | `back/api/buildspec.yml` | — |
| `logs_config.cloudwatch_logs.group_name` | `/aws/codebuild/gp-${var.env}-api-build` | — |

#### 3.9.2 buildspec.yml（API Lambda 用）

```yaml
version: 0.2
phases:
  pre_build:
    commands:
      - echo Logging in to Amazon ECR...
      - aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login --username AWS --password-stdin $ECR_REPOSITORY_URI
      - IMAGE_TAG=$(echo $CODEBUILD_RESOLVED_SOURCE_VERSION | cut -c1-7)
  build:
    commands:
      - echo Building Docker image...
      - docker buildx build --platform linux/arm64 -t $ECR_REPOSITORY_URI:$IMAGE_TAG -t $ECR_REPOSITORY_URI:latest -f back/api/Dockerfile back/api/
  post_build:
    commands:
      - echo Pushing Docker image...
      - docker push $ECR_REPOSITORY_URI:$IMAGE_TAG
      - docker push $ECR_REPOSITORY_URI:latest
      - echo Updating Lambda function...
      - aws lambda update-function-code --function-name $LAMBDA_FUNCTION_NAME --image-uri $ECR_REPOSITORY_URI:$IMAGE_TAG --region $AWS_DEFAULT_REGION
```

具体 YAML は Code Generation で確定。

#### 3.9.3 `aws_codepipeline.api`

| 設定 | 値 |
|---|---|
| `name` | `gp-${var.env}-api-pipeline` |
| `role_arn` | `aws_iam_role.codepipeline_api.arn` |
| `artifact_store.location` | `aws_s3_bucket.codepipeline_artifacts.bucket` |
| `artifact_store.type` | `S3` |

##### Source Stage
| 設定 | 値 |
|---|---|
| `category` | `Source` |
| `owner` | `AWS` |
| `provider` | `CodeStarSourceConnection` |
| `configuration.ConnectionArn` | `aws_codestarconnections_connection.github.arn` |
| `configuration.FullRepositoryId` | `{owner}/goro2pay` |
| `configuration.BranchName` | `develop` |
| `output_artifacts` | `["source_output"]` |

##### Build Stage
| 設定 | 値 |
|---|---|
| `category` | `Build` |
| `owner` | `AWS` |
| `provider` | `CodeBuild` |
| `configuration.ProjectName` | `aws_codebuild_project.api.name` |
| `input_artifacts` | `["source_output"]` |
| `output_artifacts` | `["build_output"]` |

#### 3.9.4 `aws_s3_bucket.codepipeline_artifacts`

CodePipeline の中間アーティファクト保存用。

| 設定 | 値 |
|---|---|
| `bucket` | `gp-${var.env}-codepipeline-artifacts` |
| `force_destroy` | true (dev のみ) |
| `versioning.enabled` | false (中間データのため) |
| `lifecycle_rule` | 30 日後削除 |

#### 3.9.5 IAM Roles for CD

##### `aws_iam_role.codepipeline_api`
- assume_role: `codepipeline.amazonaws.com`
- 権限: S3 artifact bucket 読書、CodeBuild StartBuild、CodeStar Connection 使用

##### `aws_iam_role.codebuild_api`
- assume_role: `codebuild.amazonaws.com`
- 権限:
  - CloudWatch Logs 書込
  - ECR push (`ecr:GetAuthorizationToken`, `ecr:BatchCheckLayerAvailability`, `ecr:PutImage`, `ecr:InitiateLayerUpload`, `ecr:UploadLayerPart`, `ecr:CompleteLayerUpload`)
  - Lambda 更新 (`lambda:UpdateFunctionCode`) on `aws_lambda_function.api.arn` のみ
  - S3 artifact bucket 読書

---

## 4. Variables / Outputs

### 4.1 Module Variables (`infra/modules/auth/variables.tf`)

| 名前 | 型 | デフォルト | 説明 |
|---|---|---|---|
| `env` | string | — | 環境識別子 (例: `dev`) |
| `region` | string | `ap-northeast-1` | AWS Region |
| `github_owner` | string | — | GitHub オーナー (例: `ryotinjpn`) |
| `github_repo_name` | string | `goro2pay` | リポジトリ名 |
| `github_branch` | string | `develop` | Frontend / API CD のソースブランチ |
| `tags` | map(string) | (default_tags で代替可) | 追加タグ |

**注意**: `api_image_uri` 変数は不要となった（ECR を Unit A で構築、初期 image は bootstrap-ecr-initial.sh で push、以降は CodeBuild が更新するため Lambda の image_uri は `lifecycle.ignore_changes`）。

### 4.2 Module Outputs (`infra/modules/auth/outputs.tf`)

他 module / 後続 Unit が参照する値:

| 名前 | 値 | 用途 |
|---|---|---|
| `user_pool_id` | `aws_cognito_user_pool.main.id` | Amplify env / 他 Unit の Lambda env |
| `user_pool_arn` | `aws_cognito_user_pool.main.arn` | 他 IAM 連携 |
| `user_pool_endpoint` | `aws_cognito_user_pool.main.endpoint` | JWT issuer URL |
| `user_pool_client_id` | `aws_cognito_user_pool_client.web.id` | Amplify env |
| `api_id` | `aws_apigatewayv2_api.main.id` | 他 Unit の route 追加時 |
| `api_endpoint` | `aws_apigatewayv2_api.main.api_endpoint` | Amplify env / Frontend が呼ぶ URL |
| `api_lambda_function_name` | `aws_lambda_function.api.function_name` | 他 Unit の route で使用、CodeBuild の lambda update |
| `api_lambda_invoke_arn` | `aws_lambda_function.api.invoke_arn` | 他 Unit の integration |
| `api_lambda_role_arn` | `aws_iam_role.api_lambda.arn` | 他 Unit が DynamoDB 等の権限を attach |
| `cognito_authorizer_id` | `aws_apigatewayv2_authorizer.cognito.id` | 他 Unit の route の `authorizer_id` |
| `amplify_app_id` | `aws_amplify_app.web.id` | デプロイ確認 / Console URL |
| `amplify_default_domain` | `aws_amplify_app.web.default_domain` | Frontend 公開 URL |
| `ecr_repository_url` | `aws_ecr_repository.api.repository_url` | 初回 image push スクリプト |
| `codepipeline_name` | `aws_codepipeline.api.name` | デプロイ状態確認 |

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

**注意**: 以下を Unit A PR の README に明記、bootstrap スクリプトとして提供:

1. **S3 tfstate bucket の作成**: `gp-tfstate-dev` を `infra/scripts/bootstrap-backend.sh` で作成
2. **ECR 初回 image push**: CodePipeline が動作する前に Lambda が起動できるよう、`infra/scripts/bootstrap-ecr-initial.sh` で「最小の Hello World image」を `:bootstrap` タグで push。これは Lambda の `image_uri` 初期値となる
3. **CodeStar Connection の手動承認**: Terraform apply 直後、AWS Console で「Pending → Available」へ手動承認（GitHub App インストール）

これら 3 手順を terraform apply 前に 1 回だけ実施する。

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
| `auth_lambda_lifecycle.tftest.hcl` | API Lambda の `lifecycle.ignore_changes = ["image_uri"]` が設定されていることを確認 (Q-I15) |
| `auth_amplify_branch.tftest.hcl` | Amplify branch の env vars に `NEXT_PUBLIC_USER_POOL_ID` 等が含まれることを確認 (Q-I14) |
| `auth_codebuild_iam.tftest.hcl` | CodeBuild IAM Role が Lambda UpdateFunctionCode 権限を最小限で持つことを確認 (Q-I15) |

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
| `aws_ecr_repository.api` | (横串、API Lambda image) | Q-I15 |
| `aws_amplify_app.web` + `aws_amplify_branch.develop` | (横串、Frontend hosting) | Q-I14 / 横串インフラ Unit A 方針 |
| `aws_codestarconnections_connection.github` | (横串、CD ソース) | Q-I14 / Q-I15 |
| `aws_codepipeline.api` + `aws_codebuild_project.api` | (横串、API Lambda CD) | Q-I15 |
| `aws_s3_bucket.codepipeline_artifacts` | (CD 中間) | Q-I15 |
| `aws_iam_role.*` | (各サービス) | Q-I9 / Q-I15 |
| `aws_cloudwatch_log_group.*` | (各 Lambda + CodeBuild) | A-NFR-OBS-01 / Q-I12 |

---

## 9. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Code Generation** | `infra/lambdas/pre-signup/index.js` (5 行) / `infra/lambdas/api/Dockerfile` (LWA arm64) / `infra/lambdas/api/buildspec.yml` (CodeBuild) / `back/api/` の Go コード本体 / `web/` の Next.js Frontend / `infra/scripts/bootstrap-backend.sh` (S3 tfstate bucket) / `infra/scripts/bootstrap-ecr-initial.sh` (ECR 初回 image push) / Terraform `*.tf` の HCL 本体 / mock_provider テスト |
| **将来の横串改善 PR** | Auth module から `api_lambda.tf` / `amplify.tf` / `codepipeline.tf` / `ecr.tf` を独立 module へ切り出し（`lambda_api/` / `amplify/` / `cicd/`）。本 MVP では Auth module 内に集約 |

---

## 10. 既存ドキュメント整合性メモ

### 10.1 Application Design / NFR Design との差分

- **API Gateway 種類**: 既存ドキュメント「REST」表記 → 本書で **HTTP API** に変更（Q-I2=B）。`unit-interfaces.md` §3.3 の「path」表記は HTTP API でも同じ動作。`A-NFR-SEC-04 Stage Throttling` は HTTP API では default_route_settings として表現される
- **API Lambda 構築タイミング**: unit-of-work.md §4.1 では `lambda_api/` を「Unit 横串」と記載 → 本書で **Unit A PR で先行構築する**（横串 PR の所在不明確のため、Q-I10=A4）。後続 PR で必要なら独立 module への切り出しを検討
- **Amplify Hosting / CodePipeline / ECR の所属**: unit-of-work.md §4.1 では `amplify/` / `lambda_api/` を「Unit 横串」と記載していたが、横串 PR タスク管理が計画上空白だったため、本書で **横串インフラを全て Unit A スコープに包含する** 方針に確定（Q-I14 / Q-I15）。これにより Unit A PR 単体で Frontend と API の auto deploy 環境まで構築可能
- **BFF パターン採用**: NEXT_PUBLIC_API_ENDPOINT でブラウザに API URL を露出する設計を取りやめ、`/api/*` パスは Next.js の catch-all Route Handler (`web/app/api/[...path]/route.ts`) が受けて API Gateway に proxy する BFF パターンを採用。`API_ENDPOINT` は server-only env、CORS allow_origins を Amplify ドメインだけに絞れる
- **Go ソース配置**: unit-of-work.md §4.1 の `apps/api/` / `apps/scheduler/` 表記を、本 PR 内で **`back/api/` / `back/scheduler/`** にリネーム（リポジトリルート直下の `back/` ディレクトリ）。各層 (Browser / Next.js / API Gateway / API Lambda) の URL は全て `/api/*` で統一

### 10.2 整合修正メモ

本 PR では既存ドキュメントは変更しないが、Code Generation 完了後にレビューで以下の調整を検討:

- `unit-interfaces.md` §3.3 の API path prefix 確認（`/api/...` で統一済み、PR #66）
- `unit-of-work.md` §4.1 の `lambda_api/` / `amplify/` / `api_gateway/` 横串記述に「Unit A PR で先行構築、横串改善 PR で将来切り出し」の脚注追加（任意）
- `unit-of-work.md` §4.1 の Go ソース配置 `apps/api/` / `apps/scheduler/` を `back/api/` / `back/scheduler/` にリネーム（必須）
- 各 Unit (B/C/D/E) の Functional Design で `apps/api/internal/<unit>/` 参照を `back/api/internal/<unit>/` に修正（各 Unit 担当者が自 Unit Construction 着手時に対応）
