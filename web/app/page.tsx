"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAuth } from "@/hooks/useAuth";
import { useWallet } from "@/hooks/useWallet";

import { LandingScreen } from "@/components/auth/LandingScreen";
import { MainScreen } from "@/components/order/MainScreen";

export default function HomePage() {
  const { status } = useAuth();
  const router = useRouter();
  const isAuthenticated = status === "authenticated";
  const { monthlyBudget, isLoading: isWalletLoading } = useWallet();
  const needsBudgetSetup = isAuthenticated && !isWalletLoading && monthlyBudget === 0;

  useEffect(() => {
    if (needsBudgetSetup) router.replace("/budget");
  }, [needsBudgetSetup, router]);

  if (status === "loading") {
    return null;
  }

  if (status === "unauthenticated") {
    return <LandingScreen />;
  }

  if (needsBudgetSetup) {
    return null;
  }

  return <MainScreen />;
}
