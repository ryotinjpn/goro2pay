"use client";

import { useCallback, useEffect, useRef } from "react";
import { useAtomValue, useSetAtom } from "jotai";
import { useRouter } from "next/navigation";

import { ModalPortal } from "@/components/common/ModalPortal";
import { sessionExpiredAtom } from "@/state/auth";
import { authMessages } from "@/lib/authMessages";
import { COPY } from "@/lib/copy";

import styles from "./SessionExpiredModalHost.module.css";

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
    <ModalPortal>
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="session-expired-title"
        data-testid="session-expired-modal"
        className={styles.overlay}
      >
        <div className={styles.card}>
          <h2 id="session-expired-title" className={styles.h2}>
            {COPY.sessionExpired.h2}
          </h2>
          <p className={styles.sub}>{authMessages.SESSION_EXPIRED}</p>
          <button
            type="button"
            data-testid="session-expired-ok"
            onClick={navigateToLogin}
            className={styles.button}
          >
            {COPY.sessionExpired.button}
          </button>
        </div>
      </div>
    </ModalPortal>
  );
}
