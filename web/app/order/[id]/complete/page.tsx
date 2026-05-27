// LC-32 OrderCompletionScreen (NFRC-C21).
//
// 注文完了画面。5 秒後に MainScreen へ自動遷移する (US-1-07 / NFR-DEG-01)。
// PR ⑧ で mock/src/screens/CompleteMock.{tsx,module.css} 流に強化:
//   - lastOrderAtom から店名 / 金額を読んで実データ表示 (旧来は固定文字列)
//   - .glow (radial gradient) を背景に重ねる
//   - .verdict は verdict-letter-tighten-cm 演出 (letter-spacing 0.1em → -0.02em)
//   - 残りダメ予算 + 今月度数のメトリクス 2 行を追加
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAtomValue } from "jotai";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";
import { COPY, composeMonthlyMeta } from "@/lib/copy";
import {
  balanceAtom,
  lastOrderAtom,
  monthlyCountAtom,
} from "@/state/main";

import styles from "./CompleteScreen.module.css";

const AUTO_NAVIGATE_MS = 5000;

export default function OrderCompletePage() {
  const router = useRouter();
  const balance = useAtomValue(balanceAtom);
  const monthlyCount = useAtomValue(monthlyCountAtom);
  const lastOrder = useAtomValue(lastOrderAtom);

  useEffect(() => {
    const timer = setTimeout(() => router.push("/"), AUTO_NAVIGATE_MS);
    return () => clearTimeout(timer);
  }, [router]);

  // lastOrder は GoroButton.onSuccess で書き込まれる (PR ⑤)。直接 URL 叩きで本画面に
  // 来た場合 (リロード等) は null になり得るのでフォールバック表示を持つ。
  const storeName = lastOrder?.storeName ?? "—";
  const amount = lastOrder?.amount ?? 0;

  return (
    <ScreenFrame testid="order-completion-screen">
      <BrandHeader showLogout />
      <div className={styles.complete}>
        <div className={styles.glow} aria-hidden="true" />
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
