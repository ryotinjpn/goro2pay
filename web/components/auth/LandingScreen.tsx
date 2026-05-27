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
    <ScreenFrame testid="landing-screen">
      <BrandHeader />
      <div className={styles.hero}>
        <div className={styles.h1}>
          <div>{COPY.landing.heroLine1}</div>
          <div>{COPY.landing.heroLine2}</div>
        </div>
        <div className={styles.sub}>{COPY.landing.sub}</div>
        <div className={styles.demoLabel}>
          {COPY.landing.demoBalanceLabel}{" "}
          <b>¥{demoBalance.toLocaleString("ja-JP")}</b>
        </div>
        <button
          type="button"
          className={styles.demoButton}
          onClick={handleDemo}
          disabled={demoUsed}
          data-testid="landing-demo-button"
          aria-label={demoUsed ? "体験用デモ完了" : "体験デモを試す"}
        >
          {demoUsed ? COPY.landing.demoButtonAfter : COPY.landing.demoButtonMain}
        </button>
      </div>
      <div className={styles.cta} data-visible={demoUsed}>
        <Link
          href="/signup"
          role="button"
          data-testid="landing-start-button"
          className={styles.ctaPrimary}
        >
          {COPY.landing.ctaPrimary}
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
