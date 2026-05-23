"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/hooks/useAuth";
import { LogoutConfirmModal } from "./LogoutConfirmModal";

export function LogoutButton() {
  const { logout } = useAuth();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);

  async function handleConfirm() {
    setIsProcessing(true);
    try {
      await logout();
    } finally {
      setIsProcessing(false);
      setOpen(false);
      // logout 失敗時 (API も signOut も失敗) でも / に push する。
      // 通常経路では Amplify Hub の `signedOut` を経由した useAuth status 更新で
      // AuthGuard が /login へ replace するため `/` に着地しないが、
      // signOut の Hub 反映が遅延 / 失敗したケースの保険として明示的に push する。
      // ユーザの体験を「ログアウトしたのに留まる」状態にしない目的。
      router.push("/");
    }
  }

  return (
    <>
      <button
        type="button"
        data-testid="logout-button"
        onClick={() => setOpen(true)}
        aria-label="ログアウト"
      >
        ⏏︎ ログアウト
      </button>
      <LogoutConfirmModal
        open={open}
        onCancel={() => !isProcessing && setOpen(false)}
        onConfirm={handleConfirm}
      />
    </>
  );
}
