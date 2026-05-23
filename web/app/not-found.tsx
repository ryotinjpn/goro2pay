// App Router のグローバル 404 boundary。
// 存在しないルートに到達したときに表示される。

import Link from "next/link";

export default function NotFound() {
  return (
    <main style={{ padding: 24, textAlign: "center" }}>
      <h1>ページが見つかりません</h1>
      <p>お探しのページは見つけられませんでした。</p>
      <Link href="/" style={{ display: "inline-block", marginTop: 16 }}>
        トップに戻る
      </Link>
    </main>
  );
}
