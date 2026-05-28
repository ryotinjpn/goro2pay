// LC-BUDGET-10 InsufficientBalanceModal (P-DEG-02 残高枯渇演出)。
//
// Unit C の useOrder が 402 を受信したときに insufficientBalanceAtom に true をセットし、
// 本コンポーネントが atom を購読して open する。閉じる / 増額誘導いずれでも atom を false に戻す。
"use client";

import { useAtom } from "jotai";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { ModalPortal } from "@/components/common/ModalPortal";
import { insufficientBalanceAtom } from "@/state/budget";

import styles from "./InsufficientBalanceModal.module.css";

export function InsufficientBalanceModal() {
  const [open, setOpen] = useAtom(insufficientBalanceAtom);
  const router = useRouter();

  useEffect(() => {
    if (!open) return;
    const onKeydown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", onKeydown);
    return () => document.removeEventListener("keydown", onKeydown);
  }, [open, setOpen]);

  if (!open) return null;

  const handleRaise = () => {
    setOpen(false);
    router.push("/budget");
  };

  return (
    <ModalPortal>
      <div
        data-testid="insufficient-balance-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="insufficient-balance-title"
        onClick={(e) => {
          if (e.target === e.currentTarget) setOpen(false);
        }}
        className={styles.overlay}
      >
        <div className={styles.card}>
          <h2 id="insufficient-balance-title" className={styles.h2}>
            残りダメ予算が足りません…
          </h2>
          <p className={styles.sub}>
            今月のダメ予算を使い切りました。
            <br />
            もっとダメになる準備はできていますか？
          </p>
          <div className={styles.actions}>
            <button
              data-testid="close-modal-button"
              onClick={() => setOpen(false)}
              className={styles.secondary}
            >
              閉じる
            </button>
            <button
              data-testid="raise-budget-button"
              onClick={handleRaise}
              className={styles.primary}
            >
              もっとダメになる
            </button>
          </div>
        </div>
      </div>
    </ModalPortal>
  );
}
