"use client";

interface LogoutConfirmModalProps {
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

export function LogoutConfirmModal({ open, onCancel, onConfirm }: LogoutConfirmModalProps) {
  if (!open) return null;
  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="logout-confirm-title"
      data-testid="logout-confirm-modal"
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.4)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <div style={{ background: "white", padding: 24, borderRadius: 8 }}>
        <h2 id="logout-confirm-title">もうダメ化を終わらせますか？</h2>
        <p>もう一度入るには再ログインが必要です</p>
        <div style={{ display: "flex", gap: 12, justifyContent: "flex-end" }}>
          <button type="button" onClick={onCancel} data-testid="logout-confirm-cancel">
            キャンセル
          </button>
          <button type="button" onClick={onConfirm} data-testid="logout-confirm-submit">
            ログアウト
          </button>
        </div>
      </div>
    </div>
  );
}
