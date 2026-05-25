import { describe, expect, it } from "vitest";
import { createStore } from "jotai";

import { insufficientBalanceAtom } from "@/state/budget";

describe("insufficientBalanceAtom", () => {
  it("初期値は false", () => {
    const store = createStore();
    expect(store.get(insufficientBalanceAtom)).toBe(false);
  });

  it("true に更新できる", () => {
    const store = createStore();
    store.set(insufficientBalanceAtom, true);
    expect(store.get(insufficientBalanceAtom)).toBe(true);
  });

  it("true → false に戻せる", () => {
    const store = createStore();
    store.set(insufficientBalanceAtom, true);
    store.set(insufficientBalanceAtom, false);
    expect(store.get(insufficientBalanceAtom)).toBe(false);
  });
});
