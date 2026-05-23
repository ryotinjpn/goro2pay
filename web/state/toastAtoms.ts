// LC-28 ToastsAtom
//
// トースト一覧のグローバル状態 (Jotai atom)。Unit A の jotai (`getDefaultStore`)
// と統一し、グローバルストアからアクセスする (Provider 不要)。
import { atom } from "jotai";

/**
 * ToastItem は表示中のトースト 1 件分。
 *
 * id は uuid (crypto.randomUUID) で一意化、durationMs 経過後に自動消去する。
 */
export type ToastItem = {
  id: string;
  text: string;
  durationMs: number;
};

/**
 * toastsAtom は表示中・キュー中の全トースト配列を保持する。
 *
 * ToastHost は slice(0, 3) で先頭 3 件のみ描画 (P-FE-TOAST-02、最大 3 件キュー)。
 */
export const toastsAtom = atom<ToastItem[]>([]);
