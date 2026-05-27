// LC-32 OrderCompletionScreen (NFRC-C21)
//
// 注文完了画面。5 秒後に MainScreen へ自動遷移する (US-1-07 / NFR-DEG-01)。
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAtomValue } from "jotai";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";
import { COPY, composeMonthlyMeta } from "@/lib/copy";
import { balanceAtom, monthlyCountAtom } from "@/state/main";

import styles from "./CompleteScreen.module.css";

const AUTO_NAVIGATE_MS = 5000;

export default function OrderCompletePage() {
  const router = useRouter();
  const balance = useAtomValue(balanceAtom);
  const monthlyCount = useAtomValue(monthlyCountAtom);

  useEffect(() => {
    const timer = setTimeout(() => router.push("/"), AUTO_NAVIGATE_MS);
    return () => clearTimeout(timer);
  }, [router]);

  // 店名・金額: jotai に永続化していないため固定 (将来 lastOrderAtom で改善)
  const storeName = "CoCo壱番屋 新宿店";
  const amount = 1200;

  return (
    <ScreenFrame testid="order-completion-screen">
      <BrandHeader showLogout />
      <div className={styles.complete}>
        <div className={styles.verdict}>{COPY.complete.verdict}</div>
        <div className={styles.body}>{COPY.complete.body}</div>
        <div className={styles.store}>{storeName}</div>
        <div className={styles.price}>¥{amount.toLocaleString("ja-JP")}</div>
        <div className={styles.line} />
        <div className={styles.meta}>
          {composeMonthlyMeta(monthlyCount > 0 ? monthlyCount : 1)}
        </div>
        <div className={`${styles.meta} ${styles.metaMute}`}>
          残りダメ予算 ¥{balance >= 0 ? balance.toLocaleString("ja-JP") : "---"}
        </div>
        <button
          type="button"
          className={styles.next}
          onClick={() => router.push("/")}
        >
          {COPY.complete.next}
        </button>
      </div>
    </ScreenFrame>
  );
}
