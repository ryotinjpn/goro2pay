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
