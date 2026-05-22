// App Router のグローバル loading boundary。
// Server Component の suspense 中に表示される最小 placeholder。
// 個別ルートで上書きしたい場合は app/<route>/loading.tsx で override する。

export default function Loading() {
  return (
    <main
      role="status"
      aria-live="polite"
      style={{ padding: 24, textAlign: "center" }}
    >
      <p>読み込み中...</p>
    </main>
  );
}
