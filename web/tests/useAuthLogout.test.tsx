import { describe, it, expect, beforeEach, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";

// vi.mock は hoist されるため factory 内では top-level 変数を参照できない。
// hoisted-safe な vi.hoisted で各 mock 関数を生成する。
const mocks = vi.hoisted(() => {
  return {
    signOut: vi.fn().mockResolvedValue(undefined),
    fetchAuthSession: vi.fn().mockResolvedValue({ tokens: undefined }),
    getCurrentUser: vi.fn().mockRejectedValue(new Error("not signed in")),
    hubListen: vi.fn().mockReturnValue(() => undefined),
    apiRequest: vi.fn(),
  };
});

vi.mock("aws-amplify/auth", () => ({
  signUp: vi.fn(),
  signIn: vi.fn(),
  signOut: mocks.signOut,
  fetchAuthSession: mocks.fetchAuthSession,
  getCurrentUser: mocks.getCurrentUser,
}));

vi.mock("aws-amplify/utils", () => ({
  Hub: { listen: mocks.hubListen },
}));

vi.mock("@/lib/apiClient", () => ({
  apiClient: { request: mocks.apiRequest },
}));

import { useAuth } from "@/hooks/useAuth";

describe("useAuth.logout", () => {
  // 順序追跡用の共通配列
  let callOrder: string[];

  beforeEach(() => {
    mocks.signOut.mockClear();
    mocks.apiRequest.mockReset();

    callOrder = [];

    mocks.apiRequest.mockImplementation(async () => {
      callOrder.push("api");
      return new Response(null, { status: 204 });
    });
    mocks.signOut.mockImplementation(async () => {
      callOrder.push("signOut");
    });
  });

  it("F-4 順序: API 監査ログ送信 → Cognito GlobalSignOut", async () => {
    const { result } = renderHook(() => useAuth());

    await act(async () => {
      await result.current.logout();
    });

    expect(mocks.apiRequest).toHaveBeenCalledWith({
      path: "/api/auth/logout",
      method: "POST",
    });
    expect(mocks.signOut).toHaveBeenCalledWith({ global: true });
    expect(callOrder).toEqual(["api", "signOut"]);
  });

  it("API 送信が失敗しても signOut は実行される", async () => {
    mocks.apiRequest.mockImplementation(async () => {
      callOrder.push("api-fail");
      throw new Error("audit log failed");
    });

    const { result } = renderHook(() => useAuth());

    await act(async () => {
      await result.current.logout();
    });

    expect(mocks.signOut).toHaveBeenCalledWith({ global: true });
    expect(callOrder).toEqual(["api-fail", "signOut"]);
  });
});
