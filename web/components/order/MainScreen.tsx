"use client";

import { useAtomValue, useSetAtom } from "jotai";

import { InsufficientBalanceModal } from "@/components/budget/InsufficientBalanceModal";
import { MetricsPanel } from "@/components/metrics/MetricsPanel";
import { COPY } from "@/lib/copy";
import {
  monthlyBudgetAtom,
  screenStateAtom,
  winFlashCounterAtom,
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
  const winFlashCounter = useAtomValue(winFlashCounterAtom);

  const handleIncrease = (nextBudget: number) => {
    setMonthlyBudget(nextBudget);
  };

  // mock 流のグローバル演出層: 注文成功時に GoroButton.onSuccess が
  // winFlashCounterAtom を increment → ここで `key={counter}` で remount し
  // win-flash-screen / win-verdict-pop を再走させる。
  // ガード 2 段:
  //   - counter > 0     : 初回 mount 時のスタブ表示防止 (PR ① の atom 契約)
  //   - screenState=slot: navigate 後の Complete 画面に演出が漏れないようにする
  //                       (mock とは別判断: mock は flash を slot ガードしないが、
  //                        web は短い slot 滞在中に発火させた方が UX 純度が高い)
  const showWinEffects = winFlashCounter > 0 && screenState === "slot";

  return (
    <>
      {showWinEffects && (
        <div
          key={`flash-${winFlashCounter}`}
          className={styles.winFlash}
          aria-hidden="true"
        />
      )}
      {showWinEffects && (
        <div
          key={`verdict-${winFlashCounter}`}
          className={styles.winVerdict}
          role="status"
          aria-live="polite"
        >
          {COPY.main.winVerdict}
        </div>
      )}
      <ScreenFrame screenState={screenState} testid="main-screen">
        <BrandHeader showLogout />
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
    </>
  );
}
