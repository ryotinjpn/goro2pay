# Unit B `budget` — NFR Requirements

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Stage**: NFR Requirements (Construction Phase)
**Depth**: Standard
**Prerequisite**: Functional Design 承認済み

---

## 0. このドキュメントの目的

Unit B `budget` の非機能要件（NFR）を数値・しきい値・対応ポリシーとして定義する。実装手段（リトライ係数・ライブラリ選定等）は NFR Design / Infrastructure Design で扱う。

**設計前提**: 本 Unit はハッカソンのデモ用途（審査員 + チームメンバー数人〜数十人規模）を想定する。本番運用 SLA は定めない（NFR-AVAIL-02 と整合）。

---

## 1. ダメ化UX（最優先 NFR）

Unit B は「残高・消化率の常時可視化」と「残高枯渇時の増額誘導」を担い、ダメ化UX の中核を構成する。

| NFR ID | 要件 | Unit B での実現 |
|---|---|---|
| NFR-DEG-03 | 残高・消化率・今月のダメ化回数は常時可視化 | `GetBalance` の ConsistentRead + TanStack Query 30s stale time + `Deduct` 成功時の即 invalidate で鮮度を担保 |
| NFR-DEG-04 | 残高枯渇時、増額誘導を強めに提示 | `ErrInsufficientBalance`（HTTP 402）を即座に返却、Frontend が増額誘導 UI を表示 |
| NFR-DEG-05 | UI コピーはダメ化文言 | `BalanceDisplay` の「残りダメ予算」ラベル、`ResetCountdown` の「リセットまで〇日」等 |

---

## 2. パフォーマンス

### 2.1 API レイテンシ目標

| API | P95 目標 | 備考 |
|---|---|---|
| `POST /api/wallet/deduct`（ウォレット減算） | **500ms 以内** | Issue #25 要件。DynamoDB 操作のみで構成され、コールドスタート除き十分達成可能 |
| `GET /api/wallet/balance` | 目標値なし | E2E 全体（NFR-PERF-02: 5 秒以内）で管理 |
| その他 Unit B API | 目標値なし | E2E 全体（NFR-PERF-01: 3 秒以内）で管理 |

**根拠**: `Deduct` は注文フロー（Unit C）の中核パスであり、Bedrock 推論（1〜2 秒）とは独立して低レイテンシが求められるため個別 SLO を設定する。`GetBalance` はその他 API と同様に E2E で管理する（Q-N2=C）。

### 2.2 `ResetAll` 実行時間

| 項目 | 値 | 根拠 |
|---|---|---|
| 想定最大ユーザ数 | 50 人 | デモ規模（Q-N3=A） |
| 1 ユーザあたり処理時間 | 〜100ms | DynamoDB Get × 2 + Write × 2 の逐次処理 |
| 合計処理時間 | 〜5 秒 | 50 人 × 100ms |
| Lambda タイムアウト設定 | **30 秒** | 合計処理時間の 6 倍の余裕（Q-N3=A） |

### 2.3 フロントエンド鮮度

| 項目 | 値 | 根拠 |
|---|---|---|
| TanStack Query stale time | **30 秒** | Q-B6=A、Q-N10=A |
| `Deduct` 成功後の invalidate | **即時** | NFR-DEG-03 の常時可視化を保証（Q-N10=A） |
| ConsistentRead（`GetBalance`） | **有効** | Q-B6=A、最新残高を確実に返す |

**注意**: Wallet テーブルで ConsistentRead を使用するため 1 回の `GetBalance` で 2 RCU を消費する。プロビジョンドキャパシティ（1 RCU）の持続スループットを超えるが、月 1 回のリセット以外の大量同時アクセスはデモ規模では発生しないため、DynamoDB バースト容量（最大 300 RCU）で吸収する。本番化時には Wallet テーブルを 2 RCU 以上に増強すること。

### 2.4 同時利用者数

Unit A Q-N4=A と同方針（デモ最小: 同時 5 人、ピーク 10 req/s）。

---

## 3. 可用性

| NFR ID | 要件 |
|---|---|
| NFR-AVAIL-01 | AWS マネージドサービス（DynamoDB / Lambda / EventBridge Scheduler）の標準可用性に準拠。追加の冗長構成は実装しない |
| NFR-AVAIL-02 | 本番相当の SLA は定めない（デモ用途） |

---

## 4. 信頼性・データ整合性

| NFR ID | 要件 | Unit B での実現 |
|---|---|---|
| NFR-REL-01 | 残高が負の値にならない | `DeductConditional` の DynamoDB ConditionExpression: `balance >= :amount`（FD §4.3 不変条件） |
| NFR-REL-02 | 同一注文リクエストの二重処理（二重引き落とし）を防止 | 冪等性キー `{userID}:{ulid}` + TTL 24h + DynamoDB PutItem with ConditionExpression（FD §4.3） |
| NFR-REL-03 | 残高変動と注文履歴の整合性 | `IdempotencyRecord` 作成 → `DeductConditional` の順序で個別実行（TransactWriteItems は使用しない、Q-N5=B）。失敗時の冪等レコード保持（FD §2.3）で再実行時の一貫性を確保 |
| NFR-REL-04 | 月初リセット失敗時の手動リカバリ | ログ出力（CloudWatch Logs に ERROR 記録）+ 手動リカバリ手順を Infrastructure Design ドキュメントに記載（Q-N6=C）。自動リトライは実装しない |

### 4.1 冪等性 TTL

| 項目 | 値 | 根拠 |
|---|---|---|
| IdempotencyRecord TTL | **24 時間** | Q-B4=A。注文フローの重複リトライが 24 時間以内に収まる前提 |
| TTL 削除方式 | DynamoDB TTL 自動削除 | WCU 消費なし、追加実装不要 |

### 4.2 月初リセットの冪等性

`BudgetResetLog.(resetDate, userID)` 複合キーに条件付き Insert を使用し、同一ユーザへの重複リセットを防止（FD §4.4）。Lambda 再実行時も安全。

---

## 5. スケーラビリティ

| NFR ID | 要件 |
|---|---|
| NFR-SCALE-01 | Lambda + DynamoDB の自動スケーリング特性に依存 |
| NFR-SCALE-02 | 明示的なスケール設計は行わない（デモ用途） |

### 5.1 DynamoDB キャパシティモード

全テーブル **プロビジョンドキャパシティ（1 RCU / 1 WCU）** を採用する（Q-N4=C）。

| テーブル | RCU | WCU | 備考 |
|---|---|---|---|
| Wallet | 1 | 1 | ConsistentRead 使用のためバースト容量に依存（§2.3 注意参照） |
| BudgetSettings | 1 | 1 | SetBudget は低頻度 |
| IdempotencyRecord | 1 | 1 | Deduct ごとに 1 write、TTL 削除は WCU 不要 |
| BudgetResetLog | 1 | 1 | 月 1 回のリセット処理、バースト容量で吸収 |

**前提**: AWS 無料枠（25 RCU / 25 WCU / 月）内での運用を想定。デモ規模のアクセス量はバースト容量（最大 300 秒分）で吸収する。本番化時は実測値に基づきキャパシティを再設定すること。

### 5.2 Lambda スケーリング

| 項目 | 値 | 根拠 |
|---|---|---|
| Lambda メモリ | **128MB** | Go は軽量、DynamoDB 操作のみで十分（Q-N9=A） |
| Lambda タイムアウト（API） | **29 秒**（API Gateway 上限） | API Gateway のデフォルト上限に合わせる |
| Lambda タイムアウト（ResetAll） | **30 秒** | Q-N3=A |
| 同時実行数制限 | 設定しない（Lambda デフォルト） | デモ規模では不要 |

---

## 6. セキュリティ

Unit B 固有のセキュリティ要件（Security Extension は opt-out のため最低限のみ）。

| NFR ID | 要件 |
|---|---|
| NFR-SEC-01 | 全 API エンドポイントは API Gateway Cognito Authorizer で JWT 検証必須（Unit A が担う、Unit B は userID を context から受け取るのみ） |
| NFR-SEC-02 | DynamoDB データは AWS マネージドデフォルト暗号化を利用（KMS 顧客管理キーは使用しない） |

### 6.1 idempotencyKey のユーザ照合

`idempotencyKey` のプレフィックス（`userID:` の左側）が認証コンテキストの `userID` と一致することを Handler 層で検証する（VR-B-05）。他ユーザのキーを横取りできない。

---

## 7. 観測性

| NFR ID | 要件 |
|---|---|
| NFR-OBS-01 | Lambda は CloudWatch Logs に構造化 JSON ログを出力 |
| NFR-OBS-02 | 高度な観測性（カスタムメトリクス / ダッシュボード）は実装しない |

### 7.1 Unit B の構造化ログ項目

Unit A の 8 項目に Unit B 固有の金額フィールドを追加する（Q-N7=C）。

| フィールド | 型 | 説明 |
|---|---|---|
| `level` | string | `INFO` / `WARN` / `ERROR` |
| `timestamp` | string | ISO 8601 |
| `userId` | string | Cognito sub |
| `action` | string | 例: `deduct`, `set_budget`, `get_balance`, `reset_all` |
| `traceId` | string | X-Ray Trace ID（任意） |
| `requestId` | string | Lambda Request ID |
| `userAgent` | string | HTTP User-Agent ヘッダ |
| `amount` | int | `Deduct` 時の引き落とし額（円）。該当 action のみ出力 |
| `newBalance` | int | `Deduct` / `ResetAll` 後の残高（円）。該当 action のみ出力 |

**注**: `idempotencyKey` はログに出力しない（Q-N7=C）。`email_hash` は Unit B では扱わないため省略。

---

## 8. テスタビリティ

Extension: Partial（純粋関数とシリアライゼーションのみ）。Unit B は残高・予算ビジネスルールに純粋関数が多く、PBT の主要適用対象とする（Q-N8=B）。

### 8.1 Property-Based Testing 適用範囲

| プロパティグループ | 対象関数 | プロパティ例 |
|---|---|---|
| 残高不変条件 | `SetBudget` の差分調整ロジック | `0 ≤ newBalance ≤ newMonthlyBudget` が常に成立 |
| 残高不変条件 | `SetBudget` 減額時の打ち切り | `newBalance = min(oldBalance, newBudget)` |
| SetBudget べき等性 | `SetBudget(x); SetBudget(x)` | 同じ値で 2 回呼んでも状態が変わらない（delta=0） |
| バリデーション境界値 | `validateMonthlyBudget` | `1 ≤ x ≤ 100000` かつ `x % 1000 == 0` の境界を網羅 |

### 8.2 通常テスト

| テスト種別 | 対象 |
|---|---|
| ユニットテスト | ドメインサービス・バリデーション・エラーマッピング |
| 統合テスト | DynamoDB Local を使った `Deduct` / `SetBudget` / `ResetAll` の E2E |
| E2E テスト | Frontend → API Gateway → Lambda → DynamoDB の結合確認 |

---

## 9. アクセシビリティ・UX

| NFR ID | 要件 |
|---|---|
| NFR-A11Y-01 | 基本的なセマンティック HTML（`BalanceDisplay` / `BudgetForm` 等） |

ブラウザ互換性: Chrome / Edge / Safari 最新 2 バージョン（Unit A Q-N11=A と同方針）。

---

## 10. コンプライアンス

| NFR ID | 要件 |
|---|---|
| NFR-COMP-01 | 本アプリは**仮想ウォレット（数値管理のみ）**を採用し、実カード情報・決済情報を一切扱わない。Unit B が保持するのは整数値（残高・予算額）のみであり、PCI DSS の直接的対象外とする（Q-N11=B） |
| NFR-COMP-02 | 個人情報は「メールアドレス」のみ取得する（Unit A / Cognito が管理）。Unit B は userID（Cognito sub）のみを保持し、個人を特定できる情報は保存しない |
| NFR-COMP-03 | 本番運用時は、個人情報保護法・金融庁ガイドラインへの適合性評価が別途必要である旨を設計書に注記する |

---

## 11. 制約サマリ

| カテゴリ | 決定事項 | 根拠 |
|---|---|---|
| パフォーマンス | 単体 SLA なし、E2E で管理 | Q-N1=C, Q-N2=C |
| ResetAll タイムアウト | 30 秒 | Q-N3=A |
| DynamoDB | 全テーブル プロビジョンド 1 RCU/1 WCU | Q-N4=C |
| アトミック性 | TransactWriteItems 不使用（個別実行） | Q-N5=B |
| リセット失敗対応 | ログ + Infrastructure Design に手順記載 | Q-N6=C |
| ログ項目 | Unit A 8 項目 + amount + newBalance | Q-N7=C |
| PBT | 残高不変条件 + SetBudget べき等性 + バリデーション境界値 | Q-N8=B |
| Lambda メモリ | 128MB | Q-N9=A |
| TanStack Query | stale 30s + Deduct 後即 invalidate | Q-N10=A |
| コンプライアンス | NFR-COMP-01〜03 引用 + 仮想ウォレット明記 | Q-N11=B |
