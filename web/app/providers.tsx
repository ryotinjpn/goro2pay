"use client";

import { ReactNode, useEffect, useRef, useState } from "react";
import { Amplify } from "aws-amplify";
import { Provider as JotaiProvider } from "jotai";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { amplifyConfig } from "@/lib/amplifyConfig";
import { setupAuthHubListener } from "@/lib/authHubListener";
import { SessionExpiredModalHost } from "@/components/auth/SessionExpiredModalHost";

// AppProviders は アプリ全体に必要な Provider 群を組み合わせる。
// - Amplify.configure を 1 度だけ呼ぶ
// - Auth Hub Listener を 1 度だけ登録
// - Jotai / TanStack Query Provider を mount
// - SessionExpiredModalHost を 1 つだけ mount
export function AppProviders({ children }: { children: ReactNode }) {
  const configuredRef = useRef(false);
  const [queryClient] = useState(() => new QueryClient());

  if (!configuredRef.current) {
    Amplify.configure(amplifyConfig, { ssr: true });
    configuredRef.current = true;
  }

  useEffect(() => {
    setupAuthHubListener();
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
