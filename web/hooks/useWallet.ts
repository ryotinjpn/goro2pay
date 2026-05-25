// LC-BUDGET-08 useWalletHook
//
// 凍結 IF (unit-interfaces.md §9): { balance, monthlyBudget, isLoading, refetch() }
// queryKey: ['balance'] (unit-interfaces.md §9.1 クロスユニット契約)
// staleTime: 30s (PR-B-05)
// placeholderData: keepPreviousData (P-DEG-01、2 回目以降のフェッチ中は前回値を保持)
"use client";

import { keepPreviousData, useQuery } from "@tanstack/react-query";

import { getWallet, type WalletResponse } from "@/lib/api/wallet";

/**
 * Wallet TanStack Query 共通設定 (Code Review Important 7)。
 * `useWallet` と `BalanceDisplay` 両方で参照することで設定の二重管理を回避。
 *
 * 注: `useWallet` は凍結 IF §9 の都合 (`isFetching` / `isError` 非公開) で
 * `BalanceDisplay` は内部で直接 useQuery を呼ぶ。両者が同じ queryKey で同じ設定を
 * 使うことを本定数で保証する。
 */
export const WALLET_QUERY_KEY = ["balance"] as const;
// queryFn は import 時に固定参照すると vi.spyOn が効かないため lazy lookup する。
export const WALLET_QUERY_OPTIONS = {
  queryKey: WALLET_QUERY_KEY,
  queryFn: () => getWallet(),
  staleTime: 30_000,
  placeholderData: keepPreviousData,
  retry: 0,
} as const;

export function useWallet() {
  const query = useQuery<WalletResponse>(WALLET_QUERY_OPTIONS);
  return {
    balance: query.data?.balance ?? 0,
    monthlyBudget: query.data?.monthlyBudget ?? 0,
    isLoading: query.isLoading,
    refetch: () => {
      void query.refetch();
    },
  };
}
