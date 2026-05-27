// AddBudgetForm: 既存予算がある状態 (= 通常運用中) に「いくら追加するか」を入力する
// メンタルモデルに合わせたフォーム。currentBudget + 選択額を POST し、backend の
// 「予算総額」モデル (apps/api/internal/wallet/service.go:141-197) で差額が
// 残高に加算される (delta > 0 のとき `balance += delta`)。
//
// なぜ別コンポーネントか:
//   - BudgetForm は「予算の総額を入力」モデル (初回設定 / 大幅変更用)。
//   - AddBudgetForm は「追加する金額」モデル (通常運用中の継ぎ足し)。
//   - 同じ画面内で両 UI を兼ねると認知負荷が高い + テストもややこしいので分離。
//
// VR-B-01/02 (合計が 1,000-100,000 円 / 1,000 円刻み) は backend が再検証するため
// クライアント側は「合計が上限内か」をチェックして disabled にするだけで足りる。
"use client";

import { useState } from "react";

import { useSetBudget } from "@/hooks/useSetBudget";

// 共通スタイル (form / chip / input / submit / error 等) は BudgetForm.module.css と
// 同じ design language を使うため共有 import。AddBudgetForm 専用の .preview のみ
// AddBudgetForm.module.css に分離 (review 指摘 Important 2)。
import styles from "./BudgetForm.module.css";
import addStyles from "./AddBudgetForm.module.css";

const QUICK_ADDITIONS: ReadonlyArray<number> = [10_000, 20_000, 30_000, 50_000];
const MAX_TOTAL = 100_000;
const STEP = 1_000;

type AddBudgetFormProps = {
  // 現在の予算 (確定済み、非 null)。null/undefined のときは BudgetForm 側を使うこと。
  currentBudget: number;
};

function validateAddition(
  addition: number | null,
  currentBudget: number,
): string | null {
  if (addition == null) return null;
  if (addition <= 0) return "追加金額は 1,000 円以上を選んでください";
  if (addition % STEP !== 0) return "1,000 円単位で入力してください";
  const total = currentBudget + addition;
  if (total > MAX_TOTAL) {
    return `合計が ${MAX_TOTAL.toLocaleString()} 円を超えるため設定できません`;
  }
  return null;
}

// 初期選択は「合計が上限を超えない最初のチップ」を選ぶ。currentBudget が上限近傍
// (例: 95,000) のときに上限超過する +10,000 を初期選択して即時 validation error を
// 出すのを避ける (review 指摘 Important 3)。全チップが上限超過の場合は null にして
// submit を disable にする (例: currentBudget=100,000 で実質的に追加不可)。
function pickInitialAddition(currentBudget: number): number | null {
  const fits = QUICK_ADDITIONS.find((v) => currentBudget + v <= MAX_TOTAL);
  return fits ?? null;
}

export function AddBudgetForm({ currentBudget }: AddBudgetFormProps) {
  const [addition, setAddition] = useState<number | null>(() =>
    pickInitialAddition(currentBudget),
  );
  const validationError = validateAddition(addition, currentBudget);

  const { mutate, isPending, error } = useSetBudget();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (addition == null || validationError != null) return;
    // backend の SetBudget は「予算総額」を受け取り delta を残高に反映する。
    mutate(currentBudget + addition);
  };

  const submitDisabled =
    addition == null || addition <= 0 || validationError != null || isPending;

  const totalAfter = addition != null ? currentBudget + addition : currentBudget;

  return (
    <form
      data-testid="add-budget-form"
      onSubmit={handleSubmit}
      className={styles.form}
    >
      <div className={styles.quickRow}>
        {QUICK_ADDITIONS.map((v) => {
          const disabled = currentBudget + v > MAX_TOTAL;
          const isSelected = addition === v;
          return (
            <button
              key={v}
              type="button"
              data-testid={`add-budget-quick-${v}`}
              aria-label={`${v.toLocaleString()} 円を追加`}
              onClick={() => setAddition(v)}
              disabled={disabled}
              className={`${styles.chip} ${isSelected ? styles.chipSelected : ""}`}
            >
              +¥{v.toLocaleString()}
            </button>
          );
        })}
      </div>

      <div className={styles.field}>
        <label htmlFor="add-budget-input">追加金額を直接入力</label>
        <input
          id="add-budget-input"
          type="number"
          data-testid="add-budget-input"
          aria-label="追加金額"
          inputMode="numeric"
          min={STEP}
          max={MAX_TOTAL - currentBudget}
          step={STEP}
          value={addition ?? ""}
          onChange={(e) => {
            const v = e.target.value;
            if (v === "") {
              setAddition(null);
              return;
            }
            const n = parseInt(v, 10);
            setAddition(Number.isNaN(n) ? null : n);
          }}
          className={styles.input}
        />
      </div>

      <div data-testid="add-budget-preview" className={addStyles.preview}>
        合計: ¥{totalAfter.toLocaleString()} (現在 ¥
        {currentBudget.toLocaleString()})
      </div>

      {validationError && (
        <div
          data-testid="add-budget-validation-error"
          role="alert"
          className={styles.error}
        >
          {validationError}
        </div>
      )}

      {error != null && (
        <div
          data-testid="add-budget-server-error"
          role="alert"
          className={styles.error}
        >
          設定に失敗しました。しばらくしてから再試行してください。
        </div>
      )}

      <button
        type="submit"
        data-testid="add-budget-submit"
        disabled={submitDisabled}
        className={styles.submit}
      >
        {isPending ? "追加中..." : "ダメ予算を追加する"}
      </button>
    </form>
  );
}
