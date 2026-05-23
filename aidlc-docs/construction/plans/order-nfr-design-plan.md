# Order Unit (Unit C) — NFR Design Plan

**Document Version**: 0.1 (Draft, awaiting user approval)
**Created**: 2026-05-23
**Unit**: C (`order` / 代行手配コア)
**Construction Depth**: Comprehensive
**Stage**: NFR Design (Construction Phase)
**Prerequisite**: NFR Requirements 承認済み (PR #78 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit C（`order` / 代行手配コア）の **NFR Design ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。NFR Requirements (NFRC-C01〜C25) で確定した数値・しきい値を **どう実現するか（パターン・論理コンポーネント）** を Comprehensive 深度で設計する。

### 1.1 NFR Requirements からの引き継ぎ事項

| NFR ID | 値 / 制約 | NFR Design で扱う論点 |
|---|---|---|
| NFRC-C01 | E2E p95 3.0s / p99 5.0s + ケース別サブターゲット | レイテンシ計測ミドルウェア、各ステップの span 切り出しパターン |
| NFRC-C03 | 親 Context タイムアウト 5 秒 | `context.WithTimeout` 適用箇所と伝播パターン |
| NFRC-C05 | 冪等性 TTL 24h、Wallet payload から OrderID 復元 | Wallet `Deduct` 呼出・冪等命中時の OrderID 取得パターン |
| NFRC-C06 | Bedrock リトライ 1 回、対象/非対象エラー | リトライ判定ロジックのパターン化（エラー型分類） |
| NFRC-C07 | Bedrock 1 呼出 1.5s タイムアウト | `context.WithTimeout(parent, 1500ms)` 適用パターン |
| NFRC-C08 | フォールバック分岐閾値 5 件 | `BuildFromHistory` / `Default` の選択ロジックパターン |
| NFRC-C09 | OrderHistory Insert 失敗時 200 応答 | エラーログ + 200 応答の分離パターン（warning ログ） |
| NFRC-C10 | 親 Context Cancel 即時伝播 | Bedrock / DeliveryAdapter / DynamoDB へのキャンセル伝播パターン |
| NFRC-C12 | 構造化ログ 19 項目（共通 8 + Unit C 11） | `slog` + `ContextAwareSlogHandler` の Unit C 拡張、ログ出力タイミング |
| NFRC-C13 | CloudWatch アラーム 3 種 | アラーム検知に必要なログフィールド付与パターン |
| NFRC-C15 | PBT P-1 + P-3（gopter） | プロパティ実装の Generator / Property 設計、shrinking 戦略 |
| NFRC-C16 | Bedrock スタブ環境別マトリクス | mock 設計（4 シナリオ）、interface 抽出パターン |
| NFRC-C18 | Lambda 256MB / arm64 / 10s タイムアウト | コールドスタート最適化（init 関数、SDK 初期化遅延化） |
| NFRC-C19 | TanStack Query 60s + invalidate | `useOrder` / `useOrderHistory` の Loading / Error / Success state パターン |
| NFRC-C20 | Bedrock Claude 3.5 Haiku Converse API | プロンプト組み立てパターン、モデル ID 環境別切替 |
| NFRC-C22 | 自虐トースト 3 種ローテーション、402 → 0.3s 遷移 | エラー UX 実装パターン（toast / router / disable） |
| NFRC-C24 | Bedrock 本文ログ非記録、PII 境界 | ログマスキング・除外パターン |

### 1.2 NFR Design で扱うこと / 扱わないこと

| 扱う | 扱わない |
|---|---|
| 設計パターン（リトライ判定 / フォールバック分岐 / レイテンシ計測 / Toast ローテーション 等） | パターンの具体実装コード（→ Code Generation） |
| 論理コンポーネント（OrderService / OrderHandler / BedrockAdapter / FallbackSuggestProvider / GoroButton / OrderCompletionScreen / ToastHost 等） | 論理コンポーネントの内部実装ロジック（→ Code Generation） |
| Bedrock SDK init 配置（cold start 最適化） | Bedrock IAM Role 設定（→ Infrastructure Design） |
| 構造化ログハンドラの拡張パターン | ライブラリ初期化コード（→ Code Generation） |
| PBT プロパティの抽象設計（Generator / Property） | テストコード（→ Code Generation） |
| 横串 Adapter（Bedrock / Delivery / Fallback）の interface 設計と DI パターン | Adapter 内部実装（→ Code Generation） |
| Frontend ローディング/エラー UX のコンポーネント分離パターン | コンポーネント実装（→ Code Generation） |
| Mock 設計（4 シナリオの interface 抽出） | Mock コード（→ Code Generation） |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1 問ずつ、`feedback_interview_style` に従う）
- [x] 回答の曖昧さ・矛盾を点検し、必要なら追加サブ質問を挟む（10 観点で矛盾チェック実施、整合確認）
- [x] `aidlc-docs/construction/order/nfr-design/nfr-design-patterns.md` を作成（14 パターン）
- [x] `aidlc-docs/construction/order/nfr-design/logical-components.md` を作成（LC-ORDER-01〜34）
- [x] `aidlc-docs/construction/interfaces/unit-interfaces.md` の更新検討（破壊的変更なし、軽微な追記不要と判断）
- [x] aidlc-state.md / audit.md / Plan checkboxes を更新
- [ ] 完了メッセージを提示し、承認ゲート（2-option）に進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-D1: Bedrock リトライ判定ロジックのエラー分類パターン

NFRC-C06 で「対象エラー（Throttle/Timeout/ServiceUnavailable/ネットワーク全般）はリトライ、永続エラー（Validation/AccessDenied）は即フォールバック」と確定済み。Go で「リトライすべきか」の判定実装パターン。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`errors.Is` / `errors.As` で Bedrock SDK エラー型を判定** | AWS SDK v2 の `*types.ThrottlingException` 等を直接判定。SDK の型に依存 |
| B | **HTTP ステータスコード + エラーコード文字列で判定** | SDK バージョンアップに強い、ただし文字列マッチで脆い |
| C | **共通の `RetryClassifier` interface を切り出し、Bedrock 専用 implementation を注入** | テスト容易、Unit D 共有時に再利用可能（横串化） |
| D | **A + C のハイブリッド**（Bedrock SDK 型を中心に判定し、Classifier interface で抽象化） | テスト容易性 + SDK 型の安全性 |

[Answer]: **D（A + C ハイブリッド）**

追記事項:
- **論理コンポーネント**:
  - `LC-ORDER-RetryClassifier` (interface): `ShouldRetry(err error) bool`
  - `LC-ORDER-BedrockRetryClassifier` (implementation): SDK 型を `errors.As` / `errors.Is` で判定
- **設計パターン P-RETRY-01 Bedrock Retry Classification**:
  - リトライ対象: `*types.ThrottlingException` / `*types.ServiceUnavailableException` / `context.DeadlineExceeded` / その他ネットワークエラー（`*smithy.GenericAPIError` で判定）
  - 即フォールバック: `*types.ValidationException` / `*types.AccessDeniedException` / `*types.ResourceNotFoundException`
  - 抽象化: `RetryClassifier` interface 経由で `OrderService` に注入
- **テスト戦略**: `fakeClassifier` を使った PBT 互換性確保
- **Unit D との共有可能性**:
  - 同じ `RetryClassifier` interface / `BedrockRetryClassifier` implementation を Unit D が再利用
  - 横串配置: `internal/adapters/bedrock/retry.go`

### Q-D2: フォールバック分岐ロジックの配置パターン

NFRC-C08 でフォールバック分岐閾値 5 件確定。`OrderService.PlaceOrder` 内に直接書くか、別コンポーネントに分離するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`OrderService` 内に if 文で直接記述** | シンプル、可読性高、テストはサービスレベルで実施 |
| B | **`PlanBuilder` コンポーネント抽出**（`Build(ctx, history) (*Plan, error)` interface） | テスト容易、Unit D で再利用可能。Bedrock + Fallback の判定責務を 1 か所に集約 |
| C | **`FallbackSuggestProvider` interface に閾値判定を含める**（既存横串 Adapter 拡張） | 横串化最大限、ただし interface が肥大化 |

[Answer]: **B（PlanBuilder コンポーネント抽出）**

追記事項:
- **論理コンポーネント**:
  - `LC-ORDER-PlanBuilder` (interface): `Build(ctx, history []OrderRecord) (*Plan, error)`
  - `LC-ORDER-BedrockPlanBuilder` (implementation): `BedrockAdapter` + `RetryClassifier` (Q-D1) + `FallbackSuggestProvider` を依存に持つ。Bedrock 呼出 → リトライ → フォールバック分岐の責務を集約
- **設計パターン P-PLAN-01 Plan Construction Strategy**:
  - 入力: `history []OrderRecord`
  - 出力: `*Plan { StoreName, MenuName, Amount, Category, Source }`、`Source` は `"bedrock" | "fallback_history" | "fallback_default"` の enum
  - 分岐ロジック: Bedrock 呼出（NFRC-C06/C07）→ 失敗時、履歴 ≥ 5 件で `BuildFromHistory`、< 5 件で `Default`（NFRC-C08、BR-C07）
  - 配置: `internal/order/plan_builder.go`（Unit C 内、Unit D 共有時は横串移動可能な interface 設計）
- **`OrderService.PlaceOrder` の簡素化**:
  - フォールバック分岐ロジックを `PlanBuilder` に委譲
  - `OrderService` は冪等性 / Wallet / Adapter / History のオーケストレーションのみ担当
- **テスト戦略**:
  - `PlanBuilder` 単体テスト: 4 シナリオ（成功 / リトライ後成功 / フォールバック履歴 / フォールバック Default）
  - `OrderService` 統合テスト: `PlanBuilder` を mock 化、オーケストレーションのみ検証

### Q-D3: レイテンシ計測パターン

NFRC-C01 (E2E p95 3.0s / p99 5.0s)、NFRC-C12 の `bedrockLatencyMs` ログ項目。各ステップ（Bedrock / Wallet / Adapter / History）のレイテンシをどう計測・記録するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **手動 `time.Now()` / `time.Since()` で各ステップ計測、構造化ログに付与** | 明示的、依存なし、コード重複あり |
| B | **`measure(ctx, name, fn)` ヘルパ関数で計測 → log にスパン情報付与** | DRY、テスト容易、関数ラップのオーバーヘッド軽微 |
| C | **`slog` の Group 機能 + `latencyMs` を全ログに付与（middleware 化）** | E2E は middleware で、各ステップは個別 measure ヘルパ |

[Answer]: **C（middleware で E2E + measure ヘルパで各ステップ）**

追記事項:
- **設計パターン P-OBS-01 Latency Measurement**:
  - **E2E レイテンシ**: `LatencyMiddleware`（Gin handler chain の最初）で計測。全 API リクエストの `latencyMs` / `statusCode` / `path` を `request_complete` ログに付与。配置: `internal/middleware/latency.go`
  - **各ステップレイテンシ**: `measure(ctx, name, fn)` ヘルパで計測。計測対象: Bedrock / Wallet.Deduct / DeliveryAdapter.Place / OrderHistory.Insert / PlanBuilder.Build。配置: `internal/observability/measure.go`（generic 関数、横串）
- **NFRC-C12 マッピング**:
  - `bedrockLatencyMs` → `measure(ctx, "bedrock", ...)` の出力をサマリログにコピー
  - `bedrockAttempt` → リトライカウンタ、サマリログに付与
  - `fallbackTriggered` → PlanBuilder で判定、サマリログに付与
- **論理コンポーネント**:
  - `LC-ORDER-LatencyMiddleware`: Gin middleware、E2E レイテンシ計測（横串、全 Unit で利用）
  - `LC-ORDER-MeasureHelper`: generic 関数、各ステップレイテンシ計測（横串）
- **Unit A `ContextAwareSlogHandler` (LC-AUTH-05) との連携**: middleware と `measure` の出力は `slog.InfoContext(ctx, ...)` 経由のため、`traceId` / `requestId` / `userId` が自動付与される。Unit C 固有のレイテンシ項目を足せば 19 項目が揃う
- **NFRC-C13-1 アラーム要件との整合**: middleware で E2E レイテンシが必ず付与されるため、CloudWatch Logs メトリクスフィルタによる p95 検知の信頼性が最大化

### Q-D4: Bedrock SDK の init 配置（コールドスタート最適化）

NFRC-C18 でコールドスタート目標 P95 600ms。Bedrock SDK クライアント初期化を `lambda.Start` 前（init 関数）に行うか、初回リクエスト時に遅延初期化するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **package-level `init()` で SDK クライアント初期化**（即時） | コールドスタート時に必ず初期化、ウォーム後は即時応答 |
| B | **`sync.Once` で初回リクエスト時に遅延初期化** | 初回のみ遅延、以降は即時。ウォームアップ Probe 不実施なら A と差なし |
| C | **`main.go` で起動時 + `sync.Once` で再生成（context cancel 時等）** | 高度、過剰 |

[Answer]: **A（package-level `init()` で SDK 初期化）**

追記事項:
- **設計パターン P-INIT-01 Lambda Cold Start Optimization**:
  - **Bedrock SDK 初期化**: `internal/adapters/bedrock/adapter.go` の `init()` で `bedrockruntime.Client` を package-level 変数に保持
  - **DynamoDB SDK 初期化**: `internal/repo/order_history/repository.go` の `init()` で `dynamodb.Client` を package-level 変数に保持
  - **AWS Config**: `init()` 内で `config.LoadDefaultConfig(context.Background())` を呼出。`AWS_REGION` 環境変数（Lambda 標準）を使用
  - **Fail-fast**: `init()` で SDK 初期化失敗時は `panic` （Lambda 起動失敗 → CloudWatch 即検知）
- **論理コンポーネント**:
  - `LC-ORDER-BedrockClientInit`: package-level 変数 + init() （`internal/adapters/bedrock/`）
  - `LC-ORDER-DynamoClientInit`: package-level 変数 + init() （`internal/repo/order_history/`）
- **テスト戦略**:
  - Production: `NewClaudeBedrockAdapter()` で `defaultClient` を使用
  - Test: `NewClaudeBedrockAdapterWithClient(mockClient)` で mock 注入（Q-D8 mock 戦略との連携）
- **コールドスタート目標達成の根拠**:
  - INIT フェーズで AWS 提供の burst CPU を活用 → SDK 初期化を 250〜400ms で完了
  - INVOKE フェーズは SDK 再利用 → 600ms 以内（NFRC-C18 達成）
- **Unit B / Unit D との整合**: 横串 `internal/adapters/bedrock/` は Unit D との共有のため、init パターンを Unit D 側でも統一

### Q-D5: 横串 Adapter の interface 抽出と DI パターン

横串コンポーネント（`BedrockAdapter` / `DeliveryAdapter` / `FallbackSuggestProvider`）は Unit C / D で共有。NFRC-C16 の mock 設計と DI 戦略。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **手動 DI**（`main.go` で interface に implementation を注入） | シンプル、Unit B と統一、デモ規模に最適 |
| B | **`google/wire` 等のコード生成 DI** | 大規模向け、過剰 |
| C | **handler の関数引数で渡す（factory 関数経由）** | 中規模 OK、テスト容易 |

[Answer]: **A（手動 DI、`main.go` で interface に implementation を注入）**

追記事項:
- **設計パターン P-DI-01 Manual Dependency Injection**:
  - 配線箇所: `apps/api/main.go` で全 component を組み立て
  - DI 順序:
    1. Q-D4=A の `init()` で SDK client を package-level 変数に設定（Bedrock / DynamoDB）
    2. `main.go` で Adapter / Provider を生成（`NewClaudeBedrockAdapter()` 等、内部で `defaultClient` 使用）
    3. PlanBuilder / Service を組み立て（Adapter 注入）
    4. Handler を組み立て（Service 注入）
    5. Gin Router に Handler 登録
- **interface 定義配置**:
  - `internal/adapters/bedrock/adapter.go`: `BedrockAdapter` interface + `ClaudeBedrockAdapter` implementation
  - `internal/adapters/delivery/adapter.go`: `DeliveryAdapter` interface + `MockDeliveryAdapter` implementation
  - `internal/adapters/fallback/provider.go`: `FallbackSuggestProvider` interface + implementation
  - `internal/order/plan_builder.go`: `PlanBuilder` interface + `BedrockPlanBuilder` implementation (Q-D2=B)
  - `internal/order/service.go`: `OrderService` interface + struct implementation
  - `internal/order/handler.go`: `OrderHandler` struct
  - `internal/repo/order_history/repository.go`: `OrderHistoryRepository` interface + implementation
- **テスト戦略**:
  - 単体テスト: 各 component の constructor に mock を直接注入
  - 統合テスト: `main.go` の組み立てを `setupTestServer()` ヘルパで再現、mock のみ差し替え
  - PBT (Q-D7): `OrderService` レベルで全 mock 注入
- **Unit A / B / D との整合**:
  - Unit A: 手動 DI を `apps/api/main.go` で実施済み（PR #71）
  - Unit B: 同じ手動 DI パターン採用（NFR Design Q-D5）
  - Unit C: 同じパターンで Unit A の `main.go` に追記（横串 `internal/adapters/bedrock/` は Unit D 追加時にもそのまま流用）
  - Unit D / E: 同パターン継承予定
- **`wire` 不採用の根拠**: コンポーネント数 < 10 でメリット薄、AI-DLC ワークフローで `wire_gen.go` の生成・管理が複雑化

### Q-D6: 構造化ログ Handler の Unit C 拡張パターン

NFRC-C12 で 11 項目の Unit C 固有ログ項目を追加。Unit A の `ContextAwareSlogHandler`（既存 `LC-AUTH-05`）を再利用しつつ Unit C 固有の context key を付与する設計。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`ContextAwareSlogHandler` をそのまま再利用、各 handler 関数で `slog.Info(..., attr1, attr2)` で都度フィールド付与** | 既存実装の再利用最大限、ログ呼出が冗長 |
| B | **`OrderLogger` ラッパ関数を Unit C 内に定義**（`logOrder(ctx, level, msg, attrs)` で 11 項目を自動付与） | DRY、ログ呼出シンプル |
| C | **`ContextAwareSlogHandler` を Unit C 専用にサブクラス化**（Order context から自動的に 11 項目を抽出） | 最大限自動化、ハンドラ複雑化 |

[Answer]: **B（`OrderLogger` ラッパ関数を Unit C 内に定義）**

追記事項:
- **設計パターン P-OBS-02 Order LogSummary**:
  - 目的: NFRC-C12 の Unit C 固有 11 項目を `PlaceOrder` リクエスト中に蓄積し、完了時に 1 件のサマリログ（`place_order_complete` event）として出力
  - 配置: `internal/order/logger.go`
  - 責務: 11 項目の状態を保持する struct、各項目の setter、`defer summary.LogComplete()` パターンで完了時一括出力
- **Unit A `ContextAwareSlogHandler` (LC-AUTH-05) との責務分離**:
  - 共通 8 項目: `ContextAwareSlogHandler` が context から自動付与（`level / timestamp / userId / action / traceId / requestId / email_hash / userAgent`）
  - Unit C 固有 11 項目: `LogSummary` がメモリで蓄積、`LogComplete` で出力
- **使用パターン**:
  ```go
  func (s *OrderService) PlaceOrder(ctx, req) (*PlaceOrderResult, error) {
      summary := order.NewLogSummary(ctx)
      defer summary.LogComplete()
      summary.SetIdempotencyKey(req.IdempotencyKey)
      summary.SetSource(req.Source())
      // ... 各ステップで蓄積
      return result, nil
  }
  ```
- **PII 防御**: `LogSummary` struct に Bedrock プロンプト本文・レスポンス本文のフィールドを意図的に含めないことで、構造的に漏洩リスクを排除（NFRC-C24 整合、タイプセーフな防御）
- **論理コンポーネント**: `LC-ORDER-LogSummary` (`internal/order/logger.go`)
- **Q-D3=C / Q-D13 との連携**: `measure` ヘルパ出力（`bedrockLatencyMs` 等）を `LogSummary` の setter で受け、サマリログに集約
- **Unit B / D との整合**: Unit B は項目少（2 個）のため A 相当で十分、Unit D は Bedrock 関連項目多のため Unit C と同じ `LogSummary` パターン採用検討（横串化余地、Code Generation で判断）

### Q-D7: PBT P-1 / P-3 のプロパティ設計（gopter）

NFRC-C15 で PBT 確定。具体的な Generator / Property 設計。

**P-1（残高不変）**: 同一 idempotencyKey の N 回送信に対し、残高変動は最大 1 回
**P-3（冪等性レスポンス一貫性）**: 同一 idempotencyKey の N 回送信に対し、レスポンス payload は 2 回目以降完全一致

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`Deduct` を mock 化したサービスレベル PBT**（OrderService 全体を対象、Wallet/Bedrock/Delivery/History 全て mock） | 高速、Unit C スコープに集中、Wallet 実装に依存しない |
| B | **A + Bedrock のみ実呼出し（通常時のみ）** | より現実的、ただし Q-N13=D で「PBT は mock 必須」確定済みなので採用しない |
| C | **A + Repository を Inmemory 実装（`mockMap[orderID]OrderRecord`）** | 実装に近いテスト、状態遷移の整合性まで含めて検証 |

[Answer]: **C（A + Inmemory Repository ハイブリッド）**

追記事項:
- **設計パターン P-PBT-01 Order Property-Based Testing**:
  - PBT フレームワーク: `gopter`（Unit B と統一、NFRC-C15）
  - テスト構成: ハイブリッド mock + inmemory
    - mock 化対象: BedrockAdapter / DeliveryAdapter / FallbackSuggestProvider / PlanBuilder
    - inmemory 実装対象: WalletService（テスト用 stub） / OrderHistoryRepository
- **配置**:
  - `internal/order/service_pbt_test.go`: PBT テストコード
  - `internal/order/inmemory_wallet_test.go`: inmemory `WalletStub`（Unit B WalletService interface 実装）
  - `internal/repo/order_history/inmemory.go`: inmemory `OrderHistoryRepository` 実装
- **Generator 設計**:
  - `idempotencyKey`: `gen.RegexMatch("^[0-9A-HJKMNP-TV-Z]{26}$")`（ULID 形式）
  - `userId`: `gen.AlphaString`（Cognito sub 形式）
  - `category`: `gen.OneConstOf("food")`（MVP 範囲、BR-C27）
  - `historyCount`: `gen.IntRange(0, 30)`（フォールバック分岐閾値検証範囲）
- **Property 設計**:
  - **P-1**: 同一 idempotencyKey の N 回送信に対し残高変動は最大 1 回（Wallet/History inmemory で状態遷移検証）
  - **P-3**: 同一 idempotencyKey の N 回送信に対しレスポンス payload は 2 回目以降完全一致（OrderID / Amount / StoreName / MenuName）
- **論理コンポーネント**:
  - `LC-ORDER-InmemoryHistory`: `OrderHistoryRepository` interface の in-memory 実装（テスト専用）
  - `LC-ORDER-WalletStub`: Unit B `WalletService` interface の in-memory 実装（テスト専用）
- **実装/PBT 互換性の保証**: inmemory 実装は production と同じ interface を実装、production の DynamoDB-specific な race condition は別途 NFRC-C17 統合テストで検証
- **Unit B との連携**: Unit B PBT で `WalletService.Deduct` の冪等性が保証、Unit C PBT は WalletStub で Unit C 視点の N 回送信挙動を検証 → Unit 境界の契約（unit-interfaces.md §3.1）を双方向に保証
- **サンプル数 / shrinking 設定**: デフォルト 100 サンプル、CI 実行時間 100ms 以内（合計 PBT スイートで 1 秒以内）

### Q-D8: Bedrock Mock の 4 シナリオ実装パターン

NFRC-C16 で mock の 4 シナリオ（成功 / Throttle / Timeout / 永続エラー）を実装。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **シナリオごとに別 struct（`MockSuccessAdapter` / `MockThrottleAdapter` / ...）** | テストごとに使い分けやすい、struct が増える |
| B | **単一 mock struct + `SetScenario(ScenarioSuccess \| ScenarioThrottle \| ...)` メソッド** | 1 mock で全シナリオ、ステートフル |
| C | **`MockBedrockAdapter` + テストごとに closure を注入**（`mock.InferOrderPlanFunc = func(ctx) {...}`） | 最大限柔軟、mock コード最小 |

[Answer]: **C（`MockBedrockAdapter` + closure 注入）**

追記事項:
- **設計パターン P-MOCK-01 Function-Field Mock**:
  - 目的: NFRC-C16 の Bedrock スタブを Go 慣用句的な closure 注入パターンで実装
  - 配置:
    - `internal/adapters/bedrock/mock.go`: `MockBedrockAdapter` struct
    - `internal/adapters/delivery/mock.go`: `MockDeliveryAdapter` struct（既存の production MockDeliveryAdapter を closure 注入対応に拡張）
    - `internal/adapters/fallback/mock.go`: `MockFallbackProvider` struct
  - 構造: `Func` フィールド (closure) + `Calls` カウンタ。`Func` 未設定時はデフォルト成功応答を返す
- **NFRC-C16 4 シナリオ実装例**:
  - 成功: `mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) { return successPlan, nil }`
  - Throttle: `mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) { return nil, &types.ThrottlingException{...} }`
  - Timeout: `mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) { return nil, context.DeadlineExceeded }`
  - 永続エラー: `mock.InferOrderPlanFunc = func(ctx, h) (*Plan, error) { return nil, &types.ValidationException{} }`
- **PBT (Q-D7=C) との連携**: PBT 実行時は決定論的な closure をセット、Sample ごとに再代入することで「異なる Bedrock 挙動」をシミュレート可能
- **NFRC-C17 統合テスト 9 シナリオ**: テスト関数内で closure を切り替えて全シナリオを検証（成功 / リトライ成功 / フォールバック / 永続エラー / 冪等命中 / 連打 / 残高不足 / Context Timeout / Context Cancel）
- **論理コンポーネント**:
  - `LC-ORDER-MockBedrockAdapter`: テスト専用、`internal/adapters/bedrock/mock.go`
  - `LC-ORDER-MockDeliveryAdapter`: 既存 `MockDeliveryAdapter` を closure 注入対応に拡張
  - `LC-ORDER-MockFallbackProvider`: テスト専用、`internal/adapters/fallback/mock.go`
- **Unit D との整合**: Bedrock を共有するため `MockBedrockAdapter` を Unit D テストでも再利用（横串、Code Generation で確認）

### Q-D9: Frontend `useOrder` hook のエラーハンドリングパターン

NFRC-C22 で 402 → 0.3 秒以内 BudgetEmpty 遷移、500 → 自虐トースト 3 種ローテーション、連打抑制 1 秒確定。`useOrder` mutation hook の onError 実装パターン。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`useOrder` 内で `onError` callback で全エラー処理（router.push / toast.show / setTimeout で 1 秒 disable）** | 1 hook に集約、再利用しやすい |
| B | **エラー処理を `errorMappers` ユーティリティに切り出し（`mapOrderError(err) → action`）+ hook で実行** | テスト容易、エラー追加時の影響範囲狭い |
| C | **B + Sentry / Datadog 連携の placeholder**（NFR-OBS-02 で実装しないが拡張ポイント明記） | 将来性、現状は単純化 |

[Answer]: **B（`errorMappers` ユーティリティに切り出し + hook で実行）**

追記事項:
- **設計パターン P-FE-ERR-01 Order Error Mapping**:
  - 目的: NFRC-C22 のエラー UX（402 遷移 / 500 自虐トースト / 連打抑制）を純関数化
  - 配置:
    - `web/lib/errorMappers.ts`: `mapOrderError(err) → OrderErrorAction` 純関数
    - `web/lib/toasts.ts`: `getRandomToast()` / `TOAST_VARIANTS`（Q-D10 で詳細確定）
    - `web/hooks/useOrder.ts`: `mapOrderError` の戻り値に応じたアクション実行
- **`OrderErrorAction` 型**:
  ```typescript
  type OrderErrorAction =
    | { type: 'navigate'; path: string; transitionMs?: number }
    | { type: 'toast'; text: string; durationMs?: number }
    | { type: 'silent' };
  ```
- **エラーマッピング**:
  | エラー | Action |
  |---|---|
  | `ApiError(402)` | `{ type: 'navigate', path: '/budget-empty', transitionMs: 0 }` |
  | `ApiError(409)` | `{ type: 'silent' }`（冪等性衝突は実質成功扱い、BR-C39） |
  | `ApiError(500)` | `{ type: 'toast', text: getRandomToast(), durationMs: 5000 }` |
  | `NetworkError` | `{ type: 'toast', text: getRandomToast(), durationMs: 5000 }` |
  | その他 | `{ type: 'toast', text: getRandomToast(), durationMs: 5000 }` |
- **`useOrder` 実装パターン**: `useMutation` の `onError` で `mapOrderError(err)` を呼び、戻り値の `action.type` で `router.push` / `showToast` / 何もしない を分岐実行。`retry: 0`（NFRC-C19、FD Q-12=A）、`onSuccess` で `['orderHistory']` / `['balance']` invalidate
- **論理コンポーネント**:
  - `LC-ORDER-OrderErrorMapper`: 純関数、`web/lib/errorMappers.ts`
  - `LC-ORDER-UseOrderHook`: React hook、`web/hooks/useOrder.ts`
- **テスト戦略**:
  - `errorMappers.test.ts`: 純関数として全エラーケース検証（Vitest）
  - `useOrder.test.tsx`: React Testing Library + msw で mock、`onError` callback の起動を検証
  - `BudgetEmptyScreen` への遷移は E2E テスト（dev 環境、Playwright 等、Build and Test ステージで詳細化）
- **NFR-OBS-02 整合の確認**: Sentry / Datadog 等の外部観測ツール連携は含めない（C 案除外根拠、placeholder も残さない）
- **Unit A `apiClient` (LC-AUTH-09) との連携**: `apiClient` の `ApiError` 型を `mapOrderError` で受ける、HTTP ステータスコードベースの分岐（`err.status` フィールド）

### Q-D10: 自虐トースト 3 種ローテーションの実装パターン

NFRC-C22 でトースト 3 種候補を確定（最終文言は Code Generation）。ランダム選択ロジックの配置と単純さ。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`lib/toasts.ts` に配列定義 + `getRandomToast()` 関数。`useOrder` の onError で 1 つ選んで `toast.show()`** | シンプル、テスト容易（`Math.random` mock） |
| B | **A + 直前と同じトーストを選ばない（連続避け）** | UX 向上、状態を `useRef` で保持 |
| C | **A + 連打時は順番に 1→2→3 をローテート**（時系列カウンタ） | 完全均等、過剰 |

[Answer]: **A（`lib/toasts.ts` 配列 + 純関数 `getRandomToast`）**

追記事項:
- **設計パターン P-FE-TOAST-01 Random Toast Variant**:
  - 目的: NFRC-C22 の自虐トースト 3 種ローテーションを純関数で実装、NFR-DEG-05 の「予測不能性によるダメ化エンタメ」を体現
  - 配置: `web/lib/toasts.ts`
  - 構造: `TOAST_VARIANTS` readonly 配列 + `getRandomToast()` 純関数（`Math.floor(Math.random() * TOAST_VARIANTS.length)`）
- **テスト戦略**:
  - `vi.spyOn(Math, 'random').mockReturnValue(value)` で決定論的にテスト
  - 全 3 文言が選択可能であることを検証
  - 配列が空でないこと / 文言が空文字でないことのバリデーション
- **論理コンポーネント**: `LC-ORDER-ToastVariants`（純関数 + 配列定義、`web/lib/toasts.ts`）
- **Q-D9=B との連携**: `mapOrderError` が `getRandomToast()` を呼んで `OrderErrorAction.text` に格納、純関数同士の組み合わせで責務分離
- **最終文言の確定**: 例示 3 文言は NFR Design 段階の参考、最終文言は Code Generation で確定（NFR-DEG-05 文言ガイドラインに従う）
- **連続避けロジック不採用の根拠**: NFRC-C22 の連打抑制 1 秒で連続表示はほぼ起きない、ランダム性そのものが NFR-DEG-05 のダメ化エンタメ体現
- **Q-D14 ToastHost との連携**: `ToastHost` コンポーネントが `useToast.showToast(text)` を提供、`getRandomToast()` の戻り値を `showToast()` に渡す

### Q-D11: TanStack Query `useOrderHistory` のローディング / 空状態パターン

NFRC-C19 で `staleTime: 60_000` / invalidate 戦略確定。MainScreen での `useOrderHistory` のローディング表示と空（履歴 0 件）の扱い。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`isLoading` 中はスケルトン、空（0 件）はサジェスト不出力（FD §6 整合）** | Unit B Q-D3 と同方針、UX 標準 |
| B | **A + `isFetching` 時のみ薄く（`opacity: 0.5`）表示**（`placeholderData` 活用） | 常に何かが見える、TanStack Query v5 標準パターン |
| C | **`isLoading` でも前回値を使い、`isError` 時のみフォールバック表示** | キャッシュ優先、初回は短時間スピナー |

[Answer]: **B（A + `placeholderData: keepPreviousData`）**

追記事項:
- **設計パターン P-FE-LOAD-01 Order History Loading State**:
  - 目的: NFR-DEG-03（履歴常時可視化）と NFRC-C19（60 秒 + invalidate）を `placeholderData` で両立
  - 配置: `web/hooks/useOrderHistory.ts`
  - 構成: `useQuery` の `placeholderData: keepPreviousData` + `staleTime: 60_000` + `gcTime: 300_000` + `refetchOnWindowFocus: true`
- **状態遷移**:
  | 状態 | `isLoading` | `isFetching` | `data` | UI |
  |---|---|---|---|---|
  | 初回ロード中 | true | true | undefined | スケルトン |
  | データ取得済 | false | false | OrderRecord[] | 通常表示 |
  | 再 fetch 中 | false | true | OrderRecord[] (前回値) | 薄く表示（opacity: 0.5） |
  | エラー時 | false | false | OrderRecord[] (前回値) | 通常表示 + エラーバナー重畳 |
  | 0 件 | false | false | [] | 「履歴がまだありません」ダメ化文言 |
- **空状態（0 件）の扱い**:
  - 「履歴がまだありません」ダメ化文言（NFR-DEG-05 体現、最終文言は Code Generation で確定）
  - サジェストカードは出さない（FD `business-logic-model.md` §6 整合、フォールバック分岐閾値 5 件未満）
  - 「ご飯めんどくさい」ボタンは常に表示
- **論理コンポーネント**:
  - `LC-ORDER-UseOrderHistoryHook`: TanStack Query hook、`web/hooks/useOrderHistory.ts`
  - `LC-ORDER-OrderHistoryList`: リスト表示、`web/components/OrderHistoryList.tsx`
  - `LC-ORDER-OrderHistorySkeleton`: スケルトン表示、`web/components/OrderHistorySkeleton.tsx`
- **Unit B Q-D3=A との差分根拠**:
  - Unit B `BalanceDisplay`: 数値 1 つ → スケルトン切替が自然
  - Unit C `OrderHistoryList`: リスト → `placeholderData` で常時可視化が UX 良
  - データ特性に応じた適材適所（NFRC-C19 stale time の判断と同じ思想）
- **テスト戦略**: Vitest + React Testing Library で各状態（isLoading / isFetching / isError / 空配列）を検証、`placeholderData: keepPreviousData` の挙動は msw で 2 回目以降の fetch を遅延させて検証
- **Unit A `apiClient` (LC-AUTH-09) / `BffProxyRouteHandler` (LC-AUTH-18) との整合**: ネットワークエラー時も `placeholderData` で前回データ表示維持、BFF 経由で 401 受信時は Unit A の自動 logout フローに委譲

### Q-D12: 連打抑制 1 秒の実装パターン（GoroButton）

NFRC-C22 で「ボタン押下後 / エラー応答後 1 秒間 disable」確定。Button の disable 制御。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`useOrder.isPending` を `disabled` 属性に直接バインド + 成功/エラー後 `setTimeout` で 1 秒 disable 維持** | シンプル、TanStack Query 標準 |
| B | **専用 `useDisableLock(durationMs)` hook を定義し、`isPending OR isLocked` で disable** | テスト容易、再利用可 |
| C | **submit イベントで `e.preventDefault()` + `pointerEvents: none` を 1 秒適用** | DOM レベル制御、過剰 |

[Answer]: **B（`useDisableLock(durationMs)` hook 切り出し）**

追記事項:
- **設計パターン P-FE-LOCK-01 Disable Lock Hook**:
  - 目的: NFRC-C22 のボタン連打抑制 1 秒を独立 hook で実装、テスト容易性・再利用性・a11y 対応を担保
  - 配置: `web/hooks/useDisableLock.ts`
  - API: `useDisableLock(durationMs: number) → { isLocked: boolean, triggerLock: () => void }`
- **状態管理**:
  - `lockedUntil: number`（lock 解除時刻 epoch ms）を `useState` で保持
  - `isLocked = Date.now() < lockedUntil` の派生値
  - `triggerLock()` で `lockedUntil = Date.now() + durationMs` を設定
- **メモリリーク対策**: `useEffect` cleanup で `clearTimeout` 呼出、unmount 時にタイマー破棄
- **`useOrder` での合成**:
  ```typescript
  export function useOrder() {
    const { isLocked, triggerLock } = useDisableLock(1000);
    const mutation = useMutation({
      mutationFn: placeOrder,
      retry: 0,  // NFRC-C19, FD Q-12=A
      onSuccess: () => {
        triggerLock();
        queryClient.invalidateQueries({ queryKey: ['orderHistory'] });
        queryClient.invalidateQueries({ queryKey: ['balance'] });
      },
      onError: (err) => {
        triggerLock();
        const action = mapOrderError(err);
        // ...
      },
    });
    return { ...mutation, disabled: mutation.isPending || isLocked };
  }
  ```
- **論理コンポーネント**:
  - `LC-ORDER-UseDisableLockHook`: 振る舞い hook、`web/hooks/useDisableLock.ts`
  - `LC-ORDER-GoroButton`: UI コンポーネント、`web/components/GoroButton.tsx`
- **テスト戦略**:
  - `useDisableLock.test.tsx`: `vi.useFakeTimers()` + `renderHook` + `act` で時間進行を制御、各境界（999ms / 1000ms / 1001ms）を検証
  - `useOrder.test.tsx`: `useDisableLock` を mock 化、`triggerLock` 呼出を検証
  - `GoroButton.test.tsx`: `disabled` 属性の変化を検証
- **a11y 観点**: HTML `disabled` 属性使用（スクリーンリーダー「無効」読み上げ、キーボード submit 不能）、`pointerEvents: none` 等の CSS ハック不採用（C 案除外根拠）
- **設計一貫性**: Q-D9=B / Q-D10=A / Q-D11=B と同様「責務単位のファイル分割 + 振る舞い hook 化」を継承
- **将来の拡張余地**: 他箇所（増額ボタン、履歴削除等）でも `useDisableLock` を再利用可能、localStorage 永続化は MVP 範囲外（placeholder も残さない）

### Q-D13: 観測性 / アラーム検知のためのログ付与パターン

NFRC-C13 のアラーム 3 種（p95 / Bedrock リトライ / フォールバック発動）を CloudWatch Logs メトリクスフィルタで検知。**ログにどのフィールドをいつ付与するか** の設計パターン。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`PlaceOrder` 完了時に 1 件のサマリログ（19 項目全付与）、エラー時のみ別途 ERROR ログ** | アラーム検知が容易、ログ件数最小 |
| B | **各ステップで INFO ログ（Bedrock 開始/完了、Wallet 開始/完了、...）+ サマリ** | デバッグ容易、ログ件数増 |
| C | **A + Bedrock リトライ発動時のみ追加 WARNING ログ**（`bedrockAttempt=2`） | アラーム検知に特化、必要最小限の追加ログ |

[Answer]: **C（A + Bedrock リトライ / フォールバック発動時のみ追加 WARNING ログ）**

追記事項:
- **設計パターン P-OBS-03 Layered Logging Strategy**:
  - 目的: NFRC-C13 アラーム 3 種を最小ログ件数で検知可能にする 3 層ログ戦略
  - 3 層構造:
    | 層 | レベル | 出力タイミング | 用途 |
    |---|---|---|---|
    | サマリ | INFO | `defer summary.LogComplete()` で 1 リクエスト終了時 | NFRC-C12 19 項目、最終集計、p95 検知 |
    | イベント WARNING | WARN | リトライ発動時 / フォールバック発動時 | NFRC-C13-2/C13-3 のシグナル |
    | エラー | ERROR | エラー戻り直前 | 異常終了の検知（CloudWatch Errors メトリクス） |
- **WARNING ログのスキーマ**:
  ```go
  // Bedrock リトライ発動時
  slog.WarnContext(ctx, "bedrock_retry",
      "event", "bedrock_retry",
      "attempt", 2,
      "errorClass", "ThrottlingException", // RetryClassifier 経由で分類
      "elapsedMs", 1500,
  )

  // フォールバック発動時
  slog.WarnContext(ctx, "fallback_triggered",
      "event", "fallback_triggered",
      "reason", "bedrock_double_failure", // or "permanent_error"
      "historyCount", 5,
      "fallbackType", "build_from_history", // or "default"
  )
  ```
- **CloudWatch Logs Insights クエリ例（NFRC-C13 アラーム実装参考、Infrastructure Design 引き継ぎ）**:
  ```
  # NFRC-C13-2 リトライ発動率
  fields @timestamp
  | filter level = "WARN" and event = "bedrock_retry"
  | stats count() by bin(5m)

  # NFRC-C13-3 フォールバック発動率
  fields @timestamp, fallbackType
  | filter level = "WARN" and event = "fallback_triggered"
  | stats count() by bin(5m)

  # NFRC-C13-1 p95 違反検知
  fields @timestamp, latencyMs
  | filter event = "place_order_complete"
  | stats percentile(latencyMs, 95) by bin(5m)
  ```
- **論理コンポーネント**:
  - `LC-ORDER-LogSummary`: Q-D6=B で確定、INFO サマリ専用
  - `LC-ORDER-EventLogger`: WARNING ログ出力、`internal/order/event_logger.go`（PlanBuilder / OrderService から呼出）
- **責務分離**:
  - `LogSummary`: 1 リクエストの状態を蓄積、完了時 INFO 出力
  - `EventLogger`: イベント駆動の WARNING 出力
  - 標準 `slog`: ERROR 出力（既存）
- **Q-D6=B / Q-D3=C との整合**:
  - サマリログは Q-D6=B の `LogSummary` で実装
  - WARNING ログは別経路、`LogSummary` には影響しない
  - レイテンシ計測（Q-D3=C の `measure` ヘルパ）はサマリログ用、WARNING 用ではない
- **Unit B との差分根拠**: Unit B は Bedrock 呼出なし、リトライ発動シグナル不要 → A で十分。Unit C は Bedrock 呼出あり、Comprehensive 深度 → C で運用シグナル強化

### Q-D14: ToastHost コンポーネントの配置と多重表示制御

NFRC-C22 のトースト表示。Next.js App Router で ToastHost をどこに配置し、多重表示（連打時）をどう制御するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **`app/layout.tsx` に `<ToastHost />` を配置（root レイアウト）、queue ベースで順次表示（最大 3 件キュー）** | グローバル、Unit A で既存があれば再利用 |
| B | **A + 同一 toast の多重表示防止（key で de-dup）** | 連打時のスパム防止 |
| C | **`react-hot-toast` 等の標準ライブラリ採用** | 実装コスト最小、外部依存追加 |

[Answer]: **A（`app/layout.tsx` 配置 + queue ベース順次表示、最大 3 件）**

追記事項:
- **設計パターン P-FE-TOAST-02 Toast Host & Queue**:
  - 目的: NFRC-C22 のトースト表示を Jotai + React で自前実装、外部依存なし、最大 3 件キューで多重表示制御
  - 配置:
    - `web/components/ToastHost.tsx`: トースト表示コンポーネント
    - `web/components/Toast.tsx`: 個別トースト（アニメーション含む）
    - `web/hooks/useToast.ts`: `showToast` API（Jotai atom ベース）
    - `web/state/toastAtoms.ts`: `toastsAtom`（Jotai atom 定義）
- **`app/layout.tsx` 配置**: `<ToastHost />` を `<body>` 直下、`<Providers>` 配下に配置（Jotai Provider 必須）
- **`useToast` API**:
  ```typescript
  type ToastItem = { id: string; text: string; durationMs: number };
  export function useToast() {
    const [toasts, setToasts] = useAtom(toastsAtom);
    const showToast = useCallback((text: string, durationMs = 5000) => {
      const id = crypto.randomUUID();
      setToasts((prev) => [...prev, { id, text, durationMs }]);
      setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), durationMs);
    }, [setToasts]);
    return { toasts, showToast };
  }
  ```
- **キュー制御**:
  - `toastsAtom` に追加された順に `ToastHost` が `slice(0, 3)` で先頭 3 件のみ表示
  - 4 件目以降は配列に保持、先頭が dismiss されると自動的に次が表示
  - 各トーストは個別 `setTimeout` で独立に消去
- **a11y**:
  - `role="region"` + `aria-live="polite"`（コンテナ）
  - `role="status"`（個別トースト）
  - 自動消去後もスクリーンリーダーが読み上げ済み
- **論理コンポーネント**:
  - `LC-ORDER-ToastHost`: トースト一覧表示、`web/components/ToastHost.tsx`
  - `LC-ORDER-Toast`: 個別トースト、`web/components/Toast.tsx`
  - `LC-ORDER-UseToastHook`: API hook、`web/hooks/useToast.ts`
  - `LC-ORDER-ToastsAtom`: Jotai atom、`web/state/toastAtoms.ts`
- **最大 3 件キューの根拠**: 表示時間 5 秒 × 1 秒連打抑制 = 5 秒間に最大 3 回エラー発生想定。それ以上はキューイング
- **B / C 不採用根拠**:
  - B（de-dup）: Q-D12=B 連打抑制で多重表示は構造的に防止、Q-D10=A ランダム性と矛盾
  - C（react-hot-toast）: 外部依存追加、MVP スコープ過剰、自前実装で十分
- **Unit A 既存 ToastHost との整合**: Unit A Code Generation 時に `ToastHost` 等が既に実装されている可能性あり、その場合は Unit C で再利用（横串化）。なければ Unit C で新規実装
- **テスト戦略**:
  - `useToast.test.tsx`: `showToast` で atom 更新、setTimeout 後に消去を検証（fakeTimers）
  - `ToastHost.test.tsx`: 4 件追加時に最大 3 件表示を検証
  - `Toast.test.tsx`: アニメーション、a11y 属性検証

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `aidlc-docs/construction/order/nfr-design/nfr-design-patterns.md` | Unit C 向け設計パターン集（リトライ判定 / フォールバック分岐 / レイテンシ計測 / SDK init 配置 / DI / ログ拡張 / PBT 設計 / Mock 設計 / Toast / Loading state / 連打抑制 / 観測性 / ToastHost） |
| `aidlc-docs/construction/order/nfr-design/logical-components.md` | Unit C の論理コンポーネント一覧（LC-ORDER-01〜N）。Backend (OrderService / OrderHandler / OrderHistoryRepository / BedrockAdapter / DeliveryAdapter / FallbackSuggestProvider / PlanBuilder / RetryClassifier / OrderLogger 等) / Frontend (GoroButton / OrderCompletionScreen / useOrder / useOrderHistory / lib/ulid.ts / lib/api/orders.ts / errorMappers / toasts / ToastHost 等)。Unit A 再利用コンポーネント（LC-AUTH-05 ContextAwareSlogHandler / LC-AUTH-09 apiClient / LC-AUTH-18 BffProxyRouteHandler）の利用箇所明記 |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

| 引き継ぎ先 | 内容 |
|---|---|
| **Infrastructure Design** | DynamoDB `OrderHistory` テーブル定義、Bedrock IAM Role / モデル ARN 制限、API Gateway ルート / Throttle / CORS、Lambda 関数 Terraform、CloudWatch Alarms / メトリクスフィルタ Terraform、AWS Budgets 設定 |
| **Code Generation** | Bedrock プロンプトテンプレート最終、リトライ・タイムアウト・冪等性の Go 実装、PBT のテストコード（`gopter`）、Bedrock SDK mock 実装、自虐トースト文言の最終確定、`useOrder` / `useOrderHistory` フック実装、`GoroButton` / `OrderCompletionScreen` / `ToastHost` 実装、`errorMappers` / `getRandomToast` 実装 |
| **Build and Test** | dev 環境での手動 E2E 検証手順、CI ワークフロー定義、Bedrock スタブのテスト戦略実装 |

---

## 6. unit-interfaces.md への追記検討

NFR Design 段階で凍結契約に追加が必要な可能性のある事項:
- TanStack Query key `['orderHistory']`（Unit C 独自、§9.1 拡張候補）
- 横串 Adapter (`BedrockAdapter` / `DeliveryAdapter` / `FallbackSuggestProvider`) の interface（既存 §6 横串契約に詳細追加）
- エラー型 (`order.ErrInsufficientFunds` / `order.ErrIdempotencyConflict`) の Frontend マッピング（§4.3 整合）

→ Q-D ヒアリング後、必要に応じて unit-interfaces.md を更新する。

---

## 7. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-D1 〜 Q-D14 を対話形式で順にヒアリング開始
