# Order Unit (Unit C) — Infrastructure Design Plan

**Document Version**: 0.1 (Draft, awaiting user approval)
**Created**: 2026-05-24
**Unit**: C (`order` / 代行手配コア)
**Construction Depth**: Comprehensive
**Stage**: Infrastructure Design (Construction Phase)
**Prerequisite**: NFR Design 承認済み (PR #79 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit C（`order` / 代行手配コア）の **Infrastructure Design ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。NFR Design で定義した論理コンポーネント（LC-ORDER-01〜34）を **実 AWS リソース（Terraform）** に展開する。Comprehensive 深度として Bedrock IAM / CloudWatch アラーム 3 種 / AWS Budgets 等の Unit C 固有論点を含める。

### 1.1 確定済み前提（質問不要）

| 項目 | 値 | 出典 |
|---|---|---|
| IaC | Terraform | 要件 Q-12=C |
| Region | `ap-northeast-1` | NFR-COMP / Q-15 |
| 命名規則 | `gp-{env}-{resource}` | Unit A Q-I4 |
| タグ | `Project=goro2pay / Env=dev / Unit=order / ManagedBy=terraform` | Unit A Q-I11 |
| モジュール規約 | terraform-module-design / terraform-coding-rule / terraform-test 準拠 | チームベースライン |
| API Gateway | Unit A が構築済み (`aws_apigatewayv2_api.main` + JWT Authorizer + Stage Throttling 100 req/s) | Unit A Infrastructure Design |
| API Lambda | Unit A が構築済み (`gp-{env}-api-fn`、Go + Gin + LWA、ECR + CodePipeline) | Unit A Infrastructure Design |
| API Lambda IAM Role | Unit A が作成済み（Unit C で DynamoDB + Bedrock 権限を追加） | Unit A Infrastructure Design |
| API Lambda Memory | **256MB** / arm64 / タイムアウト 10s | NFRC-C18、Unit C Q-N9=B |
| Bedrock モデル | `apac.anthropic.claude-3-5-haiku-20241022-v1:0`（ap-northeast-1 inference profile） | NFRC-C20、Unit C Q-N8=B |
| 構造化ログ | `log/slog` + `ContextAwareSlogHandler`（Unit A LC-AUTH-05 再利用） | NFRC-C12、Unit B 統一 |
| OrderHistory PK/SK | unit-interfaces.md §3.2 で凍結済み（PK=`USER#{userID}` / SK=`ORDER#{orderedAt}#{orderID}`） | unit-interfaces.md |
| TTL 属性名 | `expiresAt`（90 日） | NFRC-C12 / 凍結契約整合修正済み |
| terraform module 構成 | Unit A の機能別 5 module に Unit C 用 1 module を追加 | Unit A Q-I1 系列 |

### 1.2 Unit C Infrastructure Design のスコープ

1. **DynamoDB テーブル** 1 本（`OrderHistory`、TTL 90 日、GSI なし）
2. **Bedrock 権限**（API Lambda Role に Claude 3.5 Haiku モデル ARN 限定で追加）
3. **API Gateway ルート** 2 本（`POST /api/orders` / `GET /api/orders`）— 既存 API Gateway に追加
4. **IAM 権限追加**（API Lambda Role に DynamoDB OrderHistory + Bedrock InvokeModel/Converse + WalletService 利用のための Unit B テーブル参照権限）
5. **CloudWatch Logs メトリクスフィルタ** 3 種（NFRC-C13）
6. **CloudWatch Alarms** 3 種（p95 違反 / Bedrock リトライ発動 / フォールバック発動）
7. **SNS Topic**（アラーム通知メール）
8. **AWS Budgets**（Bedrock 月 $5 アラート、NFRC-C20）
9. **Terraform モジュール** 構成（`infra/modules/order/`）
10. **CloudWatch Log Group**（API Lambda 用は Unit A 既存、Unit C 専用は不要）

### 1.3 Unit C Infrastructure Design スコープ外（他 PR / 後続 Unit が担当）

- API Lambda 本体（Unit A が ECR + CodePipeline でデプロイ済み）
- API Gateway 本体 / JWT Authorizer / Stage / Throttling（Unit A 構築済み）
- Cognito User Pool / App Client（Unit A 構築済み）
- Wallet / IdempotencyKeys / BudgetSettings DynamoDB テーブル（Unit B が構築済み、Unit C は参照のみ）
- BedrockAdapter / DeliveryAdapter / OrderHistoryRepository の Go 実装 → **Code Generation**
- CloudWatch ダッシュボード（NFRC-C14 で MVP 不実装、NFR-OBS-02 整合）
- X-Ray / カスタムメトリクス（NFRC-C14 で MVP 不実装）

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（ユーザ要請により全 13 問を推奨案で一括回答、各回答に選定理由を記載）
- [x] 回答の曖昧さ・矛盾を点検し、必要なら追加サブ質問を挟む（10 観点で矛盾チェック実施、整合確認、Q-I1/Q-I11/Q-I12 を terraform-module-design 準拠に見直し）
- [x] `aidlc-docs/construction/order/infrastructure-design/infrastructure-design.md` を作成（Comprehensive 深度、機能別 3 module 新規 + 既存 2 module 追記）
- [x] `aidlc-docs/construction/order/infrastructure-design/deployment-architecture.md` を作成（全体構成図、リクエストフロー 6 種、観測性データフロー、デプロイ手順、Code Generation 引き継ぎ）
- [x] aidlc-state.md / audit.md / Plan checkboxes を更新
- [ ] 完了メッセージを提示し、承認ゲート（2-option）に進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-I1: Terraform モジュール構成

Unit C の Terraform リソース（DynamoDB OrderHistory + Bedrock IAM + CloudWatch Alarms + AWS Budgets）をどのモジュール構成で管理するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **新規 `infra/modules/order/` モジュール**（Unit ごとに独立） | Unit B と同方針。`envs/dev/main.tf` で `module "order"` を追加呼出。Unit 分離が綺麗 |
| B | **Unit A の `infra/modules/lambda_api/` に追加** | API Lambda 関連リソースを 1 モジュールに集約。Unit 境界がぼやける |
| C | **A + Bedrock 関連のみ別モジュール `infra/modules/bedrock/` に切り出し**（Unit C / D 共有を見据えて） | Unit D 追加時に Bedrock IAM を再利用しやすい。モジュール数増加 |

[Answer]: **機能別 3 module 新規（`order_history` / `bedrock` / `observability`）+ 既存 `api_gateway` / `lambda_api` への追記**（terraform-module-design 準拠の見直し版）

**選定理由**:
- terraform-module-design は AWS リソースカテゴリ別の機能分割を推奨（例: `alb` / `cloudfront` / `databases` / `ecs` / `lambda` / `network` / `security`）
- Unit A の機能別 5 module（`codestar_connection` / `cognito` / `api_gateway` / `lambda_api` / `amplify`）と粒度を揃える
- 当初案の Unit 別 `modules/order/` 1 module 集約は Unit A の機能別パターンとズレる + 横串リソース（SNS / Budgets）が Unit C に閉じ込められ Unit D/E で再利用しにくい問題があった
- 見直し後の構成:
  - `infra/modules/order_history/` 新規: DynamoDB OrderHistory + DynamoDB アクセス用 IAM Policy（Unit C 専用）
  - `infra/modules/bedrock/` 新規: Bedrock IAM Policy（Foundation Model + Inference Profile ARN 限定、Unit C / D 共有候補）
  - `infra/modules/observability/` 新規: CloudWatch Logs metric filter 3 + Alarms 3 + SNS Topic `gp-{env}-alarms` + Budgets（横串、Unit D/E 再利用可）
  - 既存 `modules/api_gateway/routes.tf` に Unit C ルート 2 本を追記（Q-I11 連動: A → B に変更）
  - 既存 `modules/lambda_api/iam.tf` で Unit C IAM Policy を attach（Q-I12 維持: data source 不要、`variables.tf` に `additional_policy_arns` 追加で受け取る）
- **envs/dev/ の踏襲**:
  - `local.env` / `local.region` を Unit C 3 module 呼出でも踏襲
  - `locals.tf` に `alarm_email = "alerts@example.com"` を追加（Q-I6 連動）
  - `main.tf` に Unit C 3 module を追記、`module.lambda_api` に `additional_policy_arns = [module.order_history.dynamodb_policy_arn, module.bedrock.bedrock_policy_arn]` を渡す
  - `outputs.tf` に `dynamodb_table_name` / `sns_topic_arn` を追加
  - `provider` の `default_tags` は Project/Env/ManagedBy のみ、`Unit` タグは各 module で `merge()` 個別付与（Unit A 規約踏襲）
- 各 module 配下のファイル分割は terraform-module-design 準拠（`main.tf` / `data.tf` / `locals.tf` / `variables.tf` / `outputs.tf`）

### Q-I2: DynamoDB `OrderHistory` のキャパシティモード

NFRC-C11 でオンデマンドスケーリング前提。`OrderHistory` テーブルのキャパシティモードを Unit B（Q-N4=C プロビジョンド 1 RCU/1 WCU）と同方針にするか、別判断にするか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **プロビジョンド 1 RCU / 1 WCU**（Unit B と統一） | デモ規模最低コスト、無料枠内、Unit B と一貫性 |
| B | **オンデマンド** | リクエスト量変動への自動追従、デモ規模では Pay-per-request の方が安い可能性 |
| C | **プロビジョンド 1 RCU / 2 WCU**（Insert 多め想定） | `PlaceOrder` ごとに Insert 1 回、Read は MainScreen 表示時のみ |

[Answer]: **A（プロビジョンド 1 RCU / 1 WCU、Unit B と統一）**

**選定理由**:
- Unit B Q-N4=C と統一、無料枠内で運用可能（コスト最低）
- NFRC-C11 同時利用者数: 数人〜数十人で 1 RCU/1 WCU でバースト対応十分
- C 案の 1 RCU/2 WCU は `PlaceOrder` 高頻度想定だが、デモ規模では月 900 リクエスト程度で 1 WCU でバースト容量を超えない
- B 案のオンデマンドは小規模では Pay-per-request の方が割高になるケースが多く、固定コスト予測の観点で A が優位
- `OrderHistory` への Read は MainScreen 表示時 + フォールバック分岐判定（履歴 5 件取得）のみで、1 RCU で十分
- バースト容量前提（同 Unit B 採用）を Infrastructure Design ドキュメントに注記する

### Q-I3: Bedrock IAM ポリシー設計

NFRC-C20 で Claude 3.5 Haiku 確定。API Lambda Role に Bedrock 権限を追加する範囲。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`bedrock:InvokeModel` + `bedrock:Converse` を `apac.anthropic.claude-3-5-haiku-*` モデル ARN 限定で許可** | 最小権限原則、PoLP 厳守 |
| B | **A + `bedrock:InvokeModelWithResponseStream` も許可**（将来のストリーミング対応の余地） | 拡張余地、ただし MVP では未使用 |
| C | **`bedrock:*` を全モデル ARN で許可**（開発容易性） | 最低限の制約、本番化時に見直し前提 |

[Answer]: **A（`bedrock:InvokeModel` + `bedrock:Converse` を Claude 3.5 Haiku モデル ARN 限定）**

**選定理由**:
- 最小権限原則 (Principle of Least Privilege) を厳守、AWS Well-Architected Framework Security Pillar 準拠
- NFRC-C24（PII 漏洩リスク予防）の精神と整合、不要な Bedrock 機能へのアクセスを構造的に排除
- B 案の `InvokeModelWithResponseStream` は MVP では未使用で、必要になった時点で追加すればよい（YAGNI）
- C 案の `bedrock:*` はセキュリティ的に NG、本番化時の見直し前提だと忘れるリスクあり
- 許可するアクション: `bedrock:InvokeModel` / `bedrock:Converse`（NFRC-C20 で Converse API 確定）
- リソース ARN: `arn:aws:bedrock:ap-northeast-1::foundation-model/anthropic.claude-3-5-haiku-*` + Inference Profile ARN（Q-I10 連動）
- terraform-test で IAM Policy のリソース ARN 制限を自動検証（Q-I13=B/C 連動）

### Q-I4: CloudWatch Logs メトリクスフィルタ + Alarms 実装方針

NFRC-C13 で 3 種のアラーム確定（p95 違反 / Bedrock リトライ発動 / フォールバック発動）。Terraform でどう実装するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **Terraform `aws_cloudwatch_log_metric_filter` + `aws_cloudwatch_metric_alarm` リソースを Unit C モジュール内に配置** | Unit C で完結、シンプル |
| B | **Alarms 関連を別 module `infra/modules/alarms/` に切り出し**（全 Unit のアラームを一元管理） | 横串化、ただし MVP では Unit A/B にアラームなしのため Unit C 単独 |
| C | **A + アラーム閾値を `variables.tf` で外部化**（環境別に変更可能） | 柔軟性、dev/stg/prd で異なる閾値設定可能 |

[Answer]: **C（A + アラーム閾値を `variables.tf` で外部化）**

**選定理由**:
- Comprehensive 深度の品質投資として、dev/stg/prd で異なる閾値（NFRC-C13 のしきい値: p95 3000ms / リトライ 5 回 / フォールバック 3 回）を設定可能にしておく
- prd では本番負荷を考慮してしきい値を緩める可能性あり（例: リトライ 10 回 / 5 分）、その時に変数化していると Terraform コード変更不要
- `terraform.tfvars.example` にデフォルト値を記載し、環境別 `terraform.tfvars` で上書き
- B 案の独立 `infra/modules/alarms/` モジュール化は、Unit A/B にアラームなしのため Unit C 単独でモジュールを作る価値が低い。Unit D/E が同様のアラームを必要とした時点で横串リファクタリング可能
- A 案だけだと dev で発火しすぎる閾値を勝手に変更するリスク、変数化で意図的な変更を強制できる

### Q-I5: SNS Topic（アラーム通知）の構成

NFRC-C13 のアラームは SNS Topic 経由でメール通知。SNS Topic の構成方針。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **Unit C 専用 SNS Topic 1 つ**（3 アラーム共通） | シンプル、通知先メール 1 つ |
| B | **アラームごとに SNS Topic 3 つ**（Critical / Warning / Info で分離） | 通知先を分けられる、複雑化 |
| C | **横串 SNS Topic 1 つ（`gp-{env}-alarms`）を新設、全 Unit のアラームをここに集約** | 全 Unit 共通、Unit A/B/D/E で再利用可能 |

[Answer]: **C（横串 SNS Topic 1 つ `gp-{env}-alarms` を新設）**

**選定理由**:
- 全 Unit のアラームを一元管理、運用者は 1 つのメールアドレスで全アラート受信可能
- Unit A/B には現状アラームがないが、将来追加された際に Unit C で作った SNS Topic を再利用できる（Unit C モジュール出力で `sns_topic_arn` を export）
- 配置は `infra/modules/order/` に置くと Unit C 削除時に他 Unit のアラーム購読が壊れるリスクがあるため、横串の位置（`envs/dev/main.tf` 直接定義 or 新規 `infra/modules/alarms/` の検討は将来）
- 暫定: Unit C モジュール内で SNS Topic を定義し、Unit D/E 追加時に横串モジュールに昇格する設計圧力を残す（infrastructure-design.md §10 整合性メモに明記）
- B 案のアラームレベル別 SNS Topic は通知先メール 1 つ運用のデモ規模では過剰
- A 案の Unit C 専用 Topic は Unit 越境時の再利用性が低い

### Q-I6: SNS Topic のサブスクリプション

NFRC-C13 のアラーム通知先メールアドレス。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **Terraform 変数 `var.alarm_email` で受け取り、`aws_sns_topic_subscription` で email 購読** | 標準パターン、`terraform.tfvars.example` に記載 |
| B | **Slack Webhook 連携**（Lambda + SNS で通知変換） | リッチ通知、実装コスト高 |
| C | **手動購読**（Terraform で SNS Topic のみ作成、購読は AWS Console で手動設定） | デモ用シンプル、IaC 化されない |

[Answer]: **A（Terraform 変数 `var.alarm_email` + `aws_sns_topic_subscription` で email 購読）**

**選定理由**:
- 標準的な AWS Terraform パターン、IaC 化により再現性確保
- `terraform.tfvars.example` に変数記載例を提供（`alarm_email = "alerts@example.com"`）
- B 案の Slack Webhook 連携は Lambda + SNS の追加実装が必要で、MVP スコープ外
- C 案の手動購読は Terraform で IaC 化しないため、環境再構築時に購読忘れリスク
- `aws_sns_topic_subscription.confirmation_timeout` で SNS 購読確認のタイムアウトを設定可能、初回 apply 時にユーザがメールから確認する運用
- Q-I5=C の横串 SNS Topic と組み合わせ、`gp-{env}-alarms` トピックに 1 つのメール購読を attach
- 注意: SNS Subscription は AWS から確認メールが送信され、ユーザがクリックして確定する必要があり、`pending_confirmation` 状態を許容する Terraform 設定とする

### Q-I7: AWS Budgets（Bedrock コスト監視）の閾値と通知

NFRC-C20 で月次予算 $10、Q-D 段階で「AWS Budgets で Bedrock 月 $5 超過アラート」確定。具体閾値と通知方針。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **80% (= $4 / 月) で警告 + 100% (= $5 / 月) で警報、両方 SNS Topic に通知** | 段階的、想定外コスト早期検知 |
| B | **100% ($5 / 月) のみ通知** | シンプル、警告レベル不要 |
| C | **A + 予測ベース予算（forecasted spend > $5 で 1 週間以内）も追加** | 予測アラート、本番運用向け |

[Answer]: **A（80% で警告 + 100% で警報、両方 SNS Topic に通知）**

**選定理由**:
- 段階的アラート（80% で警告 → 100% で警報）により、想定外コストを早期検知できる
- $4 警告の段階で原因調査（Bedrock リトライ多発 / フォールバック発動率増加）に着手可能、$5 到達前に対応余地あり
- C 案の予測ベース予算は本番運用向けで、デモ規模（月 900 リクエスト = $1.08 / 月見込み）では検知精度が低く誤報多発の可能性
- B 案の 100% のみだと予算到達後の事後対応となり、コスト超過リスク高
- AWS Budgets の `aws_budgets_budget` リソースで `notification` ブロックを 2 つ定義、いずれも Q-I5=C の横串 SNS Topic `gp-{env}-alarms` に通知
- 月次予算 $5 はNFRC-C20 の月次予算上限 $10 の半分、本番想定の半分超過で警報発動する設計

### Q-I8: DynamoDB `OrderHistory` の Point-In-Time Recovery (PITR)

| 案 | パターン | 特徴 |
|---|---|---|
| A | **無効**（dev 環境 / ハッカソン、Unit B Q-I6=A と統一） | 追加コストなし、リカバリ不可 |
| B | **有効** | 最大 35 日分のリカバリが可能、月数十円のコスト増 |

[Answer]: **A（PITR 無効、Unit B Q-I6=A と統一）**

**選定理由**:
- Unit B Q-I6=A と統一、デモ環境でリカバリ要件がない
- TTL 90 日（NFRC-C12 / 凍結契約整合修正済み）で自動削除されるデータのため、PITR で 35 日のリカバリ可能でも実質意味がない（注文履歴は誤削除リスクが極低）
- 月数十円のコスト増は誤差レベルだが、NFR-COMP-01〜03（仮想ウォレット、PCI DSS 直接対象外）の運用要件で PITR を必須としていない
- 本番化時に NFR-REL-04（手動リカバリ手順文書化）と同様の方針で再検討する設計圧力（infrastructure-design.md §10 整合性メモに明記）
- B 案は本番運用で監査要件・SLA 要件が出てきた段階で有効化検討

### Q-I9: DynamoDB `OrderHistory` の暗号化

| 案 | パターン | 特徴 |
|---|---|---|
| A | **AWS マネージド（`AES256`、デフォルト）**（Unit B Q-I4=A 統一） | 追加コストゼロ、Unit B と統一 |
| B | **Customer Managed KMS key** | 監査要件向け、ハッカソン規模では不要 |

[Answer]: **A（AWS マネージド AES256、Unit B Q-I4=A と統一）**

**選定理由**:
- Unit B Q-I4=A と統一、追加コストゼロ
- NFR-SEC-02「データは AWS マネージドサービスのデフォルト暗号化を利用、KMS 顧客管理キーは使用しない」に整合
- NFR-COMP-01（PCI DSS 直接対象外）でハッカソン規模では CMK 不要
- B 案の Customer Managed KMS key は監査・コンプライアンス要件向けで、KMS リクエスト料金（$0.03 / 10000 requests）+ 月額料金が発生
- 本番化時に金融庁ガイドライン適合性評価（NFR-COMP-03）の段階で CMK 化検討

### Q-I10: Bedrock のモデル呼出方法（Inference Profile vs 直接呼出）

NFRC-C20 で `apac.anthropic.claude-3-5-haiku-20241022-v1:0` 確定。ap-northeast-1 では inference profile 経由が推奨。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **Inference Profile 経由**（`apac.anthropic.claude-3-5-haiku-20241022-v1:0`） | ap-northeast-1 含むアジア太平洋リージョンで自動的に最適なリージョンを選択、AWS 推奨 |
| B | **直接呼出**（`anthropic.claude-3-5-haiku-20241022-v1:0`） | 単一リージョン固定、IAM Policy のリソース ARN がシンプル |
| C | **A + Bedrock model availability の Terraform data source で動的取得** | 将来モデルバージョンアップ時の保守性向上、過剰 |

[Answer]: **A（Inference Profile 経由 `apac.anthropic.claude-3-5-haiku-20241022-v1:0`）**

**選定理由**:
- NFRC-C20 で Inference Profile 経由が確定済み（Plan §1.1 確定済み前提）
- ap-northeast-1 含むアジア太平洋リージョンで自動的に最適なリージョン（東京・ソウル・ムンバイ等）を選択、AWS 推奨パターン
- リージョン障害時の自動フェイルオーバー、可用性向上
- IAM Policy のリソース ARN: `arn:aws:bedrock:*::foundation-model/anthropic.claude-3-5-haiku-20241022-v1:0` + `arn:aws:bedrock:ap-northeast-1:{account}:inference-profile/apac.anthropic.claude-3-5-haiku-20241022-v1:0` の 2 つを許可
- B 案の直接呼出は単一リージョン固定でリージョン障害に脆弱
- C 案の data source 動的取得は MVP では過剰、モデルバージョンアップは Terraform 変数で対応可能

### Q-I11: API Gateway ルート定義の配置

`POST /api/orders` / `GET /api/orders` 2 ルートを既存 API Gateway に追加する Terraform 配置。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **Unit C モジュール (`infra/modules/order/`) 内で `aws_apigatewayv2_route` を定義し、Unit A の API Gateway ID を data source で参照** | Unit 分離が綺麗、Unit C 内で完結 |
| B | **Unit A の `infra/modules/api_gateway/` に Unit C ルートを追加** | API Gateway 関連リソース集約、Unit 境界が曖昧 |
| C | **`infra/modules/api_gateway/` を Unit ごとに routes 引数を受け取る generic な module に拡張** | 将来 Unit D/E のルート追加も容易、リファクタリング規模大 |

[Answer]: **B（既存 `infra/modules/api_gateway/routes.tf` に Unit C ルートを追記）**（Q-I1 見直しに伴い変更）

**選定理由**:
- Q-I1 で機能別 module 分割（terraform-module-design 準拠）に方針変更したため、API Gateway 関連リソースは `modules/api_gateway/` に集約する設計に統一
- Unit A の `modules/api_gateway/routes.tf` が既に全 route の集約管理を担う設計（`POST /api/auth/logout` 等を集約）。Unit C ルート 2 本（`POST /api/orders` / `GET /api/orders`）も同 routes.tf に追記
- `aws_apigatewayv2_integration` + `aws_apigatewayv2_route` の組み合わせを 2 ルート分追加、integration target は既存 `module.lambda_api.api_lambda_invoke_arn` を流用（envs/dev/main.tf 既存の依存方向を維持）
- A 案の Unit C モジュール内で route 定義する設計は当初の Unit 別 module 案で正当だったが、機能別 module 構成に揃える今回の方針では一貫性を欠く
- C 案の generic module 化は Unit C/D/E のルート構成パターンが揃ってから一度にリファクタリング可能（YAGNI）
- API Gateway リソースが 1 module に集約されることで、route の網羅的把握が容易になり運用ミスを減らせる

### Q-I12: Lambda IAM Policy の更新方式

Unit A の API Lambda Role に Unit C 用権限（DynamoDB OrderHistory + Bedrock + Wallet 参照）を追加する方式。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **Unit A の Role を data source で参照し、Unit C モジュール内で `aws_iam_role_policy_attachment` で managed policy を attach** | Unit 分離、最小変更、Unit C で完結 |
| B | **Unit A の `infra/modules/lambda_api/` 内の inline policy を Unit C 用に拡張** | Unit A モジュールに Unit C 知識が漏れる |
| C | **A + Policy を `aws_iam_policy` リソースとして独立化**（再利用可能） | 将来別 Lambda にも流用可能、過剰 |

[Answer]: **A 改定版（Policy 定義は新規 `modules/order_history/` + `modules/bedrock/` 側、attach は既存 `modules/lambda_api/` で実施、Policy ARN は envs 経由で注入）**（Q-I1 見直しに伴い改定）

**選定理由**:
- Q-I1 で機能別 module に分割したため、Policy リソース定義は責務元 module に置く方が責務分離が綺麗:
  - DynamoDB OrderHistory アクセス Policy → `modules/order_history/main.tf`
  - Bedrock InvokeModel + Converse Policy → `modules/bedrock/main.tf`
- attach は API Lambda Role を所有する `modules/lambda_api/iam.tf` で実施（attach 責任は Role 所有モジュール）
- envs/dev/main.tf で `module.lambda_api` に `additional_policy_arns = [module.order_history.dynamodb_policy_arn, module.bedrock.bedrock_policy_arn]` を渡す
- `modules/lambda_api/variables.tf` に `additional_policy_arns: list(string)` を追加（デフォルト `[]`、後方互換）
- `modules/lambda_api/iam.tf` で `for_each` を使って attach（Unit D/E 追加時も同じパターンで policy_arn を追加可能）
- B 案の Unit A モジュール内 inline policy 拡張は、Unit A `modules/lambda_api/` に Unit C 知識（DynamoDB / Bedrock）が漏れる anti-pattern
- C 案の `aws_iam_policy` 独立化は機能別 module 分割で実質的に達成される
- data source `aws_iam_role` 参照は不要（envs 経由で Policy ARN を渡す方が依存関係が明示的）
- terraform-coding-rule の IAM ポリシー記述規約（最小権限、リソース ARN 限定、jsonencode 利用）に従う

### Q-I13: terraform-test の対象範囲

Unit B Infrastructure Design では Code Generation 段階で terraform-test を書く方針。Unit C の terraform-test 対象。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **DynamoDB OrderHistory のスキーマ検証**（PK/SK/TTL 属性、暗号化、PITR） | 凍結契約 §3.2 整合確認 |
| B | **A + Bedrock IAM Policy のリソース ARN 制限検証**（`bedrock:InvokeModel` が Claude 3.5 Haiku モデル ARN 限定） | 最小権限原則の自動検証 |
| C | **A + B + CloudWatch Alarms 3 種の閾値検証**（p95 / リトライ / フォールバック） | NFRC-C13 整合確認、テストカバレッジ最大 |

[Answer]: **C（A + B + CloudWatch Alarms 3 種の閾値検証）**

**選定理由**:
- Comprehensive 深度の品質投資として、テストカバレッジを最大化
- Unit C は MVP の心臓部（unit-of-work.md §3.3）で、Infrastructure の不具合がビジネス全体に直撃するため検証範囲を広く取る
- terraform-test ファイル構成（`infra/modules/order/tests/`）:
  - `dynamodb_schema.tftest.hcl`: PK/SK/TTL 属性、暗号化（AES256）、PITR 無効、キャパシティ 1 RCU/1 WCU を検証（Q-I2/Q-I8/Q-I9 整合）
  - `bedrock_iam_least_privilege.tftest.hcl`: IAM Policy のリソース ARN が Claude 3.5 Haiku モデル ARN + Inference Profile ARN に限定されていること、`bedrock:*` でないことを検証（Q-I3/Q-I10 整合）
  - `cloudwatch_alarms.tftest.hcl`: 3 アラームの閾値（p95 3000ms / リトライ 5/5min / フォールバック 3/5min）、評価期間、SNS Topic 紐付けを検証（Q-I4/Q-I5 整合、NFRC-C13 整合）
- mock_provider 利用で AWS API 呼出なしのオフラインテスト、CI で実行可能（terraform-test 規約準拠）
- A/B 単独だと Comprehensive 深度の名に相応しくない、Unit B Standard 深度との差別化

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `aidlc-docs/construction/order/infrastructure-design/infrastructure-design.md` | Terraform リソース詳細（DynamoDB OrderHistory / Bedrock IAM / API GW ルート 2 本 / IAM 権限追加 / CloudWatch Logs メトリクスフィルタ 3 種 / CloudWatch Alarms 3 種 / SNS Topic / AWS Budgets / Terraform モジュール構成 / terraform-test 方針） |
| `aidlc-docs/construction/order/infrastructure-design/deployment-architecture.md` | Unit C のデプロイ構成図（PlaceOrder リクエストフロー、観測性データフロー、リクエスト型ごとのフロー、コスト見積、デプロイ手順、Code Generation 引き継ぎ事項） |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

| 引き継ぎ先 | 内容 |
|---|---|
| **Code Generation** | Bedrock SDK 初期化コード（P-INIT-01）、Bedrock プロンプトテンプレート最終、リトライ・タイムアウト・冪等性の Go 実装、PBT のテストコード（gopter）、`MockBedrockAdapter` 実装、Frontend `useOrder` / `useOrderHistory` / `useDisableLock` / `useToast` 実装、`GoroButton` / `OrderCompletionScreen` / `ToastHost` / `Toast` / `OrderHistoryList` 実装、`errorMappers` / `getRandomToast` 純関数実装、`apiClient.placeOrder` / `apiClient.getOrderHistory` 実装、`lib/ulid.ts` 実装 |
| **Build and Test** | dev 環境での手動 E2E 検証手順、CI ワークフロー定義、Bedrock スタブのテスト戦略実装、CloudWatch Alarms の動作確認手順 |

---

## 6. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-I1 〜 Q-I13 を対話形式で順にヒアリング開始
