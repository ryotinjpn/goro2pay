// LC-29 useToast hook
//
// トースト表示 API。toastsAtom (Jotai) に追加 → setTimeout で自動消去する。
// ToastHost (LC-30) が toastsAtom を購読し、最大 3 件を描画する。
"use client";

import { useAtom } from "jotai";
import { useCallback } from "react";

import { toastsAtom } from "@/state/toastAtoms";

/**
 * useToast() returns { toasts, showToast }.
 *
 * showToast(text, durationMs) はトースト 1 件を atom に追加し、
 * durationMs 経過後に自動で削除する setTimeout を仕掛ける。
 */
export function useToast(): {
  toasts: ReturnType<typeof useAtom<typeof toastsAtom>>[0];
  showToast: (text: string, durationMs?: number) => void;
} {
  const [toasts, setToasts] = useAtom(toastsAtom);

  const showToast = useCallback(
    (text: string, durationMs = 5000) => {
      const id =
        typeof crypto !== "undefined" && "randomUUID" in crypto
          ? crypto.randomUUID()
          : `${Date.now()}-${Math.random().toString(36).slice(2)}`;
      setToasts((prev) => [...prev, { id, text, durationMs }]);
      // unmount 後でも setTimeout は ToastHost に依存しないので、
      // atom から該当 id を消すだけで安全。
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
      }, durationMs);
    },
    [setToasts],
  );

  return { toasts, showToast };
}
