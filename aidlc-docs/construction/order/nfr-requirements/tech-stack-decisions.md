# Unit C (`order`) — Tech Stack Decisions

**Document Version**: 1.0
**Created**: 2026-05-23
**Stage**: Construction / NFR Requirements
**Unit**: C — `order` (代行手配コア)
**Depth**: Comprehensive
**Related**: [nfr-requirements.md](./nfr-requirements.md), [order-nfr-requirements-plan.md](../../plans/order-nfr-requirements-plan.md)

本ドキュメントは Unit C `order` の **Tech Stack 確定事項** を、NFR Requirements の数値・しきい値要件に基づき集約する。最終的な実装パッケージ・バージョン・設定は Code Generation ステージで確定するが、**主要な選定方針** を NFR Requirements 段階で凍結する。

---

## 1. Backend (Go)

### 1.1 言語・ランタイム

| 項目 | 決定 | 由来 |
|---|---|---|
| 言語 | **Go 1.23.x**（Unit A / B と統一） | unit-of-work.md §4.1、Unit A/B 確定 |
| Web フレームワーク | **Gin**（Unit A / B と統一） | unit-of-work.md §4.1 |
| Lambda ランタイム | **`provided.al2023`**（カスタムランタイム、Go バイナリ） | aws-lambda-go ベストプラクティス |
| アーキテクチャ | **arm64 (Graviton2)** | NFRC-C18 |
| Lambda メモリ | **256MB** | NFRC-C18、Plan Q-N9=B |
| Lambda タイムアウト | **10 秒** | NFRC-C18 |

### 1.2 AWS SDK

| 項目 | 決定 | 由来 |
|---|---|---|
| AWS SDK バージョン | **AWS SDK for Go v2** (`github.com/aws/aws-sdk-go-v2`) | Unit A / B と統一 |
| Bedrock SDK | **`bedrockruntime`** クライアント | NFRC-C20 |
| DynamoDB SDK | **`dynamodb`** クライアント + `attributevalue` パッケージ | Unit B と統一 |
| Lambda ハンドラ | **`github.com/aws/aws-lambda-go/lambda`** + `events` | Unit A / B と統一 |

### 1.3 Bedrock 関連

| 項目 | 決定 | 由来 |
|---|---|---|
| モデル ID | **`apac.anthropic.claude-3-5-haiku-20241022-v1:0`**（ap-northeast-1 inference profile） | NFRC-C20、Plan Q-N8=B |
| API | **Converse API**（`Converse` メソッド） | FD 確定済み、NFRC-C20 |
| 想定トークン | 入力 500tok / 出力 200tok / リクエスト | NFRC-C20 |
| タイムアウト | **1.5 秒/呼出**、リトライ最大 1 回 | NFRC-C06 / NFRC-C07 |
| エラーハンドリング | `ThrottlingException` / `ServiceUnavailableException` / `context.DeadlineExceeded` でリトライ、`ValidationException` 等は即フォールバック | NFRC-C06 |
| 月次予算上限 | **$10/月**（負荷 10 倍想定） | NFRC-C20 |

### 1.4 ロギング・観測性

| 項目 | 決定 | 由来 |
|---|---|---|
| 構造化ログ | **`log/slog`**（標準ライブラリ、Go 1.21+） + `ContextAwareSlogHandler` (LC-AUTH-05 を再利用) | Unit B `nfr-design-patterns.md` 整合、Unit A/B と統一 |
| ログ出力先 | **CloudWatch Logs**（Lambda 標準出力経由） | NFR-OBS-01、NFRC-C12 |
| ログフォーマット | JSON 形式（NFRC-C12 の 8+11=19 項目） | NFRC-C12 |
| メトリクス | **CloudWatch Logs メトリクスフィルタ**（PutMetricData 不採用） | NFRC-C13、NFRC-C14 |
| トレーシング | **X-Ray は不採用** | NFRC-C14 / NFR-OBS-02 |
| Bedrock 本文ログ | **記録しない**（プロンプト本文 / レスポンス本文） | NFRC-C24 |

### 1.5 ULID ライブラリ

| 項目 | 決定 | 由来 |
|---|---|---|
| ULID 生成 | **`github.com/oklog/ulid/v2`**（`OrderID` 採番に使用、Frontend と同等の ULID） | FD `domain-entities.md` |
| Frontend ULID | **`ulid` npm パッケージ** | FD `frontend-components.md` §10、`lib/ulid.ts` |

### 1.6 PBT フレームワーク

| 項目 | 決定 | 由来 |
|---|---|---|
| PBT ライブラリ | **`github.com/leanovate/gopter`**（Unit B と統一） | NFRC-C15、Plan Q-N5=B |
| 適用範囲 | P-1（残高不変）+ P-3（冪等性レスポンス一貫性）の 2 プロパティ | NFRC-C15 |
| 適用外 | P-2（履歴整合）、フォールバック分岐閾値（テーブルドリブンテストで対応） | NFRC-C15 |
| 実行環境 | **mock 経由のみ**（Bedrock 実呼出しは行わない） | NFRC-C16 |

### 1.7 Mock ライブラリ

| 項目 | 決定 | 由来 |
|---|---|---|
| Mock 手法 | **手書き mock**（or `gomock` 必要に応じて） | NFRC-C16 |
| Mock 配置 | `internal/adapters/bedrock/mock/mock_bedrock.go` | NFRC-C16 |
| Mock シナリオ | 標準応答（成功）/ ThrottlingException / Timeout / 永続エラー の 4 種 | NFRC-C16 / NFRC-C17 |

---

## 2. Frontend (Next.js)

### 2.1 言語・フレームワーク

| 項目 | 決定 | 由来 |
|---|---|---|
| 言語 | **TypeScript**（Unit A / B と統一） | unit-of-work.md §4.1 |
| フレームワーク | **Next.js (App Router)** | unit-of-work.md §4.1 |
| ホスティング | **AWS Amplify Hosting**（Unit A 横串インフラ） | Unit A Infrastructure Design |
| BFF パターン | **Next.js Server Route Handler 経由**（`/api/*` プロキシ） | Unit A Infrastructure Design |

### 2.2 状態管理・データフェッチ

| 項目 | 決定 | 由来 |
|---|---|---|
| サーバ状態 | **TanStack Query v5**（Unit B と統一） | Unit B Tech Stack |
| `useOrderHistory` 設定 | `staleTime: 60_000` / `gcTime: 300_000` / `refetchOnWindowFocus: true` | NFRC-C19 / Plan Q-N10=A |
| `useOrder` mutation 設定 | `retry: 0`、`onSuccess` で `['orderHistory']` と `['balance']` invalidate | NFRC-C19 |
| クライアント状態 | **Jotai atom**（Unit A と統一、必要に応じて） | Unit A `nfr-design-patterns.md` |
| クロスユニット query key | `['balance']` invalidate（unit-interfaces.md §9.1 凍結契約に従う） | NFRC-C19 |

### 2.3 認証・API クライアント

| 項目 | 決定 | 由来 |
|---|---|---|
| API クライアント | **手書き fetch ラッパ**（Unit A 共通の `apiClient`、LC-AUTH-09 を再利用） | Unit A `nfr-design-patterns.md` |
| API 認証ヘッダ | **`Authorization: Bearer <AccessToken>`**（Unit A AccessToken 採用と整合） | Unit A AccessToken 採用反映 |
| BFF プロキシ | **catch-all Route Handler**（Unit A LC-AUTH-18 BffProxyRouteHandler を再利用） | Unit A Infrastructure Design |

### 2.4 UI コンポーネント

| 項目 | 決定 | 由来 |
|---|---|---|
| `GoroButton` | カスタムコンポーネント（押下中の連打抑制 1 秒） | NFRC-C22 |
| `OrderCompletionScreen` | `app/order/[id]/complete/page.tsx`、5 秒後自動遷移 | FD frontend-components.md / NFRC-C21 |
| トースト | 既存ライブラリ（Unit A 採用と統一、最終選定は Code Generation） | NFRC-C22 |

---

## 3. Infrastructure（高レベル方針のみ、詳細は Infrastructure Design へ）

### 3.1 AWS リソース

| リソース | 用途 | 詳細引き継ぎ先 |
|---|---|---|
| Lambda 関数 | `POST /api/orders` / `GET /api/orders` を 1 関数で処理（Go ビルド分割不要） | Infrastructure Design |
| API Gateway | REST API、Cognito Authorizer 経由（Unit A 横串） | Infrastructure Design |
| DynamoDB `GoroPay_OrderHistory` | TTL 90 日、PK=userID、SK=orderedAt-orderID | Infrastructure Design |
| Amazon Bedrock | Claude 3.5 Haiku 呼出、IAM Role 設定 | Infrastructure Design |
| CloudWatch Logs | Lambda ログ、構造化 JSON | Infrastructure Design |
| CloudWatch Alarms | NFRC-C13 の 3 アラーム | Infrastructure Design |
| SNS Topic | アラーム通知（メール） | Infrastructure Design |
| AWS Budgets | Bedrock $5/月 アラート | Infrastructure Design |

### 3.2 IAM 権限

| 権限 | 範囲 |
|---|---|
| `bedrock:InvokeModel` / `bedrock:Converse` | `apac.anthropic.claude-3-5-haiku-*` モデル ARN に限定 |
| `dynamodb:Query` / `dynamodb:PutItem` / `dynamodb:UpdateItem` | `GoroPay_OrderHistory` テーブルのみ |
| `dynamodb:GetItem` / `dynamodb:UpdateItem` | Unit B 所有: `GoroPay_Wallet` / `GoroPay_IdempotencyRecord` 経由（Unit B `WalletService.Deduct` を呼ぶため Unit C 自体は直接アクセスしない） |
| `logs:CreateLogStream` / `logs:PutLogEvents` | Lambda 実行ロール標準 |

### 3.3 環境別 Bedrock 認証

| 環境 | Bedrock 認証 |
|---|---|
| local / CI | mock のみ、AWS 認証情報不要 |
| dev / stg / prd | Lambda 実行ロール経由 |

---

## 4. Cross-Unit 統合

### 4.1 Unit B 連携

| 項目 | 決定 |
|---|---|
| 残高引き落とし | `WalletService.Deduct(ctx, userID, idempotencyKey, amount)` 呼出（凍結契約 §3.1） |
| 残高不足時 | `wallet.ErrInsufficientFunds` 受信 → 402 応答 |
| 冪等性衝突時 | `wallet.ErrIdempotencyConflict` 受信 → 409 応答（NFRC-C05 / BR-C39） |
| 冪等命中時 | Wallet payload から `orderID` を取得（凍結契約整合修正後の方式） |

### 4.2 Unit D 連携

| 項目 | 決定 |
|---|---|
| サジェスト経由注文 | `SuggestService.ResolveSuggestion(ctx, suggestionID)` 呼出（FD Q-5）、Bedrock 再呼出しなし |
| サジェスト失効時 | 透過的に通常 Bedrock フローへフォールバック（FD Q-6） |

### 4.3 横串 Adapter 連携

| Adapter | 配置 | 用途 |
|---|---|---|
| `BedrockAdapter` | `internal/adapters/bedrock/`（Unit C / D 共有） | Claude 3.5 Haiku 呼出 |
| `DeliveryAdapter` | `internal/adapters/delivery/`（`MockDeliveryAdapter` 実装） | モック手配 |
| `FallbackSuggestProvider` | `internal/adapters/fallback/`（Unit C / D 共有） | フォールバック Plan 生成 |

---

## 5. NFR との整合性チェック

| 決定 | 関連 NFR | 整合性 |
|---|---|---|
| Lambda 256MB / arm64 | NFRC-C04（コールドスタート 600ms 以内） | ✅ |
| Bedrock 3.5 Haiku ap-northeast-1 | NFRC-C01（通常パス p95 2.5s）/ NFRC-C20 | ✅ |
| `gopter` PBT (P-1/P-3) | NFRC-C15 / NFRC-C16 | ✅ |
| TanStack Query v5 (60s + invalidate) | NFRC-C19 | ✅ |
| `log/slog` + 19 項目 | NFRC-C12 | ✅ |
| CloudWatch Logs メトリクスフィルタ | NFRC-C13 / NFRC-C14 / NFR-OBS-02 | ✅ |
| AccessToken 採用 | Unit A AccessToken 統一方針 | ✅ |
| BFF パターン | Unit A Infrastructure Design 整合 | ✅ |
| Bedrock 本文ログ非記録 | NFRC-C24 / NFR-COMP-02 | ✅ |

---

## 6. 想定外の論点（後続ステージへの引き継ぎ）

| 引き継ぎ先 | 内容 |
|---|---|
| **NFR Design** | `slog` ハンドラ初期化パターン、Bedrock SDK の context.WithTimeout 適用パターン、リトライロジックのコード設計、PBT の具体プロパティ実装方針、レイテンシ計測ミドルウェア |
| **Infrastructure Design** | DynamoDB テーブル詳細、IAM ポリシー JSON、API Gateway ルート / Throttle / CORS、Lambda 関数 Terraform、CloudWatch Alarms / メトリクスフィルタ Terraform、AWS Budgets Terraform |
| **Code Generation** | Bedrock プロンプトテンプレート最終、`gopter` テストコード、mock 実装、自虐トースト文言、`useOrder` / `useOrderHistory` 実装、`GoroButton` / `OrderCompletionScreen` 実装 |
| **Build and Test** | go test / CI 設定、dev 環境 Bedrock 実呼出し検証手順、E2E テストシナリオ |

---

## 7. 文書管理

- **承認**: ユーザ承認待ち（Construction フェーズの per-unit ループ承認ゲート）
- **凍結契約への影響**: なし
- **次ステージ**: NFR Design（per-unit、Comprehensive 深度）
