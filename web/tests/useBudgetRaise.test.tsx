import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { useBudgetRaise } from "@/hooks/useBudgetRaise";
import * as metricsApi from "@/lib/api/metrics";

function wrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 }, mutations: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
}

describe("useBudgetRaise (LC-ME-08)", () => {
  beforeEach(() => {
    vi.spyOn(metricsApi, "fetchRaiseRecommendation");
    vi.spyOn(metricsApi, "postBudgetRaise");
  });
  afterEach(() => vi.restoreAllMocks());

  it("recommendation を取得できる", async () => {
    vi.mocked(metricsApi.fetchRaiseRecommendation).mockResolvedValue({
      currentMonthlyBudget: 30000, recommendedMonthlyBudget: 45000,
    });
    const { result } = renderHook(() => useBudgetRaise(), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.recommendation.isSuccess).toBe(true));
    expect(result.current.recommendation.data?.recommendedMonthlyBudget).toBe(45000);
  });

  it("mutation 成功", async () => {
    vi.mocked(metricsApi.fetchRaiseRecommendation).mockResolvedValue({
      currentMonthlyBudget: 30000, recommendedMonthlyBudget: 45000,
    });
    vi.mocked(metricsApi.postBudgetRaise).mockResolvedValue({
      newMonthlyBudget: 45000, appliedFrom: "2026-06-01T00:00:00+09:00",
    });
    const { result } = renderHook(() => useBudgetRaise(), { wrapper: wrapper() });
    await act(async () => {
      result.current.mutation.mutate(45000);
    });
    await waitFor(() => expect(result.current.mutation.isSuccess).toBe(true));
  });
});
