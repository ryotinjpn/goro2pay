// App Router のグローバル loading boundary。
// Server Component の suspense 中に表示される最小 placeholder。
// 個別ルートで上書きしたい場合は app/<route>/loading.tsx で override する。

import styles from "./loading.module.css";

export default function Loading() {
  return (
    <main role="status" aria-live="polite" className={styles.main}>
      <p className={styles.text}>読み込み中...</p>
    </main>
  );
}
