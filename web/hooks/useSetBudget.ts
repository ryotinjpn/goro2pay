// useSetBudget mutation hook (LC-BUDGET-12 連携)。
//
// 成功時に queryKey: ['balance'] を invalidate し、MainScreen 残高を即時更新する
// (PR-B-05 + クロスユニット契約)。MainScreen `/` への遷移も合わせて行う。
"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

import { postWalletBudget, type SetBudgetResponse } from "@/lib/api/wallet";

export function useSetBudget() {
  const queryClient = useQueryClient();
  const router = useRouter();
  const mutation = useMutation<SetBudgetResponse, unknown, number>({
    mutationFn: (monthlyBudget: number) => postWalletBudget(monthlyBudget),
    retry: 0,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["balance"] });
      router.push("/");
    },
  });
  return {
    mutate: mutation.mutate,
    mutateAsync: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
