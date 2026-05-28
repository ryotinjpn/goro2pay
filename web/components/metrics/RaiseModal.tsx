"use client";

import { ModalPortal } from "@/components/common/ModalPortal";
import { useBudgetRaise } from "@/hooks/useBudgetRaise";

import styles from "./RaiseModal.module.css";

type Props = {
  isOpen: boolean;
  onClose: () => void;
};

// RaiseModal は増額誘導モーダル (LC-ME-11)。
// P-ME-FE-DEG-01 §4.3 のダメコピーを使用してユーザーを退化ループへ誘導する。
// LogoutConfirmModal / SessionExpiredModal と同じ overlay/card パターンで世界観統一。
export function RaiseModal({ isOpen, onClose }: Props) {
  const { recommendation, mutation } = useBudgetRaise();

  if (!isOpen) return null;

  const recommended = recommendation.data?.recommendedMonthlyBudget;

  const handleAccept = () => {
    if (!recommended) return;
    mutation.mutate(recommended, { onSuccess: onClose });
  };

  return (
    <ModalPortal>
      <div data-testid="raise-modal-overlay" className={styles.overlay}>
        <div
          data-testid="raise-modal"
          role="dialog"
          aria-modal="true"
          aria-labelledby="raise-modal-title"
          className={styles.card}
        >
          <h2 id="raise-modal-title" className={styles.h2}>
            翌月予算を増額しますか？
          </h2>

          {recommended ? (
            <p className={styles.recommend}>
              推奨
              <span className={styles.recommendAmount}>
                ¥{recommended.toLocaleString()}
              </span>
              <span className={styles.recommendNote}>あなたには必要です</span>
            </p>
          ) : (
            <div className={styles.recommendSkeleton} />
          )}

          <div className={styles.actions}>
            <button
              data-testid="raise-modal-accept"
              onClick={handleAccept}
              disabled={!recommended || mutation.isPending}
              className={styles.primary}
            >
              {mutation.isPending ? "処理中..." : "増額する"}
            </button>
            <button
              data-testid="raise-modal-reject"
              onClick={onClose}
              className={styles.secondary}
            >
              今月はがんばる
            </button>
          </div>
        </div>
      </div>
    </ModalPortal>
  );
}
