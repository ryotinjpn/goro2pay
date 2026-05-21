# Unit C (`order`) — Business Rules

**Document Version**: 1.0
**Created**: 2026-05-21
**Stage**: Construction / Functional Design
**Unit**: C — `order`
**Depth**: Comprehensive
**Related**: [business-logic-model.md](./business-logic-model.md), [domain-entities.md](./domain-entities.md), [frontend-components.md](./frontend-components.md), 凍結契約 [unit-interfaces.md](../../interfaces/unit-interfaces.md)
**Aligned with**: 凍結契約 §3.4 (IdempotencyKeys), §4.3 (REST エラー一覧 / 409 IDEMPOTENCY_CONFLICT)

本ドキュメントは Unit C `order` の **業務ルール / 分岐条件 / バリデーション / 閾値 / コピー文言 / 将来の補償戦略** を網羅する業務ルールカタログである。Comprehensive 深度として全ルールを ID 付きで管理する（`BR-Cxx`）。

---

## 1. ルール識別子規則

- 形式: `BR-C{番号}`
- C: Unit C 所属を示す
- 番号: 連番（カテゴリ無関係に通し番号、ID 安定化のため）
- 各ルールに **由来（Inception 成果物 or Plan 質問）** と **適用箇所（business-logic-model.md のステップ）** を明記

---

## 2. リトライ・タイムアウトに関するルール

### BR-C01: Bedrock リトライは 1 回のみ
- **由来**: Plan Q-1 = B
- **ルール**: `BedrockAdapter.InferOrderPlan` の呼び出しは最大 2 回（初回 + リトライ 1 回）。それ以上はリトライしない。
- **対象エラー**: `context.DeadlineExceeded` / `ThrottlingException` / `ServiceUnavailableException` / 任意のネットワークエラー全般
- **非対象エラー**: `ValidationException` / `AccessDeniedException` などの永続エラーは **1 回目で即フォールバック**（リトライしても無駄なため）
- **適用箇所**: business-logic-model.md §2.2

### BR-C02: Bedrock 1 回あたりのタイムアウトは 1.5 秒
- **由来**: Plan Q-2 = B
- **ルール**: 各 Bedrock 呼び出しに `context.WithTimeout(parent, 1500ms)` を適用
- **保護的措置**: `OrderService.PlaceOrder` 全体に親タイムアウト 5 秒を設ける（API Gateway 上限 29 秒 / 体感 3 秒予算超過時の非常停止）
- **適用箇所**: business-logic-model.md §2.2

### BR-C03: リトライ間の待機時間はゼロ
- **由来**: Plan Q-1 = B（exponential backoff なし）
- **ルール**: ATTEMPT 1 失敗 → 即 ATTEMPT 2 を開始（`time.Sleep` なし）
- **理由**: スロットリング対策よりも 3 秒予算遵守を優先
- **適用箇所**: business-logic-model.md §2.2

### BR-C04: 親 Context のキャンセルは即時伝播
- **由来**: Go 慣習 + NFR-PERF-01
- **ルール**: クライアントが切断 / API Gateway がタイムアウトした場合、進行中の Bedrock / DeliveryAdapter / DynamoDB 呼び出しを `ctx.Err() == context.Canceled` で打ち切る
- **理由**: ゾンビ Lambda 実行を防ぎコスト・残高整合の悪化を回避

---

## 3. フォールバック戦略に関するルール

### BR-C05: フォールバック発動条件
- **由来**: Plan Q-1 + Q-3
- **ルール**: 以下のいずれかを満たした場合に `FallbackSuggestProvider` を呼び出す
  1. Bedrock の 2 連続失敗（BR-C01 を消化）
  2. ATTEMPT 1 が永続エラー（リトライ非対象、BR-C01 注釈）
- **適用箇所**: business-logic-model.md §2.2 FALLBACK ブロック

### BR-C06: フォールバック分岐の閾値
- **由来**: Plan Q-3 = A、stories.md US-2-02 と整合
- **ルール**: フォールバック時の Plan 生成は **直近 30 日の履歴件数** で分岐
  - **5 件以上**: `BuildFromHistory(history)` を呼ぶ → 最頻パターンを返す
  - **4 件以下**: `Default()` を呼ぶ → 5 店舗からランダム選択
- **理由**: US-2-02 と同じ閾値で、Unit C / D で統一する
- **適用箇所**: business-logic-model.md §2.2

### BR-C07: `Default()` の店舗ラインナップ（5 件）
- **由来**: Plan Q-4 = D（Other）の詳細決定
- **ルール**: 以下の固定 5 店舗を `internal/adapters/fallback/stores.go` に定数配列として定義し、`rand.Intn(5)` でランダム選択する

| Index | Store | Menu | Amount | Category |
|---|---|---|---|---|
| 0 | ゴロゴロ食堂 | おまかせ定食 | ¥1,000 | food |
| 1 | ぐうたら亭 | 手抜き丼 | ¥800 | food |
| 2 | ダメ屋 | やる気なしカレー | ¥1,200 | food |
| 3 | 怠惰キッチン | 何でもよし弁当 | ¥1,500 | food |
| 4 | ふぬけ食堂 | しょうがない定食 | ¥900 | food |

- **理由**: ダメ化UX 世界観に合うネーミング、金額分散（¥800〜¥1,500）でメトリクスに変動を出す
- **テスト容易性**: 単体テストでは `rand.New(rand.NewSource(seed))` でシード固定（再現性担保）
- **適用箇所**: business-logic-model.md §2.2 `FallbackSuggestProvider.Default()`

### BR-C08: `BuildFromHistory` の最頻判定方法
- **由来**: services.md §2.3 + Comprehensive 深度の補完
- **ルール**: 履歴 30 件のうち `(StoreName, MenuName)` の出現回数で最頻のもの 1 件を返す。同点の場合は最新の OrderedAt を採用。
- **金額**: 同 `(StoreName, MenuName)` の **最新の Amount** を採用（メニュー価格変更を反映）
- **カテゴリ**: 履歴の Category をそのまま使用（MVP では "food" 固定）
- **適用箇所**: `FallbackSuggestProvider.BuildFromHistory`

---

## 4. サジェスト経由注文に関するルール

### BR-C09: `suggestionId` 経由注文では Bedrock 再検証しない
- **由来**: Plan Q-5 = A、stories.md US-2-03
- **ルール**: `req.SuggestionID != nil` の場合、Unit D `SuggestService.ResolveSuggestion` から取得した保存値をそのまま `Plan` として採用する。Bedrock 再呼び出しは行わない。
- **理由**: NFR-PERF-01（3 秒以内）と NFR-DEG-01（1 タップ完結）を最優先。サジェストはユーザに表示された時点でユーザ確認済みとみなす
- **適用箇所**: business-logic-model.md §2.3

### BR-C10: 失効した `suggestionId` は透過的に Bedrock フローへフォールバック
- **由来**: Plan Q-6 = B
- **ルール**: `SuggestService.ResolveSuggestion(suggestionID)` が `nil`（TTL 30 分超過などで保存値喪失）を返した場合、エラー応答にせず、**通常の `InferPlanViaBedrock` フローを実行** する
- **ログ**: `log.Warn("suggestion_expired", suggestionID, userID)` を出力（運用観察用）
- **ユーザへの通知**: なし（透過的、NFR-DEG-05）
- **適用箇所**: business-logic-model.md §2.3 + §2.2

### BR-C11: `suggestionId` の検証
- **由来**: ULID 仕様 + 防御的プログラミング
- **ルール**:
  - 形式: ULID (Crockford's Base32, 26 文字)
  - 不正形式（長さ違反 / 文字種違反）の場合は **400 INVALID_SUGGESTION_ID** を返す（フォールバックではなく明示エラー）
  - 形式は正しいが Unit D 側で見つからない場合は BR-C10 のフォールバック扱い
- **適用箇所**: `OrderHandler.PlaceOrder` の入力バリデーション

---

## 5. 冪等性に関するルール

### BR-C12: `idempotencyKey` は Frontend が ULID で発行
- **由来**: Plan Q-7 = A
- **ルール**:
  - 発行元: Frontend (`useOrder` hook)
  - 形式: ULID (`crypto.randomUUID` 互換 ULID 26 文字)
  - 発行タイミング: ボタンが活性化されたときに 1 個生成、画面遷移までキャッシュ
  - HTTP 送信: Header `Idempotency-Key: <ULID>` （Body にも同値 `idempotencyKey` を含める二重送信、Adapter 互換性のため）
- **適用箇所**: business-logic-model.md §3.4 + frontend-components.md §`useOrder`

### BR-C13: `idempotencyKey` の TTL とスコープ
- **由来**: Plan Q-8 = A、凍結契約 `unit-interfaces.md §3.4` に追従
- **ルール**:
  - TTL: 24 時間（DynamoDB の `expiresAt` 属性で自動削除）
  - **PK**: `key`（idempotencyKey 単独）の **グローバルユニーク**（凍結契約。userID は payload に含めて衝突検知する）
  - **payload**: `userID` / `amount` / `orderID` / `createdAt` を保存
  - 命中時の挙動:
    1. 既存 entry の `payload.userID` が現在のリクエストの userID と一致するか確認
    2. 一致 → 冪等命中、`payload.orderID` で `OrderHistoryRepository.Get(userID, orderID)` を呼び出し、初回応答と同一の OrderID / Amount / RemainingBalance を返す
    3. 不一致または amount が異なる → **409 IDEMPOTENCY_CONFLICT**（BR-C39 参照）
- **注**: IdempotencyRepository 自体は **Unit B 所有**。Unit C は呼び出し側として本ルールに従う
- **適用箇所**: business-logic-model.md §3.4 連打シナリオ

### BR-C14: 冪等命中時の応答
- **由来**: stories.md US-1-05 受入基準 + Comprehensive 深度の補完
- **ルール**: 連打 2 回目以降の応答は以下を満たす
  - HTTP 201 (新規作成と同じ)
  - body の `OrderID` / `StoreName` / `MenuName` / `Amount` / `RemainingBalance` は **初回と完全一致**
  - body に `idempotent: true` フラグを含める（Frontend が「再送だな」と検知できる、ただしユーザには見せない）
- **適用箇所**: business-logic-model.md §3.4

### BR-C15: 冪等命中時の応答の取得方法
- **由来**: Comprehensive 深度の補完、凍結契約 `unit-interfaces.md §4.4` に追従
- **ルール**: `WalletService.Deduct` が `idempotent=true` を返した場合、Unit B の IdempotencyRepository entry の **payload.orderID** を Unit B 経由で取得し、`OrderHistoryRepository.Get(userID, orderID)` で初回 Insert したレコードを引き直す
- **OrderHistory アクセスパターン**: PK=`userId`, SK=`orderId` の通常 Get で完結（**追加 GSI 不要**、凍結契約は `gsi_byCreatedAt` 1 個のみ）
- **WalletService.Deduct の戻り値**: `DeductResult.Idempotent=true` のケースで、Unit B が `payload.orderID` を `DeductResult` の追加フィールドとして返すか、または別 API（例: `IdempotencyReader.GetPayload`）で取得するかは Unit B 側の設計判断（NFR Requirements 段で Unit B と摺合せ）
- **適用箇所**: business-logic-model.md §2.1 STEP 2 idempotent 分岐

---

## 6. 残高・金額に関するルール

### BR-C39: 冪等性コンフリクト時のエラー応答
- **由来**: 凍結契約 `unit-interfaces.md §4.3` の `409 IDEMPOTENCY_CONFLICT`、Unit B `ErrIdempotencyConflict`
- **ルール**: 同一 `idempotencyKey` で既存 payload の `userID` または `amount` が現リクエストと異なる場合、Wallet 減算を行わず以下を返す
  - HTTP: 409 Conflict
  - エラーコード: `IDEMPOTENCY_CONFLICT`
  - body: `{code: "IDEMPOTENCY_CONFLICT", message: "同じキーで異なる注文が記録されています"}`
- **発生原因**:
  - 開発者起因（同一 ULID を別注文で再利用、テスト時のみ起きうる）
  - 異常系（Frontend バグで ULID 再利用）
- **本 MVP での発生確率**: 0%（BR-C12 でボタン活性化ごとに新 ULID 生成、画面遷移までキャッシュのみ）
- **理由**: 凍結契約として全 Unit に通知された安全装置。Unit C 側は実装のみで、ユーザに見せる UI は不要（致命的バグ顕在化として 500 系扱いでも良いが、契約上 409 を返す）
- **適用箇所**: `OrderHandler.PlaceOrder` の `Wallet.Deduct` 呼び出しエラーハンドリング

### BR-C16: 残高不足時のエラー応答
- **由来**: stories.md US-1-04 / US-1-06 / Plan Q-12 = A
- **ルール**:
  - HTTP: 402 Payment Required
  - エラーコード: `INSUFFICIENT_BALANCE`
  - body: `{code: "INSUFFICIENT_BALANCE", balance: <現残高>}`
- **Frontend 動作**: BR-C32 参照（BudgetEmpty 画面遷移）
- **DynamoDB CAS**: Wallet 減算は `balance >= amount` の条件付き書き込みで `ConditionalCheckFailedException` を捕捉し `ErrInsufficientBalance` に変換（Unit B 担当だが Unit C はこのエラーを受信する側）
- **適用箇所**: business-logic-model.md §3.5

### BR-C17: 金額の境界値
- **ルール**:
  - **最小**: 1 円（`Amount >= 1`）
  - **最大**: 100,000 円（`Amount <= 100_000`、ダメ予算上限と同値）
  - 範囲外を Bedrock が返した場合は `Amount = clamp(Amount, 1, 100_000)` で丸めて続行（注文を諦めずデモ継続）
  - フォールバック / 履歴ベースのプランでも本範囲を保証
- **適用箇所**: `OrderService.PlaceOrder` STEP 1 後

### BR-C18: 金額の単位は整数円
- **ルール**: float を使わない、すべて int で扱う。1000.5 円のようなデータは入力時点で reject (400 INVALID_AMOUNT)
- **理由**: 金融ドメインの誤差累積防止

---

## 7. 履歴記録・取得に関するルール

### BR-C19: 履歴 Insert 失敗は 200/201 成功で応答
- **由来**: Plan Q-10 = A
- **ルール**: `OrderHistoryRepository.Insert` が失敗しても、Wallet 減算 + Delivery 手配が成功していれば **200/201 成功応答** を返す
- **ログ**: `log.Error("order_history_insert_failed", userID, orderID, idempotencyKey, error)` を構造化ログで出力
- **検知**: CloudWatch メトリクスフィルタで `event="order_history_insert_failed"` を監視（Infrastructure Design で詳細化）
- **理由**: 履歴は副次情報。ユーザ体験を優先（ご飯が届くのに画面エラーは最悪のパターン）
- **適用箇所**: business-logic-model.md §2.1 STEP 4

### BR-C20: 履歴の保持期間
- **由来**: requirements.md FR-LEARNING-02 / stories.md US-1-03 / US-2-04、凍結契約 `unit-interfaces.md §4.4`
- **ルール**: DynamoDB の `expiresAt` 属性（DynamoDB TTL 機能）で `orderedAt + 90 日` のエポック秒を設定。期限超過は自動削除。
- **適用箇所**: `OrderHistoryRepository.Insert`、`domain-entities.md §2.1.2 expiresAt`

### BR-C21: `GetHistory` のクエリパラメータ
- **由来**: Plan Q-11 = A
- **ルール**:
  - `GET /orders?limit=<N>` でクエリパラメータ受領
  - デフォルト: `limit=20` （省略時）
  - 範囲: `1 <= limit <= 100`、範囲外は 400 INVALID_LIMIT
  - 並び順: `OrderedAt DESC` 固定（最新が先頭）
  - フィルタ: なし（カテゴリ・期間絞込は MVP スコープ外）
- **適用箇所**: `OrderHandler.GetHistory`、`OrderHistoryRepository.Query`

### BR-C22: 履歴データの記録項目
- **由来**: stories.md US-1-03 受入基準、凍結契約 `unit-interfaces.md §4.1`
- **ルール**: 1 注文につき以下を DynamoDB に記録
  - **公開フィールド**（凍結契約 §4.1 OrderRecord）: orderId / userId / category / storeName / menuName / amount / orderedAt
  - **内部追加属性**: idempotencyKey / dayOfWeek (JST 換算 Monday〜Sunday) / source ("button" or "suggest") / expiresAt (Unix epoch、orderedAt + 90 日)
- **属性名規約**: DynamoDB は camelCase、Go 型は PascalCase
- **適用箇所**: domain-entities.md §2.1 OrderRecord（公開）+ §2.1.2 内部追加属性

### BR-C23: `Source` フィールドの値
- **由来**: Comprehensive 深度の追加（学習・分析用）
- **ルール**: `req.SuggestionID == nil` → `Source = "button"`、`req.SuggestionID != nil` → `Source = "suggest"`
- **用途**: Unit D の学習プロンプト生成 / Unit E のメトリクス（ダメ化加速度の分析）

---

## 8. DeliveryAdapter（外部手配）に関するルール

### BR-C24: `MockDeliveryAdapter` は常に成功する
- **由来**: services.md §3.1 + Plan Q-9 前提
- **ルール**: 本 MVP の `MockDeliveryAdapter.PlaceOrder` は固定応答 `{ExternalOrderID: "mock-<ulid>", Status: "accepted", ETAMinutes: 30}` を返し、エラーを返さない
- **テスト**: `FailingDeliveryAdapter` を別途用意（テスト専用、business-rules には実装義務なし、Code Generation 段で必要に応じて作成）
- **適用箇所**: business-logic-model.md §3.1 / §6.2

### BR-C25: 外部手配失敗時の補償なし（本 MVP）
- **由来**: Plan Q-9 = A
- **ルール**: 万が一 `DeliveryAdapter.PlaceOrder` がエラーを返した場合、Wallet を戻さない。HTTP 500 を返却し、CloudWatch アラームで運用通知。
- **ログ**: `log.Error("delivery_failed", userID, plan, idempotencyKey)` を構造化ログで出力
- **理由**: Mock では起きない、SAGA / DLQ は requirements.md §2.4 で Out of Scope
- **適用箇所**: business-logic-model.md §2.1 STEP 3

### BR-C26: 将来の補償戦略（実装非対象、設計のみ）
- **由来**: Plan Q-9 = A の補足、Comprehensive 深度の要件
- **ルール**: 将来、実 DeliveryAdapter（UberEats / 出前館）連携時は以下を採用する
  1. **同期 Refund を第一候補**: `WalletService.Refund(userID, amount, idempotencyKey + "-refund")` を `DeliveryAdapter.PlaceOrder` 失敗直後に呼ぶ。Refund 自体に冪等性キーを別途付与（多重 Refund 防止）
  2. Refund 失敗時は SNS Topic `goropay-refund-alert` に通知し運用者にエスカレーション
  3. それでも失敗した場合のみ DLQ パターン（SQS + Lambda）で非同期補償
- **記載理由**: 実装はしないが、設計意図を残し将来の AI-DLC リターン時の意思決定を再現可能にする
- **NFR-REL-01 とのギャップ**: 本 MVP では NFR-REL-01（残高不変条件）が Mock により実害なく成立する。実連携時は本ルール適用が必須

---

## 9. カテゴリ・拡張に関するルール

### BR-C27: 対応カテゴリは "food" のみ（MVP）
- **由来**: requirements.md §2.4 In Scope（最小ユースケース「ご飯めんどくさい」）、凍結契約 `unit-interfaces.md §4.1` の `Category string  // "food" | "errand" | ...`（拡張余地あり）に整合
- **ルール（MVP）**: `req.Category` は `"food"` のみ受理。他の値は 400 UNSUPPORTED_CATEGORY
- **将来拡張**: 凍結契約上は `"errand"` 等の追加カテゴリを許容する設計。具体的に何を追加するかは将来の Inception 増分で決定。本 MVP では実装拒否、凍結契約だけが将来余地を保持
- **適用箇所**: `OrderHandler.PlaceOrder` バリデーション

### BR-C28: 入力バリデーション
- **ルール**: `OrderHandler.PlaceOrder` は以下を順番に検証
  1. JWT が有効（Unit A の middleware で実施済 → userID Context にあること）
  2. `Category == "food"` （BR-C27）
  3. `IdempotencyKey` が ULID 形式（26 文字 Crockford Base32）
  4. `SuggestionID`（任意）が ULID 形式
- 不正は **400 InvalidRequest** とエラー詳細をレスポンスボディに格納

---

## 10. コピー文言・Frontend 動作に関するルール

### BR-C29: 完了画面の表示テキスト
- **由来**: stories.md US-1-01 受入基準
- **ルール**: 完了画面に **`注文完了: <店舗名> <メニュー名> ¥<金額> - 残りダメ予算 ¥<残高>`** を表示
- **改行**: 店舗名行 / メニュー行 / 金額行 / 残高行に分けて表示（ダメ化UX: 文字を大きく、視線誘導）

### BR-C30: 完了画面の自動遷移
- **由来**: stories.md US-1-07 / Plan Q-12
- **ルール**: 完了画面表示後 **5 秒後に MainScreen へ自動遷移**
- **遷移方法**: `setTimeout(() => router.push("/"), 5000)`
- **解除条件**: ユーザが画面操作（タップ）した場合は即時遷移（後述 BR-C31 と整合）

### BR-C31: 完了画面の手動遷移
- **由来**: NFR-DEG-01 補強（待ちすら奪う）
- **ルール**: 画面のどこをタップしても即時 MainScreen 遷移。「メインに戻る」ボタンは設置しない
- **理由**: 「戻る」操作の認知負荷すら排除する

### BR-C32: 残高不足時のフロント動作
- **由来**: Plan Q-12 = A、stories.md US-1-04
- **ルール**: 402 INSUFFICIENT_BALANCE 受領時、`router.push("/budget-empty?balance=<残高>")` で BudgetEmpty 画面（Unit E 所有）へ遷移
- **エラー画面の文言**: 「今月、ダメになれません（残高: ¥{balance}）」
- **画面 Owner**: BudgetEmpty 画面自体は Unit E 所有。Unit C は遷移するだけ

### BR-C33: 一般エラー時のフロント動作
- **由来**: Plan Q-12 = A
- **ルール**: 500 / その他エラー受領時
  - トースト通知: **「ちょっとうまくいかないみたい」**（自虐的・低摩擦）
  - 画面遷移: なし（MainScreen 維持）
  - ボタン: 即時再活性化（連打耐性は BR-C12〜C15 で担保）
- **デバッグ**: 開発環境（`NODE_ENV !== "production"`）のみ console.error に詳細出力

### BR-C34: ボタン押下中の表示
- **由来**: NFR-DEG-01 + 一般的 UX
- **ルール**: ボタン押下から応答受領までの間、ボタンに **「考え中…」** のラベルを表示し、ボタンを連打不能に disabled 化
- **理由**: ただし冪等性は BR-C12〜C15 で担保するため、disabled 化は UX 補強であって必須ではない
- **タイムアウト**: 4 秒経っても応答がなければ disabled 解除（NFR-PERF-01 ワースト 3 秒 + 余裕 1 秒）

---

## 11. ログ・監査に関するルール

### BR-C35: 構造化ログのフィールド統一
- **ルール**: Unit C のすべてのログは以下のフィールドを最低限含む（CloudWatch Logs Insights で `stats by event` クエリ可能とする）

| フィールド | 必須 | 例 |
|---|---|---|
| `event` | ✓ | `"order_completed"`, `"bedrock_fallback"`, `"insufficient_balance"` |
| `userID` | ✓ | `"USER#alice"` |
| `orderID` | ✓ (発行後) | `"01HX..."` |
| `idempotencyKey` | ✓ (取得後) | `"01HY..."` |
| `severity` | ✓ | `"info"`, `"warn"`, `"error"` |
| `latencyMs` | ✓ (Step 完了後) | `1234` |

- **PII 取り扱い**: `email` などの PII は出さない（userID は Cognito sub UUID）
- **適用箇所**: 全業務ロジックの全ログ呼び出し

### BR-C36: ログ出力対象イベント一覧

| イベント | severity | 出力タイミング |
|---|---|---|
| `order_initiated` | info | OrderService.PlaceOrder 受領時 |
| `bedrock_attempt_failed` | warn | Bedrock 1 回目失敗時 |
| `bedrock_fallback` | warn | Bedrock 2 回失敗 → フォールバック発動時 |
| `suggestion_expired` | warn | suggestionId 失効時 |
| `idempotency_hit` | info | 連打 2 回目以降の命中時 |
| `insufficient_balance` | info | 残高不足で 402 応答時 |
| `idempotency_conflict` | error | 同 key で別 payload を検知し 409 応答時（BR-C39、本 MVP は通常発生しない） |
| `delivery_failed` | error | DeliveryAdapter エラー時（本 MVP は発生しない） |
| `order_history_insert_failed` | error | OrderHistory Insert 失敗時 |
| `order_completed` | info | 201 応答時 |

---

## 12. 性能・可観測性に関するルール（Functional Design 視点）

### BR-C37: レイテンシ予算の内訳
- **由来**: NFR-PERF-01（3 秒以内）+ Plan Q-1 + Q-2
- **ルール**: 各ステップに以下のレイテンシ予算を設ける（合算でワースト 3.0 秒）

| ステップ | 通常時 | ワースト時 |
|---|---|---|
| API Gateway → Lambda | 150ms | 300ms |
| Bedrock 推論 1 回 | 800ms | 1500ms |
| Bedrock リトライ 1 回 | — | 1500ms（ただしフォールバック時は不要） |
| Wallet 減算 (DynamoDB UpdateItem CAS) | 30ms | 100ms |
| Delivery (Mock) | 5ms | 10ms |
| OrderHistory Insert | 30ms | 100ms |
| Lambda → API Gateway | 50ms | 100ms |

- **NFR Requirements ステージ** で再評価し、各ステップへの SLI（target / threshold）を確定する

### BR-C38: メトリクス送出
- **ルール**: Unit C は CloudWatch カスタムメトリクスを以下の名前で出す
  - `OrderPlaced` (Count)
  - `OrderInsufficientBalance` (Count)
  - `BedrockFallbackTriggered` (Count)
  - `OrderE2ELatencyMs` (Histogram)
- **詳細**: NFR Design / Infrastructure Design で実装方式（EMF / PutMetricData）を確定

---

## 13. ルール一覧サマリ表

### 13.1 Inception 由来のルール

| BR-ID | 由来 | 一行サマリ |
|---|---|---|
| BR-C16 | US-1-04 / US-1-06 | 残高不足は 402 / DynamoDB CAS |
| BR-C19 | Q-10 + Comprehensive | 履歴 Insert 失敗でも 200 |
| BR-C20 | FR-LEARNING-02 | 履歴 90 日 TTL |
| BR-C22 | US-1-03 | 履歴必須項目 |
| BR-C25 | Q-9 + requirements §2.4 | Mock 失敗時は補償なし |
| BR-C27 | requirements §2.4 | カテゴリは food のみ |
| BR-C29 | US-1-01 | 完了画面のテキスト |
| BR-C30 | US-1-07 | 5 秒後自動遷移 |

### 13.2 Plan Q&A 由来のルール

| BR-ID | Plan Q | 一行サマリ |
|---|---|---|
| BR-C01 | Q-1 | Bedrock リトライ 1 回 |
| BR-C02 | Q-2 | タイムアウト 1.5 秒 |
| BR-C06 | Q-3 | フォールバック分岐 5 件 |
| BR-C07 | Q-4 | 5 店舗ランダム |
| BR-C09 | Q-5 | suggestionId は再検証なし |
| BR-C10 | Q-6 | 失効時透過フォールバック |
| BR-C12 | Q-7 | Frontend 発行 ULID |
| BR-C13 | Q-8 | TTL 24h / userId スコープ |
| BR-C21 | Q-11 | GetHistory 仕様 |
| BR-C32〜C33 | Q-12 | エラー UX |

### 13.3 Comprehensive 深度で追加したルール

| BR-ID | 由来 | 一行サマリ |
|---|---|---|
| BR-C03 | Comprehensive | リトライ間 wait 0 |
| BR-C04 | Go 慣習 | Context キャンセル伝播 |
| BR-C08 | Comprehensive | 最頻判定の同点解決 |
| BR-C11 | Comprehensive | suggestionId 形式バリデ |
| BR-C14〜C15 | Comprehensive | 冪等命中時の応答整合 |
| BR-C17〜C18 | Comprehensive | 金額境界 / 整数円 |
| BR-C23 | Comprehensive | Source フィールド |
| BR-C26 | Comprehensive | 将来補償戦略 |
| BR-C28 | Comprehensive | 入力バリデーション順序 |
| BR-C31 | NFR-DEG-01 補強 | 完了画面任意タップ即遷移 |
| BR-C34 | UX 補強 | ボタン押下中表示 |
| BR-C35〜C36 | Comprehensive | 構造化ログ規約 |
| BR-C37〜C38 | NFR-PERF-01 接続 | レイテンシ予算 |

### 13.4 凍結 Interface 契約由来のルール

| BR-ID | 由来 | 一行サマリ |
|---|---|---|
| BR-C39 | unit-interfaces.md §4.3 | 409 IDEMPOTENCY_CONFLICT のハンドリング |

**総ルール数**: 39

---

## 14. 次ステージ（domain-entities.md / frontend-components.md）への引き継ぎ

- **domain-entities.md**: BR-C17 / BR-C18 / BR-C22 / BR-C23 を反映した型定義。`OrderRecord` / `PlaceOrderRequest` / `PlaceOrderResult` / `Plan` / `DeliveryInput` / `DeliveryOutput` を Comprehensive 深度で記述
- **frontend-components.md**: BR-C29 / BR-C30 / BR-C31 / BR-C32 / BR-C33 / BR-C34 を反映。MainScreen の GoroButton 部 / OrderCompletionScreen / `useOrder` hook の状態遷移
