"use client";

import { useEffect, useState, ReactNode } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/hooks/useAuth";

const SHOW_LOADER_AFTER_MS = 300;

interface AuthGuardProps {
  children: ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const { status } = useAuth();
  const router = useRouter();
  const [showLoader, setShowLoader] = useState(false);

  // 300ms 経過後にだけスピナーを出す (NFR-DEG-01 整合、P-PERF-01)
  useEffect(() => {
    if (status !== "loading") {
      setShowLoader(false);
      return;
    }
    const timer = setTimeout(() => setShowLoader(true), SHOW_LOADER_AFTER_MS);
    return () => clearTimeout(timer);
  }, [status]);

  // 未認証なら /login にリダイレクト
  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/login");
    }
  }, [status, router]);

  if (status === "loading") {
    if (!showLoader) return null;
    return (
      <div
        data-testid="auth-guard-loader"
        style={{
          position: "fixed",
          inset: 0,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <p>読み込み中…</p>
      </div>
    );
  }
  if (status === "unauthenticated") return null;

  return <>{children}</>;
}
