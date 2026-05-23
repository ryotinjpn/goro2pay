"use client";

// App Router のグローバル error boundary。
// Server Component / Client Component で uncaught error が出た場合に表示される。
// 「ダメ化トーン軽」で authMessages.UNKNOWN を表示し、再試行ボタンを置く。

import { useEffect } from "react";

import { authMessages } from "@/lib/authMessages";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // 開発時の追跡用。本番では observability 側 (Amplify Hosting / CloudWatch
    // Logs) に出力される。PII を含めないよう error.message 全文ではなく
    // digest と name のみログ出力。
    // eslint-disable-next-line no-console
    console.error("[GlobalError]", error.name, error.digest);
  }, [error]);

  return (
    <main
      role="alert"
      aria-live="assertive"
      style={{ padding: 24, textAlign: "center" }}
    >
      <h1>申し訳ありません</h1>
      <p>{authMessages.UNKNOWN}</p>
      <button type="button" onClick={() => reset()} style={{ marginTop: 16 }}>
        もう一度
      </button>
    </main>
  );
}
