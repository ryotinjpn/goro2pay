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
