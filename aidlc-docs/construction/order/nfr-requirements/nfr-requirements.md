# Unit C (`order`) — NFR Requirements

**Document Version**: 1.0
**Created**: 2026-05-23
**Stage**: Construction / NFR Requirements
**Unit**: C — `order` (代行手配コア / MVP の心臓)
**Depth**: Comprehensive
**Related**:
- Plan: [order-nfr-requirements-plan.md](../../plans/order-nfr-requirements-plan.md)
- Functional Design: [business-logic-model.md](../functional-design/business-logic-model.md), [business-rules.md](../functional-design/business-rules.md), [domain-entities.md](../functional-design/domain-entities.md), [frontend-components.md](../functional-design/frontend-components.md)
- 凍結契約: [unit-interfaces.md](../../interfaces/unit-interfaces.md)
- 全体要件書: [requirements.md](../../../inception/requirements/requirements.md)

本ドキュメントは Unit C `order` の **非機能要件（NFR）** を、Functional Design の Q-1〜Q-12 と本ステージ Q-N1〜Q-N13 の確定事項に基づき、数値・しきい値・対応ポリシーとして網羅する。Comprehensive 深度として全要件を ID 付き（`NFRC-Cxx`）で管理する。

---

## 1. NFR 識別子規則

- 形式: `NFRC-C{番号}` (Non-Functional Requirement / Construction / Unit C)
- 番号: 連番（カテゴリ無関係に通し番号、ID 安定化のため）
- 各要件に **由来（全体 NFR or Plan 質問）** と **検証方法** を明記

---

## 2. パフォーマンス要件

### NFRC-C01: `PlaceOrder` E2E レイテンシ予算（NFR-PERF-01 主担当）
- **由来**: Plan Q-N1 = D、全体 NFR-PERF-01
- **要件**:
  - **E2E 全体目標**: p95 ≤ 3.0 秒 / p99 ≤ 5.0 秒
  - API Gateway 上限 29 秒に対し p99 5.0s で十分余裕
- **ケース別サブターゲット**:

  | パス | p95 目標 | 内訳（参考） |
  |---|---|---|
  | 通常パス（Bedrock 1 回成功） | ≤ 2.5s | Bedrock 1.5s + Wallet 500ms + Adapter 200ms + History 200ms + Network/UI 100ms |
  | リトライ発動パス | ≤ 4.0s | Bedrock 3.0s + 残 1.0s（NFR-PERF-01 逸脱、p99 で吸収） |
  | フォールバック発動パス | ≤ 2.5s | Bedrock 諦め → 履歴+Default 高速処理 |

- **検証方法**: CloudWatch Logs Insights でレイテンシ p95/p99 を 5 分粒度で集計、NFRC-C13 アラームで違反検知
- **適用箇所**: `OrderService.PlaceOrder`、API Gateway → Lambda → Bedrock/DynamoDB

### NFRC-C02: `GetHistory` 応答時間目標
- **由来**: Plan Q-N2 = B
- **要件**:
  - **P50 ≤ 100ms / P95 ≤ 500ms**
  - Unit B `GetBalance` と統一（コールドスタート時のテール考慮）
- **ページング戦略**:
  - 1 ページあたり最大 100 件、デフォルト 20 件（FD Q-11=A）
  - cursor-based pagination は不採用（MVP は単純 LIMIT のみ）
  - 90 日 TTL 超過分は DynamoDB が自動削除のため Query 結果に含まれない
- **検証方法**: CloudWatch Logs Insights による p95 監視
- **適用箇所**: `OrderHistoryRepository.Query`、`GET /api/orders`

### NFRC-C03: 親 Context タイムアウト
- **由来**: FD BR-C02、Plan Q-N1
- **要件**: `OrderService.PlaceOrder` 全体に親タイムアウト 5 秒を設定
- **理由**: API Gateway 上限 29 秒 / 体感 3 秒予算超過時の非常停止、ゾンビ Lambda 実行防止
- **検証方法**: 統合テストで `context.WithTimeout(5*time.Second)` を確認

### NFRC-C04: コールドスタート目標
- **由来**: Plan Q-N9 = B
- **要件**: P95 で 600ms 以内（Q-N1 の通常パス予算 1.0s 内に収まる）
- **対応**:
  - Lambda メモリ 256MB（NFRC-C18）
  - arm64 アーキテクチャ採用（Graviton2、x86_64 比 約 20% 安い）
  - Provisioned Concurrency は不採用（コスト高、デモ規模では頻度低）
- **検証方法**: CloudWatch Logs の `INIT_DURATION` を集計

---

## 3. 信頼性・可用性要件

### NFRC-C05: 冪等性保証（NFR-REL-02 主担当）
- **由来**: 全体 NFR-REL-02、FD Q-7/Q-8、Plan Q-N4 = B
- **要件**:
  - `idempotencyKey` は Frontend ULID 発行（FD Q-7=A、凍結契約 §3.4）
  - TTL 24 時間 / userId 単位（FD Q-8=A）
  - **保証**: TTL 内 100% 重複検知（同一 idempotencyKey は 1 回しか引き落としされない）
  - **明示的逸脱**: TTL 24h 超過後に同一 idempotencyKey で送信された場合、新規注文として扱う
- **逸脱の発生確率**: 極低（Frontend は起動毎に新規 ULID 発行）
- **NFR-REL-02 達成手段**: idempotencyKey TTL 24h、userId 単位での重複検知（凍結契約 §3.4 と整合）
- **検証方法**: PBT P-3（NFRC-C20）により網羅的検証

### NFRC-C06: Bedrock リトライ戦略
- **由来**: FD BR-C01、Plan Q-N3 = B
- **要件**:
  - 最大 1 回リトライ（初回 + リトライ 1 回 = 計 2 回）
  - 対象エラー: `context.DeadlineExceeded` / `ThrottlingException` / `ServiceUnavailableException` / 任意のネットワークエラー全般
  - 非対象エラー: `ValidationException` / `AccessDeniedException` などの永続エラーは即フォールバック
  - リトライ間の待機時間ゼロ（exponential backoff なし）
- **検証方法**: 統合テストで mock を使い 4 シナリオ（成功 / Throttle 後成功 / 2 連続失敗 / 永続エラー）を検証

### NFRC-C07: Bedrock 1 呼出タイムアウト
- **由来**: FD BR-C02、Plan Q-N1
- **要件**: 各 Bedrock 呼び出しに `context.WithTimeout(parent, 1500ms)` を適用
- **検証方法**: 統合テストで 1.5 秒タイムアウトを確認

### NFRC-C08: フォールバック戦略
- **由来**: FD BR-C05/BR-C06、Plan Q-N1 (案 D サブターゲット)
- **要件**:
  - フォールバック発動条件: Bedrock 2 連続失敗 OR ATTEMPT 1 永続エラー
  - 分岐閾値: 直近 30 日履歴 5 件以上 → `BuildFromHistory` / 4 件以下 → `Default`（5 店舗ランダム選択）
- **検証方法**: 境界値テーブルドリブンテスト（履歴 0/4/5/6 件）

### NFRC-C09: `OrderHistory` Insert 失敗時の応答方針
- **由来**: FD Q-10、BR-C18
- **要件**: OrderHistory Insert 失敗時も 200/201 成功応答（ユーザ体験 > データ完全性）
- **NFR-REL-03 からの逸脱許容**: 残高変動と注文履歴のアトミック実行は MVP では保証しない（補償トランザクションは将来対応）
- **検証方法**: ログに `historyInsertFailed=true` を残し、CloudWatch Logs Insights で検知

### NFRC-C10: 親 Context キャンセルの即時伝播
- **由来**: FD BR-C04
- **要件**: クライアント切断 / API Gateway タイムアウト時、進行中の Bedrock / DeliveryAdapter / DynamoDB 呼び出しを `ctx.Err() == context.Canceled` で打ち切り
- **検証方法**: 統合テストで context cancel 伝播を確認

---

## 4. スケーラビリティ要件

### NFRC-C11: スケーリング前提
- **由来**: 全体 NFR-SCALE-01、NFR-PERF-03、Plan Q-N3
- **要件**:
  - Lambda + DynamoDB のオンデマンドスケーリングに依存
  - 想定同時利用者数: 数人〜数十人（NFR-PERF-03）
  - Bedrock は別系統（モデルごとのアカウント TPS / RPM クォータあり）
- **Bedrock スロットリング時の振る舞い**:
  - NFRC-C06 のリトライ 1 回 → フォールバック発動（FD 設計通り）
  - CloudWatch メトリクスフィルタで `ThrottlingException` 検知 → SNS 通知（NFRC-C13）
  - **MVP スコープ外**: Bedrock Provisioned Throughput は本番化時の議論（コスト約 $200/月のため除外）

---

## 5. 観測性要件

### NFRC-C12: 構造化ログ項目（NFR-OBS-01 主担当）
- **由来**: 全体 NFR-OBS-01、Plan Q-N6 = D
- **要件**: CloudWatch Logs に以下の構造化ログを出力（JSON 形式）
- **共通 8 項目**（Unit A と統一）:
  - `level / timestamp / userId / action / traceId / requestId / email_hash / userAgent`
- **Unit C 固有項目（11 項目）**:

  | カテゴリ | 項目 | 内容 |
  |---|---|---|
  | 冪等性 | `idempotencyKey` | リクエスト由来の ULID |
  | 冪等性 | `idempotent` | 命中フラグ (boolean) |
  | 冪等性 | `orderId` | 注文 ID (ULID) |
  | Bedrock | `bedrockLatencyMs` | Bedrock 呼出のレイテンシ (ms) |
  | Bedrock | `bedrockAttempt` | 試行回数 (1 or 2) |
  | Bedrock | `fallbackTriggered` | フォールバック発動フラグ (boolean) |
  | 注文内容 | `category` | "food" 等のカテゴリ |
  | 注文内容 | `amount` | 金額 (整数、円) |
  | 注文内容 | `storeName` | 店舗名 |
  | 注文内容 | `menuName` | メニュー名 |
  | 運用観察 | `historyCount` | フォールバック分岐閾値の根拠（履歴件数） |
  | 運用観察 | `source` | `"button"` \| `"suggestion"`（経由元） |

- **マスキング**: なし（全項目 PII 非該当）
- **記録しない項目**（NFRC-C24 で再確認）:
  - Bedrock リクエスト本文（プロンプト全文）
  - Bedrock レスポンス本文（提案 JSON 全文）
  - エラー時のスタックトレース内のプロンプト断片
- **Unit B との整合**: `amount` は Unit B Q-N7=C と統一、`idempotencyKey` は両 Unit 共通

### NFRC-C13: CloudWatch アラーム条件
- **由来**: Plan Q-N7 = D、unit-of-work.md §3.3 Comprehensive スコープ
- **アラーム実装方針**: CloudWatch Logs メトリクスフィルタ（カスタムメトリクスではない）+ CloudWatch Alarms + SNS Topic（メール通知）
- **NFR-OBS-02 整合**: メトリクスフィルタはログから抽出のため、PutMetricData による「カスタムメトリクス」には該当しない

| アラーム | 閾値 | 評価期間 | 重要度 |
|---|---|---|---|
| **NFRC-C13-1** NFR-PERF-01 違反検知 | `PlaceOrder` レイテンシ p95 > 3000ms を 5 分間に 3 回連続 | 5 min × 3 datapoints | **High** |
| **NFRC-C13-2** Bedrock 安定性監視 | `bedrockAttempt=2` のログ件数 / 5 分 ≥ 5 | 5 min × 1 datapoint | Medium |
| **NFRC-C13-3** ビジネス価値毀損検知 | `fallbackTriggered=true` のログ件数 / 5 分 ≥ 3 | 5 min × 1 datapoint | **High** |

- **責務分離**:
  - 残高枯渇率（402 応答率）は Unit B / E の責務（Unit C スコープ外）
  - 個別ユーザの追跡は不実装（NFR-OBS-02 / NFR-COMP-02 整合）
- **しきい値の根拠**:
  - p95 アラーム: NFR-PERF-01 = 3000ms と整合
  - リトライ 5 回 / 5 分: 想定負荷（数十 req/分）の 10〜20% 程度を異常とみなす
  - フォールバック 3 回 / 5 分: ビジネス影響大のため低めに設定
- **Terraform 実装**: Infrastructure Design ステージに引き継ぎ

### NFRC-C14: NFR-OBS-02 厳密準拠
- **由来**: 全体 NFR-OBS-02、Plan Q-N3 / Q-N7
- **要件**:
  - X-Ray は実装しない
  - PutMetricData によるカスタムメトリクスは実装しない
  - CloudWatch Logs メトリクスフィルタ（ログ抽出）のみ採用
- **理由**: MVP スコープでの観測性の最小化、コスト・複雑度の抑制

---

## 6. テスト要件

### NFRC-C15: Property-Based Testing（PBT）適用範囲
- **由来**: Extension: Partial、Plan Q-N5 = B、unit-of-work.md §3.3
- **適用プロパティ**:
  - **P-1 残高不変**: 同一 idempotencyKey の N 回送信に対し、残高変動は最大 1 回（Unit B PBT との連携、Unit C スコープでは `Deduct` を mock 化）
  - **P-3 冪等性レスポンス一貫性**: 同一 idempotencyKey の N 回送信に対し、レスポンス payload (OrderID, Amount, StoreName, MenuName) は 2 回目以降完全一致
- **適用外**:
  - **P-2 履歴整合**: Unit 境界をまたぐ統合検証のため E2E テスト or 観測（NFRC-C13）で対応
  - フォールバック分岐閾値（履歴 4/5 件）: 境界値テーブルドリブンテスト（履歴 0/4/5/6 件）で対応（PBT 不採用、shrinking のメリットが薄い）
- **PBT フレームワーク**: `gopter`（Unit B と統一、tech-stack-decisions.md NFRC-C30 で確定）
- **Unit B との責務分担**: 残高不変は Unit B PBT で大半カバー、Unit C は idempotencyKey 多重送信シナリオを薄く 1 プロパティで保証

### NFRC-C16: テスト環境別 Bedrock スタブ方針
- **由来**: Plan Q-N13 = D
- **テスト戦略マトリクス**:

  | 環境 | Bedrock 呼出し | スタブ手法 | 認証情報 |
  |---|---|---|---|
  | `go test`（local） | **mock 必須** | `internal/adapters/bedrock/` の interface に対する `mock_bedrock.go`（手書き or `gomock`） | 不要 |
  | CI（GitHub Actions） | **mock 必須** | 同上 | 不要 |
  | dev 環境（AWS） | **実呼出し許容** | `IS_TEST` 環境変数なし時に実 SDK | Lambda 実行ロール |
  | stg 環境 | **実呼出し** | 同上 | Lambda 実行ロール |
  | prd 環境 | **実呼出し** | 同上 | Lambda 実行ロール |

- **mock 設計指針**:
  - `BedrockAdapter` interface に対する mock implementation を `internal/adapters/bedrock/mock/` に配置
  - 標準応答（成功）/ ThrottlingException / Timeout / 永続エラー の 4 シナリオを mock で再現
  - PBT（NFRC-C15 の P-1/P-3）は mock 経由で実行
- **CI コスト保護**: GitHub Actions に AWS シークレットを持たせない方針（PII / NFRC-C24 ログ記録方針との整合）
- **dev 環境での手動 E2E 検証手順**: NFR Design / Build and Test ステージで詳細化

### NFRC-C17: 統合テスト要件
- **由来**: 派生（NFRC-C03/C06/C07/C09/C10 検証）
- **要件**: 以下のシナリオを統合テストで検証（Bedrock は mock 経由）
  - 通常パス（Bedrock 1 回成功）
  - リトライ成功パス（Throttle → 2 回目成功）
  - フォールバック発動パス（2 連続失敗）
  - 永続エラーパス（即フォールバック）
  - 冪等性命中（同一 idempotencyKey 2 回送信）
  - 連打パス（並行 N 回送信、idempotencyKey 衝突）
  - 残高不足（Unit B `Deduct` が `ErrInsufficientFunds` を返す）
  - 親 Context タイムアウト（5 秒超過）
  - クライアント切断（`ctx.Err() == context.Canceled`）

---

## 7. インフラ・リソース要件

### NFRC-C18: Lambda 設定（Unit C 担当エンドポイント）
- **由来**: Plan Q-N9 = B
- **要件**:
  - **メモリ**: 256MB（`POST /api/orders` / `GET /api/orders` 共通）
  - **アーキテクチャ**: arm64（Graviton2、x86_64 より約 20% 安い）
  - **タイムアウト**: 10 秒（NFRC-C03 の親 Context 5s + API Gateway 余裕）
  - **Provisioned Concurrency**: 不採用
- **Unit B との差分根拠**: Bedrock SDK の追加メモリ需要（Unit B は 128MB）

### NFRC-C19: TanStack Query 設定（Frontend）
- **由来**: Plan Q-N10 = A、FD frontend-components.md
- **`useOrderHistory`**:
  - `staleTime: 60_000`（60 秒）
  - `gcTime: 300_000`（5 分、デフォルト）
  - `refetchOnWindowFocus: true`（デフォルト、タブ切替時に stale なら refetch）
  - invalidate トリガー: `PlaceOrder` mutation の `onSuccess` で `queryClient.invalidateQueries({ queryKey: ['orderHistory'] })`
- **`useOrder`（mutation）**:
  - `retry: 0`（FD Q-12=A、500 受信時は自虐トースト 1 回のみ）
  - `onSuccess` で `['orderHistory']` と `['balance']`（unit-interfaces.md §9.1 共有 query key）を invalidate
- **クロスユニット query key 契約**: `['balance']` invalidate は unit-interfaces.md §9.1（Unit B との凍結契約）に従う、`['orderHistory']` は Unit C 独自 query key
- **Unit B との差分根拠**: 残高は他経路でも変動（30 秒）、履歴は自分の注文のみで変動（60 秒）。データ特性に応じた適材適所

### NFRC-C20: Bedrock モデル選定
- **由来**: Plan Q-N8 = B
- **要件**:
  - **モデル ID**: `jp.anthropic.claude-haiku-4-5-20251001-v1:0`（ap-northeast-1 inference profile）または `anthropic.claude-haiku-4-5-20251001-v1:0` 直接呼出
  - **API**: Converse API（FD で確定済み）
  - **想定トークン**: 入力 500tok / 出力 200tok / リクエスト
  - **月次予算上限**: $10/月（負荷 10 倍 = 9,000 req/月想定）、超過時はモデル変更検討
  - **コスト監視**: AWS Budgets で「Bedrock 月 $5 超過」アラート（Infrastructure Design で実装）
- **モデル変更余地**: 体感品質が悪い場合、Tech Stack 段階で C（Sonnet）に上げる検討余地を残す
- **Claude 3 Haiku を選ばない理由**:
  - ap-northeast-1 でのネイティブ提供がなく cross-region inference が必要 → レイテンシ +100〜200ms
  - NFRC-C01 通常パス予算 2.5s が苦しくなる
  - 月額差は $0.74 と軽微（無料枠範囲内）

---

## 8. ユーザビリティ要件（ダメ化UX）

### NFRC-C21: ボタン押下動線（NFR-DEG-01 主担当）
- **由来**: 全体 NFR-DEG-01、FD Q-3
- **要件**:
  - 「ご飯めんどくさい」ボタン押下から完了画面表示まで 1 タップで完結
  - 完了画面は 5 秒後に自動でメイン画面に戻る（US-1-07）
- **検証方法**: E2E テストでタップ数とフロー時間を計測

### NFRC-C22: フロントエンド エラー UX
- **由来**: Plan Q-N11 = C、FD Q-12
- **要件**:
  - **エラー応答時間**: p95 3 秒以内（成功と同じ予算）
  - **トースト表示**:
    - 表示時間: 5 秒（自動消去）
    - 表示位置: 画面下部中央（モバイル想定）
    - **3 種ローテーション**: 配列定義 → `Math.floor(Math.random() * 3)` でランダム選択（NFR-DEG-05 体現）
  - **連打抑制**: ボタン押下後 / エラー応答後 1 秒間は再送信ボタン disable
  - **402 受信時挙動**:
    - **0.3 秒以内に `BudgetEmptyScreen` へ自動遷移**（NFR-DEG-04 退化ループの起点）
    - トースト表示なし（画面遷移自体がエラー説明）
  - **500 受信時挙動**:
    - 自虐トースト 1 回表示
    - 画面遷移なし（ユーザが再ボタン押下できる状態を維持）
    - リトライ自動化なし（FD Q-12=A、`useOrder.retry: 0`）
- **自虐トースト文言例**（最終文言は Code Generation で確定）:
  1. 「サーバーがやる気を失いました…もう一度お試しください」
  2. 「システムがふぬけてます。少し待ってあげてください」
  3. 「今日はちょっとダメ化に失敗しました。再挑戦しますか？」

### NFRC-C23: ダメ化UX 全体方針
- **由来**: 全体 NFR-DEG-01〜05、unit-of-work.md §3.3
- **要件**:
  - ボタン押下から完了画面表示まで **3 秒以内**（NFR-PERF-01 / NFR-DEG-01）
  - 操作タップ数は 1 回のみ（NFR-DEG-01）
  - Bedrock 失敗時もフォールバックで即時解決体験を維持（NFR-DEG-05）

---

## 9. セキュリティ・コンプライアンス要件

### NFRC-C24: 監査・コンプライアンス（Bedrock PII 観点）
- **由来**: 全体 NFR-COMP-01〜03、Plan Q-N12 = C
- **NFR-COMP-01〜03 を引用**（Unit A / B と同方針）
- **Bedrock 入力データの PII 非該当性**:
  - Bedrock に送信するデータ: 直近 N 件の履歴（店舗名・メニュー・カテゴリ・金額・曜日）、固定プロンプトテンプレート、現在曜日
  - Bedrock に送信しないデータ: メールアドレス、氏名、`userId`、`email_hash`、その他 PII
  - 設計圧力: 将来機能追加時もこの境界を維持する
- **CloudWatch Logs への記録方針**:
  - **記録する**: NFRC-C12 の 11 項目（technical metrics + 結果サマリ）
  - **記録しない**:
    - Bedrock リクエスト本文（プロンプト全文）
    - Bedrock レスポンス本文（提案 JSON 全文）
    - エラー時のスタックトレース内のプロンプト断片
  - **理由**: PII 混入リスクの予防的排除、CloudWatch Logs 閲覧権限経由の漏洩防止
- **本番化時の追加検討事項**（注記）:
  - PII Detection（Amazon Comprehend / Bedrock Guardrail）の導入検討
  - ログ保持期間の最適化（NFR Design / Infrastructure Design で確定）
  - 監査ログ（CloudTrail）と運用ログの分離

### NFRC-C25: 認証・認可
- **由来**: 全体 NFR-SEC-01、unit-interfaces.md §1.1
- **要件**:
  - JWT 検証は Unit A `AttachUserID` middleware に委譲（Unit C 内で再検証しない）
  - `userId` は `gin.Context` 経由で取得（凍結契約 §1.1 の前提）
- **検証方法**: 統合テストで認証なしリクエストが 401 を返すことを確認

---

## 10. NFR トレーサビリティ

### 10.1 全体 NFR との対応

| 全体 NFR ID | Unit C NFR | 主担当 |
|---|---|---|
| NFR-PERF-01 (体感 3 秒) | NFRC-C01 | ★ Unit C 主担当 |
| NFR-PERF-02 (メイン画面 5 秒) | NFRC-C02（GetHistory） | 補助 |
| NFR-PERF-03 (同時利用 数十人) | NFRC-C11 | 共通前提 |
| NFR-DEG-01 (1 タップ) | NFRC-C21 / NFRC-C23 | ★ Unit C 主担当 |
| NFR-DEG-04 (枯渇時誘導) | NFRC-C22（402 → BudgetEmpty） | 起点担当 |
| NFR-DEG-05 (自虐コピー) | NFRC-C22（自虐トースト 3 種） | ★ Unit C 主担当 |
| NFR-REL-01 (残高負化なし) | Unit B 主担当（Unit C は ErrInsufficientFunds → 402 ハンドリング） | Unit B 委譲 |
| NFR-REL-02 (二重防止) | NFRC-C05 | ★ Unit C 主担当 |
| NFR-REL-03 (アトミック性) | NFRC-C09（明示的逸脱許容） | Unit C 担当 |
| NFR-SEC-01 (JWT) | NFRC-C25 | Unit A 委譲 |
| NFR-SEC-02 (デフォルト暗号化) | Infrastructure Design 引き継ぎ | Infra |
| NFR-SCALE-01 (オンデマンド) | NFRC-C11 | 共通前提 |
| NFR-OBS-01 (構造化ログ) | NFRC-C12 | ★ Unit C 主担当 |
| NFR-OBS-02 (高度な観測性 不実装) | NFRC-C14 | 共通前提 |
| NFR-COMP-01〜03 | NFRC-C24（引用 + Bedrock PII 観点） | 共通 |

### 10.2 FD ルールとの対応

| FD ルール | Unit C NFR |
|---|---|
| BR-C01 (Bedrock リトライ 1 回) | NFRC-C06 |
| BR-C02 (Bedrock タイムアウト 1.5s) | NFRC-C07 |
| BR-C04 (親 Context キャンセル即時伝播) | NFRC-C10 |
| BR-C05/C06 (フォールバック条件・閾値) | NFRC-C08 |
| BR-C13 (idempotencyKey PK = key) | NFRC-C05（凍結契約 §3.4 整合） |
| BR-C18 (History Insert 失敗時 200) | NFRC-C09 |
| BR-C20/C22 (TTL 24h, expiresAt) | NFRC-C05 |
| BR-C39 (409 IDEMPOTENCY_CONFLICT) | NFRC-C05 |

---

## 11. 想定外の論点（後続ステージへの引き継ぎ）

| 引き継ぎ先 | 内容 |
|---|---|
| **NFR Design** | 構造化ログの Go ライブラリ設定（zap / slog 等）、ロガー初期化パターン、レイテンシ計測のミドルウェア設計、PBT の具体プロパティ実装方針 |
| **Infrastructure Design** | DynamoDB `OrderHistory` テーブル定義（GSI / TTL 設定）、Bedrock IAM Role / モデル呼出許可ポリシー、API Gateway ルート定義 / Throttle / CORS、Lambda 関数定義（256MB / arm64 / 10s）、CloudWatch アラーム の Terraform 実装、AWS Budgets 設定 |
| **Code Generation** | Bedrock プロンプトテンプレート最終、リトライ・タイムアウト・冪等性の Go 実装、PBT のテストコード（`gopter`）、Bedrock SDK mock 実装、自虐トースト文言の最終確定、Frontend `useOrder` / `useOrderHistory` フック実装 |
| **Build and Test** | dev 環境での手動 E2E 検証手順、CI ワークフロー定義、Bedrock スタブのテスト戦略実装 |

---

## 12. 文書管理

- **承認**: ユーザ承認待ち（Construction フェーズの per-unit ループ承認ゲート）
- **凍結契約への影響**: なし（NFR Requirements は内部仕様、凍結契約は変更不要）
- **次ステージ**: NFR Design（per-unit、Comprehensive 深度）
