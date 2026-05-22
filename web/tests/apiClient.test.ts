import { describe, it, expect, beforeEach, vi } from "vitest";
import { getDefaultStore } from "jotai";

import { apiClient, _resetSessionExpiredHandling } from "@/lib/apiClient";
import { sessionExpiredAtom } from "@/state/auth";
import { AuthErrorWithCode } from "@/lib/authMessages";

// Amplify Auth.fetchAuthSession のモック
vi.mock("aws-amplify/auth", () => ({
  fetchAuthSession: vi.fn(async () => ({
    tokens: {
      accessToken: { toString: () => "test-access-token" },
    },
  })),
}));

describe("apiClient", () => {
  beforeEach(() => {
    _resetSessionExpiredHandling();
    const store = getDefaultStore();
    store.set(sessionExpiredAtom, null);
    vi.restoreAllMocks();
  });

  it("Authorization: Bearer ヘッダで AccessToken を付与する", async () => {
    const fetchSpy = vi.spyOn(global, "fetch").mockResolvedValue(
      new Response(null, { status: 200 })
    );
    await apiClient.request({ path: "/api/wallet" });
    expect(fetchSpy).toHaveBeenCalledTimes(1);
    const callArgs = fetchSpy.mock.calls[0];
    const headers = (callArgs[1]?.headers ?? new Headers()) as Headers;
    expect(headers.get("Authorization")).toBe("Bearer test-access-token");
  });

  it("401 を検出したら sessionExpiredAtom が起動する", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue(new Response(null, { status: 401 }));
    await expect(apiClient.request({ path: "/api/wallet" })).rejects.toBeInstanceOf(AuthErrorWithCode);
    const store = getDefaultStore();
    expect(store.get(sessionExpiredAtom)).not.toBeNull();
  });

  it("401 が複数回来ても二重発火しない (atom は最初の値を保持)", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue(new Response(null, { status: 401 }));
    await expect(apiClient.request({ path: "/api/a" })).rejects.toBeInstanceOf(AuthErrorWithCode);
    const firstSnapshot = getDefaultStore().get(sessionExpiredAtom);
    expect(firstSnapshot).not.toBeNull();

    await expect(apiClient.request({ path: "/api/b" })).rejects.toBeInstanceOf(AuthErrorWithCode);
    const secondSnapshot = getDefaultStore().get(sessionExpiredAtom);
    expect(secondSnapshot).toBe(firstSnapshot);
  });

  it("429 を検出したら RATE_LIMIT_EXCEEDED を throw する", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue(new Response(null, { status: 429 }));
    await expect(apiClient.request({ path: "/api/wallet" })).rejects.toMatchObject({
      code: "RATE_LIMIT_EXCEEDED",
    });
  });

  it("5xx を検出したら NETWORK_ERROR を throw する", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue(new Response(null, { status: 503 }));
    await expect(apiClient.request({ path: "/api/wallet" })).rejects.toMatchObject({
      code: "NETWORK_ERROR",
    });
  });
});
