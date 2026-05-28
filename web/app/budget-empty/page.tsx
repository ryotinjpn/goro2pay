// BudgetEmptyScreen は残高 0 時に表示される画面 (LC-ME-10 / US-3-03)。
// P-ME-FE-DEG-01 §4.2: マウント時に useState(true) で RaiseModal を自動表示 (退化ループ)。
//
// PR ⑧ で Tailwind 直書きの素朴版から、ScreenFrame + 世界観準拠 (BrandHeader + 中央寄せ
// 演出) に書き直した。😔 emoji と grey 文字は世界観に合わないため廃止。
"use client";

import { useState } from "react";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";
import { RaiseModal } from "@/components/metrics/RaiseModal";
import { COPY } from "@/lib/copy";

import styles from "./page.module.css";

export default function BudgetEmptyPage() {
  const [isRaiseModalOpen, setIsRaiseModalOpen] = useState(true);

  return (
    <ScreenFrame testid="budget-empty-screen" light>
      <BrandHeader light showLogout />
      <div className={styles.body}>
        <div className={styles.glow} aria-hidden="true" />
        <div className={styles.verdict}>{COPY.budgetEmpty.verdict}</div>
        <div className={styles.sub}>{COPY.budgetEmpty.sub}</div>
      </div>
      <RaiseModal
        isOpen={isRaiseModalOpen}
        onClose={() => setIsRaiseModalOpen(false)}
      />
    </ScreenFrame>
  );
}
