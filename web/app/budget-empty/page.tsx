"use client";

import { useState } from "react";

import { RaiseModal } from "@/components/metrics/RaiseModal";

import styles from "./BudgetEmptyScreen.module.css";

// BudgetEmptyScreen は残高 0 時に表示される画面 (LC-ME-10 / US-3-03)。
// P-ME-FE-DEG-01 §4.2: マウント時に useState(true) で RaiseModal を自動表示（退化ループ）。
export default function BudgetEmptyPage() {
  const [isRaiseModalOpen, setIsRaiseModalOpen] = useState(true);

  return (
    <main data-testid="budget-empty-screen" className={styles.screen}>
      <div className={styles.inner}>
        <p className={styles.emoji}>😔</p>
        <h1 className={styles.h1}>今月はもうダメになれません</h1>
        <p className={styles.body}>翌月 1 日に予算がリセットされます</p>
      </div>

      <RaiseModal
        isOpen={isRaiseModalOpen}
        onClose={() => setIsRaiseModalOpen(false)}
      />
    </main>
  );
}
