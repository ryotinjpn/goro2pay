// LC-BUDGET-10 InsufficientBalanceModal (P-DEG-02 残高枯渇演出)。
//
// Unit C の useOrder が 402 を受信したときに insufficientBalanceAtom に true をセットし、
// 本コンポーネントが atom を購読して open する。閉じる / 増額誘導いずれでも atom を false に戻す。
"use client";

import { useAtom } from "jotai";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { insufficientBalanceAtom } from "@/state/budget";

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
    <div
      data-testid="insufficient-balance-modal"
      role="dialog"
      aria-modal="true"
      aria-labelledby="insufficient-balance-title"
      onClick={(e) => {
        if (e.target === e.currentTarget) setOpen(false);
      }}
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(17, 24, 39, 0.6)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
      }}
    >
      <div
        style={{
          background: "#ffffff",
          borderRadius: 16,
          padding: 24,
          maxWidth: 360,
          width: "calc(100% - 32px)",
          boxShadow: "0 12px 32px rgba(0,0,0,0.18)",
        }}
      >
        <h2
          id="insufficient-balance-title"
          style={{ fontSize: 20, fontWeight: 700, marginBottom: 12 }}
        >
          残りダメ予算が足りません…
        </h2>
        <p style={{ color: "#374151", fontSize: 14, marginBottom: 20 }}>
          今月のダメ予算を使い切りました。
          <br />
          もっとダメになる準備はできていますか？
        </p>
        <div style={{ display: "flex", gap: 8 }}>
          <button
            data-testid="raise-budget-button"
            onClick={handleRaise}
            style={{
              flex: 1,
              padding: "12px 16px",
              borderRadius: 8,
              background: "#dc2626",
              color: "#ffffff",
              border: "none",
              fontSize: 15,
              cursor: "pointer",
            }}
          >
            もっとダメになる
          </button>
          <button
            data-testid="close-modal-button"
            onClick={() => setOpen(false)}
            style={{
              flex: 1,
              padding: "12px 16px",
              borderRadius: 8,
              background: "#e5e7eb",
              color: "#374151",
              border: "none",
              fontSize: 15,
              cursor: "pointer",
            }}
          >
            閉じる
          </button>
        </div>
      </div>
    </div>
  );
}
