"use client";

import { useAtomValue, useSetAtom } from "jotai";

import { InsufficientBalanceModal } from "@/components/budget/InsufficientBalanceModal";
import { MetricsPanel } from "@/components/metrics/MetricsPanel";
import {
  monthlyBudgetAtom,
  screenStateAtom,
} from "@/state/main";

import { ScreenFrame } from "./ScreenFrame";
import { BrandHeader } from "./BrandHeader";
import { BalanceHero } from "./BalanceHero";
import { GoroButton } from "./GoroButton";
import { DeadVerdict } from "./DeadVerdict";
import { IncreaseBudgetButton } from "./IncreaseBudgetButton";

import styles from "./MainScreen.module.css";

export function MainScreen() {
  const screenState = useAtomValue(screenStateAtom);
  const setMonthlyBudget = useSetAtom(monthlyBudgetAtom);

  const handleIncrease = (nextBudget: number) => {
    setMonthlyBudget(nextBudget);
  };

  return (
    <ScreenFrame screenState={screenState} testid="main-screen">
      <BrandHeader />
      <BalanceHero />
      <MetricsPanel />
      <div className={styles.body}>
        <GoroButton />
      </div>
      {screenState === "dead" && (
        <>
          <DeadVerdict />
          <IncreaseBudgetButton onClick={handleIncrease} />
        </>
      )}
      <InsufficientBalanceModal />
    </ScreenFrame>
  );
}
