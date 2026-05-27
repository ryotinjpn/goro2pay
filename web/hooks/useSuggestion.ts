// LC-SUGGEST-10 useSuggestion
//
// 凍結 IF (unit-interfaces.md §9): { suggestion: {...} | null, isLoading }
// queryKey: ['suggestion']
// マウント時 1 回のみ取得、再評価なし (BR-D14 / NFRD-D17)。
// retry: 0、取得失敗は fetchSuggestion 側で hasSuggestion:false に丸め。
"use client";

import { useQuery } from "@tanstack/react-query";

import { fetchSuggestion, type SuggestionResponse } from "@/lib/api/suggest";

export const SUGGESTION_QUERY_KEY = ["suggestion"] as const;

export const SUGGESTION_QUERY_OPTIONS = {
  queryKey: SUGGESTION_QUERY_KEY,
  queryFn: () => fetchSuggestion(),
  staleTime: Infinity, // マウント時 1 回、ポーリング・再評価なし (NFRD-D17)
  refetchOnWindowFocus: false,
  retry: 0,
} as const;

export function useSuggestion() {
  const query = useQuery<SuggestionResponse>(SUGGESTION_QUERY_OPTIONS);
  return {
    suggestion: query.data ?? null,
    isLoading: query.isLoading,
  };
}
