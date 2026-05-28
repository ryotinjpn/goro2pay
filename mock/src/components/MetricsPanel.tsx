// MetricsPanel: web 側 LC-ME-09 の見た目を mock に移植。
// useMetrics 等の API は無いので、表示値は MainMock から props で渡す。
//
// 表示要素:
//   - 「今月のダメ化」ラベル + 大カウント (extra-bold sans)
//   - 消化率 (warn 60-80% / danger 80%+)
//   - グラデメーター
//   - サマリ (今月のダメ化回数 + 消化額)
import styles from './MetricsPanel.module.css';

type Props = {
  damageCount: number;
  /** 0..1 (= 消化額 / 月間予算) */
  consumptionRate: number;
  /** 当月の消化額合計 (円) */
  amountUsed: number;
};

export default function MetricsPanel({
  damageCount,
  consumptionRate,
  amountUsed,
}: Props) {
  const pct = Math.min(100, Math.round(consumptionRate * 100));
  const warn = pct >= 60 && pct < 80;
  const danger = pct >= 80;

  return (
    <div className={styles.panel}>
      <div className={styles.label}>今月のダメ化</div>
      <div className={styles.count}>{damageCount} 回</div>

      <div className={styles.rateRow}>
        <span>消化率</span>
        <span
          data-warn={warn || undefined}
          data-danger={danger || undefined}
          className={styles.rateValue}
        >
          {pct}%
        </span>
      </div>
      <div className={styles.bar}>
        <div
          data-warn={warn || undefined}
          data-danger={danger || undefined}
          className={styles.barFill}
          style={{ width: `${pct}%` }}
        />
      </div>

      <p className={styles.summary}>
        今月のダメ化回数: {damageCount} 回、消化額 ¥
        {amountUsed.toLocaleString()}
      </p>
    </div>
  );
}
