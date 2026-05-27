import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { MetricsPanel } from "@/components/metrics/MetricsPanel";
import * as metricsApi from "@/lib/api/metrics";

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({ useRouter: () => ({ push: mockPush }) }));

function renderWithClient(ui: React.ReactElement) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 } } });
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>);
}

describe("MetricsPanel (LC-ME-09)", () => {
  beforeEach(() => {
    vi.spyOn(metricsApi, "fetchMetrics");
    mockPush.mockClear();
  });
  afterEach(() => vi.restoreAllMocks());

  it("ロード中はスケルトンを表示", () => {
    vi.mocked(metricsApi.fetchMetrics).mockImplementation(() => new Promise(() => {}));
    renderWithClient(<MetricsPanel />);
    expect(screen.getByTestId("metrics-panel-loading")).toBeInTheDocument();
  });

  it("正常: ダメ化回数と消化率を表示", async () => {
    vi.mocked(metricsApi.fetchMetrics).mockResolvedValue({
      damageCount: 7, consumptionRate: 0.5, monthlyBudget: 30000,
      remainingBalance: 15000, thresholdExceeded: false,
      summaryText: "今月のダメ化回数: 7 回、消化額 ¥15,000",
    });
    renderWithClient(<MetricsPanel />);
    expect(await screen.findByTestId("metrics-panel")).toBeInTheDocument();
    expect(screen.getByTestId("consumption-rate")).toHaveTextContent("50%");
    const bar = screen.getByTestId("consumption-bar");
    expect(bar.getAttribute("data-warn")).toBe("false");
  });

  it("ThresholdExceeded=true → data-warn=true で警告状態 (P-ME-FE-DEG-01)", async () => {
    vi.mocked(metricsApi.fetchMetrics).mockResolvedValue({
      damageCount: 12, consumptionRate: 0.82, monthlyBudget: 30000,
      remainingBalance: 5400, thresholdExceeded: true,
      summaryText: "今月のダメ化回数: 12 回、消化額 ¥24,600",
    });
    renderWithClient(<MetricsPanel />);
    const bar = await screen.findByTestId("consumption-bar");
    expect(bar.getAttribute("data-warn")).toBe("true");
    expect(screen.getByTestId("consumption-rate").getAttribute("data-warn")).toBe("true");
  });
});
