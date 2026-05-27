import { COPY } from '../lib/copy';
import styles from './BalanceHero.module.css';

type Props = {
  balance: number;
  initialBudget: number;
  monthlyCount: number;
  dead?: boolean;
};

export default function BalanceHero({
  balance,
  initialBudget,
  monthlyCount,
  dead = false,
}: Props) {
  const consumed = Math.max(0, initialBudget - balance);
  const rate = initialBudget > 0 ? consumed / initialBudget : 0;
  const ratePct = Math.round(rate * 100);
  const warn = rate >= 0.6 && rate < 0.8;
  const danger = rate >= 0.8 && !dead;

  return (
    <div className={styles.wrap}>
      <div className={styles.label}>{COPY.main.balanceLabel}</div>
      <div
        className={styles.amount}
        data-warn={warn}
        data-danger={danger}
        data-dead={dead}
        role="status"
        aria-live="polite"
      >
        <span className={styles.yen}>¥</span>
        {balance.toLocaleString()}
      </div>
      <div className={styles.meter}>
        <div
          className={styles.meterFill}
          data-warn={warn}
          data-danger={danger}
          style={{ width: `${Math.min(100, ratePct)}%` }}
        />
      </div>
      <div className={styles.metrics}>
        <span>今月 {monthlyCount} 度</span>
        <span>消化 {ratePct}%</span>
      </div>
    </div>
  );
}
