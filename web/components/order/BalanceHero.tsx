// BalanceHero: mock/src/components/BalanceHero.tsx ベースに移植 (PR ④)。
// jotai 経由で値を取得しつつ、見た目と段階表現 (warn/danger/dead) を mock 準拠に揃える。
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

  // 三段階表示 (mock 準拠):
  //   - dead   : screenState=dead もしくは balance=0
  //   - danger : 消化 80% 以上 (dead でない)
  //   - warn   : 消化 60〜80% (dead/danger でない)
  const isDead = screenState === "dead" || balance === 0;
  const isDanger = !isDead && consumeRate >= 0.8;
  const isWarn = !isDead && !isDanger && consumeRate >= 0.6;
  const fillPct = Math.round(consumeRate * 100);

  const balanceText = balance < 0 ? "---" : yenFormat.format(balance);

  return (
    <section
      className={styles.wrap}
      role="status"
      aria-live="polite"
      data-testid="balance-hero"
    >
      <div className={styles.label}>{COPY.main.balanceLabel}</div>
      <div
        className={styles.amount}
        data-warn={isWarn}
        data-danger={isDanger}
        data-dead={isDead}
      >
        <span className={styles.yen}>¥</span>
        {balanceText}
      </div>
      <div className={styles.meter} aria-hidden="true">
        <div
          className={styles.meterFill}
          style={{ width: `${Math.min(100, fillPct)}%` }}
          data-warn={isWarn}
          data-danger={isDanger}
        />
      </div>
      <div className={styles.metrics}>
        <span>今月 {monthlyCount} 度</span>
        <span>消化 {fillPct}%</span>
      </div>
    </section>
  );
}
