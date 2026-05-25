// Unit B Frontend 状態管理: insufficientBalanceAtom (P-DEG-02)
//
// Unit C の useOrder hook が 402 INSUFFICIENT_BALANCE を受信したときに
// `true` をセットし、`InsufficientBalanceModal` が atom を購読して open する。
// modal が閉じられたら `false` にリセットする。
//
// クロスユニット契約:
//   - Unit C `useOrder` (web/hooks/useOrder.ts): onError で 402 受信時に atom.set(true)
//   - Unit B `InsufficientBalanceModal` (web/components/budget/InsufficientBalanceModal.tsx): subscribe して open

import { atom } from "jotai";

export const insufficientBalanceAtom = atom<boolean>(false);
