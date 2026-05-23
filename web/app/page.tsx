"use client";

import { useAuth } from "@/hooks/useAuth";
import { LandingScreen } from "@/components/auth/LandingScreen";

// ルート / は認証状態で UI 切替 (Functional Design Q-A6=A、frontend-components.md §1)。
// 認証 status を fetch するため Client Component 化が必要。
//
// NOTE: 将来 MainScreen (Unit B/C/D/E) を実装する際、表示物が SSR 可能なら
// MainScreen 自体を Server Component として切り出し、本 page.tsx は
// 「LandingScreen / MainScreen の選択ロジックだけを担う Client Component」に
// 抑える方針が望ましい。MainScreen 内の interactive 部分のみを子の
// Client Component に分けることで bundle 縮小に寄与する。
//
// MainScreen (Unit C/B/D/E) は本 PR の Unit A スコープ外なので placeholder。
export default function HomePage() {
  const { status, user } = useAuth();

  if (status === "loading") {
    // AuthGuard を経由しないルートだが、Loading 中は空白で良い (NFR-DEG-01)
    return null;
  }

  if (status === "unauthenticated") {
    return <LandingScreen />;
  }

  // authenticated: Unit A スコープ外、Unit B/C/D/E が後続で実装
  return (
    <main data-testid="main-screen-placeholder" style={{ padding: 24 }}>
      <h1>ようこそ、{user?.email}</h1>
      <p>Unit B/C/D/E の実装で MainScreen がここに表示されます。</p>
    </main>
  );
}
