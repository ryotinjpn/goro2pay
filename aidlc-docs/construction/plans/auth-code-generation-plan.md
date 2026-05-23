# Auth Unit — Code Generation Plan

**Document Version**: 0.1 (Draft, awaiting user approval)
**Created**: 2026-05-22
**Unit**: A (`auth` / 認証)
**Construction Depth**: Standard
**Stage**: Code Generation (Construction Phase)
**Prerequisite**: Functional Design / NFR Requirements / NFR Design / Infrastructure Design 全て承認済み (PR #61, #66, #67, #68, #69 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit A（認証）の **Code Generation ステージ** を遂行するための作業計画を定義する。設計フェーズ（Functional Design / NFR Design / Infrastructure Design）で凍結した仕様に従い、実コード・設定・テスト・スクリプトを生成する。

### 1.1 ストーリー範囲（unit-of-work-story-map.md より）

Unit A が実装するストーリー:

| Story ID | 内容 | 受入基準ポイント |
|---|---|---|
| **US-0-01** | 新規登録でゴロゴロPay を始める | 5 タップ以下、Cognito CONFIRMED、ログイン済みでメイン画面遷移 |
| **US-0-02** | ログインで前回の続きから始める | 1 時間以上のセッション維持、誤入力エラー表示 |

加えて以下の機能要件:
- FR-AUTH-01〜05（認証全般）
- FR-AUTH-04 ログアウト機能（ストーリー化されていないが MUST、Q-A3=A で実装範囲に含める）

### 1.2 Unit 依存と契約

[unit-interfaces.md §2](../../interfaces/unit-interfaces.md):
- **依存先**: なし（最下位レイヤ）
- **依存元**: 全 Unit (B/C/D/E) が `AttachUserID` middleware と `userId` を前提
- **公開 Interface**: `AttachUserID()` / `UserIDFromContext()` / `ErrUnauthorized`
- **公開 Output (Terraform module)**: `user_pool_id` / `user_pool_client_id` / `api_id` / `api_endpoint` / `api_lambda_invoke_arn` / `api_lambda_role_arn` / `cognito_authorizer_id` / `amplify_default_domain` / `ecr_repository_url`

### 1.3 横串インフラの責務（Q-I10=A4 / Q-I14 / Q-I15）

Unit A の Code Generation で以下の横串リソースの実コード・Terraform を生成する:

- API Gateway HTTP API 本体（後続 Unit B/C/D/E が route 追加）
- API Lambda 本体（Go + Gin + LWA、Hello World + Logout 実装）
- Amplify Hosting（Next.js App Router 配信）
- CodePipeline + CodeBuild（API Lambda CD）
- ECR Repository
- CodeStar Connection (GitHub)
- IAM Roles 5 種

---

## 2. プロジェクト構造（生成対象、Inception §4.1 整合）

```
goro2pay/
├── apps/
│   ├── api/                              # API Lambda (Go + Gin + LWA)
│   │   ├── Dockerfile                    # LWA + arm64 Go バイナリ
│   │   ├── buildspec.yml                 # CodeBuild build 仕様
│   │   ├── go.mod
│   │   ├── go.sum                        # (go mod tidy で自動生成)
│   │   ├── main.go                       # Gin 起動 + middleware 登録 + handler 配線
│   │   ├── internal/
│   │   │   ├── auth/
│   │   │   │   ├── middleware.go         # AttachUserID + UserIDFromContext + ErrUnauthorized
│   │   │   │   ├── middleware_test.go    # middleware 単体テスト
│   │   │   │   ├── email.go              # Normalize + Hash (SHA256 hex)
│   │   │   │   └── email_test.go         # PBT (gopter) + Example-based テスト
│   │   │   ├── logging/
│   │   │   │   ├── handler.go            # ContextAwareSlogHandler
│   │   │   │   ├── handler_test.go       # Handler 単体テスト (bytes.Buffer 差替)
│   │   │   │   ├── middleware.go         # RequestContextMiddleware
│   │   │   │   └── middleware_test.go    # middleware 単体テスト
│   │   │   ├── handlers/
│   │   │   │   ├── health.go             # GET /health (認証不要)
│   │   │   │   ├── logout.go             # POST /api/auth/logout
│   │   │   │   └── logout_test.go        # logout handler 単体テスト
│   │   │   └── apperrors/
│   │   │       └── errors.go             # ErrUnauthorized 等の sentinel error
│   │   └── README.md                     # apps/api/ 内のビルド・起動手順
│   └── scheduler/                        # 将来 Unit B 担当 (本 PR では作成しない)
│
├── web/                                  # Next.js App Router (Amplify Hosting)
│   ├── package.json                      # Next.js 15 + Amplify v6 + Jotai + TanStack Query
│   ├── tsconfig.json
│   ├── next.config.js
│   ├── amplify.yml                       # Amplify monorepo build (appRoot: web)
│   ├── app/
│   │   ├── layout.tsx                    # 横串 Provider 群 (AmplifyConfig + Jotai + Query)
│   │   ├── page.tsx                      # / : LandingScreen / MainScreen 切替 (Unit A 部分のみ実装)
│   │   ├── login/
│   │   │   └── page.tsx                  # /login: LoginScreen
│   │   ├── (auth)/
│   │   │   └── signup/
│   │   │       └── page.tsx              # /signup: SignupScreen
│   │   ├── (authenticated)/
│   │   │   └── layout.tsx                # AuthGuard 適用レイアウト
│   │   └── api/
│   │       └── [...path]/
│   │           └── route.ts              # catch-all proxy (BFF)
│   ├── components/
│   │   └── auth/
│   │       ├── LandingScreen.tsx
│   │       ├── LoginScreen.tsx
│   │       ├── SignupScreen.tsx
│   │       ├── LogoutButton.tsx
│   │       ├── LogoutConfirmModal.tsx
│   │       ├── SessionExpiredModalHost.tsx
│   │       └── AuthGuard.tsx
│   ├── hooks/
│   │   └── useAuth.ts                    # Unit A の中核フック
│   ├── state/
│   │   └── auth.ts                       # sessionExpiredAtom
│   ├── lib/
│   │   ├── apiClient.ts                  # Browser 側 fetch ラッパ (Authorization: Bearer)
│   │   ├── authMessages.ts               # AuthErrorCode → 日本語メッセージ
│   │   ├── authHubListener.ts            # Amplify Hub: tokenRefresh_failure
│   │   └── amplifyConfig.ts              # Amplify.configure 設定
│   └── tests/
│       └── (Vitest / Playwright 設定 + テストコード)
│
├── infra/
│   ├── modules/
│   │   └── auth/                         # Q-I1=A 単一モジュール
│   │       ├── README.md                 # bootstrap 手順含む
│   │       ├── main.tf
│   │       ├── cognito.tf
│   │       ├── pre_signup_lambda.tf
│   │       ├── api_gateway.tf
│   │       ├── api_lambda.tf
│   │       ├── ecr.tf
│   │       ├── logout_route.tf
│   │       ├── amplify.tf
│   │       ├── codepipeline.tf
│   │       ├── iam.tf
│   │       ├── log_groups.tf
│   │       ├── variables.tf
│   │       ├── outputs.tf
│   │       └── tests/
│   │           ├── auth_basic.tftest.hcl
│   │           ├── auth_outputs.tftest.hcl
│   │           ├── auth_cognito_password_policy.tftest.hcl
│   │           ├── auth_lambda_lifecycle.tftest.hcl
│   │           ├── auth_amplify_branch.tftest.hcl
│   │           └── auth_codebuild_iam.tftest.hcl
│   ├── envs/
│   │   ├── dev/
│   │   │   ├── backend.tf                # S3 + use_lockfile
│   │   │   ├── providers.tf              # default_tags
│   │   │   ├── main.tf                   # auth module 呼出
│   │   │   ├── variables.tf
│   │   │   ├── outputs.tf
│   │   │   └── terraform.tfvars.example
│   │   └── prd/
│   │       └── README.md                 # placeholder
│   ├── lambdas/
│   │   └── pre-signup/
│   │       └── index.js                  # 5 行 Node.js auto-confirm
│   └── scripts/
│       ├── bootstrap-backend.sh          # S3 tfstate bucket 作成
│       └── bootstrap-ecr-initial.sh      # ECR 初回 image push
│
└── (aidlc-docs/, README.md, .gitignore 等は既存)
```

---

## 3. 作業手順（番号付きステップ、Plan 承認後に実行）

各ステップは「実装 → 単体テスト → サマリ」の小ループを取る。サマリは `aidlc-docs/construction/auth/code/` 配下に Markdown で記録（コード本体ではなく要約）。

### Step 1: プロジェクト構造セットアップ (Greenfield)
- [x] `apps/api/`, `web/`, `infra/` の各ディレクトリ生成
- [x] `apps/api/go.mod`（module name: `github.com/ryotinjpn/goro2pay/apps/api`、Go 1.22）
- [x] `web/package.json`（Next.js 15 / React 19 / Amplify v6 / Jotai 2 / TanStack Query 5 / TypeScript 5）
- [x] `web/tsconfig.json` / `next.config.js` 雛形
- [x] リポジトリルート `.gitignore` 追記（`web/node_modules/`、`web/.next/`、`apps/api/dist/` 等）
- [x] `apps/api/README.md` / `web/README.md`（簡易）

### Step 2: Backend Business Logic Generation
- [x] `apps/api/internal/auth/email.go` — `Normalize(email string) string` / `Hash(email string) string`（SHA256 hex）
- [x] `apps/api/internal/apperrors/errors.go` — `ErrUnauthorized` 等の sentinel error

**ストーリー対応**: US-0-01 / US-0-02 の email 正規化基盤、A-NFR-SEC-08 整合

### Step 3: Backend Business Logic Unit Testing (PBT)
- [x] `apps/api/internal/auth/email_test.go` — gopter (`github.com/leanovate/gopter`) ベースの PBT 5 プロパティ:
  - `Normalize` のべき等性
  - `Normalize` の大文字小文字不問
  - `Hash` の長さ不変（64 hex chars）
  - `Hash` の `Normalize` 整合
  - `Hash` の衝突なし（高確率）
- [x] Example-based ユニットテスト（境界値、Unicode、空文字等）

**ストーリー対応**: A-NFR-TEST-02

### Step 4: Backend Business Logic Summary
- [x] `aidlc-docs/construction/auth/code/business-logic-summary.md` — 生成ファイル、関数シグネチャ、テスト結果概要

### Step 5: Backend API Layer Generation (middleware + handlers)
- [x] `apps/api/internal/logging/middleware.go` — RequestContextMiddleware（requestId/traceId/userAgent を context 注入）
- [x] `apps/api/internal/logging/handler.go` — ContextAwareSlogHandler（context から attrs 抽出 + JSON 出力）
- [x] `apps/api/internal/auth/middleware.go` — AttachUserID + UserIDFromContext（AccessToken claims から sub 抽出）
- [x] `apps/api/internal/handlers/health.go` — `GET /health` (認証不要、`{status:"ok"}`)
- [x] `apps/api/internal/handlers/logout.go` — `POST /api/auth/logout`（slog で監査ログ記録、204 を返す）
- [x] `apps/api/main.go` — Gin 起動、middleware 登録順序（RequestContext → AttachUserID）、route 配線、LWA でポート 8080 listen

**ストーリー対応**: FR-AUTH-04 / Functional Design F-3 / F-4 / NFR Design LC-AUTH-01/04/05/06

### Step 6: Backend API Layer Unit Testing
- [x] `apps/api/internal/logging/middleware_test.go` — context 注入の検証
- [x] `apps/api/internal/logging/handler_test.go` — `bytes.Buffer` 差替で JSON 出力 8 項目（emailHash 除く）assert
- [x] `apps/api/internal/auth/middleware_test.go` — claims 正常/異常パスの assert
- [x] `apps/api/internal/handlers/logout_test.go` — 204 返却 + ログ出力 assert

### Step 7: Backend API Layer Summary
- [x] `aidlc-docs/construction/auth/code/api-layer-summary.md`

### Step 8: Backend Repository Layer
- [x] **Unit A は Repository を所有しない**（unit-interfaces §2 / unit-of-work §3.1）。スキップ
- [x] サマリ: `aidlc-docs/construction/auth/code/repository-layer-summary.md` に「Unit A は Repository なし」と明記

### Step 9: Frontend Components Generation
- [x] `web/lib/amplifyConfig.ts` — Amplify.configure 設定（NEXT_PUBLIC_USER_POOL_ID 等読み出し）
- [x] `web/lib/authMessages.ts` — AuthErrorCode → 日本語ダメ化トーン軽メッセージ（9 種別）
- [x] `web/lib/authHubListener.ts` — Amplify Hub `tokenRefresh_failure` 購読 → triggerSessionExpired
- [x] `web/lib/apiClient.ts` — Browser 側 fetch ラッパ（fetchAuthSession で AccessToken → Authorization: Bearer 付与、401/429 handling）
- [x] `web/state/auth.ts` — sessionExpiredAtom（Jotai）
- [x] `web/hooks/useAuth.ts` — useAuth フック（status / user / signup / login / logout）
- [x] `web/components/auth/LandingScreen.tsx`
- [x] `web/components/auth/LoginScreen.tsx`（form + error 表示 + autoComplete + data-testid）
- [x] `web/components/auth/SignupScreen.tsx`（PW 強度ヒント + form + autoComplete + data-testid）
- [x] `web/components/auth/LogoutButton.tsx` + `LogoutConfirmModal.tsx`
- [x] `web/components/auth/SessionExpiredModalHost.tsx`（atom 購読 + setTimeout 1500ms + clearTimeout）
- [x] `web/components/auth/AuthGuard.tsx`（300ms Loading 遅延、status による分岐）
- [x] `web/app/layout.tsx` — AppProviders（Amplify config + Jotai + QueryClient + SessionExpiredModalHost）
- [x] `web/app/page.tsx` — `/`: useAuth status で LandingScreen / MainScreen 切替（MainScreen は Unit A スコープ外なので placeholder）
- [x] `web/app/login/page.tsx` — LoginScreen 配置
- [x] `web/app/(auth)/signup/page.tsx` — SignupScreen 配置
- [x] `web/app/(authenticated)/layout.tsx` — AuthGuard 適用
- [x] `web/app/api/[...path]/route.ts` — catch-all proxy (BFF、Authorization 透過、API_ENDPOINT 利用)
- [x] `web/amplify.yml` — Amplify monorepo build (appRoot: web、`AMPLIFY_MONOREPO_APP_ROOT=web` env)

**ストーリー対応**: US-0-01 / US-0-02 / Functional Design `frontend-components.md` / NFR Design LC-AUTH-08〜14, LC-AUTH-18

### Step 10: Frontend Components Unit Testing
- [x] Vitest + @testing-library/react + fast-check の設定（`web/tests/setup.ts` / `vitest.config.ts`）
- [x] `useAuth` のロジック単体テスト
- [x] `apiClient` の 401 interceptor テスト
- [x] `sessionExpiredAtom` の二重発火防止テスト
- [x] `AuthGuard` の 300ms Loading 切替テスト（jest fake timers / Vitest 同等）
- [x] `Normalize` / `emailHash` の TypeScript 版（もし Frontend にも実装するなら）の PBT (fast-check)

### Step 11: Frontend Components Summary
- [x] `aidlc-docs/construction/auth/code/frontend-summary.md`

### Step 12: Database Migration / Schema (該当なし)
- [x] **Unit A はテーブルを所有しない**。サマリのみ記録（`aidlc-docs/construction/auth/code/database-summary.md`、Unit A は Cognito User Pool のみ）

### Step 13: Pre Sign-up Lambda Code
- [x] `infra/lambdas/pre-signup/index.js` — 5 行 auto-confirm

### Step 14: Build Artifacts
- [x] `apps/api/Dockerfile` — multi-stage build（Go 1.22 builder → distroless runtime + LWA layer）
- [x] `apps/api/buildspec.yml` — CodeBuild buildspec 0.2（docker buildx → ECR push → lambda update-function-code）

### Step 15: Bootstrap Scripts
- [x] `infra/scripts/bootstrap-backend.sh` — S3 tfstate bucket 作成 + 暗号化 + Public access block
- [x] `infra/scripts/bootstrap-ecr-initial.sh` — Local docker build → ECR `:bootstrap` tag push

### Step 16: Terraform Modules (機能別 5 module、unit-of-work.md §4.1 準拠)

レビュー指摘により、当初予定していた単一 `infra/modules/auth/` は撤回し、unit-of-work.md §4.1 通り機能別 module に分割する。CodeStar Connection は CodePipeline / Amplify の双方が ARN を参照するため、独立した `codestar_connection/` module に切り出した結果、最終的に 5 module 構成 (`codestar_connection` / `cognito` / `api_gateway` / `lambda_api` / `amplify`):

#### `infra/modules/codestar_connection/` (Unit 横串)
- [x] `main.tf` / `variables.tf` / `outputs.tf` / `README.md`
- [x] CodeStar Connection (Unit=shared タグ、CodePipeline / Amplify 双方が ARN を参照)
- [x] `tests/codestar_connection_basic.tftest.hcl` — plan / outputs

#### `infra/modules/cognito/` (Auth Unit 所有)
- [x] `main.tf` / `variables.tf` / `outputs.tf` / `README.md`
- [x] `cognito.tf` — User Pool + App Client + Cognito Lambda Permission
- [x] `pre_signup_lambda.tf` — archive_file + Lambda function + Log Group
- [x] `iam.tf` — Pre Sign-up Lambda IAM Role
- [x] `tests/cognito_basic.tftest.hcl` — plan / outputs / password_policy / token_validity

#### `infra/modules/api_gateway/` (Unit 横串)
- [x] `main.tf` / `variables.tf` / `outputs.tf` / `README.md`
- [x] `api_gateway.tf` — HTTP API + Cognito JWT Authorizer + Stage Throttling
- [x] `routes.tf` — POST /api/auth/logout + GET /health + 共通 integration
- [x] `tests/api_gateway_basic.tftest.hcl` — plan / outputs / throttling / authorizer TTL

#### `infra/modules/lambda_api/` (Unit 横串)
- [x] `main.tf` / `variables.tf` / `outputs.tf` / `README.md`
- [x] `api_lambda.tf` — API Lambda (image_uri = ECR :bootstrap, lifecycle.ignore_changes = [image_uri]) + permission + Log Group
- [x] `ecr.tf` — ECR Repository + lifecycle policy
- [x] `codepipeline.tf` — CodePipeline + CodeBuild + S3 artifacts + CodeBuild Log Group
- [x] `iam.tf` — IAM Roles 3 種 (api_lambda / codepipeline_api / codebuild_api)
- [x] `tests/lambda_api_basic.tftest.hcl` — plan / outputs / image_uri :bootstrap

#### `infra/modules/amplify/` (Unit 横串)
- [x] `main.tf` / `variables.tf` / `outputs.tf` / `README.md`
- [x] `amplify.tf` — Amplify App + Branch (env vars: NEXT_PUBLIC_* + server-only API_ENDPOINT + AMPLIFY_MONOREPO_APP_ROOT)
- [x] `iam.tf` — Amplify SSR Role
- [x] `tests/amplify_basic.tftest.hcl` — plan / outputs / branch env vars (BFF パターン整合)

### Step 17: Terraform Env (`infra/envs/dev/`)
- [x] `backend.tf` — S3 + use_lockfile = true
- [x] `providers.tf` — default_tags
- [x] `locals.tf` — env / region / github_owner / github_repo / github_branch (tfvars 方式から locals.tf 方式に変更)
- [x] `main.tf` — 5 module 呼出 (codestar_connection / cognito / api_gateway / lambda_api / amplify) + `aws_lambda_permission.apigw_invoke_api` (両 module の output を必要とするため envs 側で組立、循環依存回避)
- [x] `outputs.tf` — 主要 6 種 (各 module の output を再公開)

### Step 18: Terraform Tests (各 module の tests/ ディレクトリ)
- [x] `infra/modules/codestar_connection/tests/codestar_connection_basic.tftest.hcl`
- [x] `infra/modules/cognito/tests/cognito_basic.tftest.hcl`
- [x] `infra/modules/api_gateway/tests/api_gateway_basic.tftest.hcl`
- [x] `infra/modules/lambda_api/tests/lambda_api_basic.tftest.hcl`
- [x] `infra/modules/amplify/tests/amplify_basic.tftest.hcl`

各 module の tests/ で plan 成立 + 主要 outputs + 設計判断 (password_policy / token_validity / throttling / TTL / image_uri / branch env vars) を mock_provider で検証する。

### Step 19: Documentation
- [x] `aidlc-docs/construction/auth/code/deployment-runbook.md` — 初回デプロイ手順、CodeStar 承認、ECR push、ロールバック手順
- [x] ルート `README.md` 更新（Unit A 完了後の利用手順を簡記、既存 README に追記）

### Step 20: 完了確認とサマリ
- [x] `aidlc-docs/construction/auth/code/code-generation-summary.md` — 全成果物一覧、ステップ別チェック結果
- [x] aidlc-state.md の Construction → Code Generation (Unit A) を [x]
- [x] audit.md に完了記録

---

## 4. ストーリー トレーサビリティ

| Step | 関連ストーリー / FR | 関連 NFR | 関連 LC |
|---|---|---|---|
| Step 2-3 | A-NFR-SEC-08 (email_hash トレーサビリティ) | A-NFR-TEST-02 (PBT) | LC-AUTH-02, LC-AUTH-03 |
| Step 5-6 | US-0-02 (JWT セッション維持), FR-AUTH-04 (Logout) | A-NFR-OBS-01 (構造化ログ) | LC-AUTH-01, LC-AUTH-04, LC-AUTH-05, LC-AUTH-06 |
| Step 9-10 | US-0-01 (Signup), US-0-02 (Login) | NFR-DEG-01 (低摩擦), NFR-DEG-05 (コピー), A-NFR-PERF-04 | LC-AUTH-08〜14, LC-AUTH-18 |
| Step 13 | FR-AUTH-01 (auto-confirm) | A-NFR-REL-01 | LC-AUTH-07 |
| Step 14, 16 | (横串インフラ) | A-NFR-SEC-04 (Throttling), A-NFR-OBS-01 | LC-AUTH-15/16/17 |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

- **Build & Test ステージ**: `go test ./...` / `npm test` / `terraform test` の CI 統合、E2E (Playwright) シナリオ実行、bootstrap スクリプト実行確認、初回デプロイ動作確認
- **他 Unit (B/C/D/E) Construction**: 各 Unit が `module.auth` の outputs を参照して route + integration + IAM 拡張を追加
- **本番化**: `infra/envs/prd/` 構築、deletion_protection ACTIVE、Amplify branch を main に切替、WAF / Cognito Advanced Security 導入

---

## 6. 想定スコープと工数

- 生成ファイル数: 約 **70 ファイル**（コード ~50 + Terraform ~20）
- ドキュメントサマリ: 6 種（business-logic / api-layer / repository-layer / frontend / database / code-generation 全体）
- テストファイル: ~10 種（Go ユニット + PBT、TS ユニット、Terraform tftest）

本プロジェクトは greenfield のため、新規ディレクトリ・新規ファイルが大量に発生する。Brownfield のような既存ファイル修正は基本なし（リポジトリルート README.md は追記）。

---

## 7. 承認ゲート

本 Plan の構成（生成範囲・ステップ順序・テスト方針）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — ステップ追加削除、生成範囲調整、優先順位変更等
- ✅ **Approve & Start Generation** — Plan 承認、Part 2 Generation を開始（Step 1 から順次実行）
