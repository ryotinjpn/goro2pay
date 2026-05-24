import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { getDefaultStore } from "jotai";

import { useToast } from "@/hooks/useToast";
import { toastsAtom } from "@/state/toastAtoms";

describe("useToast (LC-29 / P-FE-TOAST-02)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    getDefaultStore().set(toastsAtom, []);
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  test("showToast で atom にトーストが追加される", () => {
    const { result } = renderHook(() => useToast());
    act(() => result.current.showToast("hello"));

    const toasts = getDefaultStore().get(toastsAtom);
    expect(toasts).toHaveLength(1);
    expect(toasts[0].text).toBe("hello");
  });

  test("durationMs 経過後に自動消去される", () => {
    const { result } = renderHook(() => useToast());
    act(() => result.current.showToast("auto-clear", 3000));

    expect(getDefaultStore().get(toastsAtom)).toHaveLength(1);

    act(() => {
      vi.advanceTimersByTime(2999);
    });
    expect(getDefaultStore().get(toastsAtom)).toHaveLength(1);

    act(() => {
      vi.advanceTimersByTime(2);
    });
    expect(getDefaultStore().get(toastsAtom)).toHaveLength(0);
  });

  test("複数 showToast はキューに蓄積される", () => {
    const { result } = renderHook(() => useToast());
    act(() => {
      result.current.showToast("a", 5000);
      result.current.showToast("b", 5000);
      result.current.showToast("c", 5000);
      result.current.showToast("d", 5000);
    });
    expect(getDefaultStore().get(toastsAtom)).toHaveLength(4);
  });
});
