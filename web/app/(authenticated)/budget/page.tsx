// BudgetSetupScreen — 月間予算の設定 / 変更ページ (UC-B-01 / UC-B-02)。
//
// 既存ユーザは現在の予算を初期値に prefill (FD §2.1)。
// 新規ユーザは Wallet 未作成 (404) のため null prefill (BudgetForm 内で 30,000 デフォルト)。
"use client";

import { useWallet } from "@/hooks/useWallet";
import { BudgetForm } from "@/components/budget/BudgetForm";

export default function BudgetSetupScreen() {
  const { monthlyBudget, isLoading } = useWallet();
  const isExisting = !isLoading && monthlyBudget > 0;

  return (
    <main
      data-testid="budget-setup-screen"
      style={{
        maxWidth: 480,
        margin: "0 auto",
        padding: 20,
      }}
    >
      <h1 style={{ fontSize: 24, fontWeight: 700, marginBottom: 8 }}>
        {isExisting ? "ダメ予算を変更する" : "ダメ予算を設定する"}
      </h1>
      <p
        style={{
          color: "#6b7280",
          fontSize: 14,
          marginBottom: 20,
          lineHeight: 1.6,
        }}
      >
        月間ダメ予算を 1,000〜100,000 円 (1,000 円刻み) で設定してください。
        <br />
        即時反映され、月末にリセットされます。
      </p>
      {isLoading ? (
        <div data-testid="budget-setup-loading" style={{ color: "#9ca3af" }}>
          読込中...
        </div>
      ) : (
        <BudgetForm initialBudget={isExisting ? monthlyBudget : null} />
      )}
    </main>
  );
}
