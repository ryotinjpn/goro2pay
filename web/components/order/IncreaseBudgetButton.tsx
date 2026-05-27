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
    <button
      type="button"
      className={styles.button}
      onClick={() => onClick(nextBudget)}
      data-testid="increase-budget-button"
    >
      {label}
    </button>
  );
}
