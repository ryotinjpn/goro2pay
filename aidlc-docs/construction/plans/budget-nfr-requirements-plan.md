# Budget Unit — NFR Requirements Plan

**Document Version**: 0.1 (Draft, awaiting user answers)
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Construction Depth**: Standard
**Stage**: NFR Requirements (Construction Phase)
**Prerequisite**: Functional Design 承認済み (PR マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit B（予算・ウォレット）の **NFR Requirements ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。Plan 承認後、回答内容を反映した NFR 成果物を生成する。

### 1.1 Functional Design からの引き継ぎ事項

Functional Design ステージで「NFR Requirements で扱う」とされた論点:

- `GetBalance` の ConsistentRead コスト許容（Q-B6=A で方針確定済み、数値 SLA を明文化）
- `Deduct` の応答時間目標（残高引き落としは注文フローの中核、レイテンシ要件が重要）
- `ResetAll` の実行時間許容値（全ユーザを逐次処理、Lambda タイムアウトとの関係）
- 冪等性 TTL 24h の根拠を NFR として明文化
- DynamoDB ConditionExpression による残高不変条件の保証を NFR-REL-01 に紐付け
- Property-Based Testing の Unit B 向け適用範囲（Extension: Partial、Unit B は主要適用候補）

### 1.2 全体要件書 (`requirements.md`) で Unit B に関連する NFR

| NFR ID | 内容 | Unit B への影響 |
|---|---|---|
| NFR-DEG-01 | 1 代行手配あたり最大 2 タップ以内 | 残高確認・引き落としが高速である必要 |
| NFR-DEG-03 | 残高・消化率・今月のダメ化回数は常時可視化 | `GetBalance` の鮮度 / TanStack Query stale time に影響 |
| NFR-DEG-04 | 残高枯渇時、増額誘導を強めに提示 | 残高不足エラー（402）を即座に返す必要 |
| NFR-DEG-05 | UI コピーはダメ化文言 | `BalanceDisplay` / エラーメッセージの文言方針 |
| NFR-PERF-01 | ボタン押下→完了 3 秒以内 | `Deduct` は E2E の中核パス、低レイテンシ必須 |
| NFR-PERF-03 | 同時利用者数: 数人〜数十人 | DynamoDB スループット設計の前提 |
| NFR-REL-01 | 残高が負の値にならない | DynamoDB ConditionExpression で保証（Unit B 中核責務） |
| NFR-REL-02 | 二重引き落とし防止 | 冪等性キー設計（TTL・race condition 対策）が Unit B 中核責務 |
| NFR-REL-03 | 残高変動と注文履歴のアトミック実行 | TransactWriteItems の使用有無を明確化 |
| NFR-REL-04 | 月初リセット失敗時の手動リカバリ手順文書化 | `ResetAll` 失敗時の対応方針 |
| NFR-SCALE-01 | Lambda + DynamoDB オンデマンドの自動スケーリングに依存 | キャパシティモード確定 |
| NFR-OBS-01 | Lambda は CloudWatch Logs 構造化出力 | `Deduct` / `ResetAll` のログ項目設計 |
| NFR-COMP-01〜03 | 仮想ウォレット、個人情報最小、本番化時の別途評価 | 引用のみ |

### 1.3 NFR Requirements で扱うこと / 扱わないこと

| 扱う | 扱わない |
|---|---|
| パフォーマンス目標値（`Deduct` / `GetBalance` / `ResetAll` の応答時間） | DynamoDB テーブル設計の物理詳細（→ Infrastructure Design） |
| 可用性・信頼性方針の Unit B 向け明文化（冪等 TTL・リセット失敗対応） | Lambda の IAM Role / DynamoDB の暗号化設定（→ Infrastructure Design） |
| DynamoDB キャパシティモード確定 | TransactWriteItems の Go 実装コード（→ Code Generation） |
| EventBridge Scheduler 失敗通知の有無 | EventBridge の Terraform 詳細（→ Infrastructure Design） |
| ログ項目・構造化フォーマットの Unit B 向け確定 | CloudWatch ダッシュボード設計（→ NFR Design / Infrastructure Design） |
| Property-Based Testing の Unit B 向け適用範囲確定 | テストコードの実装（→ Code Generation） |
| Tech Stack の Unit B 向け確定（Go ライブラリ選定・DynamoDB SDK 等） | コードレベルの import 文（→ Code Generation） |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1問ずつ）
- [x] 回答の曖昧さを点検し、必要なら追加質問を挟む（Q-N2/Q-N4/Q-N5 の DynamoDB 一貫性リスクを確認、C=バースト容量前提で継続）
- [x] `aidlc-docs/construction/budget/nfr-requirements/nfr-requirements.md` を作成
- [x] `aidlc-docs/construction/budget/nfr-requirements/tech-stack-decisions.md` を作成
- [ ] aidlc-state.md と audit.md を更新（承認後）
- [ ] 完了メッセージを提示し、承認ゲートに進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-N1: `Deduct` の応答時間目標

`POST /api/wallet/deduct` は注文フロー（Unit C）から内部呼び出しされ、E2E 体感時間（NFR-PERF-01: 3 秒）の大部分を占める。Bedrock 推論（Unit C）で 1〜2 秒消費すると想定すると、Deduct に割り当てられる予算は限られる。

| 案 | P50 | P95 | 備考 |
|---|---|---|---|
| A | 50ms | 200ms | DynamoDB のみ操作、Lambda コールドスタート除き十分達成可能 |
| B | 100ms | 500ms | コールドスタート込みで現実的、デモ規模では問題なし |
| C | 目標値は定めず NFR-PERF-01 全体予算（3 秒）内で OK | — | Unit A Q-N3 と同方針 |

[Answer]: **Issue #25 に合わせて p95 ≤ 500ms**。当初 C（目標値なし）で確定したが、Issue #25 の要件「Wallet subtraction SLO (p95 ≤ 500ms)」に整合させるため変更。`GetBalance` 等その他 API は C（E2E 全体で管理）を維持。

### Q-N2: `GetBalance` の応答時間目標

`GET /api/wallet/balance` は画面初期表示（NFR-PERF-02: 5 秒以内）に影響する。Q-B6=A で ConsistentRead を採用しているため、Eventually Consistent より若干遅くなる可能性がある。

| 案 | P50 | P95 | 備考 |
|---|---|---|---|
| A | 50ms | 200ms | ConsistentRead でも DynamoDB の特性上十分達成可能 |
| B | 100ms | 500ms | コールドスタート込みで現実的 |
| C | 目標値は定めず NFR-PERF-01/02 全体予算内で OK | — | Unit A Q-N3 と同方針 |

[Answer]:

### Q-N3: `ResetAll` の実行時間許容値

月初 EventBridge Scheduler で呼び出される Lambda は、全ユーザを逐次処理する（Q-B8=A）。Lambda のデフォルトタイムアウトは 3 秒、最大 15 分。デモ規模（〜数十ユーザ）での実行時間を想定する。

| 案 | 想定ユーザ数 | 1 ユーザあたり処理時間 | 合計 | Lambda タイムアウト設定 |
|---|---|---|---|---|
| A | 最大 50 人 | 〜100ms | 〜5 秒 | 30 秒で十分 |
| B | 最大 50 人 | 〜100ms | 〜5 秒 | 余裕を持って 5 分（300 秒） |
| C | 最大 100 人 | 〜100ms | 〜10 秒 | 5 分（300 秒） |

[Answer]:

### Q-N4: DynamoDB キャパシティモード

Unit B は Wallet / BudgetSettings / IdempotencyRecord / BudgetResetLog の 4 テーブルを使用する。NFR-SCALE-01 はオンデマンドに依存すると明記しているが、テーブルごとに設定可能。

| 案 | 内容 | 備考 |
|---|---|---|
| A | **全テーブル オンデマンド** | デモ規模に最適、コスト予測が難しいが小規模では安価 |
| B | **Wallet / IdempotencyRecord はプロビジョンド（1 RCU/1 WCU）、他はオンデマンド** | コスト最適化だが管理が複雑 |
| C | **全テーブル プロビジョンド（1 RCU/1 WCU）** | 最低コスト（無料枠内）、スケールしないがデモ規模では問題なし |

[Answer]:

### Q-N5: `Deduct` における TransactWriteItems の使用有無

NFR-REL-03 に「残高変動と注文履歴の記録を可能な限りアトミックに実行」とある。Unit B の `Deduct` は `IdempotencyRecord` 作成と `Wallet.balance` 更新の 2 操作を行う。

| 案 | 内容 | トレードオフ |
|---|---|---|
| A | **TransactWriteItems で 2 操作をアトミックに実行** | 強い整合性。コストが 2 倍（1 WCU × 2）。冪等 Acquire と残高更新を 1 トランザクションで実行 |
| B | **2 操作を個別に実行（現在の FD 設計通り）** | FD の TryAcquire → DeductConditional の順序で実行。失敗時は冪等レコードが残る（FD で想定済み）。実装シンプル |

[Answer]:

### Q-N6: 月初リセット失敗時の通知・リカバリ方針

NFR-REL-04 に「月初リセット失敗時の手動リカバリ手順を文書化」とある。`ResetAll` で一部ユーザが失敗した場合の検知・対応方針。

| 案 | 内容 |
|---|---|
| A | **ログ出力のみ**（CloudWatch Logs に ERROR 記録、定期確認で気づく） |
| B | **EventBridge Scheduler の失敗を CloudWatch Alarm で検知** → SNS メール通知 |
| C | **A のみ + 手動リカバリ手順を Infrastructure Design ドキュメントに記載** |

[Answer]:

### Q-N7: ログ・観測性の項目（Unit B 向け）

NFR-OBS-01 に基づく構造化ログ項目。Unit A は `level / timestamp / userId / action / traceId / requestId / email_hash / userAgent` の 8 項目（Q-N7=B）。Unit B は金額・残高を扱うため追加項目を検討。

| 案 | 内容 |
|---|---|
| A | **Unit A と同じ 8 項目**（amount / balance 等の金額情報はログに出さない） |
| B | **Unit A 8 項目 + `amount`, `newBalance`, `idempotencyKey`（マスキングなし）** |
| C | **Unit A 8 項目 + `amount`, `newBalance` のみ（idempotencyKey は省略）** |

[Answer]:

### Q-N8: Property-Based Testing の Unit B 向け適用範囲

Extension: Partial（Unit A は メール正規化 / email_hash のみに適用）。Unit B は残高・予算のビジネスルールに多数の純粋関数があり、PBT の主要適用候補とされている。

Unit B で PBT を書くとすると以下のプロパティが候補:
- `balance` は常に `0 ≤ balance ≤ monthlyBudget`（差分調整ロジック）
- `SetBudget` のべき等性（同じ値で 2 回呼んでも結果が変わらない）
- `Deduct` の冪等性（同じ idempotencyKey × 同じ amount の再呼び出し結果が一致）
- 予算バリデーション（`1 ≤ monthlyBudget ≤ 100000` かつ `% 1000 == 0`）

| 案 | 内容 |
|---|---|
| A | **残高不変条件のみ**（`balance` の範囲、差分調整ロジックに絞る） |
| B | **残高不変条件 + SetBudget べき等性 + バリデーション境界値**（3 プロパティグループ） |
| C | **B + Deduct 冪等性**（冪等性 record の状態遷移を含む、工数大） |
| D | **PBT は適用しない**（通常のユニットテストで十分） |

[Answer]:

### Q-N9: Lambda メモリ設定

`Deduct` Lambda と `ResetAll` Lambda のメモリ設定。Lambda はメモリを増やすと CPU 割り当ても増え、コールドスタート時間も短縮される傾向がある。

| 案 | メモリ | 想定コスト（デモ規模） | 備考 |
|---|---|---|---|
| A | **128MB（デフォルト最小）** | 最低コスト | Go は軽量、128MB で十分動作する |
| B | **256MB** | 微増（無視できるレベル） | コールドスタートの改善を見込む |
| C | **512MB** | 倍増だがデモ規模では誤差 | 余裕を持たせたい場合 |

[Answer]:

### Q-N10: TanStack Query の stale time（`GetBalance`）

Q-B6=A で「TanStack Query 30s fresh + 自動 invalidate」を確定済み。NFR として明文化するにあたり、`Deduct` 成功後の invalidate 戦略を確認する。

| 案 | stale time | invalidate タイミング |
|---|---|---|
| A | **30 秒**（FD 確定済み）、`Deduct` 成功時に即 invalidate | ユーザが注文するたびに即座に残高が更新される |
| B | **30 秒**、invalidate は行わず stale time 経過を待つ | シンプルだが注文直後に古い残高が表示される可能性あり |
| C | **10 秒**（より鮮度重視）、`Deduct` 成功時に即 invalidate | 常時表示（NFR-DEG-03）を重視する場合 |

[Answer]:

### Q-N11: 監査・コンプライアンス

Unit A Q-N13=A と同様。NFR-COMP-01〜03 を引用するのみか、Unit B として追加事項があるか確認。

Unit B は「残高」という金額データを扱うが、実決済ではなく仮想ウォレット（数値のみ）のため、NFR-COMP-01 の「PCI DSS 直接対象外」が適用される。

| 案 | 内容 |
|---|---|
| A | **NFR-COMP-01〜03 を引用のみ**（Unit A と同方針、Unit B 固有の追記なし） |
| B | **A + 「仮想ウォレットは実決済データを含まない」旨を Unit B セクションに明記** |

[Answer]:

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `nfr-requirements.md` | Unit B 向け NFR の数値・しきい値・対応ポリシー一覧（性能 / 可用性 / 信頼性 / スケーラビリティ / 観測性 / テスト / コンプライアンス） |
| `tech-stack-decisions.md` | Unit B の Tech Stack 確定: Go ライブラリ（AWS SDK v2 / DynamoDB Document Client 等）、DynamoDB キャパシティモード、Lambda メモリ、TanStack Query 設定の上位概要 |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

- DynamoDB テーブル定義（GSI / LSI / TTL 設定）→ **Infrastructure Design**
- EventBridge Scheduler の Terraform 実装 → **Infrastructure Design**
- `Deduct` Lambda と `ResetAll` Lambda の IAM Role → **Infrastructure Design**
- 構造化ログの具体的フォーマット文字列・Go ライブラリ設定 → **NFR Design / Code Generation**
- PBT フレームワークのセットアップとテストコード → **Code Generation**
- `SetBudget` / `Deduct` の差分調整ロジックの具体実装 → **Code Generation**

---

## 6. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-N1 〜 Q-N11 を対話形式で順にヒアリング開始
