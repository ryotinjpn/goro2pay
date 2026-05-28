"use client";

import { ModalPortal } from "@/components/common/ModalPortal";
import { COPY } from "@/lib/copy";
import styles from "./LogoutConfirmModal.module.css";

interface LogoutConfirmModalProps {
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

export function LogoutConfirmModal({
  open,
  onCancel,
  onConfirm,
}: LogoutConfirmModalProps) {
  if (!open) return null;
  return (
    <ModalPortal>
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="logout-confirm-title"
        data-testid="logout-confirm-modal"
        className={styles.overlay}
      >
        <div className={styles.card}>
          <h2 id="logout-confirm-title" className={styles.h2}>
            {COPY.logout.h2}
          </h2>
          <p className={styles.sub}>{COPY.logout.sub}</p>
          <div className={styles.actions}>
            <button
              type="button"
              onClick={onCancel}
              data-testid="logout-confirm-cancel"
              className={styles.secondary}
            >
              {COPY.logout.secondary}
            </button>
            <button
              type="button"
              onClick={onConfirm}
              data-testid="logout-confirm-submit"
              className={styles.primary}
            >
              {COPY.logout.primary}
            </button>
          </div>
        </div>
      </div>
    </ModalPortal>
  );
}
