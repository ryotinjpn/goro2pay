// IncreaseBudgetButton: mock/src/components/IncreaseBudgetButton.tsx ベースに移植 (PR ⑦)。
// 旧来の画面下端 absolute 配置 + saturate filter からフロー内挿入 (.wrap で flex center) に変更。
// MainScreen の dead 状態で DeadVerdict の直下に自然に並ぶ。
"use client";

import { useAtomValue } from "jotai";

import { composeIncreaseBudgetLabel } from "@/lib/copy";
import { recommendNextBudget } from "@/lib/budgetMath";
import { monthlyBudgetAtom } from "@/state/main";

import styles from "./IncreaseBudgetButton.module.css";

type Props = {
  onClick: (nextBudget: number) => void;
};

export function IncreaseBudgetButton({ onClick }: Props) {
  const currentBudget = useAtomValue(monthlyBudgetAtom);
  const nextBudget = recommendNextBudget(currentBudget);
  const label = composeIncreaseBudgetLabel(nextBudget);

  return (
    <div className={styles.wrap}>
      <button
        type="button"
        className={styles.button}
        onClick={() => onClick(nextBudget)}
        data-testid="increase-budget-button"
      >
        {label}
      </button>
    </div>
  );
}
