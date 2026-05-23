"use client";

import { ReactNode, useEffect, useState } from "react";
import { Provider as JotaiProvider } from "jotai";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { ensureAmplifyConfigured } from "@/lib/amplifyConfig";
import { setupAuthHubListener, teardownAuthHubListener } from "@/lib/authHubListener";
import { SessionExpiredModalHost } from "@/components/auth/SessionExpiredModalHost";

// AppProviders は アプリ全体に必要な Provider 群を組み合わせる。
// - Amplify.configure: lib/amplifyConfig.ts の module-level guard で 1 度だけ実行
// - Auth Hub Listener を 1 度だけ登録
// - Jotai / TanStack Query Provider を mount
// - SessionExpiredModalHost を 1 つだけ mount

// Amplify は module top で configure する。React render 中の副作用呼出を避け、
// StrictMode の二重 render や複数 mount でも 1 度だけ実行されるよう保証する。
ensureAmplifyConfigured();

export function AppProviders({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => new QueryClient());

  useEffect(() => {
    // setupAuthHubListener はモジュール内の registered フラグで二重登録を
    // 防ぐ。AppProviders unmount や HMR reload 時には teardown して
    // フラグごと reset し、再 mount で setup が再度動くようにする。
    setupAuthHubListener();
    return () => {
      teardownAuthHubListener();
    };
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <JotaiProvider>
        <SessionExpiredModalHost />
        {children}
      </JotaiProvider>
    </QueryClientProvider>
  );
}
