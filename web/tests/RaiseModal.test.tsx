import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { RaiseModal } from "@/components/metrics/RaiseModal";
import * as metricsApi from "@/lib/api/metrics";

function renderWithClient(ui: React.ReactElement) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>);
}

describe("RaiseModal (LC-ME-11 / P-ME-FE-DEG-01)", () => {
  const onClose = vi.fn();

  beforeEach(() => {
    vi.spyOn(metricsApi, "fetchRaiseRecommendation");
    vi.spyOn(metricsApi, "postBudgetRaise");
    onClose.mockClear();
  });
  afterEach(() => vi.restoreAllMocks());

  it("isOpen=false のとき何も表示しない", () => {
    vi.mocked(metricsApi.fetchRaiseRecommendation).mockResolvedValue({
      currentMonthlyBudget: 30000, recommendedMonthlyBudget: 45000,
    });
    renderWithClient(<RaiseModal isOpen={false} onClose={onClose} />);
    expect(screen.queryByTestId("raise-modal")).not.toBeInTheDocument();
  });

  it("isOpen=true のときモーダルを表示", async () => {
    vi.mocked(metricsApi.fetchRaiseRecommendation).mockResolvedValue({
      currentMonthlyBudget: 30000, recommendedMonthlyBudget: 45000,
    });
    renderWithClient(<RaiseModal isOpen={true} onClose={onClose} />);
    expect(screen.getByTestId("raise-modal")).toBeInTheDocument();
    expect(await screen.findByText(/45,000/)).toBeInTheDocument();
  });

  it("「今月はがんばる」で onClose が呼ばれる", async () => {
    vi.mocked(metricsApi.fetchRaiseRecommendation).mockResolvedValue({
      currentMonthlyBudget: 30000, recommendedMonthlyBudget: 45000,
    });
    renderWithClient(<RaiseModal isOpen={true} onClose={onClose} />);
    await screen.findByTestId("raise-modal-reject");
    fireEvent.click(screen.getByTestId("raise-modal-reject"));
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("推奨額が表示されたあと「増額する」ボタンが有効になる", async () => {
    vi.mocked(metricsApi.fetchRaiseRecommendation).mockResolvedValue({
      currentMonthlyBudget: 30000, recommendedMonthlyBudget: 45000,
    });
    vi.mocked(metricsApi.postBudgetRaise).mockResolvedValue({
      newMonthlyBudget: 45000, appliedFrom: "2026-06-01T00:00:00+09:00",
    });
    renderWithClient(<RaiseModal isOpen={true} onClose={onClose} />);
    await screen.findByText(/45,000/);
    await waitFor(() =>
      expect(screen.getByTestId("raise-modal-accept")).not.toBeDisabled(),
    );
  });
});
