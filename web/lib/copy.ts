// web/lib/copy.ts
//
// リヴァイ調コピー集約 (spec §1.6)。
// UI レイヤーから直接文字列リテラルを書かず、必ずこのモジュール経由で参照する。

export const COPY = {
  brand: {
    name: "ゴロゴロPay",
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
