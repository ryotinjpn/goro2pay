"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { fetchMetrics, type MetricsResponse } from "@/lib/api/metrics";
import { ApiError } from "@/lib/api/orders";

export function useMetrics() {
  const router = useRouter();
  const query = useQuery<MetricsResponse, Error>({
    queryKey: ["metrics"],
    queryFn: fetchMetrics,
    staleTime: 0,
    retry: 1,
  });

  // BR-FE01: ERR_NO_BUDGET_SET → /budget へリダイレクト (Q-F8=A)
  useEffect(() => {
    if (query.isError && query.error instanceof ApiError && query.error.code === "ERR_NO_BUDGET_SET") {
      router.push("/budget");
    }
  }, [query.isError, query.error, router]);

  // BR-FE01: remainingBalance === 0 → /budget-empty へリダイレクト (Q-F4=A)
  useEffect(() => {
    if (query.data?.remainingBalance === 0) {
      router.push("/budget-empty");
    }
  }, [query.data?.remainingBalance, router]);

  return query;
}
