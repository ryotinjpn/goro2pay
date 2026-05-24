// LC-25 useDisableLock / P-FE-LOCK-01
//
// ボタン連打抑制 1 秒の振る舞い hook。
// useState + useEffect cleanup で setTimeout を確実に解放し、unmount 時の
// メモリリークを防ぐ (NFRC-C22 / P-FE-LOCK-01)。
"use client";

import { useCallback, useEffect, useState } from "react";

/**
 * useDisableLock(durationMs) は { isLocked, triggerLock } を返す。
 *
 * triggerLock を呼ぶと durationMs ms だけ isLocked === true となり、
 * 期限到達で自動的に false に戻る。useEffect cleanup で clearTimeout する
 * ので、unmount 時もタイマーが残らない。
 */
export function useDisableLock(durationMs: number): {
  isLocked: boolean;
  triggerLock: () => void;
} {
  const [lockedUntil, setLockedUntil] = useState<number>(0);
  const isLocked = Date.now() < lockedUntil;

  const triggerLock = useCallback(() => {
    setLockedUntil(Date.now() + durationMs);
  }, [durationMs]);

  useEffect(() => {
    if (!isLocked) return;
    const remaining = lockedUntil - Date.now();
    const timer = setTimeout(() => setLockedUntil(0), remaining);
    return () => clearTimeout(timer);
  }, [lockedUntil, isLocked]);

  return { isLocked, triggerLock };
}
