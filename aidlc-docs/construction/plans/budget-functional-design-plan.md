# Unit B `budget` — Functional Design Plan

**Document Version**: 1.0
**Created**: 2026-05-21
**Unit**: B (`budget` / ダメ予算)
**Construction Depth**: Standard
**Stage**: Functional Design (Construction Phase, per-unit)

---

## 0. このドキュメントの目的

Unit B `budget` の Functional Design ステージで生成する成果物（business-logic-model.md / business-rules.md / domain-entities.md / frontend-components.md）の前段として、ビジネスロジックを確定するために必要な意思決定を質問形式で収集する。

ヒアリング方針: チームのフィードバックメモリ「対話形式でのヒアリング」に従い、本 Plan ファイルは **監査用の記録** として作成し、実際の質問はチャットで 1 問ずつ提示する。回答受領のたびに本ファイルの `[Answer]:` を更新する。

---

## 1. Plan 実施チェックリスト

- [x] Step 1: Unit B の前段成果物を読み込み済み（unit-of-work.md / unit-of-work-story-map.md / components.md / component-methods.md / services.md / requirements.md / stories.md）
- [x] Step 2: 業務ロジック上の確認事項を質問形式で抽出（Q-1〜Q-12）
- [x] Step 3: 質問を本 Plan ファイルに記載
- [x] Step 4: チャットで 1 問ずつ提示してヒアリング
- [x] Step 5: 回答を `[Answer]:` タグへ反映、矛盾検出（Q-B11 補足の数値刻みのみ追加質問、γ で確定）
- [x] Step 6: 矛盾解消後、Functional Design 成果物の生成準備完了

---

## 2. Unit B が担当するストーリー（再掲）

| ID | ストーリー要約 | 主要関心 |
|---|---|---|
| US-0-03 | ダメ予算を設定する | 予算入力・保存・初期化 |
| US-0-04 | 初回メイン画面を表示する（残高初期表示部分） | 初期化済み残高の参照 |
| US-1-02 | 残高を常に視認できる | 残高取得・最新性 |
| US-1-04 | 残高不足時に「今月ダメになれません」を知る | 残高不足判定・遷移 |
| US-1-05 | 二重引き落としを防ぐ | 冪等性キー処理 |
| US-1-06 | 残高が負にならない | 条件付き減算 |
| US-3-05 | 月初にダメ予算が自動リセットされる | スケジューラ・履歴記録 |

---

## 3. 質問

> ヒアリングは下記の Q 順にチャットで 1 問ずつ実施する。回答ごとに `[Answer]:` を更新する。

### Q-B1: ダメ予算の設定可能範囲

`SetBudget` で許容する月間予算の範囲を確定したい。Application Design では `1 ≤ monthlyBudget ≤ 100,000` 円と記載があり、`BudgetRaiseService` は `min(current * 1.5, 100_000)` で増額推奨を計算する。この境界値の取り扱いを Functional Design として正式化する。

A) `1 ≤ monthlyBudget ≤ 100,000`（Application Design に記載の範囲をそのまま採用、上限到達時は `RaiseModal` で「もう増額できません」表示）
B) `1,000 ≤ monthlyBudget ≤ 100,000`（最低 1,000 円、ペルソナ初期 30,000 円との整合性重視）
C) `1 ≤ monthlyBudget ≤ 100,000` かつ 100 円単位に丸め込み（UI 入力の刻みも 100 円単位）
D) 上限なし（1 円以上であれば上限制限なし、退化ループの極端化を許容）
E) Other (please describe after [Answer]: tag below)

[Answer]: A

---

### Q-B2: `SetBudget` の即時反映ルール（初回 vs 変更）

Application Design / services.md では「初回設定時のみ Wallet を即座に初期化、以降は翌月リセットから適用」とある。これを以下のように具体化する。

A) **初回**: `BudgetSettings` も `Wallet.Balance` も即座に予算額で作成。**変更**: `BudgetSettings.MonthlyBudget` のみ更新、`EffectiveFrom = 翌月1日 00:00 JST`、当月 `Wallet.Balance` は据え置き
B) **初回**: 同上。**変更**: `BudgetSettings` を即時更新（`EffectiveFrom = now`）、当月の `Wallet.Balance` も差分調整（増額分加算 / 減額時は max(現残高, 新予算) で打ち切り）
C) **初回**: 同上。**変更**: `BudgetSettings.MonthlyBudget` を即時更新、当月 `Wallet.Balance` は据え置き（`EffectiveFrom` 概念を使わず、翌月リセット時に最新値が単純適用）
D) Other (please describe after [Answer]: tag below)

[Answer]: B（ダメ化UX 整合: 増額分は当月残高に即時加算、減額時は max(現残高, 新予算) で打ち切り）

---

### Q-B3: 「初回判定」のロジック

Q-B2 で「初回のみ Wallet を即時初期化」が選ばれた場合、初回判定の根拠を確定する。

A) `BudgetSettingsRepository.Get(userID)` が `nil`/NotFound を返した場合を初回とする（BudgetSettings の存否で判断）
B) `WalletRepository.Get(userID)` が `nil`/NotFound を返した場合を初回とする（Wallet の存否で判断）
C) A と B の AND（両方が NotFound）。片方だけ存在する場合はエラーログを出して整合性回復
D) Other (please describe after [Answer]: tag below)

[Answer]: B（Wallet の存否で初回判定。異常系ケース 3 でユーザに障害露呈せず暗黙リカバリ可能、開発速度も A と同等）

---

### Q-B4: 冪等性キーの仕様（`IdempotencyRepository.TryAcquire`）

`Deduct` が二重引き落としを防ぐ冪等性キーの扱い。以下を確定する。

A) **キー形式**: `{userID}:{clientGeneratedUlid}`、**TTL**: 24 時間、**衝突検出**: 同一キー × 異なる payload (amount) で `ErrIdempotencyConflict`、同一 payload なら前回結果を返す
B) **キー形式**: クライアント生成の ULID 単独（userID は別属性）、**TTL**: 24 時間、衝突検出は A と同じ
C) **キー形式**: `{userID}:{clientGeneratedUUID}`、**TTL**: 1 時間（短め、デモ用）、衝突検出は A と同じ
D) Other (please describe after [Answer]: tag below)

[Answer]: A（キー: `{userID}:{clientGeneratedUlid}`、TTL 24h、ULID 時系列ソート可、PartitionKey 単独で衝突検知）

---

### Q-B5: `Deduct` 失敗時の冪等性レコードの扱い

`Deduct` で `IdempotencyRepository.TryAcquire` が `acquired=true` を返したものの、その後の `WalletRepository.DeductConditional` が `ErrInsufficientBalance` 等で失敗した場合の処理。

A) 冪等性レコードは TryAcquire 直後にコミット済みとする。失敗結果（エラーコード）も payload として保存し、同一キーの再送に対しては「同じエラー」を返す（リトライ抑制）
B) 失敗時は冪等性レコードを削除（コンペンセイション）し、同一キーで再試行可能にする（リトライ許容）
C) Hybrid: `ErrInsufficientBalance`（業務エラー）は A、`ErrBedrockUnavailable` 系の一時的エラーは B
D) Other (please describe after [Answer]: tag below)

[Answer]: A（失敗結果を payload に保存し、同一キーの再送には保存通りに返す。race condition リスク回避、冪等性の本質に整合）

---

### Q-B6: `GetBalance` のキャッシング / 鮮度

US-1-02「残高を常に視認できる」「stale state 禁止」の運用ルールを確定する。

A) **バックエンド**: `GetBalance` は毎回 DynamoDB から ConsistentRead で取得、キャッシュなし。**フロント**: TanStack Query で 30 秒間 fresh、`Deduct` 成功時は即座に invalidate
B) バックエンドは Eventually Consistent Read で OK（コスト優先）、フロントは A と同じ
C) バックエンド: ConsistentRead、フロント: 注文/予算変更後 `useWallet().refetch()` を明示的に呼び出して即時更新（自動 invalidate なし）
D) Other (please describe after [Answer]: tag below)

[Answer]: A（バックエンド ConsistentRead、フロントは TanStack Query 30s fresh + Deduct/SetBudget 成功時に自動 invalidate）

---

### Q-B7: 残高不足時の判定主体（注文の文脈）

US-1-04「残高不足時にダメになれない」の判定責任を確定する。Unit B か Unit C どちらが主体か。

A) **Unit B** が責任主体: `WalletService.Deduct` が `ErrInsufficientBalance` を返す。Unit C はそれを受けて `402 INSUFFICIENT_BALANCE` レスポンス。事前 GetBalance チェックは不要
B) **Unit C** が事前チェック: `OrderService.PlaceOrder` 開始時に `WalletService.GetBalance` で残高確認、不足なら即時 402。`Deduct` の `ErrInsufficientBalance` は二重防御
C) ハイブリッド: Unit B が主体、Unit C は推論された Bedrock 推奨額が予測残高を超える場合に Bedrock リトライで再生成を試みる（ダメ化UXの観点で「成功体験」優先）
D) Other (please describe after [Answer]: tag below)

[Answer]: A（Unit B 主体、ConditionExpression で race condition 耐性、Unit C は ErrInsufficientBalance を受けて 402 応答）

---

### Q-B8: 月初リセットの実行タイミングと粒度

US-3-05 / Application Design の `MonthlyResetHandler` 仕様を確定する。

A) **トリガ**: EventBridge Scheduler、`cron(0 15 L * ? *)` UTC = JST 翌月1日 00:00。**処理**: 全ユーザを 1 Lambda 呼び出しで逐次処理（並列度 1）。**冪等性**: `BudgetResetLog` の `(ResetDate, UserID)` を主キーで条件付き Insert、既存ならスキップ
B) **トリガ**: 同上。**処理**: 全ユーザリストをチャンク化し、SQS で並列 worker（複数 Lambda）に分散。**冪等性**: 同上
C) **トリガ**: 同上だが手動再実行可能（CLI/管理画面）。**処理**: 並列度 1。**冪等性**: 同上 + 失敗ユーザの再処理対応
D) Other (please describe after [Answer]: tag below)

[Answer]: A（EventBridge Scheduler `cron(0 15 L * ? *)` UTC、1 Lambda 逐次処理、(ResetDate, UserID) 複合キーで条件付き Insert、Lambda 手動 invoke で再実行代用）

---

### Q-B9: 月初リセット時の予算ソース（`EffectiveFrom` の扱い）

リセット時に各ユーザに適用する `monthlyBudget` の取得ロジック。

A) `BudgetSettings.MonthlyBudget` を単純に読み出して適用（`EffectiveFrom` は記録のみで判定には使わない）
B) `BudgetSettings.RaiseHistory` の中から `EffectiveFrom <= リセット日` の最新エントリを採用（厳密適用）
C) `BudgetSettings.MonthlyBudget` の現在値を採用するが、`EffectiveFrom > リセット日` の場合は前回予算を使用（変更途中の特殊ケース対応）
D) Other (please describe after [Answer]: tag below)

[Answer]: A（`BudgetSettings.MonthlyBudget` を素直に読み出して適用、`RaiseHistory` 内の `EffectiveFrom` は監査メタデータとして記録のみ。Q-B2=B の即時反映モデルとの整合）

---

### Q-B10: リセット時の前残高（前月残高）の扱い

リセット直前の残高（前月の残額）の扱い。

A) **完全リセット**: 前残高は `BudgetResetLog` に記録するのみで、新残高は `monthlyBudget` で上書き（持ち越しなし）
B) **持ち越し**: 前残高があれば `新残高 = monthlyBudget + 前残高`（残額 carry-over）。ダメ化UX 的には NG（消化を促進したい）
C) **赤字計算**: 月内に発生した負債（万一発生時）は次月予算から差し引く。本 MVP はそもそも残高不変条件で負にならないため A と同等
D) Other (please describe after [Answer]: tag below)

[Answer]: A（完全リセット、前残高は BudgetResetLog に記録のみ。ダメ化UX = 消化促進の思想と整合、節約を肯定しない）

---

### Q-B11: BudgetSetupScreen の入力 UX とバリデーション

US-0-03 のフロント側コンポーネントの仕様。

A) 数値入力（HTML `<input type="number">`）+ 100 円刻みの 5 個程度のクイックボタン（10,000 / 30,000 / 50,000 / 80,000 / 100,000）+ 範囲外時はリアルタイムエラー表示。「ダメ予算を設定する」CTA ボタンで POST
B) スライダー（1,000 〜 100,000、刻み 1,000）+ 入力された値を大きく表示。CTA は同じ
C) 数値入力のみ（クイックボタンなし）、シンプル UI、バリデーションは送信時に確認
D) Other (please describe after [Answer]: tag below)

[Answer]: A（数値入力 + クイックボタン 10k/30k/50k/80k/100k、リアルタイム範囲外エラー、CTA「ダメ予算を設定する」。NFR-DEG-01 低摩擦オンボーディング整合）

**追加確認 (Q-B11 補足)**: 数値入力の刻みは **γ = 1,000 円刻み**（`<input step="1000">`、モバイル UI 簡素化）。
- Q-B1=A の上限 100,000 円を 1,000 円刻みで割り切れる（100 段階）
- ペルソナ初期 30,000 円も 1,000 円刻みで自然に表現可能
- 1 円刻みは細かすぎ、100 円刻みは中途半端、1,000 円刻みがモバイル UI に最適
- バックエンド `SetBudget` のバリデーションも `monthlyBudget % 1000 == 0` を追加

---

### Q-B12: BalanceDisplay コンポーネントの表示仕様

US-1-02 / NFR-DEG-03「残高常時可視化」の表示仕様。

A) MainScreen 上部に **¥XX,XXX**（フォントサイズ大、色は通常黒、消化率 80% 超で赤）+ 「残りダメ予算」ラベル + リセット日カウントダウン（あと N 日）
B) MainScreen 上部に **¥XX,XXX** のみ（消化率 80% 超で赤）。ラベルやカウントダウンは別パネル
C) MainScreen 上部にプログレスバー（消化率視覚化）+ 残高と予算 `¥X / ¥Y` 形式 + 80% 超で赤
D) Other (please describe after [Answer]: tag below)

[Answer]: A（**¥XX,XXX** 大フォント + 「残りダメ予算」ラベル + リセット日カウントダウン、消化率 80% 超で赤。退化ループ動線として「あと N 日で失効」圧力を演出、Unit E MetricsPanel との役割分担明確）

---

## 4. 矛盾検出ルール（Step 5 で適用）

回答取得後、以下の論理整合をチェックする:

- Q-B2 が C（EffectiveFrom 不使用）かつ Q-B9 が B（EffectiveFrom 厳密適用）→ 矛盾
- Q-B5 が B（失敗時削除）かつ Q-B7 が A（事前チェックなし）かつ業務エラーでもリトライ許容 → 二重引き落としリスク（再確認）
- Q-B10 が B（持ち越し）かつ プロジェクト方針の「ダメ化UX を最優先」と齟齬 → 再確認
- Q-B1 の上限と Q-B11 のクイックボタン値の整合
- Q-B11 が A（100 円刻み）かつ Q-B1 が C 以外 → 端数の扱いを再確認

---

## 5. 完了条件

- 全 12 問に回答済みかつ `[Answer]:` タグが空でない
- 矛盾検出ルールに引っかかる組み合わせがない（あれば clarification Q を追加発行）
- 完了後、Functional Design 成果物（business-logic-model.md / business-rules.md / domain-entities.md / frontend-components.md）の生成へ進む
