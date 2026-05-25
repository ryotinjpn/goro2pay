// Unit B BalanceDisplay の表示分岐検証。
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { BalanceDisplay } from "@/components/budget/BalanceDisplay";
import * as walletApi from "@/lib/api/wallet";

function renderWithClient(ui: React.ReactElement) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  return render(
    <QueryClientProvider client={client}>{ui}</QueryClientProvider>,
  );
}

describe("BalanceDisplay (LC-BUDGET-09 / P-DEG-01)", () => {
  beforeEach(() => {
    vi.spyOn(walletApi, "getWallet");
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("初回ロード中はスケルトンを表示", () => {
    vi.mocked(walletApi.getWallet).mockImplementation(() => new Promise(() => {}));
    renderWithClient(<BalanceDisplay />);
    expect(screen.getByTestId("budget-balance-skeleton")).toBeInTheDocument();
  });

  it("成功時は残高 + 月間予算 + リセット日を表示", async () => {
    vi.mocked(walletApi.getWallet).mockResolvedValue({
      balance: 28800,
      monthlyBudget: 30000,
      updatedAt: "2026-05-24T00:00:00Z",
    });
    renderWithClient(<BalanceDisplay />);
    expect(await screen.findByTestId("budget-balance-amount")).toHaveTextContent("¥28,800");
    expect(screen.getByTestId("budget-reset-countdown")).toBeInTheDocument();
  });

  it("消化率 > 0.8 で警告色 (⚠ プレフィックス)", async () => {
    vi.mocked(walletApi.getWallet).mockResolvedValue({
      balance: 5000, // 30000 → 5000 = 消化率 約 0.83
      monthlyBudget: 30000,
      updatedAt: "2026-05-24T00:00:00Z",
    });
    renderWithClient(<BalanceDisplay />);
    const amount = await screen.findByTestId("budget-balance-amount");
    expect(amount.textContent).toContain("⚠");
  });

  it("失敗時はエラー表示 + 再試行ボタン", async () => {
    vi.mocked(walletApi.getWallet).mockRejectedValue(new Error("network"));
    renderWithClient(<BalanceDisplay />);
    expect(await screen.findByTestId("budget-balance-retry")).toBeInTheDocument();
  });
});
