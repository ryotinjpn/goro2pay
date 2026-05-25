# Unit D (`suggest`) — Business Rules

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / Functional Design
**Unit**: D — `suggest`
**Depth**: Standard
**Related**: [business-logic-model.md](./business-logic-model.md)、[domain-entities.md](./domain-entities.md)、[frontend-components.md](./frontend-components.md)、凍結契約 [unit-interfaces.md](../../interfaces/unit-interfaces.md) §5・§7、[_design-system/design-spec.md](../../_design-system/design-spec.md) §3.3
**Aligned with**: Plan 回答 Q-DF1〜Q-DF10（全 A）

本ドキュメントは Unit D `suggest` の **業務ルール / 分岐条件 / 閾値 / コピー方針 / ログ規約** を ID 付き（`BR-D{番号}`）で管理する。各ルールに由来（Plan 質問 or Inception 成果物 or 凍結契約）を明記する。

---

## 1. 履歴判定に関するルール

### BR-D01: 「履歴十分」の判定閾値
- **由来**: Plan Q-DF1=A、US-2-02、Unit C BR-C06 と統一
- **ルール**: `OrderHistoryReader.ListRecent(userID, 30)` の結果のうち、`orderedAt` が **直近 30 日以内のものが 5 件以上**でサジェスト生成に進む。未満は出さない。
- **適用箇所**: business-logic-model.md §2.1 STEP 2

### BR-D02: 履歴不足時は非表示（Default を使わない）
- **由来**: Plan Q-DF2=A、US-2-02
- **ルール**: BR-D01 を満たさない場合 `Suggestion{HasSuggestion: false}` を返す。Unit C の `Default()`（架空 5 店舗）フォールバックは **使わない**（サジェストは履歴学習が価値であり、的外れな先回りはしない）。
- **適用箇所**: business-logic-model.md §2.1 STEP 2

### BR-D03: 直近注文済みのサジェスト抑制
- **由来**: Plan Q-DF3=A、design spec §3.2
- **ルール**: 履歴十分でも、**直近 3 時間以内に同カテゴリ（"food"）の注文が存在**する場合は `HasSuggestion: false` を返す（直後の二重提案を防ぐ）。判定はバックエンド（注文時刻を持つ Unit D 側）で行う。
- **ログ**: `event="suggest_suppressed_recent_order"`
- **適用箇所**: business-logic-model.md §2.1 STEP 3

---

## 2. Bedrock 推論・フォールバックに関するルール

### BR-D04: InferSuggestion のタイムアウト・リトライ
- **由来**: Plan Q-DF4=A、Unit C BR-C01/BR-C02 と統一（共有 BedrockAdapter）
- **ルール**: `BedrockAdapter.InferSuggestion` は **1 回あたり 1.5 秒タイムアウト**、失敗時 **1 回リトライ**（計 2 回、待機ゼロ）。永続エラー（ValidationException 等）は 1 回目で即フォールバック。
- **補足**: SLI（target/threshold）の数値確定は NFR Requirements ステージで追認。サジェストはマウント時取得で NFR-PERF-02（メイン画面 5 秒）管理。
- **適用箇所**: business-logic-model.md §2.2

### BR-D05: Bedrock 失敗時のフォールバック
- **由来**: Plan Q-DF5=A
- **ルール**: 2 連続失敗時、`FallbackSuggestProvider.BuildFromHistory(history)`（履歴の最頻パターン）で plan を生成する。BR-D01 により履歴は 5 件以上あるため、最頻パターンを必ず返せる。
- **ログ**: `event="suggest_bedrock_fallback"`
- **適用箇所**: business-logic-model.md §2.1 STEP 5

### BR-D06: 最頻生成も不可なら非表示
- **由来**: Plan Q-DF5=A の補足（防御的）
- **ルール**: `BuildFromHistory` が `nil` を返す異常時（理論上ほぼ起きない）は `HasSuggestion: false` を返す。

### BR-D11: Resolve 時の Bedrock 再検証なし
- **由来**: Plan Q-DF7=A、Unit C BR-C09 と整合
- **ルール**: `ResolveSuggestion` は保存済み `SuggestionPlan` をそのまま返す。Bedrock 再呼び出しはしない（NFR-PERF-01 / NFR-DEG-01 を優先、提案はユーザ表示時点で確認済みとみなす）。

### BR-D12: FallbackUsed は内部ログのみ
- **由来**: 凍結契約 §5.1（`FallbackUsed` コメント「内部ログ用、API レスポンスには出さない」）
- **ルール**: `Suggestion.FallbackUsed` は構造化ログ・メトリクス用。`GET /api/suggest` の JSON レスポンスには含めない（ユーザに AI 失敗を悟らせない、透過性）。

---

## 3. 保存・復元に関するルール

### BR-D07: suggestionId は ULID
- **由来**: 凍結契約 §5.3（PK `suggestionId`）、ULID 仕様
- **ルール**: `suggestionId` は サーバ側（Unit D）で `ulid.Make()` により採番。Crockford Base32 26 文字。時系列ソート可。

### BR-D08: Suggestion の TTL は 30 分
- **由来**: Plan Q-DF6=A、凍結契約 §5.3（TTL `expiresAt` 30分）
- **ルール**: `GoroPay_Suggestion` に `expiresAt = now + 30分`（Unix epoch）を設定。DynamoDB TTL 機能で自動削除。「起動して提案を見て、少し放置してから押す」程度の猶予をカバーする短命キャッシュ。
- **適用箇所**: business-logic-model.md §2.1 STEP 6

### BR-D09: 保存 payload の項目
- **由来**: Plan Q-DF6=A、凍結契約 §5.3
- **ルール**: `GoroPay_Suggestion` に保存するのは `suggestionId`(PK) / `userId` / `plan`(JSON: storeName,menuName,amount,category) / `createdAt` / `expiresAt`(TTL)。

### BR-D10: ResolveSuggestion の失効時は nil
- **由来**: Plan Q-DF7=A、Unit C BR-C10 と整合
- **ルール**: `SuggestionStore.Get(suggestionID)` が `nil`（TTL 失効 or 不在）を返したら、`ResolveSuggestion` も `nil` を返す。Unit C が透過的に通常 Bedrock 注文フローへフォールバックする。
- **ログ**: `event="suggestion_expired_or_missing"`

---

## 4. フロントエンド・コピーに関するルール

### BR-D13: Title はフロント表示で固定文言を優先
- **由来**: Plan Q-DF9=A、design system README §3（矛盾時 design spec 優先）
- **ルール**: API（`Suggestion.Title`）は凍結契約どおり値を返す（例「そろそろご飯めんどくさいですよね？」）が、**フロント表示は design spec §3.3 の固定「そろそろだろ。」を優先**する。`Title` フィールドは将来の文言バリエーション（spec §9 オープン項目）用に契約として保持。
- **適用箇所**: frontend-components.md `SuggestBubble`

### BR-D14: サジェスト取得はマウント時 1 回
- **由来**: Plan Q-DF10=A、design spec §2.4
- **ルール**: フロントは メイン画面マウント時に **1 回だけ** `GET /api/suggest`。定期再評価・ポーリングはしない。`useSuggestion` は staleTime を長めに設定し再フェッチを抑制。

### BR-D15: hook/コンポーネント命名の突合
- **由来**: Plan Q-DF8=A
- **ルール**: **インターフェース名は凍結契約を正**（`useSuggestion()` / `GET /api/suggest` / `Suggestion` / `SuggestionPlan`）。**視覚コンポーネントは design spec を正**（`SuggestBubble`、GoroButton 内蔵）。unit-of-work.md の `SuggestionCard` は `SuggestBubble` として実装する。対応表は frontend-components.md に明記。

### BR-D16: サジェスト時のボタン表出
- **由来**: design spec §3.3
- **ルール**: `hasSuggestion=true` のとき、GoroButton 主ラベル → `押す。`、副ラベル → `— {storeName} ¥{amount} だ。`、ボタン上に `SuggestBubble`「そろそろだろ。」。`suggest-breath`（2.4s 明滅）。`hasSuggestion=false` のときは通常の「めんどくさい」表示。

---

## 5. カテゴリ・セキュリティ・ログに関するルール

### BR-D17: 対応カテゴリは "food" のみ（MVP）
- **由来**: requirements.md §2.4、Unit C BR-C27 と整合
- **ルール**: MVP のサジェスト対象は `category == "food"` のみ。履歴の他カテゴリは推論入力に含めるが、提案・抑制判定は food を対象とする。将来拡張は凍結契約の拡張余地に従う。

### BR-D18: Bedrock への PII 非送出
- **由来**: NFR-SEC、Unit C と同方針
- **ルール**: `InferSuggestion` のプロンプトには履歴の要約（category/storeName/menuName/amount/orderedAt）のみ送る。email 等の PII は送らない（userId も推論本文には含めない）。

### BR-D19: 構造化ログのフィールド
- **由来**: NFR-OBS-01、Unit C BR-C35 と同方針
- **ルール**: Unit D のログは最低限 `event` / `userId` / `severity` を含む。サジェスト固有として `suggestionId`（採番後）/ `fallbackUsed` / `historyCount` を必要に応じ付与。`idempotencyKey` は Unit D では扱わない。PII は出さない。

### BR-D20: ログ出力対象イベント一覧

| イベント | severity | タイミング |
|---|---|---|
| `suggest_requested` | info | GetSuggestion 受領時 |
| `suggest_insufficient_history` | info | 履歴 < 5 件で非表示時（BR-D02） |
| `suggest_suppressed_recent_order` | info | 直近 3h 注文で抑制時（BR-D03） |
| `suggest_bedrock_fallback` | warn | Bedrock 2 連続失敗 → フォールバック時（BR-D05） |
| `suggest_served` | info | 200 でカード返却時 |
| `suggestion_expired_or_missing` | warn | Resolve 時に失効/不在検知時（BR-D10） |

---

## 6. ルール一覧サマリ

### 6.1 Plan Q&A 由来

| BR-ID | Plan Q | 一行サマリ |
|---|---|---|
| BR-D01 | Q-DF1 | 履歴十分 = 直近30日5件以上 |
| BR-D02 | Q-DF2 | 履歴不足は非表示（Default 不使用） |
| BR-D03 | Q-DF3 | 直近3h同カテゴリ注文で抑制 |
| BR-D04 | Q-DF4 | InferSuggestion 1.5s×1リトライ |
| BR-D05 | Q-DF5 | 失敗時 BuildFromHistory |
| BR-D07〜D09 | Q-DF6 | suggestionId ULID / TTL30分 / payload |
| BR-D10 | Q-DF7 | Resolve 失効時 nil |
| BR-D11 | Q-DF7 | Resolve は Bedrock 再検証なし |
| BR-D13 | Q-DF9 | Title は表示で固定文言優先 |
| BR-D14 | Q-DF10 | マウント時 1 回取得 |
| BR-D15 | Q-DF8 | 命名突合（IF=契約 / 視覚=spec） |

### 6.2 凍結契約 / Inception / design spec 由来

| BR-ID | 由来 | 一行サマリ |
|---|---|---|
| BR-D06 | Q-DF5 補足 | 最頻も不可なら非表示 |
| BR-D12 | 契約 §5.1 | FallbackUsed は内部ログのみ |
| BR-D16 | design spec §3.3 | サジェスト時ボタン表出 |
| BR-D17 | requirements §2.4 | カテゴリ food のみ |
| BR-D18 | NFR-SEC | Bedrock へ PII 非送出 |
| BR-D19〜D20 | NFR-OBS-01 | 構造化ログ規約 |

**総ルール数**: 20

---

## 7. 次ドキュメントへの引き継ぎ

- **domain-entities.md**: BR-D07〜D09 を反映した `SuggestionRecord`（GoroPay_Suggestion）、契約 §5.1 の `Suggestion`/`SuggestionPlan`、§7.1 `InferSuggestionInput/Output` の型・制約
- **frontend-components.md**: BR-D13〜D16 を反映した `useSuggestion` / `SuggestBubble` の設計、契約⇄design spec 命名対応表
