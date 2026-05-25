// LC-BUDGET-12 BudgetSetupScreen のフォーム本体 (FD §2.1)。
//
// クイックボタン 5 個 + 数値入力 + 送信ボタン + リアルタイムバリデーション。
// VR-B-01 (1,000-100,000 円) / VR-B-02 (1,000 円刻み) を入力時点で弾く。
"use client";

import { useState } from "react";

import { useSetBudget } from "@/hooks/useSetBudget";

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

  return (
    <form
      data-testid="budget-form"
      onSubmit={handleSubmit}
      style={{ display: "flex", flexDirection: "column", gap: 16 }}
    >
      <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
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
              style={{
                padding: "10px 14px",
                borderRadius: 8,
                border: isSelected ? "2px solid #2563eb" : "1px solid #d1d5db",
                background: isSelected ? "#dbeafe" : "#ffffff",
                cursor: "pointer",
                fontSize: 14,
              }}
            >
              {isRecommended ? "★ " : ""}
              ¥{v.toLocaleString()}
            </button>
          );
        })}
      </div>

      <label
        style={{
          display: "flex",
          flexDirection: "column",
          gap: 4,
          fontSize: 13,
          color: "#374151",
        }}
      >
        金額を直接入力
        <input
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
          style={{
            padding: "10px 12px",
            borderRadius: 8,
            border: "1px solid #d1d5db",
            fontSize: 16,
          }}
        />
      </label>

      {validationError && (
        <div
          data-testid="budget-validation-error"
          role="alert"
          style={{ color: "#dc2626", fontSize: 13 }}
        >
          {validationError}
        </div>
      )}

      {error != null && (
        <div
          data-testid="budget-server-error"
          role="alert"
          style={{ color: "#dc2626", fontSize: 13 }}
        >
          設定に失敗しました。しばらくしてから再試行してください。
        </div>
      )}

      <button
        type="submit"
        data-testid="budget-submit"
        disabled={monthlyBudget == null || validationError != null || isPending}
        style={{
          padding: "12px 16px",
          borderRadius: 8,
          background:
            monthlyBudget == null || validationError != null || isPending
              ? "#9ca3af"
              : "#2563eb",
          color: "#ffffff",
          border: "none",
          fontSize: 16,
          cursor:
            monthlyBudget == null || validationError != null || isPending
              ? "not-allowed"
              : "pointer",
        }}
      >
        {isPending ? "設定中..." : "ダメ予算を設定する"}
      </button>
    </form>
  );
}
