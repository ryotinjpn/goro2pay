# ゴロゴロPay メイン画面 — デザイン仕様

**Document Version**: 1.0
**Created**: 2026-05-24
**Scope**: Web (PWA) のメイン画面 1 枚に集中 (`web/app/page.tsx` の認証後ビュー)
**Concept**: Slot Machine
**Source**: `aidlc-docs/idea.md`, `aidlc-docs/inception/requirements/requirements.md`, `aidlc-docs/inception/user-stories/personas.md`

---

## 0. 背景

ゴロゴロPay は「めんどくさい」を 1 ボタンで即時キャッシュ解決し、ユーザを快適にダメ化させていく金融グループ発のライフスタイル代行アプリ。requirements.md `§4.1 ダメ化UX` を最優先 NFR とし、ユーザストーリーは太郎ペルソナのフェーズ 1 → 2 → 3 の感情変遷に沿って組まれている。

本 spec はそのユーザ体験の要となる「メイン画面 1 枚」を、現在の暫定 UI (`GoroButton` のオレンジ角丸ボタン) から、コンセプトを体現するデザインへ刷新するための仕様を定める。

---

## 1. コンセプト & デザイントークン

### 1.1 コンセプトステートメント

> **「Slot Machine」** — 金融グループ発の、手のひらサイズのスロット筐体。残高は気品ある運命の数字としてセリフ書体で鎮座し、金色のレバー型ボタンが光ってこちらを誘惑する。リールが回り、面倒が消え、また数字が削れていく。

差別化ポイント：
- スロット演出 (押す → リール回転 → 停止) は太郎ペルソナのフェーズ 2 独白「アプリを開く前から何が出るだろうと期待する (パチンコ化)」を直接視覚化する。
- 主役カラーを真鍮ゴールドに置き、赤は「警告・退化の可視化」だけに温存することで、ハッカソン審査軸の「金融グループ発」の品格を担保する。
- 「Slot Machine」はゲーム筐体・ラスベガス由来の一般語彙として日本でも全年齢に通る無毒な比喩。

### 1.2 カラートークン

| トークン | 値 | 用途 |
|---|---|---|
| `bg-deep` | `#050402` | 画面最暗部 |
| `bg-mid` | `#0e0a06` | ベース背景 |
| `bg-warm` | `#1a1208` | 上部グラデのハイライト |
| `gold-100` | `#f4d990` | ボタン上部ハイライト |
| `gold-500` | `#c9a96b` | **主役カラー (ボタン本体・¥ 記号・ライン)** |
| `gold-700` | `#8a6f3a` | ボタン陰影 |
| `gold-900` | `#3a2a14` | ボタン最暗部 |
| `cream` | `#f4ecd8` | 残高数字・主要白文字 |
| `paper` | `#ffe7c2` | 副次テキスト |
| `mute` | `#8a7a5a` | キャプション・小さいラベル |
| `accent-warn` | `#ffb14a` | 消化 60–80% |
| `accent-danger` | `#ff4f4f` | 80% 超・残高低下警告・DEAD 時の増額ボタン |
| `dead-gray` | `#6a6a6a` | DEAD 状態の非アクティブ要素 |

実装は CSS variables (`--color-bg-deep` 形式) で `web/app/globals.css` に定義する。

### 1.3 タイポトークン

| 役割 | フォント | 使用箇所 |
|---|---|---|
| `display-serif` | Cormorant Garamond / Italic 700 | 残高数字・サブコピー (`— 押せ。考えるな。`) |
| `body-mincho` | Zen Old Mincho / 700–900 | ブランドマーク `ゴロゴロ`・主要見出し |
| `mono-pixel` | DotGothic16 / 400 | キャプション・LCD ラベル・ティッカー |
| `body-sans` | Inter / 600–800 | 補助的な英数字 (時刻・% 表記) |

**カタカナ字形のチューニング**: ブランドマークは `letter-spacing: 0.04em–0.06em` を適用し、濁点が潰れないよう字間をやや開ける。ロゴ向け大判ウェイト (Zen Old Mincho 900) と UI チップ向け小判ウェイトを字形上で分け、UI 上部の小チップは過度な装飾 (`▼ ▼` 等) を**付けない**。

### 1.4 スペーシング & 形状

- **コーナー半径**: ボタンは円 (`borderRadius: 50%`)、カード/モーダルは `borderRadius: 18px`
- **画面パディング**: `28px 22px 24px` (モバイル基準)
- **ボタンサイズ**: 画面幅の 78%、`aspect-ratio: 1`
- **CRT スキャンライン**: `repeating-linear-gradient(0deg, rgba(255,255,255,.025) 0 1px, transparent 1px 4px)` を全画面に薄く適用

### 1.5 モーション原則

- **基本トランジション**: `200ms ease`
- **スロット回転**: `0.4s linear infinite` の strip スクロール、停止時に 1 回 overshoot
- **残高減算**: `450ms cubic-bezier(.2,.8,.2,1)` で数字を旧 → 新へ縦スクロール
- **DEAD 遷移**: `filter: saturate(.2) brightness(.7)` を `600ms ease-out` で適用
- **モーション軽減**: `@media (prefers-reduced-motion: reduce)` でアニメーションを `0.01ms` まで短縮

---

## 2. コンポーネント分解とレイアウト構造

### 2.1 全体レイアウト (モバイル基準)

```
┌─────────────────────────────┐
│  [HEADER] ゴロゴロPay  21:07 │ ← BrandHeader
│                             │
│        残りダメ予算           │ ← BalanceLabel
│        ¥ 28,800             │ ← BalanceHero
│        ────────────         │ ← MeterBar
│        今月 5 回   消化 4%   │ ← MetricsRow
│                             │
│         ┌───────┐           │
│         │       │           │ ← GoroButton
│         │ めんどくさい │     │
│         │       │           │
│         └───────┘           │
└─────────────────────────────┘
```

### 2.2 コンポーネント一覧

| コンポーネント | 役割 | 既存 | 改修区分 |
|---|---|---|---|
| `MainScreen` (`app/page.tsx`) | 認証後のメイン画面コンテナ | あり | **改修** |
| `ScreenFrame` | CRT スキャンライン＋背景グラデを与えるラッパ | なし | **新規** |
| `BrandHeader` | 上部ヘッダー (`ゴロゴロPay` + 時刻) | なし | **新規** |
| `BalanceHero` | 残高表示 (ラベル＋金額＋メーター＋メトリクス) | なし | **新規** |
| `GoroButton` | 中央のメイン CTA。状態: `idle / suggested / slot / dead` | あり | **改修** |
| `SuggestBubble` | サジェスト時にボタン上に乗る吹き出し | なし | **新規 (GoroButton 内部)** |
| `SlotReel` | スロット押下時に現れるリール窓 | なし | **新規 (GoroButton 内部)** |
| `DeadVerdict` | 残高 0 時のターミネーター画面オーバーレイ | なし | **新規** |
| `IncreaseBudgetButton` | DEAD 時の赤い増額 CTA | なし | **新規** |
| `OrderHistoryList` | 履歴 | あり | **メイン画面から外す** (コンポーネントは温存) |

### 2.3 状態モデル

```
idle ──[履歴あり & サジェストあり]──→ suggested
  │                                    │
  └──[ボタン押下]─────────────────────→ slot ──[成功]──→ idle (残高更新)
                                       │
                                       └──[残高不足]──→ dead

dead ──[増額決定 & 月初]──→ idle
```

明示的に管理する状態:

- `screenState: 'idle' | 'suggested' | 'slot' | 'dead'`
- `balance: number` (リアルタイム残高、jotai atom)
- `monthlyCount: number` (今月のダメ化回数)
- `consumeRate: number` (0–1、消化率)

`screenState` はルート要素の `data-screen-state` 属性として CSS に渡し、状態に応じたスタイル分岐は CSS セレクタ側で行う。

**注**: §3.1 で扱う EMPTY (履歴なし初回) は独立した状態ではなく、`screenState === 'idle' && monthlyCount === 0` の派生表示。サジェストが出ないだけで挙動は IDLE と同一。

### 2.4 データフロー

- 認証: `useAuth()` (既存) — そのまま流用
- 注文: `useOrder()` (既存 react-query mutation) — `onMutate` で `screenState = 'slot'`、`onSuccess` で `idle` 復帰、`onError` の残高不足エラーで `dead` 遷移
- 残高: 新規 `useBalance()` hook (react-query で `GET /balance` を取得し、注文 mutation の `onSuccess` で invalidate)
- サジェスト: 新規 `useSuggest()` hook (メイン画面マウント時に `GET /suggest` を 1 回叩く)

### 2.5 階層図

```
<MainScreen>
  <ScreenFrame>
    <BrandHeader />
    <BalanceHero />
    <GoroButton>
      <SuggestBubble />     ← suggested 状態でだけ表示
      <SlotReel />          ← slot 状態でだけ表示
    </GoroButton>
    <DeadVerdict>           ← dead 状態でだけ表示 (オーバーレイ)
      <IncreaseBudgetButton />
    </DeadVerdict>
  </ScreenFrame>
</MainScreen>
```

---

## 3. 各状態の詳細仕様

### 3.1 EMPTY (履歴なし初回)

**遷移条件**: 認証済み & `monthlyCount === 0` & サジェスト履歴なし

| 要素 | 表示 |
|---|---|
| BrandHeader | `ゴロゴロPay` ／ 現在時刻 |
| BalanceHero ラベル | `残りダメ予算` |
| BalanceHero 数字 | `¥30,000` (初期値) |
| MeterBar | 0% (`gold-500` 線のみ) |
| MetricsRow | `今月 0 回` ／ `消化 0%` |
| GoroButton ラベル主 | `めんどくさい` |
| GoroButton ラベル副 | `— 押せ。考えるな。` (display-serif italic) |

**マウント時モーション**: `BalanceHero → MeterBar → MetricsRow → GoroButton` を **80ms 間隔のステアード fade-up** (`translateY: 8px → 0`, `opacity: 0 → 1`)。

**遷移先**: ボタン押下 → `SLOT`

### 3.2 IDLE (履歴あり通常)

**遷移条件**: 認証済み & `monthlyCount >= 1` & サジェスト無効 (履歴不十分 or 直近 3 時間以内に同カテゴリ注文済み)

| 要素 | 変化点 |
|---|---|
| BalanceHero 数字 | 現在残高 (消化率 80% 超で `accent-danger` 色) |
| MeterBar | 消化率に応じて `gold-500 → accent-warn (60% 超) → accent-danger (80% 超)` |
| MetricsRow | `今月 N 回` ／ `消化 X%` |
| GoroButton | EMPTY と同じ |

**モーション**: 消化率が 80% を超えた瞬間、MeterBar が **400ms ease で色遷移＋微パルス** (`box-shadow: 0 0 10px accent-danger` を 1 回)。

### 3.3 SUGGESTED (先回り提案)

**遷移条件**: 認証済み & サジェスト API (`GET /suggest`) がカードを返した。「履歴十分」(requirements.md `FR-SUGGEST-04`) の判定はバックエンド側で行い、フロントは応答の有無だけで分岐する

| 要素 | 変化点 |
|---|---|
| GoroButton ラベル主 | `YES` |
| GoroButton ラベル副 | サジェスト内容 (例: `— CoCo壱 ¥1,200 を手配します`) |
| SuggestBubble | ボタン上部に吹き出し: `そろそろご飯めんどくさいですよね？` |

**コピーバリエーション** (Bedrock 側が時刻・履歴に応じて選択):

- `そろそろご飯めんどくさいですよね？`
- `今日もダメになります？`
- `いつもの、いきます？`

**モーション**:

- マウント時、`SuggestBubble` は `opacity: 0` + `translateY: -4px` から 600ms ease で表示
- ボタン本体は **2.4s で gold グラデが微妙に明滅** (`box-shadow` opacity を `.4 ↔ .6` でループ) → 「呼吸している」感
- `prefers-reduced-motion` 時は明滅停止

**遷移先**: ボタン押下 → `SLOT` (サジェスト内容を idempotencyKey 付きで送信)

### 3.4 SLOT (注文処理中)

**遷移条件**: ボタン押下 → mutation 開始 → `onMutate` で遷移

| 要素 | 変化点 |
|---|---|
| GoroButton 内側 | 「押し込まれた」陰影に変化 (box-shadow 反転＋ `scale: .98`) |
| SlotReel | ボタン中央に黒地 LCD 窓が出現、店名候補が高速縦スクロール |
| BalanceHero | わずかに彩度を落とす (pending の視覚予告) |
| BrandHeader / MetricsRow | 固定 |

**SlotReel 仕様**:

- 黒地 `#0a0405`、ネオン枠 `gold-500` 1px
- 文字色 `gold-100`、`mono-pixel` 14px
- 候補リスト: API 待ちの間はローカルのダミー候補を `0.4s linear infinite` でループ
- `mutation.onSuccess` で API 応答の店名・金額を受領 → リールを 2 回転後にその文字で停止 (`cubic-bezier(.2,.8,.2,1)` で overshoot 後収束)
- 停止と同時に `BalanceHero` の数字が旧 → 新へ縦スクロール (450ms)

**所要時間**:

- 最低 1.2s (API が早く返ってもこの演出時間は確保)
- 最大 5s (API 失敗時はタイムアウトでエラートースト)

**遷移先**:

- 成功 → 残高更新 → `IDLE` (自動復帰)
- 残高不足エラー → `DEAD`
- ネットワークエラー → トースト表示 (既存 `Toast` コンポーネント) → `IDLE` 復帰

### 3.5 DEAD (残高 0 / 今月ダメになれません)

**遷移条件**: `balance === 0` または注文時に残高不足エラー

| 要素 | 変化点 |
|---|---|
| ScreenFrame 全体 | `filter: saturate(.2) brightness(.7)` を 600ms ease で適用 |
| BalanceHero 数字 | `¥0` (`dead-gray` 色、グロー消失) |
| GoroButton | グレー無効化 (`dead-gray` グラデ)、ラベル副は `— 残高不足` |
| DeadVerdict (オーバーレイ) | 画面中央に大判タイポで宣告 |
| IncreaseBudgetButton | 画面下部、唯一彩度 100% の `accent-danger` で生きている |

**DeadVerdict コピー** (display-serif italic 22px、3 行):

```
今月、
ダメに
なれません。
```

**IncreaseBudgetButton ラベル**: `来月から ¥50,000 にする`
- 増額値は AI 推奨。本 spec ではフロント側の表示として、まず固定で「現在予算 × 5/3 (端数を 10,000 円単位に丸め)」を初期実装とする。例: 30,000 → 50,000 / 50,000 → 80,000 / 80,000 → 130,000。AI による精緻な推奨は将来検討 (§7)。requirements.md `FR-METRICS-04` 準拠

**モーション**:

- ScreenFrame の脱色は 600ms ease-out
- DeadVerdict は脱色完了後 200ms 遅延で `opacity: 0 → 1` + `scale: .96 → 1` (500ms cubic-bezier)
- IncreaseBudgetButton は脱色対象外 (`isolation: isolate` でレイヤー分離)、さらに 400ms 遅延で下からスライドイン

**遷移先**:

- IncreaseBudgetButton タップ → 確認モーダル → 翌月予算を更新 → DEAD のまま (残高は 0、月初リセット待ち)
- 翌月 1 日 00:00 JST に EventBridge Scheduler 経由で残高リセット → `IDLE`

---

## 4. マイクロインタラクション・アクセシビリティ・既存実装への差分

### 4.1 マイクロインタラクション

| 名前 | トリガー | 仕様 |
|---|---|---|
| **press-down** | ボタン押下中 | `transform: scale(0.98)` ＋ box-shadow 1 段階弱化 (120ms ease-out) |
| **press-release** | リリース | scale 1.0 へ戻る (200ms cubic-bezier(.2,.8,.2,1)、軽く overshoot) |
| **slot-spin** | SLOT 突入 | リール strip を `0.4s linear infinite` で縦スクロール |
| **slot-stop** | mutation onSuccess | strip 停止 → 当選文字に止まる (2 回転後 600ms cubic-bezier、最後 12px 行き過ぎて戻す) |
| **balance-tick** | 残高更新 | 古い値 → 新しい値へ縦スクロール (450ms cubic-bezier、桁ごとに 30ms ステアード) |
| **suggest-breath** | SUGGESTED | ボタンの `box-shadow` opacity を `.4 ↔ .6` で 2.4s ループ (ease-in-out) |
| **meter-warn** | 消化率 80% 超過直後 | MeterBar が `accent-danger` に色遷移、1 回だけ glow パルス (400ms) |
| **dead-fade** | DEAD 突入 | ScreenFrame に `filter: saturate(.2) brightness(.7)` を 600ms ease-out で適用 |
| **verdict-appear** | dead-fade 完了 200ms 後 | DeadVerdict が `opacity: 0 → 1`, `scale: .96 → 1` (500ms cubic-bezier) |
| **increase-rise** | verdict-appear 400ms 後 | IncreaseBudgetButton が `translateY(60px) → 0` (500ms cubic-bezier) |
| **error-shake** | ネットワークエラー | ボタンが ±4px で 3 回横揺れ (300ms) |

**実装方針**: Motion ライブラリは追加しない。全て CSS `@keyframes` ＋ `transition` で完結。状態切り替えは `data-screen-state` 属性で CSS 側分岐。

### 4.2 アクセシビリティ

| 観点 | 仕様 |
|---|---|
| セマンティック | `<button>`, `<main>`, `<header>` を正しく使う |
| キーボード | GoroButton は Tab で focusable、Enter / Space で発火 |
| フォーカスリング | 金色のボタンに `outline: 3px solid #f4d990; outline-offset: 4px` |
| `aria-label` | 状態別: `idle` → `ご飯めんどくさい` ／ `suggested` → `YES — CoCo壱を注文` ／ `slot` → `注文処理中` ／ `dead` → `残高不足` |
| `aria-live` | 残高表示と DeadVerdict は `role="status" aria-live="polite"` |
| カラーコントラスト | `cream` on `bg-mid` ≒ 14:1 (AAA)／`gold-500` on `bg-mid` ≒ 7.8:1 (AAA Large) |
| モーション軽減 | `@media (prefers-reduced-motion: reduce)` で全アニメを `0.01ms` に短縮 (DEAD の filter は即時適用) |
| スクリーンリーダー | SlotReel の高速文字は `aria-hidden="true"`、当選結果は `<span class="sr-only">` で読み上げ |

### 4.3 レスポンシブ

| 幅 | 振る舞い |
|---|---|
| `<= 480px` | モバイル基準 |
| `481–768px` | コンテンツ最大幅 480px、画面中央寄せ |
| `> 768px` | 同様、最大幅 480px。背景は画面いっぱいに継続 (額装的に黒背景が広がる) |

### 4.4 既存実装への差分

| ファイル | 変更内容 |
|---|---|
| `web/app/page.tsx` | インラインスタイル削除、`MainScreen` コンポーネントへ委譲。認証分岐は据え置き |
| `web/components/order/GoroButton.tsx` | **大幅改修**。`screenState` を受け取り 4 状態を表現。スタイルはインラインから CSS Modules (`*.module.css`) へ移行 |
| `web/components/order/MainScreen.tsx` | **新規**。状態管理＋子コンポーネント orchestration |
| `web/components/order/ScreenFrame.tsx` | **新規**。背景＋ CRT スキャンライン |
| `web/components/order/BrandHeader.tsx` | **新規** |
| `web/components/order/BalanceHero.tsx` | **新規**。残高数字＋ MeterBar ＋ MetricsRow |
| `web/components/order/SlotReel.tsx` | **新規** |
| `web/components/order/DeadVerdict.tsx` | **新規** |
| `web/components/order/IncreaseBudgetButton.tsx` | **新規** |
| `web/components/order/SuggestBubble.tsx` | **新規** (GoroButton 内部) |
| `web/hooks/useBalance.ts` | **新規**。react-query で残高取得 |
| `web/hooks/useSuggest.ts` | **新規**。マウント時にサジェスト取得 |
| `web/components/order/OrderHistoryList.tsx` | メイン画面から外す。コンポーネントは将来用に温存 |
| `web/app/globals.css` | **新規**。デザイントークン (CSS variables) と共通アニメーション定義 |
| `web/app/layout.tsx` | Google Fonts (Cormorant Garamond / Zen Old Mincho / DotGothic16) を `next/font` 経由でロード |

**フォントロード戦略**: `next/font/google` で `display: swap`、subset は Cormorant Garamond=latin / Zen Old Mincho=japanese / DotGothic16=japanese。FCP 自主目標 2s 以内。

### 4.5 テスト戦略

| 種類 | 内容 |
|---|---|
| ビジュアル | 5 状態 (empty/idle/suggested/slot/dead) のスナップショットを Playwright で取得 |
| ユニット | 状態遷移ロジック (`useScreenState` のレデューサ) を vitest で検証 |
| アクセシビリティ | 主要状態で `axe-core` を Playwright 経由で実行、Violations 0 を assert (`@axe-core/playwright` を新規追加) |
| モーション軽減 | `prefers-reduced-motion: reduce` 環境でアニメ停止を Playwright で確認 |

既存の `web/tests/` 構成 (vitest + Playwright) に乗せる。`@axe-core/playwright` は本 spec で唯一追加する dev 依存。

---

## 5. 要件トレーサビリティ

| 要件 ID | 充足箇所 |
|---|---|
| FR-AUTH-* | 既存 `useAuth` を流用、認証分岐は `app/page.tsx` で据え置き (本 spec の対象外) |
| FR-BUDGET-05 (残高常時表示) | §2.1, §3 全状態の `BalanceHero` |
| FR-ORDER-01 (中央のボタン) | §2.1, §3.1 EMPTY |
| FR-ORDER-07 (完了画面) | 注文完了画面は別ファイル (`web/app/order/[id]/complete/page.tsx`) — 本 spec のスコープ外 |
| FR-ORDER-08 (体感 3 秒) | §3.4 SLOT 演出時間 (最低 1.2s、API 含めて 3s 目標) |
| FR-SUGGEST-02 (サジェスト表示) | §3.3 SUGGESTED, `SuggestBubble` |
| FR-SUGGEST-04 (履歴不十分でサジェスト出さない) | §3.3 遷移条件 |
| FR-METRICS-01,02 (今月のダメ化回数・消化率) | §2.1, §3 `MetricsRow` |
| FR-METRICS-03 (80% で強調) | §3.2 IDLE / §4.1 meter-warn |
| FR-METRICS-04 (残高 0 → 増額提案) | §3.5 DEAD, `IncreaseBudgetButton` |
| FR-UX-01 (操作 1 タップ完結) | §3.3 SUGGESTED でも 1 タップ |
| FR-UX-02 (音声・自由入力なし) | UI に入力フィールドなし |
| NFR-DEG-01 (タップ最大 2 回) | 全状態で 1 タップ完結 |
| NFR-DEG-02 (起動時サジェスト) | §3.3 |
| NFR-DEG-03 (常時可視化) | §3 全状態で残高・回数・消化率を表示 |
| NFR-DEG-04 (増額誘導) | §3.5 |
| NFR-DEG-05 (依存促進的コピー) | `押せ。考えるな。` `そろそろご飯めんどくさいですよね？` `今月ダメになれません。` |
| NFR-PERF-01 (3 秒以内) | §3.4 SLOT 所要時間 |
| NFR-PERF-02 (5 秒以内初期表示) | §4.4 フォント `display: swap`、FCP 自主目標 2s |
| NFR-A11Y-01 (セマンティック HTML) | §4.2 |

---

## 6. 既存 spec / ドキュメントとの関係

- 本 spec は `aidlc-docs/inception/application-design/` の `components.md`, `services.md` の延長線上にあり、メイン画面の **見た目とマイクロインタラクション** を補完するもの。
- バックエンド API 仕様 (`/order`, `/balance`, `/suggest`) は既存 `aidlc-docs/construction/order/` 配下と整合させる。本 spec ではフロント側の利用方法のみ規定する。
- AI-DLC ワークフロー上は Construction フェーズの追加成果物として扱える (Functional Design 補完 / NFR Design 補完)。

---

## 7. オープン項目 (将来検討)

- **AI 推奨の増額値の決定ロジック**: 現状は「現在予算 × 1.5–1.7」を仮置き。Bedrock 側の推論ロジックは別途設計。
- **サジェストの timing 学習**: 現 spec は「マウント時に 1 回叩く」で固定。フェーズ 2 を深掘りするなら定期的なサジェスト再評価 (notification API なしでの) を将来検討。
- **多カテゴリ対応**: 本 spec は「ご飯」カテゴリのみ。requirements.md `FR-ORDER-09` の SHOULD 要件である清掃・役所手続き等は、ボタンを縦に並べる or サジェストカードで掘り下げる将来案あり。
- **完了画面 (`/order/[id]/complete`) のリデザイン**: 本 spec の世界観に合わせる必要あり。別 spec として切り出す。
- **ログイン / サインアップ画面のリデザイン**: 本 spec のスコープ外。コンセプトを継承した別 spec が必要。
