# User Stories Generation Plan — ゴロゴロPay

**Created**: 2026-05-07
**Role**: Product Owner
**Prerequisite**: `aidlc-docs/inception/requirements/requirements.md` 承認済み

---

## 0. 目的

`requirements.md` に定義された機能要件（FR-AUTH / FR-BUDGET / FR-ORDER / FR-LEARNING / FR-SUGGEST / FR-METRICS / FR-UX）を、**ペルソナ起点の INVEST 準拠ユーザストーリー群** と **受入基準** に翻訳する。

「ダメ化UX」という独自 NFR を、ペルソナの感情・行動・体験の連なりとして具体化し、AWS Summit Japan 2026 ハッカソン審査観点（Intent 明確さ / 創造性 / Unit 分解 / ドキュメント品質）に貢献する。

---

## 1. 事前確認質問 — ストーリー生成の方針

ストーリー生成に入る前に、以下の方針を確認させてください。各 `[Answer]:` タグに **A / B / C / X** の文字で回答してください。

### Question A: ペルソナの詳細度
ペルソナ（`personas.md`）はどの粒度で作成しますか？

A) 1人のメインペルソナ（「ゴロゴロ太郎」）に絞り、フェーズ1→2→3の感情変遷を深く描写する
B) 3人のペルソナ（例: 一人暮らし社会人 / リモートワーカー / 学生）で多角的に描写
C) 4人以上（管理者ペルソナ・運用者ペルソナも含む）
X) Other (please describe after [Answer]: tag below)

**推奨**: A — 本 MVP はエンドユーザ向け体験が主役。単一ペルソナの感情変遷を深く描く方が、テーマ「人をダメにする」への訴求力が強い。

[Answer]: A — メインペルソナ「ゴロゴロ太郎」1人に絞り、ダメ化フェーズ1（快感）→2（依存）→3（退化）の感情変遷を深く描写する。

---

### Question B: ストーリーの分解方針
ユーザストーリー（`stories.md`）の分解方針はどれですか？

A) **Feature-Based** — 機能領域（AUTH / BUDGET / ORDER / ...）ごとに分解
B) **User Journey-Based** — ユーザ体験の流れ（初回ログイン → 初代行手配 → 習慣化 → 先回り → 無力感）ごとに分解
C) **Epic-Based** — 「ダメ化フェーズ1」「フェーズ2」「フェーズ3」を Epic として、配下にストーリーを束ねる
X) Other (please describe after [Answer]: tag below)

**推奨**: C — 審査観点「創造性とテーマ適合性」への訴求が最も強い。idea.md のフェーズ1→2→3 のストーリー構造と呼応し、Intent の縦串が効く。

[Answer]: C — Epic-Based 分解。「ダメ化フェーズ0: 準備（認証・ダメ予算設定）」「フェーズ1: ボタンを押すだけ（代行手配の快感）」「フェーズ2: ボタンすら不要（先回り提案）」「フェーズ3: 完全に人がダメになる（ダメ化メトリクス・増額誘導）」を Epic として、配下にストーリーを束ねる。各ストーリーに機能領域（FR-AUTH / FR-BUDGET 等）のタグを補助表記として付与する。

---

### Question C: 受入基準の形式
各ストーリーの受入基準（Acceptance Criteria）の形式は？

A) **Given / When / Then**（Gherkin 形式、BDD 寄り、テスト設計に直結）
B) 箇条書きチェックリスト（シンプル、視認性良い）
C) 両方併記（Given/When/Then + 追加のチェックリスト）
X) Other (please describe after [Answer]: tag below)

**推奨**: A — 後続の Build and Test ステージでの受入テスト作成に直結し、ハッカソン審査の「ドキュメント品質」にも寄与。

[Answer]: A — Given / When / Then（Gherkin 形式）で受入基準を記述する。

---

### Question D: ダメ化UX 体験の扱い
独自 NFR「ダメ化UX」をユーザストーリーにどう織り込みますか？

A) 機能ストーリーの **Acceptance Criteria** の中に、ダメ化要素（即時性・低摩擦・無力感の演出等）を埋め込む
B) ダメ化UX 専用の独立ストーリー（「ユーザは残高不足時に翌月予算増額の誘惑を感じる」等、体験ストーリー）を別セクションで定義
C) 両方併用 — 機能ストーリーの受入基準に織り込みつつ、ダメ化UX 専用のストーリー章を別途立てる
X) Other (please describe after [Answer]: tag below)

**推奨**: C — 「ダメ化UX」を独立章にすることで審査員への訴求が明確になり、同時に機能ストーリー側にもダメ化要素を埋め込むことで実装時のブレを防ぐ。

[Answer]: C — 機能ストーリーの受入基準にダメ化要素を織り込みつつ、別章で「ダメ化UX 体験ストーリー」を独立セクションとして立てる。

---

### Question E: ストーリー数の目標
MVP スコープで目指すストーリー本数は？

A) 10〜15 本（精選、コア機能のみ）
B) 15〜25 本（標準、MVP の全機能要件をカバー）
C) 25〜40 本（網羅的、将来対応も一部含む）
X) Other (please describe after [Answer]: tag below)

**推奨**: B — 締切（2026-05-10）を踏まえつつ、MVP の全機能要件をユーザ視点でカバーする標準的な本数。

[Answer]: B — 15〜25 本を目標とする。

---

## 2. ストーリー生成タスク一覧（PART 2 で実行）

以下のチェックリストは、ユーザ承認後の PART 2 Generation フェーズで順次実行する項目です。ユーザ回答をもとに具体化されます。

### ペルソナ作成
- [x] `aidlc-docs/inception/user-stories/personas.md` を作成
- [x] メインペルソナ（「ゴロゴロ太郎」）の基本属性・背景・目標・痛み・価値観を記述
- [x] フェーズ1（快感）・フェーズ2（依存）・フェーズ3（退化）それぞれの感情変遷を記述
- [x] Question A の回答に応じて追加ペルソナ（必要なら）を記述 → 不要（単一ペルソナで描写）

### ストーリー作成
- [x] `aidlc-docs/inception/user-stories/stories.md` を作成
- [x] Question B で選ばれた分解方針に従って Epic / セクション構造を定義
- [x] 機能要件 FR-AUTH を User Story に変換（Epic 0: US-0-01, US-0-02）
- [x] 機能要件 FR-BUDGET を User Story に変換（US-0-03, US-1-02, US-1-04, US-1-05, US-1-06, US-3-05）
- [x] 機能要件 FR-ORDER を User Story に変換（US-1-01, US-1-03, US-1-07）
- [x] 機能要件 FR-LEARNING を User Story に変換（US-1-03, US-2-04）
- [x] 機能要件 FR-SUGGEST を User Story に変換（US-2-01, US-2-02, US-2-03）
- [x] 機能要件 FR-METRICS を User Story に変換（US-3-01, US-3-02, US-3-03, US-3-04）
- [x] 機能要件 FR-UX を User Story に変換（横串として全ストーリーに通底、US-0-04, US-1-07）
- [x] Question D の回答に応じて「ダメ化UX 体験ストーリー」章を追加（Epic X: US-X-01〜US-X-03）

### 受入基準と INVEST 適合性チェック
- [x] 各ストーリーに Question C で選ばれた形式の受入基準を記載（Gherkin: Given/When/Then）
- [x] 各ストーリーが INVEST 基準を満たすことを確認
- [x] ペルソナと各ストーリーの紐付けを明示（ゴロゴロ太郎のフェーズを Epic 単位で紐付け）

### 審査観点トレーサビリティ
- [x] 各ストーリーが「ダメ化UX」NFR のどの側面に対応するかを注記（タグ `#DegenerationUX:*`）
- [x] 各 Epic / ストーリー群と、後続 Units Generation ステージの Unit 候補の対応を stories.md 末尾に明示

### 状態更新
- [x] `aidlc-docs/aidlc-state.md` の Stage Progress で User Stories を [x] に更新
- [x] `aidlc-docs/audit.md` に完了記録を追加

---

## 3. 完了条件

- 全 `[Answer]:` タグに回答あり
- 上記チェックリストが全て [x]
- `personas.md` と `stories.md` がユーザ承認済み
