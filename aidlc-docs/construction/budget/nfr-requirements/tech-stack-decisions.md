# Unit B `budget` — Tech Stack Decisions

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Stage**: NFR Requirements (Construction Phase)

---

## 0. このドキュメントの目的

Unit B `budget` の技術スタック選定を記録する。選定理由・バージョン方針・後続ステージへの引き継ぎ事項を含む。

---

## 1. バックエンド

### 1.1 ランタイム・フレームワーク

| 技術 | 選定内容 | 根拠 |
|---|---|---|
| 言語 | **Go** | Application Design Q-A=Go+Gin+LWA で確定済み。Lambda 上での軽量動作、128MB メモリで十分（Q-N9=A） |
| HTTP フレームワーク | **Gin** | Application Design Q-A 確定済み。Lambda Web Adapter 経由で Lambda 上で動作 |
| Lambda デプロイ方式 | **Lambda Web Adapter（LWA）** | Application Design Q-A 確定済み |

### 1.2 AWS SDK

| 技術 | 選定内容 | 根拠 |
|---|---|---|
| AWS SDK | **AWS SDK for Go v2** | 現行の推奨 SDK。DynamoDB Document Client を含む |
| DynamoDB クライアント | `github.com/aws/aws-sdk-go-v2/service/dynamodb` + `dynamodbattribute` | 標準的な Go × DynamoDB 実装 |

### 1.3 DynamoDB テーブル設計方針

| テーブル | キャパシティモード | RCU | WCU | 備考 |
|---|---|---|---|---|
| Wallet | プロビジョンド | 1 | 1 | ConsistentRead 使用、バースト容量前提（本番化時は 2 RCU 以上に変更） |
| BudgetSettings | プロビジョンド | 1 | 1 | SetBudget は低頻度 |
| IdempotencyRecord | プロビジョンド | 1 | 1 | TTL 24h（DynamoDB TTL 自動削除） |
| BudgetResetLog | プロビジョンド | 1 | 1 | 月 1 回のリセット処理 |

物理テーブル定義（PK/SK/GSI/TTL 詳細）は Infrastructure Design で決定する（凍結 IF §6.3 の Key Design を参照）。

### 1.4 冪等性・整合性

| 技術 | 選定内容 | 根拠 |
|---|---|---|
| 冪等性実装 | DynamoDB `PutItem` with `ConditionExpression: attribute_not_exists(idempotencyKey)` | Q-B4=A、FD §2.3 |
| 残高更新 | DynamoDB `UpdateItem` with `ConditionExpression: balance >= :amount` | NFR-REL-01、FD §4.3 |
| トランザクション | **使用しない**（個別操作） | Q-N5=B。TransactWriteItems の 2 倍 WCU コストを避ける |
| ULID 生成 | `github.com/oklog/ulid/v2` | idempotencyKey の ULID 部分生成（クライアント側で生成） |

### 1.5 構造化ログ

| 技術 | 選定内容 | 根拠 |
|---|---|---|
| ログライブラリ | **`log/slog`（Go 標準）** または `github.com/rs/zerolog` | NFR-OBS-01。具体選定は NFR Design で確定 |
| ログフォーマット | JSON | CloudWatch Logs Insights でのクエリ対応 |
| 出力項目 | `level`, `timestamp`, `userId`, `action`, `traceId`, `requestId`, `userAgent`, `amount`（該当時）, `newBalance`（該当時） | Q-N7=C、nfr-requirements.md §7.1 |

### 1.6 テストツール

| 技術 | 選定内容 | 根拠 |
|---|---|---|
| ユニットテスト | Go 標準 `testing` パッケージ | 追加依存なし |
| Property-Based Testing | `github.com/leanovate/gopter` または `pgregory.net/rapid` | Q-N8=B。具体選定は NFR Design / Code Generation で確定 |
| DynamoDB ローカル | **DynamoDB Local**（Docker）または **localstack** | 統合テスト用。具体選定は Code Generation で確定 |

---

## 2. フロントエンド（Unit B 担当コンポーネント）

### 2.1 フレームワーク・状態管理

Application Design で確定済みの共通スタックを使用する。

| 技術 | 選定内容 | 根拠 |
|---|---|---|
| フレームワーク | **Next.js（App Router）** | Application Design Q-B 確定済み |
| 状態管理 | **Jotai + TanStack Query** | Application Design Q-C 確定済み |
| デプロイ | **AWS Amplify Hosting** | Application Design Q-B 確定済み |

### 2.2 Unit B 固有のフロントエンド設定

| 技術 | 選定内容 | 根拠 |
|---|---|---|
| TanStack Query stale time | **30 秒** | Q-N10=A、Q-B6=A |
| `GetBalance` invalidate | `Deduct` 成功時に即 `invalidateQueries` | Q-N10=A、NFR-DEG-03 |
| API 通信 | **REST（JSON over HTTPS）** | Application Design Q-D 確定済み |

---

## 3. インフラ・スケジューラ

### 3.1 EventBridge Scheduler（`ResetAll`）

| 項目 | 選定内容 | 根拠 |
|---|---|---|
| スケジュール | `cron(0 15 L * ? *)` UTC（= JST 月末最終日 0:00） | Q-B8=A |
| Lambda タイムアウト | **30 秒** | Q-N3=A |
| Lambda メモリ | **128MB** | Q-N9=A |
| 失敗時の対応 | CloudWatch Logs に ERROR 記録 + 手動リカバリ手順（Infrastructure Design に記載） | Q-N6=C、NFR-REL-04 |

### 3.2 API Lambda

| 項目 | 選定内容 | 根拠 |
|---|---|---|
| Lambda タイムアウト | **29 秒**（API Gateway 上限） | AWS 制約 |
| Lambda メモリ | **128MB** | Q-N9=A |
| IaC | **Terraform** | Application Design Q-C（Q-12=C）確定済み |
| リージョン | **ap-northeast-1** | requirements.md 確定済み |

---

## 4. 後続ステージへの引き継ぎ

| 項目 | 引き継ぎ先 |
|---|---|
| DynamoDB テーブル物理設計（PK/SK/GSI/TTL 詳細） | Infrastructure Design |
| EventBridge Scheduler の Terraform 実装詳細 | Infrastructure Design |
| Lambda の IAM Role / DynamoDB 暗号化設定 | Infrastructure Design |
| 構造化ログライブラリの確定（slog vs zerolog） | NFR Design |
| PBT フレームワークの確定（gopter vs rapid） | NFR Design / Code Generation |
| DynamoDB Local / localstack の選定 | Code Generation |
| `idempotencyKey` バリデーション正規表現の実装 | Code Generation |
| `ResetAll` 手動リカバリ手順の文書化 | Infrastructure Design |
