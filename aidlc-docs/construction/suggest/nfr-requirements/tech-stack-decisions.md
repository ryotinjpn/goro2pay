# Unit D (`suggest`) — Tech Stack Decisions

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: D — `suggest`
**Stage**: NFR Requirements / Construction
**方針**: Unit A/B/C の確定スタックを**全面継承**。Unit D は新規技術を導入しない（Bedrock も Unit C と同一スタック）。本書は継承内容の確認と suggest 固有の適用先を明記する。
**参照**: [auth/nfr-requirements/tech-stack-decisions.md](../../auth/nfr-requirements/tech-stack-decisions.md)、[order/nfr-requirements/](../../order/nfr-requirements/)、[unit-interfaces.md](../../interfaces/unit-interfaces.md)

---

## 1. Backend (Go on Lambda) — 継承

| 役割 | 採用 | 出典 | Unit D での適用 |
|---|---|---|---|
| ランタイム | `provided.al2023`（custom runtime）/ Go 1.22 系 | Unit A | 共有 API Lambda に同居 |
| HTTP framework | `github.com/gin-gonic/gin` | Unit A | `suggest_handler.go`（`GET /api/suggest`） |
| Lambda adapter | AWS Lambda Web Adapter (LWA) | Unit A | 共有 |
| Logger | `log/slog` + JSON Handler | Unit A | suggest 構造化ログ（NFRD-D11） |
| AWS SDK | `aws-sdk-go-v2` | 共通 | DynamoDB（`GoroPay_Suggestion`）、Bedrock |
| **Bedrock SDK** | `aws-sdk-go-v2` `bedrockruntime` + **Converse API** | **Unit C NFRC-C20 継承** | 横串 `BedrockAdapter.InferSuggestion`（Unit C/D 共有） |
| エラー型 | 標準 `errors`（sentinel） | Unit A | — |

### 1.1 テスト（継承）

| 役割 | ライブラリ | Unit D での適用 |
|---|---|---|
| 単体テスト | `testing` + `github.com/stretchr/testify` | SuggestService / SuggestionStore |
| Property-Based Testing | `github.com/leanovate/gopter` | `BuildFromHistory` 最頻判定の決定性（NFRD-D12、軽量） |
| Bedrock mock | `BedrockAdapter` interface の mock（Unit C NFRC-C16 継承） | go test / CI は mock 必須 |

### 1.2 採用しないもの（Unit C と同じ）

- AWS X-Ray（NFRD-D10 / NFR-OBS-02）
- `PutMetricData` カスタムメトリクス（NFRD-D10）
- Bedrock Provisioned Throughput（NFRD-D06、MVP スコープ外）

---

## 2. Frontend (Next.js / TypeScript) — 継承

| 役割 | 採用 | Unit D での適用 |
|---|---|---|
| Next.js / React / TS | 15 系 / 19 / 5.x | `SuggestBubble`（GoroButton 内蔵） |
| State / データフェッチ | `jotai` / `@tanstack/react-query` v5 | `useSuggestion`（queryKey `['suggestion']`、NFRD-D17） |
| HTTP クライアント | 共有 `apiClient`（LC-AUTH-09、BFF 経由） | `GET /api/suggest`（独自クライアント作らない、契約 §12） |
| PBT | `fast-check` | レスポンス→props 変換ラウンドトリップ（NFRD-D12） |
| 単体/コンポーネント | Vitest + `@testing-library/react` | useSuggestion / SuggestBubble |
| E2E | Playwright | サジェスト表示 → 1 タップ注文（Unit C 結合） |

> 視覚（CSS トークン・アニメ）は横串デザインシステム（PR #94 `_design-system/`）が担う。Unit D は `SuggestBubble` の構造・データのみ。

---

## 3. AWS マネージドサービス

| サービス | 用途 | 設定（Infra Design で確定） |
|---|---|---|
| **Amazon DynamoDB** | `GoroPay_Suggestion`（Unit D 所有） | PK `suggestionId` / TTL `expiresAt`(30分) / **プロビジョンド 1 RCU/1 WCU**（NFRD-D08）/ GSI なし |
| **Amazon Bedrock** | サジェスト推論（Unit C/D 共有） | `claude-haiku-4-5`（ap-northeast-1 IP）/ Converse（NFRD-D16、Unit C IAM module 共有） |
| **AWS Lambda** | `GET /api/suggest`（共有 API Lambda） | 256MB / arm64 / 10s（NFRD-D15、Unit C 設定済み） |
| **Amazon API Gateway** | ルート | `GET /api/suggest`（Cognito Authorizer、Stage Throttling 100req/s 共有） |
| **Amazon CloudWatch Logs** | 構造化ログ | メトリクスフィルタのみ（NFRD-D10） |
| **AWS Budgets** | Bedrock コスト監視 | Unit C 設定済みに相乗り（NFRD-D16、合算 $10/月） |

---

## 4. インフラ管理（継承）

| 項目 | 採用 |
|---|---|
| IaC | Terraform（terraform-module-design / terraform-coding-rule / terraform-test 準拠） |
| 新規 module | `suggestion`（DynamoDB `GoroPay_Suggestion`）。Bedrock/observability は Unit C 既存 module を共有/参照（Infra Design で確定） |

---

## 5. バージョンマトリクス

Unit A/B/C と同一。Unit D は新規依存を追加しない（既存の gin / slog / gopter / aws-sdk-go-v2 / next / react-query / fast-check を流用）。最終 pin は Code Generation で実施。

---

## 6. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **NFR Design** | suggest のロガー初期化・レイテンシ計測パターン（Unit C 流用）、PBT 具体プロパティ |
| **Infrastructure Design** | `suggestion` module（DynamoDB）、`GET /api/suggest` ルート、Bedrock IAM 共有参照 |
| **Code Generation** | `SuggestService` / `SuggestionStore` / `suggest_handler.go` / `InferSuggestion` プロンプト / `useSuggestion` / `SuggestBubble` の最終実装、go.mod は既存流用 |
