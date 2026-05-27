import { FormEvent, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './BudgetSetupMock.module.css';

const QUICK_BUDGETS: ReadonlyArray<number> = [10_000, 30_000, 50_000, 80_000, 100_000];
const RECOMMENDED = 30_000;
const MIN = 1_000;
const MAX = 100_000;
const STEP = 1_000;

function validateBudget(value: number | null): string | null {
  if (value == null) return null;
  if (value < MIN || value > MAX) {
    return '1,000 〜 100,000 円の範囲で入力してください';
  }
  if (value % STEP !== 0) {
    return '1,000 円単位で入力してください';
  }
  return null;
}

export default function BudgetSetupMock() {
  const navigate = useNavigate();
  const [monthlyBudget, setMonthlyBudget] = useState<number | null>(RECOMMENDED);
  const [submitting, setSubmitting] = useState(false);

  const validationError = validateBudget(monthlyBudget);
  const submitDisabled =
    monthlyBudget == null || validationError != null || submitting;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (submitDisabled) return;
    setSubmitting(true);
    setTimeout(() => {
      setSubmitting(false);
      navigate('/main');
    }, 800);
  };

  return (
    <ScreenFrame>
      <BrandHeader />
      <h1 className={styles.h1}>{COPY.budget.h1}</h1>
      <p className={styles.desc}>{COPY.budget.desc}</p>

      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.quickRow}>
          {QUICK_BUDGETS.map((v) => {
            const isSelected = monthlyBudget === v;
            const isRecommended = v === RECOMMENDED;
            return (
              <button
                key={v}
                type="button"
                aria-label={`${v.toLocaleString()} 円を選択`}
                onClick={() => setMonthlyBudget(v)}
                className={`${styles.chip} ${isSelected ? styles.chipSelected : ''}`}
              >
                {isRecommended ? <span className={styles.recommendStar}>★</span> : null}
                ¥{v.toLocaleString()}
              </button>
            );
          })}
        </div>

        <div className={styles.field}>
          <label htmlFor="budget-input">{COPY.budget.inputLabel}</label>
          <input
            id="budget-input"
            type="number"
            inputMode="numeric"
            min={MIN}
            max={MAX}
            step={STEP}
            value={monthlyBudget ?? ''}
            onChange={(e) => {
              const v = e.target.value;
              if (v === '') {
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
          <div role="alert" className={styles.error}>
            {validationError}
          </div>
        )}

        <button
          type="submit"
          disabled={submitDisabled}
          className={styles.submit}
        >
          {submitting ? '設定中...' : COPY.budget.submit}
        </button>
      </form>
    </ScreenFrame>
  );
}
