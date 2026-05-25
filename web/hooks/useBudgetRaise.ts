"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  fetchRaiseRecommendation,
  postBudgetRaise,
  type BudgetRaiseResponse,
  type RecommendationResponse,
} from "@/lib/api/metrics";

export function useBudgetRaise() {
  const queryClient = useQueryClient();

  const recommendation = useQuery<RecommendationResponse, Error>({
    queryKey: ["budget-raise-recommendation"],
    queryFn: fetchRaiseRecommendation,
    staleTime: 0,
    retry: 1,
  });

  const mutation = useMutation<BudgetRaiseResponse, Error, number>({
    mutationFn: postBudgetRaise,
    retry: 0,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["metrics"] });
      queryClient.invalidateQueries({ queryKey: ["balance"] });
      queryClient.invalidateQueries({ queryKey: ["budget-raise-recommendation"] });
    },
  });

  return { recommendation, mutation };
}
