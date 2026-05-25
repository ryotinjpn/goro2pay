# ゴロゴロPay 全画面デザイン刷新 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `docs/superpowers/specs/2026-05-25-app-design.md` (v2.0) で定義した「Slot Machine」コンセプト & リヴァイ調コピーで、ゴロゴロPay の全主要画面 (ランディング / サインアップ / ログイン / メイン / 注文完了 / モーダル各種 / Toast) を刷新する。

**Architecture:**
共通基盤として CSS 変数 (デザイントークン) と次フォントを導入し、`ScreenFrame` をラッパとして全画面に適用。状態管理は既存 react-query mutation + jotai を踏襲し、新規 hook は最小限。コピーは `lib/copy.ts` に集約してテスト可能にする。ビジュアル要素は CSS Modules で世界観を実装し、ロジック部分のみ unit test、画面全体の見た目は Playwright で動作確認する。

**Tech Stack:** Next.js 15 App Router / React 19 / TypeScript / CSS Modules / next/font/google / jotai / @tanstack/react-query / vitest / Playwright

---

## File Structure (実装するファイル)

### 新規作成

| パス | 責務 |
|---|---|
| `web/app/globals.css` | CSS variables (デザイントークン §1.2)、共通 `@keyframes`、`html/body` ベース |
| `web/lib/copy.ts` | リヴァイ調コピー集約 (§1.6) と動的置換 (`composeSuggestSubLabel(store, amount)` 等) |
| `web/lib/copy.test.ts` | 上記の unit test |
| `web/lib/budgetMath.ts` | 増額値計算 (`recommendNextBudget(current)`、現在予算 × 5/3 を 10000 単位丸め) |
| `web/lib/budgetMath.test.ts` | 上記の unit test (PBT) |
| `web/state/main.ts` | `screenStateAtom` (`'idle'\|'suggested'\|'slot'\|'dead'`)、`balanceAtom` |
| `web/components/order/ScreenFrame.tsx` | 全画面共通ラッパ (CRT スキャン + 背景グラデ) |
| `web/components/order/ScreenFrame.module.css` | 上記の CSS |
| `web/components/order/BrandHeader.tsx` | 上部ブランド表示 + 時刻 |
| `web/components/order/BrandHeader.module.css` | 上記の CSS |
| `web/components/order/BalanceHero.tsx` | 残高 + メーター + メトリクス |
| `web/components/order/BalanceHero.module.css` | 上記の CSS |
| `web/components/order/SlotReel.tsx` | スロットリール |
| `web/components/order/SlotReel.module.css` | 上記の CSS |
| `web/components/order/SuggestBubble.tsx` | サジェスト吹き出し |
| `web/components/order/SuggestBubble.module.css` | 上記の CSS |
| `web/components/order/DeadVerdict.tsx` | DEAD オーバーレイ |
| `web/components/order/DeadVerdict.module.css` | 上記の CSS |
| `web/components/order/IncreaseBudgetButton.tsx` | DEAD 時の増額赤ボタン |
| `web/components/order/IncreaseBudgetButton.module.css` | 上記の CSS |
| `web/components/order/MainScreen.tsx` | 認証後メイン画面の orchestrator |
| `web/components/order/MainScreen.module.css` | 上記の CSS |
| `web/components/order/GoroButton.module.css` | (改修) 改修後の CSS |
| `web/components/auth/LandingScreen.module.css` | (改修) 改修後の CSS |
| `web/components/auth/SignupScreen.module.css` | (改修) 改修後の CSS |
| `web/components/auth/LoginScreen.module.css` | (改修) 改修後の CSS |
| `web/components/auth/LogoutConfirmModal.module.css` | (改修) 改修後の CSS |
| `web/components/auth/SessionExpiredModalHost.module.css` | (改修) 改修後の CSS |
| `web/components/order/Toast.module.css` | (改修) 改修後の CSS |
| `web/app/order/[id]/complete/CompleteScreen.module.css` | (改修) 完了画面の CSS |
| `web/tests/copy.test.ts` | (vitest 配置上 `web/lib/` ではなく `web/tests/` に置く既存規約に合わせる場合) |

### 既存ファイル改修

| パス | 変更概要 |
|---|---|
| `web/app/layout.tsx` | `next/font/google` で 3 フォントロード、`globals.css` import |
| `web/app/page.tsx` | インラインスタイル削除、`<MainScreen />` へ委譲 |
| `web/components/order/GoroButton.tsx` | `screenState` props を受け取り 4 状態を表現、CSS Module 化 |
| `web/components/auth/LandingScreen.tsx` | §5.1 のレイアウト + デモボタン挙動 |
| `web/components/auth/SignupScreen.tsx` | §5.2 の見出し、CSS Module 化 |
| `web/components/auth/LoginScreen.tsx` | §5.3 の見出し、CSS Module 化 |
| `web/app/order/[id]/complete/page.tsx` | §5.4 のレイアウト + 演出 |
| `web/components/auth/LogoutConfirmModal.tsx` | §5.5 のスタイル + コピー差し替え |
| `web/components/auth/SessionExpiredModalHost.tsx` | §5.6 のスタイル + コピー差し替え |
| `web/components/order/Toast.tsx` | §5.7 のスタイル変更 (色 / フォント) |
| `web/components/order/ToastHost.tsx` | bottom 余白の微調整のみ |
| `web/state/auth.ts` | (変更なし) |
| `web/components/order/OrderHistoryList.tsx` | メイン画面から外す (削除はせず) |

### 注意事項

- 既存テストファイル (`web/tests/*.test.tsx`) のうち、UI に依存するもの (例: `signupScreen.test.tsx` の文言検証) は新コピーに合わせて修正が必要。詳細は各タスクで明記。
- `useBalance()` hook は spec で言及していたが、本プランでは実装しない。既存の `PlaceOrderResponse.remainingBalance` を `balanceAtom` (jotai) に書き込む方式とする (新規 GET API を作らない)。
- フォント追加: `next/font/google` の Cormorant Garamond / Zen Old Mincho / DotGothic16。`package.json` への追加 dep は不要 (next/font は標準同梱)。
- a11y テスト用 `@axe-core/playwright` は spec で予告したが、本プランでは Playwright スナップショットだけにとどめる (axe 追加は将来課題)。

---

## Task Order

タスクは依存関係に沿ってこの順で進める。各タスクの最後に commit 1 回を入れる。

1. デザイントークン (globals.css) + フォント
2. コピー集約モジュール (`lib/copy.ts`)
3. 増額値計算 (`lib/budgetMath.ts`)
4. メイン画面 state atom (`state/main.ts`)
5. `ScreenFrame` コンポーネント
6. `BrandHeader`
7. `BalanceHero` (MeterBar + MetricsRow 含む)
8. `SuggestBubble`
9. `SlotReel`
10. `GoroButton` 改修 (4 状態)
11. `DeadVerdict` + `IncreaseBudgetButton`
12. `MainScreen` 統合
13. `app/page.tsx` 結線
14. ランディング画面
15. サインアップ画面
16. ログイン画面
17. 注文完了画面
18. ログアウトモーダル
19. セッション切れモーダル
20. Toast スタイル更新
21. 既存テストの文言修正 + Playwright スモーク
22. 仕上げ (lint / type-check / 全画面手動確認)

---

## Task 1: デザイントークン (globals.css) + フォント

**Files:**
- Create: `web/app/globals.css`
- Modify: `web/app/layout.tsx`

- [ ] **Step 1: `globals.css` を作成し、CSS variables とリセットを書く**

```css
/* web/app/globals.css */

:root {
  /* Color tokens (spec §1.2) */
  --color-bg-deep: #050402;
  --color-bg-mid: #0e0a06;
  --color-bg-warm: #1a1208;
  --color-gold-100: #f4d990;
  --color-gold-500: #c9a96b;
  --color-gold-700: #8a6f3a;
  --color-gold-900: #3a2a14;
  --color-cream: #f4ecd8;
  --color-paper: #ffe7c2;
  --color-mute: #8a7a5a;
  --color-accent-warn: #ffb14a;
  --color-accent-danger: #ff4f4f;
  --color-dead-gray: #6a6a6a;

  /* Spacing & shape */
  --radius-card: 18px;
  --pad-screen: 28px 22px 24px;

  /* Motion (spec §1.5) */
  --easing-spring: cubic-bezier(0.2, 0.8, 0.2, 1);
  --duration-fast: 200ms;
  --duration-mid: 450ms;
  --duration-dead: 600ms;
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html,
body {
  background:
    radial-gradient(ellipse at top, var(--color-bg-warm) 0%, var(--color-bg-mid) 60%, var(--color-bg-deep) 100%);
  color: var(--color-cream);
  min-height: 100vh;
  font-family: var(--font-body-sans), system-ui, sans-serif;
}

/* Common animations (spec §4.1) */
@keyframes slot-spin {
  0% { transform: translateY(0); }
  100% { transform: translateY(-100%); }
}

@keyframes balance-tick {
  0% { transform: translateY(0); }
  100% { transform: translateY(-1em); }
}

@keyframes suggest-breath {
  0%, 100% { box-shadow: 0 0 30px rgba(201, 169, 107, 0.4); }
  50% { box-shadow: 0 0 30px rgba(201, 169, 107, 0.6); }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

/* sr-only for screen readers */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
```

- [ ] **Step 2: `layout.tsx` でフォントをロードし `globals.css` を import**

```tsx
// web/app/layout.tsx
import type { Metadata } from "next";
import {
  Cormorant_Garamond,
  Zen_Old_Mincho,
  DotGothic16,
  Inter,
} from "next/font/google";

import { AppProviders } from "./providers";
import { ToastHost } from "@/components/order/ToastHost";

import "./globals.css";

const fontDisplaySerif = Cormorant_Garamond({
  subsets: ["latin"],
  weight: ["500", "700"],
  style: ["normal", "italic"],
  display: "swap",
  variable: "--font-display-serif",
});

const fontBodyMincho = Zen_Old_Mincho({
  subsets: ["latin"],
  weight: ["500", "700", "900"],
  display: "swap",
  variable: "--font-body-mincho",
});

const fontMonoPixel = DotGothic16({
  subsets: ["latin"],
  weight: ["400"],
  display: "swap",
  variable: "--font-mono-pixel",
});

const fontBodySans = Inter({
  subsets: ["latin"],
  weight: ["600", "800"],
  display: "swap",
  variable: "--font-body-sans",
});

export const metadata: Metadata = {
  title: "ゴロゴロPay",
  description: "面倒は、こちらで引き受ける。",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="ja"
      className={[
        fontDisplaySerif.variable,
        fontBodyMincho.variable,
        fontMonoPixel.variable,
        fontBodySans.variable,
      ].join(" ")}
    >
      <body>
        <AppProviders>
          {children}
          <ToastHost />
        </AppProviders>
      </body>
    </html>
  );
}
```

注意: `Zen_Old_Mincho` と `DotGothic16` は subset として `latin` を指定 (next/font は `japanese` subset 非対応のため、Google Fonts の latin range をフォールバックとして使う。日本語グリフは self-host 設定で別途取り込む必要があれば後続で対応)。

- [ ] **Step 3: 動作確認**

Run: `cd web && npm run lint`
Expected: `globals.css` import に lint エラー無し

Run: `cd web && npm run dev` で `/` を開く
Expected: 既存画面が暗茶背景 + cream 文字で表示される (UI 自体は未刷新)

- [ ] **Step 4: Commit**

```bash
git add web/app/globals.css web/app/layout.tsx
git commit -m "🎨 feat(web): デザイントークンと next/font を導入"
```

---

## Task 2: コピー集約モジュール (`lib/copy.ts`)

**Files:**
- Create: `web/lib/copy.ts`
- Create: `web/tests/copy.test.ts`

- [ ] **Step 1: 失敗テストを書く**

```ts
// web/tests/copy.test.ts
import { describe, it, expect } from "vitest";
import {
  COPY,
  composeSuggestSubLabel,
  composeIncreaseBudgetLabel,
  composeMonthlyMeta,
} from "@/lib/copy";

describe("COPY constants", () => {
  it("contains landing hero", () => {
    expect(COPY.landing.heroLine1).toBe("考えるな。");
    expect(COPY.landing.heroLine2).toBe("押せ。");
    expect(COPY.landing.sub).toBe("面倒は、こちらで引き受ける。");
  });
  it("contains main button labels", () => {
    expect(COPY.main.idleMain).toBe("めんどくさい");
    expect(COPY.main.idleSub).toBe("— 押せ。考えるな。");
    expect(COPY.main.suggestMain).toBe("押す。");
  });
  it("contains dead state copy", () => {
    expect(COPY.main.deadVerdict).toBe("今月は、終わりだ。");
    expect(COPY.main.deadButtonSub).toBe("— 上出来だ。使い切ったな。");
  });
  it("contains complete screen copy", () => {
    expect(COPY.complete.verdict).toBe("いい判断だ。");
    expect(COPY.complete.body).toBe("面倒は片付いた。");
    expect(COPY.complete.next).toBe("次を待て。");
  });
});

describe("composeSuggestSubLabel", () => {
  it("formats store + amount with リヴァイ調 ending", () => {
    expect(composeSuggestSubLabel("CoCo壱", 1200)).toBe("— CoCo壱 ¥1,200 だ。");
  });
  it("formats large numbers with thousand separator", () => {
    expect(composeSuggestSubLabel("ロイヤルホスト", 12500)).toBe(
      "— ロイヤルホスト ¥12,500 だ。",
    );
  });
});

describe("composeIncreaseBudgetLabel", () => {
  it("formats yen with thousand separator", () => {
    expect(composeIncreaseBudgetLabel(50000)).toBe("¥50,000。来月もこの調子だ。");
  });
});

describe("composeMonthlyMeta", () => {
  it("counts good judgments", () => {
    expect(composeMonthlyMeta(6)).toBe("今月 6 度、いい判断だった。");
  });
  it("works with 1", () => {
    expect(composeMonthlyMeta(1)).toBe("今月 1 度、いい判断だった。");
  });
});
```

- [ ] **Step 2: テストが失敗することを確認**

Run: `cd web && npm run test -- copy`
Expected: FAIL — `Cannot find module '@/lib/copy'`

- [ ] **Step 3: `lib/copy.ts` を実装**

```ts
// web/lib/copy.ts
//
// リヴァイ調コピー集約 (spec §1.6)。
// UI レイヤーから直接文字列リテラルを書かず、必ずこのモジュール経由で参照する。

export const COPY = {
  brand: {
    name: "ゴロゴロPay", // ロゴ用
  },
  landing: {
    heroLine1: "考えるな。",
    heroLine2: "押せ。",
    sub: "面倒は、こちらで引き受ける。",
    demoBalanceLabel: "体験用ダメ予算",
    demoButtonMain: "押す。",
    demoButtonAfter: "— 上出来だ。",
    ctaPrimary: "始めろ",
    ctaSecondary: "戻る",
  },
  signup: {
    h1Line1: "面倒は、",
    h1Line2: "こちらで引き受ける。",
    sub: "30 秒で済む。",
    submit: "登録する",
    submitting: "登録中…",
  },
  login: {
    h1: "戻ってきたか。",
    sub: "また面倒になったか。",
    submit: "ログイン",
    submitting: "ログイン中…",
  },
  main: {
    balanceLabel: "残りダメ予算",
    idleMain: "めんどくさい",
    idleSub: "— 押せ。考えるな。",
    suggestMain: "押す。",
    suggestBubble: "そろそろだろ。",
    slotCaption: "▼ 選定中 ▼",
    deadVerdict: "今月は、終わりだ。",
    deadButtonSub: "— 上出来だ。使い切ったな。",
  },
  complete: {
    verdict: "いい判断だ。",
    body: "面倒は片付いた。",
    next: "次を待て。",
  },
  logout: {
    h2: "やめるのか？",
    sub: "戻ってこい。",
    primary: "やめる。",
    secondary: "戻る。",
  },
  sessionExpired: {
    h2: "離れすぎたな。",
    button: "戻る。",
  },
} as const;

const yenFormat = new Intl.NumberFormat("ja-JP");

export function composeSuggestSubLabel(storeName: string, amount: number): string {
  return `— ${storeName} ¥${yenFormat.format(amount)} だ。`;
}

export function composeIncreaseBudgetLabel(nextBudget: number): string {
  return `¥${yenFormat.format(nextBudget)}。来月もこの調子だ。`;
}

export function composeMonthlyMeta(count: number): string {
  return `今月 ${count} 度、いい判断だった。`;
}
```

- [ ] **Step 4: テストが通ることを確認**

Run: `cd web && npm run test -- copy`
Expected: PASS — 12 件のアサーション全部通る

- [ ] **Step 5: Commit**

```bash
git add web/lib/copy.ts web/tests/copy.test.ts
git commit -m "🎨 feat(web): リヴァイ調コピーを lib/copy.ts に集約"
```

---

## Task 3: 増額値計算 (`lib/budgetMath.ts`)

**Files:**
- Create: `web/lib/budgetMath.ts`
- Create: `web/tests/budgetMath.test.ts`

- [ ] **Step 1: 失敗テストを書く (Property-Based Testing)**

```ts
// web/tests/budgetMath.test.ts
import { describe, it, expect } from "vitest";
import * as fc from "fast-check";
import { recommendNextBudget } from "@/lib/budgetMath";

describe("recommendNextBudget", () => {
  it("returns 50000 for 30000", () => {
    expect(recommendNextBudget(30000)).toBe(50000);
  });
  it("returns 80000 for 50000", () => {
    expect(recommendNextBudget(50000)).toBe(80000);
  });
  it("returns 130000 for 80000", () => {
    expect(recommendNextBudget(80000)).toBe(130000);
  });

  it("rounds to nearest 10000", () => {
    fc.assert(
      fc.property(fc.integer({ min: 10000, max: 1_000_000 }), (current) => {
        const next = recommendNextBudget(current);
        expect(next % 10000).toBe(0);
      }),
    );
  });

  it("always increases (next > current)", () => {
    fc.assert(
      fc.property(fc.integer({ min: 10000, max: 1_000_000 }), (current) => {
        expect(recommendNextBudget(current)).toBeGreaterThan(current);
      }),
    );
  });

  it("approximates 5/3 ratio (within rounding tolerance)", () => {
    fc.assert(
      fc.property(fc.integer({ min: 30000, max: 500_000 }), (current) => {
        const next = recommendNextBudget(current);
        const ratio = next / current;
        expect(ratio).toBeGreaterThanOrEqual(1.5);
        expect(ratio).toBeLessThanOrEqual(2.0);
      }),
    );
  });
});
```

- [ ] **Step 2: テストが失敗することを確認**

Run: `cd web && npm run test -- budgetMath`
Expected: FAIL — `Cannot find module '@/lib/budgetMath'`

- [ ] **Step 3: `lib/budgetMath.ts` を実装**

```ts
// web/lib/budgetMath.ts
//
// DEAD 時の増額提案値を計算する (spec §3.5 / FR-METRICS-04)。
// 算出式: 現在予算 × 5/3 を 10,000 円単位で四捨五入。
// AI による推奨は将来検討 (spec §9)。

const RATIO = 5 / 3;
const ROUND_UNIT = 10000;

export function recommendNextBudget(current: number): number {
  const raw = current * RATIO;
  const rounded = Math.round(raw / ROUND_UNIT) * ROUND_UNIT;
  // 安全側として、四捨五入で current 以下になるケースを防ぐ
  return rounded > current ? rounded : current + ROUND_UNIT;
}
```

- [ ] **Step 4: テストが通ることを確認**

Run: `cd web && npm run test -- budgetMath`
Expected: PASS — 全 PBT が緑

- [ ] **Step 5: Commit**

```bash
git add web/lib/budgetMath.ts web/tests/budgetMath.test.ts
git commit -m "🎨 feat(web): 増額提案値の計算ロジックを追加"
```

---

## Task 4: メイン画面 state atom (`state/main.ts`)

**Files:**
- Create: `web/state/main.ts`

- [ ] **Step 1: atom を実装**

```ts
// web/state/main.ts
//
// メイン画面の有限状態 (spec §2.3) と残高を保持する jotai atom 群。
// useBalance hook を作らずに、注文 mutation の onSuccess で
// balanceAtom を直接書き込む方針。

import { atom } from "jotai";

export type ScreenState = "idle" | "suggested" | "slot" | "dead";

export const screenStateAtom = atom<ScreenState>("idle");

// 残高は注文成功時に PlaceOrderResponse.remainingBalance から書き込む。
// 初期値は -1 (未取得) とし、未取得状態は UI 側で「---」表示する。
export const balanceAtom = atom<number>(-1);

// 月内のダメ化回数。注文成功時にインクリメント。
export const monthlyCountAtom = atom<number>(0);

// 月予算 (固定 30000、設定機能は将来)。消化率計算に使う。
export const monthlyBudgetAtom = atom<number>(30000);

// 消化率 (派生 atom)。0–1 の範囲。
export const consumeRateAtom = atom((get) => {
  const balance = get(balanceAtom);
  const budget = get(monthlyBudgetAtom);
  if (balance < 0 || budget <= 0) return 0;
  const used = Math.max(0, budget - balance);
  return Math.min(1, used / budget);
});

// 現在のサジェスト (Bedrock 応答)。null = サジェストなし。
export type Suggestion = {
  storeName: string;
  amount: number;
  suggestionId: string;
};

export const suggestionAtom = atom<Suggestion | null>(null);
```

- [ ] **Step 2: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 3: Commit**

```bash
git add web/state/main.ts
git commit -m "🎨 feat(web): メイン画面の screenState / balance atom を追加"
```

---

## Task 5: ScreenFrame コンポーネント

**Files:**
- Create: `web/components/order/ScreenFrame.tsx`
- Create: `web/components/order/ScreenFrame.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/order/ScreenFrame.module.css */

.frame {
  position: relative;
  min-height: 100vh;
  padding: var(--pad-screen);
  max-width: 480px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  isolation: isolate;
}

/* CRT scanlines overlay */
.frame::before {
  content: "";
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(
    0deg,
    rgba(255, 255, 255, 0.025) 0 1px,
    transparent 1px 4px
  );
  z-index: 0;
}

/* Vignette glow at top */
.frame::after {
  content: "";
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: radial-gradient(
    ellipse at 50% -20%,
    rgba(201, 169, 107, 0.22),
    transparent 55%
  );
  z-index: 0;
}

.frame > * {
  position: relative;
  z-index: 1;
}

/* DEAD 状態で saturate を落とす */
.frame[data-screen-state="dead"] {
  filter: saturate(0.2) brightness(0.7);
  transition: filter var(--duration-dead) ease-out;
}
```

- [ ] **Step 2: コンポーネントを実装**

```tsx
// web/components/order/ScreenFrame.tsx
//
// 全画面共通のラッパ。CRT スキャンライン + 背景ヴィネットを与え、
// data-screen-state を CSS に渡して状態別スタイリングを可能にする (spec §2)。
"use client";

import { ReactNode } from "react";
import type { ScreenState } from "@/state/main";

import styles from "./ScreenFrame.module.css";

type Props = {
  children: ReactNode;
  screenState?: ScreenState;
  testid?: string;
};

export function ScreenFrame({ children, screenState, testid }: Props) {
  return (
    <main
      className={styles.frame}
      data-screen-state={screenState}
      data-testid={testid}
    >
      {children}
    </main>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/order/ScreenFrame.tsx web/components/order/ScreenFrame.module.css
git commit -m "🎨 feat(web): ScreenFrame ラッパを追加"
```

---

## Task 6: BrandHeader

**Files:**
- Create: `web/components/order/BrandHeader.tsx`
- Create: `web/components/order/BrandHeader.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/order/BrandHeader.module.css */

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 10px;
  letter-spacing: 0.2em;
  color: var(--color-mute);
  margin-top: 4px;
}

.brand {
  font-family: var(--font-body-mincho), serif;
  font-weight: 900;
  font-size: 12px;
  letter-spacing: 0.04em;
  color: var(--color-gold-500);
}

.brand .pay {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  margin-left: 2px;
}

.time {
  font-feature-settings: "tnum";
}
```

- [ ] **Step 2: コンポーネントを実装**

```tsx
// web/components/order/BrandHeader.tsx
//
// 上部のブランド + 現在時刻 (spec §2.1)。時刻は client 側で 1 分ごとに更新。
"use client";

import { useEffect, useState } from "react";

import { COPY } from "@/lib/copy";

import styles from "./BrandHeader.module.css";

function formatTime(d: Date): string {
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${hh}:${mm}`;
}

export function BrandHeader() {
  const [time, setTime] = useState<string>(() => formatTime(new Date()));

  useEffect(() => {
    const id = setInterval(() => setTime(formatTime(new Date())), 60_000);
    return () => clearInterval(id);
  }, []);

  // COPY.brand.name = "ゴロゴロPay"。Pay 部分を Cormorant Italic にするため分割。
  const name = COPY.brand.name;
  const splitAt = name.indexOf("Pay");
  const left = splitAt >= 0 ? name.slice(0, splitAt) : name;
  const right = splitAt >= 0 ? name.slice(splitAt) : "";

  return (
    <header className={styles.header}>
      <span className={styles.brand}>
        {left}
        {right && <span className={styles.pay}>{right}</span>}
      </span>
      <span className={styles.time}>{time}</span>
    </header>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/order/BrandHeader.tsx web/components/order/BrandHeader.module.css
git commit -m "🎨 feat(web): BrandHeader を追加"
```

---

## Task 7: BalanceHero (MeterBar + MetricsRow 含む)

**Files:**
- Create: `web/components/order/BalanceHero.tsx`
- Create: `web/components/order/BalanceHero.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/order/BalanceHero.module.css */

.block {
  margin-top: 18px;
  text-align: center;
}

.label {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 9px;
  letter-spacing: 0.35em;
  color: var(--color-mute);
}

.amount {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 56px;
  letter-spacing: -0.02em;
  line-height: 1;
  margin-top: 6px;
  color: var(--color-cream);
  text-shadow: 0 0 22px rgba(244, 236, 216, 0.18);
  font-variant-numeric: tabular-nums;
}

.amount[data-low="true"] {
  color: var(--color-accent-danger);
  text-shadow: 0 0 22px rgba(255, 79, 79, 0.4);
}

.amount[data-dead="true"] {
  color: var(--color-dead-gray);
  text-shadow: none;
}

.yen {
  font-style: normal;
  color: var(--color-gold-500);
  font-size: 28px;
  margin-right: 4px;
}

.amount[data-dead="true"] .yen {
  color: var(--color-dead-gray);
}

.meter {
  height: 2px;
  background: rgba(201, 169, 107, 0.18);
  margin: 12px 14px 0;
  position: relative;
  overflow: hidden;
}

.meterFill {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  background: var(--color-gold-500);
  transition: width var(--duration-fast) ease, background-color 400ms ease;
}

.meterFill[data-warn="true"] {
  background: var(--color-accent-warn);
}

.meterFill[data-danger="true"] {
  background: var(--color-accent-danger);
  animation: meter-pulse 400ms ease-out;
}

@keyframes meter-pulse {
  0% { box-shadow: 0 0 0 var(--color-accent-danger); }
  50% { box-shadow: 0 0 10px var(--color-accent-danger); }
  100% { box-shadow: 0 0 0 var(--color-accent-danger); }
}

.metrics {
  margin-top: 10px;
  display: flex;
  justify-content: space-between;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 9px;
  letter-spacing: 0.2em;
  color: var(--color-mute);
}
```

- [ ] **Step 2: コンポーネントを実装**

```tsx
// web/components/order/BalanceHero.tsx
//
// 残高 + メーター + メトリクス (spec §2.1, §3.1, §3.2)。
"use client";

import { useAtomValue } from "jotai";

import { COPY } from "@/lib/copy";
import {
  balanceAtom,
  consumeRateAtom,
  monthlyCountAtom,
  screenStateAtom,
} from "@/state/main";

import styles from "./BalanceHero.module.css";

const yenFormat = new Intl.NumberFormat("ja-JP");

export function BalanceHero() {
  const balance = useAtomValue(balanceAtom);
  const consumeRate = useAtomValue(consumeRateAtom);
  const monthlyCount = useAtomValue(monthlyCountAtom);
  const screenState = useAtomValue(screenStateAtom);

  const isDead = screenState === "dead" || balance === 0;
  const isLow = !isDead && consumeRate >= 0.8;
  const isWarn = !isDead && !isLow && consumeRate >= 0.6;
  const fillPct = Math.round(consumeRate * 100);

  const balanceText =
    balance < 0 ? "---" : yenFormat.format(balance);

  return (
    <section
      className={styles.block}
      role="status"
      aria-live="polite"
      data-testid="balance-hero"
    >
      <div className={styles.label}>{COPY.main.balanceLabel}</div>
      <div
        className={styles.amount}
        data-low={isLow}
        data-dead={isDead}
      >
        <span className={styles.yen}>¥</span>
        {balanceText}
      </div>
      <div className={styles.meter} aria-hidden="true">
        <div
          className={styles.meterFill}
          style={{ width: `${fillPct}%` }}
          data-warn={isWarn}
          data-danger={isLow}
        />
      </div>
      <div className={styles.metrics}>
        <span>今月 {monthlyCount} 度</span>
        <span>消化 {fillPct}%</span>
      </div>
    </section>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/order/BalanceHero.tsx web/components/order/BalanceHero.module.css
git commit -m "🎨 feat(web): BalanceHero (残高 / メーター / メトリクス) を追加"
```

---

## Task 8: SuggestBubble

**Files:**
- Create: `web/components/order/SuggestBubble.tsx`
- Create: `web/components/order/SuggestBubble.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/order/SuggestBubble.module.css */

.bubble {
  position: absolute;
  bottom: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
  background: var(--color-bg-warm);
  color: var(--color-cream);
  font-family: var(--font-body-mincho), serif;
  font-weight: 700;
  font-size: 12px;
  letter-spacing: 0.04em;
  padding: 8px 12px;
  border: 1px solid var(--color-gold-500);
  box-shadow: 0 0 14px rgba(201, 169, 107, 0.5);
  white-space: nowrap;
  opacity: 0;
  animation: bubble-in 600ms ease forwards;
}

.bubble::after {
  content: "";
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 6px solid transparent;
  border-top-color: var(--color-gold-500);
}

@keyframes bubble-in {
  from {
    opacity: 0;
    transform: translate(-50%, -4px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}
```

- [ ] **Step 2: コンポーネントを実装**

```tsx
// web/components/order/SuggestBubble.tsx
//
// サジェスト時にボタン上部に乗る吹き出し (spec §3.3)。
"use client";

import { COPY } from "@/lib/copy";
import styles from "./SuggestBubble.module.css";

export function SuggestBubble() {
  return (
    <div className={styles.bubble} data-testid="suggest-bubble">
      {COPY.main.suggestBubble}
    </div>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/order/SuggestBubble.tsx web/components/order/SuggestBubble.module.css
git commit -m "🎨 feat(web): SuggestBubble (サジェスト吹き出し) を追加"
```

---

## Task 9: SlotReel

**Files:**
- Create: `web/components/order/SlotReel.tsx`
- Create: `web/components/order/SlotReel.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/order/SlotReel.module.css */

.reelWrap {
  width: 80%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.window {
  width: 100%;
  background: #0a0405;
  border: 2px solid var(--color-gold-500);
  padding: 10px 4px;
  overflow: hidden;
  position: relative;
  height: 56px;
  border-radius: 4px;
}

.strip {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 14px;
  color: var(--color-gold-100);
  text-shadow: 0 0 8px rgba(244, 217, 144, 0.6);
  letter-spacing: 0.04em;
  line-height: 1.4;
  animation: slot-spin 0.4s linear infinite;
  text-align: center;
}

.strip[data-stopped="true"] {
  animation: none;
  transition: transform 600ms var(--easing-spring);
}

.caption {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 9px;
  color: var(--color-paper);
  letter-spacing: 0.25em;
}
```

- [ ] **Step 2: コンポーネントを実装**

```tsx
// web/components/order/SlotReel.tsx
//
// 注文中のスロットリール (spec §3.4)。候補を高速スクロールし、
// `winningText` が確定したら停止する。
"use client";

import { COPY } from "@/lib/copy";

import styles from "./SlotReel.module.css";

const DUMMY_CANDIDATES = [
  "CoCo壱番屋",
  "すき家",
  "ロイヤルホスト",
  "松屋",
  "吉野家",
  "サブウェイ",
];

type Props = {
  winningText: string | null; // null = まだ回転中
};

export function SlotReel({ winningText }: Props) {
  const stopped = winningText !== null;
  const items = stopped ? [winningText] : DUMMY_CANDIDATES;

  return (
    <div className={styles.reelWrap} data-testid="slot-reel">
      <div className={styles.window}>
        <div
          className={styles.strip}
          data-stopped={stopped}
          aria-hidden="true"
        >
          {[...items, ...items].map((item, i) => (
            <div key={i}>{item}</div>
          ))}
        </div>
      </div>
      <div className={styles.caption}>{COPY.main.slotCaption}</div>
      {stopped && winningText && (
        <span className="sr-only">注文確定: {winningText}</span>
      )}
    </div>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/order/SlotReel.tsx web/components/order/SlotReel.module.css
git commit -m "🎨 feat(web): SlotReel (注文中スロット) を追加"
```

---

## Task 10: GoroButton 改修 (4 状態)

**Files:**
- Modify: `web/components/order/GoroButton.tsx`
- Create: `web/components/order/GoroButton.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/order/GoroButton.module.css */

.wrap {
  flex: 1;
  display: grid;
  place-items: center;
  margin-top: 18px;
  position: relative;
}

.button {
  width: 78%;
  aspect-ratio: 1;
  border-radius: 50%;
  background: radial-gradient(
    circle at 35% 30%,
    var(--color-gold-100) 0%,
    var(--color-gold-500) 35%,
    var(--color-gold-700) 70%,
    var(--color-gold-900) 100%
  );
  box-shadow:
    0 0 0 4px var(--color-gold-500),
    0 0 50px rgba(201, 169, 107, 0.4),
    inset 0 -16px 28px rgba(0, 0, 0, 0.45),
    inset 0 6px 14px rgba(255, 240, 200, 0.35);
  border: 4px solid rgba(0, 0, 0, 0.4);
  display: grid;
  place-items: center;
  text-align: center;
  cursor: pointer;
  transition: transform 200ms var(--easing-spring),
    box-shadow 200ms ease;
  position: relative;
}

.button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.button:active:not(:disabled) {
  transform: scale(0.98);
  box-shadow:
    0 0 0 4px var(--color-gold-500),
    0 0 30px rgba(201, 169, 107, 0.3),
    inset 0 -8px 16px rgba(0, 0, 0, 0.5);
  transition-duration: 120ms;
}

.button:focus-visible {
  outline: 3px solid var(--color-gold-100);
  outline-offset: 4px;
}

.button[data-state="suggested"] {
  animation: suggest-breath 2.4s ease-in-out infinite;
}

.button[data-state="slot"] {
  background: radial-gradient(
    circle at 35% 30%,
    #6a4f1a 0%,
    #3a2a14 100%
  );
  cursor: progress;
}

.button[data-state="dead"] {
  background: radial-gradient(
    circle at 35% 30%,
    #6a6a6a 0%,
    #3a3a3a 100%
  );
  box-shadow:
    0 0 0 4px #4a4a4a,
    inset 0 0 30px rgba(0, 0, 0, 0.7);
  cursor: not-allowed;
}

.labelMain {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 18px;
  letter-spacing: 0.14em;
  color: var(--color-bg-warm);
}

.button[data-state="dead"] .labelMain {
  color: #4a4a4a;
}

.labelSub {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-size: 11px;
  letter-spacing: 0.15em;
  color: var(--color-gold-900);
  margin-top: 6px;
}

.button[data-state="dead"] .labelSub {
  color: #4a4a4a;
}
```

- [ ] **Step 2: GoroButton.tsx を改修**

```tsx
// web/components/order/GoroButton.tsx
//
// メイン CTA (spec §2.2, §3)。screenState を受け取り、
// idle / suggested / slot / dead の 4 状態を表現する。
"use client";

import { useRouter } from "next/navigation";
import { useAtomValue, useSetAtom } from "jotai";
import { useEffect, useState } from "react";

import { useOrder } from "@/hooks/useOrder";
import { generateUlid } from "@/lib/ulid";
import { COPY, composeSuggestSubLabel } from "@/lib/copy";
import {
  balanceAtom,
  monthlyCountAtom,
  screenStateAtom,
  suggestionAtom,
  type ScreenState,
} from "@/state/main";

import { SlotReel } from "./SlotReel";
import { SuggestBubble } from "./SuggestBubble";

import styles from "./GoroButton.module.css";

const SLOT_MIN_DURATION_MS = 1200;

export function GoroButton() {
  const router = useRouter();
  const screenState = useAtomValue(screenStateAtom);
  const setScreenState = useSetAtom(screenStateAtom);
  const setBalance = useSetAtom(balanceAtom);
  const setMonthlyCount = useSetAtom(monthlyCountAtom);
  const suggestion = useAtomValue(suggestionAtom);
  const { mutate, disabled, isPending } = useOrder();

  // SLOT 演出の最低時間を確保するため、API 応答後に表示する winningText を
  // 別 state で管理する。
  const [winningText, setWinningText] = useState<string | null>(null);
  const [slotStartedAt, setSlotStartedAt] = useState<number>(0);

  // SLOT 状態が解けて IDLE に戻ったとき winningText もリセット
  useEffect(() => {
    if (screenState !== "slot") {
      setWinningText(null);
    }
  }, [screenState]);

  const handleClick = () => {
    if (screenState !== "idle" && screenState !== "suggested") return;
    if (disabled || isPending) return;

    setScreenState("slot");
    setSlotStartedAt(Date.now());
    setWinningText(null);

    mutate(
      {
        category: "food",
        idempotencyKey: generateUlid(),
        suggestionId: suggestion?.suggestionId,
      },
      {
        onSuccess: (res) => {
          const elapsed = Date.now() - slotStartedAt;
          const remaining = Math.max(0, SLOT_MIN_DURATION_MS - elapsed);
          setTimeout(() => {
            setWinningText(`${res.storeName} ¥${res.amount.toLocaleString("ja-JP")}`);
            setBalance(res.remainingBalance);
            setMonthlyCount((c) => c + 1);
            setTimeout(() => {
              setScreenState(res.remainingBalance === 0 ? "dead" : "idle");
              router.push(`/order/${res.orderId}/complete`);
            }, 800);
          }, remaining);
        },
        onError: () => {
          // useOrder 側で Toast 表示済み。残高不足時は dead へ。
          setScreenState("idle");
        },
      },
    );
  };

  const labelMain = labelMainFor(screenState);
  const labelSub = labelSubFor(screenState, suggestion);
  const ariaLabel = ariaLabelFor(screenState, suggestion);

  return (
    <div className={styles.wrap}>
      <button
        type="button"
        onClick={handleClick}
        disabled={disabled || screenState === "slot" || screenState === "dead"}
        data-testid="goro-button"
        data-state={screenState}
        aria-label={ariaLabel}
        className={styles.button}
      >
        {screenState === "suggested" && <SuggestBubble />}
        {screenState === "slot" ? (
          <SlotReel winningText={winningText} />
        ) : (
          <div>
            <div className={styles.labelMain}>{labelMain}</div>
            {labelSub && <div className={styles.labelSub}>{labelSub}</div>}
          </div>
        )}
      </button>
    </div>
  );
}

function labelMainFor(state: ScreenState): string {
  if (state === "suggested") return COPY.main.suggestMain;
  return COPY.main.idleMain;
}

function labelSubFor(
  state: ScreenState,
  suggestion: { storeName: string; amount: number } | null,
): string | null {
  if (state === "dead") return COPY.main.deadButtonSub;
  if (state === "slot") return null;
  if (state === "suggested" && suggestion) {
    return composeSuggestSubLabel(suggestion.storeName, suggestion.amount);
  }
  return COPY.main.idleSub;
}

function ariaLabelFor(
  state: ScreenState,
  suggestion: { storeName: string; amount: number } | null,
): string {
  if (state === "dead") return "残高不足のため注文できません";
  if (state === "slot") return "注文処理中";
  if (state === "suggested" && suggestion) {
    return `${suggestion.storeName} ¥${suggestion.amount.toLocaleString("ja-JP")} を注文`;
  }
  return "ご飯めんどくさい";
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/order/GoroButton.tsx web/components/order/GoroButton.module.css
git commit -m "🎨 feat(web): GoroButton を 4 状態 + 金色グラデへ刷新"
```

---

## Task 11: DeadVerdict + IncreaseBudgetButton

**Files:**
- Create: `web/components/order/DeadVerdict.tsx`
- Create: `web/components/order/DeadVerdict.module.css`
- Create: `web/components/order/IncreaseBudgetButton.tsx`
- Create: `web/components/order/IncreaseBudgetButton.module.css`

- [ ] **Step 1: DeadVerdict CSS**

```css
/* web/components/order/DeadVerdict.module.css */

.verdict {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-size: 26px;
  color: var(--color-cream);
  text-align: center;
  line-height: 1.25;
  width: 90%;
  letter-spacing: -0.02em;
  filter: none; /* parent の saturate を打ち消す */
  isolation: isolate;
  opacity: 0;
  animation: verdict-in 500ms var(--easing-spring) 800ms forwards;
}

@keyframes verdict-in {
  from {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.96);
  }
  to {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1);
  }
}
```

- [ ] **Step 2: DeadVerdict コンポーネント**

```tsx
// web/components/order/DeadVerdict.tsx
//
// DEAD 状態のオーバーレイ宣告 (spec §3.5)。
"use client";

import { COPY } from "@/lib/copy";

import styles from "./DeadVerdict.module.css";

export function DeadVerdict() {
  return (
    <div
      className={styles.verdict}
      role="status"
      aria-live="polite"
      data-testid="dead-verdict"
    >
      {COPY.main.deadVerdict}
    </div>
  );
}
```

- [ ] **Step 3: IncreaseBudgetButton CSS**

```css
/* web/components/order/IncreaseBudgetButton.module.css */

.button {
  position: absolute;
  bottom: 28px;
  left: 22px;
  right: 22px;
  background: linear-gradient(180deg, var(--color-accent-danger), #c2272d);
  color: #fff;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 13px;
  padding: 14px;
  letter-spacing: 0.15em;
  border: 1px solid #ff8a8a;
  box-shadow: 0 0 30px rgba(255, 79, 79, 0.7);
  cursor: pointer;
  filter: saturate(2) brightness(1.4); /* parent saturate を打ち消す */
  isolation: isolate;
  opacity: 0;
  transform: translateY(60px);
  animation: increase-rise 500ms var(--easing-spring) 1200ms forwards;
}

.button:hover {
  filter: saturate(2) brightness(1.6);
}

.button:focus-visible {
  outline: 3px solid #ff8a8a;
  outline-offset: 3px;
}

@keyframes increase-rise {
  from {
    opacity: 0;
    transform: translateY(60px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
```

- [ ] **Step 4: IncreaseBudgetButton コンポーネント**

```tsx
// web/components/order/IncreaseBudgetButton.tsx
//
// DEAD 時の増額提案ボタン (spec §3.5 / FR-METRICS-04)。
"use client";

import { useAtomValue } from "jotai";

import { composeIncreaseBudgetLabel } from "@/lib/copy";
import { recommendNextBudget } from "@/lib/budgetMath";
import { monthlyBudgetAtom } from "@/state/main";

import styles from "./IncreaseBudgetButton.module.css";

type Props = {
  onClick: (nextBudget: number) => void;
};

export function IncreaseBudgetButton({ onClick }: Props) {
  const currentBudget = useAtomValue(monthlyBudgetAtom);
  const nextBudget = recommendNextBudget(currentBudget);
  const label = composeIncreaseBudgetLabel(nextBudget);

  return (
    <button
      type="button"
      className={styles.button}
      onClick={() => onClick(nextBudget)}
      data-testid="increase-budget-button"
    >
      {label}
    </button>
  );
}
```

- [ ] **Step 5: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 6: Commit**

```bash
git add web/components/order/DeadVerdict.tsx web/components/order/DeadVerdict.module.css web/components/order/IncreaseBudgetButton.tsx web/components/order/IncreaseBudgetButton.module.css
git commit -m "🎨 feat(web): DEAD 状態の宣告と増額誘導ボタンを追加"
```

---

## Task 12: MainScreen 統合

**Files:**
- Create: `web/components/order/MainScreen.tsx`
- Create: `web/components/order/MainScreen.module.css`

- [ ] **Step 1: MainScreen CSS**

```css
/* web/components/order/MainScreen.module.css */

.body {
  display: flex;
  flex-direction: column;
  flex: 1;
}
```

- [ ] **Step 2: MainScreen コンポーネント**

```tsx
// web/components/order/MainScreen.tsx
//
// 認証後メイン画面 orchestrator (spec §2.5)。各サブコンポーネントを並べ、
// screenState に応じて DEAD オーバーレイを出す。
"use client";

import { useAtomValue, useSetAtom } from "jotai";

import {
  monthlyBudgetAtom,
  screenStateAtom,
} from "@/state/main";

import { ScreenFrame } from "./ScreenFrame";
import { BrandHeader } from "./BrandHeader";
import { BalanceHero } from "./BalanceHero";
import { GoroButton } from "./GoroButton";
import { DeadVerdict } from "./DeadVerdict";
import { IncreaseBudgetButton } from "./IncreaseBudgetButton";

import styles from "./MainScreen.module.css";

export function MainScreen() {
  const screenState = useAtomValue(screenStateAtom);
  const setMonthlyBudget = useSetAtom(monthlyBudgetAtom);

  const handleIncrease = (nextBudget: number) => {
    // 月予算更新 = ローカル state 反映。サーバ側 API 連携は将来。
    setMonthlyBudget(nextBudget);
  };

  return (
    <ScreenFrame screenState={screenState} testid="main-screen">
      <BrandHeader />
      <BalanceHero />
      <div className={styles.body}>
        <GoroButton />
      </div>
      {screenState === "dead" && (
        <>
          <DeadVerdict />
          <IncreaseBudgetButton onClick={handleIncrease} />
        </>
      )}
    </ScreenFrame>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/order/MainScreen.tsx web/components/order/MainScreen.module.css
git commit -m "🎨 feat(web): MainScreen orchestrator を追加"
```

---

## Task 13: app/page.tsx 結線

**Files:**
- Modify: `web/app/page.tsx`

- [ ] **Step 1: page.tsx を改修**

```tsx
// web/app/page.tsx
"use client";

import { useAuth } from "@/hooks/useAuth";

import { LandingScreen } from "@/components/auth/LandingScreen";
import { MainScreen } from "@/components/order/MainScreen";

export default function HomePage() {
  const { status } = useAuth();

  if (status === "loading") {
    return null;
  }

  if (status === "unauthenticated") {
    return <LandingScreen />;
  }

  return <MainScreen />;
}
```

- [ ] **Step 2: dev サーバで動作確認**

Run: `cd web && npm run dev` で `/` を開く (要認証 / 仮認証で OK)
Expected: 暗茶背景 + 金色のスロットボタン + 残高表示が出る (まだランディング/フォーム類は旧 UI)

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/app/page.tsx
git commit -m "🎨 feat(web): app/page.tsx を MainScreen に結線"
```

---

## Task 14: ランディング画面

**Files:**
- Modify: `web/components/auth/LandingScreen.tsx`
- Create: `web/components/auth/LandingScreen.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/auth/LandingScreen.module.css */

.hero {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  gap: 12px;
}

.h1 {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 36px;
  color: var(--color-cream);
  line-height: 0.95;
}

.h1 div {
  display: block;
}

.sub {
  font-family: var(--font-body-mincho), serif;
  font-weight: 700;
  font-size: 11px;
  color: var(--color-gold-500);
  letter-spacing: 0.04em;
  margin-top: 4px;
}

.demoLabel {
  margin-top: 14px;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 8px;
  letter-spacing: 0.3em;
  color: var(--color-mute);
}

.demoLabel b {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 22px;
  color: var(--color-cream);
  letter-spacing: 0;
  margin-left: 4px;
  font-variant-numeric: tabular-nums;
}

.demoButton {
  width: 130px;
  height: 130px;
  border-radius: 50%;
  background: radial-gradient(
    circle at 35% 30%,
    var(--color-gold-100) 0%,
    var(--color-gold-500) 35%,
    var(--color-gold-700) 70%,
    var(--color-gold-900) 100%
  );
  box-shadow:
    0 0 0 3px var(--color-gold-500),
    0 0 24px rgba(201, 169, 107, 0.4),
    inset 0 -8px 14px rgba(0, 0, 0, 0.45),
    inset 0 4px 8px rgba(255, 240, 200, 0.35);
  display: grid;
  place-items: center;
  text-align: center;
  cursor: pointer;
  border: 3px solid rgba(0, 0, 0, 0.4);
  margin: 12px auto 0;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 13px;
  letter-spacing: 0.12em;
  color: var(--color-bg-warm);
  transition: transform 200ms var(--easing-spring);
}

.demoButton:active {
  transform: scale(0.96);
}

.demoButton:disabled {
  background: radial-gradient(circle at 35% 30%, #6a6a6a 0%, #3a3a3a 100%);
  box-shadow: 0 0 0 3px #4a4a4a;
  color: #4a4a4a;
  cursor: not-allowed;
}

.cta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 14px;
  opacity: 0;
  transform: translateY(12px);
  transition: opacity 500ms ease, transform 500ms var(--easing-spring);
}

.cta[data-visible="true"] {
  opacity: 1;
  transform: translateY(0);
}

.cta a {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 11px;
  letter-spacing: 0.15em;
  padding: 10px;
  text-align: center;
  text-decoration: none;
  display: block;
  border-radius: 4px;
}

.ctaPrimary {
  background: linear-gradient(180deg, var(--color-gold-500), var(--color-gold-700));
  color: var(--color-bg-warm);
  border: 1px solid var(--color-gold-500);
}

.ctaSecondary {
  background: transparent;
  color: var(--color-gold-500);
  border: 1px solid var(--color-gold-900);
}
```

- [ ] **Step 2: LandingScreen.tsx を改修**

```tsx
// web/components/auth/LandingScreen.tsx
//
// 未ログイン者向けランディング (spec §5.1)。中央のデモボタンは 1 回押下でき、
// 押すと体験用残高が ¥0 に減算され、CTA がスライドインする。
"use client";

import Link from "next/link";
import { useState } from "react";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";
import { COPY } from "@/lib/copy";

import styles from "./LandingScreen.module.css";

const DEMO_INITIAL_BALANCE = 1000;

export function LandingScreen() {
  const [demoUsed, setDemoUsed] = useState(false);
  const [demoBalance, setDemoBalance] = useState(DEMO_INITIAL_BALANCE);

  const handleDemo = () => {
    if (demoUsed) return;
    setDemoUsed(true);
    setDemoBalance(0);
  };

  return (
    <ScreenFrame testid="landing-screen">
      <BrandHeader />
      <div className={styles.hero}>
        <div className={styles.h1}>
          <div>{COPY.landing.heroLine1}</div>
          <div>{COPY.landing.heroLine2}</div>
        </div>
        <div className={styles.sub}>{COPY.landing.sub}</div>
        <div className={styles.demoLabel}>
          {COPY.landing.demoBalanceLabel}{" "}
          <b>¥{demoBalance.toLocaleString("ja-JP")}</b>
        </div>
        <button
          type="button"
          className={styles.demoButton}
          onClick={handleDemo}
          disabled={demoUsed}
          data-testid="landing-demo-button"
          aria-label={demoUsed ? "体験用デモ完了" : "体験デモを試す"}
        >
          {demoUsed ? COPY.landing.demoButtonAfter : COPY.landing.demoButtonMain}
        </button>
      </div>
      <div className={styles.cta} data-visible={demoUsed}>
        <Link
          href="/signup"
          role="button"
          data-testid="landing-start-button"
          className={styles.ctaPrimary}
        >
          {COPY.landing.ctaPrimary}
        </Link>
        <Link
          href="/login"
          data-testid="landing-login-link"
          className={styles.ctaSecondary}
        >
          {COPY.landing.ctaSecondary}
        </Link>
      </div>
    </ScreenFrame>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: 既存テスト確認**

Run: `cd web && npm run test`
Expected: 既存テストが新コピーと整合 (もし `LandingScreen` テストが直接文字列を見ていたら次タスクで修正)

- [ ] **Step 5: Commit**

```bash
git add web/components/auth/LandingScreen.tsx web/components/auth/LandingScreen.module.css
git commit -m "🎨 feat(web): ランディング画面をリヴァイ調＋デモボタンに刷新"
```

---

## Task 15: サインアップ画面

**Files:**
- Modify: `web/components/auth/SignupScreen.tsx`
- Create: `web/components/auth/SignupScreen.module.css`
- Modify: `web/tests/signupScreen.test.tsx` (旧文言の参照があれば修正)

- [ ] **Step 1: CSS を書く (LoginScreen と共有させたいが、画面ごとに微差あるので個別作成)**

```css
/* web/components/auth/SignupScreen.module.css */

.h1 {
  margin-top: 14px;
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 22px;
  color: var(--color-cream);
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.h1 div {
  display: block;
}

.sub {
  font-family: var(--font-body-mincho), serif;
  font-weight: 700;
  font-size: 10px;
  color: var(--color-gold-500);
  letter-spacing: 0.04em;
  margin-top: 4px;
}

.field {
  margin-top: 14px;
}

.field label {
  display: block;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 9px;
  color: var(--color-mute);
  letter-spacing: 0.2em;
  margin-bottom: 4px;
}

.field input {
  width: 100%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--color-gold-900);
  color: var(--color-cream);
  padding: 8px 10px;
  font-size: 14px;
  border-radius: 4px;
  font-family: var(--font-body-sans), system-ui, sans-serif;
}

.field input:focus {
  outline: 2px solid var(--color-gold-500);
  outline-offset: 0;
}

.hints {
  list-style: none;
  margin-top: 6px;
  padding: 0;
}

.hints li {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 9px;
  color: var(--color-mute);
  letter-spacing: 0.04em;
}

.hints li.ok {
  color: var(--color-gold-500);
}

.error {
  margin-top: 8px;
  color: var(--color-accent-danger);
  font-size: 13px;
}

.submit {
  margin-top: auto;
  background: linear-gradient(180deg, var(--color-gold-500), var(--color-gold-700));
  color: var(--color-bg-warm);
  border: 1px solid var(--color-gold-500);
  padding: 12px;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 12px;
  letter-spacing: 0.15em;
  cursor: pointer;
}

.submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

- [ ] **Step 2: SignupScreen.tsx を改修**

```tsx
// web/components/auth/SignupScreen.tsx
"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/hooks/useAuth";
import { authMessages, AuthErrorWithCode } from "@/lib/authMessages";
import { COPY } from "@/lib/copy";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";

import styles from "./SignupScreen.module.css";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

type PwdHints = {
  lengthOk: boolean;
  upperOk: boolean;
  lowerOk: boolean;
  digitOk: boolean;
};

function checkPasswordHints(pw: string): PwdHints {
  return {
    lengthOk: pw.length >= 8,
    upperOk: /[A-Z]/.test(pw),
    lowerOk: /[a-z]/.test(pw),
    digitOk: /\d/.test(pw),
  };
}

function isPasswordValid(h: PwdHints): boolean {
  return h.lengthOk && h.upperOk && h.lowerOk && h.digitOk;
}

export function SignupScreen() {
  const { signup } = useAuth();
  const router = useRouter();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const hints = checkPasswordHints(password);
  const isEmailValid = EMAIL_REGEX.test(email);
  const canSubmit = isEmailValid && isPasswordValid(hints) && !isSubmitting;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setIsSubmitting(true);
    setErrorMessage(null);
    try {
      await signup(email, password);
      router.push("/");
    } catch (err) {
      const code = err instanceof AuthErrorWithCode ? err.code : "UNKNOWN";
      setErrorMessage(authMessages[code]);
    } finally {
      setPassword("");
      setIsSubmitting(false);
    }
  }

  return (
    <ScreenFrame testid="signup-screen">
      <BrandHeader />
      <h1 className={styles.h1}>
        <div>{COPY.signup.h1Line1}</div>
        <div>{COPY.signup.h1Line2}</div>
      </h1>
      <div className={styles.sub}>{COPY.signup.sub}</div>
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
          <label htmlFor="signup-email">メールアドレス</label>
          <input
            id="signup-email"
            data-testid="signup-email-input"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div className={styles.field}>
          <label htmlFor="signup-password">パスワード</label>
          <input
            id="signup-password"
            data-testid="signup-password-input"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <ul className={styles.hints} data-testid="signup-password-hints" aria-live="polite">
            <li className={hints.lengthOk ? "ok" : ""}>
              {hints.lengthOk ? "✓" : "・"} 8 文字以上
            </li>
            <li className={hints.upperOk ? "ok" : ""}>
              {hints.upperOk ? "✓" : "・"} 英大文字を含む
            </li>
            <li className={hints.lowerOk ? "ok" : ""}>
              {hints.lowerOk ? "✓" : "・"} 英小文字を含む
            </li>
            <li className={hints.digitOk ? "ok" : ""}>
              {hints.digitOk ? "✓" : "・"} 数字を含む
            </li>
          </ul>
        </div>
        {errorMessage && (
          <p
            className={styles.error}
            data-testid="signup-error-message"
            role="alert"
            aria-live="polite"
          >
            {errorMessage}
          </p>
        )}
        <button
          type="submit"
          className={styles.submit}
          data-testid="signup-submit-button"
          disabled={!canSubmit}
        >
          {isSubmitting ? COPY.signup.submitting : COPY.signup.submit}
        </button>
      </form>
    </ScreenFrame>
  );
}
```

注意: hints の `.ok` クラス参照は CSS Module 経由ではなくグローバル文字列扱いになる。CSS Module 化したいなら `styles.ok` で書き換え。下記の修正を加える:

```tsx
// 上記 ul.hints 内を以下のように差し替える
<li className={hints.lengthOk ? styles.ok : ""}>
  {hints.lengthOk ? "✓" : "・"} 8 文字以上
</li>
```

- [ ] **Step 3: 既存テストを確認・修正**

Run: `cd web && cat tests/signupScreen.test.tsx | head -60` で既存内容を確認。

Run: `cd web && npm run test -- signupScreen`
Expected: 文言依存があれば FAIL → 失敗内容を見て、`30 秒でダメ化体験スタート` 等の旧文言が新コピー (例: `面倒は、こちらで引き受ける。` `30 秒で済む。`) に変わっていることを確認した上で、テスト側の文字列アサーションを更新。

- [ ] **Step 4: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 5: Commit**

```bash
git add web/components/auth/SignupScreen.tsx web/components/auth/SignupScreen.module.css web/tests/signupScreen.test.tsx
git commit -m "🎨 feat(web): サインアップ画面をリヴァイ調見出し＋金色フォームに刷新"
```

---

## Task 16: ログイン画面

**Files:**
- Modify: `web/components/auth/LoginScreen.tsx`
- Create: `web/components/auth/LoginScreen.module.css`

- [ ] **Step 1: CSS を書く (Signup と同じ規約)**

```css
/* web/components/auth/LoginScreen.module.css */

.h1 {
  margin-top: 14px;
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 28px;
  color: var(--color-cream);
  letter-spacing: -0.02em;
}

.sub {
  font-family: var(--font-body-mincho), serif;
  font-weight: 700;
  font-size: 11px;
  color: var(--color-gold-500);
  letter-spacing: 0.04em;
  margin-top: 4px;
}

.sessionHint {
  margin-top: 12px;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 11px;
  color: var(--color-paper);
  letter-spacing: 0.04em;
}

.field {
  margin-top: 14px;
}

.field label {
  display: block;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 9px;
  color: var(--color-mute);
  letter-spacing: 0.2em;
  margin-bottom: 4px;
}

.field input {
  width: 100%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--color-gold-900);
  color: var(--color-cream);
  padding: 8px 10px;
  font-size: 14px;
  border-radius: 4px;
  font-family: var(--font-body-sans), system-ui, sans-serif;
}

.field input:focus {
  outline: 2px solid var(--color-gold-500);
  outline-offset: 0;
}

.error {
  margin-top: 8px;
  color: var(--color-accent-danger);
  font-size: 13px;
}

.submit {
  margin-top: auto;
  background: linear-gradient(180deg, var(--color-gold-500), var(--color-gold-700));
  color: var(--color-bg-warm);
  border: 1px solid var(--color-gold-500);
  padding: 12px;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 12px;
  letter-spacing: 0.15em;
  cursor: pointer;
}

.submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

- [ ] **Step 2: LoginScreen.tsx を改修**

```tsx
// web/components/auth/LoginScreen.tsx
"use client";

import { useState, FormEvent } from "react";
import { useSearchParams, useRouter } from "next/navigation";

import { useAuth } from "@/hooks/useAuth";
import { authMessages, AuthErrorWithCode } from "@/lib/authMessages";
import { COPY } from "@/lib/copy";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";

import styles from "./LoginScreen.module.css";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function LoginScreen() {
  const { login } = useAuth();
  const router = useRouter();
  const params = useSearchParams();
  const fromSessionExpired = params.get("from") === "session_expired";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const isEmailValid = EMAIL_REGEX.test(email);
  const canSubmit = isEmailValid && password.length > 0 && !isSubmitting;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setIsSubmitting(true);
    setErrorMessage(null);
    try {
      await login(email, password);
      router.push("/");
    } catch (err) {
      const code = err instanceof AuthErrorWithCode ? err.code : "UNKNOWN";
      setErrorMessage(authMessages[code]);
    } finally {
      setPassword("");
      setIsSubmitting(false);
    }
  }

  return (
    <ScreenFrame testid="login-screen">
      <BrandHeader />
      <h1 className={styles.h1}>{COPY.login.h1}</h1>
      <div className={styles.sub}>{COPY.login.sub}</div>
      {fromSessionExpired && (
        <p
          className={styles.sessionHint}
          data-testid="login-session-expired-hint"
          role="status"
        >
          {authMessages.SESSION_EXPIRED}
        </p>
      )}
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
          <label htmlFor="login-email">メールアドレス</label>
          <input
            id="login-email"
            data-testid="login-email-input"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div className={styles.field}>
          <label htmlFor="login-password">パスワード</label>
          <input
            id="login-password"
            data-testid="login-password-input"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        {errorMessage && (
          <p
            className={styles.error}
            data-testid="login-error-message"
            role="alert"
            aria-live="polite"
          >
            {errorMessage}
          </p>
        )}
        <button
          type="submit"
          className={styles.submit}
          data-testid="login-submit-button"
          disabled={!canSubmit}
        >
          {isSubmitting ? COPY.login.submitting : COPY.login.submit}
        </button>
      </form>
    </ScreenFrame>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/auth/LoginScreen.tsx web/components/auth/LoginScreen.module.css
git commit -m "🎨 feat(web): ログイン画面をリヴァイ調見出しに刷新"
```

---

## Task 17: 注文完了画面

**Files:**
- Modify: `web/app/order/[id]/complete/page.tsx`
- Create: `web/app/order/[id]/complete/CompleteScreen.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/app/order/[id]/complete/CompleteScreen.module.css */

.complete {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 8px;
  position: relative;
}

.complete::before {
  content: "";
  position: absolute;
  inset: -40px;
  background: radial-gradient(
    ellipse at center,
    rgba(201, 169, 107, 0.32),
    transparent 60%
  );
  pointer-events: none;
  opacity: 0;
  animation: glow-in 800ms ease forwards;
}

@keyframes glow-in {
  to { opacity: 1; }
}

.verdict {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 38px;
  color: var(--color-cream);
  letter-spacing: -0.02em;
  position: relative;
  opacity: 0;
  animation: verdict-in 700ms var(--easing-spring) forwards;
}

@keyframes verdict-in {
  from {
    opacity: 0;
    letter-spacing: 0.1em;
  }
  to {
    opacity: 1;
    letter-spacing: -0.02em;
  }
}

.body,
.store,
.price,
.line,
.meta,
.next {
  position: relative;
  opacity: 0;
  transform: translateY(8px);
  animation: stagger-in 500ms ease forwards;
}

.body { animation-delay: 200ms; }
.store { animation-delay: 280ms; }
.price { animation-delay: 360ms; }
.line  { animation-delay: 440ms; }
.meta  { animation-delay: 520ms; }
.next  { animation-delay: 600ms; }

@keyframes stagger-in {
  to { opacity: 1; transform: translateY(0); }
}

.body {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 11px;
  color: var(--color-mute);
  letter-spacing: 0.15em;
  margin-top: 6px;
}

.store {
  font-family: var(--font-body-mincho), serif;
  font-weight: 700;
  font-size: 18px;
  color: var(--color-cream);
  margin-top: 6px;
}

.price {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 32px;
  color: var(--color-gold-500);
  font-variant-numeric: tabular-nums;
}

.line {
  width: 60%;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent,
    var(--color-gold-500),
    transparent
  );
  margin: 14px auto;
}

.meta {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 10px;
  color: var(--color-gold-500);
  letter-spacing: 0.15em;
}

.metaMute {
  color: var(--color-mute);
}

.next {
  margin-top: auto;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 11px;
  color: var(--color-mute);
  letter-spacing: 0.2em;
  padding: 12px;
  background: transparent;
  border: none;
  cursor: pointer;
}

.next:hover {
  color: var(--color-paper);
}
```

- [ ] **Step 2: 完了画面を改修**

```tsx
// web/app/order/[id]/complete/page.tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAtomValue } from "jotai";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";
import { COPY, composeMonthlyMeta } from "@/lib/copy";
import { balanceAtom, monthlyCountAtom } from "@/state/main";

import styles from "./CompleteScreen.module.css";

const AUTO_NAVIGATE_MS = 5000;

export default function OrderCompletePage() {
  const router = useRouter();
  const balance = useAtomValue(balanceAtom);
  const monthlyCount = useAtomValue(monthlyCountAtom);

  useEffect(() => {
    const timer = setTimeout(() => router.push("/"), AUTO_NAVIGATE_MS);
    return () => clearTimeout(timer);
  }, [router]);

  // 注文情報は jotai 経由 (GoroButton onSuccess で書き込み済み) ではなく、
  // mvp ではダミー固定の方が単純。ただし spec §5.4 で「店名・金額は API
  // 応答の値」を参照するため、別途 lastOrderAtom を導入する余地あり。
  // 現段階ではダミー値をフォールバックとして使う。
  const storeName = "CoCo壱番屋 新宿店";
  const amount = 1200;

  return (
    <ScreenFrame testid="order-completion-screen">
      <BrandHeader />
      <div className={styles.complete}>
        <div className={styles.verdict}>{COPY.complete.verdict}</div>
        <div className={styles.body}>{COPY.complete.body}</div>
        <div className={styles.store}>{storeName}</div>
        <div className={styles.price}>¥{amount.toLocaleString("ja-JP")}</div>
        <div className={styles.line} />
        <div className={styles.meta}>
          {composeMonthlyMeta(monthlyCount > 0 ? monthlyCount : 1)}
        </div>
        <div className={`${styles.meta} ${styles.metaMute}`}>
          残りダメ予算 ¥{balance >= 0 ? balance.toLocaleString("ja-JP") : "---"}
        </div>
        <button
          type="button"
          className={styles.next}
          onClick={() => router.push("/")}
        >
          {COPY.complete.next}
        </button>
      </div>
    </ScreenFrame>
  );
}
```

(注意: 動的注文情報の伝搬は将来課題。現実装は jotai に書き込まれた最新の `monthlyCount` `balance` だけ使い、店名・金額はダミーのまま。実 API を完了画面で再 GET したい場合は別 spec で対応。)

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/app/order/[id]/complete/page.tsx web/app/order/[id]/complete/CompleteScreen.module.css
git commit -m "🎨 feat(web): 注文完了画面を黄金グロー＋いい判断だ。に刷新"
```

---

## Task 18: ログアウトモーダル

**Files:**
- Modify: `web/components/auth/LogoutConfirmModal.tsx`
- Create: `web/components/auth/LogoutConfirmModal.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/auth/LogoutConfirmModal.module.css */

.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(4px);
  display: grid;
  place-items: center;
  z-index: 1100;
  animation: overlay-in 200ms ease;
}

@keyframes overlay-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.card {
  background:
    repeating-linear-gradient(
      0deg,
      rgba(255, 255, 255, 0.02) 0 1px,
      transparent 1px 4px
    ),
    radial-gradient(
      ellipse at top,
      var(--color-bg-warm) 0%,
      var(--color-bg-mid) 100%
    );
  border: 1px solid var(--color-gold-500);
  box-shadow: 0 0 30px rgba(201, 169, 107, 0.3);
  padding: 22px 18px;
  width: min(420px, 86%);
  text-align: center;
  border-radius: var(--radius-card);
  animation: card-in 300ms var(--easing-spring);
}

@keyframes card-in {
  from { opacity: 0; transform: scale(0.94); }
  to { opacity: 1; transform: scale(1); }
}

.h2 {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 24px;
  color: var(--color-cream);
  letter-spacing: -0.02em;
}

.sub {
  font-family: var(--font-body-mincho), serif;
  font-weight: 700;
  font-size: 11px;
  color: var(--color-gold-500);
  margin-top: 8px;
  letter-spacing: 0.04em;
}

.actions {
  display: flex;
  gap: 8px;
  margin-top: 14px;
  justify-content: center;
}

.actions button {
  flex: 1;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 12px;
  letter-spacing: 0.15em;
  padding: 10px;
  cursor: pointer;
  border-radius: 4px;
}

.primary {
  background: linear-gradient(180deg, var(--color-gold-500), var(--color-gold-700));
  color: var(--color-bg-warm);
  border: 1px solid var(--color-gold-500);
}

.secondary {
  background: transparent;
  color: var(--color-gold-500);
  border: 1px solid var(--color-gold-900);
}
```

- [ ] **Step 2: LogoutConfirmModal.tsx を改修**

```tsx
// web/components/auth/LogoutConfirmModal.tsx
"use client";

import { COPY } from "@/lib/copy";
import styles from "./LogoutConfirmModal.module.css";

interface LogoutConfirmModalProps {
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

export function LogoutConfirmModal({
  open,
  onCancel,
  onConfirm,
}: LogoutConfirmModalProps) {
  if (!open) return null;
  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="logout-confirm-title"
      data-testid="logout-confirm-modal"
      className={styles.overlay}
    >
      <div className={styles.card}>
        <h2 id="logout-confirm-title" className={styles.h2}>
          {COPY.logout.h2}
        </h2>
        <p className={styles.sub}>{COPY.logout.sub}</p>
        <div className={styles.actions}>
          <button
            type="button"
            onClick={onCancel}
            data-testid="logout-confirm-cancel"
            className={styles.secondary}
          >
            {COPY.logout.secondary}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            data-testid="logout-confirm-submit"
            className={styles.primary}
          >
            {COPY.logout.primary}
          </button>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/auth/LogoutConfirmModal.tsx web/components/auth/LogoutConfirmModal.module.css
git commit -m "🎨 feat(web): ログアウト確認モーダルを世界観統一"
```

---

## Task 19: セッション切れモーダル

**Files:**
- Modify: `web/components/auth/SessionExpiredModalHost.tsx`
- Create: `web/components/auth/SessionExpiredModalHost.module.css`

- [ ] **Step 1: CSS を書く (LogoutConfirmModal と同じ規約)**

```css
/* web/components/auth/SessionExpiredModalHost.module.css */

.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(4px);
  display: grid;
  place-items: center;
  z-index: 1200;
  animation: overlay-in 200ms ease;
}

@keyframes overlay-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.card {
  background:
    repeating-linear-gradient(0deg, rgba(255,255,255,.02) 0 1px, transparent 1px 4px),
    radial-gradient(ellipse at top, var(--color-bg-warm) 0%, var(--color-bg-mid) 100%);
  border: 1px solid var(--color-gold-500);
  box-shadow: 0 0 30px rgba(201, 169, 107, 0.3);
  padding: 22px 18px;
  width: min(420px, 86%);
  text-align: center;
  border-radius: var(--radius-card);
  animation: card-in 300ms var(--easing-spring);
}

@keyframes card-in {
  from { opacity: 0; transform: scale(0.94); }
  to { opacity: 1; transform: scale(1); }
}

.h2 {
  font-family: var(--font-display-serif), serif;
  font-style: italic;
  font-weight: 700;
  font-size: 24px;
  color: var(--color-cream);
  letter-spacing: -0.02em;
}

.sub {
  font-family: var(--font-mono-pixel), monospace;
  font-size: 11px;
  color: var(--color-paper);
  margin-top: 8px;
  letter-spacing: 0.04em;
}

.button {
  margin-top: 14px;
  background: linear-gradient(180deg, var(--color-gold-500), var(--color-gold-700));
  color: var(--color-bg-warm);
  border: 1px solid var(--color-gold-500);
  font-family: var(--font-mono-pixel), monospace;
  font-size: 12px;
  letter-spacing: 0.15em;
  padding: 10px 24px;
  cursor: pointer;
  border-radius: 4px;
}
```

- [ ] **Step 2: SessionExpiredModalHost.tsx を改修**

```tsx
// web/components/auth/SessionExpiredModalHost.tsx
"use client";

import { useCallback, useEffect, useRef } from "react";
import { useAtomValue, useSetAtom } from "jotai";
import { useRouter } from "next/navigation";

import { sessionExpiredAtom } from "@/state/auth";
import { authMessages } from "@/lib/authMessages";
import { COPY } from "@/lib/copy";

import styles from "./SessionExpiredModalHost.module.css";

const REDIRECT_DELAY_MS = 1500;

export function SessionExpiredModalHost() {
  const state = useAtomValue(sessionExpiredAtom);
  const setState = useSetAtom(sessionExpiredAtom);
  const router = useRouter();
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const navigateToLogin = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    setState(null);
    router.push("/login?from=session_expired");
  }, [router, setState]);

  useEffect(() => {
    if (state === null) return;
    timerRef.current = setTimeout(navigateToLogin, REDIRECT_DELAY_MS);
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [state, navigateToLogin]);

  if (state === null) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="session-expired-title"
      data-testid="session-expired-modal"
      className={styles.overlay}
    >
      <div className={styles.card}>
        <h2 id="session-expired-title" className={styles.h2}>
          {COPY.sessionExpired.h2}
        </h2>
        <p className={styles.sub}>{authMessages.SESSION_EXPIRED}</p>
        <button
          type="button"
          data-testid="session-expired-ok"
          onClick={navigateToLogin}
          className={styles.button}
        >
          {COPY.sessionExpired.button}
        </button>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: 既存テスト実行**

Run: `cd web && npm run test -- sessionExpired`
Expected: 文言依存があれば修正。`離れすぎたな。` が表示されることを確認するアサーションが必要なら追加。

- [ ] **Step 5: Commit**

```bash
git add web/components/auth/SessionExpiredModalHost.tsx web/components/auth/SessionExpiredModalHost.module.css web/tests/sessionExpired.test.tsx
git commit -m "🎨 feat(web): セッション切れモーダルを世界観統一"
```

---

## Task 20: Toast スタイル更新

**Files:**
- Modify: `web/components/order/Toast.tsx`
- Create: `web/components/order/Toast.module.css`

- [ ] **Step 1: CSS を書く**

```css
/* web/components/order/Toast.module.css */

.toast {
  background: rgba(26, 18, 8, 0.92);
  color: var(--color-cream);
  padding: 10px 16px;
  border-radius: 6px;
  margin-top: 8px;
  border: 1px solid rgba(201, 169, 107, 0.3);
  box-shadow: 0 4px 14px rgba(201, 169, 107, 0.15);
  max-width: 360px;
  font-family: var(--font-mono-pixel), monospace;
  font-size: 13px;
  letter-spacing: 0.04em;
  line-height: 1.4;
}
```

- [ ] **Step 2: Toast.tsx を改修**

```tsx
// web/components/order/Toast.tsx
"use client";

import styles from "./Toast.module.css";

type ToastProps = {
  id: string;
  text: string;
  durationMs: number;
};

export function Toast({ text }: ToastProps) {
  return (
    <div role="status" className={styles.toast}>
      {text}
    </div>
  );
}
```

- [ ] **Step 3: lint / type-check**

Run: `cd web && npm run lint && npx tsc --noEmit`
Expected: エラー無し

- [ ] **Step 4: Commit**

```bash
git add web/components/order/Toast.tsx web/components/order/Toast.module.css
git commit -m "🎨 feat(web): Toast を世界観カラーに更新"
```

---

## Task 21: 既存テストの文言修正 + Playwright スモーク

**Files:**
- Modify: 既存テストファイル (文言不一致のもの)
- Modify: `web/tests/*.test.tsx` のうち UI コピーを assert しているもの

- [ ] **Step 1: 全テストを走らせて失敗箇所を特定**

Run: `cd web && npm run test`
Expected: いくつか文言不一致で FAIL する可能性 (e.g. SignupScreen の「30 秒でダメ化体験スタート」が `30 秒で済む。` に変わったため)

- [ ] **Step 2: 失敗テストを 1 つずつ修正**

各失敗について、テストの assertion を新コピー (COPY 経由で取得した値) と比較するように更新。
- 文字列リテラルを assertion に書く代わりに `COPY.signup.sub` のように import する形が望ましい。

例: `web/tests/signupScreen.test.tsx` で `getByText("30 秒でダメ化体験スタート")` を見つけたら、

```ts
import { COPY } from "@/lib/copy";
// ...
screen.getByText(COPY.signup.sub);
```

に置き換える。

- [ ] **Step 3: 再度全テスト**

Run: `cd web && npm run test`
Expected: PASS

- [ ] **Step 4: Playwright で全画面の表示を確認**

Run: `cd web && npm run e2e`
Expected: 既存 e2e (もしあれば) PASS。失敗があれば文言更新が必要。

- [ ] **Step 5: Commit**

```bash
git add web/tests
git commit -m "🎨 test(web): 新コピーに合わせて文言依存テストを更新"
```

---

## Task 22: 仕上げ (lint / type-check / 全画面手動確認)

**Files:** なし (確認のみ)

- [ ] **Step 1: 全体の lint / type-check / test**

Run: `cd web && npm run lint && npx tsc --noEmit && npm run test`
Expected: 全部 PASS

- [ ] **Step 2: 開発サーバで全画面を手動確認**

Run: `cd web && npm run dev`

確認項目:
- [ ] `/` (未ログイン) → ランディング: 暗茶背景 + 「考えるな。押せ。」+ 金色デモボタン + CTA
- [ ] デモボタンを押す → ¥0 に減算 + CTA がスライドイン
- [ ] `/signup` → 「面倒は、こちらで引き受ける。」見出し + フォーム
- [ ] `/login` → 「戻ってきたか。」見出し + フォーム
- [ ] ログイン後 `/` → メイン画面 EMPTY 状態 (履歴なし)
- [ ] ボタン押下 → SLOT 演出 → 完了画面 → メイン IDLE 復帰
- [ ] 残高 0 にしてから注文 → DEAD 状態の脱色 + 増額赤ボタン
- [ ] ログアウトボタン → 「やめるのか？」モーダル
- [ ] (E2E でセッション切れを再現できるなら) セッション切れモーダル表示

- [ ] **Step 3: 不要ファイルの削除確認**

Run: `cd web && grep -r "OrderHistoryList" app components` で参照箇所を確認。
Expected: メイン画面 (`app/page.tsx`) からの参照は無い (Task 13 で外している)。`OrderHistoryList.tsx` 自体は残してよい (将来履歴画面で使うため、spec §4.4 で温存と明記)。

- [ ] **Step 4: 最終 Commit (任意)**

何も変更がなければ commit 不要。あれば:

```bash
git add web/
git commit -m "🎨 chore(web): 仕上げ修正"
```

---

## Self-Review Checklist (実装者向け)

実装が一通り終わったら、以下を確認:

1. **Spec coverage**: spec の §1.6 コピー一覧の全項目が `lib/copy.ts` に入っているか / §3 各状態が GoroButton で表現されているか / §5 各画面の H1 / サブが該当画面に反映されているか
2. **Cross-browser**: Chrome / Safari / Firefox で `backdrop-filter` `aspect-ratio` `radial-gradient` が動くか
3. **Mobile**: 320px 幅 / 480px 幅 / iPad 幅で崩れないか
4. **Reduced motion**: macOS のシステム環境設定でモーション軽減を ON にして、SLOT・サジェスト明滅・DEAD fade が抑止されるか
5. **a11y**: ボタンの `aria-label`、各画面の `data-testid` が spec 通りか

---

## 開発時 tips

- `web/app/globals.css` を変更したら HMR が走らないことがある → dev サーバを再起動
- `next/font/google` は初回ビルド時にダウンロードされる → 初回 `npm run dev` は時間がかかる
- CSS Module の class が見つからないエラーは `*.module.css` の拡張子と import path を再確認
- `data-testid` を変えた場合、既存 e2e の selector も更新する
