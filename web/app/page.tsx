"use client";

import { useAuth } from "@/hooks/useAuth";
import { LandingScreen } from "@/components/auth/LandingScreen";
// Unit C MainScreen 構成要素
import { GoroButton } from "@/components/order/GoroButton";
import { OrderHistoryList } from "@/components/order/OrderHistoryList";

// ルート / は認証状態で UI 切替 (Functional Design Q-A6=A、frontend-components.md §1)。
// 認証 status を fetch するため Client Component 化が必要。
//
// 認証済みのとき Unit C の MainScreen を描画する。Unit B (BalanceDisplay) /
// Unit D (Suggest) は後続 PR で追加。
export default function HomePage() {
  const { status, user } = useAuth();

  if (status === "loading") {
    // AuthGuard を経由しないルートだが、Loading 中は空白で良い (NFR-DEG-01)
    return null;
  }

  if (status === "unauthenticated") {
    return <LandingScreen />;
  }

  // authenticated: Unit C MainScreen
  return (
    <main data-testid="main-screen" style={{ padding: 24, maxWidth: 480, margin: "0 auto" }}>
      <header style={{ marginBottom: 24, fontSize: 14, color: "#666" }}>
        ようこそ、{user?.email}
      </header>

      <section style={{ display: "flex", justifyContent: "center", marginBottom: 32 }}>
        <GoroButton />
      </section>

      <section>
        <h2 style={{ fontSize: 16, color: "#444", marginBottom: 12 }}>履歴</h2>
        <OrderHistoryList />
      </section>
    </main>
  );
}
