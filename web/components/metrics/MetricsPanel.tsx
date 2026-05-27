"use client";

import { useMetrics } from "@/hooks/useMetrics";

import styles from "./MetricsPanel.module.css";

// MetricsPanel は MainScreen に埋め込まれる今月のダメ化メトリクス表示 (LC-ME-09)。
// P-ME-FE-DEG-01: ThresholdExceeded=true のとき消化率を accent-danger でパルス演出 (NFRE-E08)。
// 警告状態は data-warn 属性で表現し、CSS Module で世界観 (Slot Machine) に統一。
export function MetricsPanel() {
  const { data: metrics, isLoading } = useMetrics();

  if (isLoading) {
    return (
      <div data-testid="metrics-panel-loading" className={styles.skeleton}>
        <div className={`${styles.skeletonBar} ${styles.skeletonBarLong}`} />
        <div className={`${styles.skeletonBar} ${styles.skeletonBarShort}`} />
      </div>
    );
  }

  if (!metrics) return null;

  const warn = metrics.thresholdExceeded;
  const pct = Math.round(metrics.consumptionRate * 100);

  return (
    <div data-testid="metrics-panel" className={styles.panel}>
      <div className={styles.label}>今月のダメ化</div>
      <div className={styles.count}>{metrics.damageCount} 回</div>

      <div className={styles.rateRow}>
        <span>消化率</span>
        <span
          data-testid="consumption-rate"
          data-warn={warn}
          className={styles.rateValue}
        >
          {pct}%
        </span>
      </div>
      <div className={styles.bar}>
        <div
          data-testid="consumption-bar"
          data-warn={warn}
          className={styles.barFill}
          style={{ width: `${pct}%` }}
        />
      </div>

      <p className={styles.summary}>{metrics.summaryText}</p>
    </div>
  );
}
