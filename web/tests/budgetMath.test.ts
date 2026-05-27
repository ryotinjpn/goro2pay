// web/tests/budgetMath.test.ts
import { describe, it, expect } from "vitest";
import * as fc from "fast-check";
import { recommendNextBudget } from "@/lib/budgetMath";

describe("recommendNextBudget", () => {
  it("returns 50000 for 30000", () => {
    expect(recommendNextBudget(30000)).toBe(50000);
  });
  it("returns 80000 for 50000", () => {
    expect(recommendNextBudget(50000)).toBe(80000);
  });
  it("returns 130000 for 80000", () => {
    expect(recommendNextBudget(80000)).toBe(130000);
  });

  it("rounds to nearest 10000", () => {
    fc.assert(
      fc.property(fc.integer({ min: 10000, max: 1_000_000 }), (current) => {
        const next = recommendNextBudget(current);
        expect(next % 10000).toBe(0);
      }),
    );
  });

  it("always increases (next > current)", () => {
    fc.assert(
      fc.property(fc.integer({ min: 10000, max: 1_000_000 }), (current) => {
        expect(recommendNextBudget(current)).toBeGreaterThan(current);
      }),
    );
  });

  it("approximates 5/3 ratio (within rounding tolerance)", () => {
    fc.assert(
      fc.property(fc.integer({ min: 30000, max: 500_000 }), (current) => {
        const next = recommendNextBudget(current);
        const ratio = next / current;
        expect(ratio).toBeGreaterThanOrEqual(1.5);
        expect(ratio).toBeLessThanOrEqual(2.0);
      }),
    );
  });
});
