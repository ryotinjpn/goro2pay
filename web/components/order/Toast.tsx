// LC-31 Toast コンポーネント
//
// 個別トーストの表示。role="status" でスクリーンリーダーにステータス情報として通知。
"use client";

import styles from "./Toast.module.css";

type ToastProps = {
  id: string;
  text: string;
  durationMs: number;
};

export function Toast({ text }: ToastProps) {
  return (
    <div role="status" className={styles.toast}>
      {text}
    </div>
  );
}
