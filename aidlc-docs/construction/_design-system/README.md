# 横串デザインシステム (Cross-Cutting Design System)

**Created**: 2026-05-25
**Position**: AI-DLC `CONSTRUCTION PHASE` の **横串成果物** (per-Unit ループとは別軸)

---

## 0. このディレクトリは何か

`aidlc-docs/construction/{auth,budget,order}/` の各 Unit は、AI-DLC ワークフローに従って Functional Design → NFR Design → Infrastructure Design → Code Generation の per-Unit ループで成果物を出してきた。

しかし **「フロントエンドの世界観 (Slot Machine + リヴァイ調コピー)」** は単一 Unit に閉じない横串の関心事である:
- Unit A `auth` の画面 (ランディング / サインアップ / ログイン / セッション切れ)
- Unit B `budget` の表現 (残高 / メーター / DEAD 状態)
- Unit C `order` の中核体験 (めんどくさいボタン / スロット演出 / 完了画面)
- Unit D `suggest` の挙動 (吹き出し / ボタン一体化)
- Unit E `metrics` の表現 (今月 N 度 / 消化率 / 増額誘導)

これら全てに共通する **デザイントークン・コピーシステム・状態モデル・コンポーネント分解** をまとめたのが本ディレクトリ。

---

## 1. AI-DLC ワークフロー上の位置づけ

```
INCEPTION
  └─ Application Design (済)

CONSTRUCTION
  ├─ Unit A `auth`        ─ Functional Design 〜 Code Generation (済)
  ├─ Unit B `budget`      ─ Functional Design 〜 Infrastructure Design (済) / Code Generation (未)
  ├─ Unit C `order`       ─ Functional Design 〜 Code Generation (済、ユーザ承認待ち)
  ├─ Unit D `suggest`     ─ 未着手
  ├─ Unit E `metrics`     ─ 未着手
  ├─ _design-system       ─ ★本ディレクトリ (横串・追加スコープ)
  └─ Build and Test       ─ 全 Unit 完了後
```

本ディレクトリは **追加スコープ**として扱う:
- ハッカソン審査軸「創造性とテーマ適合性」を最大化するため、UI/UX の世界観を一貫させる必要が出た (2026-05-24 のヒアリングで決定)
- 各 Unit の Functional Design (frontend-components.md) は **構造**を定義しているが、**世界観・トーン・コピー**は明示されていなかった
- 本ディレクトリは各 Unit の `frontend-components.md` を **補完**する位置づけ

---

## 2. 成果物

### `design-spec.md`

全画面の詳細仕様 (740 行):
- §1 コンセプト & デザイントークン (色 / タイポ / モーション / コピーシステム)
- §2 メイン画面のコンポーネント分解
- §3 メイン画面の状態モデル (idle / suggested / slot / dead)
- §4 マイクロインタラクション・アクセシビリティ・既存実装差分
- §5 他画面の詳細仕様 (ランディング / サインアップ / ログイン / 完了 / モーダル / Toast)
- §6 共通アクセシビリティ・レスポンシブ
- §7 要件トレーサビリティ
- §8 既存 spec / ドキュメントとの関係
- §9 オープン項目

### `implementation-plan.md`

22 タスク構成の実装プラン (3263 行):
- TDD ベースのチェックボックス形式
- 各タスクにファイルパス・コード本体・lint/test コマンド・commit メッセージを明記
- Task 1〜4 は基盤 (トークン / コピー / 計算ロジック / state)
- Task 5〜13 はメイン画面コンポーネント
- Task 14〜20 は他画面
- Task 21〜22 は仕上げ

---

## 3. 各 Unit との関係

| Unit | 既存成果物 | 本ディレクトリでの扱い |
|---|---|---|
| Unit A `auth` | LC-AUTH-01〜18 (Code Generation 済) | ランディング / サインアップ / ログイン / セッション切れ画面の **見た目とコピーを刷新**。既存の `useAuth` hook やフォーム挙動はそのまま流用 |
| Unit B `budget` | Functional/NFR/Infra Design 済 (Code Generation 未) | 残高表示 / メーター / DEAD 状態を本 design system に従って実装。Code Generation 着手時に `implementation-plan.md` の関連タスクを参照 |
| Unit C `order` | LC-ORDER-01〜34 (Code Generation 済、ユーザ承認待ち) | 既存 `GoroButton` などを **大幅改修**。本 design-spec が新コンポーネント仕様を提示 |
| Unit D `suggest` | 未着手 | サジェスト吹き出し・ボタン一体化挙動の仕様を本 spec で先取り定義 |
| Unit E `metrics` | 未着手 | 増額誘導モーダル・ダメ化メトリクス表示の仕様を本 spec で先取り定義 |

**重要**: 本デザインシステムが定義する内容と、既存の各 Unit の Functional Design `frontend-components.md` で**矛盾**する部分がある場合、原則として **本デザインシステムが優先**される (より新しい・より一貫した仕様であるため)。Code Generation 担当はまず本デザインシステムを参照し、各 Unit 固有の業務ロジック (バリデーション / API 呼び出し / state 管理) は当該 Unit の成果物を参照する。

---

## 4. 実装の進め方

実装担当者は次の順で読む:

1. **本 README** (位置づけと方針)
2. **`design-spec.md`** (世界観・トークン・コピー・状態モデル・各画面詳細)
3. **`implementation-plan.md`** (Task 1 から順に着手)
4. 各タスクで関連する Unit の `aidlc-docs/construction/{unit}/` 配下を参照 (業務ロジック確認用)

`implementation-plan.md` の各タスクは独立した commit に対応するので、フィードバック反映や巻き戻しが容易。subagent-driven-development 方式で 1 タスク 1 サブエージェント、各タスク完了後にレビューを推奨。

---

## 5. 用語

- **Slot Machine**: 本デザインシステムのコンセプト名 (spec §1.1)
- **リヴァイ調コピー**: 「進撃の巨人」リヴァイの口調をパロディ参照点とした、短文体言止め・終止形限定・上から目線のコピー規約 (spec §1.6)
- **デザイントークン**: CSS variables として `web/app/globals.css` に集約される色・タイポ・スペーシング・モーション値の総称
- **横串**: 単一の Unit に閉じない、全 Unit を貫く関心事

---

## 6. 履歴

| Date | Event |
|---|---|
| 2026-05-24 | brainstorming セッション開始 (Visual Companion で Slot Machine コンセプトを確定) |
| 2026-05-25 | spec v2.0 (全画面拡張版) と implementation-plan を作成、PR #94 提出 |
| 2026-05-25 | AI-DLC `_design-system` 配下に整理し直し、本 README を追加 |
