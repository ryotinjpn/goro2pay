# Unit D (`suggest`) — NFR Requirements

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / NFR Requirements
**Unit**: D — `suggest`（学習・先回り）
**Depth**: Standard
**Related**: Plan [suggest-nfr-requirements-plan.md](../../plans/suggest-nfr-requirements-plan.md)、FD [suggest/functional-design/](../functional-design/)、継承元 [order/nfr-requirements/](../../order/nfr-requirements/)（NFRC-Cxx）、凍結契約 [unit-interfaces.md](../../interfaces/unit-interfaces.md)
**方針**: Bedrock 横串関連は Unit C NFR-R を継承（明示）。Unit D 固有の論点のみ新規確定。Q-ND1〜Q-ND8（全 A）に基づく。

---

## 1. NFR 識別子規則

- 形式: `NFRD-D{番号}`（Non-Functional Requirement / Construction / Unit D）
- 各要件に **由来（全体 NFR / Plan Q-ND / Unit C 継承）** と **検証方法** を明記

---

## 2. パフォーマンス要件

### NFRD-D01: GetSuggestion E2E レイテンシ
- **由来**: Q-ND1=A、全体 NFR-PERF-02
- **要件**: `GET /api/suggest` は **p95 ≤ 2.5s**（Bedrock 1 回成功時）、リトライ発動時最悪 ~3.0s。NFR-PERF-02（メイン画面初期表示 5 秒）の予算内に収める。単体ハード SLO は設けず CloudWatch Logs で監視。
- **内訳（参考）**: 履歴読取（OrderHistoryReader）~200ms + Bedrock 1.5s + 保存（PutItem）~100ms + Network/Lambda ~400ms
- **検証方法**: CloudWatch Logs Insights で `GET /api/suggest` レイテンシ p95 を集計

### NFRD-D02: ResolveSuggestion レイテンシ
- **由来**: 派生（Unit C の 3 秒注文パス内で呼ばれる）
- **要件**: `ResolveSuggestion` は DynamoDB GetItem 1 回のみ（Bedrock 再検証なし、BR-D11）。**p95 ≤ 100ms**。Unit C `PlaceOrder` の E2E 3 秒予算（NFRC-C01）にほぼ影響しない。
- **検証方法**: 統合テストでレイテンシ計測、Unit C 注文パスの p95 に内包

---

## 3. 信頼性・可用性要件

### NFRD-D03: Bedrock リトライ・タイムアウト（Unit C 継承）
- **由来**: Q-ND2=A、**Unit C NFRC-C06 / NFRC-C07 を継承**
- **要件**: `BedrockAdapter.InferSuggestion` は 1 回あたり `context.WithTimeout(1500ms)`、失敗時 1 リトライ（計 2 回、待機 0）。永続エラー（ValidationException 等）は即フォールバック。FD BR-D04 と一致。
- **検証方法**: mock で 4 シナリオ（成功 / Throttle 後成功 / 2 連続失敗 / 永続エラー）

### NFRD-D04: フォールバック戦略
- **由来**: FD BR-D05 / BR-D06、Q-ND2
- **要件**: Bedrock 2 連続失敗時 `FallbackSuggestProvider.BuildFromHistory(history)`。履歴は NFRD 前提（5 件以上）で最頻パターンを返せる。万一 nil なら `hasSuggestion=false`。`FallbackUsed=true` は内部ログのみ（API 非出力）。
- **検証方法**: 境界値テスト（履歴 5/6 件）+ mock 失敗注入

### NFRD-D05: Suggestion 一時保存の信頼性
- **由来**: FD BR-D08 / BR-D10、Q-ND3
- **要件**: `GoroPay_Suggestion` に TTL 30 分で保存。`ResolveSuggestion` は失効・不在時 `nil`（Unit C BR-C10 透過フォールバック）。TTL 削除は DynamoDB 機能（WCU 消費なし）。
- **検証方法**: 統合テストで保存→取得→TTL 失効（時刻 mock）

### NFRD-D06: Bedrock スロットリング時の振る舞い（Unit C 継承）
- **由来**: Q-ND2=A、**Unit C NFRC-C11 を継承**
- **要件**: `ThrottlingException` は NFRD-D03 のリトライ→フォールバックで吸収。CloudWatch Logs メトリクスフィルタで検知（NFRD-D10）。Bedrock Provisioned Throughput は MVP スコープ外。

### NFRD-D07: 可用性
- **由来**: 全体 NFR-AVAIL-01/02
- **要件**: AWS マネージドサービス標準可用性に準拠。本番 SLA なし（デモ用途）。追加冗長なし。サジェスト取得失敗時はカード非表示に丸めるのみで、メイン画面の主機能（注文）は阻害しない。

---

## 4. スケーラビリティ要件

### NFRD-D08: DynamoDB キャパシティ
- **由来**: Q-ND3=A、Unit B/C と統一
- **要件**: `GoroPay_Suggestion` は **プロビジョンド 1 RCU / 1 WCU**（無料枠前提、デモ規模はバースト容量で吸収）。GSI なし。
- **検証方法**: Infrastructure Design の Terraform tftest

### NFRD-D09: スケーリング前提（Unit C 継承）
- **由来**: **Unit C NFRC-C11 を継承**、全体 NFR-SCALE-01 / NFR-PERF-03
- **要件**: Lambda + DynamoDB のオンデマンドスケーリング依存。想定同時利用 数人〜数十人。Bedrock はアカウントクォータに依存（throttle 時 NFRD-D06）。

---

## 5. 観測性要件

### NFRD-D10: 観測性方針（Unit C 継承 = カスタムメトリクス不実装）
- **由来**: Q-ND4=A、**Unit C NFRC-C14 を継承**、全体 NFR-OBS-02
- **要件**:
  - X-Ray 不実装 / `PutMetricData` カスタムメトリクス不実装
  - CloudWatch Logs メトリクスフィルタのみ採用
  - Unit D 用アラーム（フォールバック率等）の要否は Infrastructure Design で判断（Unit C NFRC-C13-3 と同方式）

### NFRD-D11: 構造化ログ項目（NFR-OBS-01）
- **由来**: Q-ND4=A、Unit C NFRC-C12 と同方針
- **要件**: JSON 構造化ログ。共通 8 項目（`level/timestamp/userId/action/traceId/requestId/email_hash/userAgent`、Unit A 統一）+ suggest 固有項目:

  | 項目 | 内容 |
  |---|---|
  | `suggestionId` | 採番した ULID |
  | `hasSuggestion` | 提案有無 |
  | `fallbackUsed` | フォールバック発動フラグ |
  | `historyCount` | 履歴件数（履歴十分判定の根拠） |
  | `suppressedRecentOrder` | 直近注文抑制の発動フラグ |
  | `bedrockLatencyMs` | Bedrock 呼出レイテンシ |
  | `bedrockAttempt` | 試行回数（1/2） |

- **記録しない**: Bedrock プロンプト本文 / レスポンス本文（NFRD-D16）
- **イベント**: FD BR-D20 のイベント一覧（`suggest_requested` / `_insufficient_history` / `_suppressed_recent_order` / `_bedrock_fallback` / `_served` / `suggestion_expired_or_missing`）

---

## 6. テスト要件

### NFRD-D12: Property-Based Testing 適用範囲
- **由来**: Q-ND7=A、Extension Partial
- **要件**: 軽量適用。
  - backend（`gopter`）: `BuildFromHistory` の最頻判定の決定性（同入力→同出力、同点時は最新 OrderedAt 優先）
  - frontend（`fast-check`）: `GET /api/suggest` レスポンス JSON → `useSuggestion` props 変換のラウンドトリップ
- **適用外**: 主要分岐（履歴十分 / 抑制 / フォールバック）は通常ユニットテスト + 境界値テストで担保

### NFRD-D13: テスト環境別 Bedrock スタブ（Unit C 継承）
- **由来**: Q-ND2=A、**Unit C NFRC-C16 を継承**
- **要件**: `go test` / CI は mock 必須（`BedrockAdapter` interface の mock）、dev/stg/prd は実呼出（`IS_TEST` 分岐）。CI に AWS シークレットを持たせない。

### NFRD-D14: 統合テスト要件
- **由来**: 派生（NFRD-D03〜D06 検証）
- **要件**: 履歴十分→Bedrock成功 / 履歴不足→非表示 / 直近注文→抑制 / Bedrock失敗→BuildFromHistory / ResolveSuggestion 有効・失効 の各シナリオ（Bedrock は mock）

---

## 7. インフラ・リソース要件

### NFRD-D15: Lambda 設定（Unit C 継承）
- **由来**: Q-ND5=A、**Unit C NFRC-C18 を継承**
- **要件**: `GET /api/suggest` は共有 API Lambda 上（256MB / arm64 / タイムアウト 10s）。Bedrock SDK のメモリ需要のため 256MB が妥当（Unit C が設定済み）。Provisioned Concurrency なし。

### NFRD-D16: Bedrock モデル・コスト・PII（Unit C 継承）
- **由来**: Q-ND2 / Q-ND8=A、**Unit C NFRC-C20 / NFRC-C24 を継承**
- **要件**:
  - モデル `jp.anthropic.claude-haiku-4-5-20251001-v1:0`（ap-northeast-1 IP）/ Converse API
  - 月次予算 Unit C と共有（合算 $10/月、AWS Budgets は Unit C 設定済みに相乗り）。suggest は起動毎 1 回・入力~500tok/出力~200tok で増分小
  - Bedrock へ送るのは履歴要約（store/menu/category/amount/orderedAt）のみ。email/userId/PII は送らない。プロンプト/レスポンス本文はログ非記録

### NFRD-D17: TanStack Query 設定（Frontend）
- **由来**: Q-ND6=A、FD BR-D14
- **要件**: `useSuggestion` は `queryKey: ['suggestion']`、`staleTime` 実質無限（マウント時 1 回）、`refetchOnWindowFocus: false`、`retry: 0`。取得失敗は `hasSuggestion=false` 相当に丸めカード非表示。

---

## 8. セキュリティ・コンプライアンス要件

### NFRD-D18: 認証・認可（Unit C 継承）
- **由来**: **Unit C NFRC-C25 を継承**、全体 NFR-SEC-01
- **要件**: JWT 検証は Unit A `AttachUserID` middleware に委譲。`userId` は `gin.Context` から取得。`GET /api/suggest` は認証必須。

### NFRD-D19: コンプライアンス
- **由来**: 全体 NFR-COMP-01〜03、Unit C NFRC-C24 と同方針
- **要件**: NFR-COMP-01〜03 を引用。Unit D は個人特定情報を保存しない（`GoroPay_Suggestion` は userId + plan のみ、TTL 30 分で自動削除）。

---

## 9. ユーザビリティ要件（ダメ化UX）

### NFRD-D20: 先回り体験（NFR-DEG-02 主担当）
- **由来**: 全体 NFR-DEG-02 / NFR-DEG-05、US-X-02
- **要件**:
  - 起動時（マウント時）にサジェストを提示（NFR-DEG-02 フェーズ2 体現）
  - 「押す。」1 タップで Unit C 注文フローに合流（NFR-DEG-01）
  - Bedrock 失敗・suggestionId 失効はユーザに気づかせず透過（NFR-DEG-05）
- **検証方法**: E2E（履歴十分→SuggestBubble 表示→1 タップ注文）

---

## 10. NFR トレーサビリティ

### 10.1 全体 NFR との対応

| 全体 NFR | Unit D NFR | 主担当/継承 |
|---|---|---|
| NFR-PERF-02（メイン画面 5s） | NFRD-D01 | 補助（GetSuggestion） |
| NFR-DEG-02（起動時サジェスト） | NFRD-D20 | ★ Unit D 主担当 |
| NFR-DEG-05（透過コピー） | NFRD-D04 / D20 | 共通 |
| NFR-AVAIL-01/02 | NFRD-D07 | 共通前提 |
| NFR-SCALE-01 | NFRD-D09 | Unit C 継承 |
| NFR-OBS-01（構造化ログ） | NFRD-D11 | Unit C 方針踏襲 |
| NFR-OBS-02（高度観測性 不実装） | NFRD-D10 | ★ Unit C 継承 |
| NFR-SEC-01（JWT） | NFRD-D18 | Unit A 委譲 |
| NFR-COMP-01〜03 | NFRD-D19 | 共通 |

### 10.2 Unit C 継承マップ

| Unit D NFR | 継承元（Unit C） |
|---|---|
| NFRD-D03（リトライ/タイムアウト） | NFRC-C06 / C07 |
| NFRD-D06（スロットリング） | NFRC-C11 |
| NFRD-D10（観測性方針） | NFRC-C14 |
| NFRD-D13（テスト mock） | NFRC-C16 |
| NFRD-D15（Lambda 256MB/arm64） | NFRC-C18 |
| NFRD-D16（モデル/コスト/PII） | NFRC-C20 / C24 |
| NFRD-D18（認証委譲） | NFRC-C25 |

### 10.3 FD ルールとの対応

| FD ルール | Unit D NFR |
|---|---|
| BR-D04（Bedrock 1.5s×1） | NFRD-D03 |
| BR-D05/D06（フォールバック） | NFRD-D04 |
| BR-D08/D10（TTL30分/失効nil） | NFRD-D05 |
| BR-D14（マウント時1回） | NFRD-D17 |
| BR-D18（PII 非送出） | NFRD-D16 |
| BR-D19/D20（ログ） | NFRD-D11 |

---

## 11. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **NFR Design** | suggest のロガー/レイテンシ計測パターン（Unit C nfr-design-patterns P-RETRY-01 等を参照展開）、論理コンポーネント定義、PBT 具体プロパティ |
| **Infrastructure Design** | `GoroPay_Suggestion` テーブル定義（PK/TTL/1RCU1WCU）、`GET /api/suggest` ルート、Bedrock IAM（Unit C と共有 module 参照）、フォールバック率アラームの要否判断 |
| **Code Generation** | `SuggestService` / `SuggestionStore` 実装、`InferSuggestion` プロンプト、PBT（gopter/fast-check）、`useSuggestion` / `SuggestBubble` 実装 |

## 12. 文書管理

- **承認**: ユーザ承認待ち（per-unit ループ承認ゲート）
- **凍結契約への影響**: なし（NFR Requirements は内部仕様）
- **次ステージ**: NFR Design（Standard）
