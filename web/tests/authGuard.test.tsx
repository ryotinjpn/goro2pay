import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, act } from "@testing-library/react";

import { AuthGuard } from "@/components/auth/AuthGuard";

// useAuth の status を test ごとに差し替えられるよう mock 化
const useAuthMock = vi.fn();
vi.mock("@/hooks/useAuth", () => ({
  useAuth: () => useAuthMock(),
}));

const replaceMock = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: replaceMock, push: replaceMock }),
}));

describe("AuthGuard", () => {
  beforeEach(() => {
    replaceMock.mockClear();
    useAuthMock.mockReset();
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("loading 中は 300ms 経過するまでスピナーを描画しない (P-PERF-01)", () => {
    useAuthMock.mockReturnValue({ status: "loading" });
    render(
      <AuthGuard>
        <div data-testid="protected">secret</div>
      </AuthGuard>
    );

    // 直後はスピナー無し (loading flicker 抑止)
    expect(screen.queryByTestId("auth-guard-loader")).toBeNull();

    // 200ms 経過時点でもまだ表示されない
    act(() => {
      vi.advanceTimersByTime(200);
    });
    expect(screen.queryByTestId("auth-guard-loader")).toBeNull();

    // 350ms (>= 300ms) 経過するとスピナーが描画される
    act(() => {
      vi.advanceTimersByTime(150);
    });
    expect(screen.getByTestId("auth-guard-loader")).toBeInTheDocument();
  });

  it("authenticated なら children をそのまま描画する", () => {
    useAuthMock.mockReturnValue({ status: "authenticated" });
    render(
      <AuthGuard>
        <div data-testid="protected">secret</div>
      </AuthGuard>
    );
    expect(screen.getByTestId("protected")).toBeInTheDocument();
    expect(replaceMock).not.toHaveBeenCalled();
  });

  it("unauthenticated なら /login に router.replace する", () => {
    useAuthMock.mockReturnValue({ status: "unauthenticated" });
    render(
      <AuthGuard>
        <div data-testid="protected">secret</div>
      </AuthGuard>
    );
    expect(screen.queryByTestId("protected")).toBeNull();
    expect(replaceMock).toHaveBeenCalledWith("/login");
  });
});
