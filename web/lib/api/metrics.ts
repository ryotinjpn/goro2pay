import { apiClient } from "../apiClient";

export type MetricsResponse = {
  damageCount: number;
  consumptionRate: number;
  monthlyBudget: number;
  remainingBalance: number;
  thresholdExceeded: boolean;
  summaryText: string;
};

export type RecommendationResponse = {
  currentMonthlyBudget: number;
  recommendedMonthlyBudget: number;
};

export type BudgetRaiseResponse = {
  newMonthlyBudget: number;
  appliedFrom: string;
};

export async function fetchMetrics(): Promise<MetricsResponse> {
  const res = await apiClient.request({ path: "/api/metrics", method: "GET" });
  return (await res.json()) as MetricsResponse;
}

export async function fetchRaiseRecommendation(): Promise<RecommendationResponse> {
  const res = await apiClient.request({ path: "/api/budget/raise/recommendation", method: "GET" });
  return (await res.json()) as RecommendationResponse;
}

export async function postBudgetRaise(newMonthlyBudget: number): Promise<BudgetRaiseResponse> {
  const res = await apiClient.request({
    path: "/api/budget/raise",
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ newMonthlyBudget }),
  });
  return (await res.json()) as BudgetRaiseResponse;
}
