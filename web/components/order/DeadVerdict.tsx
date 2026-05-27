"use client";

import { COPY } from "@/lib/copy";

import styles from "./DeadVerdict.module.css";

export function DeadVerdict() {
  return (
    <div
      className={styles.verdict}
      role="status"
      aria-live="polite"
      data-testid="dead-verdict"
    >
      {COPY.main.deadVerdict}
    </div>
  );
}
