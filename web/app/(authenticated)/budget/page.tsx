// BudgetSetupScreen — 月間予算の設定 / 変更ページ (UC-B-01 / UC-B-02)。
//
// 初回 (Wallet 未作成 = 404 / monthlyBudget=0): BudgetForm で「月間予算の総額」を入力させる
// 既存ユーザ                                    : AddBudgetForm で「いくら追加するか」を入力させる
//
// backend (apps/api/internal/wallet/service.go SetBudget) は「予算の総額」を受け取り
// 差額を残高に反映する仕様。ユーザのメンタルモデル「追加金額を残高に足したい」に
// 揃えるため、AddBudgetForm が currentBudget + addition を計算して送る役割を担う。
"use client";

import { useWallet } from "@/hooks/useWallet";
import { BudgetForm } from "@/components/budget/BudgetForm";
import { AddBudgetForm } from "@/components/budget/AddBudgetForm";
import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";

import styles from "./BudgetSetupScreen.module.css";

export default function BudgetSetupScreen() {
  const { monthlyBudget, isLoading } = useWallet();
  const isExisting = !isLoading && monthlyBudget > 0;

  return (
    <ScreenFrame testid="budget-setup-screen">
      <BrandHeader showLogout />
      <h1 className={styles.h1}>
        {isExisting ? "ダメ予算を追加する" : "ダメ予算を設定する"}
      </h1>
      <p className={styles.sub}>
        {isExisting ? (
          <>
            追加分は即時に残高に加算される。
            <br />
            合計 1,000〜100,000 円 (1,000 円刻み) で設定可能。
          </>
        ) : (
          <>
            月間ダメ予算を 1,000〜100,000 円 (1,000 円刻み) で設定してください。
            <br />
            即時反映され、月末にリセットされます。
          </>
        )}
      </p>
      <div className={styles.body}>
        {isLoading ? (
          <div data-testid="budget-setup-loading" className={styles.loading}>
            読込中...
          </div>
        ) : isExisting ? (
          <AddBudgetForm currentBudget={monthlyBudget} />
        ) : (
          <BudgetForm initialBudget={null} />
        )}
      </div>
    </ScreenFrame>
  );
}
