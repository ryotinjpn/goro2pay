# Auth Unit — Code Generation Summary

**Date**: 2026-05-22
**Stage**: Code Generation 完了 (Step 1〜20)

## 生成成果物

### Backend (Go + Gin + LWA)

| ファイル | 役割 |
|---|---|
| `apps/api/go.mod` | Module定義 (Go 1.22, gin v1.10, gopter, testify) |
| `apps/api/main.go` | Gin 起動 + middleware 登録 + route 配線 |
| `apps/api/Dockerfile` | Multi-stage build (Go builder → al2023-arm64 + LWA v1.0.0) |
| `apps/api/buildspec.yml` | CodeBuild 用 (docker build + ECR push + lambda update) |
| `apps/api/internal/auth/email.go` | Normalize / Hash (SHA256 hex) |
| `apps/api/internal/auth/email_test.go` | gopter PBT 5 プロパティ + Example tests |
| `apps/api/internal/auth/middleware.go` | AttachUserID + UserIDFromContext |
| `apps/api/internal/auth/middleware_test.go` | claims 正常/異常 (R-JWT-4) テスト |
| `apps/api/internal/logging/middleware.go` | RequestContextMiddleware |
| `apps/api/internal/logging/middleware_test.go` | requestId/traceId/userAgent 注入テスト |
| `apps/api/internal/logging/handler.go` | ContextAwareSlogHandler (8 項目構造化ログ) |
| `apps/api/internal/logging/handler_test.go` | bytes.Buffer 差替で JSON 内容 assert |
| `apps/api/internal/handlers/health.go` | GET /health (認証不要) |
| `apps/api/internal/handlers/logout.go` | POST /api/auth/logout (監査ログ + 204) |
| `apps/api/internal/handlers/logout_test.go` | 204 + ログ JSON 出力テスト |
| `apps/api/internal/apperrors/errors.go` | ErrUnauthorized sentinel |
| `apps/api/README.md` | apps/api/ 内のビルド手順 |

### Frontend (Next.js 15 + Amplify Auth v6 + Jotai + TanStack Query)

| ファイル | 役割 |
|---|---|
| `web/package.json` / `tsconfig.json` / `next.config.js` | プロジェクト設定 |
| `web/amplify.yml` | Amplify monorepo build (appRoot: web) |
| `web/vitest.config.ts` / `web/tests/setup.ts` | Vitest 設定 |
| `web/lib/amplifyConfig.ts` | Amplify.configure 設定 |
| `web/lib/authMessages.ts` | AuthErrorCode → 日本語ダメ化メッセージ |
| `web/lib/apiClient.ts` | Browser fetch ラッパ + 401/429/5xx interceptor |
| `web/lib/authHubListener.ts` | Amplify Hub `tokenRefresh_failure` 購読 |
| `web/state/auth.ts` | sessionExpiredAtom (Jotai) |
| `web/hooks/useAuth.ts` | useAuth フック (signup / login / logout) |
| `web/components/auth/LandingScreen.tsx` | / 未認証時 |
| `web/components/auth/LoginScreen.tsx` | /login |
| `web/components/auth/SignupScreen.tsx` | /signup (PW 強度ヒント) |
| `web/components/auth/LogoutButton.tsx` | ヘッダボタン |
| `web/components/auth/LogoutConfirmModal.tsx` | 確認モーダル |
| `web/components/auth/SessionExpiredModalHost.tsx` | セッション失効 modal |
| `web/components/auth/AuthGuard.tsx` | 認証必須 + 300ms Loading 遅延 |
| `web/app/layout.tsx` / `providers.tsx` | RootLayout + AppProviders |
| `web/app/page.tsx` | / Landing/Main 切替 |
| `web/app/login/page.tsx` | /login |
| `web/app/(auth)/signup/page.tsx` | /signup |
| `web/app/(authenticated)/layout.tsx` | AuthGuard 適用 |
| `web/app/api/[...path]/route.ts` | BFF catch-all proxy |
| `web/tests/apiClient.test.ts` | Authorization 付与 / 401 / 二重発火防止 / 429 / 5xx |
| `web/tests/authMessages.test.ts` | 9 種別メッセージ / PII 漏洩防止 |
| `web/tests/sessionExpired.test.tsx` | atom 起動 + 1.5s タイマー → router.push |
| `web/README.md` | Frontend ビルド手順 |

### Infra (Terraform)

| ファイル | 役割 |
|---|---|
| `infra/lambdas/pre-signup/index.js` | 5 行 auto-confirm |
| `infra/scripts/bootstrap-backend.sh` | S3 tfstate bucket 作成 |
| `infra/scripts/bootstrap-ecr-initial.sh` | ECR :bootstrap 初回 push |
| `infra/lambdas/pre-signup/index.js` | 5 行 auto-confirm |
| `infra/scripts/bootstrap-backend.sh` | S3 tfstate bucket 作成 |
| `infra/scripts/bootstrap-ecr-initial.sh` | ECR :bootstrap 初回 push |
| **`infra/modules/cognito/`** (Auth Unit 所有) | User Pool + App Client (Token 8h/30d) + Pre Sign-up Lambda (Node.js arm64) + IAM。main.tf / variables.tf / cognito.tf / pre_signup_lambda.tf / iam.tf / outputs.tf / README.md / tests/cognito_basic.tftest.hcl |
| **`infra/modules/api_gateway/`** (Unit 横串) | HTTP API + JWT Authorizer (TTL 60s) + Stage Throttling + Logout/Health route + 共通 integration。main.tf / variables.tf / api_gateway.tf / routes.tf / outputs.tf / README.md / tests/api_gateway_basic.tftest.hcl |
| **`infra/modules/lambda_api/`** (Unit 横串) | API Lambda (image_uri = :bootstrap, ignore_changes) + ECR + CodePipeline + CodeBuild + S3 artifacts + IAM Role 3 種。main.tf / variables.tf / api_lambda.tf / ecr.tf / codepipeline.tf / iam.tf / outputs.tf / README.md / tests/lambda_api_basic.tftest.hcl |
| **`infra/modules/amplify/`** (Unit 横串) | Amplify App + Branch (Next.js SSR、env: NEXT_PUBLIC_* + server-only API_ENDPOINT + AMPLIFY_MONOREPO_APP_ROOT=web) + SSR Role。main.tf / variables.tf / amplify.tf / iam.tf / outputs.tf / README.md / tests/amplify_basic.tftest.hcl |
| `infra/envs/dev/backend.tf` | S3 + use_lockfile |
| `infra/envs/dev/providers.tf` | default_tags |
| `infra/envs/dev/main.tf` | 4 module 呼出 (cognito / api_gateway / lambda_api / amplify) + CodeStar Connection (lambda_api と amplify で共有) |
| `infra/envs/dev/locals.tf` / `outputs.tf` | env / region / GitHub 値の固定 + 主要 outputs |
| `infra/envs/prd/README.md` | placeholder |

### ルート

| ファイル | 役割 |
|---|---|
| `.gitignore` | node_modules / .next / .terraform / *.tfstate 等を除外 |

### ドキュメント (aidlc-docs/construction/auth/code/)

| ファイル | 役割 |
|---|---|
| `business-logic-summary.md` | Step 2-4 の生成サマリ |
| `api-layer-summary.md` | Step 5-7 の生成サマリ |
| `repository-layer-summary.md` | Unit A はスキップ (Step 8) |
| `frontend-summary.md` | Step 9-11 の生成サマリ |
| `database-summary.md` | Unit A は対象外 (Step 12) |
| `deployment-runbook.md` | 初回デプロイ手順 / 通常運用 / トラブルシュート (Step 19) |
| `code-generation-summary.md` | 本ドキュメント、全成果物一覧 (Step 20) |

## ストーリー対応状況

| Story | 対応コード |
|---|---|
| **US-0-01** (Signup) | `SignupScreen.tsx` + `useAuth.signup` (auto-confirm + 自動 signIn)、`Pre Sign-up Lambda` |
| **US-0-02** (Login) | `LoginScreen.tsx` + `useAuth.login`、Cognito Token Validity 8h |
| **FR-AUTH-04** (Logout) | `LogoutButton.tsx` + `LogoutConfirmModal.tsx` + `useAuth.logout` + `internal/handlers/logout.go` |

全 Story を実装済み。E2E (US-0-01 受入: 総タップ 5 回以下、CONFIRMED 状態作成、メイン画面遷移 etc.) は Build & Test ステージで Playwright で検証する。

## 主要な設計判断の反映

| 判断 | 実コード反映 |
|---|---|
| BFF パターン | `web/app/api/[...path]/route.ts` catch-all proxy + `web/lib/apiClient.ts` |
| AccessToken 採用 | `useAuth` で `session.tokens?.accessToken` 取得、`apiClient` で Authorization: Bearer |
| email_hash 認証前限定 | middleware では emailHash 生成しない (handler.go の context 参照のみ) |
| Cognito auto-confirm | `infra/lambdas/pre-signup/index.js` 5 行 |
| Authorizer JWT (Cognito 連携) | `api_gateway.tf` aws_apigatewayv2_authorizer |
| CodePipeline + CodeBuild | `codepipeline.tf` + `apps/api/buildspec.yml` |
| AMPLIFY_MONOREPO_APP_ROOT=web | `amplify.tf` aws_amplify_branch.environment_variables |
| Token Validity 8h/30d | `cognito.tf` aws_cognito_user_pool_client |
| Stage Throttling 100 req/s, Burst 200 | `api_gateway.tf` aws_apigatewayv2_stage |
| Authorizer TTL 60s | `api_gateway.tf` |
| ECR :bootstrap + lifecycle.ignore_changes [image_uri] | `api_lambda.tf` |
| LWA v1.0.0 GA (公式 README 推奨) | `apps/api/Dockerfile` |
| Inception §4.1 整合 (apps/api/) | ディレクトリ構造、CodeBuild source path |

## 想定外の論点 (Build & Test ステージへの引き継ぎ)

- `go test ./...` 実行 (PBT + ユニット + 統合)
- `npm install && npm test` (Vitest 単体)
- `npm run e2e` (Playwright E2E、CodeStar 承認 + ECR push 後、実環境に対して)
- `terraform fmt` / `terraform validate` / `terraform test` 実行
- bootstrap スクリプト実行確認
- 初回デプロイ動作確認 (deployment-runbook.md §1〜8 順次実行)
