"use client";

import { COPY } from "@/lib/copy";
import styles from "./SuggestBubble.module.css";

export function SuggestBubble() {
  return (
    <div className={styles.bubble} data-testid="suggest-bubble">
      {COPY.main.suggestBubble}
    </div>
  );
}
