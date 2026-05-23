// LC-32 OrderCompletionScreen (NFRC-C21)
//
// 注文完了画面。5 秒後に MainScreen へ自動遷移する (US-1-07 / NFR-DEG-01)。
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

const AUTO_NAVIGATE_MS = 5000;

export default function OrderCompletePage() {
  const router = useRouter();

  useEffect(() => {
    const timer = setTimeout(() => {
      router.push("/");
    }, AUTO_NAVIGATE_MS);
    return () => clearTimeout(timer);
  }, [router]);

  return (
    <main
      data-testid="order-completion-screen"
      style={{ padding: 24, textAlign: "center" }}
    >
      <h1 style={{ fontSize: 32 }}>注文完了</h1>
      <p style={{ color: "#666", marginTop: 16 }}>
        ダメ化を続けるため、5 秒後にメイン画面に戻ります…
      </p>
      <button
        type="button"
        onClick={() => router.push("/")}
        style={{
          marginTop: 24,
          padding: "8px 16px",
          background: "transparent",
          border: "1px solid #888",
          borderRadius: 6,
          cursor: "pointer",
        }}
      >
        今すぐ戻る
      </button>
    </main>
  );
}
