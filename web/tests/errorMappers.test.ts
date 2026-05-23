import { afterEach, describe, expect, test, vi } from "vitest";

import { mapOrderError } from "@/lib/errorMappers";
import { ApiError } from "@/lib/api/orders";
import { TOAST_VARIANTS } from "@/lib/toasts";

describe("mapOrderError (LC-26 / P-FE-ERR-01)", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("402 → BudgetEmpty への即時遷移", () => {
    const action = mapOrderError(new ApiError(402, "INSUFFICIENT_FUNDS", "no funds"));
    expect(action).toEqual({
      type: "navigate",
      path: "/budget-empty",
      transitionMs: 0,
    });
  });

  test("409 → silent (冪等性衝突は実質成功扱い、BR-C39)", () => {
    const action = mapOrderError(new ApiError(409, "IDEMPOTENCY_CONFLICT", "dup"));
    expect(action).toEqual({ type: "silent" });
  });

  test("401 → silent (Unit A の hub listener が拾う)", () => {
    const action = mapOrderError(new ApiError(401, "UNAUTHORIZED", "expired"));
    expect(action).toEqual({ type: "silent" });
  });

  test("500 → 自虐トースト (NFR-DEG-05)", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    const action = mapOrderError(new ApiError(500, "INTERNAL_ERROR", "boom"));
    expect(action).toEqual({
      type: "toast",
      text: TOAST_VARIANTS[0],
      durationMs: 5000,
    });
  });

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
