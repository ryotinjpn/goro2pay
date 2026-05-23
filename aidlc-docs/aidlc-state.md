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
- [ ] Code Generation (Unit B) - **EXECUTE**

**他 Unit (worktree 別)**
- [ ] Functional Design (per-unit, Unit A/C/D/E) - **EXECUTE**
- [ ] NFR Requirements (per-unit, Unit A/C/D/E) - **EXECUTE**
- [ ] NFR Design (per-unit, Unit A/C/D/E) - **EXECUTE**
- [ ] Infrastructure Design (per-unit, Unit A/C/D/E) - **EXECUTE**
- [ ] Code Generation (per-unit, Unit A/C/D/E) - **EXECUTE**

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
