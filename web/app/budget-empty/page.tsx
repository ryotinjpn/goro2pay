"use client";

import { useState } from "react";

import { RaiseModal } from "@/components/metrics/RaiseModal";

// BudgetEmptyScreen は残高 0 時に表示される画面 (LC-ME-10 / US-3-03)。
// P-ME-FE-DEG-01 §4.2: マウント時に useState(true) で RaiseModal を自動表示（退化ループ）。
export default function BudgetEmptyPage() {
  const [isRaiseModalOpen, setIsRaiseModalOpen] = useState(true);

  return (
    <main
      data-testid="budget-empty-screen"
      className="min-h-screen flex flex-col items-center justify-center p-6 text-center"
    >
      <div className="space-y-4 max-w-sm">
        <p className="text-4xl">😔</p>
        <h1 className="text-xl font-bold text-gray-800">今月はもうダメになれません</h1>
        <p className="text-sm text-gray-500">翌月 1 日に予算がリセットされます</p>
      </div>

      <RaiseModal
        isOpen={isRaiseModalOpen}
        onClose={() => setIsRaiseModalOpen(false)}
      />
    </main>
  );
}
