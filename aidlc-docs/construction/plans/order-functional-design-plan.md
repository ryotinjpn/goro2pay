# Unit C (`order`) — Functional Design Plan

**Document Version**: 1.0
**Created**: 2026-05-21
**Stage**: Construction / Functional Design
**Unit**: C — `order` (代行手配コア)
**Depth**: **Comprehensive**（本 MVP の心臓 / 唯一の Comprehensive Unit）
**Prerequisite**: Inception 全成果物承認済み

---

## 1. Plan の目的

Unit C `order` の Functional Design ステージで生成する成果物の方針を確定するためのヒアリング Plan。Comprehensive 深度に従い、以下の業務ロジック上の意思決定ポイントを質問形式で洗い出す。回答を受けてから Functional Design Artifacts（business-logic-model.md / business-rules.md / domain-entities.md / frontend-components.md）を生成する。

## 2. 担当ストーリー（再掲、unit-of-work-story-map.md より）

| ID | タイトル | 種別 |
|---|---|---|
| **US-1-01** | 「ご飯めんどくさい」ボタンで代行手配する | **Primary（コア）** |
| US-1-03 | 注文履歴に代行手配が記録される | Primary |
| US-1-07 | 注文完了画面から自動でメインへ戻る | Primary |
| US-2-03 | サジェストカードから 1 タップで注文する（Resolve 側） | Secondary（Unit D と共同） |
| US-X-01 | 決定疲れから解放される快感（フェーズ 1 体験） | Primary |

横串担当: BedrockAdapter / DeliveryAdapter / FallbackSuggestProvider の **interface 設計 + Mock 実装**

---

## 3. 生成する Functional Design Artifacts

| # | ファイル | 内容（Comprehensive） |
|---|---|---|
| F1 | `aidlc-docs/construction/order/functional-design/business-logic-model.md` | OrderService.PlaceOrder の詳細シーケンス図（成功 / リトライ / フォールバック / 冪等性 / 残高不足の全 5 パターン）, GetHistory の処理フロー, 状態遷移図 |
| F2 | `aidlc-docs/construction/order/functional-design/business-rules.md` | 全業務ルール（Bedrock リトライ回数 / フォールバック条件 / suggestionId 失効 / 補償戦略 / カテゴリホワイトリスト / 金額境界 / コピー文言 等） |
| F3 | `aidlc-docs/construction/order/functional-design/domain-entities.md` | OrderRecord / PlaceOrderRequest / PlaceOrderResult / OrderPlan（Bedrock 推論結果） / DeliveryInput / DeliveryOutput の構造とフィールド制約 |
| F4 | `aidlc-docs/construction/order/functional-design/frontend-components.md` | MainScreen の GoroButton 部分 / OrderCompletionScreen / useOrder hook / API 呼び出しと state 遷移 / エラーハンドリング |

---

## 4. Plan 実行ステップ（チェックリスト）

- [ ] §5 の質問群（Q-1 〜 Q-12）にユーザが回答
- [ ] 回答内容を本 Plan に反映、矛盾・曖昧さチェック
- [ ] ユーザによる Plan 承認（Step 2 完了）
- [ ] business-logic-model.md 生成
- [ ] business-rules.md 生成
- [ ] domain-entities.md 生成
- [ ] frontend-components.md 生成
- [ ] Functional Design 完了メッセージ提示（2-option ゲート）
- [ ] aidlc-state.md / audit.md 更新

---

## 5. 確認質問（Q-1 〜 Q-12）

### Question 1 — Bedrock 推論失敗時のリトライ戦略

`OrderService.PlaceOrder` で `BedrockAdapter.InferOrderPlan` が失敗した場合のリトライ方針を確定したい。services.md §2.3 では「Bedrock 失敗 × リトライ × フォールバック」と記載されているが、リトライ回数・対象エラー・タイムアウトは未確定。

A) リトライなし（即フォールバック）。シンプル、3 秒以内体感を最優先
B) **1 回のみ即時リトライ**（exponential backoff なし）→ 失敗ならフォールバック。services.md SuggestService 側と同じ戦略
C) 1 回リトライ + 100ms 待機 → 失敗ならフォールバック
D) 2 回まで指数バックオフ（100ms → 300ms） → 失敗ならフォールバック
E) Other（[Answer] 後に記述）

[Answer]: B（1 回のみ即時リトライ → 失敗ならフォールバック。SuggestService 側 `withRetry(1, ...)` と対称化）

---

### Question 2 — Bedrock 推論のタイムアウト値

`BedrockAdapter.InferOrderPlan` 呼び出しの単発タイムアウトを決めたい。NFR-PERF-01 で全体 3 秒以内が目標。

A) 1.0 秒（厳しめ。Wallet/Adapter に時間を残す）
B) **1.5 秒**（Bedrock が支配的、Wallet+Adapter は 1 秒以内に収まる前提）
C) 2.0 秒（Bedrock 優先、Wallet/Adapter 短縮前提）
D) タイムアウト無制限（API Gateway の上限に依存）
E) Other（[Answer] 後に記述）

[Answer]: B（1.5 秒。Bedrock 1 回 + リトライ 1 回でワースト 3.0 秒、NFR-PERF-01 ぎりぎり収まる）

---

### Question 3 — フォールバック生成戦略の優先順位

Bedrock 失敗時、`FallbackSuggestProvider` の挙動として以下のどちらを優先するか。

A) **履歴あり (`history>=5`)** → `BuildFromHistory(history)` の最頻パターン、**履歴なし** → `Default()` の固定プラン（吉野家 牛丼 ¥500 のような一般的・安価なもの）
B) 履歴の有無を問わず常に `Default()` を返す（実装簡素化）
C) 履歴の有無を問わず常に `BuildFromHistory()` を試み、サンプル不足時は内部で Default に落とす
D) Other（[Answer] 後に記述）

[Answer]: A（履歴 5 件以上 → BuildFromHistory、それ未満 → Default。閾値は US-2-02 と整合）

---

### Question 4 — `Default()` プランの中身

質問 3 で `Default()` を採用するシナリオがある場合、固定値として何を採用するか（ハッカソンデモ・テスト容易性の観点）。

A) **店舗: "吉野家"、メニュー: "牛丼", 金額: ¥500、カテゴリ: food**（業界トップ・最安級）
B) 店舗: "ゴロゴロ食堂"、メニュー: "おまかせ定食", 金額: ¥1,000（架空店舗、自虐的なネーミング）
C) 店舗: "CoCo壱番屋", メニュー: "カレーライス", 金額: ¥1,200（idea.md 例示と同一）
D) Other（[Answer] 後に記述）

[Answer]: D（架空の店舗ラインナップ 5 件をコード内定数配列で持ち、ランダム選択）

#### Q-4 詳細決定事項

**店舗ラインナップ**（`internal/adapters/fallback/stores.go` の定数配列として保持）

| # | 店舗名 | メニュー | 金額 | カテゴリ |
|---|---|---|---|---|
| 1 | ゴロゴロ食堂 | おまかせ定食 | ¥1,000 | food |
| 2 | ぐうたら亭 | 手抜き丼 | ¥800 | food |
| 3 | ダメ屋 | やる気なしカレー | ¥1,200 | food |
| 4 | 怠惰キッチン | 何でもよし弁当 | ¥1,500 | food |
| 5 | ふぬけ食堂 | しょうがない定食 | ¥900 | food |

**選択ロジック**: ランダム（`rand.Intn(len(stores))` 1 行）
- テスト時はシード固定で再現性を担保（business-rules.md に明記）
- フォールバックは Bedrock 失敗時の救済なので低頻度前提、多様性を優先

**保管場所**: Go コード内の定数配列（`internal/adapters/fallback/`）。DynamoDB / 環境変数は採用しない（実装最小・レイテンシゼロ・PR diff で世界観レビュー可能）

**ネーミング基準**: 自虐的・食欲のなさ・金額バラつき（¥800〜¥1,500）でダメ化メトリクスに変動を出す

---

### Question 5 — `suggestionId` 経由注文の再検証

`PlaceOrder(req)` で `req.suggestionId` が指定されている場合、`SuggestService.ResolveSuggestion(suggestionId)` で plan を復元する。この時、Bedrock を**もう一度叩いて検証する**か、**保存値をそのまま使う**か。

A) **保存値をそのまま使う**（Bedrock 再呼び出しなし、コスト最小・3 秒以内必達）
B) Bedrock で検証する（plan の妥当性確認、コスト増・遅延増）
C) suggestionId が「30 分以内」の新鮮なものなら保存値、古ければ Bedrock 再呼び出し
D) Other（[Answer] 後に記述）

[Answer]: A（保存値をそのまま使う。Bedrock 再検証なし。US-2-03 と整合、NFR-PERF-01 / NFR-DEG-01 を最優先）

---

### Question 6 — `suggestionId` の失効と扱い

`SuggestService` 側で SuggestionID は TTL 30 分で消える設計（services.md §2.4）。`OrderService.PlaceOrder(suggestionId=expired)` を受け取った場合の挙動は？

A) **エラー応答 410 Gone** "サジェストの有効期限が切れました。もう一度お試しください"（明示的にユーザに知らせる）
B) フォールバックで通常の Bedrock 推論フローに移行（透過的、ユーザは何が起きたか分からない）
C) フォールバックで `Default()` に移行（最も低摩擦）
D) Other（[Answer] 後に記述）

[Answer]: B（透過的に通常の Bedrock 推論フローへフォールバック。ユーザに失効を見せない、サーバログには `suggestion_expired` 警告を残す）

---

### Question 7 — `idempotencyKey` の発行元と形式

US-1-05 の二重引き落とし防止に使う冪等性キーの**発行元**と**形式**を確定したい。

A) **クライアント (Frontend) 発行 / ULID**（Frontend が `useOrder` hook 内で `crypto.randomUUID()` 互換 ULID を生成、API リクエスト時に Header `Idempotency-Key` で送信）
B) クライアント発行 / UUID v4
C) サーバ発行（リクエストごとに採番）→ 連打時の重複検知不可能なので NG
D) Other（[Answer] 後に記述）

[Answer]: A（Frontend 発行 ULID。`useOrder` hook がボタン表示時に 1 個生成、HTTP Header `Idempotency-Key` で送信。OrderID と発行体系を統一）

---

### Question 8 — `idempotencyKey` の有効期限とスコープ

冪等性キーの**有効範囲**と**保存期間**を確定したい（IdempotencyRepository は Unit B 所有、Unit C は呼び出し側）。

A) **有効期限 24 時間 / userId 単位でユニーク**（同一ユーザの同じキー再送は 24h 内なら無害化）
B) 有効期限 1 時間 / userId 単位
C) 永続（TTL なし）
D) Other（[Answer] 後に記述）

[Answer]: A（TTL 24 時間 / userId 単位でユニーク。DynamoDB の TTL 自動削除で運用ゼロ。Unit B 所有の IdempotencyRepository 仕様として合意）

---

### Question 9 — `DeliveryAdapter` 失敗時の補償（Wallet 戻し）

`OrderService` のステップ 3 で `DeliveryAdapter.PlaceOrder` が失敗した場合、既に減算済みの Wallet 残高をどう扱うか。services.md は「本 MVP は実装せず、ログで警告のみ」と書かれているが、Comprehensive 深度なので明示確定したい。

A) **本 MVP は補償実装なし（log.Error のみ、500 応答）**。ハッカソンスコープ外、設計のみ business-rules.md に記載
B) 同期的に `WalletService.Refund(userID, amount, idempotencyKey)` を呼ぶ（best-effort、失敗しても 500 応答）
C) DLQ パターン（SQS）で非同期補償。Out of Scope（マイクロサービス的）
D) Other（[Answer] 後に記述）

[Answer]: A（本 MVP は補償実装なし。log.Error 構造化ログ + CloudWatch アラーム検知のみ。MockDeliveryAdapter は失敗しないため発生しない。business-rules.md に「将来の補償戦略」セクションで B 案を第一候補として明文化）

---

### Question 10 — `OrderHistoryRepository.Insert` 失敗時の扱い

`OrderService` のステップ 4（履歴 Insert）が失敗した場合、注文自体は成功（Wallet 減算 + Delivery 完了済み）。応答はどうするか。

A) **履歴失敗を log.Error し、応答は 200/201 成功を返す**（履歴は付随情報、注文自体は完了。ユーザ体験優先）
B) 履歴失敗で 500 応答（一貫性優先、ユーザはエラー画面を見るが Wallet/Delivery は残る → 補償が複雑）
C) リトライ 3 回 → 失敗なら EventBridge にイベント発火して非同期補償（Out of Scope）
D) Other（[Answer] 後に記述）

[Answer]: A（log.Error 構造化ログ + 応答は 200/201 成功。「ユーザ体験 > データ完全性」の優先順位を Comprehensive 深度で明文化。CloudWatch アラームで欠落検知）

---

### Question 11 — `GetHistory` の取得件数とソート

`OrderService.GetHistory(userID, limit)` の挙動。stories.md US-1-03 では「直近 3 ヶ月」「TTL 90 日」「曜日記録」とある。

A) **デフォルト limit=20、最大 100、降順（OrderedAt DESC）、TTL 90 日内のみ**（DynamoDB Query + ScanIndexForward=false）
B) limit=10 デフォルト、最大 50、降順
C) limit クエリ無視で常に最新 30 件
D) Other（[Answer] 後に記述）

[Answer]: A（デフォルト limit=20、最大 100、降順、TTL 90 日内のみ。`GET /orders?limit=20`、limit=0 / limit>100 は 400 エラー、ページング無しで MVP）

---

### Question 12 — Frontend (`useOrder` hook) のエラーハンドリング UX

API レスポンスがエラーだった場合の Frontend 動作（OrderCompletionScreen / MainScreen の動き）。

A) **402 INSUFFICIENT_BALANCE → BudgetEmpty 画面遷移**、**410 SUGGESTION_EXPIRED → トースト「サジェストの有効期限が切れました」+ MainScreen 維持**、**500 → トースト「ちょっとうまくいかないみたい」+ MainScreen 維持**（自虐的・低摩擦）
B) 全エラーで共通のエラー画面に遷移、再試行ボタン表示（一般的）
C) エラーを表示せずデフォルトプラン（Default）で完了画面を出す（究極の低摩擦、ユーザに失敗を悟らせない）
D) Other（[Answer] 後に記述）

[Answer]: A（402 → BudgetEmpty 画面遷移 / 500 → 自虐トースト「ちょっとうまくいかないみたい」+ MainScreen 維持。Q-6 決定により 410 は実際には発生しないためハンドラ不要。idempotencyKey で再送時の二重引き落とし防止）

---

## ✅ 全 12 問 回答完了サマリ

| Q | テーマ | 採用 |
|---|---|---|
| Q-1 | Bedrock リトライ戦略 | B（1 回リトライ → フォールバック） |
| Q-2 | Bedrock タイムアウト | B（1.5 秒） |
| Q-3 | フォールバック優先順位 | A（履歴 5 件以上 → BuildFromHistory、それ未満 → Default） |
| Q-4 | Default プラン中身 | D（架空店舗 5 件をコード内定数配列、ランダム選択） |
| Q-5 | suggestionId 経由再検証 | A（保存値そのまま、Bedrock 再呼び出しなし） |
| Q-6 | suggestionId 失効時 | B（透過的に通常 Bedrock フローへフォールバック） |
| Q-7 | idempotencyKey 発行元/形式 | A（Frontend 発行 ULID） |
| Q-8 | idempotencyKey TTL/スコープ | A（24 時間 / userId 単位） |
| Q-9 | DeliveryAdapter 失敗時 | A（log.Error のみ、補償は将来対応） |
| Q-10 | OrderHistory Insert 失敗時 | A（log.Error して 200/201 成功応答） |
| Q-11 | GetHistory 取得仕様 | A（デフォルト 20 / 最大 100 / 降順 / TTL 90 日内） |
| Q-12 | Frontend エラー UX | A（402 → BudgetEmpty 遷移 / 500 → 自虐トースト） |

## 矛盾チェック結果

- ✅ **Q-1 + Q-2 ↔ NFR-PERF-01**: Bedrock 1.5 秒 × 2 回リトライ = ワースト 3.0 秒。NFR ぎりぎり収まる（許容範囲、NFR Requirements で再評価）
- ✅ **Q-5 + Q-6 整合**: 保存値そのまま使う + 失効時は Bedrock フォールバック → 矛盾なし
- ✅ **Q-7 + Q-8 ↔ Unit B Wallet**: Frontend 発行 ULID / 24h TTL / userId スコープ → Unit B IdempotencyRepository の前提として整合
- ✅ **Q-9 + Q-10 ↔ NFR-REL-01**: 補償なしを明文化したうえで、business-rules.md の「将来の補償戦略」セクションで NFR-REL-01 とのギャップを記録 → 矛盾なし

矛盾なしのため、ユーザ Plan 承認を得てから Functional Design Artifacts 4 ファイルの生成へ進む。

---

## 9. 事後追記: 凍結 Interface 契約への整合（2026-05-21 同日）

PR #64（develop）で `aidlc-docs/construction/interfaces/unit-interfaces.md` が追加され、Wave 1 並列化のための Unit 間 Interface 契約が凍結された。本 Plan の決定事項と凍結契約の差分を以下のように Functional Design Artifacts に反映済み:

| 項目 | 本 Plan の決定 | 凍結契約 | 反映先 |
|---|---|---|---|
| `idempotencyKey` テーブル PK | `(userId, key)` 複合（誤り） | `key` 単独、payload で衝突検知 | BR-C13 修正 |
| 冪等命中時の OrderID 取得 | OrderHistory に GSI を追加して GetByIdempotencyKey | Wallet payload から OrderID を取得 → 通常 Get | BR-C15 修正 / S-04 シーケンス図修正 |
| OrderHistory GSI | `GSI_IdempotencyKey` を新設 | `gsi_byCreatedAt` 1 個のみ | domain-entities.md §7.4 で不採用を明文化 |
| TTL 属性名 | `TTL` | `expiresAt` | BR-C20 / BR-C22 / domain-entities.md 全体 |
| 409 IDEMPOTENCY_CONFLICT | 言及なし | エラー一覧に存在 | BR-C39 新設 / 状態遷移図 / Frontend ハンドラ |
| カテゴリ | `"food"` のみ（MVP） | `"food" \| "errand" \| ...` 拡張余地 | BR-C27 注記追加 |

これらは Q-1〜Q-12 のヒアリング結論（リトライ / フォールバック / suggestionId / 1 タップ完結）には影響せず、永続化方式とエラー契約の追従修正に留まる。


---

## 6. 矛盾チェック観点

回答受領後、AI が以下の整合性を検証する:

- Q-1 リトライ回数 ↔ Q-2 タイムアウト ↔ NFR-PERF-01 (3 秒以内) の合算
- Q-5 (suggestionId 検証なし) ↔ Q-6 (失効時 Bedrock フォールバック) の組合せが矛盾しないか
- Q-7 (Frontend 発行 ULID) ↔ Q-8 (有効期限 24h) ↔ Wallet (Unit B) 側の Idempotency 仕様
- Q-9 / Q-10 (失敗時補償なし) ↔ NFR-REL-01 (残高不変) ↔ business-rules.md で明示する補償戦略

矛盾が見つかった場合は `order-functional-design-clarification-questions.md` を生成して再ヒアリング。

---

## 7. 完了条件

- 全 12 問の `[Answer]:` が埋まっている
- 矛盾なしを確認
- ユーザが Plan を承認（Part 1 完了）
- Part 2: Functional Design Artifacts 4 ファイルを生成
- Functional Design 完了メッセージで「Continue to Next Stage」または「Request Changes」を確認

---

## 8. 次ステージへの引き継ぎ

Functional Design 承認後は **NFR Requirements (Comprehensive)** へ進む。Functional Design で確定したリトライ回数・タイムアウト値・冪等性スコープは NFR Requirements の入力となる。
