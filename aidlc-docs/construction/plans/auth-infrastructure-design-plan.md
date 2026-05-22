# Auth Unit — Infrastructure Design Plan

**Document Version**: 0.1 (Draft, awaiting user answers)
**Created**: 2026-05-22
**Unit**: A (`auth` / 認証)
**Construction Depth**: Standard
**Stage**: Infrastructure Design (Construction Phase)
**Prerequisite**: NFR Design 承認済み (PR #68 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit A（認証）の **Infrastructure Design ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。NFR Design で確定した論理コンポーネント（特に LC-15/16/17）を **Terraform 構成** に展開する。

### 1.1 NFR Design からの引き継ぎ事項

| LC ID | 名前 | 本ステージで Terraform 化する対象 |
|---|---|---|
| LC-AUTH-07 | `PreSignUpTriggerLambda` (Node.js) | Lambda 関数 + IAM ロール + zip パッケージング + Cognito 連携 |
| LC-AUTH-15 | `CognitoUserPoolConfig` | `aws_cognito_user_pool` リソース |
| LC-AUTH-16 | `CognitoAppClientConfig` | `aws_cognito_user_pool_client` リソース |
| LC-AUTH-17 | `ApiGatewayStageThrottlingConfig` | API Gateway Stage の throttling |
| その他 | API Gateway Cognito Authorizer | `aws_apigatewayv2_authorizer` or `aws_api_gateway_authorizer` |

### 1.2 既知の前提（Inception / 既決定）

- **Region**: `ap-northeast-1` のみ (NFR-COMP / Q-15)
- **IaC**: Terraform (要件 Q12=C, A-NFR-MAINT-01)
- **Module 規約**: terraform-module-design / terraform-coding-rule / terraform-test プラグイン規約準拠
- **AWS 命名規約**: aws-naming-convention プラグイン規約準拠
- **Lambda Web Adapter**: API Lambda は Go + Gin + LWA（コンテナイメージ）
- **API Gateway**: REST 形式（`requirements.md` と Application Design）
- **Frontend ホスティング**: AWS Amplify Hosting (PR #10)

### 1.3 Infrastructure Design で扱うこと / 扱わないこと

| 扱う | 扱わない |
|---|---|
| Cognito User Pool / App Client の Terraform 構成 | Cognito ユーザの初期投入データ・運用手順 |
| Pre Sign-up Lambda のリソース定義 + zip パッケージング | Lambda 関数本体の Node.js コード（→ Code Generation） |
| API Gateway の Authorizer / Stage Throttling | 業務 API の handler 実装（→ 他 Unit / Code Generation） |
| Lambda 用 IAM Role / Policy（Auth 関連分のみ） | API Lambda 全体の IAM（他 Unit と共有、shared モジュールで定義） |
| Terraform モジュール構成 / 配置（auth 用 module） | Terraform state バックエンド構成（プロジェクト全体方針、別途確定） |
| 環境分離（dev / prd）の方針 | 本 MVP では dev のみ実構築、prd は将来の枠 |
| Tagging 戦略（Project / Env / Owner / Unit） | Cost Allocation Tag の有効化（運用フェーズ） |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1問ずつ）
- [x] 回答の曖昧さを点検し、必要なら追加質問を挟む（Q-I2 / Q-I6 / Q-I10 で重要な仕様確認・スコープ拡大あり）
- [x] `aidlc-docs/construction/auth/infrastructure-design/infrastructure-design.md` を作成
- [x] `aidlc-docs/construction/auth/infrastructure-design/deployment-architecture.md` を作成
- [x] 横串リソース判定: Q-I10=A4 で API Gateway 本体・API Lambda 本体を Unit A PR で先行構築する方針確定。`shared-infrastructure.md` は本 PR では作成せず、Code Generation 完了後の整合性レビューで再判定
- [ ] aidlc-state.md と audit.md を更新（承認後）
- [ ] 完了メッセージを提示し、承認ゲートに進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-I1: Terraform モジュール分割粒度

Auth 関連のリソースをどう分割するか。terraform-module-design 規約に従いつつ。

| 案 | 分割 |
|---|---|
| A | **単一モジュール `infra/modules/auth/`** に Cognito / Pre Sign-up Lambda / API Gateway Authorizer をまとめる |
| B | **2 モジュール**: `infra/modules/cognito/`（User Pool + App Client + Pre Sign-up Lambda）と `infra/modules/api_gateway_authorizer/`（Authorizer のみ）に分離 |
| C | **3 モジュール**: `infra/modules/cognito/`, `infra/modules/lambda_pre_signup/`, `infra/modules/api_gateway/`（Authorizer 含む） |

unit-of-work.md §4.1 の `infra/modules/cognito/` を 1 つの module として例示しているので B / C 寄り。

[Answer]: **A**（単一モジュール `infra/modules/auth/`）。Cognito User Pool + App Client + Pre Sign-up Lambda + API Gateway Authorizer を 1 module にまとめる。Unit A 単位で完結、Auth 機能の追加変更が 1 module に閉じる。unit-of-work.md §4.1 の `cognito/` 例示よりも Unit A 全体を扱う Auth ファサードとして整理。

### Q-I2: API Gateway の HTTP API vs REST API

NFR Design / Application Design では「API Gateway REST」と前提しているが、Cognito Authorizer の組合せやコスト面で再検討の余地がある。

| 案 | 種類 | 特徴 |
|---|---|---|
| A | **REST API**（既存方針通り） | Cognito Authorizer ネイティブサポート、Stage Throttling 設定可能、料金は HTTP API より高い |
| B | **HTTP API** に切替 | コスト約 70% 安、レイテンシ低い、Cognito Authorizer もサポート、ただし Stage 単位 Throttling は Route Throttling として表現が違う |

requirements.md の §5.2 では「API Gateway」とのみ。

[Answer]: **B**（HTTP API に切替）。LWA は AWS 公式 README で API Gateway REST / HTTP API / Lambda Function URLs / ALB を全てサポート明記。HTTP API はコスト約 70% 安・レイテンシ低・JWT Authorizer ネイティブで Cognito も対応。Stage Throttling は Route Throttling として表現が変わるが、A-NFR-SEC-04 (100 req/s, Burst 200) の数値目標は満たせる。Application Design / NFR Design 内の「REST」記述は本ステージ完了時に整合修正する旨を備考に記載。なおユーザから「lwaはどっちかじゃないとダメじゃないっけ？」「公式ドキュメントにも書いてる？」の確認 2 回あり、本 AI 応答で LWA サポート Event Source 表 + 公式 README 引用を提示済み。

### Q-I3: Pre Sign-up Lambda のパッケージング

LC-AUTH-07 (Node.js) の zip 化方法。

| 案 | パターン |
|---|---|
| A | **Terraform `archive_file` data source** で生 JS をリポジトリ内から zip 化 → `aws_lambda_function.filename` に渡す | 依存ゼロ、CI シンプル、5 行コードに最適 |
| B | **CDK / SAM build** | Terraform に乗らない |
| C | **コンテナイメージ Lambda** | 5 行 JS には過剰 |

[Answer]: **A**（Terraform `archive_file` data source）。`infra/lambdas/pre-signup/index.js` を `archive_file` で zip 化 → `aws_lambda_function.filename` に渡す。依存ゼロ、CI シンプル、Q-D7 Node.js 選択と整合、5 行コードに最適。

### Q-I4: Cognito User Pool 名 / App Client 名 / その他リソース命名

aws-naming-convention プラグインに従うが、具体プレフィックス未確定。

| 案 | パターン例 |
|---|---|
| A | **`goro2pay-{env}-{resource}`** 形式（`goro2pay-dev-userpool`, `goro2pay-dev-appclient-web`, `goro2pay-dev-presignup-fn`） |
| B | **`gp-{env}-{resource}`** 形式（短縮、文字数節約） |
| C | aws-naming-convention プラグインの推奨をそのまま採用（プラグイン参照） |

[Answer]: **B**（`gp-{env}-{resource}`）。例: `gp-dev-userpool` / `gp-dev-appclient-web` / `gp-dev-presignup-fn` / `gp-dev-auth-authorizer`。短縮で文字数節約、AWS リソース名長制約に余裕、aws-naming-convention プラグインの命名規約レイヤと組合せ可能。

### Q-I5: 環境分離戦略

`infra/envs/dev/` と `infra/envs/prd/` の使い分け。

| 案 | 内容 |
|---|---|
| A | **dev のみ実構築、prd ディレクトリは空のまま枠だけ作成** | MVP 範囲（本 MVP は本番運用しない、Q14） |
| B | **dev と prd 両方ディレクトリ構築、prd は同じ module を呼ぶが variable 値を分離** | 構造を整えるが prd 実体は作らない |
| C | **環境分離せず、root で 1 つだけ作る** | 最小構成だが将来の本番化で全面書換 |

[Answer]: **A**（dev のみ実構築、prd は空ディレクトリだけ作成）。`infra/envs/dev/` に Auth module 呼出を実装、`infra/envs/prd/` は README placeholder のみ作成。MVP 範囲（Q14: 本番運用しない）と整合、将来ビジョンを可視化、しばらくは dev のみオペレーション。

### Q-I6: Terraform State バックエンド

本 Plan のスコープ外という整理だが、auth モジュールの output が他 module から参照されるため最低限の方針確認。

| 案 | 内容 |
|---|---|
| A | **既にプロジェクト共通の S3 + DynamoDB バックエンドが定義済みと仮定**（auth 側で気にしない、`backend` ブロックは root に既存とする） |
| B | **本 Unit A の PR で初期構築**（`infra/envs/dev/backend.tf` を auth PR で作る） |
| C | **本 Unit A の PR では言及せず、Code Generation で扱う**（実構築タイミングを遅らせる） |

[Answer]: **B (改訂版)**（Unit A PR で初期構築、ただし **DynamoDB 不要**）。Terraform 公式ドキュメントによれば DynamoDB Lock は deprecated、S3 ネイティブ `use_lockfile = true` が最新推奨。`infra/envs/dev/backend.tf` で S3 backend + `use_lockfile = true` を設定、bucket は手動 or bootstrap スクリプトで先行作成。なおユーザから「BだがDynamoDBは不要」の指摘あり、本 AI 応答で Terraform 公式ドキュメント引用 + 改訂版仕様チェックを提示済み。

### Q-I7: Cognito User Pool の Email 配信

A-NFR-I18N-01 で Q-A1 (auto-confirm) なので **メール送信は基本発生しない** が、念のためメール設定をどうするか。

| 案 | 内容 |
|---|---|
| A | **Cognito デフォルト**（Cognito 自身の SES 経由、from `no-reply@verificationemail.com`、月 50 通制限） | 本 MVP は送信を想定していないので問題なし |
| B | **SES 統合**（独自ドメインから送信） | 設定重い、本 MVP には不要 |

[Answer]: **A**（Cognito デフォルト）。auto-confirm 採用 (Q-A1) によりメール送信ルートが基本なし、Terraform で `email_configuration` ブロックを省略しデフォルト設定（`COGNITO_DEFAULT`）で運用。本番化時に SES 統合を再評価する旨を備考に明記。

### Q-I8: API Gateway Authorizer の TTL

Cognito Authorizer の `authorizer_result_ttl_in_seconds` 設定。

| 案 | 値 | 特徴 |
|---|---|---|
| A | **0 秒（キャッシュ無効）** | 毎リクエストで Cognito 検証、IdToken 取り消し（Logout）を即時反映 |
| B | **300 秒（5 分、Cognito Authorizer デフォルト）** | キャッシュで Cognito 呼出削減、Logout からの最大反映遅延 5 分 |
| C | **60 秒（中間）** | バランス、Logout の体感遅延を 1 分以内に |

A-NFR-SEC-03 で Token Validity 8h なので、キャッシュ TTL は短い方がセキュア。Logout の即時反映が UX 的にも望ましい。

[Answer]: **C**（60 秒）。`authorizer_result_ttl_in_seconds = 60`。Cognito 呼出を中庸に抑えつつ、Logout (GlobalSignOut) の反映遅延を最大 1 分以内に。Q-I2=B の HTTP API でも JWT Authorizer に同等の TTL 設定が可能。

### Q-I9: Pre Sign-up Lambda の IAM Role 権限範囲

最小権限の原則。

| 案 | 権限 |
|---|---|
| A | **CloudWatch Logs 書込のみ**（`logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents`） | 本 Trigger は AWS API を呼ばないため十分 |
| B | **A + AWS X-Ray 書込権限** | A-NFR-OBS-03 で X-Ray 不採用なので不要 |
| C | **AWSLambdaBasicExecutionRole 管理ポリシーをアタッチ**（簡便） | 過剰権限ではないが、最小権限明示性に劣る |

[Answer]: **A**（CloudWatch Logs 書込のみをインラインポリシー）。`logs:CreateLogGroup` / `logs:CreateLogStream` / `logs:PutLogEvents` のみ、Resource は対象 Log Group ARN に絞る。X-Ray 不採用 (A-NFR-OBS-03) なので追加権限なし、terraform-coding-rule の IAM ポリシー記述規約に準拠。

### Q-I10: API Gateway → Lambda 連携方式（Auth に関わる部分のみ）

Auth Unit が触るのは `POST /api/auth/logout` だが、API Gateway 全体の Lambda 連携を Unit A の PR で先に決めるか。

| 案 | 内容 |
|---|---|
| A | **Auth は API Gateway Authorizer のみ定義、Lambda 連携の本体は Unit B 以降の PR で構築**（Unit A の責務最小化） |
| B | **Unit A で API Lambda の最低限の枠（aws_lambda_function + aws_apigatewayv2_integration の Logout エンドポイント分）も定義** | Auth 単体で完結するが他 Unit に影響 |

unit-of-work.md §4.1 では `infra/modules/lambda_api/` が Unit 横串。Unit A は Authorizer のみ寄りが妥当。

[Answer]: **A4**（Cognito + Pre Sign-up + API Gateway 本体 + Authorizer + Logout route/integration + API Lambda 本体 Gin Hello World）。Unit A PR で Auth 機能の動作確認に必要なすべてのインフラを構築する。他 Unit は route + integration を追加するだけ。unit-of-work.md §4.1 の `lambda_api/` 「横串」表記からは逸脱するが、横串 PR の所在が計画上空白で実装漏れリスクが高いため、Unit A で先行構築する方針を deployment-architecture.md に明記。`infra/modules/lambda_api/` モジュールは Unit A PR で作成し他 Unit からも参照する形とする。なおユーザから複数回の確認 (LWA 制約 / Authorizer = Lambda? / 横串 PR タスク管理 / Logout integration / タイムライン) あり、本 AI 応答で各論点を図解 + 公式ドキュメント引用 + リソース構造図で解説済み。

### Q-I11: Tagging 戦略

CloudWatch Cost Allocation や運用識別のためのタグ。

| 案 | タグセット |
|---|---|
| A | **`Project=goro2pay`, `Env=dev`, `Unit=auth`, `ManagedBy=terraform`** | 4 種類、十分な識別性 |
| B | **A + `Owner` / `CostCenter` 等を追加** | 個人プロジェクト相当には過剰 |
| C | **タグなし** | MVP として最小、コスト分析できなくなる |

[Answer]: **A**（Project / Env / Unit / ManagedBy の 4 種類）。`Project=goro2pay` / `Env=dev` / `Unit=auth` / `ManagedBy=terraform`。Terraform `default_tags` provider 設定で全リソースに自動付与、各リソース個別の tags でも override 可能。

### Q-I12: CloudWatch Log Group の保持期間

A-NFR-OBS-01 / NFR-OBS-02 整合。

| 案 | 保持日数 | 月額コスト目安 (5 user, 軽負荷) |
|---|---|---|
| A | **7 日**（最短設定、デモ用途） | $0.05 程度 |
| B | **30 日** | $0.20 程度 |
| C | **無期限**（CloudWatch Logs デフォルト） | 増え続ける |

ハッカソン審査用なので短期で十分。

[Answer]: **A**（7 日）。`retention_in_days = 7`。デモ用途では十分、月額コスト最小化（$0.05 程度）。本番化時に 30 日以上への変更を再評価する旨を明記。

### Q-I13: Cognito ユーザ削除運用

A-NFR-COMP-02 「ユーザ削除手順は本番化時に対応」だが、Terraform リソースとして `deletion_protection` を設定するか。

| 案 | 内容 |
|---|---|
| A | **`deletion_protection = "INACTIVE"`** | terraform destroy で User Pool 削除可能、検証用途で楽 |
| B | **`deletion_protection = "ACTIVE"`** | terraform destroy 不可、手動解除が必要、誤削除防止 |

MVP / dev 環境想定なら A、本番化時に B に切替。

[Answer]: **A**（INACTIVE）。`deletion_protection = "INACTIVE"`。MVP / dev 環境では検証サイクルの柔軟性を優先、`terraform destroy` で User Pool 削除可能。本番化時に B (`ACTIVE`) への切替を A-NFR-COMP-02 と合わせて明記。

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `infrastructure-design.md` | Terraform リソース定義: User Pool / App Client / Pre Sign-up Lambda / Authorizer / IAM Role / Log Group。論理パラメータ（LC-15/16/17）→ HCL マッピング、変数・出力定義、モジュール分割 |
| `deployment-architecture.md` | デプロイメントアーキテクチャ図（ASCII）: AWS リソース配置、Cognito ↔ Pre Sign-up Lambda ↔ API Gateway ↔ API Lambda ↔ Frontend のリクエストフロー、Terraform モジュール依存関係、環境分離戦略、デプロイ手順 |

---

### Q-I14 (追加): AWS Amplify Hosting のデプロイソース連携

横串インフラを Unit A 担当方針に伴い追加。Frontend のホスティングと CD 経路。

| 案 | 内容 |
|---|---|
| A | **GitHub 連携 (auto deploy from main/develop)**: Amplify Hosting の標準パターン、CodeStar Connections / GitHub App 経由 |
| B | **手動デプロイ (Manual deploy)**: zip / Build Output を手動アップロード、Terraform `aws_amplify_app` のみ |
| C | **Webhook トリガだけ設定**: GitHub Actions 等から Amplify Webhook を叩く |

[Answer]: **A**（GitHub 連携 auto deploy）。develop ブランチへの push で自動デプロイ、CodeStar Connection 経由。Personal Access Token は使わず GitHub App 連携 / CodeStar Connection を採用。

### Q-I15 (追加): API Lambda の CD 戦略

API Lambda の image 更新フロー。Pre Sign-up Lambda は archive_file + terraform apply（暗黙的に手動）のまま据え置き。

| 案 | 内容 |
|---|---|
| A | **手動デプロイ**: ECR push + terraform apply、Frontend の自動デプロイと不整合 |
| B | **GitHub Actions で ECR build/push + terraform apply**: フル自動、IAM OIDC 設定必要 |
| C | **ECR push のみ自動、terraform 手動**: Lambda image_uri 更新は update_function_code |
| D | **AWS CodePipeline + CodeBuild**: AWS ネイティブ CD、buildspec.yml + IAM Role |

[Answer]: **D**（CodePipeline + CodeBuild）。推奨構成: Source (GitHub via CodeStar Connection) → Build (CodeBuild Docker build → ECR push + `aws lambda update-function-code --image-uri`)。Terraform 側 `aws_lambda_function.image_uri` は `lifecycle.ignore_changes = ["image_uri"]` で CD を妨げない。Frontend (Q-I14=A) との CD 整合性を確保、AWS ネイティブで Terraform 内完結。Pre Sign-up Lambda は archive_file 方式のまま手動 (terraform apply) 運用。

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

- DynamoDB テーブル定義（Unit B/C/D 所有）→ それぞれの Unit Infrastructure Design
- Pre Sign-up Lambda の Node.js コード本体 → Code Generation
- API Lambda の Go コード本体 + Dockerfile → Code Generation
- Backend Go middleware / Frontend hook 実装 → Code Generation
- buildspec.yml の具体記述 → Code Generation

---

## 6. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-I1 〜 Q-I13 を対話形式で順にヒアリング開始
