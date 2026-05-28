"use client";

import Link from "next/link";
import { useState } from "react";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";
import { COPY } from "@/lib/copy";

import styles from "./LandingScreen.module.css";

const DEMO_INITIAL_BALANCE = 1000;

export function LandingScreen() {
  const [demoUsed, setDemoUsed] = useState(false);
  const [demoBalance, setDemoBalance] = useState(DEMO_INITIAL_BALANCE);

  const handleDemo = () => {
    if (demoUsed) return;
    setDemoUsed(true);
    setDemoBalance(0);
  };

  return (
    <ScreenFrame testid="landing-screen" light>
      <BrandHeader light />

      <div className={styles.hero}>
        <p className={styles.eyebrow}>{COPY.landing.sub}</p>
        <h1 className={styles.h1}>{COPY.landing.heroLine1}</h1>
        <h1 className={styles.h1}>{COPY.landing.heroLine2}</h1>
      </div>

      <div className={styles.demoLabel}>{COPY.landing.demoBalanceLabel}</div>
      <div className={styles.demoAmount}>
        ¥{demoBalance.toLocaleString("ja-JP")}
      </div>

      <div className={styles.demoBtnWrap}>
        <button
          type="button"
          className={styles.demoBtn}
          data-done={demoUsed}
          onClick={handleDemo}
          disabled={demoUsed}
          data-testid="landing-demo-button"
          aria-label={demoUsed ? "体験用デモ完了" : "体験デモを試す"}
        >
          <span>{demoUsed ? "済" : COPY.landing.demoButtonMain}</span>
          {demoUsed && (
            <span className={styles.demoBtnSub}>
              {COPY.landing.demoButtonAfter}
            </span>
          )}
        </button>
      </div>

      <div className={styles.divider} />

      <div className={styles.ctaWrap} data-hidden={!demoUsed}>
        <Link
          href="/signup"
          role="button"
          data-testid="landing-start-button"
          className={styles.ctaPrimary}
        >
          {COPY.landing.ctaPrimary}
          <span className={styles.arrow} aria-hidden>
            →
          </span>
        </Link>
        <Link
          href="/login"
          data-testid="landing-login-link"
          className={styles.ctaSecondary}
        >
          {COPY.landing.ctaSecondary}
        </Link>
      </div>
    </ScreenFrame>
  );
}
