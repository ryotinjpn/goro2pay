"use client";

import { useAtomValue } from "jotai";

import { COPY } from "@/lib/copy";
import {
  balanceAtom,
  consumeRateAtom,
  monthlyCountAtom,
  screenStateAtom,
} from "@/state/main";

import styles from "./BalanceHero.module.css";

const yenFormat = new Intl.NumberFormat("ja-JP");

export function BalanceHero() {
  const balance = useAtomValue(balanceAtom);
  const consumeRate = useAtomValue(consumeRateAtom);
  const monthlyCount = useAtomValue(monthlyCountAtom);
  const screenState = useAtomValue(screenStateAtom);

  const isDead = screenState === "dead" || balance === 0;
  const isLow = !isDead && consumeRate >= 0.8;
  const isWarn = !isDead && !isLow && consumeRate >= 0.6;
  const fillPct = Math.round(consumeRate * 100);

  const balanceText =
    balance < 0 ? "---" : yenFormat.format(balance);

  return (
    <section
      className={styles.block}
      role="status"
      aria-live="polite"
      data-testid="balance-hero"
    >
      <div className={styles.label}>{COPY.main.balanceLabel}</div>
      <div
        className={styles.amount}
        data-low={isLow}
        data-dead={isDead}
      >
        <span className={styles.yen}>¥</span>
        {balanceText}
      </div>
      <div className={styles.meter} aria-hidden="true">
        <div
          className={styles.meterFill}
          style={{ width: `${fillPct}%` }}
          data-warn={isWarn}
          data-danger={isLow}
        />
      </div>
      <div className={styles.metrics}>
        <span>今月 {monthlyCount} 度</span>
        <span>消化 {fillPct}%</span>
      </div>
    </section>
  );
}
