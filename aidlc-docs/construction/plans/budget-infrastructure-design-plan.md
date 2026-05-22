# Budget Unit — Infrastructure Design Plan

**Document Version**: 0.1 (Draft, awaiting user answers)
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Construction Depth**: Standard
**Stage**: Infrastructure Design (Construction Phase)
**Prerequisite**: NFR Design 承認済み (PR #75 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit B（予算・ウォレット）の **Infrastructure Design ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。NFR Design で定義した論理コンポーネント（LC-BUDGET-01〜12）を **実 AWS リソース（Terraform）** に展開する。

### 1.1 確定済み前提（質問不要）

| 項目 | 値 | 出典 |
|---|---|---|
| IaC | Terraform | 要件 Q-12=C |
| 命名規則 | `gp-{env}-{resource}` | Unit A Q-I4 |
| タグ | `Project=goro2pay / Env=dev / ManagedBy=terraform` | Unit A Q-I11 |
| DynamoDB テーブル PK/SK | unit-interfaces.md §3.4 で凍結済み | unit-interfaces.md |
| API Gateway | Unit A が構築済み (`aws_apigatewayv2_api.main`) | Unit A Infra Design |
| Cognito Authorizer | Unit A が構築済み | Unit A Infra Design |
| API Lambda | Unit A が構築済み (`gp-{env}-api-fn`) | Unit A Infra Design |
| API Lambda IAM Role | Unit A が作成済み（Unit B で DynamoDB 権限を追加） | Unit A Infra Design |
| DynamoDB Provisioned 1 RCU / 1 WCU | NFR Requirements Q-N4=B | nfr-requirements.md §5.1 |
| Scheduler Lambda: タイムアウト 30 秒 / メモリ 128MB | NFR Requirements Q-N3=A / Q-N9=A | nfr-requirements.md §2.2 |
| EventBridge Scheduler cron | `cron(0 15 L * ? *)` UTC | Q-B8=A |
| Region | `ap-northeast-1` | NFR-COMP / Q-15 |

### 1.2 Unit B Infrastructure Design のスコープ

1. **DynamoDB テーブル** 4 本（Wallet / BudgetSettings / IdempotencyKeys / BudgetResetLog）
2. **Scheduler Lambda**（`apps/scheduler/` — EventBridge Scheduler から起動）
3. **API Gateway ルート** 2 本（`GET /api/wallet`, `POST /api/wallet/budget`）— 既存 API Gateway に追加
4. **IAM 権限追加**（API Lambda Role + Scheduler Lambda Role）
5. **Terraform モジュール** 構成
6. **CloudWatch Log Group**（Scheduler Lambda 用）

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1問ずつ）
- [x] 回答の曖昧さを点検し、必要なら追加質問を挟む
- [x] `aidlc-docs/construction/budget/infrastructure-design/infrastructure-design.md` を作成
- [x] `aidlc-docs/construction/budget/infrastructure-design/deployment-architecture.md` を作成
- [x] aidlc-state.md と audit.md を更新（承認後）
- [x] 完了メッセージを提示し、承認ゲートに進む

---

## 3. 確認質問（対話ヒアリング対象）

### Q-I1: Terraform モジュール構成

Unit B の Terraform リソース（DynamoDB + Scheduler Lambda）をどのモジュール構成で管理するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **新規 `infra/modules/budget/` モジュール** | Unit ごとに独立。他 Unit への影響ゼロ。`envs/dev/main.tf` で `module "budget"` を追加呼出 |
| B | **Unit A の `infra/modules/auth/` に追加** | ファイル数を増やさず既存モジュールに追記。Unit 分離なし |

[Answer]:

### Q-I2: Scheduler Lambda のデプロイ方式

Scheduler Lambda（`apps/scheduler/`、Go）をどのデプロイ方式で配備するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **zip（`aws-lambda-go` ランタイム）** | Go バイナリを直接 zip 化。LWA 不要（EventBridge から直接 Go handler を呼ぶ）。シンプル・軽量 |
| B | **Container image（ECR、Unit A と同じ方式）** | Dockerfile + ECR。一貫性あり。Scheduler には LWA 不要なため少し過剰 |

[Answer]:

### Q-I3: Scheduler Lambda の CI/CD

Scheduler Lambda のデプロイ方式（Q-I2）確定後、コード変更時のデプロイ方法。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **手動 `terraform apply`** のみ（CI/CD なし） | 月 1 回呼ばれる Scheduler は変更頻度が低い。ハッカソン向けにシンプル |
| B | **Unit A の CodePipeline に Scheduler の build ステップを追加** | 一元管理。API Lambda と同時 push でデプロイ |
| C | **Scheduler 専用 CodePipeline を追加** | 独立したパイプライン。実装コスト高め |

[Answer]:

### Q-I4: DynamoDB 暗号化

| 案 | パターン | 特徴 |
|---|---|---|
| A | **AWS マネージド（`AES256`、デフォルト）** | 追加コストゼロ。Unit A ECR 等と統一 |
| B | **Customer Managed KMS key** | 監査要件向け。ハッカソン規模では不要 |

[Answer]:

### Q-I5: EventBridge Scheduler の失敗ハンドリング

`ResetAll` Lambda が全体失敗（タイムアウト等）した場合の対応。個別ユーザの失敗は P-OBS-02 で ERROR ログ対応済み。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **再試行なし + CloudWatch Logs に記録のみ**（手動リカバリ） | NFR-REL-04 の方針と一致。シンプル。ハッカソン規模で十分 |
| B | **再試行 1 回 + DLQ（SQS）** | 自動リカバリ付き。実装・コスト増 |

[Answer]:

### Q-I6: `BudgetResetLog` テーブルの DynamoDB Point-In-Time Recovery (PITR)

| 案 | パターン | 特徴 |
|---|---|---|
| A | **無効**（dev 環境 / ハッカソン） | 追加コストなし |
| B | **有効** | 最大 35 日分のリカバリが可能。月 1 回リセットログの永続性を高める |

[Answer]:

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `infrastructure-design.md` | Terraform リソース詳細（DynamoDB 4 テーブル / Scheduler Lambda / API GW ルート 2 本 / IAM 権限追加）|
| `deployment-architecture.md` | Unit B のデプロイ構成図（EventBridge Scheduler → Lambda → DynamoDB フロー）|

---

## 5. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-I1 〜 Q-I6 を対話形式で順にヒアリング開始
