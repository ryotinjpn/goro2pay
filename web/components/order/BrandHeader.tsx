// BrandHeader: mock/src/components/BrandHeader.tsx ベースに移植 (PR ⑤)。
// 左に「ゴロゴロPay」、右に時刻 + ログアウトボタン (web 独自) を並べる。
"use client";

import { useEffect, useState } from "react";

import { LogoutButton } from "@/components/auth/LogoutButton";
import { COPY } from "@/lib/copy";

import styles from "./BrandHeader.module.css";

function formatTime(d: Date): string {
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${hh}:${mm}`;
}

export function BrandHeader() {
  const [time, setTime] = useState<string>(() => formatTime(new Date()));

  useEffect(() => {
    // mock 同様 30 秒間隔。1 分単位の表示なので 60 秒にしても良いが、
    // ユーザが画面を開いた直後に分が変わったときに最大 60 秒のラグが出るのを防ぐ。
    const id = setInterval(() => setTime(formatTime(new Date())), 30_000);
    return () => clearInterval(id);
  }, []);

  const name = COPY.brand.name;
  const splitAt = name.indexOf("Pay");
  const left = splitAt >= 0 ? name.slice(0, splitAt) : name;
  const right = splitAt >= 0 ? name.slice(splitAt) : "";

  return (
    <header className={styles.header}>
      <span className={styles.brand}>
        <span className={styles.brandKana}>{left}</span>
        {right && <span className={styles.brandLatin}>{right}</span>}
      </span>
      <div className={styles.right}>
        <span className={styles.time}>{time}</span>
        <span className={styles.logoutSlot}>
          <LogoutButton />
        </span>
      </div>
    </header>
  );
}
