import { afterEach, describe, expect, test, vi } from "vitest";
import { TOAST_VARIANTS, getRandomToast } from "@/lib/toasts";

describe("getRandomToast (LC-27 / P-FE-TOAST-01)", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("3 種ラインナップが空でなく短すぎない", () => {
    expect(TOAST_VARIANTS).toHaveLength(3);
    for (const v of TOAST_VARIANTS) {
      expect(v.length).toBeGreaterThan(5);
    }
  });

  test("Math.random=0 で先頭文言を返す", () => {
    vi.spyOn(Math, "random").mockReturnValue(0);
    expect(getRandomToast()).toBe(TOAST_VARIANTS[0]);
  });

  test("Math.random=0.5 で中央文言を返す", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.5);
    expect(getRandomToast()).toBe(TOAST_VARIANTS[1]);
  });

  test("Math.random=0.9 で末尾文言を返す", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.9);
    expect(getRandomToast()).toBe(TOAST_VARIANTS[2]);
  });
});
