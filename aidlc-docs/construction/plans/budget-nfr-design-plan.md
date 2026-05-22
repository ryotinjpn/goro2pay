# Budget Unit — NFR Design Plan

**Document Version**: 0.1 (Draft, awaiting user answers)
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Construction Depth**: Standard
**Stage**: NFR Design (Construction Phase)
**Prerequisite**: NFR Requirements 承認済み (PR #74 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit B（予算・ウォレット）の **NFR Design ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。NFR Requirements で確定した数値・しきい値を **どう実現するか（パターン・論理コンポーネント）** を設計する。

### 1.1 NFR Requirements からの引き継ぎ事項

| 項目 | 値 | NFR Design で扱う論点 |
|---|---|---|
| Deduct p95 ≤ 500ms | 500ms | ConditionalCheckFailedException の判定・処理パターン |
| DynamoDB ConditionExpression | UpdateItem `balance >= :amount` | 残高不足 vs 競合の区別パターン、リトライ戦略 |
| 冪等性 TTL 24h | 24h | TryAcquire / SaveResponse の実装パターン |
| BudgetResetLog 冪等性 | PutItem with ConditionExpression | 多重リセット防止の実装パターン |
| TransactWriteItems 不使用 | 個別実行（Q-N5=B） | SetBudget 2 操作の部分失敗リカバリ方針 |
| 構造化ログ 9 項目 | Unit A 8 項目 + amount / newBalance | ライブラリ選定（slog vs zerolog）、出力タイミング |
| PBT 適用範囲 | 残高不変条件 / SetBudget べき等性 / バリデーション | フレームワーク選定（gopter vs rapid） |
| NFR-DEG-03 残高常時可視化 | TanStack Query 30s + 即 invalidate | Frontend loading state パターン |
| NFR-DEG-04 残高枯渇演出 | 402 受信時の増額誘導 | 演出 UI パターン |

### 1.2 NFR Design で扱うこと / 扱わないこと

| 扱う | 扱わない |
|---|---|
| 設計パターン（ConditionalCheck 判定 / Retry / Loading state / 演出） | パターンの具体実装コード（→ Code Generation） |
| 論理コンポーネント（WalletRepository / IdempotencyRepository / BalanceDisplay / InsufficientBalanceModal 等） | 論理コンポーネントの内部実装ロジック（→ Code Generation） |
| ログライブラリ・PBT フレームワークの確定 | ライブラリの設定ファイル / コード（→ Code Generation） |
| Frontend の残高鮮度パターン（invalidate 戦略） | TanStack Query の具体オプション値（→ Code Generation） |
| ResetAll のエラー収集パターン | Lambda の IAM Role 設定（→ Infrastructure Design） |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1問ずつ）
- [x] 回答の曖昧さを点検し、必要なら追加質問を挟む
- [x] `aidlc-docs/construction/budget/nfr-design/nfr-design-patterns.md` を作成
- [x] `aidlc-docs/construction/budget/nfr-design/logical-components.md` を作成
- [x] aidlc-state.md と audit.md を更新（承認後）
- [x] 完了メッセージを提示し、承認ゲートに進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-D1: `DeductConditional` の ConditionalCheckFailedException 処理パターン

`UpdateItem` の `ConditionExpression: balance >= :amount` が失敗した場合、Go SDK は `ConditionalCheckFailedException` を返す。これを「残高不足」として即座に `ErrInsufficientBalance` にマップするか、リトライを挟むか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **即マップ（リトライなし）** | `ConditionalCheckFailedException` → `ErrInsufficientBalance`。残高不足は再試行しても変わらないため即返す |
| B | **1 回リトライ後にマップ** | 競合書き込み（別リクエストが同時に残高を変更）でたまたま失敗した可能性に備えてリトライ |
| C | **指数バックオフ 2 回リトライ後にマップ** | 競合が多い環境向け。デモ規模では過剰 |

[Answer]:

### Q-D2: `SetBudget` 2 操作の部分失敗リカバリパターン

`SetBudget` は BudgetSettings の Upsert と Wallet の作成/差分調整の 2 操作を個別実行（TransactWriteItems 不使用）。1 操作目成功・2 操作目失敗の場合の方針。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **エラーをそのまま返す（リカバリなし）** | クライアントが再試行 → べき等設計のため 2 回目は一貫した状態に収束。FD の SetBudget べき等性（PBT P-2）が保証 |
| B | **補償トランザクション**（1 操作目をロールバック） | 実装複雑。DynamoDB の Delete/Put でロールバックするが競合が増える |
| C | **DLQ に失敗ペイロードを送り非同期リカバリ** | MVP 範囲外、過剰 |

[Answer]:

### Q-D3: 残高ローディング状態パターン（NFR-DEG-03 常時可視化）

`GetBalance` 取得中（TanStack Query の `isLoading` 状態）の `BalanceDisplay` の表示パターン。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **スケルトン UI**（グレーのプレースホルダ） | 画面レイアウトが安定、UX 標準的 |
| B | **前回値を薄く表示**（`isFetching` 時のみ薄くなる） | 常に何かが表示される、TanStack Query の `placeholderData` 活用 |
| C | **スピナー**（`BalanceDisplay` 全体をスピナーに置換） | シンプル実装だが残高エリアがちらつく |

[Answer]:

### Q-D4: 残高枯渇時の演出パターン（NFR-DEG-04）

`Deduct` が `ErrInsufficientBalance`（HTTP 402）を返した場合の Frontend 演出。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **インラインエラー + 増額ボタン表示**（モーダルなし） | 注文フロー内でそのまま増額 CTA を表示。タップ数最小 |
| B | **モーダル演出**（「残りダメ予算が足りません…」ダメ化文言 + 増額誘導ボタン） | NFR-DEG-05 のダメ化文言を強く演出できる |
| C | **画面遷移**（BudgetSetupScreen へリダイレクト） | コンテキスト切替が起きる、タップ数増加 |

[Answer]:

### Q-D5: 構造化ログライブラリの選定

NFR Requirements §7.1 で「NFR Design で確定」とした Go のログライブラリ。

| 案 | ライブラリ | 特徴 |
|---|---|---|
| A | **`log/slog`（Go 1.21 標準）** | 外部依存ゼロ。JSON ハンドラ組み込み。Unit A と揃えやすい |
| B | **`github.com/rs/zerolog`** | ゼロアロケーション設計で高速。サードパーティ依存あり |
| C | **`go.uber.org/zap`** | 高機能・高速。API が独特でやや複雑 |

[Answer]:

### Q-D6: PBT フレームワークの選定

NFR Requirements §8.1 で「NFR Design / Code Generation で確定」とした Go の PBT ライブラリ。

| 案 | ライブラリ | 特徴 |
|---|---|---|
| A | **`pgregory.net/rapid`** | シンプルな API。Go テストと自然に統合。スター数が少なめだが実用的 |
| B | **`github.com/leanovate/gopter`** | 機能豊富（Shrink サポート）。API がやや複雑 |
| C | **PBT フレームワークは使わず `testing/quick` のみ** | Go 標準。機能が限定的だが依存ゼロ |

[Answer]:

### Q-D7: `Deduct` 成功後の TanStack Query invalidate パターン

`Deduct` 成功（Unit C の注文完了）後に `GetBalance` を即 invalidate する実装パターン。invalidate は Unit C（注文フロー）側で行うか、Unit B の `BalanceDisplay` が自律的に検知するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **Unit C の注文完了 callback で `invalidateQueries(['balance'])`** | 注文フローが残高更新を責任持つ。Unit C との疎結合が若干低下 |
| B | **Jotai の `orderCompleted` atom を購読して `BalanceDisplay` が自律 invalidate** | Unit B と Unit C を atom で疎結合にできる |
| C | **`refetchInterval` で定期ポーリング**（30 秒ごと自動 refetch） | invalidate 不要でシンプル。Q-N10=A（stale 30s）と同等効果 |

[Answer]:

### Q-D8: `ResetAll` のエラー収集・ログ出力パターン

月初リセットで一部ユーザが失敗した場合、FD §4.4 では `errors []error` に収集して継続する設計。ログ出力の粒度。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **ユーザごとに ERROR ログ出力、最後に集計サマリも出力** | CloudWatch Logs でユーザ単位のエラーを追跡可能 |
| B | **全ユーザ処理後に失敗ユーザ一覧をまとめて 1 件 ERROR 出力** | ログ件数を抑えるが個別追跡が難しい |
| C | **A + 成功ユーザも INFO で出力（全件ログ）** | 監査には良いがデモ規模でもログが増える |

[Answer]:

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `nfr-design-patterns.md` | Unit B 向けの設計パターン集（ConditionalCheck 判定 / Retry / Loading state / 残高枯渇演出 / ログ / PBT） |
| `logical-components.md` | Unit B の論理コンポーネント一覧（WalletRepository / IdempotencyRepository / BudgetSettingsRepository / BudgetResetLogRepository / BalanceDisplay / InsufficientBalanceModal / useWallet フック等） |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

- DynamoDB テーブル物理設計（PK/SK/GSI 詳細）→ **Infrastructure Design**
- Lambda の IAM Role / DynamoDB テーブル名の環境変数 → **Infrastructure Design**
- EventBridge Scheduler の Terraform 実装 → **Infrastructure Design**
- ログライブラリの初期化コード・設定 → **Code Generation**
- PBT のテストコード実装 → **Code Generation**
- `InsufficientBalanceModal` の文言・アニメーション実装 → **Code Generation**

---

## 6. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-D1 〜 Q-D8 を対話形式で順にヒアリング開始
