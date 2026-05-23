"use client";

import { useCallback, useEffect, useRef } from "react";
import { useAtomValue, useSetAtom } from "jotai";
import { useRouter } from "next/navigation";

import { sessionExpiredAtom } from "@/state/auth";
import { authMessages } from "@/lib/authMessages";

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
    // atom を null に戻すと triggerSessionExpired の compare-and-set ガードも
    // 自動で解除される (もうモジュールフラグは存在せず atom 単一が source of truth)。
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
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.4)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
      }}
    >
      <div style={{ background: "white", padding: 24, borderRadius: 8 }}>
        <h2 id="session-expired-title">お疲れ様でした</h2>
        <p>{authMessages.SESSION_EXPIRED}</p>
        <button type="button" data-testid="session-expired-ok" onClick={navigateToLogin}>
          OK
        </button>
      </div>
    </div>
  );
}
