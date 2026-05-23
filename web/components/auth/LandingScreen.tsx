"use client";

import Link from "next/link";

export function LandingScreen() {
  // Link の中に button をネストすると HTML5 spec 違反 (interactive content の
  // ネスト) になり、a11y も壊れるため Link 単独で使う。「はじめる」も同様に
  // Link を button 風に CSS で扱う。
  return (
    <main data-testid="landing-screen" style={{ padding: 24, textAlign: "center" }}>
      <h1>ゴロゴロPay</h1>
      <p>めんどくさいを丸投げ</p>
      <div style={{ marginTop: 24, display: "flex", flexDirection: "column", gap: 12 }}>
        <Link
          href="/signup"
          role="button"
          data-testid="landing-start-button"
          style={{
            display: "inline-block",
            padding: "8px 16px",
            border: "1px solid currentColor",
            borderRadius: 4,
            textDecoration: "none",
          }}
        >
          はじめる
        </Link>
        <Link href="/login" data-testid="landing-login-link">
          ログイン
        </Link>
      </div>
    </main>
  );
}
