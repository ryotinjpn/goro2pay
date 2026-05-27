// web/lib/budgetMath.ts
//
// DEAD 時の増額提案値を計算する (spec §3.5 / FR-METRICS-04)。
// 算出式: 現在予算 × 5/3 を 10,000 円単位で四捨五入。

const RATIO = 5 / 3;
const ROUND_UNIT = 10000;

export function recommendNextBudget(current: number): number {
  const raw = current * RATIO;
  const rounded = Math.round(raw / ROUND_UNIT) * ROUND_UNIT;
  return rounded > current ? rounded : current + ROUND_UNIT;
}
