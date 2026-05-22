# Auth Unit — Deployment Architecture

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: A (`auth`)
**Stage**: Infrastructure Design / Construction
**Predecessors**: [infrastructure-design.md](./infrastructure-design.md)

本ドキュメントは Unit A の **デプロイメント時の AWS リソース構成と相互関係、デプロイ手順、環境分離戦略** を記述する。

---

## 1. AWS リソース全体図

### 1.1 Unit A PR でデプロイされる AWS リソース（Q-I10=A4 確定）

```
                          ┌──────────────────────────────────────────────┐
                          │            ap-northeast-1 (Tokyo)            │
                          │                                              │
   ┌────────────┐         │  ┌────────────────────────────────────────┐  │
   │  Frontend  │         │  │    Cognito User Pool (gp-dev-userpool) │  │
   │  (Amplify  │ ── 1 ──►│  │      ├─ App Client (web, no secret)    │  │
   │   Hosting, │         │  │      ├─ Pre Sign-up Lambda Trigger ────┼──┐
   │   別 PR)   │         │  │      └─ User attributes: email only    │  │
   └─────┬──────┘         │  └────────────┬───────────────────────────┘  │
         │                │               │                              │
         │ 2: Sign-up     │               ▼                              │
         │ 2: Sign-in     │  ┌────────────────────────────────────────┐  │
         │ 2: Token取得   │  │  Pre Sign-up Lambda (Node.js, 5 行)   │  │
         │                │  │  gp-dev-presignup-fn                   │◄─┘
         │                │  │  - autoConfirmUser = true              │
         │                │  │  - autoVerifyEmail = true              │
         │                │  └─────────┬──────────────────────────────┘
         │                │            │ logs                           │
         │                │            ▼                                │
         │                │  ┌────────────────────────────────────┐    │
         │                │  │  CloudWatch Log Group              │    │
         │                │  │  /aws/lambda/gp-dev-presignup-fn   │    │
         │                │  │  retention 7 日                    │    │
         │                │  └────────────────────────────────────┘    │
         │                │                                              │
         │ 3: API call    │  ┌────────────────────────────────────────┐  │
         │ Authorization: │  │  API Gateway HTTP API (gp-dev-api)     │  │
         │ Bearer<Token>  │  │  Stage: $default                       │  │
         └───────────────►│  │  Throttling: 100 req/s, Burst 200      │  │
                          │  │  ├─ JWT Authorizer (Cognito 連携)      │  │
                          │  │  │   TTL 60s                           │  │
                          │  │  ├─ Route: POST /api/auth/logout       │  │
                          │  │  │       └─ Authorizer: JWT            │  │
                          │  │  └─ Integration: API Lambda (AWS_PROXY)│  │
                          │  └─────────┬──────────────────────────────┘  │
                          │            │                                  │
                          │            │ invoke                           │
                          │            ▼                                  │
                          │  ┌────────────────────────────────────────┐  │
                          │  │  API Lambda (Go + Gin + LWA, arm64)   │  │
                          │  │  gp-dev-api-fn                         │  │
                          │  │  Hello World + Logout handler          │  │
                          │  │  ENV: COGNITO_USER_POOL_ID,            │  │
                          │  │       COGNITO_APP_CLIENT_ID,           │  │
                          │  │       AWS_LWA_PORT=8080,               │  │
                          │  │       LOG_LEVEL=info                   │  │
                          │  └─────────┬──────────────────────────────┘  │
                          │            │ logs                              │
                          │            ▼                                  │
                          │  ┌────────────────────────────────────────┐  │
                          │  │  CloudWatch Log Group                  │  │
                          │  │  /aws/lambda/gp-dev-api-fn             │  │
                          │  │  retention 7 日                        │  │
                          │  └────────────────────────────────────────┘  │
                          │                                              │
                          │  ┌────────────────────────────────────────┐  │
                          │  │  ECR (gp-dev-api-image)                │  │
                          │  │  Go + Gin + LWA コンテナイメージ       │  │
                          │  └────────────────────────────────────────┘  │
                          │                                              │
                          │  ┌────────────────────────────────────────┐  │
                          │  │  IAM Roles                             │  │
                          │  │  ├─ gp-dev-presignup-role (logs only) │  │
                          │  │  └─ gp-dev-api-role (logs only)       │  │
                          │  └────────────────────────────────────────┘  │
                          │                                              │
                          └──────────────────────────────────────────────┘

凡例:
   1 = User Pool 設定取得 (config / amplify_outputs.json 経由、別 PR)
   2 = Cognito 認証フロー (Amplify Auth ライブラリ経由)
   3 = 業務 API 呼出 (IdToken 付与)
```

### 1.2 後続 PR で追加されるリソース（参考）

```
別 PR (Frontend):
   - AWS Amplify Hosting (Next.js App Router)
   - Amplify ↔ Cognito 設定連携 (Auth Unit の output を参照)

各 Unit (B/C/D/E):
   - DynamoDB Tables (Wallet / OrderHistory / IdempotencyKeys 等)
   - 各 Unit の handler 用 route + integration
     (Auth Unit 既設 API Gateway / API Lambda に乗る)
   - Bedrock IAM (Unit C/D)
   - EventBridge Scheduler + 月初リセット Lambda (Unit B)
```

---

## 2. リクエストフロー（主要シナリオ）

### 2.1 Sign-up フロー

```
[太郎ブラウザ]
   │
   │ 1. POST /signup form 送信
   ▼
[Frontend Amplify Auth.signUp(email, password)]
   │
   │ 2. Cognito SignUp API 直接呼出
   ▼
[Cognito User Pool]
   │
   │ 3. Pre Sign-up Lambda Trigger 起動
   ▼
[Pre Sign-up Lambda]
   │   event.response.autoConfirmUser = true
   │   event.response.autoVerifyEmail = true
   │   return event
   ▼
[Cognito User Pool]
   │
   │ 4. ユーザを CONFIRMED 状態で作成
   │
   ▼
[Frontend Amplify]
   │
   │ 5. Auth.signIn(email, password) を自動実行
   ▼
[Cognito InitiateAuth]
   │
   │ 6. IdToken / AccessToken / RefreshToken 返却
   ▼
[Frontend]
   │
   │ 7. localStorage に Token 保存 (Amplify 内部)
   │ 8. router.push("/")
```

### 2.2 認証付き API 呼出フロー（例: Logout）

```
[太郎ブラウザ]
   │
   │ 1. ヘッダの「ログアウト」ボタン押下 (確認モーダル後)
   ▼
[Frontend useAuth().logout()]
   │
   │ 2. Auth.signOut({ global: true })
   ▼
[Cognito GlobalSignOut]
   │
   │ 3. Refresh Token 取り消し
   │
   ▼
[Frontend Amplify]
   │
   │ 4. localStorage から Token 削除
   │
   │ 5. POST /api/auth/logout (Authorization: Bearer <IdToken>)
   ▼
[API Gateway HTTP API]
   │
   │ 6. JWT Authorizer 検証 (cache 60s, miss なら Cognito JWKS 検証)
   │    → claims を route 通過時に context へ
   │
   ▼
[API Lambda (Gin + LWA)]
   │
   │ 7. RequestContextMiddleware (LC-04)
   │    → requestId / traceId / userAgent を context に
   │
   │ 8. AttachUserIDMiddleware (LC-01)
   │    → claims.sub を userId として context に
   │    → claims.email から emailHash を生成して context に
   │
   │ 9. LogoutHandler (LC-06)
   │    → slog.InfoContext(ctx, "user logout", "action", "logout")
   │    → c.Status(204)
   ▼
[CloudWatch Logs]
   │
   │ JSON ログ出力（8 項目構造化）
   │
   ▼
[Frontend]
   │
   │ 10. router.push("/") (LandingScreen 表示)
```

### 2.3 セッション失効フロー

```
[Frontend API call any endpoint]
   │
   │ 1. Authorization: Bearer <expired IdToken>
   ▼
[API Gateway JWT Authorizer]
   │
   │ 2. exp < now → 401 Unauthorized
   ▼
[Frontend apiClient interceptor (P-RES-01)]
   │
   │ 3. 401 検出 → triggerSessionExpired()
   │ 4. sessionExpiredAtom 書込 (atomic, 二重発火防止 P-RES-02)
   ▼
[SessionExpiredModalHost (LC-11)]
   │
   │ 5. Modal 表示「お疲れ様でした…」
   │ 6. setTimeout(1500ms) → atom クリア + router.push('/login?from=session_expired')
```

---

## 3. Terraform モジュール依存関係

### 3.1 本 PR 内の依存

```
infra/envs/dev/
   └─ main.tf
        └─ module "auth" (= infra/modules/auth/)
              ├─ aws_cognito_user_pool.main
              ├─ aws_cognito_user_pool_client.web
              ├─ aws_lambda_function.pre_signup ─┐
              ├─ data.archive_file.pre_signup ──┘
              ├─ aws_apigatewayv2_api.main
              ├─ aws_apigatewayv2_authorizer.cognito
              ├─ aws_apigatewayv2_stage.default
              ├─ aws_apigatewayv2_integration.api_lambda
              ├─ aws_apigatewayv2_route.logout
              ├─ aws_lambda_function.api ─┐
              ├─ aws_lambda_permission.* ─┘
              ├─ aws_iam_role.* + aws_iam_role_policy.*
              └─ aws_cloudwatch_log_group.*
```

### 3.2 後続 PR からの参照

```
[別 PR: infra/modules/amplify/]
   └─ var.user_pool_id          ← module.auth.user_pool_id
   └─ var.user_pool_client_id   ← module.auth.user_pool_client_id
   └─ var.api_endpoint          ← module.auth.api_endpoint

[各 Unit (B/C/D/E) PR]
   └─ var.api_id                ← module.auth.api_id
   └─ var.cognito_authorizer_id ← module.auth.cognito_authorizer_id
   └─ var.api_lambda_invoke_arn ← module.auth.api_lambda_invoke_arn
   └─ var.api_lambda_role_arn   ← module.auth.api_lambda_role_arn (DynamoDB 権限を追加)
```

---

## 4. デプロイ手順

### 4.1 前提作業（Unit A PR 適用前に 1 回だけ実施）

1. **AWS アカウント準備**: Region `ap-northeast-1` でアクセス可能な IAM ユーザ
2. **Terraform State Backend bucket 作成**:
   ```
   aws s3api create-bucket --bucket gp-tfstate-dev --region ap-northeast-1 \
     --create-bucket-configuration LocationConstraint=ap-northeast-1
   aws s3api put-bucket-versioning --bucket gp-tfstate-dev --versioning-configuration Status=Enabled
   aws s3api put-bucket-encryption --bucket gp-tfstate-dev \
     --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
   ```
   このコマンドは Code Generation で `infra/scripts/bootstrap-backend.sh` として配置
3. **ECR リポジトリ作成**:
   ```
   aws ecr create-repository --repository-name gp-dev-api-image --region ap-northeast-1
   ```
4. **API Lambda 用 Docker image を build & push**（Code Generation で自動化スクリプト用意）
   - Go バイナリビルド (arm64)
   - Dockerfile: LWA + バイナリ
   - ECR タグ: `gp-dev-api-image:latest` （初回は手動 push、CI/CD は本 MVP 範囲外）

### 4.2 Terraform 適用

```
cd infra/envs/dev
terraform init
terraform plan -var="api_image_uri=<ECR URI>:<tag>"
terraform apply -var="api_image_uri=<ECR URI>:<tag>"
```

### 4.3 動作確認

1. **Sign-up 動作確認**: AWS CLI 経由で Cognito SignUp → User が `CONFIRMED` で作成されることを確認
2. **Logout 動作確認**: IdToken を取得して `POST /api/auth/logout` に送信 → 204 / ログ確認
3. **CloudWatch Logs 確認**: 構造化ログ 8 項目 (level/timestamp/userId/action/traceId/requestId/email_hash/userAgent) が出ることを確認

### 4.4 デプロイ削除

```
cd infra/envs/dev
terraform destroy -var="api_image_uri=<ECR URI>:<tag>"
```

`deletion_protection = INACTIVE` (Q-I13) のため User Pool も削除可能。

---

## 5. 環境分離戦略 (Q-I5=A 確定版)

### 5.1 本 PR で構築する環境

```
infra/envs/
├── dev/
│   ├── backend.tf      # S3 + use_lockfile (Q-I6)
│   ├── providers.tf    # default_tags (Q-I11)
│   ├── main.tf         # module "auth" 呼出
│   ├── variables.tf
│   └── outputs.tf
└── prd/
    └── README.md       # placeholder のみ
```

### 5.2 prd/README.md の内容（placeholder）

```markdown
# Production Environment

This directory is reserved for the production environment.

The MVP is designed for hackathon demo only and is not deployed to production.
When productization is decided, the following changes are required:

- Cognito `deletion_protection = "ACTIVE"` (Q-I13)
- API Gateway access logging を有効化
- WAF / Cognito Advanced Security Features 導入 (A-NFR-SEC-09)
- CloudWatch Custom Metrics + EMF 出力 (A-NFR-OBS-02 を再評価)
- CloudWatch Log retention 30 日以上に拡大 (Q-I12)
- Email 配信を SES 統合に変更 (Q-I7)
- パスワードリセット機能の追加 (要件外、本番化時に追加)
```

---

## 6. 環境変数とシークレットの取扱

### 6.1 Lambda 環境変数（[unit-interfaces.md §10](../../interfaces/unit-interfaces.md) 整合）

| 変数 | 値の出所 | 用途 |
|---|---|---|
| `AWS_REGION` | Lambda 実行時に AWS が自動付与 | SDK |
| `COGNITO_USER_POOL_ID` | `aws_cognito_user_pool.main.id` (Terraform) | API Lambda が claims 検証参考用 / 他 Unit からも参照 |
| `COGNITO_APP_CLIENT_ID` | `aws_cognito_user_pool_client.web.id` (Terraform) | 同上 |
| `LOG_LEVEL` | `info` (default) / `debug` (デバッグ時) | slog Handler の Level |
| `AWS_LWA_PORT` | `8080` (固定) | LWA 標準 |

### 6.2 シークレット

本 Unit A は **保護対象シークレットを扱わない**:
- パスワードは Cognito 内部、アプリには到達しない
- IdToken はクライアント側のみ、Lambda は claims を読むだけ
- AWS API キー類は IAM Role 経由（明示的な Secret は不要）

そのため AWS Secrets Manager / Parameter Store の利用は **本 Unit A では不要**。

---

## 7. 監視・運用 (本 MVP の最小構成)

### 7.1 観測ポイント

| 項目 | 観測手段 |
|---|---|
| Pre Sign-up Lambda 失敗 | CloudWatch Logs `/aws/lambda/gp-dev-presignup-fn` の ERROR 検索 |
| API Lambda エラー | CloudWatch Logs `/aws/lambda/gp-dev-api-fn` の ERROR 検索 |
| API Gateway 4xx/5xx | API Gateway Console で確認（detailed_metrics 無効、A-NFR-OBS-02） |
| Cognito Sign-in 失敗 | CloudWatch Metrics の Cognito 標準メトリクス (Console 確認のみ) |

### 7.2 アラート

本 MVP では CloudWatch Alarm を**設定しない**（A-NFR-OBS-02 / Q-N8）。本番化時に再評価。

---

## 8. コスト見積（dev 環境、月額）

ハッカソン審査用途想定（同時 5 user / ピーク 10 req/s）の概算:

| サービス | 想定使用量 | 月額（USD） |
|---|---|---|
| Cognito User Pool | 50 MAU 程度 | $0.00（50 MAU 無料枠） |
| Pre Sign-up Lambda | 100 invocations 程度 | $0.00（無料枠内） |
| API Lambda | 10000 invocations 程度 | $0.00（無料枠内） |
| API Gateway HTTP API | 10000 リクエスト | $0.01 |
| ECR | 1 image, 数百 MB | $0.10 |
| CloudWatch Logs | 50MB 程度 / 月 (retention 7 日) | $0.05 |
| S3 (tfstate) | 数 MB | $0.00（無料枠内） |
| **合計** | | **約 $0.16 / 月** |

---

## 9. 後続ステージへの引き継ぎ

### 9.1 Code Generation で実装するもの

1. `infra/lambdas/pre-signup/index.js` (5 行 auto-confirm)
2. `infra/lambdas/api/` の Go コード (Gin + LWA + Logout handler + middleware)
3. `infra/lambdas/api/Dockerfile` (LWA + arm64 Go バイナリ)
4. `infra/scripts/bootstrap-backend.sh` (S3 bucket 作成)
5. `infra/scripts/build-and-push-api.sh` (ECR push 自動化)
6. `infra/modules/auth/*.tf` の HCL 本体
7. `infra/envs/dev/*.tf` の HCL 本体
8. terraform-test の `*.tftest.hcl` テストファイル

### 9.2 整合性レビュー対象

Code Generation 完了後に確認:

- 既存 [unit-interfaces.md §3.3](../../interfaces/unit-interfaces.md) の `path` 表記と HTTP API での実 route が一致するか
- 既存 [unit-of-work.md §4.1](../../../inception/application-design/unit-of-work.md) の `lambda_api/` 横串記述に脚注追加が必要か
- Frontend Amplify 設定（Amplify Hosting 別 PR で構築時）が `module.auth.user_pool_id` / `user_pool_client_id` / `api_endpoint` を正しく参照できるか
