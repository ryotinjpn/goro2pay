// App Router のグローバル 404 boundary。
// 存在しないルートに到達したときに表示される。

import Link from "next/link";

import styles from "./not-found.module.css";

export default function NotFound() {
  return (
    <main className={styles.main}>
      <h1 className={styles.h1}>ページが見つかりません</h1>
      <p className={styles.body}>お探しのページは見つけられませんでした。</p>
      <Link href="/" className={styles.link}>
        トップに戻る
      </Link>
    </main>
  );
}
