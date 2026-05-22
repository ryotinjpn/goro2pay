"use client";

import Link from "next/link";

export function LandingScreen() {
  return (
    <main data-testid="landing-screen" style={{ padding: 24, textAlign: "center" }}>
      <h1>ゴロゴロPay</h1>
      <p>めんどくさいを丸投げ</p>
      <div style={{ marginTop: 24, display: "flex", flexDirection: "column", gap: 12 }}>
        <Link href="/signup" data-testid="landing-signup-link">
          <button type="button" data-testid="landing-start-button">
            はじめる
          </button>
        </Link>
        <Link href="/login" data-testid="landing-login-link">
          ログイン
        </Link>
      </div>
    </main>
  );
}
