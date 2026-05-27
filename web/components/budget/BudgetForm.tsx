// LC-BUDGET-12 BudgetSetupScreen のフォーム本体 (FD §2.1)。
//
// クイックボタン 5 個 + 数値入力 + 送信ボタン + リアルタイムバリデーション。
// VR-B-01 (1,000-100,000 円) / VR-B-02 (1,000 円刻み) を入力時点で弾く。
"use client";

import { useState } from "react";

import { useSetBudget } from "@/hooks/useSetBudget";

import styles from "./BudgetForm.module.css";

const QUICK_BUDGETS: ReadonlyArray<number> = [10_000, 30_000, 50_000, 80_000, 100_000];
const RECOMMENDED = 30_000;

const MIN = 1_000;
const MAX = 100_000;
const STEP = 1_000;

function validateBudget(value: number | null): string | null {
  if (value == null) return null; // submit ボタンを disabled にするだけ
  if (value < MIN || value > MAX) {
    return "1,000 〜 100,000 円の範囲で入力してください";
  }
  if (value % STEP !== 0) {
    return "1,000 円単位で入力してください";
  }
  return null;
}

type BudgetFormProps = {
  /** 既存予算の prefill (変更画面用)。新規ユーザは null。 */
  initialBudget?: number | null;
};

export function BudgetForm({ initialBudget = null }: BudgetFormProps) {
  const [monthlyBudget, setMonthlyBudget] = useState<number | null>(
    initialBudget ?? RECOMMENDED,
  );
  const validationError = validateBudget(monthlyBudget);

  const { mutate, isPending, error } = useSetBudget();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (monthlyBudget == null || validationError != null) return;
    mutate(monthlyBudget);
  };

  const submitDisabled =
    monthlyBudget == null || validationError != null || isPending;

  return (
    <form data-testid="budget-form" onSubmit={handleSubmit} className={styles.form}>
      <div className={styles.quickRow}>
        {QUICK_BUDGETS.map((v) => {
          const isSelected = monthlyBudget === v;
          const isRecommended = v === RECOMMENDED;
          return (
            <button
              key={v}
              type="button"
              data-testid={`budget-quick-${v}`}
              aria-label={`${v.toLocaleString()} 円を選択`}
              onClick={() => setMonthlyBudget(v)}
              className={`${styles.chip} ${isSelected ? styles.chipSelected : ""}`}
            >
              {isRecommended ? <span className={styles.recommendStar}>★</span> : null}
              ¥{v.toLocaleString()}
            </button>
          );
        })}
      </div>

      <div className={styles.field}>
        <label htmlFor="budget-input">金額を直接入力</label>
        <input
          id="budget-input"
          type="number"
          data-testid="budget-input"
          aria-label="月間ダメ予算"
          inputMode="numeric"
          min={MIN}
          max={MAX}
          step={STEP}
          value={monthlyBudget ?? ""}
          onChange={(e) => {
            const v = e.target.value;
            if (v === "") {
              setMonthlyBudget(null);
              return;
            }
            const n = parseInt(v, 10);
            setMonthlyBudget(Number.isNaN(n) ? null : n);
          }}
          className={styles.input}
        />
      </div>

      {validationError && (
        <div
          data-testid="budget-validation-error"
          role="alert"
          className={styles.error}
        >
          {validationError}
        </div>
      )}

      {error != null && (
        <div
          data-testid="budget-server-error"
          role="alert"
          className={styles.error}
        >
          設定に失敗しました。しばらくしてから再試行してください。
        </div>
      )}

      <button
        type="submit"
        data-testid="budget-submit"
        disabled={submitDisabled}
        className={styles.submit}
      >
        {isPending ? "設定中..." : "ダメ予算を設定する"}
      </button>
    </form>
  );
}
