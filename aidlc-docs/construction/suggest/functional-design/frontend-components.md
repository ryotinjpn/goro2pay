# Unit D (`suggest`) — Frontend Components

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / Functional Design
**Unit**: D — `suggest`
**Depth**: Standard
**Related**: [business-logic-model.md](./business-logic-model.md)、[business-rules.md](./business-rules.md)、[domain-entities.md](./domain-entities.md)、凍結契約 [unit-interfaces.md](../../interfaces/unit-interfaces.md) §5・§9、横串 [_design-system/design-spec.md](../../_design-system/design-spec.md) §3.3（PR #94）
**Aligned with**: Plan 回答 Q-DF8（命名突合）/ Q-DF9（Title 文言）/ Q-DF10（取得タイミング）、design spec §2.4・§3.3

本ドキュメントは Unit D のフロントエンド構造を定義する。**インターフェースは凍結契約を正、視覚表現は design spec §3.3 を正**とする（BR-D15）。Unit D は独自画面を持たず、メイン画面（Unit C 所有の `GoroButton` / `MainScreen`）の **SUGGESTED 状態に寄生する形**で表示される。

---

## 1. 凍結契約 ⇄ design spec 命名対応表（重要）

Plan Q-DF8=A により、両資料の名前の食い違いを以下のとおり確定する。

| 概念 | 凍結契約（IF の正） | design spec（視覚の正） | 本 Unit D での確定 |
|---|---|---|---|
| データ取得 hook | `useSuggestion()`（§9） | `useSuggest()`（§2.4） | **`useSuggestion()`** を採用（IF=契約） |
| API | `GET /api/suggest`（§5.2） | `GET /suggest`（§3.3） | **`GET /api/suggest`**（BFF 経由、契約） |
| 提案表示部品 | `SuggestionCard`（unit-of-work） | `SuggestBubble`（§2.2/§3.3） | **`SuggestBubble`** を採用（視覚=spec）。GoroButton 内蔵の吹き出し |
| 戻り DTO | `Suggestion`/`SuggestionPlan`（§5.1） | （未定義） | 契約 §5.1 を採用 |

> この対応表が Unit D フロントの単一の真実。Code Generation はこの表に従う。

---

## 2. コンポーネント階層

Unit D は Unit C / design system の `MainScreen` ツリー内に表示される（design spec §2.5）。Unit D 所有は **`SuggestBubble`** と **`useSuggestion`** のみ。

```
<MainScreen>                      ← Unit C / design system 所有
  <ScreenFrame>
    <BrandHeader />
    <BalanceHero />
    <GoroButton screenState="suggested">   ← Unit C 所有。state を受けて表出変化
      <SuggestBubble />           ← ★Unit D 所有（suggested 状態でのみ表示）
      <SlotReel />                ← Unit C 所有（slot 状態）
    </GoroButton>
  </ScreenFrame>
</MainScreen>
```

- `useSuggestion()`（Unit D）は `MainScreen` マウント時に呼ばれ、結果を `GoroButton` の `screenState` と `SuggestBubble` の表示内容に供給する。
- `screenState='suggested'` への遷移条件は「`useSuggestion` が `hasSuggestion:true` を返した」こと（design spec §3.3）。

---

## 3. `useSuggestion` hook（Unit D 所有）

### 3.1 公開型（凍結契約 §9 準拠）

```typescript
function useSuggestion(): {
  suggestion: {
    hasSuggestion: boolean;
    suggestionId?: string;
    title?: string;
    plan?: { storeName: string; menuName: string; amount: number; category: string };
  } | null;
  isLoading: boolean;
};
```

### 3.2 振る舞い

| 項目 | 仕様 | 由来 |
|---|---|---|
| 取得タイミング | `MainScreen` マウント時に **1 回** `GET /api/suggest` | BR-D14 / Q-DF10 / spec §2.4 |
| 再評価 | なし（ポーリング・interval なし）。`staleTime` を長め（例: Infinity 相当）に設定し再フェッチ抑制 | BR-D14 |
| データ層 | TanStack Query（`queryKey: ['suggestion']`）、共有 `apiClient`（LC-AUTH-09、BFF 経由）を利用 | 契約 §12 |
| 認証 | `apiClient` が `Authorization: Bearer <accessToken>` を自動付与 | Unit A 共有 |
| エラー時 | 取得失敗（ネットワーク等）は `suggestion: {hasSuggestion:false}` 相当に丸め、カード非表示（先回り提案が出ないだけ、致命的でない） | BR-D02 方針と整合 |
| `hasSuggestion:false` | `SuggestBubble` 非表示、`GoroButton` は通常（idle）表示 | US-2-02 |

### 3.3 API 統合契約

| 項目 | 値 |
|---|---|
| Method / Path | `GET /api/suggest`（契約 §5.2） |
| 認証 | 必須（Cognito Authorizer、Unit A） |
| レスポンス(200, 履歴十分) | `{hasSuggestion:true, suggestionId, title, plan:{storeName,menuName,amount,category}}` |
| レスポンス(200, 履歴不足/抑制) | `{hasSuggestion:false}` |
| クライアント | 共有 `apiClient` ラッパ（独自 fetch を作らない、契約 §12） |

---

## 4. `SuggestBubble` コンポーネント（Unit D 所有、design spec §3.3）

### 4.1 役割

`GoroButton` が `screenState='suggested'` のとき、ボタン上部に乗る吹き出し。先回り提案を「断定形」で提示する（design spec §0.5.2 フェーズ2の体現）。

### 4.2 Props

```typescript
type SuggestBubbleProps = {
  visible: boolean;          // screenState === 'suggested'
  // 表示文言は固定（Q-DF9）。plan は GoroButton 副ラベル側で使用するため、
  // SuggestBubble 自体は文言を受け取らない設計でもよい（実装裁量）
};
```

### 4.3 表示仕様（design spec §3.3、BR-D16）

| 要素 | 表示 | 由来 |
|---|---|---|
| 吹き出し文言 | **固定「そろそろだろ。」**（API `title` は使わず固定優先） | BR-D13 / Q-DF9 |
| GoroButton 主ラベル（連動） | `押す。` | spec §3.3 |
| GoroButton 副ラベル（連動） | `— {storeName} ¥{amount} だ。`（例: `— CoCo壱 ¥1,200 だ。`） | spec §3.3 |
| モーション | マウント時 `opacity:0 + translateY:-4px → 600ms ease` で表示。ボタンは `suggest-breath`（2.4s 明滅） | spec §3.3 / §4.1 |
| reduced-motion | 明滅停止 | spec §4.2 |

> 副ラベルの `{storeName}`/`{amount}` は `useSuggestion().suggestion.plan` の動的値。文言の語尾（`だ。`）だけがリヴァイ調（design spec §1.6）。

### 4.4 アクセシビリティ（design spec §4.2 準拠）

| 観点 | 仕様 |
|---|---|
| `aria-label`（suggested 時の GoroButton） | `{storeName} ¥{amount} を注文`（機能ベース、視覚は世界観コピー） |
| SuggestBubble | 装飾的吹き出しは `aria-hidden` 可。提案内容は GoroButton の `aria-label` で読み上げ |
| キーボード | GoroButton（Unit C）が Tab focusable / Enter・Space 発火。SuggestBubble は非インタラクティブ |

### 4.5 `data-testid`（自動化対応、CLAUDE.md code-generation 準拠）

| 要素 | data-testid |
|---|---|
| SuggestBubble ルート | `suggest-bubble` |
| 提案副ラベル（店名・金額） | `suggest-plan-label` |

---

## 5. ユーザ操作フロー（フロント視点）

```
メイン画面マウント
  └─ useSuggestion() → GET /api/suggest
       ├─ hasSuggestion:false → SuggestBubble 非表示 / GoroButton 通常「めんどくさい」
       └─ hasSuggestion:true  → screenState='suggested'
              ├─ SuggestBubble「そろそろだろ。」表示（fade-up + breath）
              ├─ GoroButton 主「押す。」/ 副「— {店} ¥{金額} だ。」
              └─ ユーザが「押す。」1 タップ
                    └─ useOrder().placeOrder({ suggestionId })   ← Unit C の hook
                          └─ 成功時 invalidateQueries(['balance'])（契約 §9.1、Unit C 責務）
                          └─ slot → 完了画面（Unit C / design system）
```

> 「押す。」押下時のリクエストは **Unit C の `useOrder`** が担う（`suggestionId` を渡すだけ）。Unit D フロントは「提案の表示」までが責務で、注文送信は Unit C に委譲する（US-2-03 の共同担当の境界）。

---

## 6. 既存実装・他 Unit との差分

| ファイル | 区分 | 内容 |
|---|---|---|
| `web/hooks/useSuggestion.ts` | **新規（Unit D）** | §3。マウント時 1 回取得 |
| `web/components/order/SuggestBubble.tsx` | **新規（Unit D）** | §4。GoroButton 内蔵の吹き出し |
| `web/components/order/GoroButton.tsx` | 改修（Unit C / design system 所有） | `screenState='suggested'` 表出。Unit D は props 経由で提案内容を供給 |
| `web/components/order/MainScreen.tsx` | 連携（design system 所有） | `useSuggestion()` を呼び `screenState` を決定 |
| `web/hooks/useOrder.ts` | 参照（Unit C 所有） | 「押す。」押下時 `placeOrder({suggestionId})` |

> 視覚（CSS トークン・アニメ・globals.css）は design system 実装プラン（PR #94 `_design-system/implementation-plan.md`）が担う。Unit D FD はサジェスト固有の **構造・データ・文言方針**を定義する。

---

## 7. テスト戦略（Standard）

| 種類 | 対象 |
|---|---|
| ユニット（Vitest） | `useSuggestion` の状態（loading / hasSuggestion true・false / エラー丸め）、API レスポンス → 表示マッピング |
| コンポーネント（Testing Library） | `SuggestBubble` の visible 切替、固定文言「そろそろだろ。」、副ラベルの動的値埋め込み |
| PBT（fast-check、任意） | レスポンス JSON → props 変換のラウンドトリップ（Unit A/C と同方針、NFR で範囲確定） |
| E2E（Playwright） | 履歴十分時に SuggestBubble 表示 → 「押す。」1 タップで注文（Unit C と結合） |

---

## 8. オープン項目（後続ステージ）

- 吹き出し文言の AI 動的化（`Title` 活用）は spec §9 / 将来。FD では固定「そろそろだろ。」（BR-D13）
- `useSuggestion` の `staleTime` 具体値・エラー丸めの詳細は NFR Requirements / Code Generation で確定
- `_design-system/design-spec.md` への参照は PR #94 マージ後に解決
