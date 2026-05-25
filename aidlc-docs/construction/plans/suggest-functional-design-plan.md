# Unit D (`suggest`) — Functional Design Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / Functional Design
**Unit**: D — `suggest`（学習・先回り）
**Depth**: **Standard**（unit-of-work.md §6.2 確定）
**Prerequisite**: Inception 全成果物承認済み / 凍結契約 unit-interfaces.md / Unit C `order` 完了済み（develop マージ済み）
**Related**: 凍結契約 [unit-interfaces.md](../interfaces/unit-interfaces.md) §5・§7、横串デザインシステム [_design-system/design-spec.md](../_design-system/design-spec.md) §3.3（※ PR #94 worktree-add-desing、develop 未マージ）、Unit C [order/functional-design/](../order/functional-design/)

---

## 1. Plan の目的と範囲

Unit D `suggest`（学習・先回り）の **業務ロジック設計**（技術非依存）を Functional Design として確定する。Standard 深度として、GetSuggestion / ResolveSuggestion の 2 ユースケース、ドメインエンティティ、業務ルール、フロントコンポーネントを定義する。Plan 承認後、回答を反映した FD 成果物 4 種を生成する。

### 1.1 担当ストーリー（unit-of-work-story-map.md）

| Story | 概要 | Unit D の担当範囲 |
|---|---|---|
| US-2-01 | アプリを開いた瞬間に先回りサジェスト表示 | Bedrock 推論 + サジェストカード返却 |
| US-2-02 | 履歴が不十分なときはサジェストを出さない | 履歴件数による非表示判定 |
| US-2-03 | サジェストカードから 1 タップ注文（**Unit C 共同**） | `ResolveSuggestion` で保存済み plan を返す（Unit C が注文フローに合流） |
| US-2-04 | 行動履歴が蓄積されて学習される | Unit C 所有 OrderHistory の読取参照（Insert は Unit C） |
| US-X-02 | 自分より自分を知る AI に委ねる心地よさ | サジェスト精度・受諾速度（フェーズ2体験） |

### 1.2 入力（凍結契約・横串・他 Unit 連携）

| 入力 | 内容 |
|---|---|
| 凍結契約 §5.1 | `SuggestService`（`GetSuggestion` / `ResolveSuggestion`）、`Suggestion` / `SuggestionPlan` DTO |
| 凍結契約 §5.2 | `GET /api/suggest`（200 履歴十分 / 200 hasSuggestion:false） |
| 凍結契約 §5.3 | `GoroPay_Suggestion`（PK `suggestionId` / TTL `expiresAt` 30分 / `userId`,`plan`(JSON),`createdAt`） |
| 凍結契約 §5.4 | 内部依存: `OrderHistoryReader`（Unit C 読取）/ `BedrockAdapter`（横串）/ `FallbackSuggestProvider`（横串）/ `SuggestionStore`（内部） |
| 凍結契約 §7.1 | `BedrockAdapter.InferSuggestion(InferSuggestionInput) → InferSuggestionOutput{HasSuggestion, Title, Plan}` |
| 凍結契約 §7.3 | `FallbackSuggestProvider.BuildFromHistory(history) → *SuggestionPlan` |
| 凍結契約 §9 | Frontend `useSuggestion()` hook 型契約 |
| デザイン spec §3.3 | SUGGESTED 状態の表出（`SuggestBubble`「そろそろだろ。」/ ボタン主`押す。`/ 副`— {店名} ¥{金額} だ。`/ suggest-breath 2.4s） |
| Unit C 連携 | `OrderService` が `ResolveSuggestion(suggestionID)` を呼ぶ（Unit C BR-C09/C10: 保存値そのまま使用・失効時透過フォールバック） |

### 1.3 扱うこと / 扱わないこと

| 扱う | 扱わない（別ステージ/別 Unit） |
|---|---|
| GetSuggestion / ResolveSuggestion の業務ロジック擬似コード | Bedrock プロンプト本文の最終確定（→ NFR Design / Code Generation） |
| 履歴十分判定・抑制条件・フォールバック分岐の業務ルール | リトライ/タイムアウトの SLI 数値確定（→ NFR Requirements） |
| `Suggestion` / `SuggestionPlan` / `GoroPay_Suggestion` のドメイン定義 | DynamoDB 物理設計詳細・GSI（→ Infrastructure Design） |
| frontend-components.md（SuggestBubble/useSuggestion の構造・状態・API 統合） | CSS トークン・アニメ実装（→ デザインシステム実装プラン / Code Generation） |
| 凍結契約とデザイン spec の命名/文言の突合方針 | OrderHistory の Insert（Unit C 所有） |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順で実施する。

- [ ] §3 の確認質問（Q-DF1〜Q-DF10）にユーザが回答
- [ ] 回答の曖昧さ・矛盾を点検（特に契約 ⇄ デザイン spec の突合）、必要なら追加質問
- [ ] ユーザによる Plan 承認（Part 1 完了）
- [ ] `aidlc-docs/construction/suggest/functional-design/business-logic-model.md` 生成
- [ ] `aidlc-docs/construction/suggest/functional-design/business-rules.md` 生成
- [ ] `aidlc-docs/construction/suggest/functional-design/domain-entities.md` 生成
- [ ] `aidlc-docs/construction/suggest/functional-design/frontend-components.md` 生成
- [ ] Functional Design 完了メッセージ提示（2-option ゲート）
- [ ] aidlc-state.md / audit.md 更新、#65 とは別の本ブランチ（develop 起点）へコミット

---

## 3. 確認質問（Q-DF1 〜 Q-DF10）

> 各 `[Answer]:` に記号（A/B/C/D）を記入してください。**(推奨)** は調査に基づく既定案。「全部推奨で」で一括採用可。D は自由記述。

### —— 業務ルール / 履歴判定 ——

#### Q-DF1 — 「履歴十分」の判定閾値（US-2-02）
GetSuggestion が「履歴十分でサジェストを出す」と判定する閾値。Unit C の BR-C06（フォールバック分岐）は直近30日 5 件で統一されている。

- A) **直近 30 日の OrderHistory が 5 件以上でサジェスト生成、未満は出さない（Unit C BR-C06 と統一）** **(推奨)**
- B) 別閾値（[Answer] に件数・期間を記載）
- C) 件数ではなく「直近 N 日に注文あり」で判定
- D) Other

[Answer]:

#### Q-DF2 — 履歴不足時（閾値未満）の挙動
履歴が閾値未満のとき。Unit C の注文フローは `Default()`（架空 5 店舗）にフォールバックするが、サジェストは US-2-02 で「出さない」が要件。

- A) **`hasSuggestion=false` を返す（`Default()` フォールバックは使わない。US-2-02 準拠）** **(推奨)**
- B) `Default()` で適当なサジェストを出す（order と同じ挙動）
- C) Other

[Answer]:

#### Q-DF3 — サジェスト抑制条件（直近注文済み）
デザイン spec §3.2 は「直近 3 時間以内に同カテゴリ注文済みならサジェスト無効（IDLE 表示）」とする。これを Unit D バックエンドで判定するか。

- A) **バックエンドで「直近 3 時間以内に同カテゴリ注文あり」なら `hasSuggestion=false`（design spec 準拠、二重提案を防ぐ）** **(推奨)**
- B) 抑制しない（履歴十分なら常に出す。シンプル優先）
- C) フロント側で抑制（バックエンドは常に返す）
- D) Other（時間窓を変える等）

[Answer]:

### —— 連携 / Bedrock ——

#### Q-DF4 — `InferSuggestion` のタイムアウト・リトライ
横串 `BedrockAdapter` は Unit C / D 共有。Unit C（order）は 1.5s/回・1 リトライ（BR-C01/C02）。サジェストはマウント時取得で NFR-PERF-02（メイン画面 5 秒）管理であり、order ほど厳しくない。

- A) **order と同じ 1.5s/回・1 リトライ（共有 Adapter ポリシーの一貫性）** **(推奨)**
- B) サジェストは緩め（例: 2.0s/回・リトライなし、または 2 リトライ）
- C) NFR Requirements で数値確定（FD では「リトライあり/フォールバックあり」の方針のみ）
- D) Other

[Answer]:

#### Q-DF5 — Bedrock 失敗時のフォールバック
`InferSuggestion` がリトライ後も失敗したとき。

- A) **`FallbackSuggestProvider.BuildFromHistory(history)`（履歴最頻パターン）でサジェスト生成。`FallbackUsed=true`（内部ログのみ、API 応答には出さない）** **(推奨)**
- B) 即 `hasSuggestion=false`（フォールバックせず、サジェストを諦める）
- C) Other

[Answer]:

### —— ドメインモデル / 保存 ——

#### Q-DF6 — `GoroPay_Suggestion` の保存内容と TTL
凍結契約 §5.3: PK `suggestionId` / TTL `expiresAt`(30分) / `userId`,`plan`(JSON),`createdAt`。

- A) **`suggestionId`=ULID、TTL 30 分（契約準拠）、payload=`{userId, plan{storeName,menuName,amount,category}, createdAt}`。30 分は「起動直後の 1 タップ注文を想定した短命キャッシュ」根拠** **(推奨)**
- B) TTL を変更（[Answer] に値・根拠）
- C) Other

[Answer]:

#### Q-DF7 — `ResolveSuggestion` の挙動（Unit C が呼ぶ）
Unit C `OrderService` が 1 タップ注文時に `ResolveSuggestion(suggestionID)` を呼ぶ。Unit C 側は BR-C09（保存値そのまま使用）/ BR-C10（失効時は透過的に Bedrock フローへ）。

- A) **`suggestionId` で `GoroPay_Suggestion` を Get → あれば `SuggestionPlan` 返却 / TTL 失効・不在なら `nil` 返却（Unit C が透過フォールバック）。Bedrock 再検証なし（Unit C BR-C09 と整合）** **(推奨)**
- B) Other

[Answer]:

### —— フロントエンド ——

#### Q-DF8 — 凍結契約 ⇄ デザイン spec の hook/component 命名の突合
契約 §9 は `useSuggestion()`、デザイン spec §2.4/§4.4 は `useSuggest()`・`SuggestBubble`。unit-of-work.md は `SuggestionCard`・`useSuggestion`。

- A) **インターフェースは凍結契約を正（`useSuggestion()` / `GET /api/suggest`）。視覚は design spec §3.3 を正（GoroButton 内包の `SuggestBubble`「そろそろだろ。」/ ボタン主`押す。`/ 副`— {店名} ¥{金額} だ。`）。`SuggestionCard` は design spec の `SuggestBubble` として実装し、frontend-components.md にこの対応表を明記** **(推奨)**
- B) デザイン spec 側の命名（useSuggest）に寄せる（契約更新が必要）
- C) Other

[Answer]:

#### Q-DF9 — `Suggestion.Title` フィールドの扱い（契約 ⇄ デザイン spec の文言差）
契約 §5.1 例は `Title: "そろそろご飯めんどくさいですよね？"`、デザイン spec §3.3 吹き出しは固定 `そろそろだろ。`（README §3: 矛盾時デザイン spec 優先）。

- A) **API は `Title` を返すが、フロント表示は design spec の固定 `そろそろだろ。` を優先（design system 優先原則）。`Title` フィールドは将来の文言バリエーション（spec §9 オープン項目）用に契約維持** **(推奨)**
- B) API の `Title` をそのまま表示（design spec の吹き出しを動的化）
- C) Other

[Answer]:

#### Q-DF10 — サジェスト取得タイミング
デザイン spec §2.4「マウント時に 1 回 `GET /api/suggest`」。spec §9 で定期再評価は将来検討。

- A) **メイン画面マウント時に 1 回だけ `GET /api/suggest`（再評価・ポーリングなし。spec §2.4 準拠）。`useSuggestion` は staleTime を長めに設定し再フェッチを抑制** **(推奨)**
- B) 定期再評価あり（[Answer] に間隔）
- C) Other

[Answer]:

---

## 4. 矛盾チェック観点

回答収集後、以下を点検してから成果物生成へ進む。

1. Q-DF1 の閾値が Unit C BR-C06（5 件）と US-2-02 で整合するか
2. Q-DF3 の抑制条件が design spec §3.2 と矛盾しないか
3. Q-DF4 のタイムアウト方針が NFR Requirements に渡す前提として妥当か（数値確定は NFR-R）
4. Q-DF8/Q-DF9 の命名・文言突合が凍結契約とデザイン spec の双方に矛盾しないか（契約=IF、spec=視覚）
5. ResolveSuggestion の失効時 `nil` 返却が Unit C BR-C10 と整合するか
6. 「depends / たぶん / 標準で」等の曖昧回答がないか

---

## 5. 完了条件

- ユーザが Q-DF1〜Q-DF10 に回答（Part 1）
- 矛盾なしを確認しユーザが Plan を承認
- FD 成果物 4 種（business-logic-model / business-rules / domain-entities / frontend-components）を生成
- 2-option 完了ゲートで「Continue to Next Stage」または「Request Changes」を確認

---

## 6. 次ステージ（NFR Requirements）への引き継ぎ

FD 確定後、② NFR Requirements（Standard）へ。Q-DF4 のタイムアウト/リトライ方針は NFR Requirements で SLI 数値として確定。Bedrock 共有 Adapter のスロットリング方針は Unit C NFR と整合させる。デザインシステム（PR #94）がマージされ次第、frontend-components.md の `_design-system/design-spec.md` 参照が解決する。
