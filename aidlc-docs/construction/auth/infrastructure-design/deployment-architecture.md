# Auth Unit — Deployment Architecture

**Document Version**: 1.2
**Created**: 2026-05-22
**Updated**: 2026-05-22 (Q-I14/Q-I15 追加: Amplify Hosting + CodePipeline/CodeBuild/ECR を Unit A スコープに追加)
**Updated**: 2026-05-22 (BFF パターン採用 + back/ ディレクトリリネーム + API 認証を AccessToken に統一)
**Unit**: A (`auth`)
**Stage**: Infrastructure Design / Construction
**Predecessors**: [infrastructure-design.md](./infrastructure-design.md)

本ドキュメントは Unit A の **デプロイメント時の AWS リソース構成と相互関係、デプロイ手順、環境分離戦略** を記述する。

---

## 1. AWS リソース全体図

### 1.1 Unit A PR でデプロイされる AWS リソース（Q-I10=A4 / Q-I14 / Q-I15 確定）

```
                       ┌────────────────────────────────────────────────────────┐
                       │                  ap-northeast-1 (Tokyo)                 │
                       │                                                          │
                       │  ┌─────────────── ランタイム ────────────────┐           │
                       │  │                                              │          │
   ┌────────────┐      │  │ ┌─────────────────────────────────────────┐ │          │
   │ ブラウザ    │ ─1──┼──┼►│ Amplify Hosting (gp-dev-web)            │ │          │
   │ (太郎)     │      │  │ │  Next.js App Router (SSR)               │ │          │
   └─────┬──────┘      │  │ │  branch: develop                        │ │          │
         │              │  │ │  env (Browser 露出):                    │ │          │
         │              │  │ │   NEXT_PUBLIC_USER_POOL_ID              │ │          │
         │              │  │ │   NEXT_PUBLIC_USER_POOL_CLIENT_ID       │ │          │
         │              │  │ │   NEXT_PUBLIC_AWS_REGION                │ │          │
         │              │  │ │  env (server-only, BFF):                │ │          │
         │              │  │ │   API_ENDPOINT                          │ │          │
         │              │  │ │  catch-all Route Handler:               │ │          │
         │              │  │ │   /api/[...path]/route.ts → API GW      │ │          │
         │              │  │ └─────────────────────────────────────────┘ │          │
         │              │  │                                              │          │
         │ 2: Sign-up   │  │ ┌─────────────────────────────────────────┐ │          │
         │ 2: Sign-in   │  │ │ Cognito User Pool (gp-dev-userpool)     │ │          │
         │ 2: Token取得 │  │ │  ├─ App Client (web, no secret)         │ │          │
         └─────────────►│  │ │  ├─ Pre Sign-up Lambda Trigger          │ │          │
                       │  │ │  └─ User attributes: email only         │ │          │
                       │  │ └────────────┬────────────────────────────┘ │          │
                       │  │              │                                │          │
                       │  │              ▼                                │          │
                       │  │ ┌─────────────────────────────────────────┐ │          │
                       │  │ │ Pre Sign-up Lambda (Node.js, arm64)     │ │          │
                       │  │ │ gp-dev-presignup-fn                     │ │          │
                       │  │ │  autoConfirmUser=true / verifyEmail=true│ │          │
                       │  │ └─────────────────────────────────────────┘ │          │
                       │  │                                              │          │
                       │  │ ┌─────────────────────────────────────────┐ │          │
   ┌────────────┐      │  │ │ API Gateway HTTP API (gp-dev-api)       │ │          │
   │ ブラウザ    │ ─3──┼──┼►│  Stage: $default                        │ │          │
   │ Bearer<Tok>│      │  │ │  Throttling 100 req/s / Burst 200       │ │          │
   └────────────┘      │  │ │  JWT Authorizer (Cognito, TTL 60s)      │ │          │
                       │  │ │  Route: POST /api/auth/logout           │ │          │
                       │  │ └────────────┬────────────────────────────┘ │          │
                       │  │              │ AWS_PROXY                     │          │
                       │  │              ▼                                │          │
                       │  │ ┌─────────────────────────────────────────┐ │          │
                       │  │ │ API Lambda (Go + Gin + LWA, arm64)      │ │          │
                       │  │ │ gp-dev-api-fn                           │ │          │
                       │  │ │  Hello World + Logout                   │ │          │
                       │  │ │  ENV: COGNITO_USER_POOL_ID, etc         │ │          │
                       │  │ │  lifecycle.ignore_changes=[image_uri]   │ │          │
                       │  │ └─────────────────────────────────────────┘ │          │
                       │  │                                              │          │
                       │  │ ┌─────────────────────────────────────────┐ │          │
                       │  │ │ CloudWatch Log Groups (retention 7日)  │ │          │
                       │  │ │  /aws/lambda/gp-dev-presignup-fn        │ │          │
                       │  │ │  /aws/lambda/gp-dev-api-fn              │ │          │
                       │  │ │  /aws/codebuild/gp-dev-api-build        │ │          │
                       │  │ └─────────────────────────────────────────┘ │          │
                       │  └──────────────────────────────────────────────┘          │
                       │                                                              │
                       │  ┌─────────────── CD パイプライン ─────────────┐            │
                       │  │                                              │           │
                       │  │     ┌──────────────────────────────────┐     │           │
                       │  │ ─4─►│ CodeStar Connection (GitHub)     │     │           │
                       │  │     └────────────┬─────────────────────┘     │           │
                       │  │                  │ used by                    │           │
                       │  │     ┌────────────┴─────────────────────┐     │           │
                       │  │     ▼                                  ▼     │           │
                       │  │ ┌──────────────┐         ┌──────────────────┐│           │
                       │  │ │ Amplify      │         │ CodePipeline     ││           │
                       │  │ │ branch:      │         │ gp-dev-api-      ││           │
                       │  │ │ develop      │         │ pipeline         ││           │
                       │  │ │ auto_build   │         │  Source: GitHub  ││           │
                       │  │ └──────────────┘         │   develop branch ││           │
                       │  │                          │  Build:          ││           │
                       │  │                          │   CodeBuild      ││           │
                       │  │                          └────────┬─────────┘│           │
                       │  │                                   │          │           │
                       │  │                                   ▼          │           │
                       │  │                          ┌──────────────────┐│           │
                       │  │                          │ CodeBuild        ││           │
                       │  │                          │ gp-dev-api-build ││           │
                       │  │                          │  buildspec.yml   ││           │
                       │  │                          │   - docker build ││           │
                       │  │                          │   - ECR push     ││           │
                       │  │                          │   - lambda update││           │
                       │  │                          └────────┬─────────┘│           │
                       │  │                                   │          │           │
                       │  │                                   ▼          │           │
                       │  │   ┌───────────────────┐  ┌──────────────────┐│           │
                       │  │   │ S3 (artifacts)    │  │ ECR              ││           │
                       │  │   │ gp-dev-           │  │ gp-dev-api-image ││           │
                       │  │   │ codepipeline-     │  │  :bootstrap (init)││          │
                       │  │   │ artifacts         │  │  :latest          ││          │
                       │  │   │ (lifecycle 30d)   │  │  :sha-* (commit)  ││          │
                       │  │   └───────────────────┘  └──────────────────┘│           │
                       │  └──────────────────────────────────────────────┘           │
                       │                                                              │
                       │  ┌─────────────── IAM Roles ───────────────────┐             │
                       │  │  - gp-dev-presignup-role (logs only)        │             │
                       │  │  - gp-dev-api-role (logs only, 他Unitで拡張)│             │
                       │  │  - gp-dev-amplify-ssr-role (logs only)      │             │
                       │  │  - gp-dev-codepipeline-role (S3+CB+CSC)     │             │
                       │  │  - gp-dev-codebuild-role (logs+ECR+λ update)│             │
                       │  └──────────────────────────────────────────────┘            │
                       │                                                              │
                       └──────────────────────────────────────────────────────────────┘

                       ┌──────────────────────────────────────────┐
                       │  GitHub: ryotinjpn/goro2pay              │
                       │   - branch: develop  (push triggers CD)  │
                       └──────────────────────────────────────────┘

凡例:
   1 = ブラウザが Amplify Hosting から Frontend を取得
   2 = Cognito 認証フロー (Amplify Auth ライブラリ経由、ブラウザ → Cognito 直接)
   3 = 業務 API 呼出 (BFF パターン: ブラウザ → /api/* (同一オリジン)
       → Next.js catch-all Route Handler が server-only env API_ENDPOINT
       で API Gateway に proxy → API Lambda)
   4 = GitHub push → Amplify auto build (Frontend) + CodePipeline (API Lambda CD)

BFF パターンの利点:
   - API Gateway URL がブラウザバンドルに露出しない (server-only env)
   - CORS 設定不要 (ブラウザ ↔ Next.js は同一オリジン)
   - Authorization ヘッダ付与を Next.js Server で一元化
```

### 1.2 後続 PR で追加されるリソース（参考）

```
横串インフラは Unit A で全構築済み (Q-I14/I15)。後続 PR で追加されるのは:


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
   │ 5. POST /api/auth/logout (Authorization: Bearer <AccessToken>)
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
   │ 1. Authorization: Bearer <expired AccessToken>
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

### 2.4 CD フロー（Q-I14 / Q-I15）

```
[開発者]
   │
   │ 1. develop ブランチへ git push
   ▼
[GitHub: ryotinjpn/goro2pay]
   │
   │ 2. CodeStar Connection 経由で Webhook 通知
   │
   ├──────────────────────────────────────┐
   │                                       │
   ▼                                       ▼
[Amplify Hosting]                  [CodePipeline (gp-dev-api-pipeline)]
   │                                       │
   │ 3a. develop branch を pull           │ 3b. Source Stage: GitHub develop branch を pull
   │ 4a. npm ci && npm run build (web/)   │ 4b. S3 artifact bucket に source.zip 保存
   │ 5a. Next.js build artifacts を deploy │ 5b. Build Stage: CodeBuild 起動
   │ 6a. 公開 URL 更新                    │     ↓
   ▼                                       │   ┌─ docker buildx build --platform linux/arm64
[Amplify default domain で Frontend 公開] │   ├─ ECR login + docker push :sha-XXX, :latest
                                           │   └─ aws lambda update-function-code --image-uri
                                           ▼
                                   [API Lambda (gp-dev-api-fn) image_uri 更新]
                                           │
                                           │ 6b. 次の cold start 時に新 image を起動
                                           │ 7b. 以降の API リクエストは新コード動作
                                           ▼
                                   [動作確認]
```

push 〜 デプロイ完了まで: Frontend ~3-5 分、API Lambda ~5-8 分 (Docker build + push + update)。

### 2.5 Pre Sign-up Lambda 更新フロー（手動運用）

CD なしの暗黙手動運用。

```
[開発者]
   │
   │ 1. infra/lambdas/pre-signup/index.js を編集
   │
   ▼
[ローカル端末]
   │
   │ 2. cd infra/envs/dev && terraform apply
   ▼
[Terraform]
   │
   │ 3. archive_file が再 zip 化、source_code_hash 変化検出
   │ 4. aws_lambda_function.pre_signup を update
   ▼
[Pre Sign-up Lambda 更新完了]
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
              ├─ aws_ecr_repository.api ─┐
              ├─ aws_ecr_lifecycle_policy.api ┘
              ├─ aws_lambda_function.api ─┐
              │   (lifecycle.ignore_changes = [image_uri])
              ├─ aws_codestarconnections_connection.github
              ├─ aws_amplify_app.web ─┐
              ├─ aws_amplify_branch.develop ┘
              ├─ aws_codepipeline.api ─┐
              ├─ aws_codebuild_project.api ─┤
              ├─ aws_s3_bucket.codepipeline_artifacts ┘
              ├─ aws_lambda_permission.* ─┘
              ├─ aws_iam_role.* + aws_iam_role_policy.*
              └─ aws_cloudwatch_log_group.*
```

### 3.2 後続 PR からの参照

```
[Unit A 完結済み: infra/modules/amplify/ も Auth module に同居]
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
   bash infra/scripts/bootstrap-backend.sh
   ```
   内部で実行されるコマンド:
   ```
   aws s3api create-bucket --bucket gp-tfstate-dev --region ap-northeast-1 \
     --create-bucket-configuration LocationConstraint=ap-northeast-1
   aws s3api put-bucket-versioning --bucket gp-tfstate-dev --versioning-configuration Status=Enabled
   aws s3api put-bucket-encryption --bucket gp-tfstate-dev \
     --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
   aws s3api put-public-access-block --bucket gp-tfstate-dev \
     --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
   ```

### 4.2 Terraform 適用 (1 回目)

ECR Repository を含む全リソースを Terraform で構築。**この時点では Lambda には image がまだ無いため、Lambda 自体の作成は skip される設計**:

```
cd infra/envs/dev
terraform init
terraform plan -var="github_owner=<your-github-account>"
terraform apply -var="github_owner=<your-github-account>"
```

**注意**: 初回 apply 時は CodeStar Connection が `Pending` 状態。次のステップで承認する。

### 4.3 CodeStar Connection の承認 (手動、1 回のみ)

AWS Console で以下を実行:
1. Developer Tools → Settings → Connections を開く
2. `gp-dev-github-conn` を選択
3. "Update pending connection" をクリック
4. GitHub App をインストール（goro2pay リポジトリへのアクセス許可）
5. 状態が `Available` になることを確認

### 4.4 ECR 初回 image push

CodePipeline が動作する前に Lambda が起動できるよう、bootstrap image を push:

```
bash infra/scripts/bootstrap-ecr-initial.sh
```

内部処理:
1. Local で `infra/lambdas/api/Dockerfile` を arm64 build
2. ECR にログイン
3. `:bootstrap` タグで push
4. このタグが Lambda の初期 `image_uri` となる

### 4.5 動作確認

1. **Frontend 自動デプロイ確認**: `develop` ブランチへ commit / push → Amplify Hosting で auto build / deploy が走ることを確認
2. **API Lambda CD 確認**: 同じ push で CodePipeline が起動 → CodeBuild が ECR push → Lambda image_uri が更新されることを確認
3. **Sign-up 動作確認**: Frontend の `/signup` から登録 → User が CONFIRMED で作成されることを確認
4. **Logout 動作確認**: ログイン後にログアウトボタン押下 → `POST /api/auth/logout` 204 / ログ確認
5. **CloudWatch Logs 確認**: 構造化ログ 8 項目 (level/timestamp/userId/action/traceId/requestId/email_hash/userAgent) が出ることを確認

### 4.6 デプロイ削除

```
cd infra/envs/dev
terraform destroy -var="github_owner=<your-github-account>"
```

`deletion_protection = INACTIVE` (Q-I13) のため User Pool も削除可能。Amplify App / CodePipeline / ECR / S3 Artifacts すべて自動削除。

**注意**: ECR Repository に image が残っているとデフォルトで削除エラー。`force_delete = true` を `aws_ecr_repository.api` に設定済み（dev のみ）。本番化時は `false` に切替。

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
- ECR `force_delete = false` (Q-I15)
- API Gateway access logging を有効化
- WAF / Cognito Advanced Security Features 導入 (A-NFR-SEC-09)
- CloudWatch Custom Metrics + EMF 出力 (A-NFR-OBS-02 を再評価)
- CloudWatch Log retention 30 日以上に拡大 (Q-I12)
- Email 配信を SES 統合に変更 (Q-I7)
- パスワードリセット機能の追加 (要件外、本番化時に追加)
- Amplify branch を `main` に切替、stage を PRODUCTION に変更 (Q-I14)
- CodePipeline Source branch を `main` に切替 (Q-I15)
- CORS の allow_origins を Frontend ドメインに絞る
- 横串改善: Auth module から amplify / codepipeline / lambda_api を独立 module に切り出し
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

### 6.2 Amplify Hosting (Next.js Server) 環境変数（BFF パターン整合）

| 変数 | 値の出所 | スコープ | 用途 |
|---|---|---|---|
| `NEXT_PUBLIC_USER_POOL_ID` | `aws_cognito_user_pool.main.id` | **Browser 露出 OK** | Amplify Auth がブラウザで利用 |
| `NEXT_PUBLIC_USER_POOL_CLIENT_ID` | `aws_cognito_user_pool_client.web.id` | **Browser 露出 OK** | 同上 |
| `NEXT_PUBLIC_AWS_REGION` | `ap-northeast-1` | **Browser 露出 OK** | Amplify Auth がブラウザで利用 |
| `API_ENDPOINT` | `aws_apigatewayv2_api.main.api_endpoint` | **server-only** (NEXT_PUBLIC_ なし) | Next.js catch-all Route Handler が proxy 先として使用 |

**重要**: `API_ENDPOINT` は `NEXT_PUBLIC_` プレフィックスを**付けない**。これにより Next.js のビルド時にブラウザバンドルから除外され、API Gateway URL が公開されない。catch-all Route Handler (`web/app/api/[...path]/route.ts`) のみが `process.env.API_ENDPOINT` でアクセス可能。

### 6.3 シークレット

本 Unit A は **保護対象シークレットを扱わない**:
- パスワードは Cognito 内部、アプリには到達しない
- AccessToken / IdToken はクライアント側のみ、API 認証には AccessToken を使う、Lambda は claims を読むだけ
- AWS API キー類は IAM Role 経由（明示的な Secret は不要）
- `API_ENDPOINT` は機密ではないが、ブラウザ露出を避けることで攻撃面を縮小する目的（BFF パターン）

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
| Frontend デプロイ状態 | Amplify Hosting Console |
| API Lambda CD 状態 | CodePipeline Console / CodeBuild ログ `/aws/codebuild/gp-dev-api-build` |

### 7.2 アラート

本 MVP では CloudWatch Alarm を**設定しない**（A-NFR-OBS-02 / Q-N8）。本番化時に再評価。

---

## 8. コスト見積（dev 環境、月額）

ハッカソン審査用途想定（同時 5 user / ピーク 10 req/s、CD は週 5-10 回程度の commit）の概算:

| サービス | 想定使用量 | 月額（USD） |
|---|---|---|
| Cognito User Pool | 50 MAU 程度 | $0.00（50 MAU 無料枠） |
| Pre Sign-up Lambda | 100 invocations 程度 | $0.00（無料枠内） |
| API Lambda | 10000 invocations 程度 | $0.00（無料枠内） |
| API Gateway HTTP API | 10000 リクエスト | $0.01 |
| ECR | 1 image, 数百 MB | $0.10 |
| CloudWatch Logs (Lambda + CodeBuild) | 50MB 程度 / 月 (retention 7 日) | $0.05 |
| S3 (tfstate + codepipeline-artifacts) | 数 MB | $0.00（無料枠内） |
| Amplify Hosting | 無料枠内（Build minute 1000/月、Storage 5GB/月、Data Transfer 15GB/月） | $0.00 |
| CodePipeline | $1.00/active pipeline/月（無料枠 1 pipeline/月、最初の 30 日無料） | $1.00 |
| CodeBuild | 100 build minutes 程度（5-10 build × 10-15min） | $0.50 |
| CodeStar Connection | 無料 | $0.00 |
| **合計** | | **約 $1.66 / 月** |

※ CodePipeline / CodeBuild の追加でコストが増えるが、デモ運用範囲では月 $2 未満で収まる見込み。

---

## 9. 後続ステージへの引き継ぎ

### 9.1 Code Generation で実装するもの

1. `infra/lambdas/pre-signup/index.js` (5 行 auto-confirm)
2. `back/api/` の Go コード (Gin + LWA + Logout handler + middleware) ※ Code 配置はリポジトリルート直下の `back/` ディレクトリを採用 (unit-of-work.md §4.1 の `apps/api/` 表記は本 PR 内で `back/api/` にリネーム決定、Code Generation 完了後に Inception ドキュメントへ別途反映)
3. `infra/lambdas/api/Dockerfile` (LWA + arm64 Go バイナリ build)
4. `infra/lambdas/api/buildspec.yml` (CodeBuild: docker build + ECR push + lambda update)
5. `web/` の Next.js Frontend 雛形（Auth Unit 担当ページ部分）
6. `infra/scripts/bootstrap-backend.sh` (S3 tfstate bucket 作成)
7. `infra/scripts/bootstrap-ecr-initial.sh` (ECR 初回 image push)
8. `infra/modules/auth/*.tf` の HCL 本体（cognito.tf / pre_signup_lambda.tf / api_gateway.tf / api_lambda.tf / ecr.tf / logout_route.tf / amplify.tf / codepipeline.tf / iam.tf / log_groups.tf）
9. `infra/envs/dev/*.tf` の HCL 本体（backend.tf / providers.tf / main.tf）
10. `infra/envs/prd/README.md` placeholder
11. terraform-test の `*.tftest.hcl` テストファイル 6 種

### 9.2 整合性レビュー対象

Code Generation 完了後に確認:

- 既存 [unit-interfaces.md §3.3](../../interfaces/unit-interfaces.md) の `path` 表記と HTTP API での実 route が一致するか
- 既存 [unit-of-work.md §4.1](../../../inception/application-design/unit-of-work.md) の `lambda_api/` 横串記述に脚注追加が必要か
- Frontend Amplify 設定（Amplify Hosting 別 PR で構築時）が `module.auth.user_pool_id` / `user_pool_client_id` / `api_endpoint` を正しく参照できるか
