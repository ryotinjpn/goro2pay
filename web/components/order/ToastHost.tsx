// LC-30 ToastHost / P-FE-TOAST-02
//
// アプリ全画面で利用可能なトースト一覧コンテナ。app/layout.tsx の <body> 直下、
// <AppProviders> 配下に配置する (Jotai store 必須)。
//
// キュー制御:
//   - toastsAtom には追加された順に蓄積される
//   - ToastHost は slice(0, 3) で先頭 3 件のみ描画 (P-FE-TOAST-02 最大 3 件)
//   - 4 件目以降は配列にとどまり、useToast の setTimeout が先頭を消すと自動表示
"use client";

import { useToast } from "@/hooks/useToast";

import { Toast } from "./Toast";

const MAX_VISIBLE = 3;

export function ToastHost() {
  const { toasts } = useToast();
  const visible = toasts.slice(0, MAX_VISIBLE);

  return (
    <div
      role="region"
      aria-live="polite"
      aria-label="通知"
      data-testid="toast-host"
      style={{
        position: "fixed",
        bottom: 24,
        left: "50%",
        transform: "translateX(-50%)",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        zIndex: 1000,
        pointerEvents: "none",
      }}
    >
      {visible.map((t) => (
        <Toast key={t.id} {...t} />
      ))}
    </div>
  );
}
