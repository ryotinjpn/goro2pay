// LC-23 useOrderHistory / P-FE-LOAD-01
//
// 注文履歴の取得 hook。
// placeholderData: keepPreviousData で再 fetch 中も前回データ表示維持
// (NFR-DEG-03 常時可視化、NFRC-C19)。
"use client";

import {
  keepPreviousData,
  useQuery,
} from "@tanstack/react-query";

import {
  fetchOrderHistory,
  type OrderRecord,
} from "@/lib/api/orders";

/**
 * useOrderHistory は GET /api/orders を購読する hook。
 *
 * 設定値 (NFRC-C19):
 *   - staleTime           : 60_000 ms (履歴は注文時のみ変動、Unit B 残高 30s より長め)
 *   - gcTime              : 300_000 ms (5 分、デフォルト)
 *   - refetchOnWindowFocus: true (タブ切替時に stale なら再 fetch)
 *   - placeholderData     : keepPreviousData (再 fetch 中も前回値を維持)
 */
export function useOrderHistory(limit = 20) {
  return useQuery<OrderRecord[]>({
    queryKey: ["orderHistory", limit],
    queryFn: () => fetchOrderHistory(limit),
    placeholderData: keepPreviousData,
    staleTime: 60_000,
    gcTime: 300_000,
    refetchOnWindowFocus: true,
  });
}
