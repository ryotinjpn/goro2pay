# AI-DLC State Tracking

## Project Information
- **Project Name**: ゴロゴロPay
- **Project Type**: Greenfield
- **Start Date**: 2026-05-07T00:00:00Z
- **Current Stage**: INCEPTION - Workspace Detection (完了)

## Workspace State
- **Existing Code**: No
- **Programming Languages**: N/A
- **Build System**: N/A
- **Project Structure**: Empty (aidlc-docs/idea.md のみ存在)
- **Reverse Engineering Needed**: No
- **Workspace Root**: /Users/ryota_matsushita/work/goro2pay

## Code Location Rules
- **Application Code**: Workspace root (NEVER in aidlc-docs/)
- **Documentation**: aidlc-docs/ only
- **Structure patterns**: See code-generation.md Critical Rules

## Extension Configuration
| Extension | Enabled | Decided At |
|---|---|---|
| Security Baseline | No | Requirements Analysis |
| Property-Based Testing | Partial | Requirements Analysis |

## Stage Progress
### 🔵 INCEPTION PHASE
- [x] Workspace Detection (2026-05-07)
- [x] Requirements Analysis (2026-05-07, 要件書: aidlc-docs/inception/requirements/requirements.md)
- [x] User Stories (2026-05-07, personas.md + stories.md 生成済み、23 ストーリー / 1 ペルソナ / Epic-Based / Gherkin 受入基準)
- [x] Workflow Planning (2026-05-07, execution-plan.md 生成済み)
- [x] Application Design (2026-05-07, components.md / component-methods.md / services.md / component-dependency.md / application-design.md の 5 分冊を生成)
- [x] Units Generation (2026-05-07, unit-of-work.md / unit-of-work-dependency.md / unit-of-work-story-map.md を生成、5 Unit 構成: auth / budget / order / suggest / metrics) — ユーザー承認待ち

### 🟢 CONSTRUCTION PHASE (per-unit ループ + 最終 Build)

**Unit B `budget` 進捗** (worktree: future-unit-b)
- [x] Functional Design (Unit B, 2026-05-21) — business-logic-model / business-rules / domain-entities / frontend-components 生成済み、ユーザ承認待ち
- [x] NFR Requirements (Unit B, 2026-05-22) — nfr-requirements.md / tech-stack-decisions.md 生成済み、ユーザ承認待ち
- [x] NFR Design (Unit B, 2026-05-22) — nfr-design-patterns.md / logical-components.md 生成済み、unit-interfaces.md 更新済み
- [x] Infrastructure Design (Unit B, 2026-05-22) — infrastructure-design.md / deployment-architecture.md 生成済み
- [x] Code Generation (Unit B, 2026-05-25) — Backend (apps/api/internal/wallet/ + 4 Repo + cmd/scheduler) + Frontend (hooks 2 + components 3 + state 1 + page + tests) + Infrastructure (modules/budget/ 全 8 + tests 4) + Docs (5 サマリ)。約 50 ファイル生成。検証: Backend Go test 17 packages PASS / Frontend Vitest 75 PASS / TypeScript PASS / Terraform validate + tftest 11 + 既存 9 PASS / bootstrap 14MB build PASS。worktree feat/unit-b-code-generation。ユーザ承認待ち

**Unit A `auth` 進捗** (worktree: future-unit-a)
- [x] Functional Design (Unit A, 2026-05-21, PR #61, #66) — Cognito + Pre Sign-up + JWT 検証戦略確定
- [x] NFR Requirements (Unit A, 2026-05-21, PR #67) — Token 8h/30d、PBT Partial、Stage Throttling 100 req/s 等
- [x] NFR Design (Unit A, 2026-05-22, PR #68) — 18 パターン + 17 論理コンポーネント (LC-AUTH-01〜18)
- [x] Infrastructure Design (Unit A, 2026-05-22, PR #69) — 横串インフラ Unit A 包含 (Amplify + CodePipeline + CodeBuild + ECR + IAM 5 種) + BFF パターン + AccessToken 統一
- [x] Code Generation (Unit A, 2026-05-22) — 約 70 ファイル生成、apps/api/ + web/ + infra/、PBT 5 + Vitest 3 + tftest 6、deployment-runbook.md

**Unit C `order` 進捗** (worktree: future-unit-c, Comprehensive)
- [x] Functional Design (Unit C, 2026-05-21, PR #65 マージ済み) — business-logic-model / business-rules / domain-entities / frontend-components 生成済み、凍結契約整合修正反映済み
- [x] NFR Requirements (Unit C, 2026-05-23, PR #78 マージ済み) — nfr-requirements.md (NFRC-C01〜C25) / tech-stack-decisions.md 生成済み、Q-N1〜Q-N13 全 13 問対話ヒアリング完了
- [x] NFR Design (Unit C, 2026-05-24, PR #79 マージ済み) — nfr-design-patterns.md (P-RETRY-01 / P-PLAN-01 / P-OBS-01〜03 / P-INIT-01 / P-DI-01 / P-MOCK-01 / P-PBT-01 / P-FE-ERR-01 / P-FE-TOAST-01〜02 / P-FE-LOAD-01 / P-FE-LOCK-01 の 14 パターン) / logical-components.md (LC-ORDER-01〜34 の 34 コンポーネント) 生成済み、Q-D1〜Q-D14 全 14 問対話ヒアリング完了
- [x] Infrastructure Design (Unit C, 2026-05-24, PR #80 マージ済み) — infrastructure-design.md / deployment-architecture.md 生成済み、Q-I1〜Q-I13 全 13 問推奨案で確定（ユーザ要請の一括回答 + Q-I1/Q-I11/Q-I12 を terraform-module-design 準拠に見直し）。新規 module 3 種（order_history / bedrock / observability）+ 既存 2 module（api_gateway / lambda_api）への追記
- [x] Code Generation (Unit C, 2026-05-24) — 約 101 ファイル生成 (Backend 41 + Frontend 24 + Infra 28 + Docs 8)、Step 1〜19 全完了。Backend Go テスト全パス、Frontend Vitest 44 全パス、Terraform tftest 12 全パス。ユーザ承認待ち

**他 Unit (worktree 別、Unit D/E)**
- [ ] Functional Design (per-unit, Unit D/E) - **EXECUTE**
- [ ] NFR Requirements (per-unit, Unit D/E) - **EXECUTE**
- [ ] NFR Design (per-unit, Unit D/E) - **EXECUTE**
- [ ] Infrastructure Design (per-unit, Unit D/E) - **EXECUTE**
- [ ] Code Generation (per-unit, Unit D/E) - **EXECUTE**

**全 Unit 統合**
- [ ] Build and Test - **EXECUTE**

### 🟡 OPERATIONS PHASE
- [ ] Operations - **PLACEHOLDER**

## Key Decisions (要件分析より)
- **Business Intent**: 「人をダメにする」を最優先価値として定義、ダメ化UXを独自NFRとして設定
- **Project Deadline**: 2026-05-10 までに Inception フェーズ完了必須
- **Frontend**: PWA（AWS Amplify Hosting 配信）
- **Backend**: AWS サーバレス（Lambda + API Gateway + DynamoDB + Cognito + Bedrock + EventBridge Scheduler）
- **Auth**: Amazon Cognito User Pool（メール+パスワード）
- **AI**: Amazon Bedrock Claude (Converse API, Lambda から直接呼び出し)
- **Payment**: 仮想ウォレット（DynamoDB 上の数値、実決済なし）、月初 EventBridge Scheduler でリセット
- **External Integrations**: アダプタ層パターン、初期実装は `MockDeliveryAdapter` のみ
- **IaC**: Terraform（プロジェクトの Terraform 規約プラグインに準拠）
- **Region**: ap-northeast-1 のみ
- **Target**: 日本国内のみ / 日本語のみ / 円建て
- **Minimum Use Case**: 「ご飯めんどくさい」→ モックデリバリー注文 + 予算引き落とし
- **Scope Out**: プッシュ通知 / 音声入力 / ネイティブアプリ / 実外部API / 実決済 / 多言語対応 / 厳格コンプライアンス
