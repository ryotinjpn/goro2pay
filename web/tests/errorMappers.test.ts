import { afterEach, describe, expect, test, vi } from "vitest";

import { mapOrderError } from "@/lib/errorMappers";
import { ApiError } from "@/lib/api/orders";
import { AuthErrorWithCode } from "@/lib/authMessages";
import { TOAST_VARIANTS } from "@/lib/toasts";

describe("mapOrderError (LC-26 / P-FE-ERR-01)", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  // ===== ApiError ベース =====

  test("ApiError(402) → BudgetEmpty への即時遷移", () => {
    const action = mapOrderError(new ApiError(402, "INSUFFICIENT_FUNDS", "no funds"));
    expect(action).toEqual({
      type: "navigate",
      path: "/budget-empty",
      transitionMs: 0,
    });
  });

  test("ApiError(409) → 自虐トースト + refresh:true (F-I4 修正、BR-C39)", () => {
    const action = mapOrderError(new ApiError(409, "IDEMPOTENCY_CONFLICT", "dup"));
    expect(action).toEqual({
      type: "toast",
      text: "もう注文済みでした…まあ、ダメ化中ですからね",
      durationMs: 5000,
      refresh: true,
    });
  });

  test("ApiError(503) → 自虐トースト (B-C2 連動: SERVICE_UNAVAILABLE)", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    const action = mapOrderError(new ApiError(503, "SERVICE_UNAVAILABLE", "wallet not ready"));
    expect(action).toEqual({
      type: "toast",
      text: TOAST_VARIANTS[0],
      durationMs: 5000,
    });
  });

  test("ApiError(500) → 自虐トースト (NFR-DEG-05)", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    const action = mapOrderError(new ApiError(500, "INTERNAL_ERROR", "boom"));
    expect(action).toEqual({
      type: "toast",
      text: TOAST_VARIANTS[0],
      durationMs: 5000,
    });
  });

  // ===== AuthErrorWithCode ベース (F-C1 修正) =====
  // apiClient.request は 401/429/500+ を AuthErrorWithCode として throw する。
  // 旧実装は ApiError しか判定しなかったため到達不能だった分岐をここでテスト。

  test("AuthErrorWithCode(SESSION_EXPIRED) → silent (Unit A hub listener が処理)", () => {
    const action = mapOrderError(new AuthErrorWithCode("SESSION_EXPIRED"));
    expect(action).toEqual({ type: "silent" });
  });

  test("AuthErrorWithCode(RATE_LIMIT_EXCEEDED) → 自虐トースト", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.5);
    const action = mapOrderError(new AuthErrorWithCode("RATE_LIMIT_EXCEEDED"));
    expect(action).toEqual({
      type: "toast",
      text: TOAST_VARIANTS[1],
      durationMs: 5000,
    });
  });

  test("AuthErrorWithCode(NETWORK_ERROR) → 自虐トースト", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.9);
    const action = mapOrderError(new AuthErrorWithCode("NETWORK_ERROR"));
    expect(action).toEqual({
      type: "toast",
      text: TOAST_VARIANTS[2],
      durationMs: 5000,
    });
  });

  test("AuthErrorWithCode(UNKNOWN) → デフォルト 自虐トースト", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    const action = mapOrderError(new AuthErrorWithCode("UNKNOWN"));
    expect(action).toEqual({
      type: "toast",
      text: TOAST_VARIANTS[0],
      durationMs: 5000,
    });
  });

  // ===== その他 =====

  test("ネットワークエラー (Error instance) → 自虐トースト", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.5);
    const action = mapOrderError(new Error("network down"));
    expect(action).toEqual({
      type: "toast",
      text: TOAST_VARIANTS[1],
      durationMs: 5000,
    });
  });

  test("不明な値 (string) → 自虐トースト", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.9);
    const action = mapOrderError("nope");
    expect(action).toEqual({
      type: "toast",
      text: TOAST_VARIANTS[2],
      durationMs: 5000,
    });
  });
});
