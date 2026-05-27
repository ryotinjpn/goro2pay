"use client";

import { useBudgetRaise } from "@/hooks/useBudgetRaise";

type Props = {
  isOpen: boolean;
  onClose: () => void;
};

// RaiseModal は増額誘導モーダル (LC-ME-11)。
// P-ME-FE-DEG-01 §4.3 のダメコピーを使用してユーザーを退化ループへ誘導する。
export function RaiseModal({ isOpen, onClose }: Props) {
  const { recommendation, mutation } = useBudgetRaise();

  if (!isOpen) return null;

  const recommended = recommendation.data?.recommendedMonthlyBudget;

  const handleAccept = () => {
    if (!recommended) return;
    mutation.mutate(recommended, { onSuccess: onClose });
  };

  return (
    <div
      data-testid="raise-modal-overlay"
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
    >
      <div
        data-testid="raise-modal"
        className="bg-white rounded-2xl p-6 mx-4 max-w-sm w-full space-y-4"
      >
        <h2 className="text-lg font-bold text-center">翌月予算を増額しますか？</h2>

        {recommended ? (
          <p className="text-center text-gray-600 text-sm">
            推奨: ¥{recommended.toLocaleString()}
            <span className="text-xs text-gray-400 block">（あなたには必要です）</span>
          </p>
        ) : (
          <div className="h-8 bg-gray-100 rounded animate-pulse" />
        )}

        <div className="space-y-2">
          <button
            data-testid="raise-modal-accept"
            onClick={handleAccept}
            disabled={!recommended || mutation.isPending}
            className="w-full bg-blue-500 text-white py-3 rounded-xl font-bold disabled:opacity-50"
          >
            {mutation.isPending ? "処理中..." : "増額する"}
          </button>
          <button
            data-testid="raise-modal-reject"
            onClick={onClose}
            className="w-full text-gray-500 py-2 text-sm"
          >
            今月はがんばる
          </button>
        </div>
      </div>
    </div>
  );
}
