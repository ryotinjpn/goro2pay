"use client";

import { useEffect } from "react";
import { useAtomValue, useSetAtom } from "jotai";

import { InsufficientBalanceModal } from "@/components/budget/InsufficientBalanceModal";
import { MetricsPanel } from "@/components/metrics/MetricsPanel";
import { useMetrics } from "@/hooks/useMetrics";
import { useWallet } from "@/hooks/useWallet";
import { COPY } from "@/lib/copy";
import {
  balanceAtom,
  monthlyBudgetAtom,
  monthlyCountAtom,
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
  const setBalance = useSetAtom(balanceAtom);
  const setMonthlyCount = useSetAtom(monthlyCountAtom);
  const winFlashCounter = useAtomValue(winFlashCounterAtom);

  // BalanceHero / GoroButton 側は jotai atom を一次ソースとして読むが、
  // 実残高 / 実回数は API から取得する必要がある。MainScreen mount 時に
  // useWallet() / useMetrics() で fetch し、結果を atom に同期する。
  // GoroButton.onSuccess も同じ atom を更新するため、最終値は楽観的更新と
  // 整合する (注文成功 → setBalance/setMonthlyCount → 次回 invalidate で
  // refetch されたら同じ値で上書き)。
  const { balance: walletBalance, monthlyBudget: walletMonthlyBudget } = useWallet();
  const metricsQuery = useMetrics();
  const damageCount = metricsQuery.data?.damageCount;
  useEffect(() => {
    if (walletBalance !== undefined) setBalance(walletBalance);
    if (walletMonthlyBudget !== undefined) setMonthlyBudget(walletMonthlyBudget);
    if (damageCount !== undefined) setMonthlyCount(damageCount);
  }, [
    walletBalance,
    walletMonthlyBudget,
    damageCount,
    setBalance,
    setMonthlyBudget,
    setMonthlyCount,
  ]);

  const handleIncrease = (nextBudget: number) => {
    setMonthlyBudget(nextBudget);
  };

  // mock 流のグローバル演出層: 注文成功時に GoroButton.onSuccess が
  // winFlashCounterAtom に Date.now() を書き込み、ここで `key={counter}` で
  // remount し win-flash-screen / win-verdict-pop を再走させる。
  // (handleClick 冒頭で 0 にリセットされるため、slot 突入の瞬間は false)
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
