"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAuth } from "@/hooks/useAuth";
import { WALLET_QUERY_OPTIONS } from "@/hooks/useWallet";
import { ApiError } from "@/lib/api/orders";
import { type WalletResponse } from "@/lib/api/wallet";

import { LandingScreen } from "@/components/auth/LandingScreen";
import { MainScreen } from "@/components/order/MainScreen";

export default function HomePage() {
  const { status } = useAuth();
  const router = useRouter();
  const isAuthenticated = status === "authenticated";

  // useWallet ではなく useQuery を直接 subscribe するのは
  // (1) status が unauthenticated/loading のとき fetch を抑止する enabled が必要
  // (2) 新規ユーザは 404 を投げる仕様 (wallet.ts) なので isError の中身を判定する必要がある
  // ためで、BalanceDisplay と同じパターン。
  const wallet = useQuery<WalletResponse>({
    ...WALLET_QUERY_OPTIONS,
    enabled: isAuthenticated,
  });

  // 新規ユーザは 404 (WALLET_NOT_FOUND) を返す仕様 (wallet.ts:47-56) なので
  // 404 のときは「予算未設定」と同等に扱う。それ以外のエラーは MainScreen に渡し、
  // BalanceDisplay 側のエラー UI でリトライさせる。
  const isWalletNotFound =
    wallet.error instanceof ApiError && wallet.error.status === 404;
  const isOtherError = wallet.isError && !isWalletNotFound;
  const isWalletReady = wallet.data !== undefined || isOtherError || isWalletNotFound;
  const needsBudgetSetup =
    isAuthenticated &&
    isWalletReady &&
    (isWalletNotFound || (wallet.data?.monthlyBudget ?? 0) === 0);

  useEffect(() => {
    if (needsBudgetSetup) router.replace("/budget");
  }, [needsBudgetSetup, router]);

  if (status === "loading") {
    return null;
  }

  if (status === "unauthenticated") {
    return <LandingScreen />;
  }

  // wallet 確定前のフラッシュを防ぐ。エラーで確定済みなら通す。
  if (!isWalletReady) {
    return null;
  }

  if (needsBudgetSetup) {
    return null;
  }

  return <MainScreen />;
}
