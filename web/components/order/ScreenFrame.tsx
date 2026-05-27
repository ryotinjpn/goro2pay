"use client";

import { ReactNode } from "react";
import type { ScreenState } from "@/state/main";

import styles from "./ScreenFrame.module.css";

type Props = {
  children: ReactNode;
  screenState?: ScreenState;
  testid?: string;
};

export function ScreenFrame({ children, screenState, testid }: Props) {
  return (
    <main
      className={styles.frame}
      data-screen-state={screenState}
      data-testid={testid}
    >
      {children}
    </main>
  );
}
