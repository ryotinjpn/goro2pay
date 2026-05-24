// LC-22 useOrder / P-FE-ERR-01
//
// PlaceOrder mutation の React hook。
// エラーハンドリング (mapOrderError)、連打抑制 (useDisableLock)、
// 共有 query key invalidate (['orderHistory'] / ['balance']) を統合する。
"use client";

import {
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { startTransition } from "react";

import { placeOrder, type PlaceOrderRequest, type PlaceOrderResponse } from "@/lib/api/orders";
import { mapOrderError } from "@/lib/errorMappers";

import { useDisableLock } from "./useDisableLock";
import { useToast } from "./useToast";

const DISABLE_LOCK_MS = 1000;

/**
 * useOrder mutation hook.
 *
 * 戻り値:
 *   - mutate / mutateAsync : useMutation 標準
 *   - isPending            : useMutation 標準 (送信中)
 *   - disabled             : isPending OR isLocked (連打抑制 1 秒含む)
 *
 * NFRC-C22 / NFRC-C19 整合:
 *   - retry: 0
 *   - onSuccess: triggerLock + invalidate(orderHistory) + invalidate(balance)
 *   - onError  : triggerLock + mapOrderError → router.push / showToast / silent
 */
export function useOrder() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { showToast } = useToast();
  const { isLocked, triggerLock } = useDisableLock(DISABLE_LOCK_MS);

  const mutation = useMutation<PlaceOrderResponse, unknown, PlaceOrderRequest>({
    mutationFn: placeOrder,
    retry: 0,
    onSuccess: () => {
      triggerLock();
      queryClient.invalidateQueries({ queryKey: ["orderHistory"] });
      queryClient.invalidateQueries({ queryKey: ["balance"] }); // unit-interfaces.md §9.1
    },
    onError: (err) => {
      triggerLock();
      const action = mapOrderError(err);
      // F-I4 修正: action.refresh = true の場合は履歴/残高を invalidate。
      // 409 冪等性衝突時に履歴を最新化することで「もう注文済み」状態を反映。
      if (action.refresh) {
        queryClient.invalidateQueries({ queryKey: ["orderHistory"] });
        queryClient.invalidateQueries({ queryKey: ["balance"] });
      }
      switch (action.type) {
        case "navigate":
          startTransition(() => router.push(action.path));
          return;
        case "toast":
          showToast(action.text, action.durationMs);
          return;
        case "silent":
          return;
      }
    },
  });

  return {
    ...mutation,
    disabled: mutation.isPending || isLocked,
  };
}
