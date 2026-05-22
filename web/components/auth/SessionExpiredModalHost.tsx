"use client";

import { useEffect, useRef } from "react";
import { useAtomValue, useSetAtom } from "jotai";
import { useRouter } from "next/navigation";

import { sessionExpiredAtom } from "@/state/auth";
import { authMessages } from "@/lib/authMessages";
import { _resetSessionExpiredHandling } from "@/lib/apiClient";

const REDIRECT_DELAY_MS = 1500;

export function SessionExpiredModalHost() {
  const state = useAtomValue(sessionExpiredAtom);
  const setState = useSetAtom(sessionExpiredAtom);
  const router = useRouter();
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  function navigateToLogin() {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    setState(null);
    _resetSessionExpiredHandling();
    router.push("/login?from=session_expired");
  }

  useEffect(() => {
    if (state === null) return;
    timerRef.current = setTimeout(navigateToLogin, REDIRECT_DELAY_MS);
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state]);

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
