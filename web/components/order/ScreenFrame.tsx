"use client";

import { ReactNode } from "react";
import type { ScreenState } from "@/state/main";

import styles from "./ScreenFrame.module.css";

type Props = {
  children: ReactNode;
  screenState?: ScreenState;
  testid?: string;
  /** Light Aurora テーマ (移行済み画面のみ true)。未指定は従来のダーク。 */
  light?: boolean;
};

export function ScreenFrame({ children, screenState, testid, light }: Props) {
  return (
    <div className={styles.outer} data-theme={light ? "light" : undefined}>
      <main
        className={styles.frame}
        data-screen-state={screenState}
        data-testid={testid}
      >
        {children}
      </main>
    </div>
  );
}
