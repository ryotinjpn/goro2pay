# Order Unit (Unit C) — NFR Requirements Plan

**Document Version**: 0.1 (Draft, awaiting user approval)
**Created**: 2026-05-23
**Unit**: C (`order` / 代行手配コア)
**Construction Depth**: Comprehensive
**Stage**: NFR Requirements (Construction Phase)
**Prerequisite**: Functional Design 承認済み (PR #65 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit C（`order` / 代行手配コア）の **NFR Requirements ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。Unit C は MVP の心臓であり Comprehensive 深度のため、性能予算（3 秒）配分、Bedrock 関連メトリクス、Property-Based Testing（残高不変・履歴整合・冪等性）、CloudWatch アラーム条件など、他 Unit より踏み込んだ NFR を確定する。Plan 承認後、回答内容を反映した NFR 成果物を生成する。

### 1.1 Functional Design からの引き継ぎ事項

Unit C Functional Design (Q-1〜Q-12) で確定済み・NFR Requirements で数値・SLA として明文化する論点:

| FD 確定事項 | NFR Requirements で扱う観点 |
|---|---|
| Bedrock リトライ 1 回 (BR-C01) | リトライ回数・対象エラー分類を NFR-REL として明文化 |
| Bedrock 1 呼出 1.5 秒タイムアウト (BR-C02) | レイテンシ予算配分（Bedrock 占有時間 / 残予算で Wallet+Adapter+History を実行） |
| 親 Context 5 秒タイムアウト (BR-C02) | API Gateway 上限 29s との関係、E2E 3 秒予算との整合 |
| フォールバック分岐閾値 5 件 (BR-C06) | 履歴件数読取の応答時間目標 |
| `idempotencyKey` TTL 24h / userId 単位 (Q-7/8) | 冪等性保証 SLA、TTL 根拠を NFR-REL として明文化 |
| `OrderHistory` Insert 失敗時 200 応答 (Q-10, BR-C18) | NFR-REL-03（アトミック実行）からの逸脱を NFR で許容明文化 |
| `GetHistory` デフォルト 20 / 最大 100 (Q-11) | 応答時間目標、ページング戦略 |
| Frontend エラー UX (Q-12) | エラー応答時間 / リトライ抑制 / トースト表示時間 |

### 1.2 全体要件書 (`requirements.md`) で Unit C に関連する NFR

| NFR ID | 内容 | Unit C への影響 |
|---|---|---|
| **NFR-PERF-01** | ボタン押下→完了 **3 秒以内**（目標） | **Unit C が E2E オーケストレーション主体。本 NFR の主担当。** |
| NFR-PERF-02 | メイン画面初期表示 5 秒以内 | `GetHistory` の初回ロード時間に影響 |
| NFR-PERF-03 | 同時利用者数: 数人〜数十人 | Bedrock / DynamoDB スループット設計の前提 |
| NFR-DEG-01 | 1 代行手配あたり最大 2 タップ以内 | 「ご飯めんどくさい」ボタン → 完了画面の 1 タップ動線（再確認） |
| NFR-DEG-05 | UI コピーは自虐文言 | 完了画面 / エラートースト / フォールバック時の文言方針 |
| NFR-REL-01 | 残高が負にならない | Wallet 経由のため Unit B 主担当。Unit C は ErrInsufficientFunds → 402 ハンドリング |
| **NFR-REL-02** | 二重引き落とし防止（冪等性） | **Unit C 主担当（idempotencyKey の発行は Frontend Q-7=A、検証は Wallet）。Unit C は idempotent==true 応答整合を保証** |
| NFR-REL-03 | 残高変動と注文履歴のアトミック性 | Q-10=A により Unit C は **逸脱を許容**。NFR でその根拠を明記 |
| NFR-SEC-01 | JWT 検証必須 | Unit A AuthMiddleware に依存、Unit C 内で再検証しない |
| NFR-SCALE-01 | Lambda + DynamoDB の自動スケール | Bedrock 呼び出しは Throttle のリスクあり → アラーム条件で対応 |
| **NFR-OBS-01** | Lambda は CloudWatch Logs 構造化出力 | **Unit C は Bedrock latency / fallback 発動回数 / idempotent 命中率の可視化が中核** |
| NFR-OBS-02 | X-Ray・カスタムメトリクスは MVP では不実装 | カスタムメトリクスを **明示的に出さない** ことを NFR 確認 |
| NFR-COMP-01〜03 | 仮想ウォレット / 個人情報最小 / 本番化時の別途評価 | 引用のみ |

### 1.3 NFR Requirements で扱うこと / 扱わないこと

| 扱う | 扱わない |
|---|---|
| `PlaceOrder` E2E レイテンシ予算（3 秒の内訳配分） | Bedrock プロンプト本文・モデル選定の最終決定（→ Functional Design 完了済み・Tech Stack で確認のみ） |
| `GetHistory` の応答時間目標とページング戦略 | DynamoDB GSI の物理設計詳細（→ Infrastructure Design） |
| Bedrock リトライ回数・タイムアウト・スロットル時の振る舞い NFR 化 | リトライロジックの Go 実装（→ Code Generation） |
| 冪等性 TTL の根拠 / 同時送信時の race condition 観点 | 冪等レコード DynamoDB のキー設計詳細（→ 凍結契約 §3.4 で確定済み） |
| Property-Based Testing の Unit C 向け適用範囲（P-1/P-2/P-3） | PBT テストコード（→ Code Generation） |
| 構造化ログ項目の Unit C 向け確定（Bedrock 関連項目を含む） | 構造化ログの Go ライブラリ設定詳細（→ NFR Design / Code Generation） |
| CloudWatch メトリクスフィルタ / アラーム条件の方針 | CloudWatch Terraform 詳細（→ Infrastructure Design） |
| Bedrock コスト・スループットの月次予算観点 | Tech Stack の最終確定（Go ライブラリ・Bedrock SDK 等） |
| Frontend エラー UX の応答時間 / 自動遷移時間 | Next.js コンポーネント実装（→ Code Generation） |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1 問ずつ、`feedback_interview_style` に従う）
- [x] 回答の曖昧さ・矛盾を点検し、必要なら追加サブ質問を挟む（10 観点で矛盾チェック実施、整合確認）
- [x] `aidlc-docs/construction/order/nfr-requirements/nfr-requirements.md` を作成
- [x] `aidlc-docs/construction/order/nfr-requirements/tech-stack-decisions.md` を作成
- [x] aidlc-state.md を Unit C セクションを追加して更新
- [x] audit.md に NFR Requirements 経緯を追記
- [ ] 完了メッセージを提示し、承認ゲート（2-option）に進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-N1: `PlaceOrder` E2E レイテンシ予算配分（NFR-PERF-01 内訳）

NFR-PERF-01「ボタン押下→完了 3 秒以内」を Unit C E2E でどう配分するか。BR-C02 で Bedrock 1 呼出 1.5 秒、最悪リトライで 3.0 秒。Wallet.Deduct (Unit B Q-N1=p95 500ms)、Adapter Place、History Insert、Network/Frontend を含めて 3 秒に収める必要がある。

| 案 | Bedrock 占有想定 | 残予算配分 | 備考 |
|---|---|---|---|
| A | 通常時 1.5s（リトライなし） | Wallet 500ms + Adapter 200ms + History 200ms + Network/UI 600ms | NFR-PERF-01 を **通常時のみ目標、リトライ・フォールバック時は逸脱許容** |
| B | リトライ 2 回まで含めて p95 3.0s 以内 | Bedrock 最悪 3.0s でも全体 3.0s 内に押し込む | E2E 厳守、リトライ時はフォールバック判断を急ぐ運用 |
| C | 通常時 1.5s + フォールバック時 1.0s（履歴 1 件取得 + Default 配列処理） | フォールバック時のみ 2.5s 上限 | フォールバック発動時は時間予算をやや緩める |
| D | E2E 目標は p95 3.0s、p99 5.0s（API GW 上限内） | リトライ・フォールバック時の逸脱を p99 で吸収 | 統計的目標に切り替え |

[Answer]: **D（推奨案: p95 3.0s / p99 5.0s + ケース別サブターゲット）**

ケース別サブターゲット:

| パス | p95 目標 | 内訳（参考） |
|---|---|---|
| 通常パス（Bedrock 1 回成功） | ≤ 2.5s | Bedrock 1.5s + Wallet/Adapter/History/UI 1.0s |
| リトライ発動パス | ≤ 4.0s | Bedrock 3.0s + 残 1.0s（NFR-PERF-01 逸脱、p99 で吸収） |
| フォールバック発動パス | ≤ 2.5s | Bedrock 諦め → 履歴+Default 高速 |
| **E2E 全体** | **p95 ≤ 3.0s / p99 ≤ 5.0s** | NFR-PERF-01 を統計的に厳守、API GW 上限 29s に余裕 |

### Q-N2: `GetHistory` の応答時間目標とページング戦略

`GET /api/orders` は MainScreen のサジェスト判断（履歴 5 件以上か）と完了画面の履歴一覧表示で使われる。BR-C06 の閾値判定にも使う。Q-11=A で デフォルト 20 / 最大 100 / 降順 / TTL 90 日内が確定済み。

| 案 | P50 | P95 | invalidate 戦略 |
|---|---|---|---|
| A | 50ms | 200ms | DynamoDB Query (PK=userID, ORDER BY orderedAt DESC LIMIT N)、ConsistentRead 不要 |
| B | 100ms | 500ms | 同上、コールドスタート込みの現実的目標 |
| C | 目標値なし、NFR-PERF-02（メイン画面 5 秒）の全体予算内で OK | — | Unit A/B Q-N3 と同方針 |

[Answer]: **B（P50 100ms / P95 500ms、Unit B `GetBalance` と統一）**

追記事項:
- ページング: 1 ページあたり最大 100 件、cursor-based pagination は不採用（MVP は単純 LIMIT のみ）
- 90 日 TTL 超過分は DynamoDB が自動削除のため Query 結果に含まれない
- フォールバック分岐用（BR-C06）の高速 count 専用パスは設けない（500ms 内で十分）
- コールドスタート時のテール考慮で A（P95 200ms）ではなく B 採用

### Q-N3: Bedrock スロットリング時の振る舞い

NFR-SCALE-01 では Lambda + DynamoDB のオンデマンドスケーリングは保証されているが、**Bedrock は別系統**（モデルごとのアカウント TPS / RPM クォータあり）。BR-C01 でリトライは 1 回のみ。デモ中に Bedrock スロットリング（`ThrottlingException`）が発生した場合の振る舞いを NFR として明文化する。

| 案 | 内容 |
|---|---|
| A | **BR-C01 のリトライ 1 回 → フォールバック発動**（FD 設計通り、追加施策なし） |
| B | A + **CloudWatch メトリクスフィルタで `ThrottlingException` 検知 → アラーム** |
| C | A + B + **Bedrock の Provisioned Throughput 検討メモを Infrastructure Design に引き継ぐ**（実装はしない） |

[Answer]: **B（A + CloudWatch メトリクスフィルタで ThrottlingException 検知 → アラーム）**

追記事項:
- CloudWatch メトリクスフィルタはログから抽出のため NFR-OBS-02（カスタムメトリクス不実装）に違反しない
- フォールバック発動率も併せて監視（Q-N7 連動）
- 実装は Infrastructure Design / NFR Design ステージに引き継ぐ
- Bedrock Provisioned Throughput は本番化時の議論として除外（MVP スコープ外）

### Q-N4: 冪等性 TTL 24 時間の根拠と SLA

Q-7=A / Q-8=A で `idempotencyKey` は Frontend ULID 発行 / TTL 24 時間 / userId 単位確定済み。NFR-REL-02 にどう紐付けるか。

| 案 | 内容 |
|---|---|
| A | **24h は「ユーザがアプリを 1 日放置・再起動しても二重送信を防ぐ」根拠で固定**。NFR-REL-02 達成手段として明記、SLA は「TTL 内 100% 重複検知」 |
| B | **A + 24h 経過後の重複は許容**（極稀ケース、ビジネス影響軽微）と明記 |
| C | **TTL を 7 日に拡張**（ユーザの長期不在に対応） |

[Answer]: **B（A + 24h 経過後の重複は許容と明記）**

追記事項:
- NFR-REL-02 達成手段: idempotencyKey TTL 24h、userId 単位での重複検知
- 保証: TTL 内 100% 重複検知（同一 idempotencyKey は 1 回しか引き落としされない）
- 明示的逸脱: TTL 24h 超過後に同一 idempotencyKey で送信された場合、新規注文として扱う
- 逸脱の発生確率: 極低（Frontend は起動毎に新規 ULID 発行、凍結契約 §3.4 と整合）

### Q-N5: Property-Based Testing の Unit C 向け適用範囲

Extension: Partial。Unit A は メール正規化 / email_hash のみ、Unit B は **B（残高不変条件 + SetBudget べき等性 + バリデーション境界値）** で確定済み。Unit C は unit-of-work.md §3.3 で 3 プロパティが明記されている:

- **P-1 残高不変**: 同一 idempotencyKey の多重送信で残高は 1 回分しか減らない
- **P-2 履歴整合**: 全注文履歴の sum == 初期残高 - 現在残高 + リセット差分
- **P-3 冪等性レスポンス一貫性**: 同一 key の 2 回目以降は初回と同じ payload（OrderID / Amount / Store / Menu）

| 案 | 内容 | 工数 |
|---|---|---|
| A | **P-1 のみ**（残高不変、Unit B PBT で大半カバー済みのためスコープ外との解釈） | 小 |
| B | **P-1 + P-3**（冪等性レスポンス一貫性、Unit C 単体で完結する純粋関数で書きやすい） | 中 |
| C | **P-1 + P-2 + P-3 全て**（Comprehensive 深度の名にふさわしい全範囲） | 大 |
| D | **P-3 + フォールバック判定境界**（履歴 4/5 件の閾値分岐をプロパティ化） | 中 |
| E | **C + フォールバック判定境界**（最大限） | 特大 |

[Answer]: **B（P-1 + P-3）**

追記事項:
- **PBT 適用プロパティ**:
  - **P-1**: 同一 idempotencyKey の N 回送信に対し、残高変動は最大 1 回（Unit B PBT との連携、Unit C スコープでは `Deduct` をスタブ化）
  - **P-3**: 同一 idempotencyKey の N 回送信に対し、レスポンス payload (OrderID, Amount, StoreName, MenuName) は 2 回目以降完全一致
- **PBT 適用外**:
  - P-2（履歴整合）: Unit 境界をまたぐ統合検証のため E2E テスト or 観測（Q-N7 アラーム）で対応
  - フォールバック分岐閾値（履歴 4/5 件）: 境界値テーブルドリブンテストで対応（PBT 不採用、shrinking のメリットが薄い）
- **PBT フレームワーク**: Unit B と統一（`gopter` 想定、tech-stack-decisions.md で確定）

### Q-N6: 構造化ログ項目（Unit C 向け）

NFR-OBS-01 構造化ログ。Unit A は 8 項目、Unit B は **Unit A 8 項目 + amount / newBalance**（Q-N7=C）で確定済み。Unit C は Bedrock / Adapter / フォールバック / 冪等命中の 4 観点が中核。

Unit A 共通 8 項目: `level / timestamp / userId / action / traceId / requestId / email_hash / userAgent`

| 案 | 追加項目 |
|---|---|
| A | **Unit A 8 項目のみ**（Unit C 固有の追加なし） |
| B | A + `idempotencyKey` / `idempotent`（命中フラグ） / `orderId` |
| C | B + `bedrockLatencyMs` / `bedrockAttempt`（1 or 2） / `fallbackTriggered`（boolean） / `category` / `amount` |
| D | C + `storeName` / `menuName`（フォールバック時の店舗特定用） / `historyCount`（フォールバック分岐閾値の根拠） |
| E | D + `suggestionId`（サジェスト経由フラグ） |

[Answer]: **D（C + storeName / menuName / historyCount + source 派生）**

追記事項:
- **共通 8 項目**: `level / timestamp / userId / action / traceId / requestId / email_hash / userAgent`
- **Unit C 固有項目（11 項目）**:
  - 冪等性: `idempotencyKey`, `idempotent`, `orderId`
  - Bedrock 関連: `bedrockLatencyMs`, `bedrockAttempt`, `fallbackTriggered`
  - 注文内容: `category`, `amount`, `storeName`, `menuName`
  - 運用観察: `historyCount`, `source`（`"button"` | `"suggestion"`、E の suggestionId を低カーディナリティで代替）
- **マスキング**: なし（全項目 PII 非該当）
- **Bedrock プロンプト本文/レスポンス本文**: ログに含めない（Q-N12 で再確認）
- **Unit B との整合**: `amount` は Unit B Q-N7=C と統一、`idempotencyKey` は両 Unit 共通

### Q-N7: CloudWatch メトリクスフィルタとアラーム条件

NFR-OBS-02 で X-Ray・カスタムメトリクスは MVP 不実装と明記。ただし CloudWatch Logs Insights / メトリクスフィルタは **ログから抽出する形** なら追加コストなし。unit-of-work.md §3.3 では「CloudWatch アラーム条件（体感 3 秒超えの p95 検知等）」が Comprehensive スコープに含まれる。

| 案 | アラーム条件 |
|---|---|
| A | **アラームなし**（NFR-OBS-02 厳密準拠、本 MVP では不実装） |
| B | **`PlaceOrder` p95 > 3000ms を 5 分間に 3 回検知 → CloudWatch Alarm**（NFR-PERF-01 違反検知） |
| C | B + **Bedrock リトライ発動回数 / 5 分（Errors の傾向把握）** |
| D | B + C + **フォールバック発動回数 / 5 分**（フォールバック乱発はビジネスインパクト大） |
| E | D + **`ErrInsufficientFunds`（402）応答率 / 5 分**（ユーザ体験悪化指標） |

[Answer]: **D（B + C + フォールバック発動回数）**

追記事項:
- **アラーム実装方針**: CloudWatch Logs メトリクスフィルタ（カスタムメトリクスではない）+ CloudWatch Alarms + SNS Topic（メール通知）
- **アラーム 3 種と推奨閾値**:
  | アラーム | 閾値 | 評価期間 | 重要度 |
  |---|---|---|---|
  | NFR-PERF-01 違反検知（p95 > 3000ms） | 5 分間に 3 回 | 5 min × 3 datapoints | High |
  | Bedrock リトライ発動（`bedrockAttempt=2`） | 5 分間に 5 回以上 | 5 min × 1 datapoint | Medium |
  | フォールバック発動（`fallbackTriggered=true`） | 5 分間に 3 回以上 | 5 min × 1 datapoint | High |
- **責務分離**:
  - 残高枯渇率（402 応答率）は Unit B / E の責務（Unit C スコープ外、E 案除外の根拠）
  - 個別ユーザの追跡は不実装（NFR-OBS-02 / NFR-COMP-02 整合）
- **Terraform 実装**: Infrastructure Design ステージに引き継ぎ

### Q-N8: Bedrock モデル / コスト前提の確定

unit-of-work.md §3.3 では「Amazon Bedrock: Claude 系モデル」と概要のみ。NFR Requirements で **モデル ID と概算コスト方針** を確定する（最終 Tech Stack は tech-stack-decisions.md に記載）。

参考価格（ap-northeast-1, 2026-05 時点想定、概算）:
- Claude 3 Haiku: 入力 $0.25/Mtok, 出力 $1.25/Mtok（**最廉価**）
- Claude 3.5 Sonnet v2: 入力 $3.00/Mtok, 出力 $15.00/Mtok
- Claude Haiku 4.5: 入力 $0.80/Mtok, 出力 $4.00/Mtok（中間）
- Claude Opus 4.7: 入力 $15.00/Mtok, 出力 $75.00/Mtok（**高品質・高コスト**）

1 リクエストあたりトークン想定: 入力 500tok（履歴 + プロンプト）、出力 200tok（Plan JSON）

| 案 | モデル | 1 リクエストコスト概算 | 月次予算（10 ユーザ × 3 注文/日 × 30 日 = 900 req/月） |
|---|---|---|---|
| A | Claude 3 Haiku | $0.000375 | **$0.34/月** |
| B | Claude Haiku 4.5 | $0.0012 | **$1.08/月** |
| C | Claude 3.5 Sonnet v2 | $0.0045 | **$4.05/月** |
| D | Claude Opus 4.7 | $0.0225 | **$20.25/月**（無料枠超過注意） |

[Answer]: **B（Claude Haiku 4.5）**

追記事項:
- **モデル ID**: `jp.anthropic.claude-haiku-4-5-20251001-v1:0`（ap-northeast-1 inference profile）または `anthropic.claude-haiku-4-5-20251001-v1:0` 直接呼出
- **API**: Converse API（FD で確定済み）
- **想定トークン**: 入力 500tok / 出力 200tok / リクエスト
- **月次予算上限**: $10/月（負荷 10 倍 = 9,000 req/月想定）、超過時はモデル変更検討
- **コスト監視**: AWS Budgets で「Bedrock 月 $5 超過」アラート（Infrastructure Design で実装）
- **A 案（Claude 3 Haiku）を選ばない理由**:
  - ap-northeast-1 でのネイティブ提供がなく cross-region inference が必要 → レイテンシ +100〜200ms
  - Q-N1 の通常パス予算 2.5s が苦しくなる
  - 月額差は $0.74 と軽微（無料枠範囲内）
- **モデル変更余地**: 体感品質が悪い場合、Tech Stack 段階で C（Sonnet）に上げる検討余地を残す

### Q-N9: Lambda メモリ設定（API Lambda、Unit C 担当エンドポイント）

Unit B は 128MB 確定（Q-N9=A）。Unit C は Bedrock SDK + DynamoDB SDK + Adapter ロジックの 3 系統を扱うため、より重い可能性がある。`POST /api/orders` Lambda のメモリ設定。

| 案 | メモリ | 想定コールドスタート | 備考 |
|---|---|---|---|
| A | **128MB**（Unit B と統一） | やや遅い（Bedrock SDK 初期化込み） | コスト最低、デモ規模では機能動作する |
| B | **256MB** | 改善 | Bedrock SDK ウォームアップを考慮、無料枠内で十分余裕 |
| C | **512MB** | 顕著に改善 | 体感 3 秒予算を守りやすい、無料枠超過軽微 |
| D | **GET /api/orders は 128MB / POST /api/orders は 512MB**（エンドポイント別） | — | きめ細かい最適化、Lambda 関数が 2 つに分割される |

[Answer]: **B（256MB、全エンドポイント統一）**

追記事項:
- **Lambda メモリ**: 256MB（`POST /api/orders` / `GET /api/orders` 共通）
- **アーキテクチャ**: arm64（Graviton2、x86_64 より約 20% 安い）
- **コールドスタート目標**: P95 で 600ms 以内（Q-N1 の通常パス予算 1.0s 内）
- **Provisioned Concurrency**: 不採用（コスト高、デモ規模では頻度低）
- **タイムアウト**: 10 秒（BR-C02 の親 Context 5s + API GW 余裕）
- **Unit B との差分根拠**: Bedrock SDK の追加メモリ需要（Unit B は 128MB、Unit C は Bedrock 含むため +128MB）
- **D 案除外根拠**: Lambda 関数 2 分割は運用複雑化、Go ビルドで 1 バイナリ多エンドポイント対応可能

### Q-N10: TanStack Query 設定（`useOrder` / `useOrderHistory`）

Functional Design `frontend-components.md` で `useOrder` hook と `useOrderHistory` hook を定義済み。NFR として stale time / cacheTime を確定する。Unit B は GetBalance に対し **stale time 30 秒 + Deduct 成功時 invalidate**（Q-N10=A）で確定。

| 案 | useOrderHistory stale time | invalidate タイミング |
|---|---|---|
| A | **60 秒**、`PlaceOrder` 成功時に即 invalidate | 履歴は注文後にしか変わらないため stale time 長め |
| B | **30 秒**（Unit B GetBalance と統一）、`PlaceOrder` 成功時に即 invalidate | 統一感重視 |
| C | **stale: Infinity（手動 invalidate のみ）** | 最低限のリクエスト数、`PlaceOrder` 成功時のみ更新 |

`useOrder` は mutation のため stale time の概念なし、retry: 0（Q-12=A の意図に沿う）想定。

[Answer]: **A（60 秒 + PlaceOrder 成功時 invalidate）**

追記事項:
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

### Q-N11: フロントエンド エラー UX レイテンシ目標

Q-12=A で「402 → BudgetEmpty 遷移 / 500 → 自虐トースト」確定済み。NFR としてエラー応答時間と UI リトライ抑制を明文化。

| 案 | 内容 |
|---|---|
| A | **エラー応答も p95 3 秒以内**（成功と同じ予算） + トースト表示 5 秒 + 自動消去 + 連打抑制 1 秒 |
| B | A + **402 受信時の遷移は 0.3 秒以内**（即座に BudgetEmptyScreen へ） |
| C | A + B + **500 受信時の自虐トーストは 3 種ローテーション**（NFR-DEG-05 体現、UI 多様性） |
| D | A のみ（最低限） |

[Answer]: **C（A + B + 自虐トースト 3 種ローテーション）**

追記事項:
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

### Q-N12: 監査・コンプライアンス

Unit A Q-N13=A / Unit B Q-N11=B と同様の論点。Unit C は Bedrock を呼び出すため、**プロンプトに渡すデータの種類** が論点になりうる。

Unit C で Bedrock に渡すデータ: 直近履歴の店舗名・メニュー・カテゴリ・金額・曜日（**個人情報非該当**、メールアドレス・氏名は含まない）

| 案 | 内容 |
|---|---|
| A | **NFR-COMP-01〜03 を引用のみ**（Unit A と同方針） |
| B | **A + 「Bedrock に送信するデータには個人情報（氏名・メール）を含まない」旨を Unit C セクションに明記**（PII 漏洩リスクへの先回り対応） |
| C | **B + Bedrock リクエスト/レスポンスのログ記録方針**（CloudWatch Logs にプロンプト本文を残すか）も NFR 化 |

[Answer]: **C（B + Bedrock プロンプト/レスポンスのログ記録方針）**

追記事項:
- **NFR-COMP-01〜03 を引用**（Unit A / B と同方針）
- **Bedrock 入力データの PII 非該当性**:
  - Bedrock に送信するデータ: 直近 N 件の履歴（店舗名・メニュー・カテゴリ・金額・曜日）、固定プロンプトテンプレート、現在曜日
  - Bedrock に送信しないデータ: メールアドレス、氏名、`userId`、`email_hash`、その他 PII
  - 設計圧力: 将来機能追加時もこの境界を維持する
- **CloudWatch Logs への記録方針**:
  - **記録する**: Q-N6=D の 11 項目（technical metrics + 結果サマリ）
  - **記録しない**:
    - Bedrock リクエスト本文（プロンプト全文）
    - Bedrock レスポンス本文（提案 JSON 全文）
    - エラー時のスタックトレース内のプロンプト断片
  - **理由**: PII 混入リスクの予防的排除、CloudWatch Logs 閲覧権限経由の漏洩防止
- **本番化時の追加検討事項**（注記）:
  - PII Detection（Amazon Comprehend / Bedrock Guardrail）の導入検討
  - ログ保持期間の最適化（NFR Design / Infrastructure Design で確定）
  - 監査ログ（CloudTrail）と運用ログの分離

### Q-N13: テスト環境での Bedrock スタブ方針

Q-N5（PBT 範囲）と関連。実装フェーズで Bedrock 呼び出しを **テスト時にスタブ化** するかを NFR として方針確定する。

| 案 | 内容 |
|---|---|
| A | **テスト時は `BedrockAdapter` の interface mock**（go test 内で完結、外部呼び出しなし） |
| B | A + **CI でも Bedrock 実呼び出しは行わない**（コスト・遅延・スロットル考慮） |
| C | B + **dev 環境のみ Bedrock 実呼び出しを許容**（手動 E2E テスト時） |
| D | A + B + C のフル方針 |

[Answer]: **D（A + B + C のフル方針）**

追記事項:
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
  - PBT（Q-N5=B の P-1/P-3）は mock 経由で実行
- **PBT フレームワーク**: `gopter`（Unit B と統一）
- **CI コスト保護**: GitHub Actions に AWS シークレットを持たせない方針（Q-N12 ログ記録方針との整合）
- **dev 環境での手動 E2E 検証手順**（注記）: NFR Design / Build and Test ステージで詳細化

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `aidlc-docs/construction/order/nfr-requirements/nfr-requirements.md` | Unit C 向け NFR の数値・しきい値・対応ポリシー一覧（性能 / 可用性 / 信頼性 / スケーラビリティ / 観測性 / テスト / コンプライアンス）、E2E レイテンシ予算配分図、CloudWatch アラーム条件カタログ、Bedrock 関連 NFR |
| `aidlc-docs/construction/order/nfr-requirements/tech-stack-decisions.md` | Unit C の Tech Stack 確定: Bedrock モデル ID / SDK / Lambda メモリ / Go ライブラリ（DynamoDB Document Client / ulid / structured logging）/ TanStack Query 設定 / PBT フレームワーク選定方針 |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

- DynamoDB `OrderHistory` テーブル定義（GSI / TTL 設定、項目型）→ **Infrastructure Design**
- Bedrock IAM Role / モデル呼出許可ポリシー → **Infrastructure Design**
- API Gateway ルート定義 / Throttle / CORS → **Infrastructure Design**
- 構造化ログの Go ライブラリ設定（zap / slog 等）→ **NFR Design / Code Generation**
- Bedrock プロンプトテンプレート最終 → **Code Generation**（FD §2 で擬似コード確定済み）
- CloudWatch アラーム の Terraform 実装 → **Infrastructure Design**
- リトライ・タイムアウト・冪等性の Go 実装コード → **Code Generation**
- Property-Based Testing のテストコード → **Code Generation**

---

## 6. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-N1 〜 Q-N13 を対話形式で順にヒアリング開始
