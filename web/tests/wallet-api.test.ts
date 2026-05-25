// Unit B Wallet API ラッパ (web/lib/api/wallet.ts) の単体テスト。
// apiClient.request を spy で mock する。
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import * as apiClientMod from "@/lib/apiClient";
import { ApiError } from "@/lib/api/orders";
import { getWallet, postWalletBudget } from "@/lib/api/wallet";

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("getWallet", () => {
  beforeEach(() => {
    vi.spyOn(apiClientMod.apiClient, "request");
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("200 で WalletResponse を返す", async () => {
    vi.mocked(apiClientMod.apiClient.request).mockResolvedValue(
      jsonResponse(200, {
        balance: 28800,
        monthlyBudget: 30000,
        updatedAt: "2026-05-24T00:00:00Z",
      }),
    );
    const w = await getWallet();
    expect(w.balance).toBe(28800);
    expect(w.monthlyBudget).toBe(30000);
  });

  it("404 で ApiError を throw する", async () => {
    vi.mocked(apiClientMod.apiClient.request).mockResolvedValue(
      jsonResponse(404, { code: "WALLET_NOT_FOUND", message: "予算が未設定です" }),
    );
    await expect(getWallet()).rejects.toBeInstanceOf(ApiError);
  });
});

describe("postWalletBudget", () => {
  beforeEach(() => {
    vi.spyOn(apiClientMod.apiClient, "request");
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("200 で SetBudgetResponse を返す", async () => {
    vi.mocked(apiClientMod.apiClient.request).mockResolvedValue(
      jsonResponse(200, {
        monthlyBudget: 30000,
        appliedFrom: "2026-05-24T00:00:00Z",
      }),
    );
    const r = await postWalletBudget(30000);
    expect(r.monthlyBudget).toBe(30000);
  });

  it("400 VALIDATION_FAILED は ApiError(400) として伝播", async () => {
    vi.mocked(apiClientMod.apiClient.request).mockResolvedValue(
      jsonResponse(400, { code: "VALIDATION_FAILED", message: "範囲外です" }),
    );
    await expect(postWalletBudget(999_999)).rejects.toMatchObject({
      status: 400,
      code: "VALIDATION_FAILED",
    });
  });
});
