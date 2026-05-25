import { apiClient } from "../apiClient";
import { ApiError } from "./orders";

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

type ErrorResponseBody = {
  error?: { code?: string; message?: string };
};

async function parseError(res: Response): Promise<ApiError> {
  let body: ErrorResponseBody = {};
  try {
    body = (await res.json()) as ErrorResponseBody;
  } catch {
    /* JSON でない応答は status code のみで判定 */
  }
  const code = body.error?.code ?? `HTTP_${res.status}`;
  const message = body.error?.message ?? res.statusText;
  return new ApiError(res.status, code, message);
}

export async function fetchMetrics(): Promise<MetricsResponse> {
  const res = await apiClient.request({ path: "/api/metrics", method: "GET" });
  if (!res.ok) throw await parseError(res);
  return (await res.json()) as MetricsResponse;
}

export async function fetchRaiseRecommendation(): Promise<RecommendationResponse> {
  const res = await apiClient.request({ path: "/api/budget/raise/recommendation", method: "GET" });
  if (!res.ok) throw await parseError(res);
  return (await res.json()) as RecommendationResponse;
}

export async function postBudgetRaise(newMonthlyBudget: number): Promise<BudgetRaiseResponse> {
  const res = await apiClient.request({
    path: "/api/budget/raise",
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ newMonthlyBudget }),
  });
  if (!res.ok) throw await parseError(res);
  return (await res.json()) as BudgetRaiseResponse;
}
