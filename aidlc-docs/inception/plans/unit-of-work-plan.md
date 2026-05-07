# Unit of Work Plan — ゴロゴロPay

**Created**: 2026-05-07
**Role**: Software Architect (as Planner)
**Prerequisites**: Requirements Analysis / User Stories / Workflow Planning / Application Design 承認済み
**Depth**: Comprehensive（execution-plan.md §6 に従う）

---

## 0. 目的

Application Design で示唆された **6 Unit 候補**（A: 認証 / B: ダメ予算 / C: 代行手配コア / D: 学習・先回り / E: ダメ化メトリクス / F: ダメ化UX 横串）を、**正式な Unit of Work** として確定する。

「Unit of Work」の本プロジェクト定義:
- **論理的なコードモジュールの集合**（物理 Lambda とは独立、Q-E=B' のモノリシック Lambda 構成）
- **1 Unit = 1 ビジネスドメイン = 1 Go パッケージ系列**
- **各 Unit が per-unit Construction フェーズの単位になる**

スコープ外:
- 物理 Lambda の分割（Application Design で確定済、単一 API Lambda）
- Terraform モジュール分割（Application Design §8 で確定）

---

## 1. 事前確認質問

本 MVP の Unit 分解方針を固めるため、以下の質問にお答えください。各 `[Answer]:` タグに **A / B / C / X** の文字で回答してください。

### Question A: Unit の最終数

Application Design で示唆した 6 Unit 候補のうち、どの粒度で確定しますか？

A) **Unit A〜F の 6 個すべて**（ダメ化UX を独立 Unit F として維持）
B) **Unit A〜E の 5 個**（ダメ化UX は NFR 横串として独立 Unit にしない、各 Unit 内の受入基準に織り込む）
C) **Unit A〜E の 5 個 + スケジューラ**（SchedulerLambda / 月初リセットを独立 Unit として切り出す）
X) Other

**推奨**: **B (5 Unit)** — ダメ化UX は NFR-DEG として全 Unit に通底するため、独立 Unit にすると「同じ要件が複数 Unit に跨る」状態を作ってしまう。各 Unit の受入基準（stories.md の Epic X を対応する機能 Unit に統合）に織り込む方が実装時に扱いやすい。SchedulerLambda は Unit B (ダメ予算) の配下に位置付ける（WalletService を共有）。

[Answer]: B — 5 Unit 構成（A: 認証 / B: ダメ予算 / C: 代行手配コア / D: 学習・先回り / E: ダメ化メトリクス）。ダメ化UX は独立 Unit とせず、各 Unit の受入基準に横串として織り込む。SchedulerLambda は Unit B 配下に位置付ける。

---

### Question B: 共有コンポーネントの扱い

`BedrockAdapter` と `FallbackSuggestProvider` は Unit C（代行手配コア）と Unit D（学習・先回り）の両方から使われる。どう扱いますか？

A) **共有パッケージ `pkg/` に切り出す**（`pkg/bedrock/`, `pkg/fallback/`）
B) **Unit C 側に配置し、Unit D から import**（C が D の先、依存方向が明確になる）
C) **Unit D 側に配置し、Unit C から import**（D のほうが Bedrock の主利用者）
X) Other

**推奨**: **A (pkg/)** — どちらかの Unit に寄せると「この Unit は相手に依存する」という非対称が生まれる。本来は横串ユーティリティなので `internal/adapters/` 直下、または `pkg/` に置くのが素直。ただし Go の `internal/` ルールを考えると `apps/api/internal/adapters/bedrock/` として API 側に置き、Unit C/D の両方から参照するのも可（本 MVP は単一 Lambda なので `internal/` で十分）。

[Answer]: A — BedrockAdapter と FallbackSuggestProvider は共有コンポーネントとして Unit 非所属に配置する。実装パスは `apps/api/internal/adapters/bedrock/` および `apps/api/internal/adapters/fallback/`。Unit C と Unit D の両方から同じ interface 経由で参照する。

---

### Question C: 推奨実装順序

Construction フェーズでの Unit 実装順序はどれが良いですか？

A) **A → B → C → D → E**（依存関係に素直、下位基盤から積む）
B) **C → B → A → D → E**（コアユースケース US-1-01 を最優先、認証は後回し）
C) **A → B → C → (D と E 並列)**（D/E は独立度が高いため並列化）
X) Other

**推奨**: **A (A → B → C → D → E)** — 認証（A）を先に固めると C の受入テストで「ログイン済ユーザ」前提が使え、B を固めると C の Wallet.Deduct が動作可能になる。D は C で蓄積された履歴が前提、E は全 Unit の集計のため最後が自然。ハッカソン締切を意識した最短経路。

[Answer]: A — 実装順序は A（認証） → B（ダメ予算） → C（代行手配コア） → D（学習・先回り） → E（ダメ化メトリクス）。依存関係に沿った下位基盤からの積み上げ方式。

---

### Question D: Unit 間通信パターン

Unit 間の通信（例: OrderService が WalletService を呼ぶ）の実装方式は？

A) **直接 Go 関数呼び出し**（モノリシック Lambda なので同一プロセス、interface 経由で疎結合化）
B) **Event Bus（EventBridge 経由）で非同期化**（将来の Lambda 分離を見据える）
C) **ドメインイベント風に内部チャネル**（goroutine + channel）
X) Other

**推奨**: **A** — Q-E=B'（モノリシック Lambda）と整合。MVP スコープで同期呼び出しが最もシンプル。将来の Unit ごと Lambda 分離は、interface 経由で疎結合にしておけば容易に置き換え可能。

[Answer]: A — Unit 間通信は Go の interface 経由の直接関数呼び出し（同期）とする。モノリシック Lambda 内で同一プロセスの関数呼び出しとして実装し、将来の Unit 分離に備えて interface で疎結合を担保する。

---

### Question E: per-unit Construction の深度

Construction フェーズで各 Unit の Functional Design / NFR Design / Infrastructure Design の深度はどうしますか？

A) **全 Unit を Standard 深度で統一**（MVP スコープ、締切重視）
B) **Unit C (コア) のみ Comprehensive、他は Standard**（コアだけ厚く）
C) **全 Unit を Comprehensive**（審査観点「ドキュメント品質」最優先、工数大）
X) Other

**推奨**: **B** — 本 MVP の命運を握るのは Unit C（「ご飯めんどくさい」コア）。ここは Bedrock 連携 + 冪等性 + アダプタ層が絡み、ハッカソンで最も複雑な部分。他 Unit は概ね標準的な CRUD + 集計で Standard で十分。

[Answer]: B — Unit C（代行手配コア）のみ Comprehensive 深度、他の 4 Unit（A/B/D/E）は Standard 深度。Unit C の Functional Design / NFR Requirements / NFR Design / Infrastructure Design は詳細シーケンス図・境界値・PBT プロパティ定義まで記述し、ハッカソン審査での技術的見せ場とする。

---

### Question F: Unit 命名規則

正式な Unit 名の命名規則は？

A) **英語 PascalCase**（`Authentication`, `Budget`, `Order`, `Suggestion`, `Metrics`）
B) **英語 kebab-case / パッケージ名と同期**（`auth`, `budget`, `order`, `suggest`, `metrics`）
C) **日本語で命名**（「認証」「ダメ予算」「代行手配」「先回り提案」「ダメ化メトリクス」）
X) Other

**推奨**: **B** — Go パッケージ名（`internal/wallet`, `internal/order` 等）と合わせて検索性・一貫性を保つ。日本語は Unit 説明文で利用し、ID 化には英語を採用。

[Answer]: B — Unit 名は英語 kebab-case / パッケージ名と同期した識別子（`auth`, `budget`, `order`, `suggest`, `metrics`）を採用。日本語名は説明文でのみ使用し、ID 化は英語で統一する。

---

## 2. Part 2 で生成する成果物（ユーザ承認後に実行）

### ドキュメントファイル
- [ ] `aidlc-docs/inception/application-design/unit-of-work.md` — Unit 定義と責務、コード組織戦略
- [ ] `aidlc-docs/inception/application-design/unit-of-work-dependency.md` — Unit 間依存マトリクス + Mermaid 依存図
- [ ] `aidlc-docs/inception/application-design/unit-of-work-story-map.md` — Story → Unit マッピング（全 23 ストーリーを Unit に割り当て）

### 検証項目
- [ ] 全 23 ストーリー（US-0-* / US-1-* / US-2-* / US-3-* / US-X-*）が Unit に割り当てられる
- [ ] 全機能要件（FR-AUTH / FR-BUDGET / FR-ORDER / FR-LEARNING / FR-SUGGEST / FR-METRICS / FR-UX）が Unit にマッピングされる
- [ ] Unit 境界に循環依存がない
- [ ] Unit 間の共有コンポーネント（BedrockAdapter 等）の扱いが明文化される
- [ ] per-unit Construction の順序と深度が定義される

### 状態更新
- [ ] `aidlc-docs/aidlc-state.md` の Stage Progress で Units Generation を [x]
- [ ] `aidlc-docs/audit.md` に完了記録を追加

---

## 3. 完了条件

- 全 `[Answer]:` タグに回答あり
- 上記チェックリストが全て [x]
- 3 つの成果物ドキュメントがユーザ承認済み
- Inception フェーズの全成果物（要件書 / ユーザストーリー / ペルソナ / 実行計画 / アプリケーション設計 / Unit 分解）が揃い、2026-05-10 のハッカソン応募要件を満たす
