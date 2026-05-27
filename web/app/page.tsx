"use client";

import { useAuth } from "@/hooks/useAuth";

import { LandingScreen } from "@/components/auth/LandingScreen";
import { MainScreen } from "@/components/order/MainScreen";

export default function HomePage() {
  const { status } = useAuth();

  if (status === "loading") {
    return null;
  }

  if (status === "unauthenticated") {
    return <LandingScreen />;
  }

  return <MainScreen />;
}
