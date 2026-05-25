import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { useMetrics } from "@/hooks/useMetrics";
import * as metricsApi from "@/lib/api/metrics";

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({ useRouter: () => ({ push: mockPush }) }));

function wrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 } } });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
}

describe("useMetrics (LC-ME-07)", () => {
  beforeEach(() => {
    vi.spyOn(metricsApi, "fetchMetrics");
    mockPush.mockClear();
  });
  afterEach(() => vi.restoreAllMocks());

  it("正常: data が返る", async () => {
    vi.mocked(metricsApi.fetchMetrics).mockResolvedValue({
      damageCount: 5, consumptionRate: 0.5, monthlyBudget: 30000,
      remainingBalance: 15000, thresholdExceeded: false,
      summaryText: "今月のダメ化回数: 5 回、消化額 ¥15,000",
    });
    const { result } = renderHook(() => useMetrics(), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.damageCount).toBe(5);
  });

  it("remainingBalance===0 → /budget-empty へリダイレクト", async () => {
    vi.mocked(metricsApi.fetchMetrics).mockResolvedValue({
      damageCount: 10, consumptionRate: 1.0, monthlyBudget: 30000,
      remainingBalance: 0, thresholdExceeded: true,
      summaryText: "今月のダメ化回数: 10 回、消化額 ¥30,000",
    });
    renderHook(() => useMetrics(), { wrapper: wrapper() });
    await waitFor(() => expect(mockPush).toHaveBeenCalledWith("/budget-empty"));
  });
});
