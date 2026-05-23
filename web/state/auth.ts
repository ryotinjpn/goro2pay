// Jotai atoms for Auth Unit
// NFR Design P-RES-02 / LC-AUTH-10: 二重発火防止のための atomic state

import { atom } from "jotai";

export type SessionExpiredState = {
  openedAt: number;
} | null;

// session 失効モーダルの表示状態。null なら未発火。
export const sessionExpiredAtom = atom<SessionExpiredState>(null);
