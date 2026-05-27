// web/state/main.ts
//
// メイン画面の有限状態 (spec §2.3) と残高を保持する jotai atom 群。

import { atom } from "jotai";

export type ScreenState = "idle" | "suggested" | "slot" | "dead";

export const screenStateAtom = atom<ScreenState>("idle");

export const balanceAtom = atom<number>(-1);

export const monthlyCountAtom = atom<number>(0);

export const monthlyBudgetAtom = atom<number>(30000);

export const consumeRateAtom = atom((get) => {
  const balance = get(balanceAtom);
  const budget = get(monthlyBudgetAtom);
  if (balance < 0 || budget <= 0) return 0;
  const used = Math.max(0, budget - balance);
  return Math.min(1, used / budget);
});

export type Suggestion = {
  storeName: string;
  amount: number;
  suggestionId: string;
};

export const suggestionAtom = atom<Suggestion | null>(null);

// 直近の注文結果。Complete 画面で実データ表示するため onSuccess 時に setter で書き込む。
// PlaceOrderResponse (lib/api/orders.ts) が全 field 非 optional で返すため、producer 側で
// `useOrder.onSuccess` から欠損なくセットできる。
//   - orderId    : 将来の冪等性チェック / Complete URL 検証用に保持
//   - storeName  : Complete 画面の店名表示
//   - menuName   : Complete 画面のサブテキスト or OrderHistory との整合用 (PR ⑧で wiring)
//   - amount     : Complete 画面の金額表示
export type LastOrder = {
  orderId: string;
  storeName: string;
  menuName: string;
  amount: number;
};

export const lastOrderAtom = atom<LastOrder | null>(null);

// 注文成功時の金色フラッシュ + verdict ポップ演出を再走させるためのキー。
//
// GoroButton.handleClick の責務:
//   1. クリック時に setWinFlash(0) でリセット → MainScreen の showWinEffects=false
//      に倒し、API 応答前に前回の演出が "double-fire" するのを防ぐ
//   2. onSuccess で setWinFlash(Date.now()) を立てる → クリックごとにユニーク値で
//      `key={counter}` が更新され演出が clean に remount される (c=>c+1 にすると
//      リセット 0 → 1 で同じキーが繰り返され React が remount しない bug)
//
// MainScreen consumer の責務:
//   - `counter > 0` ガードで初期 mount 時のスタブ演出を防ぐ
//   - counter の値そのものは意味を持たない (タイムスタンプ or 0)
export const winFlashCounterAtom = atom<number>(0);
